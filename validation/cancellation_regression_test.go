package validation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-config/v2/validation"
)

type cancelingSelf struct{ cancel context.CancelFunc }

func (v cancelingSelf) Validate() error { v.cancel(); return nil }
func TestRunRechecksCancellationAfterFinalCallback(t *testing.T) {
	for _, self := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		var err error
		if self {
			err = validation.Run(ctx, cancelingSelf{cancel})
		} else {
			err = validation.Run(ctx, "value", func(context.Context, string) error { cancel(); return nil })
		}
		cancel()
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("self=%t returned success after cancellation: %v", self, err)
		}
	}
}
