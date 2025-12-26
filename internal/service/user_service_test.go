package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/redis"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

const mockErrorFmt = "mock error: %w"
const testEmail = "email@example.com"
const testUserFullName = "Target User"
const testNewUsername = "newusername"

var (
	errMockArgs                = errors.New("mock: missing args")
	errMockInvalidUser         = errors.New("invalid type assertion for User")
	errMockInvalidPrivacy      = errors.New("invalid type assertion for PrivacyPreferences")
	errMockInvalidNotifPrefs   = errors.New("invalid type assertion for NotificationPreferences")
	errMockInvalidDisplayPrefs = errors.New("invalid type assertion for DisplayPreferences")
	errRedis                   = errors.New("redis error")
	errDB                      = errors.New("database error")
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
			return nil, fmt.Errorf(mockErrorFmt, err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.User); ok {
		return val, nil
	}

	return nil, errMockInvalidUser
}

func (m *MockUserRepository) FindPrivacyPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.PrivacyPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockErrorFmt, err) // wrapcheck
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.PrivacyPreferences); ok {
		return val, nil
	}

	return nil, errMockInvalidPrivacy
}

func (m *MockUserRepository) FindNotificationPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.NotificationPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockErrorFmt, err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.NotificationPreferences); ok {
		return val, nil
	}

	return nil, errMockInvalidNotifPrefs
}

func (m *MockUserRepository) FindDisplayPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.DisplayPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockErrorFmt, err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.DisplayPreferences); ok {
		return val, nil
	}

	return nil, errMockInvalidDisplayPrefs
}

func (m *MockUserRepository) IsFollowing(ctx context.Context, followerID, followedID uuid.UUID) (bool, error) {
	args := m.Called(ctx, followerID, followedID)
	// wrapcheck: Error might be nil, but if it's not, we should technically wrap it or just ignore.
	// However, for bool returns it's simpler.
	err := args.Error(1)
	if err != nil {
		return args.Bool(0), fmt.Errorf(mockErrorFmt, err)
	}

	return args.Bool(0), nil
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
			return nil, fmt.Errorf(mockErrorFmt, err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.User); ok {
		return val, nil
	}

	return nil, errMockInvalidUser
}

func (m *MockUserRepository) SearchUsers(
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
	args := m.Called(ctx)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf(mockErrorFmt, err)
	}

	result, _ := args.Get(0).(*dto.UserStatsResponse)

	return result, nil
}

// MockTokenStore is a mock implementation of repository.TokenStore.
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

type userServiceTestCase struct {
	name          string
	requesterID   uuid.UUID
	targetUser    *dto.User
	targetPrivacy *dto.PrivacyPreferences
	isFollowing   bool
	expectedErr   error
	validateResp  func(*testing.T, *dto.UserProfileResponse)
}

func TestUserServiceGetUserProfile(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	requesterID := uuid.New()
	followerID := uuid.New()

	baseUser := createBaseUser(targetID)

	tests := getUserServiceTestCases(targetID, requesterID, followerID, baseUser)

	runUserServiceTest(t, tests, targetID)
}

func getUserServiceTestCases(
	targetID, requesterID, followerID uuid.UUID,
	baseUser *dto.User,
) []userServiceTestCase {
	tests := getSuccessTestCases(targetID, requesterID, followerID, baseUser)
	tests = append(tests, getErrorTestCases(requesterID, baseUser)...)

	return tests
}

func getSuccessTestCases(
	targetID, requesterID, followerID uuid.UUID,
	baseUser *dto.User,
) []userServiceTestCase {
	return []userServiceTestCase{
		{
			name:        "Self Profile - Always Allow",
			requesterID: targetID,
			targetUser:  baseUser,
			targetPrivacy: &dto.PrivacyPreferences{
				ProfileVisibility: "private",
				ShowEmail:         false,
			},
			validateResp: func(t *testing.T, r *dto.UserProfileResponse) {
				t.Helper()
				assert.Equal(t, testEmail, *r.Email)
				assert.Equal(t, testUserFullName, *r.FullName)
			},
		},
		{
			name:        "Public Profile - Can View",
			requesterID: requesterID,
			targetUser:  baseUser,
			targetPrivacy: &dto.PrivacyPreferences{
				ProfileVisibility: "public",
				ShowEmail:         false,
				ShowFullName:      true,
			},
			validateResp: func(t *testing.T, r *dto.UserProfileResponse) {
				t.Helper()
				assert.Nil(t, r.Email)
				assert.Equal(t, testUserFullName, *r.FullName)
			},
		},
		{
			name:        "Followers Only - Following - Allow",
			requesterID: followerID,
			targetUser:  baseUser,
			targetPrivacy: &dto.PrivacyPreferences{
				ProfileVisibility: "followers_only",
				ShowEmail:         true,
			},
			isFollowing: true,
			validateResp: func(t *testing.T, r *dto.UserProfileResponse) {
				t.Helper()
				assert.Equal(t, testEmail, *r.Email)
			},
		},
	}
}

