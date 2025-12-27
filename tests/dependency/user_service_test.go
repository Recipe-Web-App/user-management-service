package dependency_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/redis"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	headerUserID = "X-User-Id"
	baseURL      = "/api/v1/user-management/users"
	reqPathFmt   = "%s/%s/profile"
	mockErrorFmt = "mock error: %w"
)

var (
	errUnexpectedUserType         = errors.New("unexpected type for User")
	errUnexpectedPrefsType        = errors.New("unexpected type for PrivacyPreferences")
	errMockReturnedNilResult      = errors.New("mock returned nil result")
	errMockUnexpectedNotifPrefs   = errors.New("unexpected type for NotificationPreferences")
	errMockUnexpectedDisplayPrefs = errors.New("unexpected type for DisplayPreferences")
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

func (m *MockUserRepository) FindNotificationPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.NotificationPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("find notification preferences: %w", err)
		}

		return nil, errMockReturnedNilResult
	}

	prefs, ok := args.Get(0).(*dto.NotificationPreferences)
	if !ok {
		return nil, errMockUnexpectedNotifPrefs
	}

	err := args.Error(1)
	if err != nil {
		return prefs, fmt.Errorf(mockErrorFmt, err)
	}

	return prefs, nil
}

func (m *MockUserRepository) FindDisplayPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.DisplayPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("find display preferences: %w", err)
		}

		return nil, errMockReturnedNilResult
	}

	prefs, ok := args.Get(0).(*dto.DisplayPreferences)
	if !ok {
		return nil, errMockUnexpectedDisplayPrefs
	}

	err := args.Error(1)
	if err != nil {
		return prefs, fmt.Errorf(mockErrorFmt, err)
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

func (m *MockUserRepository) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]dto.UserSearchResult, int, error) {
	args := m.Called(ctx, query, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf("search users: %w", err)
	}

	results, _ := args.Get(0).([]dto.UserSearchResult)

	return results, args.Int(1), nil
}

func (m *MockUserRepository) UpdateNotificationPreferences(
	ctx context.Context,
	userID uuid.UUID,
	prefs *dto.NotificationPreferences,
) error {
	args := m.Called(ctx, userID, prefs)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf(mockErrorFmt, err)
	}

	return nil
}

func (m *MockUserRepository) UpdatePrivacyPreferences(
	ctx context.Context,
	userID uuid.UUID,
	prefs *dto.PrivacyPreferences,
) error {
	args := m.Called(ctx, userID, prefs)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf(mockErrorFmt, err)
	}

	return nil
}

func (m *MockUserRepository) UpdateDisplayPreferences(
	ctx context.Context,
	userID uuid.UUID,
	prefs *dto.DisplayPreferences,
) error {
	args := m.Called(ctx, userID, prefs)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf(mockErrorFmt, err)
	}

	return nil
}

func (m *MockUserRepository) GetUserStats(ctx context.Context) (*dto.UserStatsResponse, error) {
	return &dto.UserStatsResponse{}, nil
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

		var resp dto.UserProfileResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, targetUser.Username, resp.Username)
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

		var resp dto.UserProfileResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "newusername", resp.Username)
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

// testFixtureWithRedis includes a real miniredis instance for token storage tests.
type testFixtureWithRedis struct {
	handler     http.Handler
	mockRepo    *MockUserRepository
	redisServer *miniredis.Miniredis
	requesterID uuid.UUID
}

func setupTestWithRedis(t *testing.T) *testFixtureWithRedis {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	port, _ := strconv.Atoi(mr.Port())

	redisConfig := &config.RedisConfig{
		Host:     mr.Host(),
		Port:     port,
		Database: 0,
		Password: "",
	}

	redisService, err := redis.New(redisConfig)
	require.NoError(t, err)

	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}

	container, err := app.NewContainer(app.ContainerConfig{
		Config:     cfg,
		UserRepo:   mockRepo,
		TokenStore: redisService,
	})
	require.NoError(t, err)

	srv := server.NewServerWithContainer(container)

	return &testFixtureWithRedis{
		handler:     srv.Handler,
		mockRepo:    mockRepo,
		redisServer: mr,
		requesterID: uuid.New(),
	}
}

func newDeleteRequestRequest(t *testing.T, userID uuid.UUID) *http.Request {
	t.Helper()

	reqPath := baseURL + "/account/delete-request"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set(headerUserID, userID.String())

	return req
}

