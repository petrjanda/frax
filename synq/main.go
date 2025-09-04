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

	"github.com/petrjanda/frax/synq/tasks"

	_ "embed"
)

func main() {
	ctx := context.Background()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	litellm, err := getAdapter()
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	queries := []string{
		"Monitor freshness of data of all dbt sources with P1 priority tag.",
		"Ensure fields used in join clauses are tested for uniqueness where appropriate.",
		"All sources upstream of P1 and P2 data products should have freshness and volume test.",
		"Tables impacting ML data products should test for data freshness and drift on feature columns.",
		// "Tables feeding into high priority Tableau dashboards should have business rules tests reflective of the most common queries.",
	}

	for _, q := range queries {
		// Run the agent
		response, err := tasks.PlannerTask.Invoke(ctx, litellm, ai.NewHistory(
			ai.NewUserMessage(q),
		))

		if err != nil {
			log.Fatalf("llm failed: %v", err)
		}

		payload, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			log.Fatalf("failed marshalling: %v", err)
		}

		fmt.Println("OUTPUT:")
		fmt.Println(string(payload))
		fmt.Println("========================================")

	}
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
