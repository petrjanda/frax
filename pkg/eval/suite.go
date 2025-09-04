package eval

import (
	"context"

	"github.com/petrjanda/frax/pkg/llm"

	"github.com/petrjanda/frax/pkg/eval/expectations"

	_ "embed"
)

type Suite struct {
	Cases    []*Case
	Model    llm.LLM
	Variants []*llm.LLMRequest

	Usage *llm.LLMUsage

	events SuiteEvents
}

type SuiteEvents interface {
	OnSuiteStart(suite *Suite)
	OnSuiteEnd(suite *Suite)
	OnSuiteError(error error)

	OnCaseStart(variant *llm.LLMRequest, case_ *Case)
	OnCaseEnd(variant *llm.LLMRequest, case_ *Case, errors []error)
	OnCaseError(variant *llm.LLMRequest, case_ *Case, error error)

	OnExpectationStart(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation)
	OnExpectationEnd(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation, err error)
	OnExpectationError(variant *llm.LLMRequest, case_ *Case, actual string, expectation expectations.Expectation, error error)
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

func NewSuite(events SuiteEvents, cases []*Case, model llm.LLM, req []*llm.LLMRequest) *Suite {
	return &Suite{
		Cases:    cases,
		Model:    model,
		Variants: req,

		Usage: llm.NewLLMUsage(0, 0, 0),

		events: events,
	}
}

func (s *Suite) Run(ctx context.Context) error {
	s.events.OnSuiteStart(s)
	for _, variant := range s.Variants {
		for _, q := range s.Cases {
			// Create conversation history with the travel request
			history := llm.NewHistory(
				llm.NewUserMessage(q.Input),
			)

			variant = variant.Clone(llm.WithHistory(history))

			s.events.OnCaseStart(variant, q)

			// Run the agent
			response, err := s.Model.Invoke(ctx, variant)

			if err != nil {
				s.events.OnCaseError(variant, q, err)
				continue
			}

			s.Usage.Add(response.Usage)

			lastMessage, ok := response.Messages[len(response.Messages)-1].(*llm.TextMessage)
			if !ok {
				continue
			}

			q.Eval(ctx, s.events, variant, lastMessage.Content)
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
