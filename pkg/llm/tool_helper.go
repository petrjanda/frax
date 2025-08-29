package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petrjanda/frax/pkg/adapters/openai/schemas"
)

// GenericTool is a generic tool implementation that handles JSON marshalling/unmarshalling
// automatically for input type I and output type O
type GenericTool[I, O any] struct {
	name        string
	description string
	runner      func(ctx context.Context, input I) (O, error)
}

// NewGenericTool creates a new generic tool with the given name, description, and runner function
func NewGenericTool[I, O any](name, description string, runner func(ctx context.Context, input I) (O, error)) *GenericTool[I, O] {
	return &GenericTool[I, O]{
		name:        name,
		description: description,
		runner:      runner,
	}
}

// Name returns the name of the tool
func (g *GenericTool[I, O]) Name() string {
	return g.name
}

// Description returns the description of the tool
func (g *GenericTool[I, O]) Description() string {
	return g.description
}

// InputSchemaRaw returns the JSON schema for the tool's input type I
func (g *GenericTool[I, O]) InputSchemaRaw() json.RawMessage {
	generator := schemas.NewOpenAISchemaGenerator()
	return generator.MustGenerateSchema((*I)(nil))
}

// OutputSchemaRaw returns the JSON schema for the tool's output type O
func (g *GenericTool[I, O]) OutputSchemaRaw() json.RawMessage {
	generator := schemas.NewOpenAISchemaGenerator()
	return generator.MustGenerateSchema((*O)(nil))
}

// Run executes the tool with the given arguments, automatically handling JSON marshalling/unmarshalling
func (g *GenericTool[I, O]) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// Unmarshal the input arguments to type I
	var input I
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Run the tool with the typed input
	output, err := g.runner(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Marshal the output to JSON
	result, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return json.RawMessage(result), nil
}

// GenericToolBuilder provides a fluent interface for building generic tools
type GenericToolBuilder[I, O any] struct {
	name        string
	description string
	runner      func(ctx context.Context, input I) (O, error)
}

// NewGenericToolBuilder starts building a new generic tool
func NewGenericToolBuilder[I, O any]() *GenericToolBuilder[I, O] {
	return &GenericToolBuilder[I, O]{}
}

// WithName sets the name of the tool
func (b *GenericToolBuilder[I, O]) WithName(name string) *GenericToolBuilder[I, O] {
	b.name = name
	return b
}

// WithDescription sets the description of the tool
func (b *GenericToolBuilder[I, O]) WithDescription(description string) *GenericToolBuilder[I, O] {
	b.description = description
	return b
}

// WithRunner sets the runner function for the tool
func (g *GenericToolBuilder[I, O]) WithRunner(runner func(ctx context.Context, input I) (O, error)) *GenericToolBuilder[I, O] {
	g.runner = runner
	return g
}

// Build creates the final GenericTool instance
func (b *GenericToolBuilder[I, O]) Build() (*GenericTool[I, O], error) {
	if b.name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if b.description == "" {
		return nil, fmt.Errorf("tool description is required")
	}
	if b.runner == nil {
		return nil, fmt.Errorf("tool runner function is required")
	}

	return &GenericTool[I, O]{
		name:        b.name,
		description: b.description,
		runner:      b.runner,
	}, nil
}

// MustBuild creates the final GenericTool instance or panics if validation fails
func (b *GenericToolBuilder[I, O]) MustBuild() *GenericTool[I, O] {
	tool, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build tool: %v", err))
	}
	return tool
}

// Helper function to create a tool with a simple function signature
func CreateTool[I, O any](name, description string, runner func(ctx context.Context, input I) (O, error)) *GenericTool[I, O] {
	return NewGenericTool(name, description, runner)
}
