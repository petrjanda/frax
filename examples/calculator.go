package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/llm"
)

// MathCalculator implements the Tool interface for basic math operations
type MathCalculator struct{}

func (c *MathCalculator) Name() string {
	return "calculator"
}

func (c *MathCalculator) Description() string {
	return "A simple calculator that can perform basic arithmetic operations (add, subtract, multiply, divide). Use this tool when you need to perform mathematical calculations."
}

func (c *MathCalculator) InputSchemaRaw() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"operation": {
				"type": "string",
				"enum": ["add", "subtract", "multiply", "divide"],
				"description": "The arithmetic operation to perform"
			},
			"a": {
				"type": "number",
				"description": "First number"
			},
			"b": {
				"type": "number",
				"description": "Second number"
			}
		},
		"required": ["operation", "a", "b"]
	}`)
}

func (c *MathCalculator) OutputSchemaRaw() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"result": {
				"type": "number",
				"description": "The result of the calculation"
			},
			"operation": {
				"type": "string",
				"description": "The operation that was performed"
			}
		}
	}`)
}

func (c *MathCalculator) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Operation string  `json:"operation"`
		A         float64 `json:"a"`
		B         float64 `json:"b"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	var result float64
	switch input.Operation {
	case "add":
		result = input.A + input.B
	case "subtract":
		result = input.A - input.B
	case "multiply":
		result = input.A * input.B
	case "divide":
		if input.B == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = input.A / input.B
	default:
		return nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	output := map[string]interface{}{
		"result":    result,
		"operation": input.Operation,
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return outputBytes, nil
}

func main() {
	fmt.Println("Calculator Tool Example with LLM Integration")
	fmt.Println("==========================================")

	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set your OpenAI API key to run this example:")
		fmt.Println("export OPENAI_API_KEY=your_api_key_here")
		fmt.Println("\nFalling back to demonstration mode...")
		demoMode()
		return
	}

	// Create OpenAI adapter
	fmt.Println("🔑 Creating OpenAI adapter...")
	openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel("gpt-4o"))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}
	fmt.Println("✅ OpenAI adapter created successfully!")

	// Create calculator tool
	calculator := &MathCalculator{}
	fmt.Printf("🧮 Calculator tool loaded: %s\n", calculator.Description())

	// Test various math operations through OpenAI
	testCases := []struct {
		question string
		expected string
	}{
		{
			question: "What is 15 plus 27?",
			expected: "add",
		},
		{
			question: "Can you multiply 8 by 9?",
			expected: "multiply",
		},
		{
			question: "Calculate 100 divided by 5",
			expected: "divide",
		},
		{
			question: "What's 50 minus 23?",
			expected: "subtract",
		},
	}

	ctx := context.Background()

	for i, testCase := range testCases {
		fmt.Printf("\n--- Test Case %d: %s ---\n", i+1, testCase.question)

		// Create conversation history
		history := []llm.Message{
			&llm.UserMessage{Content: testCase.question},
		}

		// Create request with tools enabled
		request := llm.NewLLMRequest(
			history,
			llm.WithToolUsage(true),
			llm.WithTools(calculator),
		)

		// Make the request
		fmt.Printf("🤖 Asking OpenAI: %s\n", testCase.question)
		response, err := openaiLLM.Invoke(ctx, request)
		if err != nil {
			log.Printf("❌ OpenAI request failed: %v", err)
			continue
		}

		// Process the response
		fmt.Printf("📊 Response received:\n")
		fmt.Printf("  Tool calls: %d\n", len(response.ToolCalls))
		fmt.Printf("  Messages: %d\n", len(response.Messages))

		// Check if tools were used
		if response.CalledTool() {
			fmt.Println("  ✅ Tools were used!")
			for j, toolCall := range response.ToolCalls {
				fmt.Printf("    Tool call %d: %s\n", j+1, toolCall.Name)
				fmt.Printf("    Arguments: %s\n", string(toolCall.Args))

				// Execute the tool and show the result
				if toolCall.Name == "calculator" {
					fmt.Printf("    🧮 Executing calculator tool...\n")
					result, err := calculator.Run(ctx, toolCall.Args)
					if err != nil {
						fmt.Printf("    ❌ Tool execution failed: %v\n", err)
					} else {
						// Pretty print the JSON result
						var prettyJSON bytes.Buffer
						if err := json.Indent(&prettyJSON, result, "      ", "  "); err == nil {
							fmt.Printf("    ✅ Tool result:\n%s\n", prettyJSON.String())
						} else {
							fmt.Printf("    ✅ Tool result: %s\n", string(result))
						}
					}
				}
			}
		} else {
			fmt.Println("  ⚠️  No tools were used")
		}

		// Display messages
		for j, msg := range response.Messages {
			switch m := msg.(type) {
			case *llm.UserMessage:
				fmt.Printf("  Message %d (User): %s\n", j+1, m.Content)
			case *llm.ToolResultMessage:
				fmt.Printf("  Message %d (Tool Result): %s\n", j+1, string(m.Result))
			default:
				fmt.Printf("  Message %d (Unknown): %+v\n", j+1, m)
			}
		}
	}

	fmt.Println("\n🎉 Calculator Tool Integration Example completed!")
	fmt.Println("This demonstrates:")
	fmt.Println("- Real AI-powered tool selection")
	fmt.Println("- Natural language processing of math questions")
	fmt.Println("- Tool execution through OpenAI's function calling")
	fmt.Println("- Pretty JSON output display")
}

// demoMode runs without OpenAI API key for demonstration purposes
func demoMode() {
	fmt.Println("\n📚 Demonstration Mode")
	fmt.Println("====================")

	calculator := &MathCalculator{}
	fmt.Printf("🧮 Calculator tool: %s\n", calculator.Description())

	// Show what the tool can do
	fmt.Println("\nThe calculator tool can perform these operations:")
	fmt.Println("- add: Addition of two numbers")
	fmt.Println("- subtract: Subtraction of two numbers")
	fmt.Println("- multiply: Multiplication of two numbers")
	fmt.Println("- divide: Division of two numbers")

	fmt.Println("\nExample usage:")
	fmt.Println("User: 'What is 15 plus 27?'")
	fmt.Println("AI: [Recognizes math operation, calls calculator tool]")
	fmt.Println("Tool: Executes add(15, 27) = 42")
	fmt.Println("AI: 'The result is 42'")

	fmt.Println("\nTo run with real AI:")
	fmt.Println("1. Get an OpenAI API key from https://platform.openai.com/")
	fmt.Println("2. Set the environment variable: export OPENAI_API_KEY=your_key")
	fmt.Println("3. Run this example again")
}
