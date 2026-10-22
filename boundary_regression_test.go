package config_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"
	"testing"
	"time"

	config "github.com/faustbrian/go-config/v3"
	"github.com/faustbrian/go-config/v3/merge"
)

type boundaryArrays struct {
	Items [1]map[string][]string `config:"items"`
}

func (v *boundaryArrays) UnmarshalConfigValue(any) error {
	v.Items = [1]map[string][]string{{"key": {"original"}}}
	return nil
}

func TestSnapshotClonesMutableArrayReferences(t *testing.T) {
	plan, err := config.NewPlan(source{info: config.SourceInfo{Name: "arrays"}, tree: map[string]any{"value": "ignored"}})
	if err != nil {
		t.Fatal(err)
	}
	type settings struct {
		Value boundaryArrays `config:"value"`
	}
	snapshot, err := config.LoadWithValidators(context.Background(), plan, func(_ context.Context, value settings) error {
		value.Value.Items[0]["key"][0] = "validator"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	first := snapshot.Value()
	if first.Value.Items[0]["key"][0] != "original" {
		t.Fatal("validator escaped candidate copy through array")
	}
	first.Value.Items[0]["key"][0] = "reader"
	if snapshot.Value().Value.Items[0]["key"][0] != "original" {
		t.Fatal("reader escaped snapshot copy through array")
	}
	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() {
			for range 8 {
				copy := snapshot.Value()
				copy.Value.Items[0]["key"][0] = "reader"
			}
		})
	}
	readers.Wait()
	if snapshot.Value().Value.Items[0]["key"][0] != "original" {
		t.Fatal("concurrent readers escaped snapshot copy")
	}
}

type boundaryPrivatePayload struct{ hidden map[string]string }

func (v *boundaryPrivatePayload) UnmarshalConfigValue(any) error {
	v.hidden = map[string]string{"key": "value"}
	return nil
}
func TestOptionalPayloadIsAdmittedBeforeValidation(t *testing.T) {
	type settings struct {
		Value config.Optional[boundaryPrivatePayload] `config:"value"`
	}
	plan, err := config.NewPlan(source{info: config.SourceInfo{Name: "optional"}, tree: map[string]any{"value": "ignored"}})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	snapshot, err := config.LoadWithValidators(context.Background(), plan, func(context.Context, settings) error { called = true; return nil })
	var rejected *config.SnapshotValueError
	if snapshot != nil || !errors.As(err, &rejected) || called {
		t.Fatalf("unsupported Optional published or validated: snapshot=%t rejected=%t callback=%t", snapshot != nil, errors.As(err, &rejected), called)
	}
}

type boundaryCyclePayload struct{ Dynamic any }

