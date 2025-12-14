// Package handler provides HTTP request handlers for the API.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/validation"
)

// Sentinel errors for request binding.
var (
	ErrEmptyBody            = errors.New("request body is empty")
	ErrUnsupportedMediaType = errors.New("unsupported content type")
	ErrInvalidJSON          = errors.New("invalid JSON")
	ErrInvalidFieldType     = errors.New("invalid field type")
	ErrValidationFailed     = errors.New("validation failed")
)

// RequestBinder handles binding and validating HTTP request bodies.
type RequestBinder struct {
	validator *validation.Validator
}

// NewRequestBinder creates a new RequestBinder with a validator.
func NewRequestBinder() *RequestBinder {
	return &RequestBinder{
		validator: validation.New(),
	}
}

// BindJSON reads JSON from the request body and unmarshals it into the target.
func (b *RequestBinder) BindJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return ErrEmptyBody
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") && contentType != "" {
		return fmt.Errorf("%w: %s", ErrUnsupportedMediaType, contentType)
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(target)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyBody
		}

		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return fmt.Errorf("%w at position %d", ErrInvalidJSON, syntaxErr.Offset)
		}

		var unmarshalErr *json.UnmarshalTypeError
		if errors.As(err, &unmarshalErr) {
			return fmt.Errorf("%w for field %s: expected %s",
				ErrInvalidFieldType, unmarshalErr.Field, unmarshalErr.Type.String())
		}

		if strings.HasPrefix(err.Error(), "json: unknown field") {
			return fmt.Errorf("%w: %w", ErrInvalidJSON, err)
		}

		return fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	}

	return nil
}

// Validate validates a struct using the validator.
func (b *RequestBinder) Validate(target any) error {
	err := b.validator.Validate(target)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}

	return nil
}

// BindAndValidate combines binding and validation in a single call.
func (b *RequestBinder) BindAndValidate(r *http.Request, target any) error {
	err := b.BindJSON(r, target)
	if err != nil {
		return err
	}

	return b.Validate(target)
}
