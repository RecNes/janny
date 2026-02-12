//go:build !darwin

package organizer

import (
	"io/fs"
)

// isCloudFile checks if the file is a dataless cloud placeholder.
// For non-Darwin systems, we currently don't support checking system specific flags
// for cloud placeholders, so we always return false.
func (o *Organizer) isCloudFile(info fs.FileInfo) bool {
	return false
}
