package llm

import (
	"context"
	"fmt"
	"time"
)

// Agent represents an agent that can use tools and interact with an LLM
type Agent struct {
	llm          LLM
	tools        []Tool
	history      []Message
	maxRetries   int
	retryDelay   time.Duration
	retryBackoff float64
}

// AgentOpts represents options for configuring an agent
type AgentOpts = func(*Agent)

// WithInitialHistory sets the initial conversation history for the agent
func WithInitialHistory(history []Message) AgentOpts {
	return func(a *Agent) {
		a.history = history
	}
}

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
		for _, toolCall := range response.ToolCalls {
			message, err := a.CallToolWithRetry(ctx, toolCall)
			if err != nil {
				// If all retries failed, create an error message and continue
				errorMessage := &ToolErrorMessage{
					ToolCall: toolCall,
					Error:    err.Error(),
				}
				response.AddMessage(errorMessage)
			} else {
				response.AddMessage(message)
			}
		}

		// Continue the loop
		return a.Invoke(ctx, NewLLMRequest(response.Messages))
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

		// Create an error message to feed back to the LLM for correction
		errorMessage := &ToolErrorMessage{
			ToolCall: currentToolCall,
			Error:    err.Error(),
		}

		// Add the error message to the conversation and ask LLM to retry
		// This allows the LLM to understand what went wrong and correct the input
		retryRequest := NewLLMRequest(NewHistory(errorMessage))
		retryRequest = retryRequest.Clone(
			WithTools(a.tools...),
			WithToolUsage(AutoToolSelection()),
		)

		// Get a corrected tool call from the LLM
		retryResponse, retryErr := a.llm.Invoke(ctx, retryRequest)
		if retryErr != nil {
			// If we can't get a retry, continue to the next retry attempt
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
			}
		}
	}

	return nil, fmt.Errorf("tool call failed after %d retries: %w", a.maxRetries+1, lastErr)
}

// CallTool executes a tool call and returns the result message
func (a *Agent) CallTool(ctx context.Context, toolCall *ToolCall) (Message, error) {
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
