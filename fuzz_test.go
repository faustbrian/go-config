package config_test

import (
	"context"
	"reflect"
	"testing"

	config "github.com/faustbrian/go-config"
	"github.com/faustbrian/go-config/decode"
	"github.com/faustbrian/go-config/dotenv"
	"github.com/faustbrian/go-config/environment"
	jsonsource "github.com/faustbrian/go-config/json"
	tomlsource "github.com/faustbrian/go-config/toml"
	yamlsource "github.com/faustbrian/go-config/yaml"
)

func FuzzStructuredSources(f *testing.F) {
	f.Add([]byte(`{"name":"worker","port":8080}`))
	f.Add([]byte("name: worker\nport: 8080\n"))
	f.Add([]byte("name = \"worker\"\nport = 8080\n"))
	f.Add([]byte("{\x00\xff"))

	f.Fuzz(func(t *testing.T, data []byte) {
		sources := []struct {
			build func([]byte) (config.Source, error)
		}{
			{build: func(data []byte) (config.Source, error) {
				return jsonsource.Bytes(data, jsonsource.Options{Name: "fuzz-json"})
			}},
			{build: func(data []byte) (config.Source, error) {
				return yamlsource.Bytes(data, yamlsource.Options{Name: "fuzz-yaml"})
			}},
			{build: func(data []byte) (config.Source, error) {
				return tomlsource.Bytes(data, tomlsource.Options{Name: "fuzz-toml"})
			}},
		}
		for _, fixture := range sources {
			source, err := fixture.build(data)
			if err != nil {
				t.Fatalf("construct source: %v", err)
			}
			_, _ = source.Load(context.Background())
		}
	})
}

func FuzzDotenvInterpolation(f *testing.F) {
	f.Add("APP_NAME=${NAME:-worker}\n", "worker")
	f.Add("APP_NAME=${APP_NAME}\n", "")
	f.Add("APP_NAME='${NAME}'\n", "literal")
	f.Add("\x00=${MISSING}\n", "value")

	type settings struct {
		Name string `config:"name"`
	}
	f.Fuzz(func(t *testing.T, contents, external string) {
		source, err := dotenv.BytesFor[settings]([]byte(contents), dotenv.Options{
			Name: "fuzz-dotenv", Prefix: "APP_",
			Limits: dotenv.Limits{
				MaxBytes: 4096, MaxLines: 64, MaxLineBytes: 1024, MaxKeys: 64,
			},
			Interpolation: &dotenv.Interpolation{
				Variables: map[string]string{"NAME": external}, IncludeFile: true,
				MaxDepth: 8, MaxExpandedBytes: 4096,
			},
		})
		if err != nil {
			t.Fatalf("construct source: %v", err)
		}
		_, _ = source.Load(context.Background())
	})
}

func FuzzEnvironmentMapping(f *testing.F) {
	f.Add("APP_VALUE", "42")
	f.Add("app_value", "true")
	f.Add("APP_VALUE", "[1,2,3]")
	f.Add("BAD-NAME", "secret")

	type settings struct {
		Value int `config:"value"`
	}
	f.Fuzz(func(t *testing.T, name, value string) {
		if len(name)+len(value) > 4096 {
			t.Skip()
		}
		source, err := environment.EnvironFor[settings](
			[]string{name + "=" + value},
			environment.Options{
				Name: "fuzz-environment", Prefix: "APP_",
				Limits: environment.Limits{
					MaxVariables: 4, MaxBytes: 4096, MaxValueBytes: 2048,
				},
			},
		)
		if err != nil {
			t.Fatalf("construct source: %v", err)
		}
		_, _ = source.Load(context.Background())
	})
}

func FuzzDecodeTagsAndDestinationTypes(f *testing.F) {
	f.Add("value", uint8(0), "42")
	f.Add("value,required", uint8(1), "true")
	f.Add("-", uint8(2), "ignored")
	f.Add(",secret", uint8(3), "value")

	f.Fuzz(func(t *testing.T, tag string, selector uint8, input string) {
		if len(tag)+len(input) > 2048 {
			t.Skip()
		}
		types := []reflect.Type{
			reflect.TypeFor[int](),
			reflect.TypeFor[bool](),
			reflect.TypeFor[[]string](),
			reflect.TypeFor[map[string]string](),
		}
		typeOf := reflect.StructOf([]reflect.StructField{{
			Name: "Value", Type: types[int(selector)%len(types)],
			Tag: reflect.StructTag(`config:"` + tag + `"`),
		}})
		destination := reflect.New(typeOf).Interface()
		_ = decode.Value(map[string]any{"value": input}, destination)
	})
}
