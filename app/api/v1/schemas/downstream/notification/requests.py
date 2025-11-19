"""Request schemas for notification service API."""

from pydantic import BaseModel, Field


class NewFollowerNotificationRequest(BaseModel):
    """Request to send new follower notification.

    Matches the notification service API specification for
    POST /notifications/new-follower endpoint.

    Attributes:
        recipient_ids: List of user IDs who should receive the notification
            (the users being followed). Max 100 recipients per request.
        follower_id: UUID of the user who is now following the recipients
    """

    recipient_ids: list[str] = Field(
        ...,
        min_length=1,
        max_length=100,
        description="List of recipient user IDs (users being followed)",
        examples=[
            [
                "550e8400-e29b-41d4-a716-446655440001",
                "550e8400-e29b-41d4-a716-446655440002",
            ]
        ],
    )
    follower_id: str = Field(
        ...,
        description="UUID of the new follower",
        examples=["550e8400-e29b-41d4-a716-446655440003"],
    )


class EmailChangedNotificationRequest(BaseModel):
    """Request to send email change notification.

    Matches the notification service API specification for
    POST /notifications/email-changed endpoint.

    This is a security-critical notification sent to both old and new
    email addresses when a user changes their email.

    Attributes:
        recipient_ids: Single user ID whose email changed. Max 1 recipient.
        old_email: Previous email address (notification sent here as security measure)
        new_email: New email address (notification sent here for confirmation)
    """

    recipient_ids: list[str] = Field(
        ...,
        min_length=1,
        max_length=1,
        description="Single recipient (user whose email changed)",
        examples=[["550e8400-e29b-41d4-a716-446655440001"]],
    )
    old_email: str = Field(
        ...,
        description=(
            "Previous email address (notification sent here as security measure)"
        ),
        examples=["old.email@example.com"],
    )
    new_email: str = Field(
        ...,
        description="New email address (notification sent here for confirmation)",
        examples=["new.email@example.com"],
    )
