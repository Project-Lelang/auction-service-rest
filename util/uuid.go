package util

import "github.com/google/uuid"

func NewUuid() string {
	return uuid.New().String()
}

func IsUuid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
