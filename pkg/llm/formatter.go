package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

// Formatter implements the LLM interface to provide structured output formatting
// It ignores tool call directives and forces the use of its own formatting tool
type Formatter interface {
	Tool
	LLM
	FormatName() string
	ValidateInput(input json.RawMessage) error
}

// BaseFormatter provides a basic implementation of the Formatter interface
type BaseFormatter struct {
	name        string
	description string
	inputSchema json.RawMessage
	llm         LLM // The underlying LLM to delegate to
}

// NewBaseFormatter creates a new base formatter
func NewBaseFormatter(name, description string, inputSchema json.RawMessage, llm LLM) *BaseFormatter {
	return &BaseFormatter{
		name:        name,
		description: description,
		inputSchema: inputSchema,
		llm:         llm,
	}
}

// Name returns the name of the formatter
func (f *BaseFormatter) Name() string {
	return f.name
}

// FormatName returns the format name (same as Name for base formatter)
func (f *BaseFormatter) FormatName() string {
	return f.name
}

// Description returns the description of the formatter
func (f *BaseFormatter) Description() string {
	return f.description
}

// InputSchemaRaw returns the input schema as raw JSON
func (f *BaseFormatter) InputSchemaRaw() json.RawMessage {
	return f.inputSchema
}

// OutputSchemaRaw returns the output schema as raw JSON (same as input for formatters)
func (f *BaseFormatter) OutputSchemaRaw() json.RawMessage {
	return f.inputSchema
}

// ValidateInput validates the input against the schema
func (f *BaseFormatter) ValidateInput(input json.RawMessage) error {
	// For now, just return nil - could add JSON schema validation here
	return nil
}

// Run executes the formatter tool, returning the input as output (echo behavior)
func (f *BaseFormatter) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// Validate input first
	if err := f.ValidateInput(args); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// For formatters, we return the input as output to enforce structure
	return args, nil
}

// Invoke implements the LLM interface
// It ignores tool call directives and forces the use of this formatter
func (f *BaseFormatter) Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	if f.llm == nil {
		return nil, fmt.Errorf("no underlying LLM configured")
	}

	// Create a new request that forces the use of this formatter
	// We ignore any existing tool usage and tool configurations
	forcedRequest := NewLLMRequest(
		request.History,
		WithTools(f),                       // Only include this formatter as a tool
		WithToolUsage(ForceTool(f.Name())), // Force the use of this formatter
	)

	// Delegate to the underlying LLM
	response, err := f.llm.Invoke(ctx, forcedRequest)
	if err != nil {
		return nil, fmt.Errorf("underlying LLM invocation failed: %w", err)
	}

	// Execute the formatter tool with the tool call arguments
	if len(response.ToolCalls) > 0 {
		toolCall := response.ToolCalls[0]
		result, err := f.Run(ctx, toolCall.Args)
		if err != nil {
			return nil, fmt.Errorf("formatter tool execution failed: %w", err)
		}

		content, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		// Create a new response with the formatted result
		return &LLMResponse{
			Messages: []Message{
				&UserMessage{
					Content: string(content),
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("no tool call found in response - LLM did not follow forced tool usage")
}
