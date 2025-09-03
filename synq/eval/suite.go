package eval

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petrjanda/frax/pkg/llm"

	_ "embed"
)

type Suite[T any] struct {
	Cases   []*Case
	Model   llm.LLM
	Request *llm.LLMRequest

	Total *SuiteResult
	Usage *llm.LLMUsage
}

type SuiteResult struct {
	ExpectationsTotal  int
	ExpectationsOk     int
	ExpectationsErrors int

	CasesTotal  int
	CasesOk     int
	CasesErrors int
	CasesFatal  int
}

func NewSuiteResult() *SuiteResult {
	return &SuiteResult{
		ExpectationsTotal:  0,
		ExpectationsOk:     0,
		ExpectationsErrors: 0,

		CasesTotal:  0,
		CasesOk:     0,
		CasesErrors: 0,
		CasesFatal:  0,
	}
}

func (s *SuiteResult) Add(result *CaseResult) {
	s.CasesTotal += 1

	if result.Errors > 0 {
		s.CasesErrors += 1
	} else {
		s.CasesOk += 1
	}

	s.ExpectationsTotal += result.Total
	s.ExpectationsOk += result.Ok
	s.ExpectationsErrors += result.Errors
}

func (s *SuiteResult) AddFatal() {
	s.CasesFatal += 1
}

func NewSuite[T any](cases []*Case, model llm.LLM, req *llm.LLMRequest) *Suite[T] {
	return &Suite[T]{
		Cases:   cases,
		Model:   model,
		Request: req,

		Total: NewSuiteResult(),
		Usage: llm.NewLLMUsage(0, 0, 0),
	}
}

func (s *Suite[T]) Run(ctx context.Context) error {

	fmt.Printf("")
	fmt.Printf("running suite with %d cases\n", len(s.Cases))
	fmt.Println("")

	for _, q := range s.Cases {
		// Create conversation history with the travel request
		history := llm.NewHistory(
			llm.NewUserMessage(q.Input),
		)

		fmt.Printf("case %s\n", q)

		// Run the agent
		response, err := s.Model.Invoke(ctx,
			s.Request.Clone(llm.WithHistory(history)),
		)

		if err != nil {
			s.Total.AddFatal()

			fmt.Printf("  - failed:\n    \033[31m%v\033[0m\n", err)
			fmt.Println("")
			continue
		}

		s.Usage.Add(response.Usage)

		lastMessage, ok := response.Messages[len(response.Messages)-1].(*llm.TextMessage)
		if !ok {
			fmt.Printf("last message is not a text message: %v", response.Messages[len(response.Messages)-1])
		}

		result := q.Eval(lastMessage.Content)
		s.Total.Add(result)

		if result.Errors > 0 {
			var directive T
			if err = json.Unmarshal([]byte(lastMessage.Content), &directive); err != nil {
				fmt.Printf("failed unmarshalling: %v", lastMessage.Content)
			}

			if payload, err := json.MarshalIndent(directive, "", "  "); err != nil {
				fmt.Printf("failed marshalling: %v", err)
			} else {
				fmt.Println("actual output was ```")
				fmt.Println(string(payload))
				fmt.Println("```")
			}
		}
	}

	fmt.Printf("--------------------------------\n")
	fmt.Printf(" = total=%d, ok=%d, error=%d, fatal=%d | expectations total=%d, ok=%d, error=%d | score=%.1f\n",
		s.Total.CasesTotal, s.Total.CasesOk, s.Total.CasesErrors, s.Total.CasesFatal,
		s.Total.ExpectationsTotal, s.Total.ExpectationsOk, s.Total.ExpectationsErrors,
		float64(s.Total.ExpectationsOk)/float64(s.Total.ExpectationsTotal)*100,
	)

	fmt.Printf(" ~ usage input=%d, output=%d, total=%d\n",
		s.Usage.PromptTokens, s.Usage.CompletionTokens, s.Usage.TotalTokens,
	)

	return nil
}
