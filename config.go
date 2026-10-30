// Package config composes explicit Sources into deterministic configuration
// snapshots. NewPlan fixes low-to-high precedence; LoadTree merges canonical
// documents, and Load adds strict decoding and validation. No source is discovered
// or refreshed implicitly, and no watcher, goroutine or process environment is owned.
//
// Sources, metadata and extension hooks are trusted application collaborators;
// context cancellation is cooperative, not forced preemption. LoadTree isolates
// canonical maps/slices and admitted typed containers, including Optional payloads.
// Typed state is admitted before recursive owned traversal. See the
// maintained layering and security guides for composition and trust boundaries.
package config

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/faustbrian/go-config/v3/decode"
	"github.com/faustbrian/go-config/v3/internal/safeerror"
	"github.com/faustbrian/go-config/v3/merge"
	"github.com/faustbrian/go-config/v3/validation"
)

// SourceInfo describes source policy and diagnostic labels, not its values.
// Names are not sanitized; callers must keep credentials out of
// this metadata. NewPlan captures Info once, rather than rereading it on load.
type SourceInfo struct {
	// Name is the exact nonblank plan identity and diagnostic label; keep it non-secret.
	Name string
	// Priority orders sources from low to high; later equal-priority entries win.
	Priority int
	// Sensitive marks provenance, not encryption, value erasure or typed redaction.
	Sensitive bool
	// Optional lets LoadTree suppress only errors matching ErrNotFound.
	Optional bool
}

const (
	// PriorityDefaults is the conventional lowest category (10).
	PriorityDefaults = 10
	// PriorityDiscoveredBase orders supplied base documents at 20.
	PriorityDiscoveredBase = 20
	// PriorityDiscoveredProfile orders supplied profile documents at 30.
	PriorityDiscoveredProfile = 30
	// PriorityExplicitFiles orders supplied explicit files at 40.
	PriorityExplicitFiles = 40
	// PriorityDotenv orders supplied dotenv sources at 50.
	PriorityDotenv = 50
	// PriorityEnvironment orders supplied environment sources at 60.
	PriorityEnvironment = 60
	// PriorityOverrides orders supplied override sources at 70, not above arbitrary
	// custom priorities.
	PriorityOverrides = 70
)

// Document is the intermediate tree produced by a Source. LoadTree copies and
// validates canonical maps/slices before merging; a Source must not mutate them
// concurrently. Origins are caller-declared provenance, not authenticated evidence.
type Document struct {
	// Tree contains canonical maps/slices/scalars, nil or merge.Delete markers.
	Tree map[string]any
	// Origins optionally overrides declared provenance by dotted path; values are
	// never stored here.
	Origins map[string]Origin
	// Digest identifies the exact bytes a byte-backed source accepted for this
	// Tree. It is nil for sources without a byte representation. Custom sources
	// building Document with unkeyed literals must add this field.
	Digest *SourceDigest
}

// SourceDigest is the SHA-256 of the exact bytes one source load read and
// parsed, before any normalization. It identifies those bytes; it does not
// authenticate their origin or prove they were not replaced before the read.
// Snapshots never record it for a Sensitive source, because an unsalted digest
// of a short secret lets anyone holding it test guesses offline.
type SourceDigest struct {
	// SHA256 is the digest of the accepted bytes.
	SHA256 [sha256.Size]byte
}

// TreeValueError reports a non-canonical value returned by a Source. It never
// formats the rejected value.
type TreeValueError struct {
	// Path identifies the rejected value by trusted key/index text.
	Path string
	// Type identifies its Go type without formatting the rejected value.
	Type string
}

// TreeCycleError reports a cyclic source tree without traversing it.
type TreeCycleError struct {
	// Path identifies the encountered map/slice cycle without traversing it again.
	Path string
}

// Error returns the failure category and declared metadata, not the rejected value.
func (e *TreeCycleError) Error() string {
	return fmt.Sprintf("config tree at %q: cyclic value", e.Path)
}

// TreeLimitError reports a source-tree structural bound without formatting a
// value.
type TreeLimitError struct {
	// Path identifies the value exceeding a per-source structural bound.
	Path string
	// Kind is the exceeded depth, keys or items category.
	Kind string
	// Limit is the admitted maximum for that category.
	Limit int
}

// Error returns the failure category and declared metadata, not the rejected value.
func (e *TreeLimitError) Error() string {
	return fmt.Sprintf("config tree at %q: %s exceeds %d limit", e.Path, e.Kind, e.Limit)
}

