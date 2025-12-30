package component_test

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
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const mockErrorFmt = "mock error: %w"

var (
	errMockArgs           = errors.New("mock: missing args")
	errMockInvalidUser    = errors.New("invalid type assertion for User")
	errMockInvalidPrivacy = errors.New("invalid type assertion for PrivacyPreferences")
)

// MockUserRepo for component tests.
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) FindUserByID(ctx context.Context, userID uuid.UUID) (*dto.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockErrorFmt, err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.User); ok {
		return val, nil
	}

	return nil, errMockInvalidUser
}

func (m *MockUserRepo) FindPrivacyPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.PrivacyPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockErrorFmt, err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.PrivacyPreferences); ok {
		return val, nil
	}

	return nil, errMockInvalidPrivacy
}

func (m *MockUserRepo) IsFollowing(ctx context.Context, followerID, followedID uuid.UUID) (bool, error) {
	args := m.Called(ctx, followerID, followedID)

	err := args.Error(1)
	if err != nil {
		return args.Bool(0), fmt.Errorf(mockErrorFmt, err)
	}

	return args.Bool(0), nil
}

func (m *MockUserRepo) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.UserProfileUpdateRequest,
) (*dto.User, error) {
	args := m.Called(ctx, userID, update)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockErrorFmt, err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.User); ok {
		return val, nil
	}

	return nil, errMockInvalidUser
}

func (m *MockUserRepo) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]dto.UserSearchResult, int, error) {
	args := m.Called(ctx, query, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf(mockErrorFmt, err)
	}

	results, _ := args.Get(0).([]dto.UserSearchResult)

	return results, args.Int(1), nil
}

func (m *MockUserRepo) GetUserStats(ctx context.Context) (*dto.UserStatsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, fmt.Errorf("mock error: %w", args.Error(1))
	}

	if val, ok := args.Get(0).(*dto.UserStatsResponse); ok {
		return val, nil
	}

	return nil, fmt.Errorf("mock error: %w", args.Error(1))
}

// MockTokenStore for component tests.
type MockTokenStore struct {
	mock.Mock
}

func (m *MockTokenStore) StoreDeleteToken(
	ctx context.Context,
	userID uuid.UUID,
	token string,
	ttl time.Duration,
) error {
	args := m.Called(ctx, userID, token, ttl)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf(mockErrorFmt, err)
	}

	return nil
}

func (m *MockTokenStore) GetDeleteToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)

	err := args.Error(1)
	if err != nil {
		return args.String(0), fmt.Errorf(mockErrorFmt, err)
	}

	return args.String(0), nil
}

func (m *MockTokenStore) DeleteDeleteToken(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf(mockErrorFmt, err)
	}

	return nil
}

func TestUserProfileComponent(t *testing.T) {
	t.Parallel()

	// Create Mock Repo
	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)

	// Create Service using Mock Repo
	svc := service.NewUserService(mockRepo, mockTokenStore)

	// Create Container
	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	// Setup Health so router doesn't panic
	c.HealthService = service.NewHealthService(nil, nil)

	// Create Server
	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	// Setup Data
	user := &dto.User{
		UserID:    userID.String(),
		Username:  "componentuser",
		FullName:  func() *string { s := "Component User"; return &s }(),
		Email:     func() *string { s := "test@example.com"; return &s }(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	privacy := &dto.PrivacyPreferences{
		ProfileVisibility: "public",
		ShowEmail:         true,
		ShowFullName:      true,
	}

	// Mock Expectations
	mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, userID).Return(privacy, nil)

	// Perform Request
	requesterID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String()+"/profile", nil)
	req.Header.Set("X-User-Id", requesterID.String())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Assert Response
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp dto.UserProfileResponse

	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), resp.UserID)
	assert.NotNil(t, resp.FullName)
	assert.Equal(t, "Component User", *resp.FullName)
	assert.NotNil(t, resp.Email)
	assert.Equal(t, "test@example.com", *resp.Email)
}

func TestUpdateUserProfileComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	now := time.Now()

	existingUser := &dto.User{
		UserID:    userID.String(),
		Username:  "oldusername",
		Email:     func() *string { s := "old@example.com"; return &s }(),
		FullName:  func() *string { s := "Old Name"; return &s }(),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	updatedUser := &dto.User{
		UserID:    userID.String(),
		Username:  "newusername",
		Email:     func() *string { s := "old@example.com"; return &s }(),
		FullName:  func() *string { s := "Old Name"; return &s }(),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(existingUser, nil)
	mockRepo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(updatedUser, nil)

	reqBody := `{"username": "newusername"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user-management/users/profile", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp dto.UserProfileResponse

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, "newusername", apiResp.Username)
}

func TestUpdateUserProfileComponent_NotFound(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(nil, repository.ErrUserNotFound)

	reqBody := `{"username": "newusername"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user-management/users/profile", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateUserProfileComponent_DuplicateUsername(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	now := time.Now()

	existingUser := &dto.User{
		UserID:    userID.String(),
		Username:  "oldusername",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(existingUser, nil)
	mockRepo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(nil, repository.ErrDuplicateUsername)

	reqBody := `{"username": "existinguser"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user-management/users/profile", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestUpdateUserProfileComponent_ValidationError(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	// Invalid username (too short)
	reqBody := `{"username": "ab"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user-management/users/profile", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

func TestRequestAccountDeletionComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	now := time.Now()

	user := &dto.User{
		UserID:    userID.String(),
		Username:  "testuser",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)
	mockTokenStore.On("StoreDeleteToken", mock.Anything, userID, mock.Anything, service.DeleteTokenTTL).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-management/users/account/delete-request", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp dto.UserAccountDeleteRequestResponse

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), apiResp.UserID)
	assert.NotEmpty(t, apiResp.ConfirmationToken)
	assert.False(t, apiResp.ExpiresAt.IsZero())
}

func TestRequestAccountDeletionComponent_Unauthorized(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	// Missing X-User-Id header
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-management/users/account/delete-request", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequestAccountDeletionComponent_NotFound(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(nil, repository.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-management/users/account/delete-request", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
}

func TestRequestAccountDeletionComponent_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	// Use nil token store to simulate cache unavailable
	svc := service.NewUserService(mockRepo, nil)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-management/users/account/delete-request", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Contains(t, rr.Body.String(), "SERVICE_UNAVAILABLE")
}

func TestConfirmAccountDeletionComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	token := uuid.New().String()
	now := time.Now()

	deactivatedUser := &dto.User{
		UserID:    userID.String(),
		Username:  "testuser",
		IsActive:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockTokenStore.On("GetDeleteToken", mock.Anything, userID).Return(token, nil)
	mockRepo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(deactivatedUser, nil)
	mockTokenStore.On("DeleteDeleteToken", mock.Anything, userID).Return(nil)

	reqBody := fmt.Sprintf(`{"confirmationToken": "%s"}`, token)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/users/account", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp dto.UserConfirmAccountDeleteResponse

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), apiResp.UserID)
	assert.False(t, apiResp.DeactivatedAt.IsZero())
}

