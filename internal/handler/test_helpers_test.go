package handler_test

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/middleware"
)

// setAuthenticatedUser adds an authenticated user to the request context.
// This helper replaces the X-User-Id header approach for tests.
func setAuthenticatedUser(req *http.Request, userID uuid.UUID) *http.Request {
	user := &middleware.AuthenticatedUser{
		UserID:    userID,
		ClientID:  "test-client",
		IsService: false,
	}
	ctx := middleware.SetAuthenticatedUser(req.Context(), user)

	return req.WithContext(ctx)
}

// setAuthenticatedUserFromString adds an authenticated user to the request context
// after parsing the user ID string. If parsing fails, the context is not modified.
func setAuthenticatedUserFromString(req *http.Request, userIDStr string) *http.Request {
	if userIDStr == "" {
		return req
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Return unchanged request - this simulates invalid UUID scenario
		// The handler will detect the missing context and return 401
		return req
	}

	return setAuthenticatedUser(req, userID)
}