func getErrorTestCases(requesterID uuid.UUID, baseUser *dto.User) []userServiceTestCase {
	return []userServiceTestCase{
		{
			name:        "Private Profile - Not Self - Deny",
			requesterID: requesterID,
			targetUser:  baseUser,
			targetPrivacy: &dto.PrivacyPreferences{
				ProfileVisibility: "private",
			},
			expectedErr: service.ErrProfilePrivate,
		},
		{
			name:        "Followers Only - Not Following - Deny",
			requesterID: requesterID,
			targetUser:  baseUser,
			targetPrivacy: &dto.PrivacyPreferences{
				ProfileVisibility: "followers_only",
			},
			isFollowing: false,
			expectedErr: service.ErrProfilePrivate,
		},
		{
			name:        "User Not Found",
			targetUser:  nil,
			expectedErr: service.ErrUserNotFound,
		},
	}
}

func createBaseUser(targetID uuid.UUID) *dto.User {
	return &dto.User{
		UserID:    targetID.String(),
		Username:  "targetuser",
		Email:     func() *string { s := testEmail; return &s }(),
		FullName:  func() *string { s := testUserFullName; return &s }(),
		Bio:       func() *string { s := "Bio"; return &s }(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func runUserServiceTest(t *testing.T, tests []userServiceTestCase, targetID uuid.UUID) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)
			mockTokenStore := new(MockTokenStore)
			svc := service.NewUserService(mockRepo, mockTokenStore)

			ctx := context.Background()

			setupMockExpectations(ctx, mockRepo, tt, targetID)

			resp, err := svc.GetUserProfile(ctx, tt.requesterID, targetID)

			verifyTestResult(t, tt, resp, err)
		})
	}
}

func setupMockExpectations(
	ctx context.Context,
	mockRepo *MockUserRepository,
	tt userServiceTestCase,
	targetID uuid.UUID,
) {
	// Setup expectations
	if tt.targetUser == nil {
		mockRepo.On("FindUserByID", ctx, targetID).Return(nil, repository.ErrUserNotFound)
	} else {
		mockRepo.On("FindUserByID", ctx, targetID).Return(tt.targetUser, nil)
		mockRepo.On("FindPrivacyPreferencesByUserID", ctx, targetID).Return(tt.targetPrivacy, nil)

		if tt.targetPrivacy.ProfileVisibility == "followers_only" && tt.requesterID != targetID {
			mockRepo.On("IsFollowing", ctx, tt.requesterID, targetID).Return(tt.isFollowing, nil)
		}
	}
}

func verifyTestResult(
	t *testing.T,
	tt userServiceTestCase,
	resp *dto.UserProfileResponse,
	err error,
) {
	t.Helper()

	if tt.expectedErr != nil {
		require.Error(t, err)
		require.ErrorIs(t, err, tt.expectedErr) // testifylint
		assert.Nil(t, resp)
	} else {
		require.NoError(t, err)
		assert.NotNil(t, resp)

		if tt.validateResp != nil {
			tt.validateResp(t, resp)
		}
	}
}

