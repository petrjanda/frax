# Tools Helper Package - Before vs After

This document shows the dramatic reduction in boilerplate code when using the tools helper package.

## Before: Manual Implementation

```go
// Manual tool implementation - lots of boilerplate
type FlightBookingTool struct{}

func (f *FlightBookingTool) Name() string {
    return "book_flight"
}

func (f *FlightBookingTool) Description() string {
    return "Book a flight from one city to another..."
}

func (f *FlightBookingTool) InputSchemaRaw() json.RawMessage {
    generator := schemas.NewOpenAISchemaGenerator()
    return generator.MustGenerateSchema(FlightRequest{})
}

func (f *FlightBookingTool) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
    // Your actual tool logic here
    var request FlightRequest
    if err := json.Unmarshal(args, &request); err != nil {
        return nil, err
    }

    // Process the request...
    flight := Flight{...}
    return json.Marshal(flight)
}
```

**Lines of code: ~25 lines**
**Boilerplate: ~15 lines (60%)**
**Actual logic: ~10 lines (40%)**

## After: Using Tools Helper

```go
// Using tools helper - minimal boilerplate
flightTool := tools.NewSimpleTool(tools.ToolConfig{
    Name:        "book_flight",
    Description: "Book a flight from one city to another...",
    InputType:   FlightRequest{},
    OutputType:  Flight{},
}, func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
    // Your actual tool logic here
    var request FlightRequest
    if err := json.Unmarshal(args, &request); err != nil {
        return nil, err
    }

    // Process the request...
    flight := Flight{...}
    return json.Marshal(flight)
})
```

**Lines of code: ~15 lines**
**Boilerplate: ~5 lines (33%)**
**Actual logic: ~10 lines (67%)**

## Even More Compact: Auto-Generation

```go
// Auto-generated tool - minimal boilerplate
flightTool := tools.AutoToolWithName("book_flight",
    FlightRequest{},
    Flight{},
    func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
        // Your actual tool logic here
        var request FlightRequest
        if err := json.Unmarshal(args, &request); err != nil {
            return nil, err
        }

        // Process the request...
        flight := Flight{...}
        return json.Marshal(flight)
    })
```

**Lines of code: ~12 lines**
**Boilerplate: ~2 lines (17%)**
**Actual logic: ~10 lines (83%)**

## Benefits Summary

| Aspect             | Manual   | Helper   | Auto-Generated |
| ------------------ | -------- | -------- | -------------- |
| **Total Lines**    | 25       | 15       | 12             |
| **Boilerplate**    | 15 (60%) | 5 (33%)  | 2 (17%)        |
| **Actual Logic**   | 10 (40%) | 10 (67%) | 10 (83%)       |
| **Code Reduction** | -        | 40%      | 52%            |
| **Maintenance**    | High     | Low      | Very Low       |

## What's Automatically Generated

✅ **Input Schema**: JSON schema from your Go types  
✅ **Output Schema**: JSON schema from your Go types  
✅ **Name**: Can be auto-generated from type names  
✅ **Description**: Can be auto-generated from type names  
✅ **Interface Implementation**: All required methods

## What You Still Control

🎯 **Tool Logic**: Your `Run` function implementation  
🎯 **Custom Names**: Override auto-generated names if needed  
🎯 **Custom Descriptions**: Override auto-generated descriptions if needed  
🎯 **Type Safety**: Full Go type safety for inputs/outputs

## Migration Path

1. **Start Simple**: Use `NewSimpleTool()` for full control
2. **Add Auto-Generation**: Use `AutoToolWithName()` for custom names
3. **Go Full Auto**: Use `AutoTool()` for maximum automation
4. **Customize as Needed**: Mix and match approaches

The tools helper package transforms tool creation from a boilerplate-heavy task to a focused, logic-first experience!
