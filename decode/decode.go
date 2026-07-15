// Package decode maps format-independent configuration trees into Go values.
package decode

import (
	"encoding"
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/faustbrian/go-config/internal/safeerror"
)

var (
	textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
	durationType        = reflect.TypeFor[time.Duration]()
	urlType             = reflect.TypeFor[url.URL]()
)

// ValueUnmarshaler decodes a format-independent configuration value. It is
// intended for presence-aware wrappers that must observe explicit null.
type ValueUnmarshaler interface {
	UnmarshalConfigValue(any) error
}

// TextTargeter lets presence-aware wrappers declare the underlying type that
// textual sources should convert before calling ValueUnmarshaler.
type TextTargeter interface {
	ConfigTextTarget() reflect.Type
}

var valueUnmarshalerType = reflect.TypeFor[ValueUnmarshaler]()

// FieldError describes a decode failure without including the received value.
type FieldError struct {
	Path     string
	Source   string
	Location string
	Expected string
	Received string
	Cause    error
}

func (e *FieldError) Error() string {
	message := fmt.Sprintf(
		"decode config field %q: expected %s, received %s",
		e.Path,
		e.Expected,
		e.Received,
	)
	if e.Cause != nil {
		message += ": conversion failed"
	}
	if e.Source != "" {
		message += fmt.Sprintf(" from source %q", e.Source)
	}
	if e.Location != "" {
		message += fmt.Sprintf(" at %q", e.Location)
	}
	return message
}

func (e *FieldError) Unwrap() error { return redactCause(e.Cause) }

// Errors contains independent field failures in lexical path order.
type Errors struct {
	Fields []*FieldError
}

// PanicError reports a recovered extension panic without retaining its value.
type PanicError struct {
	Operation string
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("config %s panicked", e.Operation)
}

func (e *Errors) Error() string {
	return fmt.Sprintf("decode config: %d field errors", len(e.Fields))
}

func (e *Errors) Unwrap() []error {
	errors := make([]error, len(e.Fields))
	for index, field := range e.Fields {
		errors[index] = field
	}
	return errors
}

// Into atomically decodes tree into a non-nil pointer. Destination is assigned
// only after the complete tree has decoded successfully.
func Into(tree map[string]any, destination any) error {
	return Value(tree, destination)
}

// Value atomically decodes one format-independent value into destination.
func Value(input any, destination any) error {
	pointer := reflect.ValueOf(destination)
	if !pointer.IsValid() || pointer.Kind() != reflect.Pointer || pointer.IsNil() {
		return errors.New("decode destination must be a non-nil pointer")
	}

	candidate := reflect.New(pointer.Elem().Type()).Elem()
	if err := decodeValue(input, candidate, ""); err != nil {
		return normalizeErrors(err)
	}
	pointer.Elem().Set(candidate)

	return nil
}

