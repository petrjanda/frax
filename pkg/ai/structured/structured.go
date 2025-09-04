package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/tools"

	"github.com/xeipuuv/gojsonschema"
)

// StructuredLLM implements the StructuredLLM interface to provide structured output formatting
// It ignores tool call directives and forces the use of its own formatting tool
type StructuredLLM interface {
	tools.Tool
	ai.LLM
	FormatName() string
	ValidateInput(input json.RawMessage) error
}

// LLM provides a basic implementation of the LLMWithStructuredOutput interface
type LLM struct {
	name        string
	description string
	inputSchema json.RawMessage
	llm         ai.LLM // The underlying LLM to delegate to

	events ai.LLMEvents
}

// LLMOpts represents options for configuring an LLM with structured output
type LLMOpts = func(*LLM)

// WithName sets a custom name for the tool (defaults to "formatter")
func WithName(name string) LLMOpts {
	return func(f *LLM) {
		f.name = name
	}
}

// WithDescription sets a custom description for the tool
func WithDescription(description string) LLMOpts {
	return func(f *LLM) {
		f.description = description
	}
}

func WithEvents(events ai.LLMEvents) LLMOpts {
	return func(f *LLM) {
		f.events = events
	}
}

// NewLLM creates a new base LLM with structured output
// Uses sensible defaults: name="formatter", description="Must be called to provide structured output"
func NewLLM(inputSchema json.RawMessage, llm_ ai.LLM, opts ...LLMOpts) StructuredLLM {
	f := &LLM{
		name:        "formatter",
		description: "Must be called to provide structured output",
		inputSchema: inputSchema,
		llm:         llm_,
		events:      ai.NewNoopAgentEvents(),
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Name returns the name of the LLM with structured output
func (f *LLM) Name() string {
	return f.name
}

// FormatName returns the format name (same as Name for base LLM with structured output)
func (f *LLM) FormatName() string {
	return f.name
}

// Description returns the description of the LLM with structured output
func (f *LLM) Description() string {
	return f.description
}

// InputSchemaRaw returns the input schema as raw JSON
func (f *LLM) InputSchemaRaw() json.RawMessage {
	return f.inputSchema
}

// ValidateInput validates the input against the schema
func (f *LLM) ValidateInput(input json.RawMessage) error {
	// @todo: add json schema validation here

	// Create a schema loader and compile the schema
	schemaLoader := gojsonschema.NewBytesLoader(f.inputSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	// Create a document loader for the input
	documentLoader := gojsonschema.NewBytesLoader(input)

	// Validate the input against the schema
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid() {
		// Collect all validation errors
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		return fmt.Errorf("validation errors: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

// Execute executes the LLM with structured output tool, returning the input as output (echo behavior)
func (f *LLM) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {

	// prettyArgs, err := json.MarshalIndent(args, "", "  ")
	// if err != nil {
	// 	panic(fmt.Sprintf("failed to marshal args: %v", err))
	// }
	// log.Println(string(prettyArgs))
	// panic(f.ValidateInput(args))

	// Validate input first
	if err := f.ValidateInput(args); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// For LLMs with structured output, we return the input as output to enforce structure
	return args, nil
}

// Invoke implements the LLM interface
// It ignores tool call directives and forces the use of this LLM with structured output
func (f *LLM) Invoke(ctx context.Context, request *ai.LLMRequest) (*ai.LLMResponse, error) {
	if f.llm == nil {
		return nil, fmt.Errorf("no underlying LLM configured")
	}

	// Create a new request that forces the use of this LLM with structured output
	// We ignore any existing tool usage and tool configurations
	forcedRequest := request.Clone(
		ai.WithTools(f), // Only include this LLM with structured output as a tool
		ai.WithToolUsage(tools.ForceTool(f.Name())), // Force the use of this LLM with structured output
	)

	// Log actual internal request
	f.events.OnRequest(ctx, forcedRequest)

	// Delegate to the underlying LLM
	response, err := f.llm.Invoke(ctx, forcedRequest)
	if err != nil {
		return nil, fmt.Errorf("underlying LLM invocation failed: %w", err)
	}

	// Execute the LLM with structured output tool with the tool call arguments
	toolCalls := response.ToolCalls()
	if len(toolCalls) > 0 {
		toolCall := toolCalls[0]
		result, err := f.Execute(ctx, toolCall.Args)
		if err != nil {
			f.events.OnToolError(ctx, toolCall, 0, err)
			return nil, fmt.Errorf("structured output generation failed: %w", err)
		}

		content, err := json.Marshal(result)
		if err != nil {
			f.events.OnToolError(ctx, toolCall, 0, err)
			return nil, fmt.Errorf("structured output marshalling failed: %w", err)
		}

		// Reset response so we don't expose internal `formatter` tool call details.
		// response := NewLLMResponse()

		response := response.Clone()
		response.AddMessage(ai.NewUserMessage(string(content)))
		f.events.OnResponse(ctx, request, response)

		// Create a new response with the formatted result
		return response, nil
	}

	return nil, fmt.Errorf("no tool call found in response - LLM did not follow forced tool usage")
}

// MarshalJSON implements custom JSON marshaling for BaseLLMWithStructuredOutput
// This ensures that when the struct is serialized to JSON, it reports the tool name
func (f *LLM) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":         f.name,
		"description":  f.description,
		"input_schema": f.inputSchema,
	})
}
