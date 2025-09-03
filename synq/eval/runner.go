package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/petrjanda/frax/pkg/llm"

	_ "embed"
)

func RunSuite[T any](ctx context.Context, cases []*Case, model llm.LLM, req *llm.LLMRequest) error {
	for _, q := range cases {
		// Create conversation history with the travel request
		history := llm.NewHistory(
			llm.NewUserMessage(q.Input),
		)

		// Run the agent
		response, err := model.Invoke(ctx,
			req.Clone(llm.WithHistory(history)),
		)

		if err != nil {
			log.Fatalf("llm failed: %v", err)
		}

		lastMessage, ok := response.Messages[len(response.Messages)-1].(*llm.TextMessage)
		if !ok {
			log.Fatalf("last message is not a text message: %v", response.Messages[len(response.Messages)-1])
		}

		var directive T
		if err = json.Unmarshal([]byte(lastMessage.Content), &directive); err != nil {
			log.Fatalf("failed unmarshalling: %v", lastMessage.Content)
		}

		if payload, err := json.MarshalIndent(directive, "", "  "); err != nil {
			log.Fatalf("failed marshalling: %v", err)
		} else {
			fmt.Println("OUTPUT:")
			fmt.Println(string(payload))
			fmt.Println("========================================")
		}

		q.Validate([]byte(lastMessage.Content))
	}

	return nil
}
