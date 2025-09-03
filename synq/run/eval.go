package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/adapters/openai/schemas"
	"github.com/petrjanda/frax/pkg/llm"
	"github.com/petrjanda/frax/synq/dsl"
	"github.com/petrjanda/frax/synq/eval"
	. "github.com/petrjanda/frax/synq/eval"
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

	directiveSchema := schemas.NewOpenAISchemaGenerator().MustGenerateSchema(dsl.Directive{})
	structuredLLM := llm.NewBaseLLMWithStructuredOutput(
		directiveSchema, openaiLLM,
		llm.LLMWithStructuredOutputWithEvents(llm.NewJSONFileLogAgentEvents("eval.json")),
	)

	req := llm.NewLLMRequest(
		llm.WithSystem(prompts.PlannerSystem),
		llm.WithTemperature(0.0),
		llm.WithMaxCompletionTokens(1000),
	)

	cases := []*Case{
		NewCase("Monitor freshness of data of all dbt sources with P1 priority tag.",
			[]*Expectation{
				Expect("entities.query.entity_type", Eq("dbt_source")),
				Expect("entities.query.data_product_impact.importance.0", Eq("P1")),
				Expect("tests.table_stats_monitor.freshness", Eq(true)),
				Expect("tests.table_stats_monitor.volume", Eq(true)),
				Expect("tests.table_stats_monitor.change_delay", Eq(true)),
			}),

		NewCase("Ensure fields used in join clauses are tested for uniqueness where appropriate.",
			[]*Expectation{
				Expect("entities.query.entity_type", Eq("dbt_source")),
				Expect("entities.query.data_product_impact.importance.0", Eq("P1")),
				Expect("tests.table_stats_monitor.freshness", Eq(true)),
				Expect("tests.table_stats_monitor.volume", Eq(true)),
				Expect("tests.table_stats_monitor.change_delay", Eq(true)),
			}),

		NewCase("Tables impacting ML data products should test for data freshness and drift on feature columns.",
			[]*Expectation{
				Expect("entities.query.entity_type", Eq("dbt_model")),
				Expect("entities.query.data_product_impact.importance.0", Eq("P1")),
				Expect("tests.drift_monitor.columns.0", Eq("feature_columns")),
			}),

		NewCase("All sources upstream of P1 and P2 data products should have freshness and volume test.",
			[]*Expectation{
				Expect("entities.query.entity_type", Eq("dbt_source")),
				Expect("entities.query.data_product_impact.importance.0", Eq("P1")),
				Expect("tests.table_stats_monitor.freshness", Eq(true)),
				Expect("tests.table_stats_monitor.volume", Eq(true)),
				Expect("tests.table_stats_monitor.change_delay", Eq(true)),
			}),
	}

	eval.RunSuite[dsl.Directive](ctx, cases, structuredLLM, req)
}

func getAdapter() (llm.LLM, error) {
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

	return openaiLLM, nil
}