// Error returns the failure category and declared metadata, not the rejected value.
func (e *TreeValueError) Error() string {
	return fmt.Sprintf("config tree at %q: unsupported value type %s", e.Path, e.Type)
}

// SnapshotValueError reports cyclic or uncopyable typed state without formatting
// its value. Optional payloads follow the same admission rules as ordinary fields.
type SnapshotValueError struct {
	// Path identifies typed state rejected before validation/publication.
	Path string
	// Type identifies its Go type, not its value.
	Type string
}

// SourceError identifies a failed source without exposing arbitrary source
// error text. Cause identity remains available through errors.Is.
type SourceError struct {
	// Name is the caller-controlled failed source label.
	Name string
	// Cause retains the error internally; Unwrap hides its text/type, preserving errors.Is.
	Cause error
}

// Error returns the failure category and declared metadata, not the rejected value.
func (e *SourceError) Error() string {
	return fmt.Sprintf("load config source %q: source failed", e.Name)
}

// Unwrap preserves errors.Is identity without exposing arbitrary cause text/type.
func (e *SourceError) Unwrap() error {
	return safeerror.Redact(e.Cause, "config source cause redacted")
}

// Format keeps every fmt verb on the redacted message and trusted source label.
func (e *SourceError) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(e.Error()))
}

// MarshalText serializes only the redacted diagnostic message.
func (e *SourceError) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}

// Error returns the failure category and declared metadata, not the rejected value.
func (e *SnapshotValueError) Error() string {
	return fmt.Sprintf("config snapshot at %q: value type %s is not safely cloneable", e.Path, e.Type)
}

// Source loads one configuration document without mutating global state.
// Implementations own input admission before their allocations, must return
// promptly under a bounded context and must not panic. Concurrent/repeated loads
// require a repeatable, concurrency-safe implementation; Plan adds no source lock.
// Return ErrNotFound (or a matching error) for optional absence, not empty success.
type Source interface {
	// Info returns stable, non-secret diagnostic metadata. NewPlan captures it once.
	Info() SourceInfo
	// Load returns a complete candidate or an error. The caller may copy its
	// canonical tree; do not mutate returned maps/slices during that copy.
	Load(context.Context) (Document, error)
}

// ContextFS extends fs.FS with a cancellable open operation. Filesystem sources
// prefer it on load when implemented and otherwise call fs.FS.Open.
type ContextFS interface {
	fs.FS
	// OpenContext opens an fs.ValidPath name under ctx and returns a file whose
	// lifecycle belongs to the source loader, including cleanup after failure.
	OpenContext(context.Context, string) (fs.File, error)
}

// ContextFile extends fs.File with cancellable read and metadata operations.
// Filesystem sources use these methods when implemented.
type ContextFile interface {
	fs.File
	// ReadContext follows io.Reader count/error rules and cooperates with ctx.
	ReadContext(context.Context, []byte) (int, error)
	// StatContext returns non-nil trustworthy file metadata or an error under ctx.
	StatContext(context.Context) (fs.FileInfo, error)
}

// ContextCloser exposes a cancellable close operation for remote resources.
// fs.FS-backed sources use an independent cleanup context; Reader factories
// use ordinary Close instead. The implementation must cooperate with deadlines.
type ContextCloser interface {
	// CloseContext releases the resource under an independent one-second
	// cleanup context. The implementation must cooperate; no goroutine is abandoned.
	CloseContext(context.Context) error
}

// GenerationFile exposes an opaque stable generation token. Filesystem
// sources compare it before and after reads to reject mixed generations.
type GenerationFile interface {
	fs.File
	// GenerationContext returns an opaque token that changes whenever the readable
	// generation changes. Stable tokens must cover same-size/same-time rewrites.
	GenerationContext(context.Context) (string, error)
}

// Plan is an inspectable, structurally fixed low-to-high precedence source plan.
// It borrows Sources, not their mutable state. The zero Plan has no sources.
// Concurrent loads are safe only when every borrowed Source is safe for them.
type Plan struct {
	sources []plannedSource
}

type plannedSource struct {
	source Source
	info   SourceInfo
}

