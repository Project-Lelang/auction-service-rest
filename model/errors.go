package model

import "errors"

var (
	ErrDbtxNotFound     = errors.New("db transaction not found")
	ErrDbtxAlreadyExist = errors.New("db transaction already exist")
)
