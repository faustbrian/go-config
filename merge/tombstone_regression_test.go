package merge_test

import (
	"testing"

	"github.com/faustbrian/go-config/v2/merge"
)

func TestNewObjectSubtreesConsumeDeletionMarkers(t *testing.T) {
	for _, lower := range []map[string]any{nil, {"nested": nil}, {"nested": map[string]any{}}} {
		upper := map[string]any{"nested": map[string]any{"removed": merge.Delete{}, "keep": "value", "deeper": map[string]any{"removed": merge.Delete{}}}}
		tree, err := merge.Trees(lower, upper)
		if err != nil {
			t.Fatal(err)
		}
		nested := tree["nested"].(map[string]any)
		if _, exists := nested["removed"]; exists {
			t.Fatal("new object retained deletion marker")
		}
		if _, exists := nested["deeper"].(map[string]any)["removed"]; exists {
			t.Fatal("deeper object retained deletion marker")
		}
		if nested["keep"] != "value" {
			t.Fatal("ordinary sibling lost")
		}
		if _, exists := upper["nested"].(map[string]any)["removed"]; !exists {
			t.Fatal("upper input mutated")
		}
	}
}
