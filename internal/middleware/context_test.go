package middleware_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/middleware"
)

func TestGetAuthenticatedUser(t *testing.T) {
	t.Parallel()

	t.Run("returns user when present in context", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New()
		expectedUser := &middleware.AuthenticatedUser{
			UserID:    userID,
			ClientID:  "test-client",
			Scopes:    []string{"read", "write"},
			IsService: false,
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), expectedUser)

		user, ok := middleware.GetAuthenticatedUser(ctx)
		require.True(t, ok)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser.UserID, user.UserID)
		assert.Equal(t, expectedUser.ClientID, user.ClientID)
		assert.Equal(t, expectedUser.Scopes, user.Scopes)
		assert.Equal(t, expectedUser.IsService, user.IsService)
	})

	t.Run("returns false when no user in context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		user, ok := middleware.GetAuthenticatedUser(ctx)
		assert.False(t, ok)
		assert.Nil(t, user)
	})

	t.Run("returns false when context has wrong type", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), middleware.AuthUserKey, "not a user")

		user, ok := middleware.GetAuthenticatedUser(ctx)
		assert.False(t, ok)
		assert.Nil(t, user)
	})

	t.Run("returns false when context has nil user", func(t *testing.T) {
		t.Parallel()

		ctx := middleware.SetAuthenticatedUser(context.Background(), nil)

		user, ok := middleware.GetAuthenticatedUser(ctx)
		assert.False(t, ok)
		assert.Nil(t, user)
	})
}

func TestSetAuthenticatedUser(t *testing.T) {
	t.Parallel()

	t.Run("stores user in context", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New()
		user := &middleware.AuthenticatedUser{
			UserID:   userID,
			ClientID: "my-client",
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		retrieved, ok := middleware.GetAuthenticatedUser(ctx)
		require.True(t, ok)
		assert.Equal(t, user, retrieved)
	})

	t.Run("overwrites existing user", func(t *testing.T) {
		t.Parallel()

		user1 := &middleware.AuthenticatedUser{UserID: uuid.New(), ClientID: "client1"}
		user2 := &middleware.AuthenticatedUser{UserID: uuid.New(), ClientID: "client2"}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user1)
		ctx = middleware.SetAuthenticatedUser(ctx, user2)

		retrieved, ok := middleware.GetAuthenticatedUser(ctx)
		require.True(t, ok)
		assert.Equal(t, user2.ClientID, retrieved.ClientID)
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Parallel()

	t.Run("returns user ID for regular user", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New()
		user := &middleware.AuthenticatedUser{
			UserID:    userID,
			ClientID:  "test-client",
			IsService: false,
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		id, ok := middleware.GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, userID, id)
	})

	t.Run("returns false for service account", func(t *testing.T) {
		t.Parallel()

		user := &middleware.AuthenticatedUser{
			UserID:    uuid.Nil,
			ClientID:  "service-client",
			IsService: true,
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		id, ok := middleware.GetUserIDFromContext(ctx)
		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("returns false when no user in context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		id, ok := middleware.GetUserIDFromContext(ctx)
		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("returns false for nil user ID", func(t *testing.T) {
		t.Parallel()

		user := &middleware.AuthenticatedUser{
			UserID:    uuid.Nil,
			ClientID:  "test-client",
			IsService: false,
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		id, ok := middleware.GetUserIDFromContext(ctx)
		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, id)
	})
}

//nolint:funlen // table-driven test
func TestHasScope(t *testing.T) {
	t.Parallel()

	t.Run("returns true when scope is present", func(t *testing.T) {
		t.Parallel()

		user := &middleware.AuthenticatedUser{
			UserID: uuid.New(),
			Scopes: []string{"read", "write", "admin"},
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		assert.True(t, middleware.HasScope(ctx, "read"))
		assert.True(t, middleware.HasScope(ctx, "write"))
		assert.True(t, middleware.HasScope(ctx, "admin"))
	})

	t.Run("returns false when scope is not present", func(t *testing.T) {
		t.Parallel()

		user := &middleware.AuthenticatedUser{
			UserID: uuid.New(),
			Scopes: []string{"read"},
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		assert.False(t, middleware.HasScope(ctx, "write"))
		assert.False(t, middleware.HasScope(ctx, "admin"))
	})

	t.Run("returns false when no user in context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		assert.False(t, middleware.HasScope(ctx, "read"))
	})

	t.Run("returns false when scopes is empty", func(t *testing.T) {
		t.Parallel()

		user := &middleware.AuthenticatedUser{
			UserID: uuid.New(),
			Scopes: []string{},
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		assert.False(t, middleware.HasScope(ctx, "read"))
	})

	t.Run("returns false when scopes is nil", func(t *testing.T) {
		t.Parallel()

		user := &middleware.AuthenticatedUser{
			UserID: uuid.New(),
			Scopes: nil,
		}

		ctx := middleware.SetAuthenticatedUser(context.Background(), user)

		assert.False(t, middleware.HasScope(ctx, "read"))
	})
}
