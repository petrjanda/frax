package eval

import (
	"fmt"

	"github.com/petrjanda/frax/synq/eval/expectations"
)

type SuiteEvents interface {
	OnSuiteStart(suite *Suite)
	OnSuiteEnd(suite *Suite)
	OnSuiteError(error error)

	OnCaseStart(case_ *Case)
	OnCaseEnd(case_ *Case, errors []error)
	OnCaseError(case_ *Case, error error)

	OnExpectationStart(expectation expectations.Expectation)
	OnExpectationEnd(expectation expectations.Expectation, err error)
	OnExpectationError(expectation expectations.Expectation, error error)
}

type LoggingSuiteEvents struct{}

func NewLoggingSuiteEvents() SuiteEvents {
	return &LoggingSuiteEvents{}
}

func (e *LoggingSuiteEvents) OnSuiteStart(s *Suite) {
	fmt.Printf("")
	fmt.Printf("running suite with %d cases\n", len(s.Cases))
	fmt.Println("")
}

func (e *LoggingSuiteEvents) OnSuiteEnd(s *Suite) {
	fmt.Printf("  = total=%d, ok=%d, error=%d, fatal=%d | expectations total=%d, ok=%d, error=%d | score=%.1f\n",
		s.Total.CasesTotal, s.Total.CasesOk, s.Total.CasesErrors, s.Total.CasesFatal,
		s.Total.ExpectationsTotal, s.Total.ExpectationsOk, s.Total.ExpectationsErrors,
		float64(s.Total.ExpectationsOk)/float64(s.Total.ExpectationsTotal)*100,
	)

	fmt.Printf("  ~ usage input=%d, output=%d, total=%d\n",
		s.Usage.PromptTokens, s.Usage.CompletionTokens, s.Usage.TotalTokens,
	)
}

func (e *LoggingSuiteEvents) OnSuiteError(error error) {
	// @TODO
}

func (e *LoggingSuiteEvents) OnCaseStart(case_ *Case) {
	fmt.Printf("case '%s'\n", case_)
}

func (e *LoggingSuiteEvents) OnCaseEnd(case_ *Case, errors []error) {
	fmt.Printf("  = total=%d, ok=%d, error=%d\n", len(case_.Expectations), len(case_.Expectations)-len(errors), len(errors))
	fmt.Println("")

}

func (e *LoggingSuiteEvents) OnCaseError(case_ *Case, err error) {
	fmt.Printf("  — %v ... [\033[31mERR\033[0m]\n", err)
	fmt.Println("")
}

func (e *LoggingSuiteEvents) OnExpectationStart(expectation expectations.Expectation) {
	// @TODO
}

func (e *LoggingSuiteEvents) OnExpectationEnd(expectation expectations.Expectation, err error) {
	// fmt.Printf("  — %v ... [\033[32mOK\033[0m]\n", expectation)
}

func (e *LoggingSuiteEvents) OnExpectationError(expectation expectations.Expectation, err error) {
	// fmt.Printf("  — %v ... [\033[31mERR\033[0m]\n", expectation)
}
