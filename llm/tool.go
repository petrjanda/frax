package llm

import (
	"context"
	"encoding/json"
)

// Tool represents a tool that can be called by the agent
type Tool interface {
	// Name returns the name of the tool
	Name() string

	// Description returns a description of what the tool does
	Description() string

	// InputSchemaRaw returns the JSON schema for the tool's input
	InputSchemaRaw() json.RawMessage

	// OutputSchemaRaw returns the JSON schema for the tool's output
	OutputSchemaRaw() json.RawMessage

	// Run executes the tool with the given arguments
	Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}