// NewPlan validates sources and orders them from lowest to highest priority.
// Sources with equal priority retain caller order, so later entries win.
// Nil/typed-nil sources, blank names and duplicate exact names fail. Arbitrary
// priorities are accepted; names are not trimmed or sanitized after validation.
// Source.Info calls are synchronous and are not panic-contained.
func NewPlan(sources ...Source) (Plan, error) {
	planned := make([]plannedSource, 0, len(sources))
	names := make(map[string]struct{}, len(sources))

	for index, source := range sources {
		if source == nil || isNilSource(source) {
			return Plan{}, fmt.Errorf("config plan source %d: nil source", index)
		}

		info := source.Info()
		if strings.TrimSpace(info.Name) == "" {
			return Plan{}, fmt.Errorf("config plan source %d: empty name", index)
		}
		if _, exists := names[info.Name]; exists {
			return Plan{}, fmt.Errorf("config plan: duplicate source name %q", info.Name)
		}
		names[info.Name] = struct{}{}
		planned = append(planned, plannedSource{source: source, info: info})
	}

	sort.SliceStable(planned, func(left, right int) bool {
		return planned[left].info.Priority < planned[right].info.Priority
	})

	return Plan{sources: planned}, nil
}

func isNilSource(source Source) bool {
	value := reflect.ValueOf(source)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// Sources returns a fresh metadata slice in precedence order; labels remain caller-trusted.
func (p Plan) Sources() []SourceInfo {
	result := make([]SourceInfo, len(p.sources))
	for index, source := range p.sources {
		result[index] = source.info
	}

	return result
}

// Origin identifies declared winning-source metadata for a dotted path.
// It never includes a field value, but names/locations/path components are trusted
// caller or source text. It is not proof of authenticity or complete provenance;
// supplied Origin overrides can replace presence/state and source labels.
type Origin struct {
	// Source is the declared source label, which overrides may replace.
	Source string
	// Location is trusted source-provided path/location metadata, not sanitized text.
	Location string
	// Sensitive marks policy metadata; it does not redact the actual typed value.
	Sensitive bool
	// Deprecated records the config tag's deprecation annotation, not rejection.
	Deprecated bool
	// Present records declared presence and may be overridden independently of State.
	Present bool
	// State distinguishes absence, null, present and defaulted provenance.
	State Presence
}

// Snapshot holds a loaded value and declared provenance. LoadTree snapshots
// isolate canonical maps/slices. Typed copying recurses through pointers, maps,
// slices, arrays and exported struct fields. Optional payloads are admitted
// before their clone hook runs; cyclic or unsupported mutable private state
// fails before validation/publication. Strings and immutable scalar state remain shared.
//
// A successfully loaded snapshot may be read concurrently when custom typed
// state is immutable. The zero Snapshot contains T's zero value, no origins and
// no source digests; a nil *Snapshot is not usable. No secure memory erasure
// is provided.
type Snapshot[T any] struct {
	value   T
	origins map[string]Origin
	digests map[string]SourceDigest
}

// Value returns a defensive copy of the loaded value within Snapshot's documented
// copying shape, including array elements and Optional payloads. Strings and
// immutable scalar internals are shared.
func (s *Snapshot[T]) Value() T {
	return cloneTyped(s.value)
}

// Origin returns provenance without exposing the field value.
func (s *Snapshot[T]) Origin(path string) (Origin, bool) {
	origin, ok := s.origins[path]
	return origin, ok
}

// SourceDigest returns the digest of the bytes the named source accepted during
// the load that produced this snapshot. It reports false for sources that
// supply no digest, Sensitive sources, optional sources skipped as not found,
// and unknown names. A source overridden by a later one still reports its own
// digest.
func (s *Snapshot[T]) SourceDigest(name string) (SourceDigest, bool) {
	digest, ok := s.digests[name]
	return digest, ok
}

// LoadTree loads Sources synchronously in precedence order and publishes a tree
// only after every required source and merge succeeds. Atomicity is publication
// only: it neither rolls back source side effects nor reads a cross-source instant.
// Optional sources suppress only errors matching ErrNotFound.
//
// Each source tree is copied and checked separately for cycles, depth 64,
// 100000 keys and 100000 array elements. Counters reset per source; source count,
// aggregate merged bytes and string sizes have no total-plan cap. Source.Load
// and metadata are trusted; context must be non-nil and cancellation is cooperative.
func LoadTree(ctx context.Context, plan Plan) (*Snapshot[map[string]any], error) {
	merged := make(map[string]any)
	origins := make(map[string]Origin)
	digests := make(map[string]SourceDigest)

	for _, planned := range plan.sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		document, err := planned.source.Load(ctx)
		if err != nil {
			if planned.info.Optional && errors.Is(err, ErrNotFound) {
				continue
			}
			return nil, &SourceError{
				Name:  planned.info.Name,
				Cause: safeerror.Redact(err, "config source cause redacted"),
			}
		}
		tree, err := canonicalTree(ctx, document.Tree)
		if err != nil {
			return nil, fmt.Errorf("load config source %q: %w", planned.info.Name, err)
		}

		next, err := merge.Trees(merged, tree)
		if err != nil {
			return nil, fmt.Errorf("merge config source %q: %w", planned.info.Name, err)
		}
		merged = next
		applyOrigins(origins, tree, document.Origins, "", planned.info)
		// An unsalted digest of a short secret is an offline guessing oracle.
		if document.Digest != nil && !planned.info.Sensitive {
			digests[planned.info.Name] = *document.Digest
		}
	}

	return &Snapshot[map[string]any]{value: cloneMap(merged), origins: cloneOrigins(origins), digests: digests}, nil
}

