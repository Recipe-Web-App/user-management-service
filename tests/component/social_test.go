package component_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// MockSocialRepoComponent for component tests.
type MockSocialRepoComponent struct {
	mock.Mock
}

func (m *MockSocialRepoComponent) GetFollowing(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf(mockErrorFmt, err)
	}

	users, _ := args.Get(0).([]dto.User)

	return users, args.Int(1), nil
}

func createTestUserComponent(userID uuid.UUID, username string) *dto.User {
	now := time.Now()
	fullName := "Test User"
	email := "test@example.com"

	return &dto.User{
		UserID:    userID.String(),
		Username:  username,
		Email:     &email,
		FullName:  &fullName,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func createFollowedUsersComponent(count int) []dto.User {
	users := make([]dto.User, count)
	now := time.Now()

	for i := range count {
		fullName := fmt.Sprintf("Followed User %d", i+1)
		users[i] = dto.User{
			UserID:    uuid.New().String(),
			Username:  fmt.Sprintf("followeduser%d", i+1),
			FullName:  &fullName,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return users
}

// ============================================================================
// GetFollowing Component Tests
// ============================================================================

func TestGetFollowingComponent_Success(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	targetUser := createTestUserComponent(targetUserID, "targetuser")
	publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}
	followedUsers := createFollowedUsersComponent(2)

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
	mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 20, 0).Return(followedUsers, 2, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/following", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                         `json:"success"`
		Data    dto.GetFollowedUsersResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)
	assert.Equal(t, 2, apiResp.Data.TotalCount)
	assert.Len(t, apiResp.Data.FollowedUsers, 2)
	require.NotNil(t, apiResp.Data.Limit)
	require.NotNil(t, apiResp.Data.Offset)
	assert.Equal(t, 20, *apiResp.Data.Limit)
	assert.Equal(t, 0, *apiResp.Data.Offset)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowingComponent_CountOnly(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	targetUser := createTestUserComponent(targetUserID, "targetuser")
	publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
	mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 20, 0).Return([]dto.User{}, 42, nil).Once()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/following?count_only=true",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":42`)
	// count_only mode should not include followedUsers, limit, offset
	assert.NotContains(t, rr.Body.String(), `"followedUsers"`)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowingComponent_WithPagination(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	targetUser := createTestUserComponent(targetUserID, "targetuser")
	publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}
	followedUsers := createFollowedUsersComponent(5)

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
	mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 50, 10).Return(followedUsers, 100, nil).Once()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/following?limit=50&offset=10",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":100`)
	assert.Contains(t, rr.Body.String(), `"limit":50`)
	assert.Contains(t, rr.Body.String(), `"offset":10`)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowingComponent_OwnProfile(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	targetUser := createTestUserComponent(userID, "ownuser")
	followedUsers := createFollowedUsersComponent(3)

	// When viewing own profile, privacy check is skipped
	mockUserRepo.On("FindUserByID", mock.Anything, userID).Return(targetUser, nil).Once()
	mockSocialRepo.On("GetFollowing", mock.Anything, userID, 20, 0).Return(followedUsers, 3, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String()+"/following", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":3`)

	// Privacy preferences should NOT be fetched when viewing own profile
	mockUserRepo.AssertNotCalled(t, "FindPrivacyPreferencesByUserID", mock.Anything, mock.Anything)
	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowingComponent_NotFound(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(nil, repository.ErrUserNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/following", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertNotCalled(t, "GetFollowing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFollowingComponent_Forbidden_PrivateProfile(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	targetUser := createTestUserComponent(targetUserID, "privateuser")
	privatePrivacy := &dto.PrivacyPreferences{ProfileVisibility: "private"}

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(privatePrivacy, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/following", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertNotCalled(t, "GetFollowing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFollowingComponent_Forbidden_FollowersOnlyNotFollowing(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	targetUser := createTestUserComponent(targetUserID, "followersonly")
	followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(followersOnlyPrivacy, nil).Once()
	mockUserRepo.On("IsFollowing", mock.Anything, requesterID, targetUserID).Return(false, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/following", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertNotCalled(t, "GetFollowing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFollowingComponent_Success_FollowersOnlyWhenFollowing(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	targetUser := createTestUserComponent(targetUserID, "followersonly")
	followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}
	followedUsers := createFollowedUsersComponent(1)

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(followersOnlyPrivacy, nil).Once()
	mockUserRepo.On("IsFollowing", mock.Anything, requesterID, targetUserID).Return(true, nil).Once()
	mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 20, 0).Return(followedUsers, 1, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/following", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":1`)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowingComponent_Unauthorized(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()

	// Missing X-User-Id header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/following", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
}

func TestGetFollowingComponent_ValidationError_InvalidUUID(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	requesterID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/invalid-uuid/following", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

func TestGetFollowingComponent_ValidationError_InvalidLimit(t *testing.T) {
	t.Parallel()

	mockUserRepo := new(MockUserRepo)
	mockSocialRepo := new(MockSocialRepoComponent)
	mockTokenStore := new(MockTokenStore)

	userSvc := service.NewUserService(mockUserRepo, mockTokenStore)
	socialSvc := service.NewSocialService(mockUserRepo, mockSocialRepo)

	c := &app.Container{
		UserService:   userSvc,
		SocialService: socialSvc,
		Config:        config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	targetUserID := uuid.New()
	requesterID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/following?limit=0",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}
