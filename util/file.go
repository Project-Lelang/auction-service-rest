package util

import "path"

// GetFilenameFromPath returns the base filename from a storage path.
// e.g. "user/uuid/identity.jpg" → "identity.jpg"
func GetFilenameFromPath(storagePath string) string {
	return path.Base(storagePath)
}

// StringInSlice reports whether s is present in the slice.
func StringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