// Load resolves and strictly decodes a Plan into a typed snapshot, including
// any Validate method. It is LoadWithValidators without additional validators;
// the latter documents publication, copying and callback limitations.
func Load[T any](ctx context.Context, plan Plan) (*Snapshot[T], error) {
	return LoadWithValidators[T](ctx, plan)
}

// LoadWithValidators resolves and decodes into a fresh candidate, propagates
// field presence/metadata, rejects known uncopyable typed state, then validates a
// copy before publication. Validation mutation is not an update mechanism.
// Typed admission precedes recursive presence/metadata work and cloning.
// A failure returns nil, but callback side effects are not rolled back.
//
// Decode/validation callbacks run inline with panic recovery at their own seams.
// Arbitrary blocking code cannot be preempted. Validation rechecks ctx after
// the final callback and does not publish canceled candidates. Sources are borrowed.
func LoadWithValidators[T any](
	ctx context.Context,
	plan Plan,
	validators ...validation.Validator[T],
) (*Snapshot[T], error) {
	tree, err := LoadTree(ctx, plan)
	if err != nil {
		return nil, err
	}

	var value T
	if err := decode.IntoContext(ctx, tree.value, &value); err != nil {
		annotateDecodeError(err, tree.origins)
		return nil, err
	}
	// Hooks may return cyclic or uncopyable state. Admit it before recursive
	// presence/metadata traversal as well as before cloning and callbacks.
	if err := validateSnapshotValue(reflect.ValueOf(value), "", make(map[cloneVisit]bool)); err != nil {
		return nil, err
	}
	applyPresence(reflect.ValueOf(&value).Elem(), "", tree.origins)
	applyFieldMetadata(reflect.ValueOf(value), "", tree.origins)
	if err := validation.Run(ctx, cloneTyped(value), validators...); err != nil {
		return nil, err
	}

	return &Snapshot[T]{value: cloneTyped(value), origins: cloneOrigins(tree.origins), digests: tree.digests}, nil
}

func applyFieldMetadata(value reflect.Value, path string, origins map[string]Origin) {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}
	for index := 0; index < value.NumField(); index++ {
		definition := value.Type().Field(index)
		if definition.IsExported() {
			applyDefinitionMetadata(definition, value.Field(index), path, origins)
		}
	}
}

func applyDefinitionMetadata(
	definition reflect.StructField,
	value reflect.Value,
	path string,
	origins map[string]Origin,
) {
	name, options, _ := strings.Cut(definition.Tag.Get("config"), ",")
	if name == "-" {
		return
	}
	if name == "" {
		name = strings.ToLower(definition.Name)
	}
	fieldPath := name
	if path != "" {
		fieldPath = path + "." + name
	}
	sensitive := metadataOption(options, "secret")
	deprecated := metadataOption(options, "deprecated")
	if sensitive || deprecated {
		markOriginMetadata(origins, fieldPath, sensitive, deprecated)
	}
	applyFieldMetadata(value, fieldPath, origins)
}

func metadataOption(options, wanted string) bool {
	for _, option := range strings.Split(options, ",") {
		if option == wanted {
			return true
		}
	}
	return false
}

