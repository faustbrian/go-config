package config_test

import (
	"context"
	stdjson "encoding/json"
	"fmt"
	"testing"

	config "github.com/faustbrian/go-config/v3"
	"github.com/faustbrian/go-config/v3/configtest"
	"github.com/faustbrian/go-config/v3/decode"
	jsonsource "github.com/faustbrian/go-config/v3/json"
	"github.com/faustbrian/go-config/v3/merge"
	"github.com/faustbrian/go-config/v3/validation"
)

type benchmarkSettings struct {
	Name   string         `config:"name"`
	Port   int            `config:"port"`
	Labels map[string]int `config:"labels"`
}

// BenchmarkSourcePlanLoad measures repeated typed loading from one in-memory source;
// construction is outside b.Loop.
func BenchmarkSourcePlanLoad(b *testing.B) {
	source := configtest.NewSource(
		config.SourceInfo{Name: "benchmark", Priority: 10},
		config.Document{Tree: map[string]any{"name": "worker", "port": int64(8080)}},
	)
	plan := configtest.MustPlan(b, source)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := config.Load[benchmarkSettings](context.Background(), plan); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecode measures strict decoding of a fixed scalar/two-label tree into a
// fresh candidate.
func BenchmarkDecode(b *testing.B) {
	tree := map[string]any{
		"name": "worker", "port": int64(8080),
		"labels": map[string]any{"one": int64(1), "two": int64(2)},
	}
	b.ReportAllocs()
	for b.Loop() {
		var destination benchmarkSettings
		if err := decode.Into(tree, &destination); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMerge measures copying/merging two fixed nested layers; no source I/O is included.
func BenchmarkMerge(b *testing.B) {
	lower := map[string]any{"server": map[string]any{"host": "localhost", "port": int64(80)}}
	upper := map[string]any{"server": map[string]any{"port": int64(443)}, "debug": true}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := merge.Trees(lower, upper); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidation measures one successful callback on a fixed value without
// source/decode work.
func BenchmarkValidation(b *testing.B) {
	validator := func(context.Context, benchmarkSettings) error { return nil }
	value := benchmarkSettings{Name: "worker", Port: 8080}
	b.ReportAllocs()
	for b.Loop() {
		if err := validation.Run(context.Background(), value, validator); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLargeConfig measures repeated parsing of a fixed 10000-key JSON document;
// generation/source construction is outside b.Loop.
func BenchmarkLargeConfig(b *testing.B) {
	large := make(map[string]any, 10_000)
	for index := range 10_000 {
		large[fmt.Sprintf("key_%05d", index)] = index
	}
	data, err := stdjson.Marshal(large)
	if err != nil {
		b.Fatal(err)
	}
	source, err := jsonsource.Bytes(data, jsonsource.Options{
		Name: "large", Limits: jsonsource.Limits{MaxBytes: int64(len(data) + 1)},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := source.Load(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}
