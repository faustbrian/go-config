//go:build linux

package sourceio

import (
	"fmt"
	"io/fs"
	"syscall"
)

// Linux change time detects some same-size/mtime-preserving writes; it is
// optional metadata, not a generation/authenticity guarantee.
func systemChangeToken(info fs.FileInfo) (string, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", false
	}
	return fmt.Sprint(stat.Ctim), true
}
