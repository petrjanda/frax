package tasks

import (
	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/workflows"
)

var PlannerTask = workflows.NewStructuredTask[Directive](
	"planner",
	ai.NewLLMRequest(
		ai.WithModel("claude-4-sonnet"),
		ai.WithSystem(`
			You are expert data engineer who turns plan text requests into directives that can be used to deploy data tests.
					
			<task>
				Turn user input into structured description of how tests should be deployed with 3 components:
				* Entity selection - which defines which data entities (tables, views, etc.) should be tested
				* Column selection - which defines which columns should be tested
				* Tests - which defines the tests that should be deployed
			</task>

			<critical>
				* By default, try to create deterministic definitions (listing, dsl query, etc.)
				* If you can't conclusively generate deterministic definition, use LLM powered definition.
				* Avoid guessing deterministic definition. You should only use it if the query can be clearly expressed.
				* Try to keep the definitions dynamic, and avoid inferring specific values. This means the definition can apply dynamically on entities that will be created in the future.
			</critical>

			<entity_selection>
				Entity selection chooses which tables should be tested. It could be done in two ways:
				1. By using query DSL language that describes rules that will match specific tables.
				2. By using LLM powered definition.

				<examples>
					Select all tables/entities that contain JOIN clauses: "entities": {"query": {"is_join": true}}
				</examples>
			</entity_selection>
		`),
		ai.WithTemperature(0.0),
		ai.WithMaxCompletionTokens(1000),
	),
)

// Directive describes how tests should deploy
type Directive struct {
	// Defines which entities (typically tables) should be tested. Use either query or llm, not both.
	Entities *EntitiesDirective `json:"entities,omitempty" jsonschema:"required"`

	// Add one or multiple tests to deploy.
	Tests *Tests `json:"tests,omitempty" jsonschema:"required"`
}

type EntitiesDirective struct {
	// Structured query selector. Operands are ANDed together.
	// Query *Operand `json:"query,omitempty" jsonschema:"oneof_required=query"`
	Query *Operand `json:"query,omitempty"`

	// Prompt should only contain instructions to select entities. It should not specify what tests to deploy or columns to test.
	// LLM string `json:"llm,omitempty" jsonschema:"oneof_required=llm"`
	LLM string `json:"llm,omitempty"`
}

type Operand struct {
	// By how its materialized in the database (table, view, materialized_view)
	Materialization string `json:"materialization,omitempty" jsonschema:"anyof_required=materialization"`

	// By its type (dbt_model, dbt_source, table, view, airflow_task, tableau_dashboard, etc.)
	EntityType string `json:"entity_type,omitempty" jsonschema:"anyof_required=entity_type"`

	// By its name, typically table name.
	Name string `json:"name,omitempty" jsonschema:"anyof_required=name"`

	// By its schema
	Schema string `json:"schema,omitempty" jsonschema:"anyof_required=schema"`

	// Matches entities that are data sources
	IsSource bool `json:"is_source,omitempty" jsonschema:"anyof_required=is_source"`

	// Matches entities that JOIN the data.
	IsJoin bool `json:"is_join,omitempty" jsonschema:"anyof_required=is_join"`

	// Matches entities that are UNION the data.
	IsUnion bool `json:"is_union,omitempty" jsonschema:"anyof_required=is_union"`

	// We are matching all entities upstream of data products.
	DataProductImpact *struct {
		// Match data products by importance (often referred to as data product severity or priority).
		Importance []string `json:"importance,omitempty" jsonschema:"oneof_required=importance,enum=P1,enum=P2,enum=P3"`

		// Use to express criteria about data products that can't be expressed with other operations.
		LLM string `json:"llm,omitempty" jsonschema:"oneof_required=llm"`
	} `json:"data_product_impact,omitempty" jsonschema:"anyof_required=data_product_impact"`

	// Matches entity that has specific key/value annotation. Tags, meta fields and other attributes are expressed as annotations.
	Annotation *struct {
		Key   string `json:"key,omitempty" jsonschema:"required"`
		Value string `json:"value,omitempty" jsonschema:"required"`
	} `json:"annotation,omitempty" jsonschema:"anyof_required=annotation"`

	// // It will match all our operands and then expand selection to all nodes upstream via lineage.
	// ExpandUpstream bool `json:"expand_upstream,omitempty"`

	// // It will match all our operands and then expand selection to all nodes downstream via lineage.
	// ExpandDownstream bool `json:"expand_downstream,omitempty"`
}

type LLMEntitySelector struct {
	Prompt string `json:"prompt,omitempty" jsonschema:"required"`
}

type Tests struct {
	Unique *struct {
		Columns *ColumnsDirective `json:"columns,omitempty" jsonschema:"required"`
	} `json:"unique,omitempty" jsonschema:"anyof_required=unique"`

	NotNull *struct {
		Columns *ColumnsDirective `json:"columns,omitempty" jsonschema:"required"`
	} `json:"not_null,omitempty" jsonschema:"anyof_required=not_null"`

	Pattern *struct {
		Column  string `json:"column,omitempty" jsonschema:"required"`
		Pattern string `json:"pattern,omitempty" jsonschema:"required"`
	} `json:"pattern,omitempty" jsonschema:"anyof_required=pattern"`

	Length *struct {
		Column string `json:"column,omitempty" jsonschema:"required"`
		Length int    `json:"length,omitempty" jsonschema:"required"`
	} `json:"length,omitempty" jsonschema:"anyof_required=length"`

	Range *struct {
		Column string `json:"column,omitempty" jsonschema:"required"`

		// Minimal value. If specified alone, the test acts as >= Min.
		Min int `json:"min"`

		// Maximum value. If specified alone, the test acts as <= Max.
		Max int `json:"max"`
	} `json:"range,omitempty" jsonschema:"anyof_required=range"`

	AcceptedValues *struct {
		Column string   `json:"column,omitempty" jsonschema:"required"`
		Values []string `json:"values,omitempty" jsonschema:"required"`
	} `json:"accepted_values,omitempty" jsonschema:"anyof_required=accepted_values"`

	// Monitor table for key metrics such as volume of rows, freshness of data, and delay between changes to number of rows. Doesn't need column selection.
	TableStatsMonitor *struct {
		Volume      bool `json:"volume,omitempty"`
		Freshness   bool `json:"freshness,omitempty"`
		ChangeDelay bool `json:"change_delay,omitempty"`
	} `json:"table_stats_monitor,omitempty" jsonschema:"anyof_required=table_stats_monitor"`

	DriftMonitor *struct {
		Columns *ColumnsDirective `json:"columns,omitempty" jsonschema:"required"`
	} `json:"drift_monitor,omitempty" jsonschema:"anyof_required=drift_monitor"`

	// Use to express ambiguous criteria that can't be expressed with other operations.
	LLM string `json:"llm,omitempty" jsonschema:"anyof_required=llm"`
}

type ColumnsDirective struct {
	// Use to specify exact column names.
	Names []string `json:"names,omitempty" jsonschema:"oneof_required=names"`

	// Use to select all columns of a certain types.
	Types []string `json:"types,omitempty" jsonschema:"oneof_required=types"`

	// Use to express ambiguous criteria that can't be expressed with other operations. Prompt should only contain instructions to select columns.
	LLM string `json:"llm,omitempty" jsonschema:"oneof_required=llm"`
}
