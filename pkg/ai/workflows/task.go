package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/adapters/openai"
	"github.com/petrjanda/frax/pkg/ai/structured"
)

type Task interface {
	Name() string
	Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error)
	Clone() Task

	WithName(name string) Task
}

// TEST OUTPUT

type TextTask struct {
	Name_   string
	Request *ai.LLMRequest
}

func NewTask(name string, request *ai.LLMRequest) Task {
	return &TextTask{
		Name_:   name,
		Request: request,
	}
}

func (t *TextTask) Name() string {
	return t.Name_
}

func (t *TextTask) Clone() Task {
	return NewTask(t.Name_, t.Request.Clone())
}

func (t *TextTask) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	return llm.Invoke(ctx, t.Request.Clone(ai.WithHistory(history)))
}

func (t *TextTask) WithRequestOpts(opts ...ai.LLMRequestOpts) *TextTask {
	new := t.Clone().(*TextTask)
	for _, opt := range opts {
		opt(new.Request)
	}

	return t
}

func (t *TextTask) WithName(name string) Task {
	new := t.Clone().(*TextTask)
	new.Name_ = name
	return new
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

func NewStructuredTask[T any](name string, request *ai.LLMRequest) Task {
	return &StructuredTask[T]{
		Name_:   name,
		Request: request,
	}
}

func (t *StructuredTask[T]) Name() string {
	return t.Name_
}

func (t *StructuredTask[T]) Clone() Task {
	return &StructuredTask[T]{
		Name_:   t.Name_,
		LLM:     t.LLM,
		Request: t.Request.Clone(),
	}
}

func (t *StructuredTask[T]) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	structuredLLM := structured.NewLLM(openai.NewOpenAISchemaGenerator().MustGenerateSchema((*T)(nil)), llm, structured.WithEvents(ai.NewJSONFileLogAgentEvents("pirates.json")))
	response, err := structuredLLM.Invoke(ctx, t.Request.Clone(ai.WithHistory(history)))
	if err != nil {
		return nil, err
	}

	lastMessage := response.Messages[len(response.Messages)-1]
	if lastMessage.Kind() != ai.MessageKindText {
		return nil, fmt.Errorf("last message is not a text message")
	}

	var result T
	err = json.Unmarshal([]byte(lastMessage.(*ai.TextMessage).Content), &result)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (t *StructuredTask[T]) With(opt StructuredTaskOpts[T]) *StructuredTask[T] {
	opt(t)
	return t
}

func (t *StructuredTask[T]) WithRequestOpts(opts ...ai.LLMRequestOpts) *StructuredTask[T] {
	new := t.Clone().(*StructuredTask[T])
	for _, opt := range opts {
		opt(new.Request)
	}
	return new
}

func (t *StructuredTask[T]) WithName(name string) Task {
	new := t.Clone().(*StructuredTask[T])
	new.Name_ = name
	return new
}
