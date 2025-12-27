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

func (m *MockSocialRepoComponent) GetFollowers(
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

func (m *MockSocialRepoComponent) FollowUser(
	ctx context.Context,
	followerID, followeeID uuid.UUID,
) error {
	args := m.Called(ctx, followerID, followeeID)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf(mockErrorFmt, err)
	}

	return nil
}

func (m *MockSocialRepoComponent) UnfollowUser(
	ctx context.Context,
	followerID, followeeID uuid.UUID,
) error {
	args := m.Called(ctx, followerID, followeeID)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf(mockErrorFmt, err)
	}

	return nil
}

func (m *MockSocialRepoComponent) GetRecentRecipes(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.RecipeSummary, error) {
	args := m.Called(ctx, userID, limit)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf(mockErrorFmt, err)
	}

	recipes, _ := args.Get(0).([]dto.RecipeSummary)

	return recipes, nil
}

func (m *MockSocialRepoComponent) GetRecentFollows(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.UserSummary, error) {
	args := m.Called(ctx, userID, limit)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf(mockErrorFmt, err)
	}

	follows, _ := args.Get(0).([]dto.UserSummary)

	return follows, nil
}

func (m *MockSocialRepoComponent) GetRecentReviews(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.ReviewSummary, error) {
	args := m.Called(ctx, userID, limit)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf(mockErrorFmt, err)
	}

	reviews, _ := args.Get(0).([]dto.ReviewSummary)

	return reviews, nil
}

func (m *MockSocialRepoComponent) GetRecentFavorites(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.FavoriteSummary, error) {
	args := m.Called(ctx, userID, limit)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf(mockErrorFmt, err)
	}

	favorites, _ := args.Get(0).([]dto.FavoriteSummary)

	return favorites, nil
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

	var apiResp dto.GetFollowedUsersResponse

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, 2, apiResp.TotalCount)
	assert.Len(t, apiResp.FollowedUsers, 2)
	require.NotNil(t, apiResp.Limit)
	require.NotNil(t, apiResp.Offset)
	assert.Equal(t, 20, *apiResp.Limit)
	assert.Equal(t, 0, *apiResp.Offset)

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
		"/api/v1/user-management/users/"+targetUserID.String()+"/following?countOnly=true",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":42`)
	// countOnly mode should not include followedUsers, limit, offset
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

// ============================================================================
// GetFollowers Component Tests
// ============================================================================

func TestGetFollowersComponent_Success(t *testing.T) {
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
	followers := createFollowedUsersComponent(2)

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
	mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 20, 0).Return(followers, 2, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/followers", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp dto.GetFollowedUsersResponse

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, 2, apiResp.TotalCount)
	assert.Len(t, apiResp.FollowedUsers, 2)
	require.NotNil(t, apiResp.Limit)
	require.NotNil(t, apiResp.Offset)
	assert.Equal(t, 20, *apiResp.Limit)
	assert.Equal(t, 0, *apiResp.Offset)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowersComponent_CountOnly(t *testing.T) {
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
	mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 20, 0).Return([]dto.User{}, 42, nil).Once()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/followers?countOnly=true",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":42`)
	// countOnly mode should not include followedUsers, limit, offset
	assert.NotContains(t, rr.Body.String(), `"followedUsers"`)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowersComponent_WithPagination(t *testing.T) {
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
	followers := createFollowedUsersComponent(5)

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
	mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 50, 10).Return(followers, 100, nil).Once()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/followers?limit=50&offset=10",
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

func TestGetFollowersComponent_OwnProfile(t *testing.T) {
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
	followers := createFollowedUsersComponent(3)

	// When viewing own profile, privacy check is skipped
	mockUserRepo.On("FindUserByID", mock.Anything, userID).Return(targetUser, nil).Once()
	mockSocialRepo.On("GetFollowers", mock.Anything, userID, 20, 0).Return(followers, 3, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String()+"/followers", nil)
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

func TestGetFollowersComponent_NotFound(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/followers", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertNotCalled(t, "GetFollowers", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFollowersComponent_Forbidden_PrivateProfile(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/followers", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertNotCalled(t, "GetFollowers", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFollowersComponent_Forbidden_FollowersOnlyNotFollowing(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/followers", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertNotCalled(t, "GetFollowers", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFollowersComponent_Success_FollowersOnlyWhenFollowing(t *testing.T) {
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
	followers := createFollowedUsersComponent(1)

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(followersOnlyPrivacy, nil).Once()
	mockUserRepo.On("IsFollowing", mock.Anything, requesterID, targetUserID).Return(true, nil).Once()
	mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 20, 0).Return(followers, 1, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/followers", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"totalCount":1`)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetFollowersComponent_Unauthorized(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/followers", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
}

