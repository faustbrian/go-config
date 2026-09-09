package configservice_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	direct "github.com/faustbrian/go-config/adapters/service"
	"github.com/faustbrian/go-config/configservice"
	"github.com/faustbrian/go-config/dotenv"
	"github.com/faustbrian/go-config/programmatic"
	"github.com/faustbrian/go-service"
)

type settings struct {
	Port int `config:"port"`
}

func TestCompatibilityFacadePreservesLoaderAndErrorIdentity(t *testing.T) {
	defaults, err := programmatic.Defaults("defaults", map[string]any{"port": int64(8080)})
	if err != nil {
		t.Fatal(err)
	}
	options := configservice.Options[settings]{}
	options.Sources.Defaults = append(options.Sources.Defaults, defaults)
	loader, err := configservice.New(options)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	loaded, err := loader(context.Background(), service.Invocation{})
	if err != nil || loaded.Port != 8080 {
		t.Fatalf("loader() = %#v, %v", loaded, err)
	}
	if !errors.Is(configservice.ErrInvalidOptions, direct.ErrInvalidOptions) ||
		!errors.Is(direct.ErrInvalidOptions, configservice.ErrInvalidOptions) {
		t.Fatal("ErrInvalidOptions identity differs")
	}
	_, invalid := configservice.New(configservice.Options[settings]{
		Dotenv: &configservice.Dotenv{
			FS: fstest.MapFS{}, Path: ".env",
			Options: dotenv.Options{Name: "dotenv"},
		},
	})
	if !errors.Is(invalid, configservice.ErrInvalidOptions) {
		t.Fatalf("New(invalid) error = %v, want ErrInvalidOptions", invalid)
	}
	var legacyError *configservice.OptionsError
	if !errors.As(invalid, &legacyError) || legacyError.Field != "Dotenv" {
		t.Fatalf("New(invalid) error = %#v, want compatibility OptionsError", invalid)
	}
	var directError *direct.OptionsError
	if errors.As(invalid, &directError) {
		t.Fatalf("New(invalid) exposed direct OptionsError: %#v", invalid)
	}
	if legacyError.Error() == "" || len(legacyError.Unwrap()) != 1 {
		t.Fatalf("OptionsError contract = %#v", legacyError)
	}
	cause := errors.New("construction failure")
	withCause := (&configservice.OptionsError{Cause: cause}).Unwrap()
	if len(withCause) != 2 || !errors.Is(withCause[1], cause) {
		t.Fatalf("Unwrap(with cause) = %#v", withCause)
	}
}
