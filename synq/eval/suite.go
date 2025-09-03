package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/petrjanda/frax/pkg/llm"

	_ "embed"
)

type Suite[T any] struct {
	Cases   []*Case
	Model   llm.LLM
	Request *llm.LLMRequest

	Total *SuiteResult
}

type SuiteResult struct {
	ExpectationsTotal  int
	ExpectationsOk     int
	ExpectationsErrors int

	CasesTotal  int
	CasesOk     int
	CasesErrors int
}

func NewSuiteResult() *SuiteResult {
	return &SuiteResult{
		ExpectationsTotal:  0,
		ExpectationsOk:     0,
		ExpectationsErrors: 0,

		CasesTotal:  0,
		CasesOk:     0,
		CasesErrors: 0,
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

func NewSuite[T any](cases []*Case, model llm.LLM, req *llm.LLMRequest) *Suite[T] {
	return &Suite[T]{
		Cases:   cases,
		Model:   model,
		Request: req,

		Total: NewSuiteResult(),
	}
}

func (s *Suite[T]) Run(ctx context.Context) error {
	for _, q := range s.Cases {
		// Create conversation history with the travel request
		history := llm.NewHistory(
			llm.NewUserMessage(q.Input),
		)

		// Run the agent
		response, err := s.Model.Invoke(ctx,
			s.Request.Clone(llm.WithHistory(history)),
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

		s.Total.Add(q.Validate([]byte(lastMessage.Content)))
	}

	fmt.Printf("--------------------------------\n")
	fmt.Printf(" = total=%d, ok=%d, error=%d\n", s.Total.CasesTotal, s.Total.CasesOk, s.Total.CasesErrors)

	return nil
}
