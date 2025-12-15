package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	errDB        = errors.New("db error")
	errMockArgs  = errors.New("mock: missing args")
	errStartType = errors.New("invalid type assertion for UserProfileResponse")
)

// MockUserService is a mock implementation of service.UserService.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserProfile(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
) (*dto.UserProfileResponse, error) {
	args := m.Called(ctx, requesterID, targetUserID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs // Fix nilnil
	}

	if val, ok := args.Get(0).(*dto.UserProfileResponse); ok {
		return val, nil
	}

	return nil, errStartType // Fix err113
}

type userHandlerTestCase struct {
	name           string
	targetIDPath   string
	requesterIDHdr string
	mockRun        func(*MockUserService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

func TestUserHandlerGetUserProfile(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	requesterID := uuid.New()

	baseProfile := &dto.UserProfileResponse{
		UserID:    targetID.String(),
		Username:  "targetuser",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := getHandlerTestCases(targetID, requesterID, baseProfile)

	runUserHandlerTest(t, tests)
}

func getHandlerTestCases(
	targetID, requesterID uuid.UUID,
	baseProfile *dto.UserProfileResponse,
) []userHandlerTestCase {
	return []userHandlerTestCase{
		{
			name:           "Success",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, requesterID, targetID).Return(baseProfile, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, targetID.String())
				assert.Contains(t, body, "targetuser")
			},
		},
		{
			name:         "User Not Found",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, uuid.Nil, targetID).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:         "Profile Private",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, uuid.Nil, targetID).Return(nil, service.ErrProfilePrivate)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:         "Internal Error",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, uuid.Nil, targetID).Return(nil, errDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:         "Invalid ID Format",
			targetIDPath: "invalid-uuid",
			mockRun: func(m *MockUserService) {
				// Service is not called because ID validation fails first.
			},
			expectedStatus: http.StatusBadRequest,
		},
	}
}

func runUserHandlerTest(t *testing.T, tests []userHandlerTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockUserService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewUserHandler(mockSvc)

			r := chi.NewRouter()
			r.Get("/users/{user_id}/profile", h.GetUserProfile)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.targetIDPath+"/profile", nil)
			if tt.requesterIDHdr != "" {
				req.Header.Set("X-User-Id", tt.requesterIDHdr)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.String())
			}
		})
	}
}
