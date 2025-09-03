package eval

import (
	"encoding/json"
	"fmt"

	_ "embed"
)

type Case struct {
	Input        string
	Expectations []*Expectation
	Alias        string
}

func NewCase(input string, expectations []*Expectation, opts ...CaseOption) *Case {
	c := &Case{
		Input:        input,
		Expectations: expectations,
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

func (c *Case) Validate(actual json.RawMessage) []error {
	errors := []error{}

	alias := c.Input
	if c.Alias != "" {
		alias = c.Alias
	}

	fmt.Printf("running '%s'\n", alias)

	for _, expectation := range c.Expectations {
		if err := expectation.Validate(actual); err != nil {
			fmt.Printf("  — %v ... [\033[31mERR\033[0m]\n", err)
			errors = append(errors, err)
		} else {
			fmt.Printf("  — %v ... [\033[32mOK\033[0m]\n", expectation)
		}
	}

	fmt.Printf("  = total=%d, ok=%d, error=%d\n", len(c.Expectations), len(c.Expectations)-len(errors), len(errors))

	return errors
}
