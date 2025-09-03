package expectations

import (
	_ "embed"
)

type Expectation interface {
	Eval(actual string) error
	String() string
}
