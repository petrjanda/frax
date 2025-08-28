# Frax LLM Framework

A modular Go framework for building LLM-powered applications with a clean, extensible architecture.

## 🏗️ Package Structure

The framework is organized into a clean, modular structure:

```
frax/
├── pkg/                    # All packages
│   ├── llm/               # Core LLM package (generic)
│   │   ├── agent.go       # Agent interface and implementation
│   │   ├── llm.go         # LLM interface, requests, responses
│   │   ├── messages.go    # Message interface and implementations
│   │   ├── tool.go        # Tool interface
│   │   └── types.go       # Generic Task and Eval interfaces
│   └── adapters/          # LLM provider adapters
│       └── openai/        # OpenAI API adapter
│           ├── openai.go  # OpenAI-specific implementation
│           └── schemas/   # OpenAI-compatible schema generation
│               ├── openai.go      # Schema generator
│               └── README.md      # Schema package documentation
├── examples/               # Example implementations
│   ├── calculator/        # Calculator tool example
│   ├── structured_output/ # Structured output with schema generation
│   └── travel_agent/      # Travel agent with flight and hotel booking tools
├── go.mod                  # Go module file
└── go.sum                  # Go module checksums
```

## 🚀 Core Components

### 1. **Agent** (`llm/agent.go`)
The `Agent` struct orchestrates conversations between users, LLMs, and tools. It implements the conversation loop, handles tool calls, and manages conversation history.

```go
agent := llm.NewAgent(llm, tools, llm.WithInitialHistory(history))
messages, err := agent.Loop(ctx, conversationHistory)
```

### 2. **LLM** (`llm/llm.go`)
The `LLM` interface defines how to interact with language models. It handles requests, responses, and tool integration.

```go
type LLM interface {
    Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error)
}
```

### 3. **Messages** (`llm/messages.go`)
Message types for different conversation elements:
- `UserMessage`: User input
- `ToolResultMessage`: Results from tool execution
- `ToolCall`: Tool invocation requests

### 4. **Tools** (`llm/tool.go`)
The `Tool` interface defines executable functions that agents can call:

```go
type Tool interface {
    Name() string
    Description() string
    InputSchemaRaw() json.RawMessage
    OutputSchemaRaw() json.RawMessage
    Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}
```

### 5. **Tool Usage Control** (`pkg/llm/tool_usage.go`)
Simple control over tool usage behavior:

```go
// Let the LLM choose automatically (default behavior)
request := llm.NewLLMRequest(history, llm.WithToolUsage(llm.AutoToolSelection()))

// Force use of a specific tool
request := llm.NewLLMRequest(history, llm.WithToolUsage(llm.ForceTool("calculator")))
```

### 6. **OpenAI Adapter** (`pkg/adapters/openai/`)
A complete implementation of the LLM interface using OpenAI's API:

```go
openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel("gpt-4o"))
response, err := openaiLLM.Invoke(ctx, request)
```

