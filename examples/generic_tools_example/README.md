# Generic Tools Example

This example demonstrates the new generic tool helper that provides type safety and abstracts away JSON marshalling/unmarshalling.

## What is the Generic Tool Helper?

The generic tool helper (`GenericTool[I, O]`) is a type-safe wrapper around the traditional `Tool` interface that:

- **Enforces type safety** at compile time for input (I) and output (O) types
- **Automatically handles JSON marshalling/unmarshalling** 
- **Generates JSON schemas automatically** from Go structs
- **Provides a cleaner API** for creating tools

## Before vs After

### Before (Traditional Approach)
```go
type FlightBookingTool struct{}

func (f *FlightBookingTool) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
    var request FlightBookingRequest
    if err := json.Unmarshal(args, &request); err != nil {
        return nil, fmt.Errorf("failed to parse request parameters: %w", err)
    }
    
    // ... business logic ...
    
    result, err := json.Marshal(flight)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal flight: %w", err)
    }
    
    return json.RawMessage(result), nil
}
```

### After (Generic Tool Helper)
```go
func bookFlight(ctx context.Context, request FlightBookingRequest) (Flight, error) {
    // ... business logic ...
    return flight, nil
}

flightTool := llm.CreateTool[FlightBookingRequest, Flight](
    "book_flight",
    "Book a flight from one city to another...",
    bookFlight,
)
```

## Key Benefits

1. **Type Safety**: Input and output types are enforced at compile time
2. **No Manual JSON Handling**: Marshalling/unmarshalling is automatic
3. **Cleaner Code**: Focus on business logic, not serialization
4. **Automatic Schema Generation**: JSON schemas are generated from Go structs
5. **Better IDE Support**: Full autocomplete and type checking
6. **Easier Refactoring**: Type changes are caught at compile time

## Usage Examples

### Simple Tool Creation
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

### Using the Builder Pattern
```go
tool, err := llm.NewGenericToolBuilder[CalculatorInput, CalculatorOutput]().
    WithName("add_numbers").
    WithDescription("Add two numbers together").
    WithRunner(addNumbers).
    Build()
```

### Complex Types
```go
type WeatherInput struct {
    City string    `json:"city" jsonschema:"required"`
    Date time.Time `json:"date" jsonschema:"required"`
}

type WeatherOutput struct {
    City        string  `json:"city" jsonschema:"required"`
    Temperature float64 `json:"temperature" jsonschema:"required"`
    Condition   string  `json:"condition" jsonschema:"required"`
}

weatherTool := llm.CreateTool[WeatherInput, WeatherOutput](
    "get_weather",
    "Get weather information for a city",
    getWeather,
)
```

## Running the Example

1. Set your OpenAI API key:
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

2. Run the example:
   ```bash
   cd examples/generic_tools_example
   go run main.go
   ```

## How It Works

1. **Type Parameters**: The generic tool uses Go generics `[I, O]` where:
   - `I` is the input type (must be a struct with JSON tags)
   - `O` is the output type (must be a struct with JSON tags)

2. **Schema Generation**: Uses the existing schema generator to create JSON schemas from Go structs

3. **JSON Handling**: The `Run` method automatically:
   - Unmarshals the input JSON to type `I`
   - Calls your runner function with the typed input
   - Marshals the output to JSON

4. **Interface Compliance**: Implements the `Tool` interface, so it works seamlessly with existing code

## Migration Guide

To migrate existing tools to use the generic helper:

1. **Extract the business logic** into a function with the signature:
   ```go
   func(ctx context.Context, input InputType) (OutputType, error)
   ```

2. **Replace the struct-based tool** with a generic tool:
   ```go
   // Old
   toolbox := llm.NewToolbox(&MyTool{})
   
   // New
   myTool := llm.CreateTool[InputType, OutputType]("name", "description", myFunction)
   toolbox := llm.NewToolbox(myTool)
   ```

3. **Remove manual JSON handling** from your tool implementation

4. **Update any type references** to use the new generic tool

## Best Practices

1. **Use descriptive names** for your input/output types
2. **Add proper JSON schema tags** for validation
3. **Keep runner functions pure** - they should only depend on their input parameters
4. **Use the builder pattern** for complex tool configurations
5. **Leverage Go's type system** for compile-time safety

## Limitations

- Input and output types must be structs (not primitives)
- All fields must have proper JSON tags
- The runner function must have the exact signature: `func(context.Context, InputType) (OutputType, error)` 