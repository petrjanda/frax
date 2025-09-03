package eval

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petrjanda/frax/pkg/llm"

	_ "embed"
)

type Suite struct {
	Cases   []*Case
	Model   llm.LLM
	Request *llm.LLMRequest

	Total *SuiteResult
	Usage *llm.LLMUsage

	events SuiteEvents
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

func NewSuite(cases []*Case, model llm.LLM, req *llm.LLMRequest) *Suite {
	return &Suite{
		Cases:   cases,
		Model:   model,
		Request: req,

		Total: NewSuiteResult(),
		Usage: llm.NewLLMUsage(0, 0, 0),

		events: NewLoggingSuiteEvents(),
	}
}

func (s *Suite) Run(ctx context.Context) error {
	s.events.OnSuiteStart(s)

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
			s.Total.FatalResult()

			fmt.Printf("  - failed:\n    \033[31m%v\033[0m\n", err)
			fmt.Println("")
			continue
		}

		s.Usage.Add(response.Usage)

		lastMessage, ok := response.Messages[len(response.Messages)-1].(*llm.TextMessage)
		if !ok {
			s.Total.FatalResult()
			continue
		}

		result := q.Eval(ctx, s.events, lastMessage.Content)
		s.Total.Result(result)

		if result.Errors > 0 {
			var directive map[string]any
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

	s.events.OnSuiteEnd(s)

	return nil
}

func (s *SuiteResult) Result(result *CaseResult) {
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

func (s *SuiteResult) FatalResult() {
	s.CasesFatal += 1
}
