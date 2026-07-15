package sourceio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"math"
	"testing"
	"testing/fstest"

	config "github.com/faustbrian/go-config"
)

func TestBytesIsImmutableRepeatableAndBounded(t *testing.T) {
	t.Parallel()

	data := []byte("original")
	input := Bytes(data)
	data[0] = 'X'
	first, err := input.Read(context.Background(), 8)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	first[0] = 'Y'
	second, err := input.Read(context.Background(), 8)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(second) != "original" {
		t.Fatalf("second Read() = %q", second)
	}
	if _, err := input.Read(context.Background(), 7); err == nil {
		t.Fatal("Read() error = nil, want byte limit")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := input.Read(ctx, 8); !errors.Is(err, context.Canceled) {
		t.Fatalf("Read() error = %v, want context.Canceled", err)
	}
}

func TestFromFSValidatesAndReopensFiles(t *testing.T) {
	t.Parallel()

	if _, err := FromFS(nil, "config.json"); err == nil {
		t.Fatal("FromFS(nil) error = nil")
	}
	if _, err := FromFS(fstest.MapFS{}, "../config.json"); err == nil {
		t.Fatal("FromFS(invalid path) error = nil")
	}
	missing, err := FromFS(fstest.MapFS{}, "missing.json")
	if err != nil {
		t.Fatalf("FromFS() error = %v", err)
	}
	if _, err := missing.Read(context.Background(), 100); !errors.Is(err, config.ErrNotFound) {
		t.Fatalf("Read() error = %v, want ErrNotFound", err)
	}

	filesystem := fstest.MapFS{"config.json": {Data: []byte("first")}}
	input, err := FromFS(filesystem, "config.json")
	if err != nil {
		t.Fatalf("FromFS() error = %v", err)
	}
	first, err := input.Read(context.Background(), 100)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	filesystem["config.json"].Data = []byte("second")
	second, err := input.Read(context.Background(), 100)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(first) != "first" || string(second) != "second" {
		t.Fatalf("reads = %q, %q", first, second)
	}
}

func TestFromFSPreservesNonAbsenceOpenError(t *testing.T) {
	t.Parallel()

	want := errors.New("permission denied")
	input, err := FromFS(errorFS{err: want}, "config.json")
	if err != nil {
		t.Fatalf("FromFS() error = %v", err)
	}
	if _, err := input.Read(context.Background(), 100); !errors.Is(err, want) {
		t.Fatalf("Read() error = %v, want %v", err, want)
	}
}

func TestInputReadHandlesReadAndCloseFailures(t *testing.T) {
	t.Parallel()

	readFailure := errors.New("read failure")
	closeFailure := errors.New("close failure")
	tests := map[string]struct {
		reader io.ReadCloser
		want   error
	}{
		"read": {
			reader: &readCloser{Reader: errorReader{err: readFailure}, closeErr: closeFailure},
			want:   readFailure,
		},
		"close": {
			reader: &readCloser{Reader: bytes.NewBufferString("value"), closeErr: closeFailure},
			want:   closeFailure,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			input := Input{open: func() (io.ReadCloser, error) { return test.reader, nil }}
			_, err := input.Read(context.Background(), 100)
			if !errors.Is(err, test.want) {
				t.Fatalf("Read() error = %v, want %v", err, test.want)
			}
		})
	}

	openFailure := errors.New("open failure")
	input := Input{open: func() (io.ReadCloser, error) { return nil, openFailure }}
	if _, err := input.Read(context.Background(), 100); !errors.Is(err, openFailure) {
		t.Fatalf("Read() error = %v, want open failure", err)
	}
}

func TestReadValidatesBoundsErrorsAndCancellation(t *testing.T) {
	t.Parallel()

	if _, err := Read(context.Background(), nil, 1); err == nil {
		t.Fatal("Read(nil) error = nil")
	}
	if _, err := Read(context.Background(), bytes.NewBufferString("too large"), 3); err == nil {
		t.Fatal("Read() error = nil, want byte limit")
	}
	if got, err := Read(context.Background(), bytes.NewBufferString("value"), math.MaxInt64); err != nil || string(got) != "value" {
		t.Fatalf("Read(MaxInt64) = %q, %v", got, err)
	}

	want := errors.New("reader failed")
	if _, err := Read(context.Background(), errorReader{err: want}, 100); !errors.Is(err, want) {
		t.Fatalf("Read() error = %v, want %v", err, want)
	}

	ctx, cancel := context.WithCancel(context.Background())
	reader := &cancelReader{cancel: cancel}
	if _, err := Read(ctx, reader, 100); !errors.Is(err, context.Canceled) {
		t.Fatalf("Read() error = %v, want context.Canceled", err)
	}
}

type readCloser struct {
	io.Reader
	closeErr error
}

func (r *readCloser) Close() error { return r.closeErr }

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

type cancelReader struct {
	cancel context.CancelFunc
	read   bool
}

type errorFS struct{ err error }

func (f errorFS) Open(string) (fs.File, error) { return nil, f.err }

func (r *cancelReader) Read(buffer []byte) (int, error) {
	if r.read {
		return 0, io.EOF
	}
	r.read = true
	buffer[0] = 'x'
	r.cancel()
	return 1, nil
}
