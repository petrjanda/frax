package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/llm/structured"

	"github.com/petrjanda/frax/pkg/eval"
	"github.com/petrjanda/frax/pkg/eval/events"
	. "github.com/petrjanda/frax/pkg/eval/expectations"
	"github.com/petrjanda/frax/pkg/llm"
	"github.com/petrjanda/frax/synq/dsl"
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

	litellm, err := getAdapter()
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	directiveSchema := openai.NewOpenAISchemaGenerator().MustGenerateSchema(dsl.Directive{})
	structuredLLM := structured.NewLLM(directiveSchema, litellm)

	// VARIANT

	variants := []*llm.LLMRequest{
		// llm.NewLLMRequest(
		// 	llm.WithModel("claude-3-7-sonnet"),
		// 	llm.WithSystem(prompts.PlannerSystem),
		// 	llm.WithTemperature(0.0),
		// 	llm.WithMaxCompletionTokens(1000),
		// ),

		llm.NewLLMRequest(
			llm.WithModel("claude-4-sonnet"),
			llm.WithSystem(prompts.PlannerSystem),
			llm.WithTemperature(0.0),
			llm.WithMaxCompletionTokens(1000),
		),

		// llm.NewLLMRequest(
		// 	llm.WithModel("claude-3-5-haiku"),
		// 	llm.WithSystem(prompts.PlannerSystem),
		// 	llm.WithTemperature(0.0),
		// 	llm.WithMaxCompletionTokens(1000),
		// ),
	}

	// CASES

	judge := NewScoringJudge(
		"JSON", litellm, "claude-3-7-sonnet", "Is the message a valid JSON record? Valid means it's parseable as json", 50)

	cases := []*eval.Case{
		eval.NewCase(
			"Monitor freshness of data of all dbt sources with P1 priority tag.", eval.WithAlias("P1 sources fresh"),
		).
			Expect(JSONPath("entities.query.entity_type").Eq("dbt_source")).
			Expect(JSONPath("entities.query.data_product_impact.importance.0").Eq("P1")).
			Expect(JSONPath("tests.table_stats_monitor.freshness").Eq(true)).
			Expect(JSONPath("tests.table_stats_monitor.volume").DoesNotExist()).
			Expect(JSONPath("tests.table_stats_monitor.change_delay").DoesNotExist()),

		eval.NewCase(
			"Ensure fields used in join clauses are tested for uniqueness where appropriate.", eval.WithAlias("Unique join"),
		).
			Expect(JSONPath("entities.query.is_join").Eq(true)).
			Expect(JSONPath("tests.unique.columns.llm").Exists()),

		eval.NewCase(
			"Tables impacting ML data products should test for data freshness and drift on feature columns.", eval.WithAlias("ML data products"),
		).
			Expect(JSONPath("entities.query.data_product_impact.llm").Exists()).
			Expect(JSONPath("tests.table_stats_monitor.freshness").Eq(true)).
			Expect(JSONPath("tests.drift_monitor.columns.llm").Exists()),

		eval.NewCase(
			"All sources upstream of P1 and P2 data products should have freshness and volume test.", eval.WithAlias("Product upstream"),
		).
			Expect(JSONPath("entities.query.data_product_impact.importance.0").Eq("P1")).
			Expect(JSONPath("entities.query.data_product_impact.importance.1").Eq("P2")).
			Expect(JSONPath("tests.table_stats_monitor.freshness").Eq(true)).
			Expect(JSONPath("tests.table_stats_monitor.volume").Eq(true)).
			Expect(judge),
	}

	suite := eval.NewSuite(
		events.NewTableSuiteEvents(), cases, structuredLLM, variants,
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
