package eval

import (
	"fmt"

	"github.com/tidwall/gjson"

	_ "embed"
)

type JSONExpectation struct {
	Path     string
	Expected Assertion
}

func JSON(path string, assertion Assertion) *JSONExpectation {
	return &JSONExpectation{
		Path:     path,
		Expected: assertion,
	}
}

func (e *JSONExpectation) Eval(actual string) error {
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
	return fmt.Sprintf("equals %v", e.Right)
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