//nolint:dupl // component tests intentionally mirror GetFollowing tests
func TestGetFollowersComponent_ValidationError_InvalidUUID(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/invalid-uuid/followers", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

func TestGetFollowersComponent_ValidationError_InvalidLimit(t *testing.T) {
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
		"/api/v1/user-management/users/"+targetUserID.String()+"/followers?limit=0",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

// FollowUser Component Tests

const followUserTestTargetEmail = "target@example.com"

func targetEmailPtr() *string {
	s := followUserTestTargetEmail

	return &s
}

func TestFollowUserComponent_Success(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	// Target user exists and is active
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: true,
	}, nil)

	// Target user allows follows
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(&dto.PrivacyPreferences{
		AllowFollows: true,
	}, nil)

	// Follow succeeds
	mockSocialRepo.On("FollowUser", mock.Anything, followerID, targetUserID).Return(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Successfully followed user")
	assert.Contains(t, rr.Body.String(), `"isFollowing":true`)
}

func TestFollowUserComponent_Success_AdminOverride(t *testing.T) {
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

	adminID := uuid.New()
	followerID := uuid.New()
	targetUserID := uuid.New()

	// Target user exists and is active
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: true,
	}, nil)

	// Target user allows follows
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(&dto.PrivacyPreferences{
		AllowFollows: true,
	}, nil)

	// Follow succeeds
	mockSocialRepo.On("FollowUser", mock.Anything, followerID, targetUserID).Return(nil)

	// Admin creates follow on behalf of another user
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", adminID.String())
	req.Header.Set("X-User-Role", "admin")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Successfully followed user")
}

func TestFollowUserComponent_BadRequest_SelfFollow(t *testing.T) {
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

	// Attempt to follow self
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+userID.String()+"/follow/"+userID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	assert.Contains(t, rr.Body.String(), "Cannot follow yourself")
}

func TestFollowUserComponent_NotFound_TargetDoesNotExist(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	// Target user not found
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(nil, repository.ErrUserNotFound)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
}

func TestFollowUserComponent_NotFound_TargetInactive(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	// Target user exists but is inactive
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: false,
	}, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
}

func TestFollowUserComponent_Forbidden_UserDoesNotAllowFollows(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	// Target user exists and is active
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: true,
	}, nil)

	// Target user does not allow follows
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(&dto.PrivacyPreferences{
		AllowFollows: false,
	}, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")
}

func TestFollowUserComponent_Forbidden_UserIdMismatch(t *testing.T) {
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
	differentUserID := uuid.New()
	targetUserID := uuid.New()

	// Non-admin trying to follow on behalf of another user
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+differentUserID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")
}

func TestFollowUserComponent_Unauthorized_MissingHeader(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	// Missing X-User-Id header
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
}

func TestFollowUserComponent_ValidationError_InvalidUserUUID(t *testing.T) {
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
	targetUserID := uuid.New()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/invalid-uuid/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

func TestFollowUserComponent_ValidationError_InvalidTargetUUID(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user-management/users/"+requesterID.String()+"/follow/invalid-uuid",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

// UnfollowUser component tests.

//nolint:dupl // Test structure intentionally mirrors FollowUser tests for consistency
func TestUnfollowUserComponent_Success(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	// Target user exists and is active
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: true,
	}, nil)
	mockSocialRepo.On("UnfollowUser", mock.Anything, followerID, targetUserID).Return(nil)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Successfully unfollowed user")
	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestUnfollowUserComponent_Success_AdminOverride(t *testing.T) {
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

	adminID := uuid.New()
	userID := uuid.New()
	targetUserID := uuid.New()

	// Target user exists and is active
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: true,
	}, nil)
	mockSocialRepo.On("UnfollowUser", mock.Anything, userID, targetUserID).Return(nil)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+userID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", adminID.String())
	req.Header.Set("X-User-Role", "admin")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Successfully unfollowed user")
	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

