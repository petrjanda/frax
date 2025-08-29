package llm

import (
	"context"
)

// LLM represents a language model that can process requests and generate responses
type LLM interface {
	Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error)
}

type History = []Message

func NewHistory(messages ...Message) History {
	return messages
}