func markOriginMetadata(
	origins map[string]Origin,
	path string,
	sensitive bool,
	deprecated bool,
) {
	prefix := path + "."
	for candidate, origin := range origins {
		if candidate == path || strings.HasPrefix(candidate, prefix) {
			origin.Sensitive = origin.Sensitive || sensitive
			origin.Deprecated = origin.Deprecated || deprecated
			origins[candidate] = origin
		}
	}
}

type presenceSetter interface {
	setConfigPresence(Presence)
}

func applyPresence(value reflect.Value, path string, origins map[string]Origin) {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}
	if value.CanAddr() {
		if setter, ok := value.Addr().Interface().(presenceSetter); ok {
			state := Absent
			if origin, exists := origins[path]; exists {
				state = origin.State
			}
			setter.setConfigPresence(state)
			return
		}
	}
	if value.Kind() != reflect.Struct {
		return
	}
	for index := 0; index < value.NumField(); index++ {
		definition := value.Type().Field(index)
		if definition.IsExported() {
			applyFieldPresence(value.Field(index), definition, path, origins)
		}
	}
}

func applyFieldPresence(
	value reflect.Value,
	definition reflect.StructField,
	path string,
	origins map[string]Origin,
) {
	name, _, _ := strings.Cut(definition.Tag.Get("config"), ",")
	if name == "-" {
		return
	}
	if name == "" {
		name = strings.ToLower(definition.Name)
	}
	fieldPath := name
	if path != "" {
		fieldPath = path + "." + name
	}
	applyPresence(value, fieldPath, origins)
}

func annotateDecodeError(err error, origins map[string]Origin) {
	var failures *decode.Errors
	if errors.As(err, &failures) {
		for _, failure := range failures.Fields {
			annotateFieldError(failure, origins)
		}
		return
	}

	var failure *decode.FieldError
	if errors.As(err, &failure) {
		annotateFieldError(failure, origins)
	}
}

func annotateFieldError(failure *decode.FieldError, origins map[string]Origin) {
	origin, ok := nearestOrigin(failure.Path, origins)
	if !ok {
		return
	}
	failure.Source = origin.Source
	failure.Location = origin.Location
}

func nearestOrigin(path string, origins map[string]Origin) (Origin, bool) {
	candidate := path
	for candidate != "" {
		if origin, ok := origins[candidate]; ok {
			return origin, true
		}
		if strings.HasSuffix(candidate, "]") {
			if strings.Contains(candidate, "[") {
				bracket := strings.LastIndexByte(candidate, '[')
				candidate = candidate[:bracket]
				continue
			}
		}
		switch dot := strings.LastIndexByte(candidate, '.'); dot {
		case -1:
			return Origin{}, false
		default:
			candidate = candidate[:dot]
		}
	}
	return Origin{}, false
}

// ErrNotFound is returned by sources that are absent. Only optional sources
// suppress this error.
var ErrNotFound = errors.New("configuration source not found")

// ErrSourceChanged indicates that a filesystem source changed while it was
// being read, so no candidate snapshot was accepted.
var ErrSourceChanged = errors.New("configuration source changed during read")

// Origins are declarations, not an audit log. Replacement/deletion prunes
// descendant metadata; source overrides retain their own presence/state even
// if inconsistent with the tree.
func applyOrigins(
	origins map[string]Origin,
	tree map[string]any,
	overrides map[string]Origin,
	parent string,
	source SourceInfo,
) {
	for key, value := range tree {
		path := key
		if parent != "" {
			path = parent + "." + key
		}

		if _, deleted := value.(merge.Delete); deleted {
			delete(origins, path)
			deleteDescendants(origins, path)
		} else {
			origin := Origin{
				Source: source.Name, Sensitive: source.Sensitive, Present: true, State: Present,
			}
			if value == nil {
				origin.State = Null
			}
			if override, exists := overrides[path]; exists {
				if override.Source != "" {
					origin.Source = override.Source
				}
				origin.Sensitive = origin.Sensitive || override.Sensitive
				origin.Location = override.Location
				origin.Present = override.Present
				origin.State = override.State
			}
			origins[path] = origin
			if object, ok := value.(map[string]any); ok {
				applyOrigins(origins, object, overrides, path, source)
			} else {
				deleteDescendants(origins, path)
			}
		}
	}
}

func deleteDescendants(origins map[string]Origin, path string) {
	prefix := path + "."
	for candidate := range origins {
		if strings.HasPrefix(candidate, prefix) {
			delete(origins, candidate)
		}
	}
}

