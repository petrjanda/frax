package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	openai "github.com/petrjanda/frax/pkg/adapters/openai"
	llm "github.com/petrjanda/frax/pkg/llm"
)

// Person represents a person with structured data
type Person struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	Location string `json:"location"`
}

// Task represents a task with structured data
type Task struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Tags        []string `json:"tags"`
	DueDate     string   `json:"due_date"`
}

func main() {
	ctx := context.Background()

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI adapter
	openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// Test 1: Direct LLM usage
	fmt.Println("--- Test 1: Direct LLM Usage ---")
	response, err := openaiLLM.Invoke(ctx, &llm.LLMRequest{
		History: []llm.Message{
			&llm.UserMessage{Content: "Hello! How are you today?"},
		},
	})
	if err != nil {
		log.Printf("❌ Direct LLM failed: %v", err)
	} else {
		fmt.Printf("✅ Direct LLM response: %+v\n", response)
	}

	// Test 2: Create formatters for structured output
	fmt.Println("\n--- Test 2: Creating Formatters ---")

	// Person formatter schema
	personSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"email": {"type": "string"},
			"location": {"type": "string"}
		},
		"required": ["name", "age", "email", "location"]
	}`)

	// Task formatter schema
	taskSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {"type": "string"},
			"description": {"type": "string"},
			"priority": {"type": "string"},
			"tags": {"type": "array", "items": {"type": "string"}},
			"due_date": {"type": "string"}
		},
		"required": ["title", "description", "priority", "tags", "due_date"]
	}`)

	// Create LLMs with structured output that implement the LLM interface
	personLLM := llm.NewBaseLLMWithStructuredOutput(personSchema, openaiLLM)
	taskLLM := llm.NewBaseLLMWithStructuredOutput(taskSchema, openaiLLM)

	// Example of customizing with options (optional)
	customLLM := llm.NewBaseLLMWithStructuredOutput(
		personSchema,
		openaiLLM,
		llm.WithName("custom_person_tool"),
		llm.WithDescription("Custom tool for generating person data with specific requirements"),
	)

	fmt.Printf("✅ Created person LLM with structured output: %s\n", personLLM.Name())
	fmt.Printf("✅ Created task LLM with structured output: %s\n", taskLLM.Name())
	fmt.Printf("✅ Created custom LLM with structured output: %s\n", customLLM.Name())

	// Test 3: Use the LLM with structured output directly
	fmt.Println("\n--- Test 3: Using LLM with Structured Output Directly ---")

	// Use the person LLM with structured output directly - it implements the LLM interface!
	personResponse2, err := personLLM.Invoke(ctx, &llm.LLMRequest{
		History: []llm.Message{
			&llm.UserMessage{Content: "Generate a person who is a software engineer"},
		},
	})
	if err != nil {
		log.Printf("❌ Person LLM with structured output failed: %v", err)
	} else {
		log.Printf("✅ Person LLM with structured output response: %+v\n", personResponse2.Messages[0])
	}

	// Test 4: Use LLM with structured output directly as LLM
	fmt.Println("\n--- Test 4: Using LLM with Structured Output Directly as LLM ---")

	// Use the person LLM with structured output directly - it implements the LLM interface!
	personResponse, err := personLLM.Invoke(ctx, &llm.LLMRequest{
		History: []llm.Message{
			&llm.UserMessage{Content: "Generate a person who is a data scientist"},
		},
	})
	if err != nil {
		log.Printf("❌ Person LLM with structured output failed: %v", err)
	} else {
		log.Printf("✅ Person LLM with structured output response: %+v\n", personResponse.Messages[0])
	}

	// Test 5: Use task LLM with structured output directly
	fmt.Println("\n--- Test 5: Using Task LLM with Structured Output Directly ---")

	taskResponse, err := taskLLM.Invoke(ctx, &llm.LLMRequest{
		History: []llm.Message{
			&llm.UserMessage{Content: "Generate a task related to machine learning"},
		},
	})
	if err != nil {
		log.Printf("❌ Task LLM with structured output failed: %v", err)
	} else {
		log.Printf("✅ Task LLM with structured output response: %+v\n", taskResponse.Messages[0])
	}

	fmt.Println("\n🎉 All tests completed!")
}
