// Package configservice is the compatibility import path for the target-oriented
// service adapter at github.com/faustbrian/go-config/v2/adapters/service.
package configservice

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/faustbrian/go-config/v2"
	direct "github.com/faustbrian/go-config/v2/adapters/service"
	"github.com/faustbrian/go-config/v2/dotenv"
	"github.com/faustbrian/go-config/v2/environment"
	"github.com/faustbrian/go-config/v2/validation"
	"github.com/faustbrian/go-service"
)

// ErrInvalidOptions identifies invalid loader construction.
var ErrInvalidOptions = direct.ErrInvalidOptions

// OptionsError identifies one invalid loader option without formatting its
// underlying cause.
type OptionsError struct {
	// Field identifies the rejected option.
	Field string
	// Reason describes the safe failure category.
	Reason string
	// Cause retains the construction failure for errors.Is and errors.As.
	Cause error
}

// Error returns a safe loader-construction diagnostic.
func (err *OptionsError) Error() string {
	return fmt.Sprintf("%s: %s: %v", err.Field, err.Reason, ErrInvalidOptions)
}

// Unwrap exposes both the option classification and construction cause.
func (err *OptionsError) Unwrap() []error {
	causes := []error{ErrInvalidOptions}
	if err.Cause != nil {
		causes = append(causes, err.Cause)
	}
	return causes
}

// Dotenv describes an explicitly local dotenv source. The caller retains
// ownership of FS; each load opens Path through the config dotenv adapter.
type Dotenv struct {
	// FS contains Path.
	FS fs.FS
	// Path identifies the dotenv document within FS.
	Path string
	// Options configure bounded parsing and typed mapping.
	Options dotenv.Options
}

// Options configure one immutable typed service loader.
type Options[T any] struct {
	// Sources contains caller-owned sources grouped by default precedence.
	Sources config.DefaultSources
	// Local explicitly permits Dotenv. Dotenv is rejected when Local is false.
	Local bool
	// Dotenv adds one local dotenv source when non-nil.
	Dotenv *Dotenv
	// Environment adds the process environment when non-nil.
	Environment *environment.Options
	// Validators run after complete typed decoding.
	Validators []validation.Validator[T]
}

// Loader is directly assignable to service.CommandSpec.Load. It owns no
// resource, performs no retries, and is safe to invoke repeatedly when its
// caller-provided sources are repeatable.
type Loader[T any] func(context.Context, service.Invocation) (T, error)

// New constructs a typed loader. Configuration is resolved only when the
// selected service command invokes the loader, before component construction.
func New[T any](options Options[T]) (Loader[T], error) {
	var directDotenv *direct.Dotenv
	if options.Dotenv != nil {
		directDotenv = &direct.Dotenv{
			FS: options.Dotenv.FS, Path: options.Dotenv.Path,
			Options: options.Dotenv.Options,
		}
	}
	loader, err := direct.New(direct.Options[T]{
		Sources: options.Sources, Local: options.Local, Dotenv: directDotenv,
		Environment: options.Environment, Validators: options.Validators,
	})
	if err != nil {
		optionsError := &direct.OptionsError{Cause: err}
		// Preserve a safe fallback if the direct adapter later adds an error type.
		_ = errors.As(err, &optionsError)
		return nil, &OptionsError{
			Field: optionsError.Field, Reason: optionsError.Reason,
			Cause: optionsError.Cause,
		}
	}
	return Loader[T](loader), nil
}
