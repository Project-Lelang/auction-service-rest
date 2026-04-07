package constant

import "errors"

var (
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrForbidden        = errors.New("forbidden")

	ErrNoData              = errors.New("no data")
	ErrDuplicateData       = errors.New("duplicate data")
	ErrForeignKeyViolation = errors.New("data in use or id doesn't exist")

	ErrRequestIdNotFound = errors.New("request id not found")

	ErrNotImplemented = errors.New("not implemented")

	ErrDbtxNotFound     = errors.New("db transaction not found")
	ErrDbtxAlreadyExist = errors.New("db transaction already exist")
)
