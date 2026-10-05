package filesystem

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	config "github.com/faustbrian/go-config/v2"
	"github.com/faustbrian/go-config/v2/discover"
)

func TestReaderValidatesEveryOptionBoundary(t *testing.T) {
	t.Parallel()

	open := func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(nil)), nil
	}
	tests := map[string]Options{
		"empty name":     {Format: FormatJSON},
		"auto format":    {Name: "reader", Format: FormatAuto},
		"invalid format": {Name: "reader", Format: FormatTOML + 1},
		"negative bytes": {Name: "reader", Format: FormatJSON, Limits: Limits{MaxBytes: -1}},
		"negative depth": {Name: "reader", Format: FormatJSON, Limits: Limits{MaxDepth: -1}},
		"negative keys":  {Name: "reader", Format: FormatJSON, Limits: Limits{MaxKeys: -1}},
	}
	for name, options := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Reader(open, options); err == nil {
				t.Fatal("Reader() error = nil")
			}
		})
	}
}

func TestReaderDispatchesYAMLAndTOMLAndPreservesMetadata(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		format Format
		data   string
	}{
		"yaml": {format: FormatYAML, data: "value: yaml\n"},
		"toml": {format: FormatTOML, data: `value = "toml"`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := Reader(func(context.Context) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewBufferString(test.data)), nil
			}, Options{
				Name: name, Priority: 42, Sensitive: true, Optional: true,
				Format: test.format,
			})
			if err != nil {
				t.Fatalf("Reader() error = %v", err)
			}
			if got := source.Info(); got != (config.SourceInfo{
				Name: name, Priority: 42, Sensitive: true, Optional: true,
			}) {
				t.Fatalf("Info() = %#v", got)
			}
			document, err := source.Load(context.Background())
			if err != nil || document.Tree["value"] != name {
				t.Fatalf("Load() = %#v, %v", document, err)
			}
		})
	}
}

func TestReaderRejectsNilFactoryResultAndMalformedDocument(t *testing.T) {
	t.Parallel()

	for name, open := range map[string]OpenFunc{
		"nil reader": func(context.Context) (io.ReadCloser, error) { return nil, nil },
		"malformed": func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBufferString(`{"broken":`)), nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := Reader(open, Options{Name: "reader", Format: FormatJSON})
			if err != nil {
				t.Fatalf("Reader() error = %v", err)
			}
			if _, err := source.Load(context.Background()); err == nil {
				t.Fatal("Load() error = nil")
			}
		})
	}
}

func TestFromFSRejectsInvalidPathFormatAndLimits(t *testing.T) {
	t.Parallel()

	if _, err := FromFS(emptyFilesystem{}, "../config.json", Options{Name: "file"}); err == nil {
		t.Fatal("FromFS(invalid path) error = nil")
	}
	if _, err := FromFS(emptyFilesystem{}, "config.json", Options{
		Name: "file", Format: FormatTOML + 1,
	}); err == nil {
		t.Fatal("FromFS(invalid format) error = nil")
	}
	if _, err := FromFS(emptyFilesystem{}, "config.json", Options{
		Name: "file", Limits: Limits{MaxDepth: -1},
	}); err == nil {
		t.Fatal("FromFS(invalid limits) error = nil")
	}
}

