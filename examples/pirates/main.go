package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/adapters/openai"
	"github.com/petrjanda/frax/pkg/ai/workflows"
)

type Translation struct {
	// Input message
	In string `json:"in" jsonschema:"required"`

	// Output message
	Out string `json:"out" jsonschema:"required"`
}

var PiratesTask = workflows.NewStructuredTask[Translation](
	"arrrr",
	ai.NewLLMRequest(
		ai.WithModel("claude-4-sonnet"),
		ai.WithSystem(`
			You are a pirate translator.

			User will provide a message in English and your task is to translate it into pirate speech.

			Important:
			* Keep length of your response similar to user's message.
		`),
		ai.WithTemperature(1.0),
		ai.WithMaxCompletionTokens(1000),
	),
)

var ReversePiratesTask = workflows.NewStructuredTask[Translation](
	"no-arrrr",
	ai.NewLLMRequest(
		ai.WithModel("claude-4-sonnet"),
		ai.WithSystem(`
			You are a pirate translator.

			User will provide a message in pirate speech and your task is to translate it into English.

			Important:
			* Keep length of your response similar to user's message.
		`),
		ai.WithTemperature(1.0),
		ai.WithMaxCompletionTokens(1000),
	),
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	litellm, _ := getLLM()

	queries := []string{
		"How are you?",
		"I love you",
		"Hey",
		"Should we attack this ship?",
	}

	for _, q := range queries {
		// Translate to pirate
		response, _ := PiratesTask.Invoke(ctx, litellm, ai.NewHistory(
			ai.NewUserMessage(q),
		))

		// Print the conversation
		pirateTranslation := toTranslation(response.LastMessageAsText().Content)
		fmt.Printf("%s ==> %s\n", pirateTranslation.In, pirateTranslation.Out)

		// Translate back to English
		response, _ = ReversePiratesTask.Invoke(ctx, litellm, ai.NewHistory(
			ai.NewUserMessage(pirateTranslation.Out),
		))

		englishTranslation := toTranslation(response.LastMessageAsText().Content)
		fmt.Printf("%s ==> %s\n", englishTranslation.In, englishTranslation.Out)
		fmt.Println("========================================")
	}
}

func getLLM() (ai.LLM, error) {
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

func toTranslation(text string) Translation {
	var pirateTranslation Translation
	err := json.Unmarshal([]byte(text), &pirateTranslation)
	if err != nil {
		panic(err)
	}

	return pirateTranslation
}
