package main

import (
	"context"
	"fmt"
	"log"
	"os"

	openai "github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/adapters/openai/schemas"
	llm "github.com/petrjanda/frax/pkg/llm"
)

// Person represents a person with structured data
type Person struct {
	Name     string `json:"name" jsonschema:"required"`
	Age      int    `json:"age" jsonschema:"required"`
	Email    string `json:"email" jsonschema:"required"`
	Location string `json:"location" jsonschema:"required"`
}

// Task represents a task with structured data
type Task struct {
	Title       string   `json:"title" jsonschema:"required"`
	Description string   `json:"description" jsonschema:"required"`
	Priority    string   `json:"priority" jsonschema:"required"`
	Tags        []string `json:"tags" jsonschema:"required"`
	DueDate     string   `json:"due_date" jsonschema:"required"`
}

func main() {
	ctx := context.Background()

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("OPENAI_API_KEY environment variable is not set, skipping API tests")
		log.Println("To run the full example, set OPENAI_API_KEY and run again")
		return
	}

	// Create OpenAI adapter
	model, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	schemaGenerator := schemas.NewOpenAISchemaGenerator()

	personLLM := llm.NewBaseLLMWithStructuredOutput(
		schemaGenerator.MustGenerateSchema(Person{}), model)

	taskLLM := llm.NewBaseLLMWithStructuredOutput(
		schemaGenerator.MustGenerateSchema(Task{}), model)

	{
		// Test 3: Use the LLM with structured output directly
		fmt.Println("\n--- Test 3: Using LLM with Structured Output Directly ---")

		// Use the person LLM with structured output directly - it implements the LLM interface!
		resp, err := personLLM.Invoke(ctx,
			llm.NewLLMRequest(llm.NewHistory(
				llm.NewUserMessage("Generate a person who is a software engineer"),
			)))

		if err != nil {
			log.Printf("❌ Person LLM with structured output failed: %v", err)
		} else {
			log.Printf("✅ Person LLM with structured output response: %+v\n", resp.Messages[0])
		}
	}

	{
		// Test 4: Use LLM with structured output directly as LLM
		fmt.Println("\n--- Test 4: Using LLM with Structured Output Directly as LLM ---")

		// Use the person LLM with structured output directly - it implements the LLM interface!
		resp, err := personLLM.Invoke(ctx, &llm.LLMRequest{
			History: llm.NewHistory(
				llm.NewUserMessage("Generate a person who is a data scientist"),
			),
		})

		if err != nil {
			log.Printf("❌ Person LLM with structured output failed: %v", err)
		} else {
			log.Printf("✅ Person LLM with structured output response: %+v\n", resp.Messages[0])
		}
	}

	{
		// Test 5: Use task LLM with structured output directly
		fmt.Println("\n--- Test 5: Using Task LLM with Structured Output Directly ---")

		resp, err := taskLLM.Invoke(ctx, &llm.LLMRequest{
			History: llm.NewHistory(
				llm.NewUserMessage("Generate a task related to machine learning"),
			),
		})

		if err != nil {
			log.Printf("❌ Task LLM with structured output failed: %v", err)
		} else {
			log.Printf("✅ Task LLM with structured output response: %+v\n", resp.Messages[0])
		}
	}

	fmt.Println("\n🎉 All tests completed!")
}