func cloneOrigins(origins map[string]Origin) map[string]Origin {
	clone := make(map[string]Origin, len(origins))
	for path, origin := range origins {
		clone[path] = origin
	}
	return clone
}

func cloneMap(value map[string]any) map[string]any {
	clone := make(map[string]any, len(value))
	for key, item := range value {
		clone[key] = cloneValue(item)
	}
	return clone
}

func cloneValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		return cloneMap(value)
	case []any:
		clone := make([]any, len(value))
		for index, item := range value {
			clone[index] = cloneValue(item)
		}
		return clone
	default:
		return value
	}
}

const (
	maxCanonicalTreeDepth = 64
	maxCanonicalTreeKeys  = 100_000
	maxCanonicalTreeItems = 100_000
)

type treeVisit struct {
	kind    reflect.Kind
	pointer uintptr
	length  int
}

type treeCanonicalizer struct {
	ctx      context.Context
	visiting map[treeVisit]bool
	keys     []struct{}
	items    int
}

func canonicalTree(ctx context.Context, tree map[string]any) (map[string]any, error) {
	canonicalizer := treeCanonicalizer{
		ctx: ctx, visiting: make(map[treeVisit]bool),
	}
	return canonicalizer.object(tree, "", 1)
}

func (c *treeCanonicalizer) object(
	tree map[string]any,
	parent string,
	depth int,
) (map[string]any, error) {
	if err := c.check(depth, parent); err != nil {
		return nil, err
	}
	visit := treeVisit{kind: reflect.Map, pointer: reflect.ValueOf(tree).Pointer()}
	if c.visiting[visit] {
		return nil, &TreeCycleError{Path: parent}
	}
	c.visiting[visit] = true
	defer delete(c.visiting, visit)

	keys := make([]string, 0, len(tree))
	for key := range tree {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	canonical := make(map[string]any, len(tree))
	for _, key := range keys {
		if err := c.ctx.Err(); err != nil {
			return nil, err
		}
		path := key
		if parent != "" {
			path = parent + "." + key
		}
		c.keys = append(c.keys, struct{}{})
		if len(c.keys) > maxCanonicalTreeKeys {
			return nil, &TreeLimitError{
				Path: path, Kind: "keys", Limit: maxCanonicalTreeKeys,
			}
		}
		converted, err := c.value(tree[key], path, depth+1)
		if err != nil {
			return nil, err
		}
		canonical[key] = converted
	}
	return canonical, nil
}

func (c *treeCanonicalizer) value(value any, path string, depth int) (any, error) {
	if err := c.check(depth, path); err != nil {
		return nil, err
	}
	switch value := value.(type) {
	case nil:
		return nil, nil
	case map[string]any:
		return c.object(value, path, depth)
	case []any:
		visit := treeVisit{kind: reflect.Slice, pointer: reflect.ValueOf(value).Pointer(), length: len(value)}
		if c.visiting[visit] {
			return nil, &TreeCycleError{Path: path}
		}
		if len(value) > maxCanonicalTreeItems-c.items {
			return nil, &TreeLimitError{
				Path: path, Kind: "items", Limit: maxCanonicalTreeItems,
			}
		}
		c.items += len(value)
		c.visiting[visit] = true
		defer delete(c.visiting, visit)
		items := make([]any, len(value))
		for index, item := range value {
			converted, err := c.value(item, fmt.Sprintf("%s[%d]", path, index), depth+1)
			if err != nil {
				return nil, err
			}
			items[index] = converted
		}
		return items, nil
	case merge.Delete:
		return value, nil
	}

	typeOf := reflect.TypeOf(value)
	switch typeOf.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.String:
		return value, nil
	default:
		return nil, &TreeValueError{Path: path, Type: typeOf.String()}
	}
}

func (c *treeCanonicalizer) check(depth int, path string) error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	if depth > maxCanonicalTreeDepth {
		return &TreeLimitError{
			Path: path, Kind: "depth", Limit: maxCanonicalTreeDepth,
		}
	}
	return nil
}

func cloneTyped[T any](value T) T {
	cloned := cloneReflect(reflect.ValueOf(value))
	if !cloned.IsValid() {
		var zero T
		return zero
	}
	return cloned.Interface().(T)
}

type cloneVisit struct {
	typeOf  reflect.Type
	pointer uintptr
	length  int
}