func TestUserServiceUpdateUserProfile(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	userID := uuid.New()
	now := time.Now()

	baseUser := &dto.User{
		UserID:    userID.String(),
		Username:  "oldusername",
		Email:     func() *string { s := "old@email.com"; return &s }(),
		FullName:  func() *string { s := "Old Name"; return &s }(),
		Bio:       func() *string { s := "Old bio"; return &s }(),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name         string
		update       *dto.UserProfileUpdateRequest
		setupMock    func(*MockUserRepository)
		expectedErr  error
		validateResp func(*testing.T, *dto.UserProfileResponse)
	}{
		{
			name: "Success - Update Username",
			update: &dto.UserProfileUpdateRequest{
				Username: func() *string { s := testNewUsername; return &s }(),
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindUserByID", mock.Anything, userID).Return(baseUser, nil)
				updatedUser := *baseUser
				updatedUser.Username = testNewUsername
				m.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(&updatedUser, nil)
			},
			validateResp: func(t *testing.T, r *dto.UserProfileResponse) {
				t.Helper()
				assert.Equal(t, testNewUsername, r.Username)
			},
		},
		{
			name: "Success - Update All Fields",
			update: &dto.UserProfileUpdateRequest{
				Username: func() *string { s := testNewUsername; return &s }(),
				Email:    func() *string { s := "new@email.com"; return &s }(),
				FullName: func() *string { s := "New Name"; return &s }(),
				Bio:      func() *string { s := "New bio"; return &s }(),
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindUserByID", mock.Anything, userID).Return(baseUser, nil)
				updatedUser := &dto.User{
					UserID:    userID.String(),
					Username:  testNewUsername,
					Email:     func() *string { s := "new@email.com"; return &s }(),
					FullName:  func() *string { s := "New Name"; return &s }(),
					Bio:       func() *string { s := "New bio"; return &s }(),
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				}
				m.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(updatedUser, nil)
			},
			validateResp: func(t *testing.T, r *dto.UserProfileResponse) {
				t.Helper()
				assert.Equal(t, testNewUsername, r.Username)
				assert.Equal(t, "new@email.com", *r.Email)
				assert.Equal(t, "New Name", *r.FullName)
				assert.Equal(t, "New bio", *r.Bio)
			},
		},
		{
			name:   "Success - No Changes (empty request)",
			update: &dto.UserProfileUpdateRequest{},
			setupMock: func(m *MockUserRepository) {
				m.On("FindUserByID", mock.Anything, userID).Return(baseUser, nil)
				// UpdateUser should not be called
			},
			validateResp: func(t *testing.T, r *dto.UserProfileResponse) {
				t.Helper()
				assert.Equal(t, "oldusername", r.Username)
			},
		},
		{
			name: "User Not Found - From FindUserByID",
			update: &dto.UserProfileUpdateRequest{
				Username: func() *string { s := testNewUsername; return &s }(),
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindUserByID", mock.Anything, userID).Return(nil, repository.ErrUserNotFound)
			},
			expectedErr: service.ErrUserNotFound,
		},
		{
			name: "User Not Found - From UpdateUser",
			update: &dto.UserProfileUpdateRequest{
				Username: func() *string { s := testNewUsername; return &s }(),
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindUserByID", mock.Anything, userID).Return(baseUser, nil)
				m.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(nil, repository.ErrUserNotFound)
			},
			expectedErr: service.ErrUserNotFound,
		},
		{
			name: "Duplicate Username",
			update: &dto.UserProfileUpdateRequest{
				Username: func() *string { s := "existinguser"; return &s }(),
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindUserByID", mock.Anything, userID).Return(baseUser, nil)
				m.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(nil, repository.ErrDuplicateUsername)
			},
			expectedErr: service.ErrDuplicateUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)
			mockTokenStore := new(MockTokenStore)
			svc := service.NewUserService(mockRepo, mockTokenStore)

			tt.setupMock(mockRepo)

			resp, err := svc.UpdateUserProfile(context.Background(), userID, tt.update)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}
		})
	}
}

func TestUserServiceRequestAccountDeletion(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	userID := uuid.New()
	now := time.Now()

	baseUser := &dto.User{
		UserID:    userID.String(),
		Username:  "testuser",
		Email:     func() *string { s := testEmail; return &s }(),
		FullName:  func() *string { s := testUserFullName; return &s }(),
		Bio:       func() *string { s := "Bio"; return &s }(),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name          string
		setupMock     func(*MockUserRepository, *MockTokenStore)
		tokenStoreNil bool
		expectedErr   error
		validateResp  func(*testing.T, *dto.UserAccountDeleteRequestResponse)
	}{
		{
			name: "Success",
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				repo.On("FindUserByID", mock.Anything, userID).Return(baseUser, nil)
				ts.On("StoreDeleteToken", mock.Anything, userID, mock.Anything, service.DeleteTokenTTL).Return(nil)
			},
			validateResp: func(t *testing.T, r *dto.UserAccountDeleteRequestResponse) {
				t.Helper()
				assert.Equal(t, userID.String(), r.UserID)
				assert.NotEmpty(t, r.ConfirmationToken)
				assert.False(t, r.ExpiresAt.IsZero())
			},
		},
		{
			name:          "Cache Unavailable - Token Store Nil",
			tokenStoreNil: true,
			expectedErr:   service.ErrCacheUnavailable,
		},
		{
			name: "User Not Found",
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				repo.On("FindUserByID", mock.Anything, userID).Return(nil, repository.ErrUserNotFound)
			},
			expectedErr: service.ErrUserNotFound,
		},
		{
			name: "Cache Unavailable - Redis Error",
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				repo.On("FindUserByID", mock.Anything, userID).Return(baseUser, nil)
				ts.On("StoreDeleteToken", mock.Anything, userID, mock.Anything, service.DeleteTokenTTL).Return(errRedis)
			},
			expectedErr: service.ErrCacheUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)

			var tokenStore *MockTokenStore
			if !tt.tokenStoreNil {
				tokenStore = new(MockTokenStore)
			}

			var svc *service.UserServiceImpl
			if tt.tokenStoreNil {
				svc = service.NewUserService(mockRepo, nil)
			} else {
				svc = service.NewUserService(mockRepo, tokenStore)
			}

			if tt.setupMock != nil && tokenStore != nil {
				tt.setupMock(mockRepo, tokenStore)
			} else if tt.setupMock != nil {
				tt.setupMock(mockRepo, nil)
			}

			resp, err := svc.RequestAccountDeletion(context.Background(), userID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}
		})
	}
}