func TestFromDiscoveredValidatesPathsAndFallsBackToLexicalPath(t *testing.T) {
	t.Parallel()

	for name, result := range map[string]discover.Result{
		"missing both":    {},
		"missing lexical": {ResolvedPath: "config.json"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := FromDiscovered(result, Options{Name: "file"}); err == nil {
				t.Fatal("FromDiscovered() error = nil")
			}
		})
	}

	directory := t.TempDir()
	path := filepath.Join(directory, "config.json")
	if err := os.WriteFile(path, []byte(`{"value":"fallback"}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	source, err := FromDiscovered(discover.Result{Path: path}, Options{Name: "file"})
	if err != nil {
		t.Fatalf("FromDiscovered() error = %v", err)
	}
	document, err := source.Load(context.Background())
	if err != nil || document.Tree["value"] != "fallback" {
		t.Fatalf("Load() = %#v, %v", document, err)
	}

	missing := filepath.Join(t.TempDir(), "missing", "config.json")
	if _, err := FromDiscovered(discover.Result{
		Path: missing, ResolvedPath: missing,
	}, Options{Name: "file"}); err == nil {
		t.Fatal("FromDiscovered(missing directory) error = nil")
	}
}

func TestPathConstructorsPropagateAbsolutePathFailure(t *testing.T) {
	t.Parallel()

	failure := errors.New("absolute path failure")
	absolutePath := func(string) (string, error) { return "", failure }
	if _, err := fromPath("config.json", Options{Name: "file"}, absolutePath); !errors.Is(err, failure) {
		t.Fatal("FromPath() error = nil")
	}
	if _, err := fromDiscovered(discover.Result{
		Path: "config.json", ResolvedPath: "config.json",
	}, Options{Name: "file"}, absolutePath); !errors.Is(err, failure) {
		t.Fatal("FromDiscovered() error = nil")
	}
}

func TestNewDiscoveredFSRejectsUnavailableOrNonDirectoryRoot(t *testing.T) {
	t.Parallel()

	failure := errors.New("stat failure")
	for name, operations := range map[string]discoveredFSOperations{
		"stat failure": {
			stat: func(string) (fs.FileInfo, error) { return nil, failure },
		},
		"not directory": {
			stat: func(string) (fs.FileInfo, error) {
				return discoveredFileInfo{id: "file", mode: 0o600}, nil
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := newDiscoveredFSWithOperations("root", operations); err == nil {
				t.Fatal("newDiscoveredFSWithOperations() error = nil")
			}
		})
	}
}

func TestDiscoveredFSRejectsOpenBoundaryFailures(t *testing.T) {
	t.Parallel()

	failure := errors.New("filesystem failure")
	tests := map[string]struct {
		configure func(*stubDiscoveredRoot)
		openError error
		want      error
	}{
		"open root": {openError: failure, want: failure},
		"stat root": {
			configure: func(root *stubDiscoveredRoot) { root.statError = failure }, want: failure,
		},
		"replaced root": {
			configure: func(root *stubDiscoveredRoot) {
				root.directory = discoveredFileInfo{id: "replacement", mode: fs.ModeDir}
			},
			want: config.ErrSourceChanged,
		},
		"lstat before": {
			configure: func(root *stubDiscoveredRoot) { root.beforeError = failure }, want: failure,
		},
		"symlink before": {
			configure: func(root *stubDiscoveredRoot) {
				root.before = discoveredFileInfo{id: "file", mode: fs.ModeSymlink}
			},
			want: discover.ErrSymlink,
		},
		"non-regular before": {
			configure: func(root *stubDiscoveredRoot) {
				root.before = discoveredFileInfo{id: "file", mode: fs.ModeDir}
			},
			want: config.ErrSourceChanged,
		},
		"open file": {
			configure: func(root *stubDiscoveredRoot) { root.openError = failure }, want: failure,
		},
		"stat opened file": {
			configure: func(root *stubDiscoveredRoot) { root.file.statError = failure }, want: failure,
		},
		"lstat after": {
			configure: func(root *stubDiscoveredRoot) { root.afterError = failure }, want: failure,
		},
		"symlink after": {
			configure: func(root *stubDiscoveredRoot) {
				root.after = discoveredFileInfo{id: "file", mode: fs.ModeSymlink}
			},
			want: config.ErrSourceChanged,
		},
		"opened identity changed": {
			configure: func(root *stubDiscoveredRoot) {
				root.file.info = discoveredFileInfo{id: "other", mode: 0o600}
			},
			want: config.ErrSourceChanged,
		},
		"path identity changed": {
			configure: func(root *stubDiscoveredRoot) {
				root.after = discoveredFileInfo{id: "other", mode: 0o600}
			},
			want: config.ErrSourceChanged,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := newStubDiscoveredRoot()
			if test.configure != nil {
				test.configure(root)
			}
			operations := discoveredFSOperations{
				stat: func(string) (fs.FileInfo, error) {
					return discoveredFileInfo{id: "root", mode: fs.ModeDir}, nil
				},
				openRoot: func(string) (discoveredRoot, error) {
					if test.openError != nil {
						return nil, test.openError
					}
					return root, nil
				},
				sameFile: func(left, right fs.FileInfo) bool {
					return left.Sys() == right.Sys()
				},
			}
			filesystem, err := newDiscoveredFSWithOperations("root", operations)
			if err != nil {
				t.Fatalf("newDiscoveredFSWithOperations() error = %v", err)
			}
			file, err := filesystem.Open("config.json")
			if file != nil {
				_ = file.Close()
			}
			if !errors.Is(err, test.want) {
				t.Fatalf("Open() error = %v, want %v", err, test.want)
			}
			if test.openError == nil && !root.closed {
				t.Fatal("Open() did not close root")
			}
		})
	}
}

func TestLocationSourcePreservesExistingOriginsAndMarksNull(t *testing.T) {
	t.Parallel()

	inner := staticFilesystemSource{
		info: config.SourceInfo{Name: "inner"},
		document: config.Document{
			Tree: map[string]any{
				"existing": "value",
				"null":     nil,
				"nested":   map[string]any{"leaf": true},
			},
			Origins: map[string]config.Origin{
				"existing": {Present: true, State: config.Defaulted},
			},
		},
	}
	source := locationSource{source: inner, location: "/config.json"}
	if source.Info() != inner.info {
		t.Fatalf("Info() = %#v", source.Info())
	}
	document, err := source.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := map[string]config.Origin{
		"existing":    {Present: true, State: config.Defaulted, Location: "/config.json"},
		"null":        {Present: true, State: config.Null, Location: "/config.json"},
		"nested":      {Present: true, State: config.Present, Location: "/config.json"},
		"nested.leaf": {Present: true, State: config.Present, Location: "/config.json"},
	}
	if !reflect.DeepEqual(document.Origins, want) {
		t.Fatalf("Origins = %#v, want %#v", document.Origins, want)
	}

	failure := errors.New("load failure")
	source.source = staticFilesystemSource{err: failure}
	if _, err := source.Load(context.Background()); !errors.Is(err, failure) {
		t.Fatalf("Load() error = %v, want load failure", err)
	}
}

func TestReaderSourceRejectsInternallyInvalidFormat(t *testing.T) {
	t.Parallel()

	source := &readerSource{
		open: func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(nil)), nil
		},
		info:    config.SourceInfo{Name: "reader"},
		options: Options{Name: "reader", Format: FormatTOML + 1},
	}
	if _, err := source.Load(context.Background()); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestReaderSourceHonorsAlreadyCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := &readerSource{
		open: func(context.Context) (io.ReadCloser, error) {
			t.Fatal("reader factory called after cancellation")
			return nil, nil
		},
		info:    config.SourceInfo{Name: "reader"},
		options: Options{Name: "reader", Format: FormatJSON},
	}
	if _, err := source.Load(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Load() error = %v, want context.Canceled", err)
	}
}

type emptyFilesystem struct{}

func (emptyFilesystem) Open(string) (fs.File, error) { return nil, fs.ErrNotExist }

type discoveredFileInfo struct {
	id   string
	mode fs.FileMode
}

func (info discoveredFileInfo) Name() string       { return info.id }
func (info discoveredFileInfo) Size() int64        { return 0 }
func (info discoveredFileInfo) Mode() fs.FileMode  { return info.mode }
func (info discoveredFileInfo) ModTime() time.Time { return time.Time{} }
func (info discoveredFileInfo) IsDir() bool        { return info.mode.IsDir() }
func (info discoveredFileInfo) Sys() any           { return info.id }

type stubDiscoveredFile struct {
	info      fs.FileInfo
	statError error
	closed    bool
}

func (file *stubDiscoveredFile) Stat() (fs.FileInfo, error) { return file.info, file.statError }
func (file *stubDiscoveredFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (file *stubDiscoveredFile) Close() error {
	file.closed = true
	return nil
}

type stubDiscoveredRoot struct {
	directory   fs.FileInfo
	before      fs.FileInfo
	after       fs.FileInfo
	file        *stubDiscoveredFile
	statError   error
	beforeError error
	afterError  error
	openError   error
	lstatCalls  int
	closed      bool
}

func newStubDiscoveredRoot() *stubDiscoveredRoot {
	file := discoveredFileInfo{id: "file", mode: 0o600}
	return &stubDiscoveredRoot{
		directory: discoveredFileInfo{id: "root", mode: fs.ModeDir},
		before:    file,
		after:     file,
		file:      &stubDiscoveredFile{info: file},
	}
}

func (root *stubDiscoveredRoot) Open(string) (fs.File, error) {
	return root.file, root.openError
}
func (root *stubDiscoveredRoot) Stat(string) (fs.FileInfo, error) {
	return root.directory, root.statError
}
func (root *stubDiscoveredRoot) Lstat(string) (fs.FileInfo, error) {
	root.lstatCalls++
	if root.lstatCalls == 1 {
		return root.before, root.beforeError
	}
	return root.after, root.afterError
}
func (root *stubDiscoveredRoot) Close() error {
	root.closed = true
	return nil
}

type staticFilesystemSource struct {
	info     config.SourceInfo
	document config.Document
	err      error
}

func (s staticFilesystemSource) Info() config.SourceInfo { return s.info }
func (s staticFilesystemSource) Load(context.Context) (config.Document, error) {
	return s.document, s.err
}
