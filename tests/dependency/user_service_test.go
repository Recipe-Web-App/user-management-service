package dependency_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	headerUserID = "X-User-Id"
	baseURL      = "/api/v1/user-management/users"
	reqPathFmt   = "%s/%s/profile"
)

var (
	errUnexpectedUserType    = errors.New("unexpected type for User")
	errUnexpectedPrefsType   = errors.New("unexpected type for PrivacyPreferences")
	errMockReturnedNilResult = errors.New("mock returned nil result without error")
)

// MockUserRepository is a mock implementation of repository.UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*dto.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("find user: %w", err)
		}

		return nil, errMockReturnedNilResult
	}

	user, ok := args.Get(0).(*dto.User)
	if !ok {
		return nil, errUnexpectedUserType
	}

	err := args.Error(1)
	if err != nil {
		return user, fmt.Errorf("find user: %w", err)
	}

	return user, nil
}

func (m *MockUserRepository) FindPrivacyPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.PrivacyPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("find privacy preferences: %w", err)
		}

		return nil, errMockReturnedNilResult
	}

	prefs, ok := args.Get(0).(*dto.PrivacyPreferences)
	if !ok {
		return nil, errUnexpectedPrefsType
	}

	err := args.Error(1)
	if err != nil {
		return prefs, fmt.Errorf("find privacy preferences: %w", err)
	}

	return prefs, nil
}

func (m *MockUserRepository) IsFollowing(ctx context.Context, followerID, followedID uuid.UUID) (bool, error) {
	args := m.Called(ctx, followerID, followedID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.UserProfileUpdateRequest,
) (*dto.User, error) {
	args := m.Called(ctx, userID, update)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}

		return nil, errMockReturnedNilResult
	}

	user, ok := args.Get(0).(*dto.User)
	if !ok {
		return nil, errUnexpectedUserType
	}

	err := args.Error(1)
	if err != nil {
		return user, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

type testFixture struct {
	handler     http.Handler
	mockRepo    *MockUserRepository
	requesterID uuid.UUID
}

func setupTest(t *testing.T) *testFixture {
	t.Helper()

	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}

	container, err := app.NewContainer(app.ContainerConfig{Config: cfg, UserRepo: mockRepo})
	require.NoError(t, err)

	srv := server.NewServerWithContainer(container)

	return &testFixture{handler: srv.Handler, mockRepo: mockRepo, requesterID: uuid.New()}
}

func newProfileRequest(t *testing.T, userID, requesterID uuid.UUID) *http.Request {
	t.Helper()

	reqPath := fmt.Sprintf(reqPathFmt, baseURL, userID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set(headerUserID, requesterID.String())

	return req
}

func createTestUser(userID uuid.UUID) *dto.User {
	return &dto.User{
		UserID:    userID.String(),
		Username:  "targetuser",
		FullName:  func() *string { s := "Target User"; return &s }(),
		Email:     func() *string { s := "email@example.com"; return &s }(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestGetUserProfile(t *testing.T) {
	t.Parallel()

	t.Run("Success_PublicProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUser(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public", ShowFullName: true, ShowEmail: true}

		fix.mockRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newProfileRequest(t, targetUserID, fix.requesterID))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                    `json:"success"`
			Data    dto.UserProfileResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, targetUser.Username, resp.Data.Username)
	})

	t.Run("NotFound_UserDoesNotExist", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		nonExistentID := uuid.New()

		fix.mockRepo.On("FindUserByID", mock.Anything, nonExistentID).Return(nil, repository.ErrUserNotFound).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newProfileRequest(t, nonExistentID, fix.requesterID))

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Forbidden_PrivateProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		privateTargetID := uuid.New()
		privateUser := &dto.User{UserID: privateTargetID.String(), Username: "privateuser"}
		privatePrivacy := &dto.PrivacyPreferences{ProfileVisibility: "private"}

		fix.mockRepo.On("FindUserByID", mock.Anything, privateTargetID).Return(privateUser, nil).Once()
		fix.mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, privateTargetID).Return(privatePrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newProfileRequest(t, privateTargetID, fix.requesterID))

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

func newUpdateProfileRequest(t *testing.T, userID uuid.UUID, body string) *http.Request {
	t.Helper()

	reqPath := baseURL + "/profile"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqPath, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set(headerUserID, userID.String())
	req.Header.Set("Content-Type", "application/json")

	return req
}

func TestUpdateUserProfile(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	t.Run("Success_UpdateUsername", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		userID := fix.requesterID
		now := time.Now()

		existingUser := &dto.User{
			UserID:    userID.String(),
			Username:  "oldusername",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		updatedUser := &dto.User{
			UserID:    userID.String(),
			Username:  "newusername",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, userID).Return(existingUser, nil).Once()
		fix.mockRepo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(updatedUser, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newUpdateProfileRequest(t, userID, `{"username": "newusername"}`))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                    `json:"success"`
			Data    dto.UserProfileResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, "newusername", resp.Data.Username)
	})

	t.Run("NotFound_UserDoesNotExist", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		userID := uuid.New()

		fix.mockRepo.On("FindUserByID", mock.Anything, userID).Return(nil, repository.ErrUserNotFound).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newUpdateProfileRequest(t, userID, `{"username": "newusername"}`))

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Conflict_DuplicateUsername", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		userID := fix.requesterID
		now := time.Now()

		existingUser := &dto.User{
			UserID:    userID.String(),
			Username:  "oldusername",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, userID).Return(existingUser, nil).Once()
		fix.mockRepo.On("UpdateUser", mock.Anything, userID, mock.Anything).
			Return(nil, repository.ErrDuplicateUsername).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newUpdateProfileRequest(t, userID, `{"username": "existinguser"}`))

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		reqBody := strings.NewReader(`{"username": "newusername"}`)
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPut,
			baseURL+"/profile",
			reqBody,
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
