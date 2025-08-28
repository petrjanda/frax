package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	_ "embed"
)

// Agent represents an agent that can use tools and interact with an LLM
type Agent struct {
	llm          LLM
	tools        []Tool
	maxRetries   int
	retryDelay   time.Duration
	retryBackoff float64
}

// AgentOpts represents options for configuring an agent
type AgentOpts = func(*Agent)

// WithMaxRetries sets the maximum number of retries for tool calls
func WithMaxRetries(maxRetries int) AgentOpts {
	return func(a *Agent) {
		a.maxRetries = maxRetries
	}
}

// WithRetryDelay sets the initial delay between retries
func WithRetryDelay(delay time.Duration) AgentOpts {
	return func(a *Agent) {
		a.retryDelay = delay
	}
}

// WithRetryBackoff sets the exponential backoff multiplier for retries
func WithRetryBackoff(backoff float64) AgentOpts {
	return func(a *Agent) {
		a.retryBackoff = backoff
	}
}

// NewAgent creates a new agent with the given LLM and tools
func NewAgent(llm LLM, tools []Tool, opts ...AgentOpts) LLM {
	a := &Agent{
		llm:          llm,
		tools:        tools,
		maxRetries:   3,                      // Default: 3 retries
		retryDelay:   100 * time.Millisecond, // Default: 100ms initial delay
		retryBackoff: 2.0,                    // Default: 2x backoff
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Loop processes the conversation loop, handling tool calls and LLM responses
func (a *Agent) Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	req := request.Clone(
		WithTools(a.tools...),
		WithToolUsage(AutoToolSelection()),
	)

	response, err := a.llm.Invoke(ctx, req)
	if err != nil {
		return nil, err
	}

	toolCalls := response.ToolCalls()
	if len(toolCalls) > 0 {
		for _, toolCall := range toolCalls {
			message, err := a.CallTool(ctx, toolCall)
			if err != nil {
				response.AddMessage(NewToolResultErrorMessage(toolCall, err.Error()))
			} else {
				response.AddMessage(message)
			}
		}

		return a.Invoke(ctx, NewLLMRequest(append(request.History, response.Messages...)))
	}

	return response, nil
}

// CallTool executes a tool call with retry logic using a formatter approach
func (a *Agent) CallTool(ctx context.Context, toolCall *ToolCall) (Message, error) {
	// Find the tool to get its input schema
	targetTool, err := a.findTool(toolCall.Name)
	if err != nil {
		return nil, err
	}

	return a.executeToolWithRetry(ctx, toolCall, targetTool)
}

// findTool finds a tool by name from the agent's tool list
func (a *Agent) findTool(name string) (Tool, error) {
	for _, t := range a.tools {
		if t.Name() == name {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}

// executeToolWithRetry manages the retry loop for tool execution
func (a *Agent) executeToolWithRetry(ctx context.Context, toolCall *ToolCall, targetTool Tool) (Message, error) {
	var lastErr error
	delay := a.retryDelay
	currentToolCall := toolCall // Create a local copy

	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry (exponential backoff)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * a.retryBackoff)
			}
		}

		// Try to execute the tool
		result, err := a.executeToolAttempt(ctx, currentToolCall, targetTool)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// If this is the last attempt, don't retry
		if attempt == a.maxRetries {
			break
		}

		// Handle tool failure and get corrected parameters
		correctedToolCall, shouldContinue := a.handleToolFailure(ctx, currentToolCall, targetTool, attempt, err)
		if !shouldContinue {
			break
		}

		currentToolCall = correctedToolCall
	}

	return nil, fmt.Errorf("tool call failed after %d retries: %w", a.maxRetries+1, lastErr)
}

// executeToolAttempt executes a single tool attempt
func (a *Agent) executeToolAttempt(ctx context.Context, toolCall *ToolCall, targetTool Tool) (Message, error) {
	result, err := targetTool.Run(ctx, toolCall.Args)
	if err != nil {
		return nil, err
	}

	return &ToolResultMessage{
		ToolCall: toolCall,
		Result:   result,
	}, nil
}

// handleToolFailure handles tool failure and attempts to get corrected parameters
func (a *Agent) handleToolFailure(ctx context.Context, toolCall *ToolCall, targetTool Tool, attempt int, err error) (*ToolCall, bool) {
	// Log the retry attempt for debugging
	slog.Info("Tool call failed, asking LLM to correct parameters",
		"tool", toolCall.Name,
		"attempt", attempt+1,
		"max_attempts", a.maxRetries,
		"error", err.Error(),
	)

	// Get corrected parameters from the LLM
	correctedArgs, err := a.correctToolCall(ctx, toolCall, targetTool, err)
	if err != nil {
		slog.Warn("Failed to get corrected parameters from LLM, continuing to next retry attempt",
			"tool", toolCall.Name,
			"attempt", attempt+1,
			"error", err.Error(),
		)
		return toolCall, true // Continue with current parameters
	}

	// Update the tool call with corrected parameters
	updatedToolCall := a.updateToolCallArgs(toolCall, correctedArgs)

	slog.Info("LLM provided corrected tool call parameters",
		"tool", updatedToolCall.Name,
		"corrected_params", prettyJSON(correctedArgs),
	)

	return updatedToolCall, true
}

//go:embed prompts/tool_call_correction.txt
var correctionPromptFormat string

// correctToolCall uses the LLM to get corrected parameters for a failed tool call
func (a *Agent) correctToolCall(ctx context.Context, toolCall *ToolCall, targetTool Tool, originalErr error) (json.RawMessage, error) {
	errorMessage := NewUserMessage(fmt.Sprintf(
		correctionPromptFormat,
		toolCall.Name,
		originalErr.Error(),
		prettyJSON(toolCall.Args),
	))

	retryRequest := NewLLMRequest(NewHistory(errorMessage))
	formatter := NewBaseLLMWithStructuredOutput(targetTool.InputSchemaRaw(), a.llm)

	// Get corrected parameters from the LLM
	retryResponse, retryErr := formatter.Invoke(ctx, retryRequest)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to get corrected parameters: %w", retryErr)
	}

	// Extract the corrected parameters from the LLM response
	if len(retryResponse.Messages) > 0 {
		if userMessage, ok := retryResponse.Messages[0].(*UserMessage); ok {
			slog.Info("Retry response", "response", userMessage.Content)
			return []byte(userMessage.Content), nil
		}
	}

	return nil, fmt.Errorf("LLM did not provide corrected parameters")
}

// HELPERS
func (a *Agent) updateToolCallArgs(originalToolCall *ToolCall, newArgs json.RawMessage) *ToolCall {
	return &ToolCall{
		ID:   originalToolCall.ID,
		Name: originalToolCall.Name,
		Args: newArgs,
	}
}

func prettyJSON(data json.RawMessage) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "  ")
	if err != nil {
		return string(data)
	}
	return prettyJSON.String()
}