func decodeValue(input any, destination reflect.Value, path string) error {
	if destination.CanAddr() && destination.Addr().Type().Implements(valueUnmarshalerType) {
		unmarshaler := destination.Addr().Interface().(ValueUnmarshaler)
		if err := callValueUnmarshaler(unmarshaler, input); err != nil {
			return fieldError(path, destination.Type().String(), input, err)
		}
		return nil
	}

	if input == nil {
		switch destination.Kind() {
		case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface:
			destination.SetZero()
			return nil
		default:
			return fieldError(path, destination.Type().String(), input, nil)
		}
	}

	if destination.Kind() == reflect.Pointer {
		value := reflect.New(destination.Type().Elem())
		if err := decodeValue(input, value.Elem(), path); err != nil {
			return err
		}
		destination.Set(value)
		return nil
	}

	if destination.Type() == durationType {
		text, ok := input.(string)
		if !ok {
			return fieldError(path, "duration string", input, nil)
		}
		value, err := time.ParseDuration(text)
		if err != nil {
			return fieldError(path, "duration string", input, err)
		}
		destination.SetInt(int64(value))
		return nil
	}

	if destination.Type() == urlType {
		text, ok := input.(string)
		if !ok {
			return fieldError(path, "URL string", input, nil)
		}
		value, err := url.ParseRequestURI(text)
		if err != nil {
			return fieldError(path, "URL string", input, err)
		}
		destination.Set(reflect.ValueOf(*value))
		return nil
	}

	if destination.CanAddr() && destination.Addr().Type().Implements(textUnmarshalerType) {
		text, ok := input.(string)
		if !ok {
			return fieldError(path, "string", input, nil)
		}
		unmarshaler := destination.Addr().Interface().(encoding.TextUnmarshaler)
		if err := callTextUnmarshaler(unmarshaler, text); err != nil {
			return fieldError(path, destination.Type().String(), input, err)
		}
		return nil
	}

	switch destination.Kind() {
	case reflect.Struct:
		return decodeStruct(input, destination, path)
	case reflect.Map:
		return decodeMap(input, destination, path)
	case reflect.Slice:
		return decodeSlice(input, destination, path)
	case reflect.String:
		value, ok := input.(string)
		if !ok {
			return fieldError(path, destination.Type().String(), input, nil)
		}
		destination.SetString(value)
	case reflect.Bool:
		value, ok := input.(bool)
		if !ok {
			return fieldError(path, destination.Type().String(), input, nil)
		}
		destination.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, ok := signedInteger(input)
		if !ok || destination.OverflowInt(value) {
			return fieldError(path, destination.Type().String(), input, nil)
		}
		destination.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, ok := unsignedInteger(input)
		if !ok || destination.OverflowUint(value) {
			return fieldError(path, destination.Type().String(), input, nil)
		}
		destination.SetUint(value)
	case reflect.Float32, reflect.Float64:
		value, ok := floatingPoint(input)
		if !ok || destination.OverflowFloat(value) {
			return fieldError(path, destination.Type().String(), input, nil)
		}
		destination.SetFloat(value)
	case reflect.Interface:
		destination.Set(reflect.ValueOf(cloneUntyped(input)))
	default:
		return fieldError(path, destination.Type().String(), input, nil)
	}

	return nil
}

func callValueUnmarshaler(unmarshaler ValueUnmarshaler, input any) (err error) {
	defer func() {
		if recover() != nil {
			err = &PanicError{Operation: "value unmarshaler"}
		}
	}()
	return unmarshaler.UnmarshalConfigValue(input)
}

func callTextUnmarshaler(unmarshaler encoding.TextUnmarshaler, input string) (err error) {
	defer func() {
		if recover() != nil {
			err = &PanicError{Operation: "text unmarshaler"}
		}
	}()
	return unmarshaler.UnmarshalText([]byte(input))
}

func decodeStruct(input any, destination reflect.Value, path string) error {
	object, ok := input.(map[string]any)
	if !ok {
		return fieldError(path, "object", input, nil)
	}

	type structField struct {
		index    int
		required bool
	}
	fields := make(map[string]structField, destination.NumField())
	order := make([]string, 0, destination.NumField())
	for index := 0; index < destination.NumField(); index++ {
		definition := destination.Type().Field(index)
		if !definition.IsExported() {
			continue
		}
		name := definition.Tag.Get("config")
		options := ""
		if comma := strings.IndexByte(name, ','); comma >= 0 {
			options = name[comma+1:]
			name = name[:comma]
		}
		if name == "-" {
			continue
		}
		if name == "" {
			name = strings.ToLower(definition.Name)
		}
		if _, exists := fields[name]; exists {
			return fieldError(join(path, name), "unambiguous field", input, nil)
		}
		fields[name] = structField{
			index: index, required: hasOption(options, "required"),
		}
		order = append(order, name)
	}

	seen := make(map[string]bool, len(object))
	var failures []error
	for _, name := range sortedKeys(object) {
		value := object[name]
		field, exists := fields[name]
		if !exists {
			failures = append(failures, fieldError(join(path, name), "known field", value, nil))
			continue
		}
		if err := decodeValue(value, destination.Field(field.index), join(path, name)); err != nil {
			failures = append(failures, err)
		}
		seen[name] = true
	}
	for _, name := range order {
		if fields[name].required && !seen[name] {
			failures = append(failures, &FieldError{
				Path: join(path, name), Expected: "required field", Received: "absent",
			})
		}
	}

	return combineErrors(failures)
}

func hasOption(options, wanted string) bool {
	for _, option := range strings.Split(options, ",") {
		if option == wanted {
			return true
		}
	}
	return false
}

