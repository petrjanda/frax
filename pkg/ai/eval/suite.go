package eval

import (
	"context"

	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/ai/eval/expectations"
	"github.com/petrjanda/frax/pkg/ai/workflows"

	_ "embed"
)

type Suite struct {
	Cases  []*Case
	Tasks  []workflows.Task
	Usage  *ai.LLMUsage
	events SuiteEvents
}

type SuiteEvents interface {
	OnSuiteStart(suite *Suite)
	OnSuiteEnd(suite *Suite)
	OnSuiteError(error error)

	OnCaseStart(variant workflows.Task, case_ *Case)
	OnCaseEnd(variant workflows.Task, case_ *Case, errors []error)
	OnCaseError(variant workflows.Task, case_ *Case, error error)

	OnExpectationStart(variant workflows.Task, case_ *Case, expectation expectations.Expectation)
	OnExpectationEnd(variant workflows.Task, case_ *Case, expectation expectations.Expectation, err error)
	OnExpectationError(variant workflows.Task, case_ *Case, actual string, expectation expectations.Expectation, error error)
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

func NewSuite(events SuiteEvents, cases []*Case, variants []workflows.Task) *Suite {
	return &Suite{
		Cases:  cases,
		Tasks:  variants,
		Usage:  ai.NewLLMUsage(0, 0, 0),
		events: events,
	}
}

func (s *Suite) Run(ctx context.Context, llm ai.LLM) error {
	s.events.OnSuiteStart(s)
	for _, task := range s.Tasks {
		for _, q := range s.Cases {
			// Create conversation history with the travel request
			history := ai.NewHistory(
				ai.NewUserMessage(q.Input),
			)

			s.events.OnCaseStart(task, q)

			// Run the agent
			response, err := task.Invoke(ctx, llm, history)

			if err != nil {
				s.events.OnCaseError(task, q, err)
				continue
			}

			s.Usage.Add(response.Usage)

			lastMessage, ok := response.Messages[len(response.Messages)-1].(*ai.TextMessage)
			if !ok {
				continue
			}

			q.Eval(ctx, s.events, task, lastMessage.Content)
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
