package expectations

import (
	"fmt"

	"github.com/tidwall/gjson"

	_ "embed"
)

type JSONExpectation struct {
	Path     string
	Expected Assertion
}

type Assertion interface {
	Validate(value any) error
}

func JSONPath(path string) *JSONExpectation {
	return &JSONExpectation{
		Path: path,
	}
}

func (e *JSONExpectation) Eval(actual string) error {
	if e.Expected == nil {
		return fmt.Errorf("no assertion given for path %s", e.Path)
	}

	value := gjson.Get(actual, e.Path)
	err := e.Expected.Validate(value.Value())

	if err != nil {
		return fmt.Errorf("path %s %v", e.Path, err)
	}

	return nil
}

func (e *JSONExpectation) String() string {
	return fmt.Sprintf("%s %v", e.Path, e.Expected)
}

func (e *JSONExpectation) Eq(right any) *JSONExpectation {
	e.Expected = &Eq_{Right: right}
	return e
}

func (e *JSONExpectation) Exists() *JSONExpectation {
	e.Expected = &Exists_{}
	return e
}

func (e *JSONExpectation) DoesNotExist() *JSONExpectation {
	e.Expected = &DoesNotExist_{}
	return e
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
	return fmt.Sprintf("equals %v", e.Right)
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

var DoesNotExist = &DoesNotExist_{}

type DoesNotExist_ struct {
}

func (e *DoesNotExist_) Validate(value any) error {
	if value == nil {
		return nil
	}

	return fmt.Errorf("expected nil value, got '%v'", value)
}

func (e *DoesNotExist_) String() string {
	return "does not exist"
}
