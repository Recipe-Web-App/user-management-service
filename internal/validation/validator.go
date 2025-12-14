// Package validation provides request validation using go-playground/validator.
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// jsonTagParts is the number of parts to split a JSON tag into (name and options).
const jsonTagParts = 2

// ErrValidation is the base error for validation failures.
var ErrValidation = errors.New("validation error")

// validationMessages maps validation tags to their error message templates.
// Messages with %s will have the parameter substituted.
var validationMessages = map[string]string{
	"required": "is required",
	"email":    "must be a valid email address",
	"url":      "must be a valid URL",
	"uuid":     "must be a valid UUID",
	"alphanum": "must contain only alphanumeric characters",
	"alpha":    "must contain only alphabetic characters",
	"numeric":  "must be numeric",
}

// parameterizedMessages maps validation tags to their parameterized message formats.
var parameterizedMessages = map[string]string{
	"min":   "must be at least %s characters",
	"max":   "must be at most %s characters",
	"len":   "must be exactly %s characters",
	"oneof": "must be one of: %s",
	"gte":   "must be greater than or equal to %s",
	"lte":   "must be less than or equal to %s",
	"gt":    "must be greater than %s",
	"lt":    "must be less than %s",
}

// Validator wraps the go-playground/validator with custom error formatting.
type Validator struct {
	validate *validator.Validate
}

// ValidationError represents a validation error with field details.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}

	return strings.Join(msgs, "; ")
}

// ToMap converts ValidationErrors to a map for API responses.
func (ve ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string, len(ve))
	for _, e := range ve {
		result[e.Field] = e.Message
	}

	return result
}

// New creates a new Validator instance with custom configurations.
func New() *Validator {
	v := validator.New()

	// Use JSON tag names for field names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", jsonTagParts)[0]
		if name == "-" {
			return fld.Name
		}

		return name
	})

	return &Validator{validate: v}
}

// Validate validates a struct and returns formatted validation errors.
func (v *Validator) Validate(s any) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors

	ok := errors.As(err, &validationErrs)
	if !ok {
		return fmt.Errorf("%w: %w", ErrValidation, err)
	}

	errs := make(ValidationErrors, 0, len(validationErrs))
	for _, e := range validationErrs {
		errs = append(errs, ValidationError{
			Field:   e.Field(),
			Message: formatMessage(e),
		})
	}

	return errs
}

// formatMessage creates a human-readable error message for a validation error.
func formatMessage(e validator.FieldError) string {
	tag := e.Tag()

	// Check for static messages first
	if msg, ok := validationMessages[tag]; ok {
		return msg
	}

	// Check for parameterized messages
	if format, ok := parameterizedMessages[tag]; ok {
		return fmt.Sprintf(format, e.Param())
	}

	// Default fallback
	return fmt.Sprintf("failed %s validation", tag)
}