func TestRequestAccountDeletion(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	t.Run("Success_WithRealRedis", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		userID := fix.requesterID
		now := time.Now()

		user := &dto.User{
			UserID:    userID.String(),
			Username:  "testuser",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteRequestRequest(t, userID))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.UserAccountDeleteRequestResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, userID.String(), resp.UserID)
		assert.NotEmpty(t, resp.ConfirmationToken)
		assert.False(t, resp.ExpiresAt.IsZero())

		// Verify token is stored in Redis
		storedToken, err := fix.redisServer.Get("delete-request:" + userID.String())
		require.NoError(t, err)
		assert.Equal(t, resp.ConfirmationToken, storedToken)

		// Verify TTL is set (approximately 24 hours)
		ttl := fix.redisServer.TTL("delete-request:" + userID.String())
		assert.Positive(t, ttl)
		assert.LessOrEqual(t, ttl, service.DeleteTokenTTL)
	})

	t.Run("TokenReplacement_NewRequestReplacesOldToken", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		userID := fix.requesterID
		now := time.Now()

		user := &dto.User{
			UserID:    userID.String(),
			Username:  "testuser",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)

		// First request
		rr1 := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr1, newDeleteRequestRequest(t, userID))
		require.Equal(t, http.StatusOK, rr1.Code)

		var resp1 dto.UserAccountDeleteRequestResponse
		require.NoError(t, json.Unmarshal(rr1.Body.Bytes(), &resp1))
		firstToken := resp1.ConfirmationToken

		// Second request
		rr2 := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr2, newDeleteRequestRequest(t, userID))
		require.Equal(t, http.StatusOK, rr2.Code)

		var resp2 dto.UserAccountDeleteRequestResponse
		require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &resp2))
		secondToken := resp2.ConfirmationToken

		// Verify tokens are different
		assert.NotEqual(t, firstToken, secondToken)

		// Verify only the second token is stored
		storedToken, err := fix.redisServer.Get("delete-request:" + userID.String())
		require.NoError(t, err)
		assert.Equal(t, secondToken, storedToken)
	})

	t.Run("NotFound_UserDoesNotExist", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		nonExistentID := uuid.New()

		fix.mockRepo.On("FindUserByID", mock.Anything, nonExistentID).Return(nil, repository.ErrUserNotFound).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteRequestRequest(t, nonExistentID))

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			baseURL+"/account/delete-request",
			nil,
		)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func newConfirmDeletionRequest(t *testing.T, userID uuid.UUID, body string) *http.Request {
	t.Helper()

	reqPath := baseURL + "/account"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, reqPath, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set(headerUserID, userID.String())
	req.Header.Set("Content-Type", "application/json")

	return req
}

func TestConfirmAccountDeletion(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	t.Run("Success_WithRealRedis", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		userID := fix.requesterID
		now := time.Now()

		deactivatedUser := &dto.User{
			UserID:    userID.String(),
			Username:  "testuser",
			IsActive:  false,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Pre-store a token in Redis
		token := uuid.New().String()
		err := fix.redisServer.Set("delete-request:"+userID.String(), token)
		require.NoError(t, err)

		fix.mockRepo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(deactivatedUser, nil).Once()

		reqBody := fmt.Sprintf(`{"confirmationToken": "%s"}`, token)
		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newConfirmDeletionRequest(t, userID, reqBody))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.UserConfirmAccountDeleteResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, userID.String(), resp.UserID)
		assert.False(t, resp.DeactivatedAt.IsZero())

		// Verify token is deleted from Redis
		exists := fix.redisServer.Exists("delete-request:" + userID.String())
		assert.False(t, exists)
	})

	t.Run("InvalidToken_WrongToken", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		userID := fix.requesterID

		// Pre-store a token in Redis
		correctToken := uuid.New().String()
		err := fix.redisServer.Set("delete-request:"+userID.String(), correctToken)
		require.NoError(t, err)

		// Send wrong token
		reqBody := `{"confirmationToken": "wrong-token"}`
		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newConfirmDeletionRequest(t, userID, reqBody))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "INVALID_TOKEN")
	})

	t.Run("InvalidToken_TokenExpired", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		userID := fix.requesterID

		// No token stored in Redis (simulating expired/not requested)

		reqBody := `{"confirmationToken": "some-token"}`
		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newConfirmDeletionRequest(t, userID, reqBody))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "INVALID_TOKEN")
	})

	t.Run("FullFlow_RequestThenConfirm", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		userID := fix.requesterID
		now := time.Now()

		user := &dto.User{
			UserID:    userID.String(),
			Username:  "testuser",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		deactivatedUser := &dto.User{
			UserID:    userID.String(),
			Username:  "testuser",
			IsActive:  false,
			CreatedAt: now,
			UpdatedAt: now,
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil).Once()
		fix.mockRepo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(deactivatedUser, nil).Once()

		// Step 1: Request deletion
		rr1 := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr1, newDeleteRequestRequest(t, userID))
		require.Equal(t, http.StatusOK, rr1.Code)

		var requestResp dto.UserAccountDeleteRequestResponse
		require.NoError(t, json.Unmarshal(rr1.Body.Bytes(), &requestResp))
		token := requestResp.ConfirmationToken

		// Step 2: Confirm deletion with the received token
		reqBody := fmt.Sprintf(`{"confirmationToken": "%s"}`, token)
		rr2 := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr2, newConfirmDeletionRequest(t, userID, reqBody))

		require.Equal(t, http.StatusOK, rr2.Code)

		var confirmResp dto.UserConfirmAccountDeleteResponse
		require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &confirmResp))
		assert.Equal(t, userID.String(), confirmResp.UserID)
		assert.False(t, confirmResp.DeactivatedAt.IsZero())

		// Verify token is deleted
		exists := fix.redisServer.Exists("delete-request:" + userID.String())
		assert.False(t, exists)
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupTestWithRedis(t)
		defer fix.redisServer.Close()

		reqBody := strings.NewReader(`{"confirmationToken": "some-token"}`)
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodDelete,
			baseURL+"/account",
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

