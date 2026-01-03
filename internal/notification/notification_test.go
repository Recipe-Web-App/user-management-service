package notification_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/notification"
)

// MockDownstreamClient is a mock implementation of DownstreamClient.
type MockDownstreamClient struct {
	mock.Mock
}

func (m *MockDownstreamClient) Do(
	ctx context.Context,
	method, path string,
	body any,
	response any,
) error {
	args := m.Called(ctx, method, path, body, response)

	return args.Error(0) //nolint:wrapcheck // mock errors don't need wrapping
}

func TestNotificationClient_NotifyNewFollower_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDownstreamClient)
	recipientID := uuid.New()
	followerID := uuid.New()

	mockClient.On("Do",
		mock.Anything,
		"POST",
		"/notifications/new-follower",
		mock.MatchedBy(func(req notification.NewFollowerRequest) bool {
			return len(req.RecipientIDs) == 1 &&
				req.RecipientIDs[0] == recipientID.String() &&
				req.FollowerID == followerID.String()
		}),
		mock.Anything,
	).Return(nil)

	client := notification.NewNotificationClient(mockClient)
	client.NotifyNewFollower(context.Background(), recipientID, followerID)

	mockClient.AssertExpectations(t)
}

func TestNotificationClient_NotifyNewFollower_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDownstreamClient)
	recipientID := uuid.New()
	followerID := uuid.New()

	// Error is logged but not returned (fire-and-forget)
	mockClient.On("Do",
		mock.Anything,
		"POST",
		"/notifications/new-follower",
		mock.Anything,
		mock.Anything,
	).Return(assert.AnError)

	client := notification.NewNotificationClient(mockClient)

	// Should not panic even with error
	client.NotifyNewFollower(context.Background(), recipientID, followerID)

	mockClient.AssertExpectations(t)
}

func TestNotificationClient_NotifyEmailChanged_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDownstreamClient)
	recipientID := uuid.New()
	oldEmail := "old@example.com"
	newEmail := "new@example.com"

	mockClient.On("Do",
		mock.Anything,
		"POST",
		"/notifications/email-changed",
		mock.MatchedBy(func(req notification.EmailChangedRequest) bool {
			return len(req.RecipientIDs) == 1 &&
				req.RecipientIDs[0] == recipientID.String() &&
				req.OldEmail == oldEmail &&
				req.NewEmail == newEmail
		}),
		mock.Anything,
	).Return(nil)

	client := notification.NewNotificationClient(mockClient)
	client.NotifyEmailChanged(context.Background(), recipientID, oldEmail, newEmail)

	mockClient.AssertExpectations(t)
}

func TestNotificationClient_NotifyEmailChanged_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDownstreamClient)
	recipientID := uuid.New()

	// Error is logged but not returned (fire-and-forget)
	mockClient.On("Do",
		mock.Anything,
		"POST",
		"/notifications/email-changed",
		mock.Anything,
		mock.Anything,
	).Return(notification.ErrNotificationFailed)

	client := notification.NewNotificationClient(mockClient)

	// Should not panic even with error
	client.NotifyEmailChanged(context.Background(), recipientID, "old@test.com", "new@test.com")

	mockClient.AssertExpectations(t)
}

func TestNoopClient_NotifyNewFollower(t *testing.T) {
	t.Parallel()

	client := &notification.NoopClient{}

	// Should not panic
	client.NotifyNewFollower(context.Background(), uuid.New(), uuid.New())
}

func TestNoopClient_NotifyEmailChanged(t *testing.T) {
	t.Parallel()

	client := &notification.NoopClient{}

	// Should not panic
	client.NotifyEmailChanged(context.Background(), uuid.New(), "old@test.com", "new@test.com")
}
