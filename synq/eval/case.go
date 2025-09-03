package eval

import (
	"context"

	_ "embed"

	"github.com/petrjanda/frax/synq/eval/expectations"
)

type Case struct {
	Input        string
	Expectations []expectations.Expectation
	Alias        string
}

type CaseResult struct {
	Total  int
	Ok     int
	Errors int
}

func NewCase(input string, opts ...CaseOption) *Case {
	c := &Case{
		Input: input,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type CaseOption func(*Case)

func WithAlias(alias string) CaseOption {
	return func(c *Case) {
		c.Alias = alias
	}
}

func (c *Case) Expect(expectation expectations.Expectation) *Case {
	c.Expectations = append(c.Expectations, expectation)
	return c
}

func (c *Case) String() string {
	alias := c.Input
	if c.Alias != "" {
		alias = c.Alias
	}

	return alias
}

func (c *Case) Eval(ctx context.Context, events SuiteEvents, actual string) *CaseResult {
	errors := []error{}
	events.OnCaseStart(c)

	for _, expectation := range c.Expectations {
		if err := expectation.Eval(actual); err != nil {
			events.OnExpectationError(expectation, err)
			errors = append(errors, err)
		} else {
			events.OnExpectationEnd(expectation, nil)
		}
	}

	events.OnCaseEnd(c, errors)

	return &CaseResult{
		Total:  len(c.Expectations),
		Ok:     len(c.Expectations) - len(errors),
		Errors: len(errors),
	}
}
