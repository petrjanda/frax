package dsl

// Discriminator values for entity selection
const (
	EntitySelectionDeterministic = "deterministic"
	EntitySelectionLLM           = "llm"
)

// Base interface for entity selection
type EntitySelector interface{ isEntitySelector() }

type DeterministicEntitySelector struct {
	Kind  string                `json:"kind" jsonschema:"const=deterministic"`
	Query *EntitySelectionQuery `json:"query"`
}

func (DeterministicEntitySelector) isEntitySelector() {}

type LLMEntitySelector struct {
	Kind   string `json:"kind" jsonschema:"const=llm"`
	Prompt string `json:"prompt"`
}

func (LLMEntitySelector) isEntitySelector() {}

type EntitySelectionQuery struct {
	Query string `json:"query" jsonschema:"required"`
}

// Discriminator values for column selection
const (
	ColumnSelectionList = "list"
	ColumnSelectionType = "type"
	ColumnSelectionLLM  = "llm"
)

// Base interface for column selection
type ColumnSelector interface{ isColumnSelector() }

type ListColumnSelector struct {
	Kind    string   `json:"kind" jsonschema:"const=list"`
	Columns []string `json:"columns"`
}

func (ListColumnSelector) isColumnSelector() {}

type TypeColumnSelector struct {
	Kind  string   `json:"kind" jsonschema:"const=type"`
	Types []string `json:"types"`
}

func (TypeColumnSelector) isColumnSelector() {}

type LLMColumnSelector struct {
	Kind   string `json:"kind" jsonschema:"const=llm"`
	Prompt string `json:"prompt"`
}

func (LLMColumnSelector) isColumnSelector() {}

type TestType string

const (
	TestTypeEqual              TestType = "equal"
	TestTypeNotEqual           TestType = "not_equal"
	TestTypeGreaterThan        TestType = "greater_than"
	TestTypeLessThan           TestType = "less_than"
	TestTypeGreaterThanOrEqual TestType = "greater_than_or_equal"
	TestTypeLessThanOrEqual    TestType = "less_than_or_equal"
)

type TestCase struct {
	Type TestType `json:"type"`
}

type Directive struct {
	Entities any        `json:"entities" jsonschema:"oneof_type=DeterministicEntitySelector,LLMEntitySelector"`
	Columns  any        `json:"columns" jsonschema:"oneof_type=ListColumnSelector,TypeColumnSelector,LLMColumnSelector"`
	Tests    []TestCase `json:"tests"`
}