//nolint:dupl // Test structure intentionally mirrors FollowUser tests for consistency
func TestUnfollowUserComponent_Success_Idempotent(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	// Target user exists and is active
	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: true,
	}, nil)
	// Unfollow returns nil even if not following (idempotent)
	mockSocialRepo.On("UnfollowUser", mock.Anything, followerID, targetUserID).Return(nil)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Successfully unfollowed user")
	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestUnfollowUserComponent_BadRequest_SelfUnfollow(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+userID.String()+"/follow/"+userID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Cannot unfollow yourself")
}

func TestUnfollowUserComponent_NotFound_TargetDoesNotExist(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(nil, repository.ErrUserNotFound)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	mockUserRepo.AssertExpectations(t)
}

func TestUnfollowUserComponent_NotFound_TargetInactive(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(&dto.User{
		UserID:   targetUserID.String(),
		Username: "targetuser",
		Email:    targetEmailPtr(),
		IsActive: false,
	}, nil)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", followerID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	mockUserRepo.AssertExpectations(t)
}

//nolint:dupl // Test structure intentionally mirrors FollowUser tests for consistency
func TestUnfollowUserComponent_Forbidden_UserIdMismatch(t *testing.T) {
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
	pathUserID := uuid.New()
	targetUserID := uuid.New()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+pathUserID.String()+"/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Cannot perform unfollow action for another user")
}

func TestUnfollowUserComponent_Unauthorized_MissingHeader(t *testing.T) {
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

	followerID := uuid.New()
	targetUserID := uuid.New()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+followerID.String()+"/follow/"+targetUserID.String(),
		nil,
	)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
}

