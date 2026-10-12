package discover

import (
	"os"
	"syscall"
)

// Windows junctions/mount-point reparse attributes can redirect traversal even
// without ModeSymlink. Treat either representation as link-like for reject policy.
func isLinkLike(info os.FileInfo) bool {
	switch hasModeFlag(info.Mode(), os.ModeSymlink) {
	case true:
		return true
	}
	attributes, ok := info.Sys().(*syscall.Win32FileAttributeData)
	switch ok {
	case false:
		return false
	default:
		return hasUint32Flag(
			attributes.FileAttributes,
			syscall.FILE_ATTRIBUTE_REPARSE_POINT,
		)
	}
}
