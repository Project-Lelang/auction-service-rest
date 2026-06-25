package filesystem

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func guessMimeType(r io.ReadSeeker) (string, error) {
	buf := make([]byte, 30720)
	n, err := r.Read(buf)
	if err != nil {
		return "", err
	}
	r.Seek(0, io.SeekStart)
	mType := mimetype.Detect(buf[:n])
	return mType.String(), nil
}

func ParsePath(s string) string {
	return filepath.ToSlash(s)
}

func ParsePathTrim(s string) string {
	return strings.Trim(ParsePath(s), "/")
}
