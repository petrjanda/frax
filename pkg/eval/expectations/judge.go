package expectations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/llm"
)

type JudgeExpectation struct {
	Name        string
	Instruction string
	Model       string
	llm         llm.LLM

	judgement func(ctx context.Context, response *llm.TextMessage) error
}

func NewJudge[T any](name string, parent llm.LLM, model string, instruction string, judgement func(ctx context.Context, response *llm.TextMessage) error) *JudgeExpectation {
	directiveSchema := openai.NewOpenAISchemaGenerator().MustGenerateSchema((*T)(nil))

	return &JudgeExpectation{
		Name:        name,
		Instruction: instruction,
		Model:       model,
		llm: llm.NewBaseLLMWithStructuredOutput(
			directiveSchema, parent,
		),
		judgement: judgement,
	}
}

func (e *JudgeExpectation) Eval(ctx context.Context, actual string) error {
	req := llm.NewLLMRequest(
		llm.WithModel(e.Model),
		llm.WithSystem(e.Instruction),
		llm.WithHistory(llm.NewHistory(llm.NewUserMessage(actual))),
		llm.WithTemperature(0.0),
		llm.WithMaxCompletionTokens(1000),
	)

	response, err := e.llm.Invoke(ctx, req)
	if err != nil {
		return err
	}

	if len(response.Messages) == 0 {
		return fmt.Errorf("no response from judge")
	}

	lastMessage, ok := response.Messages[len(response.Messages)-1].(*llm.TextMessage)
	if !ok {
		return fmt.Errorf("no text message in response from judge")
	}

	if e.judgement == nil {
		return fmt.Errorf("no judgement function given for judge %s", e.Name)
	}

	return e.judgement(ctx, lastMessage)
}

func (e *JudgeExpectation) String() string {
	return e.Name
}

type ScoringJudgeExpectation struct {
	judge *JudgeExpectation
}

// SIMPLE SCORING

type ScoringJudgeVerdict struct {
	// Score that indicates to what degree you agree with the statement. 0 means you disagree completely, 100 means you agree completely.
	Score int `json:"score" jsonschema:"required,minimum=0,maximum=100"`

	// Reason for the score.
	Reason string `json:"reason" jsonschema:"required"`
}

func NewScoringJudge(name string, parent llm.LLM, model string, instruction string, threshold int) *ScoringJudgeExpectation {
	judgement := func(ctx context.Context, response *llm.TextMessage) error {
		var verdict ScoringJudgeVerdict
		if err := json.Unmarshal([]byte(response.Content), &verdict); err != nil {
			return err
		}

		if verdict.Score < 0 || verdict.Score > 100 {
			return fmt.Errorf("score must be between 0 and 100")
		}

		if verdict.Reason == "" {
			return fmt.Errorf("reason is required")
		}

		if verdict.Score < threshold {
			return fmt.Errorf("score must be greater than 50")
		}

		return nil
	}

	return &ScoringJudgeExpectation{judge: NewJudge[ScoringJudgeVerdict](name, parent, model, instruction, judgement)}
}

func (e *ScoringJudgeExpectation) Eval(ctx context.Context, actual string) error {
	return e.judge.Eval(ctx, actual)
}

func (e *ScoringJudgeExpectation) String() string {
	return e.judge.String()
}
