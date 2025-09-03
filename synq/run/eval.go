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
	"github.com/petrjanda/frax/synq/eval/events"
	. "github.com/petrjanda/frax/synq/eval/expectations"
	"github.com/petrjanda/frax/synq/prompts"

	_ "embed"
)

func main() {
	ctx := context.Background()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// SETUP

	openaiLLM, err := getAdapter()
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	directiveSchema := openai.NewOpenAISchemaGenerator().MustGenerateSchema(dsl.Directive{})
	structuredLLM := llm.NewBaseLLMWithStructuredOutput(directiveSchema, openaiLLM)

	// VARIANT

	variants := []*llm.LLMRequest{

		llm.NewLLMRequest(
			llm.WithModel("claude-3-7-sonnet"),
			llm.WithSystem(prompts.PlannerSystem),
			llm.WithTemperature(0.0),
			llm.WithMaxCompletionTokens(1000),
		),

		llm.NewLLMRequest(
			llm.WithModel("claude-4-sonnet"),
			llm.WithSystem(prompts.PlannerSystem),
			llm.WithTemperature(0.0),
			llm.WithMaxCompletionTokens(1000),
		),
	}

	// CASES

	cases := []*eval.Case{
		eval.NewCase(
			"Monitor freshness of data of all dbt sources with P1 priority tag.",
		).
			Expect(JSONPath("entities.query.entity_type").Eq("dbt_source")).
			Expect(JSONPath("entities.query.data_product_impact.importance.0").Eq("P1")).
			Expect(JSONPath("tests.table_stats_monitor.freshness").Eq(true)).
			Expect(JSONPath("tests.table_stats_monitor.volume").DoesNotExist()).
			Expect(JSONPath("tests.table_stats_monitor.change_delay").DoesNotExist()),

		eval.NewCase(
			"Ensure fields used in join clauses are tested for uniqueness where appropriate.",
		).
			Expect(JSONPath("entities.query.is_join").Eq(true)).
			Expect(JSONPath("tests.unique.columns.llm").Exists()),

		eval.NewCase(
			"Ensure fields used in join clauses are tested for uniqueness where appropriate.",
		).
			Expect(JSONPath("entities.query.is_join").Eq(true)).
			Expect(JSONPath("tests.unique.columns.llm").Exists()),

		eval.NewCase(
			"Tables impacting ML data products should test for data freshness and drift on feature columns.",
		).
			Expect(JSONPath("entities.query.data_product_impact.llm").Exists()).
			Expect(JSONPath("tests.table_stats_monitor.freshness").Eq(true)).
			Expect(JSONPath("tests.drift_monitor.columns.llm").Exists()),

		eval.NewCase(
			"All sources upstream of P1 and P2 data products should have freshness and volume test.",
		).
			Expect(JSONPath("entities.query.data_product_impact.importance.0").Eq("P1")).
			Expect(JSONPath("entities.query.data_product_impact.importance.1").Eq("P2")).
			Expect(JSONPath("tests.table_stats_monitor.freshness").Eq(true)).
			Expect(JSONPath("tests.table_stats_monitor.volume").Eq(true)),
	}

	suite := eval.NewSuite(
		events.NewLoggingSuiteEvents(), cases, structuredLLM, variants,
	)

	if err := suite.Run(ctx); err != nil {
		log.Fatalf("Failed to run suite: %v", err)
	}
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
