package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormatterRequest represents a request to the formatter with just conversation history
type FormatterRequest struct {
	History []Message
}

// NewFormatterRequest creates a new formatter request with the given history
func NewFormatterRequest(history []Message) *FormatterRequest {
	return &FormatterRequest{History: history}
}

// Invoker uses a formatter to force the model to produce structured output
// It handles all logic directly without maintaining history or complex tool call checking
type Invoker struct {
	llm       LLM
	formatter Formatter
}

// InvokerOpts represents options for configuring an invoker
type InvokerOpts = func(*Invoker)

// WithFormatter sets the formatter for the invoker
func WithFormatter(formatter Formatter) InvokerOpts {
	return func(i *Invoker) {
		i.formatter = formatter
	}
}

// NewInvoker creates a new invoker with the given LLM and formatter
func NewInvoker(llm LLM, formatter Formatter, opts ...InvokerOpts) *Invoker {
	i := &Invoker{
		llm:       llm,
		formatter: formatter,
	}

	for _, opt := range opts {
		opt(i)
	}
	return i
}

// Invoke forces the model to produce output in the specified format
// The model will be forced to call the formatter tool, ensuring structured output
func (i *Invoker) Invoke(ctx context.Context, request *FormatterRequest) (json.RawMessage, error) {
	if i.llm == nil {
		return nil, fmt.Errorf("no LLM configured")
	}
	if i.formatter == nil {
		return nil, fmt.Errorf("no formatter configured")
	}

	// Create request that forces the use of the formatter tool
	req := NewLLMRequest(
		request.History,
		WithTools(i.formatter),
		WithToolUsage(ForceTool(i.formatter.Name())),
	)

	// Invoke the LLM directly
	response, err := i.llm.Invoke(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM invocation failed: %w", err)
	}

	// Since we forced the tool usage, we assume the response contains a tool call
	// and we execute it directly to get the structured data
	if len(response.ToolCalls) > 0 {
		toolCall := response.ToolCalls[0] // Take the first tool call

		// Execute the formatter tool with the provided arguments
		result, err := i.formatter.Run(ctx, toolCall.Args)
		if err != nil {
			return nil, fmt.Errorf("formatter tool execution failed: %w", err)
		}

		return result, nil
	}

	return nil, fmt.Errorf("no tool call found in response - LLM did not follow forced tool usage")
}
