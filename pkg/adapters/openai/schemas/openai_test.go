package schemas

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestPerson represents a person with structured data
type TestPerson struct {
	Name     string `json:"name" jsonschema:"required"`
	Age      int    `json:"age" jsonschema:"required"`
	Email    string `json:"email" jsonschema:"required"`
	Location string `json:"location" jsonschema:"required"`
}

// TestTask represents a task with structured data
type TestTask struct {
	Title       string   `json:"title" jsonschema:"required"`
	Description string   `json:"description" jsonschema:"required"`
	Priority    string   `json:"priority" jsonschema:"required"`
	Tags        []string `json:"tags" jsonschema:"required"`
	DueDate     string   `json:"due_date" jsonschema:"required"`
}

func TestNewOpenAISchemaGenerator(t *testing.T) {
	generator := NewOpenAISchemaGenerator()
	if generator == nil {
		t.Fatal("Expected generator to be created")
	}
	if generator.reflector == nil {
		t.Fatal("Expected reflector to be initialized")
	}
}

func TestGenerateSchema(t *testing.T) {
	generator := NewOpenAISchemaGenerator()

	// Test Person schema
	person := TestPerson{}
	schema, err := generator.GenerateSchema(person)
	if err != nil {
		t.Fatalf("Failed to generate person schema: %v", err)
	}

	// Verify it's valid JSON
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	// Check for required fields
	if schemaMap["type"] != "object" {
		t.Errorf("Expected type to be 'object', got %v", schemaMap["type"])
	}

	// Check that required fields are present
	required, ok := schemaMap["required"].([]interface{})
	if !ok {
		t.Fatal("Expected 'required' field to be present")
	}
	if len(required) != 4 {
		t.Errorf("Expected 4 required fields, got %d", len(required))
	}

	// Test Task schema
	task := TestTask{}
	schema, err = generator.GenerateSchema(task)
	if err != nil {
		t.Fatalf("Failed to generate task schema: %v", err)
	}

	// Verify it's valid JSON
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	// Check for required fields
	if schemaMap["type"] != "object" {
		t.Errorf("Expected type to be 'object', got %v", schemaMap["type"])
	}

	// Check that required fields are present
	required, ok = schemaMap["required"].([]interface{})
	if !ok {
		t.Fatal("Expected 'required' field to be present")
	}
	if len(required) != 5 {
		t.Errorf("Expected 5 required fields, got %d", len(required))
	}
}

func TestGenerateSchemaFromStruct(t *testing.T) {
	// Test the generic function
	schema, err := GenerateSchemaFromStruct[TestPerson]()
	if err != nil {
		t.Fatalf("Failed to generate schema from struct type: %v", err)
	}

	// Verify it's valid JSON
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	if schemaMap["type"] != "object" {
		t.Errorf("Expected type to be 'object', got %v", schemaMap["type"])
	}
}

func TestValidateOpenAICompatibility(t *testing.T) {
	generator := NewOpenAISchemaGenerator()

	// Test with a valid schema
	person := TestPerson{}
	schema, err := generator.GenerateSchema(person)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Validate compatibility
	if err := generator.ValidateOpenAICompatibility(schema); err != nil {
		t.Errorf("Expected schema to be OpenAI compatible: %v", err)
	}
}

func TestGenerateSchemaFromType(t *testing.T) {
	generator := NewOpenAISchemaGenerator()

	// Test with a type
	schema, err := generator.GenerateSchemaFromType(reflect.TypeOf(TestPerson{}))
	if err != nil {
		t.Fatalf("Failed to generate schema from type: %v", err)
	}

	// Verify it's valid JSON
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	if schemaMap["type"] != "object" {
		t.Errorf("Expected type to be 'object', got %v", schemaMap["type"])
	}
}