func TestUnfollowUserComponent_ValidationError_InvalidUserUUID(t *testing.T) {
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
	targetUserID := uuid.New()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/invalid-uuid/follow/"+targetUserID.String(),
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

func TestUnfollowUserComponent_ValidationError_InvalidTargetUUID(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/user-management/users/"+requesterID.String()+"/follow/invalid-uuid",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

// ============================================================================
// GetUserActivity Component Tests
// ============================================================================

func createTestActivityDataComponent() (
	[]dto.RecipeSummary, []dto.UserSummary, []dto.ReviewSummary, []dto.FavoriteSummary,
) {
	now := time.Now()
	comment := "Great recipe!"

	recipes := []dto.RecipeSummary{
		{RecipeID: 1, Title: "Pasta Carbonara", CreatedAt: now},
		{RecipeID: 2, Title: "Chicken Tikka", CreatedAt: now.Add(-time.Hour)},
	}

	follows := []dto.UserSummary{
		{UserID: uuid.New().String(), Username: "chef1", FollowedAt: now},
	}

	reviews := []dto.ReviewSummary{
		{ReviewID: 1, RecipeID: 1, Rating: 5, Comment: &comment, CreatedAt: now},
	}

	favorites := []dto.FavoriteSummary{
		{RecipeID: 1, Title: "Apple Pie", FavoritedAt: now},
	}

	return recipes, follows, reviews, favorites
}

func TestGetUserActivityComponent_Success_Authenticated(t *testing.T) {
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
	recipes, follows, reviews, favorites := createTestActivityDataComponent()

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
	mockSocialRepo.On("GetRecentRecipes", mock.Anything, targetUserID, 15).Return(recipes, nil).Once()
	mockSocialRepo.On("GetRecentFollows", mock.Anything, targetUserID, 15).Return(follows, nil).Once()
	mockSocialRepo.On("GetRecentReviews", mock.Anything, targetUserID, 15).Return(reviews, nil).Once()
	mockSocialRepo.On("GetRecentFavorites", mock.Anything, targetUserID, 15).Return(favorites, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/activity", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp dto.UserActivityResponse

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, targetUserID.String(), apiResp.UserID)
	assert.Len(t, apiResp.RecentRecipes, 2)
	assert.Len(t, apiResp.RecentFollows, 1)
	assert.Len(t, apiResp.RecentReviews, 1)
	assert.Len(t, apiResp.RecentFavorites, 1)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetUserActivityComponent_Success_Anonymous(t *testing.T) {
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

	targetUser := createTestUserComponent(targetUserID, "targetuser")
	publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}
	recipes, follows, reviews, favorites := createTestActivityDataComponent()

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
	mockSocialRepo.On("GetRecentRecipes", mock.Anything, targetUserID, 15).Return(recipes, nil).Once()
	mockSocialRepo.On("GetRecentFollows", mock.Anything, targetUserID, 15).Return(follows, nil).Once()
	mockSocialRepo.On("GetRecentReviews", mock.Anything, targetUserID, 15).Return(reviews, nil).Once()
	mockSocialRepo.On("GetRecentFavorites", mock.Anything, targetUserID, 15).Return(favorites, nil).Once()

	// No X-User-Id header - anonymous access
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/activity", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), targetUserID.String())

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetUserActivityComponent_Success_CustomLimit(t *testing.T) {
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
	mockSocialRepo.On("GetRecentRecipes", mock.Anything, targetUserID, 50).Return([]dto.RecipeSummary{}, nil).Once()
	mockSocialRepo.On("GetRecentFollows", mock.Anything, targetUserID, 50).Return([]dto.UserSummary{}, nil).Once()
	mockSocialRepo.On("GetRecentReviews", mock.Anything, targetUserID, 50).Return([]dto.ReviewSummary{}, nil).Once()
	mockSocialRepo.On("GetRecentFavorites", mock.Anything, targetUserID, 50).Return([]dto.FavoriteSummary{}, nil).Once()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/activity?per_type_limit=50",
		nil,
	)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetUserActivityComponent_Success_OwnProfile(t *testing.T) {
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
	recipes, follows, reviews, favorites := createTestActivityDataComponent()

	mockUserRepo.On("FindUserByID", mock.Anything, userID).Return(targetUser, nil).Once()
	mockSocialRepo.On("GetRecentRecipes", mock.Anything, userID, 15).Return(recipes, nil).Once()
	mockSocialRepo.On("GetRecentFollows", mock.Anything, userID, 15).Return(follows, nil).Once()
	mockSocialRepo.On("GetRecentReviews", mock.Anything, userID, 15).Return(reviews, nil).Once()
	mockSocialRepo.On("GetRecentFavorites", mock.Anything, userID, 15).Return(favorites, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String()+"/activity", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Privacy preferences should NOT be fetched when viewing own activity
	mockUserRepo.AssertNotCalled(t, "FindPrivacyPreferencesByUserID", mock.Anything, mock.Anything)
	mockUserRepo.AssertExpectations(t)
	mockSocialRepo.AssertExpectations(t)
}

func TestGetUserActivityComponent_NotFound(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/activity", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserActivityComponent_Forbidden_PrivateProfile(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/activity", nil)
	req.Header.Set("X-User-Id", requesterID.String())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserActivityComponent_Forbidden_FollowersOnlyAnonymous(t *testing.T) {
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

	targetUser := createTestUserComponent(targetUserID, "followersonly")
	followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

	mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(followersOnlyPrivacy, nil).Once()

	// Anonymous request (no X-User-Id header)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+targetUserID.String()+"/activity", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserActivityComponent_ValidationError_InvalidUUID(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/invalid-uuid/activity", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

func TestGetUserActivityComponent_ValidationError_InvalidPerTypeLimit(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/activity?per_type_limit=0",
		nil,
	)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}

func TestGetUserActivityComponent_ValidationError_PerTypeLimitTooHigh(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/user-management/users/"+targetUserID.String()+"/activity?per_type_limit=101",
		nil,
	)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
}
