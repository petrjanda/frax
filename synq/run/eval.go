package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/petrjanda/frax/pkg/adapters/openai"

	"github.com/petrjanda/frax/pkg/llm"
	"github.com/petrjanda/frax/synq/dsl"
	"github.com/petrjanda/frax/synq/eval"
	. "github.com/petrjanda/frax/synq/eval/json"
	"github.com/petrjanda/frax/synq/prompts"

	_ "embed"
)

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
	structuredLLM := llm.NewBaseLLMWithStructuredOutput(
		directiveSchema, openaiLLM,
		llm.LLMWithStructuredOutputWithEvents(llm.NewJSONFileLogAgentEvents("eval.json")),
	)

	req := llm.NewLLMRequest(
		llm.WithModel(getModel()),
		llm.WithSystem(prompts.PlannerSystem),
		llm.WithTemperature(0.0),
		llm.WithMaxCompletionTokens(1000),
	)

	cases := []*eval.Case{
		eval.NewCase("Monitor freshness of data of all dbt sources with P1 priority tag.").
			Expect(JSON("entities.query.entity_type", Eq("dbt_source"))).
			Expect(JSON("entities.query.data_product_impact.importance.0", Eq("P1"))).
			Expect(JSON("tests.table_stats_monitor.freshness", Eq(true))).
			Expect(JSON("tests.table_stats_monitor.volume", Eq(true))).
			Expect(JSON("tests.table_stats_monitor.change_delay", DoesNotExist)).
			Expect(JSON("tests.table_stats_monitor.volume", DoesNotExist)).
			Expect(JSON("tests.table_stats_monitor.change_delay", DoesNotExist)),

		eval.NewCase("Ensure fields used in join clauses are tested for uniqueness where appropriate.").
			Expect(JSON("entities.query.is_join", Eq(true))).
			Expect(JSON("tests.unique.columns.llm", Exists)),

		eval.NewCase("Ensure fields used in join clauses are tested for uniqueness where appropriate.").
			Expect(JSON("entities.query.is_join", Eq(true))).
			Expect(JSON("tests.unique.columns.llm", Exists)),

		eval.NewCase("Tables impacting ML data products should test for data freshness and drift on feature columns.").
			Expect(JSON("entities.query.data_product_impact.llm", Exists)).
			Expect(JSON("tests.table_stats_monitor.freshness", Eq(true))).
			Expect(JSON("tests.drift_monitor.columns.llm", Exists)),

		eval.NewCase("All sources upstream of P1 and P2 data products should have freshness and volume test.").
			Expect(JSON("entities.query.data_product_impact.importance.0", Eq("P1"))).
			Expect(JSON("entities.query.data_product_impact.importance.1", Eq("P2"))).
			Expect(JSON("tests.table_stats_monitor.freshness", Eq(true))).
			Expect(JSON("tests.table_stats_monitor.volume", Eq(true))),
	}

	suite := eval.NewSuite[dsl.Directive](cases, structuredLLM, req)
	suite.Run(ctx)
}

func getModel() string {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "o3-mini"
	}
	return model
}

func getAdapter() (llm.LLM, error) {
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
