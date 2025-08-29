# Generic Tool Helper

The Generic Tool Helper provides a type-safe, generic wrapper around the traditional `Tool` interface that abstracts away JSON marshalling/unmarshalling and provides compile-time type safety.

## Overview

The `GenericTool[I, O]` type allows you to create tools with strongly-typed input and output types, eliminating the need for manual JSON handling and providing better developer experience.

## Key Features

- **Type Safety**: Input (I) and output (O) types are enforced at compile time
- **Automatic JSON Handling**: Marshalling/unmarshalling is handled automatically
- **Schema Generation**: JSON schemas are generated automatically from Go structs
- **Interface Compliance**: Implements the `Tool` interface seamlessly
- **Builder Pattern**: Optional fluent builder interface for complex configurations

## Architecture

```
GenericTool[I, O] implements Tool interface
├── Name() string
├── Description() string  
├── InputSchemaRaw() json.RawMessage
├── OutputSchemaRaw() json.RawMessage
└── Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
```

The `Run` method automatically:
1. Unmarshals input JSON to type `I`
2. Calls your runner function with the typed input
3. Marshals the output to JSON

## Usage

### Basic Usage

```go
// Define input/output types
type CalculatorInput struct {
    A int `json:"a" jsonschema:"required"`
    B int `json:"b" jsonschema:"required"`
}

type CalculatorOutput struct {
    Result int `json:"result" jsonschema:"required"`
}

// Define the runner function
func addNumbers(ctx context.Context, input CalculatorInput) (CalculatorOutput, error) {
    return CalculatorOutput{Result: input.A + input.B}, nil
}

// Create the tool
calculatorTool := llm.CreateTool[CalculatorInput, CalculatorOutput](
    "add_numbers",
    "Add two numbers together",
    addNumbers,
)
```

### Constructor Function

```go
tool := llm.NewGenericTool[InputType, OutputType](
    "tool_name",
    "tool description", 
    runnerFunction,
)
```

### Builder Pattern

```go
tool, err := llm.NewGenericToolBuilder[InputType, OutputType]().
    WithName("tool_name").
    WithDescription("tool description").
    WithRunner(runnerFunction).
    Build()
```

## Type Requirements

### Input/Output Types Must Be Structs

```go
// ✅ Valid - struct types
type Input struct {
    Field string `json:"field" jsonschema:"required"`
}

type Output struct {
    Result string `json:"result" jsonschema:"required"`
}

// ❌ Invalid - primitive types
type Input string
type Output int
```

### JSON Schema Tags

All fields must have proper JSON tags for schema generation:

```go
type ValidInput struct {
    Name     string  `json:"name" jsonschema:"required"`
    Age      int     `json:"age" jsonschema:"required"`
    Optional *string `json:"optional,omitempty"`
}
```

### Runner Function Signature

The runner function must have the exact signature:

```go
func(ctx context.Context, input InputType) (OutputType, error)
```

## Integration with Existing Code

Generic tools are fully compatible with existing code:

```go
// Create toolbox with generic tools
toolbox := llm.NewToolbox(
    calculatorTool,
    weatherTool,
    // ... other tools
)

// Use with existing agent
agent := llm.NewAgent(llm, toolbox)
```

## Error Handling

The generic tool helper provides meaningful error messages:

```go
// Input unmarshalling errors
"failed to unmarshal input: invalid character 'x' looking for beginning of value"

// Tool execution errors  
"tool execution failed: division by zero"

// Output marshalling errors
"failed to marshal output: unsupported type"
```

## Testing

Generic tools are easy to test since they're just functions:

```go
func TestCalculatorTool(t *testing.T) {
    ctx := context.Background()
    input := CalculatorInput{A: 5, B: 3}
    
    output, err := addNumbers(ctx, input)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    if output.Result != 8 {
        t.Errorf("Expected result 8, got %d", output.Result)
    }
}
```

## Performance

- **No runtime overhead**: Generic tools have the same performance as manual implementations
- **Compile-time optimization**: Go compiler optimizes the generic code
- **Memory efficient**: No additional allocations beyond what's necessary

## Migration from Traditional Tools

### Before (Traditional)

```go
type MyTool struct{}

func (t *MyTool) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
    var input MyInput
    if err := json.Unmarshal(args, &input); err != nil {
        return nil, err
    }
    
    // ... business logic ...
    
    result, err := json.Marshal(output)
    if err != nil {
        return nil, err
    }
    
    return json.RawMessage(result), nil
}
```

### After (Generic)

```go
func myFunction(ctx context.Context, input MyInput) (MyOutput, error) {
    // ... business logic ...
    return output, nil
}

myTool := llm.CreateTool[MyInput, MyOutput]("name", "description", myFunction)
```

## Best Practices

1. **Use descriptive type names**: `FlightBookingRequest` instead of `Input`
2. **Keep runner functions pure**: Only depend on input parameters
3. **Add comprehensive JSON schema tags**: Helps with validation
4. **Use the builder pattern**: For complex tool configurations
5. **Test the runner functions**: Separate from tool infrastructure

## Limitations

- Input/output types must be structs (not primitives)
- All fields must have JSON tags
- Runner function signature is fixed
- Requires Go 1.18+ for generics support

## Examples

See the `examples/generic_tools_example/` directory for complete working examples.

## Contributing

When adding new features to the generic tool helper:

1. Maintain backward compatibility
2. Add comprehensive tests
3. Update documentation
4. Follow Go best practices
5. Consider performance implications 