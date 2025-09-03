package eval

import (
	"fmt"

	_ "embed"
)

type Case struct {
	Input        string
	Expectations []Expectation
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

func (c *Case) Expect(expectation Expectation) *Case {
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

func (c *Case) Eval(actual string) *CaseResult {
	errors := []error{}

	fmt.Println("")
	fmt.Println("expectations")

	for _, expectation := range c.Expectations {
		if err := expectation.Eval(actual); err != nil {
			fmt.Printf("  — %v ... [\033[31mERR\033[0m]\n", err)
			errors = append(errors, err)
		} else {
			fmt.Printf("  — %v ... [\033[32mOK\033[0m]\n", expectation)
		}
	}

	fmt.Printf("  = total=%d, ok=%d, error=%d\n", len(c.Expectations), len(c.Expectations)-len(errors), len(errors))
	fmt.Println("")

	return &CaseResult{
		Total:  len(c.Expectations),
		Ok:     len(c.Expectations) - len(errors),
		Errors: len(errors),
	}
}
