package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
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

	if response.CalledTool() {
		// Process all tool calls first to collect results
		var toolResults []Message
		var failedToolCalls []*ToolCall

		for _, toolCall := range response.ToolCalls {
			message, err := a.CallToolWithRetry(ctx, toolCall)
			if err != nil {
				// Collect failed tool calls for later handling
				failedToolCalls = append(failedToolCalls, toolCall)
			} else {
				toolResults = append(toolResults, message)
			}
		}

		// If we have successful tool results, add them to the response
		if len(toolResults) > 0 {
			for _, result := range toolResults {
				response.AddMessage(result)
			}
		}

		// If we have failed tool calls, we need to handle them
		if len(failedToolCalls) > 0 {
			// Create a comprehensive error message for the failed tools
			errorDescription := "Some tools failed to execute. Please review the errors and provide corrected parameters."
			errorMessage := NewUserMessage(errorDescription)
			response.AddMessage(errorMessage)

			// Return the response with errors instead of continuing the loop
			// This prevents the message sequence violation
			return response, nil
		}

		return a.Invoke(ctx, NewLLMRequest(append(request.History, response.Messages...)))
	}

	return response, nil
}

// CallToolWithRetry executes a tool call with retry logic
func (a *Agent) CallToolWithRetry(ctx context.Context, toolCall *ToolCall) (Message, error) {
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

		message, err := a.CallTool(ctx, currentToolCall)
		if err == nil {
			return message, nil
		}

		lastErr = err

		// If this is the last attempt, don't retry
		if attempt == a.maxRetries {
			break
		}

		// Create a comprehensive error message that includes:
		// 1. What the tool call was trying to do
		// 2. What parameters were used
		// 3. What the specific error was
		// 4. Clear instructions on how to fix it
		errorDescription := fmt.Sprintf(
			"Tool call to '%s' failed with error: %s\n\n"+
				"Failed parameters: %s\n\n"+
				"Please correct the parameters and provide a new tool call. "+
				"Make sure to fix the specific issue mentioned in the error.",
			currentToolCall.Name,
			err.Error(),
			prettyJSON(currentToolCall.Args),
		)

		errorMessage := NewUserMessage(errorDescription)

		// Log the retry attempt for debugging
		slog.Info("Tool call failed, asking LLM to correct parameters",
			"tool", currentToolCall.Name,
			"attempt", attempt+1,
			"max_attempts", a.maxRetries,
			"error", err.Error(),
		)

		// Create a retry request that includes:
		// 1. The original conversation history (so the LLM knows the task)
		// 2. The error message with context
		// 3. The tools available
		// This gives the LLM full context to understand what went wrong and how to fix it
		retryRequest := NewLLMRequest(NewHistory(errorMessage))
		retryRequest = retryRequest.Clone(
			WithTools(a.tools...),
			WithToolUsage(AutoToolSelection()),
		)

		// Get a corrected tool call from the LLM
		retryResponse, retryErr := a.llm.Invoke(ctx, retryRequest)
		if retryErr != nil {
			// If we can't get a retry, continue to the next retry attempt
			slog.Warn("Failed to get corrected tool call from LLM, continuing to next retry attempt",
				"tool", currentToolCall.Name,
				"attempt", attempt+1,
				"error", retryErr.Error(),
			)
			continue
		}

		// If the LLM provided a corrected tool call, use it
		if len(retryResponse.ToolCalls) > 0 {
			correctedToolCall := retryResponse.ToolCalls[0]
			if correctedToolCall.Name == currentToolCall.Name {
				// Create a new copy to avoid modifying the original
				currentToolCall = &ToolCall{
					Name: correctedToolCall.Name,
					Args: correctedToolCall.Args,
				}
				slog.Info("LLM provided corrected tool call parameters",
					"tool", correctedToolCall.Name,
					"corrected_params", prettyJSON(correctedToolCall.Args),
				)
			}
		} else {
			slog.Warn("LLM did not provide a corrected tool call, continuing to next retry attempt",
				"tool", currentToolCall.Name,
				"attempt", attempt+1,
			)
		}
	}

	return nil, fmt.Errorf("tool call failed after %d retries: %w", a.maxRetries+1, lastErr)
}

// CallTool executes a tool call and returns the result message
func (a *Agent) CallTool(ctx context.Context, toolCall *ToolCall) (Message, error) {

	slog.Info("Calling tool",
		"tool", toolCall.Name,
		"params", prettyJSON(toolCall.Args))

	var tool Tool
	for _, t := range a.tools {
		if t.Name() == toolCall.Name {
			tool = t
			break
		}
	}

	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Name)
	}

	result, err := tool.Run(ctx, toolCall.Args)
	if err != nil {
		return nil, err
	}

	return &ToolResultMessage{
		ToolCall: toolCall,
		Result:   result,
	}, nil
}

func prettyJSON(data json.RawMessage) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "  ")
	if err != nil {
		return string(data)
	}
	return prettyJSON.String()
}