func newSearchRequest(t *testing.T, requesterID uuid.UUID, queryParams string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		baseURL+"/search"+queryParams,
		nil,
	)
	require.NoError(t, err)
	req.Header.Set(headerUserID, requesterID.String())

	return req
}

//nolint:funlen // Table-driven test with multiple well-organized subtests
func TestSearchUsers(t *testing.T) {
	t.Parallel()

	t.Run("Success_ReturnsSearchResults", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		now := time.Now()
		fullName := "Test User"

		searchResults := []dto.UserSearchResult{
			{
				UserID:    uuid.New().String(),
				Username:  "testuser1",
				FullName:  &fullName,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		fix.mockRepo.On("SearchUsers", mock.Anything, "test", 20, 0).Return(searchResults, 1, nil)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, "?query=test"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.UserSearchResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, 1, resp.TotalCount)
		assert.Len(t, resp.Results, 1)
		assert.Equal(t, "testuser1", resp.Results[0].Username)
	})

	t.Run("Success_CountOnly", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		fix.mockRepo.On("SearchUsers", mock.Anything, "test", 20, 0).Return([]dto.UserSearchResult{}, 10, nil)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, "?query=test&countOnly=true"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.UserSearchResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, 10, resp.TotalCount)
		assert.Empty(t, resp.Results)
	})

	t.Run("Success_WithPagination", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		fix.mockRepo.On("SearchUsers", mock.Anything, "user", 10, 5).Return([]dto.UserSearchResult{}, 25, nil)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, "?query=user&limit=10&offset=5"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.UserSearchResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, 25, resp.TotalCount)
		assert.Equal(t, 10, resp.Limit)
		assert.Equal(t, 5, resp.Offset)
	})

	t.Run("Success_EmptyQuery", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		fix.mockRepo.On("SearchUsers", mock.Anything, "", 20, 0).Return([]dto.UserSearchResult{}, 0, nil)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, ""))

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("BadRequest_InvalidLimit", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, "?query=test&limit=0"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("BadRequest_LimitOverMax", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, "?query=test&limit=101"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("BadRequest_NegativeOffset", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, "?query=test&offset=-1"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			baseURL+"/search?query=test",
			nil,
		)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InternalError_DatabaseFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)

		fix.mockRepo.On("SearchUsers", mock.Anything, "test", 20, 0).Return(nil, 0, repository.ErrUserNotFound)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newSearchRequest(t, fix.requesterID, "?query=test"))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func newGetUserByIDRequest(t *testing.T, userID uuid.UUID) *http.Request {
	t.Helper()

	reqPath := fmt.Sprintf("%s/%s", baseURL, userID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	require.NoError(t, err)
	// No X-User-Id header needed - anonymous access

	return req
}

func TestGetUserByID(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	t.Run("Success_PublicProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		targetUserID := uuid.New()
		now := time.Now()

		user := &dto.User{
			UserID:    targetUserID.String(),
			Username:  "publicuser",
			FullName:  func() *string { s := "Public User"; return &s }(),
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		publicPrivacy := &dto.PrivacyPreferences{
			ProfileVisibility: "public",
			ShowFullName:      true,
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, targetUserID).Return(user, nil).Once()
		fix.mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetUserByIDRequest(t, targetUserID))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.UserSearchResult
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, user.Username, resp.Username)
	})

	t.Run("NotFound_UserDoesNotExist", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		nonExistentID := uuid.New()

		fix.mockRepo.On("FindUserByID", mock.Anything, nonExistentID).Return(nil, repository.ErrUserNotFound).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetUserByIDRequest(t, nonExistentID))

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("NotFound_PrivateProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		privateUserID := uuid.New()
		now := time.Now()

		user := &dto.User{
			UserID:    privateUserID.String(),
			Username:  "privateuser",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		privatePrivacy := &dto.PrivacyPreferences{
			ProfileVisibility: "private",
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, privateUserID).Return(user, nil).Once()
		fix.mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, privateUserID).Return(privatePrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetUserByIDRequest(t, privateUserID))

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("NotFound_FollowersOnlyProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupTest(t)
		followersOnlyUserID := uuid.New()
		now := time.Now()

		user := &dto.User{
			UserID:    followersOnlyUserID.String(),
			Username:  "followersuser",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		followersOnlyPrivacy := &dto.PrivacyPreferences{
			ProfileVisibility: "followers_only",
		}

		fix.mockRepo.On("FindUserByID", mock.Anything, followersOnlyUserID).Return(user, nil).Once()
		fix.mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, followersOnlyUserID).
			Return(followersOnlyPrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetUserByIDRequest(t, followersOnlyUserID))

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
