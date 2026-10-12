//go:build !darwin && !linux

package sourceio

import "io/fs"

// Platforms without a portable change-time field rely on ordinary metadata
// and any caller-provided GenerationFile token; no synthetic identity is invented.
func systemChangeToken(fs.FileInfo) (string, bool) {
	return "", false
}
