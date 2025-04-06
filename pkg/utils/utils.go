package utils

import (
	"github.com/google/uuid"
)

func IsClosed[T any](ch <-chan T) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func GenerateUUID() string {
	return uuid.NewString()
}