func (v *boundaryCyclePayload) UnmarshalConfigValue(any) error {
	m := map[string]any{}
	m["self"] = m
	v.Dynamic = m
	return nil
}
func TestOptionalRuntimeCycleIsRejected(t *testing.T) {
	if os.Getenv("CONFIG_OPTIONAL_CYCLE_PROBE") == "1" {
		debug.SetMaxStack(1 << 20)
		type settings struct {
			Value config.Optional[boundaryCyclePayload] `config:"value"`
		}
		plan, err := config.NewPlan(source{info: config.SourceInfo{Name: "optional-cycle"}, tree: map[string]any{"value": "ignored"}})
		if err != nil {
			t.Fatal(err)
		}
		snapshot, err := config.Load[settings](context.Background(), plan)
		var rejected *config.SnapshotValueError
		if snapshot != nil || !errors.As(err, &rejected) {
			t.Fatal("cyclic Optional payload was not rejected")
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestOptionalRuntimeCycleIsRejected$", "-test.timeout=3s")
	command.Env = append(os.Environ(), "CONFIG_OPTIONAL_CYCLE_PROBE=1")
	if output, err := command.CombinedOutput(); err != nil {
		if len(output) > 2048 {
			output = output[:2048]
		}
		t.Fatalf("bounded Optional cycle: %v; deadline=%v; output=%s", err, ctx.Err(), output)
	}
}

func TestLoadDoesNotPublishAfterFinalValidatorCancels(t *testing.T) {
	plan, err := config.NewPlan(source{info: config.SourceInfo{Name: "cancel"}, tree: map[string]any{"value": "ok"}})
	if err != nil {
		t.Fatal(err)
	}
	type settings struct {
		Value string `config:"value"`
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	snapshot, err := config.LoadWithValidators(ctx, plan, func(context.Context, settings) error { cancel(); return nil })
	if snapshot != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled candidate published=%t error=%v", snapshot != nil, err)
	}
}

func TestLoadTreeConsumesNewNestedTombstones(t *testing.T) {
	plan, err := config.NewPlan(source{info: config.SourceInfo{Name: "delete"}, tree: map[string]any{"nested": map[string]any{"removed": merge.Delete{}, "keep": "value"}}})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := config.LoadTree(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := snapshot.Value()["nested"].(map[string]any)["removed"]; exists {
		t.Fatal("supported Source tombstone escaped merge")
	}
	if _, exists := snapshot.Origin("nested.removed"); exists {
		t.Fatal("deleted path retains provenance")
	}
}
func TestLoadTreePreservesOverlappingAcyclicSlices(t *testing.T) {
	items := make([]any, 2)
	items[1] = items[:1]
	plan, err := config.NewPlan(source{info: config.SourceInfo{Name: "finite"}, tree: map[string]any{"items": items}})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := config.LoadTree(context.Background(), plan)
	if err != nil || snapshot == nil {
		t.Fatalf("finite overlapping slices rejected: %v", err)
	}
}

type boundaryRecursiveNode struct {
	Name string                 `config:"name,secret"`
	Next *boundaryRecursiveNode `config:"next"`
}

func (v *boundaryRecursiveNode) UnmarshalConfigValue(input any) error {
	v.Name = "root"
	if text, _ := input.(string); text == "cycle" {
		v.Next = v
	} else {
		v.Next = &boundaryRecursiveNode{Name: "leaf"}
	}
	return nil
}
func TestTypedRuntimeGraphsTerminate(t *testing.T) {
	if probe := os.Getenv("CONFIG_TYPED_GRAPH_PROBE"); probe != "" {
		debug.SetMaxStack(1 << 20)
		type settings struct {
			Node boundaryRecursiveNode `config:"node"`
		}
		var input any = probe
		if probe == "finite" {
			input = map[string]any{"name": "root", "next": map[string]any{"name": "leaf", "next": nil}}
		}
		plan, err := config.NewPlan(source{info: config.SourceInfo{Name: "graph"}, tree: map[string]any{"node": input}})
		if err != nil {
			t.Fatal(err)
		}
		snapshot, err := config.Load[settings](context.Background(), plan)
		if probe == "cycle" {
			var rejected *config.SnapshotValueError
			if snapshot != nil || !errors.As(err, &rejected) {
				t.Fatal("typed runtime cycle not rejected")
			}
		} else {
			if err != nil || snapshot == nil || snapshot.Value().Node.Next.Name != "leaf" {
				t.Fatal("finite recursive typed value not preserved")
			}
			for _, path := range []string{"node.name", "node.next.name"} {
				origin, exists := snapshot.Origin(path)
				if !exists || !origin.Sensitive {
					t.Fatal("finite recursive metadata lost")
				}
			}
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []string{"cycle", "finite"} {
		t.Run(probe, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestTypedRuntimeGraphsTerminate$", "-test.timeout=3s")
			command.Env = append(os.Environ(), "CONFIG_TYPED_GRAPH_PROBE="+probe)
			if output, err := command.CombinedOutput(); err != nil {
				if len(output) > 2048 {
					output = output[:2048]
				}
				t.Fatalf("bounded typed graph: %v; deadline=%v; output=%s", err, ctx.Err(), output)
			}
		})
	}
}
