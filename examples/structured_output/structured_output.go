package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/llm"
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
	fmt.Println("Structured Output Example with Formatters and Invokers")
	fmt.Println("=====================================================")

	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set your OpenAI API key to run this example:")
		fmt.Println("export OPENAI_API_KEY=your_api_key_here")
		fmt.Println("\nFalling back to demonstration mode...")
		demoModeStructuredOutput()
		return
	}

	// Create OpenAI adapter
	fmt.Println("🔑 Creating OpenAI adapter...")
	openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel("gpt-4o"))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}
	fmt.Println("✅ OpenAI adapter created successfully!")

	// Create formatters for different data structures
	personFormatter := createPersonFormatter()
	taskFormatter := createTaskFormatter()

	fmt.Printf("🧩 Person formatter loaded: %s\n", personFormatter.Description())
	fmt.Printf("🧩 Task formatter loaded: %s\n", taskFormatter.Description())

	// Test structured output generation
	ctx := context.Background()

	// Test 1: Generate a person
	fmt.Println("\n--- Test 1: Generate Person ---")
	person, err := generatePerson(ctx, openaiLLM, personFormatter)
	if err != nil {
		log.Printf("❌ Person generation failed: %v", err)
	} else {
		fmt.Printf("✅ Generated person: %+v\n", person)
	}

	// Test 2: Generate a task
	fmt.Println("\n--- Test 2: Generate Task ---")
	task, err := generateTask(ctx, openaiLLM, taskFormatter)
	if err != nil {
		log.Printf("❌ Task generation failed: %v", err)
	} else {
		fmt.Printf("✅ Generated task: %+v\n", task)
	}

	// Test 3: Use the invoker pattern
	fmt.Println("\n--- Test 3: Using Invoker Pattern ---")
	invoker := llm.NewInvoker(openaiLLM, personFormatter)

	// Create a formatter request
	request := llm.NewFormatterRequest([]llm.Message{
		&llm.UserMessage{Content: "Generate a person who is a software engineer"},
	})

	// Use the invoker to get structured output
	structuredData, err := invoker.Invoke(ctx, request)
	if err != nil {
		log.Printf("❌ Invoker failed: %v", err)
	} else {
		// Parse the structured data
		var extractedPerson Person
		if err := json.Unmarshal(structuredData, &extractedPerson); err != nil {
			log.Printf("❌ Failed to parse person data: %v", err)
		} else {
			fmt.Printf("✅ Invoker generated person: %+v\n", extractedPerson)
		}
	}

	fmt.Println("\n🎉 Structured Output Example completed!")
	fmt.Println("This demonstrates:")
	fmt.Println("- Formatter tools that enforce output structure")
	fmt.Println("- Forced tool usage to ensure structured output")
	fmt.Println("- Invoker pattern for easy structured data extraction")
	fmt.Println("- Type-safe Go structs from LLM output")
}

// createPersonFormatter creates a formatter for Person structs
func createPersonFormatter() llm.Formatter {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "Full name of the person"
			},
			"age": {
				"type": "integer",
				"description": "Age of the person"
			},
			"email": {
				"type": "string",
				"description": "Email address of the person"
			},
			"location": {
				"type": "string",
				"description": "City and country where the person lives"
			}
		},
		"required": ["name", "age", "email", "location"]
	}`)

	return llm.NewBaseFormatter(
		"person_formatter",
		"Formats and validates person data with name, age, email, and location",
		schema,
	)
}

// createTaskFormatter creates a formatter for Task structs
func createTaskFormatter() llm.Formatter {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {
				"type": "string",
				"description": "Title of the task"
			},
			"description": {
				"type": "string",
				"description": "Detailed description of the task"
			},
			"priority": {
				"type": "string",
				"enum": ["low", "medium", "high", "urgent"],
				"description": "Priority level of the task"
			},
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"description": "Tags associated with the task"
			},
			"due_date": {
				"type": "string",
				"description": "Due date in YYYY-MM-DD format"
			}
		},
		"required": ["title", "description", "priority", "tags", "due_date"]
	}`)

	return llm.NewBaseFormatter(
		"task_formatter",
		"Formats and validates task data with title, description, priority, tags, and due date",
		schema,
	)
}

// generatePerson generates a person using the formatter
func generatePerson(ctx context.Context, model llm.LLM, formatter llm.Formatter) (*Person, error) {
	// Create an invoker that handles everything directly
	invoker := llm.NewInvoker(model, formatter)

	// Create a formatter request
	request := llm.NewFormatterRequest([]llm.Message{
		&llm.UserMessage{Content: "Generate a person who is a software engineer"},
	})

	// Use the invoker to get structured output
	structuredData, err := invoker.Invoke(ctx, request)
	if err != nil {
		return nil, err
	}

	// Parse the structured data
	var person Person
	if err := json.Unmarshal(structuredData, &person); err != nil {
		return nil, fmt.Errorf("failed to parse person data: %w", err)
	}

	return &person, nil
}

// generateTask generates a task using the formatter
func generateTask(ctx context.Context, model llm.LLM, formatter llm.Formatter) (*Task, error) {
	// Create an invoker that handles everything directly
	invoker := llm.NewInvoker(model, formatter)

	// Create a formatter request
	request := llm.NewFormatterRequest([]llm.Message{
		&llm.UserMessage{Content: "Generate a task related to software development"},
	})

	// Use the invoker to get structured output
	structuredData, err := invoker.Invoke(ctx, request)
	if err != nil {
		return nil, err
	}

	// Parse the structured data
	var task Task
	if err := json.Unmarshal(structuredData, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task data: %w", err)
	}

	return &task, nil
}

// demoModeStructuredOutput runs without OpenAI API key for demonstration purposes
func demoModeStructuredOutput() {
	fmt.Println("\n📚 Demonstration Mode")
	fmt.Println("====================")

	personFormatter := createPersonFormatter()
	taskFormatter := createTaskFormatter()

	fmt.Printf("🧩 Person formatter: %s\n", personFormatter.Description())
	fmt.Printf("🧩 Task formatter: %s\n", taskFormatter.Description())

	fmt.Println("\nThe formatter system works as follows:")
	fmt.Println("1. Create a formatter with a JSON schema")
	fmt.Println("2. Use an Invoker that handles everything internally")
	fmt.Println("3. The LLM is forced to call the formatter tool")
	fmt.Println("4. The formatter validates and returns the structured data")
	fmt.Println("5. Simple one-line API for structured output generation")

	fmt.Println("\nExample usage:")
	fmt.Println("invoker := llm.NewInvoker(llm, personFormatter)")
	fmt.Println("person, err := invoker.InvokeWithFormatAndExtract(ctx, prompt, format, &person)")

	fmt.Println("\nTo run with real AI:")
	fmt.Println("1. Get an OpenAI API key from https://platform.openai.com/")
	fmt.Println("2. Set the environment variable: export OPENAI_API_KEY=your_key")
	fmt.Println("3. Run this example again")
}
