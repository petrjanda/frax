package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/adapters/openai"
	"github.com/petrjanda/frax/pkg/ai/structured"

	"github.com/petrjanda/frax/synq/dsl"

	_ "embed"
)

//go:embed prompts/planner_system.txt
var PromptPlannerSystem string

func main() {
	ctx := context.Background()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	openaiLLM, err := getAdapter()
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	directiveSchema := openai.NewOpenAISchemaGenerator().MustGenerateSchema(dsl.Directive{})
	writeSchema(directiveSchema, "schema.json")

	structured := structured.NewLLM(
		directiveSchema, openaiLLM,
		structured.WithEvents(ai.NewJSONFileLogAgentEvents("log.json")),
	)

	system := PromptPlannerSystem

	queries := []string{
		"Monitor freshness of data of all dbt sources with P1 priority tag.",
		"Ensure fields used in join clauses are tested for uniqueness where appropriate.",
		"All sources upstream of P1 and P2 data products should have freshness and volume test.",
		"Tables impacting ML data products should test for data freshness and drift on feature columns.",
		// "Tables feeding into high priority Tableau dashboards should have business rules tests reflective of the most common queries.",
	}

	for _, q := range queries {
		fmt.Println("INPUT:")
		fmt.Println(q)
		fmt.Println("--------------------------------")

		// Create conversation history with the travel request
		history := ai.NewHistory(
			ai.NewUserMessage(q),
		)

		req := ai.NewLLMRequest(
			ai.WithModel(getModel()),
			ai.WithSystem(system),
			ai.WithHistory(history),
			ai.WithTemperature(0.0),
			ai.WithMaxCompletionTokens(1000),
		)

		// Run the agent
		response, err := structured.Invoke(ctx, req)

		if err != nil {
			log.Fatalf("llm failed: %v", err)
		}

		// Print the conversation
		for _, msg := range response.Messages {
			switch t := msg.(type) {
			case *ai.TextMessage:
				var directive dsl.Directive
				err := json.Unmarshal([]byte(t.Content), &directive)
				if err != nil {
					log.Fatalf("failed unmarshalling: %v", t.Content)
				}

				payload, err := json.MarshalIndent(directive, "", "  ")
				if err != nil {
					log.Fatalf("failed marshalling: %v", err)
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

func getModel() ai.ModelId {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "o3-mini"
	}
	return ai.ModelId(model)
}

func getAdapter() (ai.LLM, error) {
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	endpoint := os.Getenv("OPENAI_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	// Create OpenAI adapter
	openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithEndpoint(endpoint))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	return openaiLLM, nil
}

func writeSchema(schema json.RawMessage, filename string) {
	payload, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema: %v", err)
	}

	err = os.WriteFile(filename, payload, 0644)
	if err != nil {
		log.Fatalf("Failed to write schema to file: %v", err)
	}
}
