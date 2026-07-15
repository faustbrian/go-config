package config_test

import (
	"context"
	"reflect"
	"testing"

	jsonsource "github.com/faustbrian/go-config/json"
	tomlsource "github.com/faustbrian/go-config/toml"
	yamlsource "github.com/faustbrian/go-config/yaml"
)

func TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree(t *testing.T) {
	t.Parallel()

	sources := map[string]sourceLoader{
		"json": func() (map[string]any, error) {
			source, err := jsonsource.Bytes(
				[]byte(`{"name":"api","port":8080,"enabled":true,"items":["one","two"],"nested":{"host":"localhost"}}`),
				jsonsource.Options{Name: "json"},
			)
			if err != nil {
				return nil, err
			}
			document, err := source.Load(context.Background())
			return document.Tree, err
		},
		"yaml": func() (map[string]any, error) {
			source, err := yamlsource.Bytes(
				[]byte("name: api\nport: 8080\nenabled: true\nitems: [one, two]\nnested:\n  host: localhost\n"),
				yamlsource.Options{Name: "yaml"},
			)
			if err != nil {
				return nil, err
			}
			document, err := source.Load(context.Background())
			return document.Tree, err
		},
		"toml": func() (map[string]any, error) {
			source, err := tomlsource.Bytes(
				[]byte("name = \"api\"\nport = 8080\nenabled = true\nitems = [\"one\", \"two\"]\n[nested]\nhost = \"localhost\"\n"),
				tomlsource.Options{Name: "toml"},
			)
			if err != nil {
				return nil, err
			}
			document, err := source.Load(context.Background())
			return document.Tree, err
		},
	}

	var reference map[string]any
	for name, load := range sources {
		tree, err := load()
		if err != nil {
			t.Fatalf("%s load error = %v", name, err)
		}
		if reference == nil {
			reference = tree
			continue
		}
		if !reflect.DeepEqual(tree, reference) {
			t.Fatalf("%s tree = %#v, want %#v", name, tree, reference)
		}
	}
}

type sourceLoader func() (map[string]any, error)
