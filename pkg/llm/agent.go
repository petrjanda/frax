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
	var lastErr error
	delay := a.retryDelay
	currentToolCall := toolCall // Create a local copy

	// Find the tool to get its input schema
	var targetTool Tool
	for _, t := range a.tools {
		if t.Name() == currentToolCall.Name {
			targetTool = t
			break
		}
	}

	if targetTool == nil {
		return nil, fmt.Errorf("tool not found: %s", currentToolCall.Name)
	}

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

		result, err := targetTool.Run(ctx, currentToolCall.Args)

		if err == nil {
			return &ToolResultMessage{
				ToolCall: currentToolCall,
				Result:   result,
			}, nil
		}

		lastErr = err

		// If this is the last attempt, don't retry
		if attempt == a.maxRetries {
			break
		}

		// Log the retry attempt for debugging
		slog.Info("Tool call failed, asking LLM to correct parameters",
			"tool", currentToolCall.Name,
			"attempt", attempt+1,
			"max_attempts", a.maxRetries,
			"error", err.Error(),
		)

		errorMessage := NewUserMessage(fmt.Sprintf(
			"Tool call to '%s' failed with error: %s\n\n"+
				"Failed parameters: %s\n\n"+
				"Please provide corrected parameters that match the tool's input schema above.",
			currentToolCall.Name,
			err.Error(),
			prettyJSON(currentToolCall.Args),
		))

		retryRequest := NewLLMRequest(NewHistory(errorMessage))
		formatter := NewBaseLLMWithStructuredOutput(targetTool.InputSchemaRaw(), a.llm)

		// Get corrected parameters from the LLM
		retryResponse, retryErr := formatter.Invoke(ctx, retryRequest)
		if retryErr != nil {
			// If we can't get a retry, continue to the next retry attempt
			slog.Warn("Failed to get corrected parameters from LLM, continuing to next retry attempt",
				"tool", currentToolCall.Name,
				"attempt", attempt+1,
				"error", retryErr.Error(),
			)
			continue
		}

		// Extract the corrected parameters from the LLM response
		if len(retryResponse.Messages) > 0 {
			// Get the first message content as corrected parameters
			if userMessage, ok := retryResponse.Messages[0].(*UserMessage); ok {
				correctedArgs := []byte(userMessage.Content)

				slog.Info("Retry response", "response", string(correctedArgs))

				// Create a new tool call with corrected parameters
				currentToolCall = &ToolCall{
					ID:   currentToolCall.ID,
					Name: currentToolCall.Name,
					Args: correctedArgs,
				}

				slog.Info("LLM provided corrected tool call parameters",
					"tool", currentToolCall.Name,
					"corrected_params", prettyJSON(correctedArgs),
				)
			} else {
				slog.Warn("LLM did not provide corrected parameters, continuing to next retry attempt",
					"tool", currentToolCall.Name,
					"attempt", attempt+1,
				)
			}
		} else {
			slog.Warn("LLM did not provide corrected parameters, continuing to next retry attempt",
				"tool", currentToolCall.Name,
				"attempt", attempt+1,
			)
		}
	}

	return nil, fmt.Errorf("tool call failed after %d retries: %w", a.maxRetries+1, lastErr)
}

func prettyJSON(data json.RawMessage) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "  ")
	if err != nil {
		return string(data)
	}
	return prettyJSON.String()
}
