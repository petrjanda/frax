package eval

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"

	_ "embed"
)

type Expectation struct {
	Path     string
	Expected Assertion
}

func Expect(path string, assertion Assertion) *Expectation {
	return &Expectation{
		Path:     path,
		Expected: assertion,
	}
}

func (e *Expectation) Validate(actual json.RawMessage) error {
	value := gjson.Get(string(actual), e.Path)
	if value.Exists() {
		if e.Expected != nil {
			err := e.Expected.Validate(value.Value())

			if err != nil {
				return fmt.Errorf("path '%s' %v", e.Path, err)
			}

			return nil
		}
		return nil
	}

	return fmt.Errorf("path '%s' not found", e.Path)
}

func (e *Expectation) String() string {
	return fmt.Sprintf("%s%v", e.Path, e.Expected)
}

type Assertion interface {
	Validate(value any) error
}

// EQUALS

type Eq_ struct {
	Right any
}

func (e *Eq_) Validate(left any) error {
	if left == e.Right {
		return nil
	}

	return fmt.Errorf("expected '%v', got '%v'", e.Right, left)
}

func (e *Eq_) String() string {
	return fmt.Sprintf("=%v", e.Right)
}

func Eq(right any) *Eq_ {
	return &Eq_{Right: right}
}

// EXISTS

type Exists_ struct {
}

func (e *Exists_) Validate(value any) error {
	if value == nil {
		return fmt.Errorf("expected value, got nil")
	}

	return nil
}

func (e *Exists_) String() string {
	return "exists"
}

var Exists = &Exists_{}
