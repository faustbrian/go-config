package programmatic_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime/debug"
	"testing"
	"time"

	config "github.com/faustbrian/go-config/v3"
	"github.com/faustbrian/go-config/v3/merge"
	"github.com/faustbrian/go-config/v3/programmatic"
)

func TestProgrammaticRejectsRuntimeCycles(t *testing.T) {
	if probe := os.Getenv("CONFIG_PROGRAMMATIC_CYCLE_PROBE"); probe != "" {
		debug.SetMaxStack(1 << 20)
		m := map[string]any{}
		if probe == "map" {
			m["self"] = m
		} else {
			items := make([]any, 1)
			items[0] = items
			m["items"] = items
		}
		source, err := programmatic.Map("cycle", 0, m)
		var rejected *config.TreeCycleError
		if source != nil || !errors.As(err, &rejected) {
			t.Fatal("cyclic programmatic input was not rejected")
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []string{"map", "slice"} {
		t.Run(probe, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestProgrammaticRejectsRuntimeCycles$", "-test.timeout=3s")
			command.Env = append(os.Environ(), "CONFIG_PROGRAMMATIC_CYCLE_PROBE="+probe)
			if output, err := command.CombinedOutput(); err != nil {
				if len(output) > 2048 {
					output = output[:2048]
				}
				t.Fatalf("bounded programmatic cycle: %v; deadline=%v; output=%s", err, ctx.Err(), output)
			}
		})
	}
}
func TestProgrammaticPreservesSharedAcyclicContainers(t *testing.T) {
	shared := map[string]any{"value": "ok"}
	tree := map[string]any{"left": shared, "right": shared}
	source, err := programmatic.Map("acyclic", 0, tree)
	if err != nil {
		t.Fatal(err)
	}
	document, err := source.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	document.Tree["left"].(map[string]any)["value"] = "changed"
	if document.Tree["right"].(map[string]any)["value"] != "ok" {
		t.Fatal("copied sibling aliases")
	}
}

func TestProgrammaticKeepsStructuralTombstoneBoundary(t *testing.T) {
	if source, err := programmatic.Map("delete", 0, map[string]any{"removed": merge.Delete{}}); source != nil || err == nil {
		t.Fatal("arbitrary struct admission changed")
	}
}
func TestProgrammaticPreservesOverlappingAcyclicSlices(t *testing.T) {
	items := make([]any, 2)
	items[1] = items[:1]
	source, err := programmatic.Map("finite", 0, map[string]any{"items": items})
	if err != nil || source == nil {
		t.Fatalf("finite overlapping slices rejected: %v", err)
	}
}

// Child probes retain hard-bounded historical-hang proof. This direct public
// assertion also records the rejected active map/slice path in native coverage.
func TestProgrammaticRejectsActiveCyclesDirectly(t *testing.T) {
	cyclicMap := map[string]any{}
	cyclicMap["self"] = cyclicMap
	cyclicSlice := make([]any, 1)
	cyclicSlice[0] = cyclicSlice
	for _, test := range []struct {
		name string
		tree map[string]any
		path string
	}{
		{"map", cyclicMap, "self"},
		{"slice", map[string]any{"items": cyclicSlice}, "items[0]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source, err := programmatic.Map("cycle", 0, test.tree)
			var rejected *config.TreeCycleError
			if source != nil || !errors.As(err, &rejected) || rejected.Path != test.path {
				t.Fatalf("Map = %v, %v; want no source and cycle at %s", source, err, test.path)
			}
		})
	}
}
