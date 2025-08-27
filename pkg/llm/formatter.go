package llm

import (
	"context"
	"encoding/json"
)

// Formatter represents a tool that enforces a specific output structure
// The model is forced to call this tool, which validates and returns the structured data
type Formatter interface {
	Tool
	// FormatName returns the name used to identify this formatter
	FormatName() string
	// ValidateInput validates that the input matches the expected structure
	ValidateInput(input json.RawMessage) error
}

// BaseFormatter provides a basic implementation of the Formatter interface
type BaseFormatter struct {
	name        string
	description string
	inputSchema json.RawMessage
}

// NewBaseFormatter creates a new base formatter
func NewBaseFormatter(name, description string, inputSchema json.RawMessage) *BaseFormatter {
	return &BaseFormatter{
		name:        name,
		description: description,
		inputSchema: inputSchema,
	}
}

// Name returns the name of the formatter
func (f *BaseFormatter) Name() string {
	return f.name
}

// FormatName returns the format name (same as tool name for base formatter)
func (f *BaseFormatter) FormatName() string {
	return f.name
}

// Description returns the description of the formatter
func (f *BaseFormatter) Description() string {
	return f.description
}

// InputSchemaRaw returns the JSON schema for the formatter's input
func (f *BaseFormatter) InputSchemaRaw() json.RawMessage {
	return f.inputSchema
}

// OutputSchemaRaw returns the JSON schema for the formatter's output
// For formatters, output is typically the same as input (echo behavior)
func (f *BaseFormatter) OutputSchemaRaw() json.RawMessage {
	return f.inputSchema
}

// ValidateInput validates that the input matches the expected structure
func (f *BaseFormatter) ValidateInput(input json.RawMessage) error {
	// Basic validation: ensure input is valid JSON
	var parsed interface{}
	return json.Unmarshal(input, &parsed)
}

// Run executes the formatter, typically echoing back the input
func (f *BaseFormatter) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// Validate the input first
	if err := f.ValidateInput(args); err != nil {
		return nil, err
	}

	// Echo back the input (this enforces the structure)
	return args, nil
}
