package organizer

import (
	"io/fs"
	"syscall"
)

// isCloudFile checks if the file is a dataless cloud placeholder (e.g. iCloud)
// This implementation is for Darwin (macOS) which uses the SF_DATALESS flag.
func (o *Organizer) isCloudFile(info fs.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	// Check for SF_DATALESS (0x40000000)
	// This flag indicates the file's content is not locally present
	return (stat.Flags & 0x40000000) != 0
}