func validateSnapshotValue(
	value reflect.Value,
	path string,
	visiting map[cloneVisit]bool,
) error {
	if !value.IsValid() {
		return nil
	}
	if (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) && value.IsNil() {
		return nil
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice:
		visit := cloneVisit{typeOf: value.Type(), pointer: value.Pointer()}
		if value.Kind() == reflect.Slice {
			visit.length = value.Len()
		}
		if visiting[visit] {
			return snapshotValueError(path, value.Type())
		}
		visiting[visit] = true
		defer delete(visiting, visit)
	}
	if value.CanInterface() {
		if validator, ok := value.Interface().(interface {
			validateConfigValue(string, map[cloneVisit]bool) error
		}); ok {
			return validator.validateConfigValue(path, visiting)
		}
		if _, ok := value.Interface().(interface{ cloneConfigValue() any }); ok {
			return nil
		}
		if _, ok := value.Interface().(time.Time); ok {
			return nil
		}
	}

	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		return validateSnapshotValue(value.Elem(), path, visiting)
	case reflect.Map:
		iterator := value.MapRange()
		for iterator.Next() {
			if err := validateSnapshotValue(iterator.Key(), path, visiting); err != nil {
				return err
			}
			if err := validateSnapshotValue(iterator.Value(), path, visiting); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if err := validateSnapshotValue(value.Index(index), path, visiting); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			definition := value.Type().Field(index)
			fieldPath := joinConfigPath(path, definition)
			if !definition.IsExported() {
				if typeContainsMutableReferences(definition.Type) {
					return snapshotValueError(fieldPath, definition.Type)
				}
				continue
			}
			if err := validateSnapshotValue(value.Field(index), fieldPath, visiting); err != nil {
				return err
			}
		}
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return snapshotValueError(path, value.Type())
	}
	return nil
}

func snapshotValueError(path string, typeOf reflect.Type) error {
	return &SnapshotValueError{Path: path, Type: typeOf.String()}
}

func typeContainsMutableReferences(typeOf reflect.Type) bool {
	switch typeOf.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface,
		reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return true
	case reflect.Array:
		return typeContainsMutableReferences(typeOf.Elem())
	case reflect.Struct:
		for _, definition := range reflect.VisibleFields(typeOf) {
			if typeContainsMutableReferences(definition.Type) {
				return true
			}
		}
	}
	return false
}

func joinConfigPath(parent string, definition reflect.StructField) string {
	name, _, _ := strings.Cut(definition.Tag.Get("config"), ",")
	if name == "" {
		name = strings.ToLower(definition.Name)
	}
	if parent == "" {
		return name
	}
	return parent + "." + name
}

// Clone only previously admitted state; supported containers are independent.
func cloneReflect(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}
	if (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) && value.IsNil() {
		return reflect.Zero(value.Type())
	}
	if value.CanInterface() {
		if cloner, ok := value.Interface().(interface{ cloneConfigValue() any }); ok {
			cloned := reflect.ValueOf(cloner.cloneConfigValue())
			if cloned.IsValid() && cloned.Type().AssignableTo(value.Type()) {
				return cloned
			}
			if value.Kind() == reflect.Pointer {
				if cloned.IsValid() {
					if cloned.Type().AssignableTo(value.Type().Elem()) {
						pointer := reflect.New(value.Type().Elem())
						pointer.Elem().Set(cloned)
						return pointer
					}
				}
			}
		}
	}

	switch value.Kind() {
	case reflect.Pointer:
		clone := reflect.New(value.Type().Elem())
		clone.Elem().Set(cloneReflect(value.Elem()))
		return clone
	case reflect.Interface:
		clone := cloneReflect(value.Elem())
		result := reflect.New(value.Type()).Elem()
		result.Set(clone)
		return result
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		clone := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			clone.SetMapIndex(cloneReflect(iterator.Key()), cloneReflect(iterator.Value()))
		}
		return clone
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		clone := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			clone.Index(index).Set(cloneReflect(value.Index(index)))
		}
		return clone
	case reflect.Array:
		clone := reflect.New(value.Type()).Elem()
		for index := 0; index < value.Len(); index++ {
			clone.Index(index).Set(cloneReflect(value.Index(index)))
		}
		return clone
	case reflect.Struct:
		clone := reflect.New(value.Type()).Elem()
		clone.Set(value)
		for index := 0; index < value.NumField(); index++ {
			if value.Type().Field(index).IsExported() {
				clone.Field(index).Set(cloneReflect(value.Field(index)))
			}
		}
		return clone
	default:
		return value
	}
}
