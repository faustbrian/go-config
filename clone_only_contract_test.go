package config

import (
	"errors"
	"reflect"
	"testing"
)

// This sealed hook is package-owned, not a public third-party extension point.
// Its private mutable state is safe only because its clone owns a fresh slice.
type trustedCloneOnlyValue struct{ values []int }

func (value trustedCloneOnlyValue) cloneConfigValue() any {
	return trustedCloneOnlyValue{values: append([]int(nil), value.values...)}
}

func TestTrustedCloneOnlyValueAdmitsOwnedPrivateState(t *testing.T) {
	input := trustedCloneOnlyValue{values: []int{7, 11}}
	if err := validateSnapshotValue(reflect.ValueOf(input), "value", make(map[cloneVisit]bool)); err != nil {
		t.Fatalf("package-owned cloner admission: %v", err)
	}
	first := cloneTyped(input)
	input.values[0] = 99
	if first.values[0] != 7 {
		t.Fatal("clone retained input's private slice")
	}
	second := cloneTyped(first)
	first.values[1] = 88
	if second.values[1] != 11 {
		t.Fatal("second clone retained previous private slice")
	}
	second.values[0] = 33
	if first.values[0] != 7 || input.values[0] != 99 {
		t.Fatal("mutating a clone changed another owner")
	}

	// The same opaque mutable state without the sealed cloning responsibility
	// remains unsupported rather than silently borrowing the private reference.
	type untrusted struct{ values []int }
	err := validateSnapshotValue(reflect.ValueOf(untrusted{values: []int{7}}), "value", make(map[cloneVisit]bool))
	var rejected *SnapshotValueError
	if !errors.As(err, &rejected) || rejected.Path != "value.values" {
		t.Fatalf("opaque untrusted admission = %v; want SnapshotValueError at value.values", err)
	}
}
