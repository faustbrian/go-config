package config_test

import (
	"context"
	"testing"

	config "github.com/faustbrian/go-config"
	"github.com/faustbrian/go-config/configtest"
)

func TestLoadPropagatesSecretAndDeprecatedFieldMetadataToOrigins(t *testing.T) {
	t.Parallel()

	type credentials struct {
		Token string `config:"token"`
	}
	type settings struct {
		Legacy      string       `config:"legacy,deprecated"`
		Credentials *credentials `config:"credentials,secret"`
		Implicit    string
		Ignored     string `config:"-"`
		private     string
	}
	source := configtest.NewSource(
		config.SourceInfo{Name: "document"},
		config.Document{Tree: map[string]any{
			"legacy": "value",
			"credentials": map[string]any{
				"token": "canary-secret-value",
			},
			"implicit": "value",
		}},
	)
	plan := configtest.MustPlan(t, source)
	snapshot := configtest.MustLoad[settings](t, context.Background(), plan)

	legacy, ok := snapshot.Origin("legacy")
	if !ok || !legacy.Deprecated || legacy.Sensitive {
		t.Fatalf("legacy origin = %#v, %v", legacy, ok)
	}
	for _, path := range []string{"credentials", "credentials.token"} {
		origin, ok := snapshot.Origin(path)
		if !ok || !origin.Sensitive || origin.Deprecated {
			t.Fatalf("Origin(%q) = %#v, %v", path, origin, ok)
		}
	}
	implicit, ok := snapshot.Origin("implicit")
	if !ok || implicit.Sensitive || implicit.Deprecated {
		t.Fatalf("implicit origin = %#v, %v", implicit, ok)
	}
}
