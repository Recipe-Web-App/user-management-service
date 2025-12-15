package component_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
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

func TestUserProfileComponent(t *testing.T) {
	t.Parallel()

	// Create Mock Repo
	mockRepo := new(MockUserRepo)

	// Create Service using Mock Repo
	svc := service.NewUserService(mockRepo)

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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/users/"+userID.String()+"/profile", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Assert Response
	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                    `json:"success"`
		Data    dto.UserProfileResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	resp := apiResp.Data

	assert.Equal(t, userID.String(), resp.UserID)
	assert.NotNil(t, resp.FullName)
	assert.Equal(t, "Component User", *resp.FullName)
	assert.NotNil(t, resp.Email)
	assert.Equal(t, "test@example.com", *resp.Email)
}
