package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

type Toolbox = []Tool

func FindTool(name string, tools Toolbox) (Tool, error) {
	for _, tool := range tools {
		if tool.Name() == name {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}

func NewToolbox(tools ...Tool) Toolbox {
	return tools
}

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
