package eval

import (
	"context"

	_ "embed"

	"github.com/petrjanda/frax/pkg/ai"
	"github.com/petrjanda/frax/pkg/eval/expectations"
)

type Case struct {
	Input        string
	Expectations []expectations.Expectation
	Alias        string
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

type CaseResult struct {
	Total  int
	Ok     int
	Errors int
}

func (c *Case) Eval(ctx context.Context, events SuiteEvents, variant *ai.LLMRequest, actual string) *CaseResult {
	errors := []error{}

	for _, expectation := range c.Expectations {
		events.OnExpectationStart(variant, c, expectation)
		if err := expectation.Eval(ctx, actual); err != nil {
			events.OnExpectationError(variant, c, actual, expectation, err)
			errors = append(errors, err)
		} else {
			events.OnExpectationEnd(variant, c, expectation, nil)
		}
	}

	events.OnCaseEnd(variant, c, errors)

	return &CaseResult{
		Total:  len(c.Expectations),
		Ok:     len(c.Expectations) - len(errors),
		Errors: len(errors),
	}
}
