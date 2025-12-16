package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/validation"
)

// JSONResponse writes a JSON response with the given status code.
func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// SuccessResponse writes a successful JSON response.
func SuccessResponse(w http.ResponseWriter, status int, data any) {
	JSONResponse(w, status, dto.Response{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse writes an error JSON response.
func ErrorResponse(w http.ResponseWriter, status int, code, message string) {
	JSONResponse(w, status, dto.Response{
		Success: false,
		Error: &dto.Error{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationErrorResponse writes a validation error response.
func ValidationErrorResponse(w http.ResponseWriter, err error) {
	response := dto.Response{
		Success: false,
		Error: &dto.Error{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
		},
	}

	var validationErrs validation.ValidationErrors
	if errors.As(err, &validationErrs) {
		response.Error.Details = validationErrs.ToMap()
	}

	JSONResponse(w, http.StatusBadRequest, response)
}

// NotFoundResponse writes a 404 not found response.
func NotFoundResponse(w http.ResponseWriter, resource string) {
	ErrorResponse(w, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

// InternalErrorResponse writes a 500 internal server error response.
func InternalErrorResponse(w http.ResponseWriter) {
	ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
}

// UnauthorizedResponse writes a 401 unauthorized response.
func UnauthorizedResponse(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}

	ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// ForbiddenResponse writes a 403 forbidden response.
func ForbiddenResponse(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}

	ErrorResponse(w, http.StatusForbidden, "FORBIDDEN", message)
}

// ConflictResponse writes a 409 conflict response.
func ConflictResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusConflict, "CONFLICT", message)
}

// ServiceUnavailableResponse writes a 503 service unavailable response.
func ServiceUnavailableResponse(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Service temporarily unavailable"
	}

	ErrorResponse(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}