func TestConfirmAccountDeletionComponent_InvalidToken(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	storedToken := uuid.New().String()

	mockTokenStore.On("GetDeleteToken", mock.Anything, userID).Return(storedToken, nil)

	reqBody := `{"confirmationToken": "wrong-token"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/users/account", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_TOKEN")
}

func TestConfirmAccountDeletionComponent_Unauthorized(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	reqBody := `{"confirmationToken": "some-token"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/users/account", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	// Missing X-User-Id header

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfirmAccountDeletionComponent_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	// Use nil token store to simulate cache unavailable
	svc := service.NewUserService(mockRepo, nil)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	reqBody := `{"confirmationToken": "some-token"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/users/account", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Contains(t, rr.Body.String(), "SERVICE_UNAVAILABLE")
}

// ============================================================================
// SearchUsers Component Tests
// ============================================================================

func TestSearchUsersComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
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

	mockRepo.On("SearchUsers", mock.Anything, "test", 20, 0).Return(searchResults, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/search?query=test", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify response structure
	assert.Contains(t, rr.Body.String(), "testuser1")
	assert.Contains(t, rr.Body.String(), `"totalCount":1`)
}

func TestSearchUsersComponent_CountOnly(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	// When countOnly is true, service still calls repo but returns empty results
	mockRepo.On("SearchUsers", mock.Anything, "test", 20, 0).Return([]dto.UserSearchResult{}, 5, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/search?query=test&countOnly=true", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":5`)
	assert.Contains(t, rr.Body.String(), `"results":[]`)
}

func TestSearchUsersComponent_WithPagination(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	mockRepo.On("SearchUsers", mock.Anything, "test", 10, 5).Return([]dto.UserSearchResult{}, 15, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/search?query=test&limit=10&offset=5", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":15`)
	assert.Contains(t, rr.Body.String(), `"limit":10`)
	assert.Contains(t, rr.Body.String(), `"offset":5`)
}

func TestSearchUsersComponent_Unauthorized(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	// Missing X-User-Id header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/search?query=test", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestSearchUsersComponent_InvalidLimit(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/search?query=test&limit=0", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

// ============================================================================
// GetUserByID Component Tests
// ============================================================================

func TestGetUserByIDComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	now := time.Now()

	user := &dto.User{
		UserID:    userID.String(),
		Username:  "publicuser",
		FullName:  func() *string { s := "Public User"; return &s }(),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	privacy := &dto.PrivacyPreferences{
		ProfileVisibility: "public",
		ShowFullName:      true,
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, userID).Return(privacy, nil)

	requesterID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String(), nil)
	req.Header.Set("X-User-Id", requesterID.String())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp dto.UserSearchResult

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), apiResp.UserID)
	assert.Equal(t, "publicuser", apiResp.Username)
}

func TestGetUserByIDComponent_NotFound_UserDoesNotExist(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	nonExistentID := uuid.New()

	mockRepo.On("FindUserByID", mock.Anything, nonExistentID).Return(nil, repository.ErrUserNotFound)

	requesterID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+nonExistentID.String(), nil)
	req.Header.Set("X-User-Id", requesterID.String())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
}

func TestGetUserByIDComponent_NotFound_PrivateProfile(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	now := time.Now()

	user := &dto.User{
		UserID:    userID.String(),
		Username:  "privateuser",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	privacy := &dto.PrivacyPreferences{
		ProfileVisibility: "private",
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, userID).Return(privacy, nil)

	requesterID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String(), nil)
	req.Header.Set("X-User-Id", requesterID.String())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Private profile returns 404, not 403
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
}

func TestGetUserByIDComponent_NotFound_FollowersOnlyProfile(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	now := time.Now()

	user := &dto.User{
		UserID:    userID.String(),
		Username:  "followersuser",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	privacy := &dto.PrivacyPreferences{
		ProfileVisibility: "followers_only",
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, userID).Return(privacy, nil)

	requesterID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String(), nil)
	req.Header.Set("X-User-Id", requesterID.String())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// followers_only returns 404 for non-followers
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetUserByIDComponent_NotFound_InactiveUser(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	now := time.Now()

	user := &dto.User{
		UserID:    userID.String(),
		Username:  "inactiveuser",
		IsActive:  false, // Inactive user
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)

	requesterID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String(), nil)
	req.Header.Set("X-User-Id", requesterID.String())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetUserByIDComponent_ValidationError_InvalidUUID(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	svc := service.NewUserService(mockRepo, mockTokenStore)

	c := &app.Container{
		UserService: svc,
		Config:      config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	requesterID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/not-a-uuid", nil)
	req.Header.Set("X-User-Id", requesterID.String())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}
