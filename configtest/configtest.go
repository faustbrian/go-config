// Package configtest provides deterministic configuration fixtures and assertions
// without mutating process-global state. NewSource copies canonical map/slice
// containers; unsupported custom references are borrowed. Filesystem returns a
// fresh, mutable fstest.MapFS, not an immutable filesystem enforcement mechanism.
// These helpers are for tests, not source authentication or hostile-input admission;
// diagnostic names/paths and test failure output must remain non-secret.
package configtest

import (
	"context"
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"testing/fstest"

	config "github.com/faustbrian/go-config/v3"
)

// TestingT is the subset of testing.T used by configtest helpers.
type TestingT interface {
	// Helper attributes test failures to the caller of the fixture/assertion helper.
	Helper()
	// Fatalf reports failure and must stop the current test, as testing.T does.
	Fatalf(format string, args ...any)
}

type source struct {
	info     config.SourceInfo
	document config.Document
	err      error
}

// NewSource captures copied canonical map/slice containers and returns fresh
// documents per load. It does not admit arbitrary input; custom references
// outside that container shape remain shared and must be immutable.
func NewSource(info config.SourceInfo, document config.Document) config.Source {
	return &source{info: info, document: cloneDocument(document)}
}

// FailingSource returns cause on a non-canceled load; canceled contexts take
// precedence. The cause is borrowed, not redacted by this test helper.
func FailingSource(info config.SourceInfo, cause error) config.Source {
	return &source{info: info, err: cause}
}

func (s *source) Info() config.SourceInfo { return s.info }

func (s *source) Load(ctx context.Context) (config.Document, error) {
	if err := ctx.Err(); err != nil {
		return config.Document{}, err
	}
	if s.err != nil {
		return config.Document{}, s.err
	}
	return cloneDocument(s.document), nil
}

// Environment converts a map into a stable, sorted environment snapshot.
func Environment(values map[string]string) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	environment := make([]string, len(names))
	for index, name := range names {
		environment[index] = name + "=" + values[name]
	}
	return environment
}

// Filesystem returns a fresh in-memory filesystem with 0600-mode copied files.
// The returned fstest.MapFS is mutable through a type assertion; ownership and
// containment/permission enforcement are not supplied by this fixture helper.
func Filesystem(files map[string]string) fs.FS {
	fixture := make(fstest.MapFS, len(files))
	for name, contents := range files {
		fixture[name] = &fstest.MapFile{Data: []byte(contents), Mode: 0o600}
	}
	return fixture
}

// MustPlan constructs a plan or fails the current test.
func MustPlan(t TestingT, sources ...config.Source) config.Plan {
	t.Helper()
	plan, err := config.NewPlan(sources...)
	if err != nil {
		t.Fatalf("configtest.MustPlan: %v", err)
	}
	return plan
}

// MustLoad loads a typed snapshot or fails the current test.
func MustLoad[T any](t TestingT, ctx context.Context, plan config.Plan) *config.Snapshot[T] {
	t.Helper()
	snapshot, err := config.Load[T](ctx, plan)
	if err != nil {
		t.Fatalf("configtest.MustLoad: %v", err)
	}
	return snapshot
}

// DiffSecrets returns an empty string when values match or a redacted diff.
func DiffSecrets(got, want config.Secret) string {
	if got == want {
		return ""
	}
	return fmt.Sprintf("- %s\n+ %s", got, want)
}

// AssertOrigin compares safe provenance for path.
func AssertOrigin(
	t TestingT,
	snapshot interface {
		// Origin returns declared metadata for path and whether it exists; no value is exposed.
		Origin(string) (config.Origin, bool)
	},
	path string,
	want config.Origin,
) {
	t.Helper()
	got, ok := snapshot.Origin(path)
	if !ok {
		t.Fatalf("configtest.AssertOrigin: path %q has no provenance", path)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("configtest.AssertOrigin: path %q = %s, want %s", path, origin(got), origin(want))
	}
}

func cloneDocument(document config.Document) config.Document {
	return config.Document{
		Tree:    cloneMap(document.Tree),
		Origins: cloneOrigins(document.Origins),
	}
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

func cloneOrigins(origins map[string]config.Origin) map[string]config.Origin {
	clone := make(map[string]config.Origin, len(origins))
	for path, value := range origins {
		clone[path] = value
	}
	return clone
}

func origin(value config.Origin) string {
	return fmt.Sprintf(
		"{source:%q location:%q sensitive:%t present:%t state:%d}",
		value.Source,
		value.Location,
		value.Sensitive,
		value.Present,
		value.State,
	)
}
