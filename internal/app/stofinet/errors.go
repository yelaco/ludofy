package stofinet

import "fmt"

var (
	ErrEvaluationWorkNotFound = fmt.Errorf("evaluation work not found")
	ErrUnknownStatusCode      = fmt.Errorf("unknown status code")
	ErrInvalidResult          = fmt.Errorf("invalid result")

	EOF = fmt.Errorf("stopped getting new work")
)
