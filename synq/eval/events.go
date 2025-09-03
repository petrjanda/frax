package eval

import (
	"fmt"
)

type SuiteEvents interface {
	OnSuiteStart(suite *Suite)
	OnSuiteEnd(suite *Suite)
	OnCaseStart(case_ *Case)
	OnCaseEnd(case_ *Case, errors []error)
	OnCaseError(case_ *Case, error error)
	OnExpectationStart(expectation Expectation)
	OnExpectationEnd(expectation Expectation, err error)
	OnExpectationError(expectation Expectation, error error)
	OnSuiteError(error error)
}

type LoggingSuiteEvents struct {
}

func NewLoggingSuiteEvents() SuiteEvents {
	return &LoggingSuiteEvents{}
}

func (e *LoggingSuiteEvents) OnSuiteStart(s *Suite) {
	fmt.Printf("")
	fmt.Printf("running suite with %d cases\n", len(s.Cases))
	fmt.Println("")
}

func (e *LoggingSuiteEvents) OnSuiteEnd(s *Suite) {
	fmt.Printf("--------------------------------\n")
	fmt.Printf(" = total=%d, ok=%d, error=%d, fatal=%d | expectations total=%d, ok=%d, error=%d | score=%.1f\n",
		s.Total.CasesTotal, s.Total.CasesOk, s.Total.CasesErrors, s.Total.CasesFatal,
		s.Total.ExpectationsTotal, s.Total.ExpectationsOk, s.Total.ExpectationsErrors,
		float64(s.Total.ExpectationsOk)/float64(s.Total.ExpectationsTotal)*100,
	)

	fmt.Printf(" ~ usage input=%d, output=%d, total=%d\n",
		s.Usage.PromptTokens, s.Usage.CompletionTokens, s.Usage.TotalTokens,
	)
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

func (e *LoggingSuiteEvents) OnExpectationStart(expectation Expectation) {
	// @TODO
}

func (e *LoggingSuiteEvents) OnExpectationEnd(expectation Expectation, err error) {
	fmt.Printf("  — %v ... [\033[32mOK\033[0m]\n", expectation)
}

func (e *LoggingSuiteEvents) OnExpectationError(expectation Expectation, err error) {
	fmt.Printf("  — %v ... [\033[31mERR\033[0m]\n", expectation)
}

func (e *LoggingSuiteEvents) OnSuiteError(error error) {
	// @TODO
}
