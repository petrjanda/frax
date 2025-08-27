package llm

import (
	"context"
)

// Task represents a task that can be executed with input and produces output
type Task[I any, O any] interface {
	Exec(ctx context.Context, input I) (O, error)
}

// Eval represents an evaluator that can evaluate output and produce a score
type Eval[O any] interface {
	Eval(ctx context.Context, input O) (float64, error)
}
