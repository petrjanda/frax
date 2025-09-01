package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/adapters/openai/schemas"
	"github.com/petrjanda/frax/pkg/llm"
	"github.com/petrjanda/frax/synq/dsl"
)

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

	structuredLLM := llm.NewBaseLLMWithStructuredOutput(
		schemas.NewOpenAISchemaGenerator().MustGenerateSchema(dsl.Directive{}),
		openaiLLM,
	)

	system := `
		You are expert data engineer who turns plan text requests into directives that can be used to deploy data tests.
		
		Turn user input into structured description of how tests should be deployed with 3 components:
		* Entity selection - which defines which data entities (tables, views, etc.) should be tested
		* Column selection - which defines which columns should be tested
		* Tests - which defines the tests that should be deployed
	`

	// Create conversation history with the travel request
	history := llm.NewHistory(
		llm.NewUserMessage(`
			Deploy tests to default.runs table on all id fields and ensure they are unique.
		`),
	)

	// Run the agent
	response, err := structuredLLM.Invoke(ctx,
		llm.NewLLMRequest(history,
			llm.WithSystem(system),
			llm.WithTemperature(0.0),
			llm.WithMaxCompletionTokens(1000),
		),
	)

	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	// Print the conversation
	for _, msg := range response.Messages {
		payload, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("Failed to marshal message: %v", err)
		}
		fmt.Println(string(payload))
	}
}
