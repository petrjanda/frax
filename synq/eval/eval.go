package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/adapters/openai"
	"github.com/petrjanda/frax/pkg/ai/workflows"
	"github.com/petrjanda/frax/synq/tasks"

	"github.com/petrjanda/frax/pkg/ai/eval"
	"github.com/petrjanda/frax/pkg/ai/eval/events"
	. "github.com/petrjanda/frax/pkg/ai/eval/expectations"

	_ "embed"
)

func main() {
	ctx := context.Background()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// SETUP

	litellm, err := getConnector()
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// VARIANT

	current := tasks.PlannerTask

	variants := []workflows.Task{
		// ai.NewLLMRequest(
		// 	ai.WithModel("claude-3-7-sonnet"),
		// 	ai.WithSystem(prompts.PlannerSystem),
		// 	ai.WithTemperature(0.0),
		// 	ai.WithMaxCompletionTokens(1000),
		// ),

		current.WithName("gemini").WithRequestOpts(ai.WithModel("gemini-2-5-pro")),
		current,

		// ai.NewLLMRequest(
		// 	ai.WithModel("claude-3-5-haiku"),
		// 	ai.WithSystem(prompts.PlannerSystem),
		// 	ai.WithTemperature(0.0),
		// 	ai.WithMaxCompletionTokens(1000),
		// ),
	}

	// CASES

	cases := []*eval.Case{
		eval.NewCase(
			"Monitor freshness of data of all dbt sources upstream of P1 data products.", eval.WithAlias("P1 sources fresh"),
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
		// Expect(NewScoringJudge(
		// 	"Drift definition", litellm, "claude-4-sonnet",
		// 	"Is the drift monitor column selection description unambiguous? It is unambiguous if you knew how to select columns based on the description. You can assume in future ", 50,
		// )),

		// eval.NewCase(
		// 	"All sources upstream of P1 and P2 data products should have freshness and volume test.", eval.WithAlias("Product upstream"),
		// ).
		// 	Expect(JSONPath("entities.query.data_product_impact.importance.0").Eq("P1")).
		// 	Expect(JSONPath("entities.query.data_product_impact.importance.1").Eq("P2")).
		// 	Expect(JSONPath("tests.table_stats_monitor.freshness").Eq(true)).
		// 	Expect(JSONPath("tests.table_stats_monitor.volume").Eq(true)),
	}

	suite := eval.NewSuite(
		events.NewTableSuiteEvents(), cases, variants,
	)

	if err := suite.Run(ctx, litellm); err != nil {
		log.Fatalf("Failed to run suite: %v", err)
	}
}

func getConnector() (ai.LLM, error) {
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
