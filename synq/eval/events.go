package eval

import (
	"fmt"

	"github.com/petrjanda/frax/pkg/llm"
	"github.com/petrjanda/frax/synq/eval/expectations"
)

type SuiteEvents interface {
	OnSuiteStart(suite *Suite)
	OnSuiteEnd(suite *Suite)
	OnSuiteError(error error)

	OnCaseStart(variant *llm.LLMRequest, case_ *Case)
	OnCaseEnd(variant *llm.LLMRequest, case_ *Case, errors []error)
	OnCaseError(variant *llm.LLMRequest, case_ *Case, error error)

	OnExpectationStart(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation)
	OnExpectationEnd(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation, err error)
	OnExpectationError(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation, error error)
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

func (e *LoggingSuiteEvents) OnCaseStart(variant *llm.LLMRequest, case_ *Case) {
	// fmt.Printf("%s | case '%s'\n", variant.Model, case_)
}

func (e *LoggingSuiteEvents) OnCaseEnd(variant *llm.LLMRequest, case_ *Case, errors []error) {
	fmt.Printf("%s | case '%s' = total=%d, ok=%d, error=%d\n", variant.Model, case_, len(case_.Expectations), len(case_.Expectations)-len(errors), len(errors))
	// fmt.Println("")

}

func (e *LoggingSuiteEvents) OnCaseError(variant *llm.LLMRequest, case_ *Case, err error) {
	fmt.Printf("%s |  — %v ... [\033[31mERR\033[0m]\n", variant.Model, err)
	// fmt.Println("")
}

func (e *LoggingSuiteEvents) OnExpectationStart(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation) {
	// @TODO
}

func (e *LoggingSuiteEvents) OnExpectationEnd(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation, err error) {
	// fmt.Printf("  — %v ... [\033[32mOK\033[0m]\n", expectation)
}

func (e *LoggingSuiteEvents) OnExpectationError(variant *llm.LLMRequest, case_ *Case, expectation expectations.Expectation, err error) {
	// fmt.Printf("  — %v ... [\033[31mERR\033[0m]\n", expectation)
}
