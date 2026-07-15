// Package sourceio provides immutable, bounded byte and fs.FS inputs for
// format source packages.
package sourceio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"

	config "github.com/faustbrian/go-config"
)

// Input is a repeatable immutable byte or filesystem input.
type Input struct {
	data []byte
	open func() (io.ReadCloser, error)
}

// Bytes returns an immutable copy-backed input.
func Bytes(data []byte) Input { return Input{data: bytes.Clone(data)} }

// FromFS returns a repeatable input for path.
func FromFS(filesystem fs.FS, path string) (Input, error) {
	if filesystem == nil {
		return Input{}, errors.New("source filesystem must not be nil")
	}
	if !fs.ValidPath(path) {
		return Input{}, fmt.Errorf("source path %q is invalid", path)
	}
	return Input{open: func() (io.ReadCloser, error) {
		file, err := filesystem.Open(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("%w: %w", config.ErrNotFound, err)
			}
			return nil, err
		}
		return file, nil
	}}, nil
}

// Read returns at most maxBytes and checks ctx between reader operations.
func (i Input) Read(ctx context.Context, maxBytes int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if i.open == nil {
		if int64(len(i.data)) > maxBytes {
			return nil, fmt.Errorf("source input exceeds %d byte limit", maxBytes)
		}
		return bytes.Clone(i.data), nil
	}

	reader, err := i.open()
	if err != nil {
		return nil, err
	}
	data, readErr := Read(ctx, reader, maxBytes)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

// Read reads a bounded stream and checks ctx between reader operations.
func Read(ctx context.Context, reader io.Reader, maxBytes int64) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("source reader must not be nil")
	}
	limit := maxBytes + 1
	if maxBytes == math.MaxInt64 {
		limit = maxBytes
	}
	limited := io.LimitReader(reader, limit)
	var buffer bytes.Buffer
	chunk := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		count, err := limited.Read(chunk)
		if count > 0 {
			_, _ = buffer.Write(chunk[:count])
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	if int64(buffer.Len()) > maxBytes {
		return nil, fmt.Errorf("source input exceeds %d byte limit", maxBytes)
	}
	return buffer.Bytes(), nil
}
