package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s (value: %v): %s", e.Field, e.Value, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (errs ValidationErrors) Error() string {
	if len(errs) == 0 {
		return "no validation errors"
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}

	var msgs []string
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("multiple validation errors: [%s]", strings.Join(msgs, "; "))
}
