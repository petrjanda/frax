package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/adapters/openai/schemas"
	"github.com/petrjanda/frax/pkg/llm"
	"github.com/petrjanda/frax/synq/dsl"
)

func main() {
	ctx := context.Background()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "o3-mini"
	}

	endpoint := os.Getenv("OPENAI_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	// Create OpenAI adapter
	openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel(model), openai.WithEndpoint(endpoint))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	directiveSchema := schemas.NewOpenAISchemaGenerator().MustGenerateSchema(dsl.Directive{})

	payload, err := json.MarshalIndent(directiveSchema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema: %v", err)
	}

	err = os.WriteFile("schema.json", payload, 0644)
	if err != nil {
		log.Fatalf("Failed to write schema to file: %v", err)
	}

	structuredLLM := llm.NewBaseLLMWithStructuredOutput(
		directiveSchema, openaiLLM,
		llm.LLMWithStructuredOutputWithEvents(llm.NewJSONFileLogAgentEvents("log.json")),
	)

	system := `
		You are expert data engineer who turns plan text requests into directives that can be used to deploy data tests.
		
		Turn user input into structured description of how tests should be deployed with 3 components:
		* Entity selection - which defines which data entities (tables, views, etc.) should be tested
		* Column selection - which defines which columns should be tested
		* Tests - which defines the tests that should be deployed

		Instructions:
		* By default, try to create deterministic definitions (listing, dsl query, etc.)
		* If you can't conclusively generate deterministic definition, use LLM powered definition.
		* Avoid guessing deterministic definition. You should only use it if the query can be clearly expressed.
		* Try to keep the definitions dynamic, and avoid inferring specific values. This means the definition can apply dynamically on entities that will be created in the future.

		# Entity selection

		Entity selection chooses which tables should be tested. It could be done in two ways:
		1. By using query DSL language that describes rules that will match specific tables.
		2. By using LLM powered definition.
	`

	queries := []string{
		// "Monitor freshness of data of all dbt sources with P1 priority tag.",
		"Ensure fields used in join clauses are tested for uniqueness where appropriate.",
		// "All sources upstream of P1 and P2 data products should have freshness and volume test.",
		// "Tables impacting ML data products should test for data freshness and drift on feature columns.",
		// "Tables feeding into high priority Tableau dashboards should have business rules tests reflective of the most common queries.",
	}

	for _, q := range queries {
		fmt.Println("INPUT:")
		fmt.Println(q)
		fmt.Println("--------------------------------")

		// Create conversation history with the travel request
		history := llm.NewHistory(
			llm.NewUserMessage(q),
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
			switch t := msg.(type) {
			case *llm.TextMessage:

				// fmt.Println(t.Content)
				// fmt.Println("--------------------------------")
				var directive dsl.Directive
				err := json.Unmarshal([]byte(t.Content), &directive)
				if err != nil {
					log.Fatalf("Failed: %v", t.Content)
				}

				payload, err := json.MarshalIndent(directive, "", "  ")
				if err != nil {
					log.Fatalf("Failed to marshal message: %v", err)
				}

				fmt.Println("OUTPUT:")
				fmt.Println(string(payload))
				fmt.Println("========================================")

			default:
				// Quietly ignore other message types
				continue
			}

		}
	}
}