func TestUserServiceConfirmAccountDeletion(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	userID := uuid.New()
	token := uuid.New().String()
	now := time.Now()

	baseUser := &dto.User{
		UserID:    userID.String(),
		Username:  "testuser",
		Email:     func() *string { s := testEmail; return &s }(),
		FullName:  func() *string { s := testUserFullName; return &s }(),
		Bio:       func() *string { s := "Bio"; return &s }(),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name          string
		token         string
		setupMock     func(*MockUserRepository, *MockTokenStore)
		tokenStoreNil bool
		expectedErr   error
		validateResp  func(*testing.T, *dto.UserConfirmAccountDeleteResponse)
	}{
		{
			name:  "Success",
			token: token,
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				ts.On("GetDeleteToken", mock.Anything, userID).Return(token, nil)

				deactivatedUser := *baseUser
				deactivatedUser.IsActive = false
				repo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(&deactivatedUser, nil)
				ts.On("DeleteDeleteToken", mock.Anything, userID).Return(nil)
			},
			validateResp: func(t *testing.T, r *dto.UserConfirmAccountDeleteResponse) {
				t.Helper()
				assert.Equal(t, userID.String(), r.UserID)
				assert.False(t, r.DeactivatedAt.IsZero())
			},
		},
		{
			name:          "Cache Unavailable - Token Store Nil",
			token:         token,
			tokenStoreNil: true,
			expectedErr:   service.ErrCacheUnavailable,
		},
		{
			name:  "Invalid Token - Not Found in Cache",
			token: token,
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				ts.On("GetDeleteToken", mock.Anything, userID).Return("", redis.ErrTokenNotFound)
			},
			expectedErr: service.ErrInvalidToken,
		},
		{
			name:  "Invalid Token - Token Mismatch",
			token: "wrong-token",
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				ts.On("GetDeleteToken", mock.Anything, userID).Return(token, nil)
			},
			expectedErr: service.ErrInvalidToken,
		},
		{
			name:  "User Not Found During Deactivation",
			token: token,
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				ts.On("GetDeleteToken", mock.Anything, userID).Return(token, nil)
				repo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(nil, repository.ErrUserNotFound)
			},
			expectedErr: service.ErrUserNotFound,
		},
		{
			name:  "Success - Token Cleanup Error Ignored",
			token: token,
			setupMock: func(repo *MockUserRepository, ts *MockTokenStore) {
				ts.On("GetDeleteToken", mock.Anything, userID).Return(token, nil)

				deactivatedUser := *baseUser
				deactivatedUser.IsActive = false
				repo.On("UpdateUser", mock.Anything, userID, mock.Anything).Return(&deactivatedUser, nil)
				ts.On("DeleteDeleteToken", mock.Anything, userID).Return(errRedis) // Error ignored
			},
			validateResp: func(t *testing.T, r *dto.UserConfirmAccountDeleteResponse) {
				t.Helper()
				assert.Equal(t, userID.String(), r.UserID)
				assert.False(t, r.DeactivatedAt.IsZero())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)

			var tokenStore *MockTokenStore
			if !tt.tokenStoreNil {
				tokenStore = new(MockTokenStore)
			}

			var svc *service.UserServiceImpl
			if tt.tokenStoreNil {
				svc = service.NewUserService(mockRepo, nil)
			} else {
				svc = service.NewUserService(mockRepo, tokenStore)
			}

			if tt.setupMock != nil && tokenStore != nil {
				tt.setupMock(mockRepo, tokenStore)
			}

			resp, err := svc.ConfirmAccountDeletion(context.Background(), userID, tt.token)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}
		})
	}
}

func TestUserServiceGetUserStats(t *testing.T) {
	t.Parallel()

	expectedStats := &dto.UserStatsResponse{
		TotalUsers:        100,
		ActiveUsers:       80,
		InactiveUsers:     20,
		NewUsersToday:     5,
		NewUsersThisWeek:  15,
		NewUsersThisMonth: 30,
	}

	tests := []struct {
		name        string
		setupMock   func(*MockUserRepository)
		expectedErr error
		expected    *dto.UserStatsResponse
	}{
		{
			name: "Success",
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserStats", mock.Anything).Return(expectedStats, nil)
			},
			expected: expectedStats,
		},
		{
			name: "Repository Error",
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserStats", mock.Anything).Return(nil, errDB)
			},
			expectedErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)
			mockTokenStore := new(MockTokenStore)
			svc := service.NewUserService(mockRepo, mockTokenStore)

			tt.setupMock(mockRepo)

			stats, err := svc.GetUserStats(context.Background())

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, stats)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, stats)
			}
		})
	}
}
