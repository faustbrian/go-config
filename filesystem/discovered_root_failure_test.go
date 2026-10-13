package filesystem

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoveredFSRetainsNativeRootOpenFailure(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "owned")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(directory, "config.json")
	if err := os.WriteFile(filename, []byte("owned"), 0o600); err != nil {
		t.Fatal(err)
	}
	filesystem, err := newDiscoveredFS(directory)
	if err != nil {
		t.Fatal(err)
	}
	file, err := filesystem.Open("config.json")
	if err != nil {
		t.Fatalf("admitted regular file: %v", err)
	}
	data, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil || closeErr != nil || string(data) != "owned" {
		t.Fatalf("admitted read=%q, %v; close=%v", data, readErr, closeErr)
	}

	// Remove only the owned directory after construction, before OpenRoot. This
	// exercises the actual default operation, not an injected operation's error.
	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(directory); err != nil {
		t.Fatal(err)
	}
	file, err = filesystem.Open("config.json")
	var original *fs.PathError
	if file != nil || !errors.Is(err, fs.ErrNotExist) || !errors.As(err, &original) || original.Path != directory {
		t.Fatalf("removed bound root: file=%v error=%v; want no handle and original root path/cause", file, err)
	}
}
