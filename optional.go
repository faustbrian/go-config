package config

import (
	"reflect"

	"github.com/faustbrian/go-config/v2/decode"
)

// Presence distinguishes values that ordinary Go zero values cannot.
type Presence uint8

const (
	// Absent means no source supplied the field.
	Absent Presence = iota
	// Null means a source explicitly supplied null.
	Null
	// Present means a source supplied a value, including empty or zero.
	Present
	// Defaulted means a default source supplied the value.
	Defaulted
)

// Optional preserves absence, explicit null, and present zero values.
// Its zero value is Absent. Get returns T directly, not a defensive copy.
// Snapshot loading admits the payload under the ordinary typed cloneability
// rules before its clone hook runs; unsupported or cyclic state is rejected.
type Optional[T any] struct {
	value T
	state Presence
}

// State reports the value's presence state.
func (o Optional[T]) State() Presence { return o.state }

// Get returns the value when it is present or defaulted.
func (o Optional[T]) Get() (T, bool) {
	return o.value, o.state == Present || o.state == Defaulted
}

// ConfigTextTarget identifies T to typed textual sources.
func (o Optional[T]) ConfigTextTarget() reflect.Type { return reflect.TypeFor[T]() }

func (o *Optional[T]) setConfigPresence(state Presence) { o.state = state }

// UnmarshalConfigValue decodes into a fresh T and changes the receiver only on
// success. Nil input clears the value and sets Null; other success sets Present.
// A nil receiver is not supported. Extension behavior follows decode.Value.
func (o *Optional[T]) UnmarshalConfigValue(input any) error {
	var zero T
	if input == nil {
		o.value = zero
		o.state = Null
		return nil
	}

	var value T
	if err := decode.Value(input, &value); err != nil {
		return err
	}
	o.value = value
	o.state = Present
	return nil
}

func (o Optional[T]) cloneConfigValue() any {
	return Optional[T]{value: cloneTyped(o.value), state: o.state}
}

func (o Optional[T]) validateConfigValue(path string, visiting map[cloneVisit]bool) error {
	return validateSnapshotValue(reflect.ValueOf(o.value), path, visiting)
}
