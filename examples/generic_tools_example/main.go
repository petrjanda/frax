package main

import (
	"context"
	"log"
	"os"

	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/llm"
)

// Simple example types for demonstration
type CalculatorInput struct {
	A int `json:"a" jsonschema:"required"`
	B int `json:"b" jsonschema:"required"`
}

type CalculatorOutput struct {
	Result int    `json:"result" jsonschema:"required"`
	Op     string `json:"operation" jsonschema:"required"`
}

type WeatherInput struct {
	City string `json:"city" jsonschema:"required"`
	Date string `json:"date" jsonschema:"required"`
}

type WeatherOutput struct {
	City        string  `json:"city" jsonschema:"required"`
	Date        string  `json:"date" jsonschema:"required"`
	Temperature float64 `json:"temperature" jsonschema:"required"`
	Condition   string  `json:"condition" jsonschema:"required"`
}

// Calculator tool function - clean and typed!
func addNumbers(ctx context.Context, input CalculatorInput) (CalculatorOutput, error) {
	return CalculatorOutput{
		Result: input.A + input.B,
		Op:     "addition",
	}, nil
}

// Weather tool function - clean and typed!
func getWeather(ctx context.Context, input WeatherInput) (WeatherOutput, error) {
	// Mock weather data
	return WeatherOutput{
		City:        input.City,
		Date:        input.Date,
		Temperature: 22.5,
		Condition:   "sunny",
	}, nil
}

func main() {
	ctx := context.Background()

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI adapter
	openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// Create tools using the generic helper - much cleaner!
	calculatorTool := llm.CreateTool(
		"add_numbers",
		"Add two numbers together. Provide two integers to add.",
		addNumbers,
	)

	weatherTool := llm.CreateTool(
		"get_weather",
		"Get weather information for a city on a specific date. Provide city name and date.",
		getWeather,
	)

	// Create toolbox
	toolbox := llm.NewToolbox(calculatorTool, weatherTool)

	// Create agent
	agent := llm.NewAgent(openaiLLM, toolbox)

	// Create conversation history
	history := llm.NewHistory(
		llm.NewUserMessage("Please add 15 and 27, and also tell me the weather in Barcelona for tomorrow."),
	)

	log.Println("🚀 Starting Generic Tools Example...")
	log.Println("🛠️  Tools available: add_numbers, get_weather")
	log.Println("============================================================")

	// Run the agent
	response, err := agent.Invoke(ctx, llm.NewLLMRequest(history,
		llm.WithSystem("You are a helpful assistant with access to calculator and weather tools."),
		llm.WithTemperature(0.0),
	))
	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	log.Println("\n============================================================")
	log.Println("🎉 Agent completed the task!")
	log.Println("📝 Conversation:")
	log.Println("============================================================")

	// Print the conversation
	for i, msg := range response.Messages {
		switch m := msg.(type) {
		case *llm.UserMessage:
			log.Printf("\n👤 User %d: %s\n", i+1, m.Content)
		case *llm.ToolResultMessage:
			log.Printf("\n🔧 Tool Result %d: %s\n", i+1, string(m.Result))
		case *llm.AssistantMessage:
			log.Printf("\n🤖 Assistant %d: %s\n", i+1, m.Content)
		default:
			log.Printf("\n💬 Message %d: %+v\n", i+1, msg)
		}
	}

}