### 7. **OpenAI Schemas** (`pkg/adapters/openai/schemas/`)
Generates OpenAI-compatible JSON schemas from Go structs using the [invopop/jsonschema](https://github.com/invopop/jsonschema) library:

```go
import "github.com/petrjanda/frax/pkg/adapters/openai/schemas"

generator := schemas.NewOpenAISchemaGenerator()
schema, err := generator.GenerateSchema(Person{})
```

## 📦 Installation

```bash
go get github.com/petrjanda/frax
```

## 📁 Project Structure

The project follows a clean separation of concerns:

- **`pkg/llm/`**: Contains the core, generic LLM interfaces and implementations that are provider-agnostic
- **`pkg/adapters/`**: Contains specific LLM provider implementations (OpenAI, etc.)
- **`examples/`**: Contains working examples showing how to use the framework

## 🎯 Tool Usage Strategies

The framework provides two simple strategies for controlling tool usage:

1. **`AutoToolSelection()`**: Lets the LLM automatically choose when to use tools (default behavior)
2. **`ForceTool(toolName)`**: Forces the LLM to use a specific tool

For more complex scenarios (like disabling tools or restricting to a subset), simply control which tools are provided to the agent in the first place. This approach is simpler and more intuitive.

## 🔧 Usage Examples

### Basic OpenAI Integration

```go
package main

import (
    "context"
    "github.com/petrjanda/frax/pkg/llm"
    "github.com/petrjanda/frax/pkg/adapters/openai"
)

func main() {
    // Create OpenAI adapter
    openaiLLM, err := openai.NewOpenAIAdapter(os.Getenv("OPENAI_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }

    // Create conversation history
    history := []llm.Message{
        &llm.UserMessage{Content: "Hello! How are you?"},
    }

    // Example: Force the use of a specific tool
    request := llm.NewLLMRequest(
        history,
        llm.WithToolUsage(llm.ForceTool("calculator")),
        llm.WithTools(calculator),
    )

    // Make request
    request := llm.NewLLMRequest(history, llm.WithToolUsage(false))
    response, err := openaiLLM.Invoke(context.Background(), request)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %+v\n", response)
}
```

### Creating Custom Tools

```go
type CalculatorTool struct{}

func (c *CalculatorTool) Name() string { return "calculator" }
func (c *CalculatorTool) Description() string { return "Performs basic math operations" }
func (c *CalculatorTool) InputSchemaRaw() json.RawMessage { 
    return json.RawMessage(`{"type": "object", "properties": {...}}`) 
}
func (c *CalculatorTool) OutputSchemaRaw() json.RawMessage { 
    return json.RawMessage(`{"type": "object", "properties": {...}}`) 
}
func (c *CalculatorTool) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
    // Tool implementation
    return result, nil
}
```

### Using Agents with Tools

```go
// Create tools
tools := []llm.Tool{&CalculatorTool{}}

// Create agent
agent := llm.NewAgent(openaiLLM, tools)

// Run conversation loop
messages, err := agent.Loop(ctx, conversationHistory)
```

### Structured Output with Schema Generation

The framework supports structured output generation with automatic schema creation:

```go
import "github.com/petrjanda/frax/pkg/adapters/openai/schemas"

// Define your struct with jsonschema tags
type Person struct {
    Name string `json:"name" jsonschema:"required"`
    Age  int    `json:"age" jsonschema:"required"`
}

// Generate OpenAI-compatible schema
generator := schemas.NewOpenAISchemaGenerator()
schema, err := generator.GenerateSchema(Person{})

// Use with structured output LLM
personLLM := llm.NewBaseLLMWithStructuredOutput(schema, openaiLLM)
```

### Travel Agent with Multiple Tools

The framework demonstrates complex multi-tool workflows with the travel agent example:

```go
// Create travel booking tools
tools := []llm.Tool{
    &FlightBookingTool{},
    &HotelBookingTool{},
}

// Create agent with tools
agent := llm.NewAgent(openaiLLM, tools)

// The agent automatically orchestrates the booking process
response, err := agent.Invoke(ctx, request)
```

## 🌟 Features

- **Modular Design**: Clean separation of concerns with focused, single-responsibility packages
- **Interface-Based**: Easy to extend and implement new LLM providers
- **Tool Integration**: Built-in support for function calling and tool execution
- **Type Safety**: Strong typing with Go's type system
- **OpenAI Ready**: Full OpenAI API integration out of the box
- **Extensible**: Easy to add new message types, tools, and LLM providers

## 🔌 Extending the Framework

### Adding New LLM Providers

Implement the `LLM` interface:

```go
type MyLLMProvider struct{}

func (m *MyLLMProvider) Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
    // Your implementation here
    return response, nil
}
```

### Adding New Message Types

Implement the `Message` interface:

```go
type CustomMessage struct {
    Content string
    Metadata map[string]interface{}
}

func (m *CustomMessage) Kind() MessageKind {
    return "custom"
}
```

### Adding New Tool Types

Implement the `Tool` interface (see the calculator example in `examples/simple_tool.go`).

## 🧪 Testing

```bash
# Test the main package
go test ./...

# Test the llm package specifically
go test ./llm

# Run examples
go run examples/simple_tool.go
```

## 📝 Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key for the OpenAI adapter

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🚀 Roadmap

- [ ] Streaming support for LLM responses
- [ ] Additional LLM provider adapters (Anthropic, Google, etc.)
- [ ] Enhanced tool validation and error handling
- [ ] Conversation persistence and management
- [ ] Performance monitoring and metrics
- [ ] WebSocket support for real-time conversations 