func decodeMap(input any, destination reflect.Value, path string) error {
	object, ok := input.(map[string]any)
	if !ok || destination.Type().Key().Kind() != reflect.String {
		return fieldError(path, destination.Type().String(), input, nil)
	}

	result := reflect.MakeMapWithSize(destination.Type(), len(object))
	var failures []error
	for _, key := range sortedKeys(object) {
		value := object[key]
		item := reflect.New(destination.Type().Elem()).Elem()
		if err := decodeValue(value, item, join(path, key)); err != nil {
			failures = append(failures, err)
			continue
		}
		mapKey := reflect.New(destination.Type().Key()).Elem()
		mapKey.SetString(key)
		result.SetMapIndex(mapKey, item)
	}
	if err := combineErrors(failures); err != nil {
		return err
	}
	destination.Set(result)

	return nil
}

func decodeSlice(input any, destination reflect.Value, path string) error {
	items, ok := input.([]any)
	if !ok {
		return fieldError(path, destination.Type().String(), input, nil)
	}

	result := reflect.MakeSlice(destination.Type(), len(items), len(items))
	var failures []error
	for index, item := range items {
		if err := decodeValue(item, result.Index(index), fmt.Sprintf("%s[%d]", path, index)); err != nil {
			failures = append(failures, err)
		}
	}
	if err := combineErrors(failures); err != nil {
		return err
	}
	destination.Set(result)

	return nil
}

func sortedKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func combineErrors(failures []error) error {
	if len(failures) == 0 {
		return nil
	}
	if len(failures) == 1 {
		return failures[0]
	}
	return normalizeErrors(&Errors{Fields: flattenErrors(failures)})
}

func normalizeErrors(err error) error {
	fields := flattenErrors([]error{err})
	if len(fields) == 1 {
		return fields[0]
	}
	sort.SliceStable(fields, func(left, right int) bool {
		return fields[left].Path < fields[right].Path
	})
	return &Errors{Fields: fields}
}

func flattenErrors(failures []error) []*FieldError {
	fields := make([]*FieldError, 0, len(failures))
	for _, failure := range failures {
		var aggregate *Errors
		if errors.As(failure, &aggregate) {
			fields = append(fields, aggregate.Fields...)
			continue
		}
		var field *FieldError
		if errors.As(failure, &field) {
			fields = append(fields, field)
		}
	}
	return fields
}

func fieldError(path, expected string, received any, cause error) error {
	return &FieldError{
		Path:     path,
		Expected: expected,
		Received: describe(received),
		Cause:    redactCause(cause),
	}
}

func redactCause(cause error) error {
	// A wrapped arbitrary error may contain secret text. Only the exact
	// library-owned type is safe to expose through errors.As.
	//nolint:errorlint // Deliberately do not traverse an untrusted error chain.
	if _, safe := cause.(*PanicError); safe {
		return cause
	}
	return safeerror.Redact(cause, "config conversion cause redacted")
}

func describe(value any) string {
	if value == nil {
		return "null"
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Map:
		return "object"
	case reflect.Slice, reflect.Array:
		return "array"
	default:
		return reflect.TypeOf(value).String()
	}
}

func join(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func signedInteger(value any) (int64, bool) {
	switch value := value.(type) {
	case int:
		return int64(value), true
	case int8:
		return int64(value), true
	case int16:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true
	case uint:
		if uint64(value) <= math.MaxInt64 {
			return int64(value), true
		}
	case uint8:
		return int64(value), true
	case uint16:
		return int64(value), true
	case uint32:
		return int64(value), true
	case uint64:
		if value <= math.MaxInt64 {
			return int64(value), true
		}
	}
	return 0, false
}

func unsignedInteger(value any) (uint64, bool) {
	signed, ok := signedInteger(value)
	if ok && signed >= 0 {
		return uint64(signed), true
	}
	if value, ok := value.(uint64); ok {
		return value, true
	}
	return 0, false
}

func floatingPoint(value any) (float64, bool) {
	switch value := value.(type) {
	case float32:
		return float64(value), true
	case float64:
		return value, true
	default:
		signed, ok := signedInteger(value)
		return float64(signed), ok
	}
}

func cloneUntyped(value any) any {
	switch value := value.(type) {
	case map[string]any:
		clone := make(map[string]any, len(value))
		for key, item := range value {
			clone[key] = cloneUntyped(item)
		}
		return clone
	case []any:
		clone := make([]any, len(value))
		for index, item := range value {
			clone[index] = cloneUntyped(item)
		}
		return clone
	default:
		return value
	}
}
