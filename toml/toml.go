// Package toml loads explicit TOML 1.0.0 configuration into canonical documents.
// The root table is normalized; date/time categories become strings, and TOML
// non-finite floats remain float64. Depth/key admission follows backend
// materialization, not a pre-parser heap cap.
// Bytes copies input before load-time byte admission; FromFS borrows/reopens an
// fs.FS path and checks generation metadata. Sources own no watcher/goroutine,
// retry or shutdown. Context is cooperative around owned/parser work, not forced
// preemption of ordinary filesystem or backend operations. See CONFIG-DEC-001
// through CONFIG-DEC-009 and the maintained conformance guide for exact policy,
// not a claim of complete specification conformance.
package toml

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"time"

	btoml "github.com/BurntSushi/toml"
	config "github.com/faustbrian/go-config/v2"
	"github.com/faustbrian/go-config/v2/internal/safeerror"
	"github.com/faustbrian/go-config/v2/internal/sourceio"
)

const (
	defaultMaxBytes = 1 << 20
	defaultMaxDepth = 64
	defaultMaxKeys  = 100_000
)

// Limits bounds parser resource use. Zero values select conservative defaults.
type Limits struct {
	// MaxBytes caps each accepted input on load; zero selects 1 MiB.
	MaxBytes int64
	// MaxDepth caps normalized value depth; zero selects 64.
	MaxDepth int
	// MaxKeys caps mapping keys across the document; zero selects 100000.
	MaxKeys int
}

// Options configures source metadata and parser bounds.
type Options struct {
	// Name is a required nonblank, non-secret diagnostic label.
	Name string
	// Priority passes unchanged to NewPlan ordering.
	Priority int
	// Sensitive marks provenance, not encrypted or erased storage.
	Sensitive bool
	// Optional permits plan-level ErrNotFound suppression only.
	Optional bool
	// Limits chooses positive admission bounds; negative values are invalid.
	Limits Limits
}

// ParseError wraps a parser cause without including source text in diagnostics.
type ParseError struct {
	// Cause retains parser failure; Unwrap hides arbitrary text and concrete types.
	Cause error
}

// Error returns the failure category and declared metadata, not the rejected value.
func (*ParseError) Error() string { return "decode TOML config: malformed document" }

// Unwrap preserves errors.Is identity without exposing arbitrary cause text/type.
func (e *ParseError) Unwrap() error {
	return safeerror.Redact(e.Cause, "TOML parser cause redacted")
}

// Format prevents detailed formatting from traversing the parser cause.
func (e *ParseError) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(e.Error()))
}

// MarshalText serializes only the redacted diagnostic message.
func (e *ParseError) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}

type source struct {
	info   config.SourceInfo
	input  sourceio.Input
	limits Limits
}

// Bytes constructs a repeatable source from copied data and validated options.
// Size is admitted only on Load after the constructor copy; decoded plaintext
// and copied input are retained without guaranteed secure erasure.
func Bytes(data []byte, options Options) (config.Source, error) {
	info, limits, err := validate(options)
	if err != nil {
		return nil, err
	}
	return &source{info: info, input: sourceio.Bytes(data), limits: limits}, nil
}

// FromFS constructs a source for an fs.ValidPath name in a borrowed filesystem.
// It reopens the path each load, checks metadata/generation and owns closing each
// admitted file. It imposes no discovery permission/symlink policy or I/O sandbox;
// ordinary filesystem methods must return and support the caller's concurrency.
func FromFS(filesystem fs.FS, path string, options Options) (config.Source, error) {
	info, limits, err := validate(options)
	if err != nil {
		return nil, err
	}
	input, err := sourceio.FromFS(filesystem, path)
	if err != nil {
		return nil, err
	}
	return &source{info: info, input: input, limits: limits}, nil
}

func (s *source) Info() config.SourceInfo { return s.info }

func (s *source) Load(ctx context.Context) (config.Document, error) {
	data, err := s.input.Read(ctx, s.limits.MaxBytes)
	if err != nil {
		return config.Document{}, err
	}
	// Backend grammar/duplicate checks precede owned depth/key normalization;
	// those limits do not prevent backend materialization under MaxBytes.
	tree := make(map[string]any)
	if _, err := btoml.NewDecoder(sourceio.ContextReader(ctx, data)).Decode(&tree); err != nil {
		return config.Document{}, &ParseError{
			Cause: safeerror.Redact(err, "TOML parser cause redacted"),
		}
	}
	keys := 0
	value, err := normalize(ctx, tree, 1, "", s.limits, &keys)
	if err != nil {
		return config.Document{}, err
	}
	normalized := value.(map[string]any)
	return config.Document{Tree: normalized}, nil
}

func validate(options Options) (config.SourceInfo, Limits, error) {
	if strings.TrimSpace(options.Name) == "" {
		return config.SourceInfo{}, Limits{}, errors.New("TOML source name must not be empty")
	}
	limits := options.Limits
	if limits.MaxBytes < 0 || limits.MaxDepth < 0 || limits.MaxKeys < 0 {
		return config.SourceInfo{}, Limits{}, errors.New("TOML source limits must not be negative")
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = defaultMaxBytes
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = defaultMaxDepth
	}
	if limits.MaxKeys == 0 {
		limits.MaxKeys = defaultMaxKeys
	}
	return config.SourceInfo{
		Name: options.Name, Priority: options.Priority,
		Sensitive: options.Sensitive, Optional: options.Optional,
	}, limits, nil
}

func normalize(
	ctx context.Context,
	value any,
	depth int,
	path string,
	limits Limits,
	keys *int,
) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if depth > limits.MaxDepth {
		return nil, fmt.Errorf("decode TOML config: depth exceeds %d at %q", limits.MaxDepth, path)
	}
	switch value := value.(type) {
	case map[string]any:
		object := make(map[string]any, len(value))
		for key, item := range value {
			*keys++
			childPath := join(path, key)
			if *keys > limits.MaxKeys {
				return nil, fmt.Errorf("decode TOML config: keys exceed %d at %q", limits.MaxKeys, childPath)
			}
			converted, err := normalize(ctx, item, depth+1, childPath, limits, keys)
			if err != nil {
				return nil, err
			}
			object[key] = converted
		}
		return object, nil
	case []map[string]any:
		items := make([]any, len(value))
		for index, item := range value {
			converted, err := normalize(ctx, item, depth+1, fmt.Sprintf("%s[%d]", path, index), limits, keys)
			if err != nil {
				return nil, err
			}
			items[index] = converted
		}
		return items, nil
	case []any:
		items := make([]any, len(value))
		for index, item := range value {
			converted, err := normalize(ctx, item, depth+1, fmt.Sprintf("%s[%d]", path, index), limits, keys)
			if err != nil {
				return nil, err
			}
			items[index] = converted
		}
		return items, nil
	case time.Time:
		switch value.Location().String() {
		case "date-local":
			return value.Format("2006-01-02"), nil
		case "time-local":
			return value.Format("15:04:05.999999999"), nil
		case "datetime-local":
			return value.Format("2006-01-02T15:04:05.999999999"), nil
		default:
			return value.Format(time.RFC3339Nano), nil
		}
	case string, bool, int64, float64:
		return value, nil
	default:
		typeOf := reflect.TypeOf(value)
		if typeOf == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("decode TOML config at %q: unsupported %s", path, typeOf)
	}
}

func join(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}
