package filesystem

import "errors"

var (
	ErrFileNotExist error = errors.New("file not exist")

	ErrPathIsNotAllowed   error = errors.New("path is not allowed")
	ErrPathIsNotDirectory error = errors.New("path is not a directory")
)
