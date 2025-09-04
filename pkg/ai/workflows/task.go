package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/adapters/openai"
	"github.com/petrjanda/frax/pkg/ai/structured"
)

type Task[T any] interface {
	Name() string
	Invoke(ctx context.Context, llm ai.LLM, history ai.History) (T, error)
}

// TEST OUTPUT

type TextTask struct {
	Name_   string
	Request *ai.LLMRequest
}

type TextTaskOpts = func(*TextTask)

func TextTaskWithRequest(request *ai.LLMRequest) TextTaskOpts {
	return func(t *TextTask) {
		t.Request = request
	}
}

func NewTask(name string, request *ai.LLMRequest, opts ...TextTaskOpts) Task[*ai.LLMResponse] {
	t := &TextTask{
		Name_:   name,
		Request: request,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *TextTask) Name() string {
	return t.Name_
}

func (t *TextTask) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	return llm.Invoke(ctx, t.Request.Clone(ai.WithHistory(history)))
}

// STRUCTURED OUTPUT

type StructuredTask[T any] struct {
	Name_   string
	LLM     ai.LLM
	Request *ai.LLMRequest
}

type StructuredTaskOpts[T any] = func(*StructuredTask[T])

func StructuredTaskWithRequest[T any](request *ai.LLMRequest) StructuredTaskOpts[T] {
	return func(t *StructuredTask[T]) {
		t.Request = request
	}
}

func StructuredTaskWithLLM[T any](llm ai.LLM) StructuredTaskOpts[T] {
	return func(t *StructuredTask[T]) {
		t.LLM = llm
	}
}

func NewStructuredTask[T any](name string, request *ai.LLMRequest) Task[T] {
	return &StructuredTask[T]{
		Name_:   name,
		Request: request,
	}
}

func (t *StructuredTask[T]) Name() string {
	return t.Name_
}

func (t *StructuredTask[T]) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (T, error) {
	structuredLLM := structured.NewLLM(openai.NewOpenAISchemaGenerator().MustGenerateSchema((*T)(nil)), llm, structured.WithEvents(ai.NewJSONFileLogAgentEvents("pirates.json")))
	response, err := structuredLLM.Invoke(ctx, t.Request.Clone(ai.WithHistory(history)))
	if err != nil {
		return *new(T), err
	}

	lastMessage := response.Messages[len(response.Messages)-1]
	if lastMessage.Kind() != ai.MessageKindText {
		return *new(T), fmt.Errorf("last message is not a text message")
	}

	var result T
	err = json.Unmarshal([]byte(lastMessage.(*ai.TextMessage).Content), &result)
	if err != nil {
		return *new(T), err
	}

	return result, nil
}
