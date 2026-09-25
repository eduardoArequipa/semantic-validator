package providers

import (
	"context"
	"errors"
)

type Result struct {
	Value      bool
	Confidence float64
}

var ErrTimeout = errors.New("semantic provider timeout")

type Provider interface {
	Evaluate(ctx context.Context, state, question string) (Result, error)
}
