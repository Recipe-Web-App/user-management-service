"""User preferences model."""

from datetime import UTC, datetime
from typing import Any
from uuid import UUID

from sqlalchemy import JSON, Boolean, DateTime, String, Text
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID  # noqa: N811
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.db.sql.models.base_sql_model import BaseSqlModel
from app.enums.preferences import NotificationPreferenceKey, PreferenceType


class UserPreferences(BaseSqlModel):
    """User preferences model.

    Stores user preferences for various features including notifications, privacy
    settings, and other customizable options.
    """

    __tablename__ = "user_preferences"

    # Primary key
    preference_id: Mapped[UUID] = mapped_column(
        PostgresUUID(as_uuid=True),
        primary_key=True,
        doc="Unique identifier for the preference record",
    )

    # Foreign key to user
    user_id: Mapped[UUID] = mapped_column(
        PostgresUUID(as_uuid=True),
        nullable=False,
        index=True,
        doc="Reference to the user who owns these preferences",
    )

    # Preference type and key
    preference_type: Mapped[str] = mapped_column(
        String(50),
        nullable=False,
        index=True,
        doc="Type of preference (e.g., 'notification', 'privacy', 'display')",
    )

    preference_key: Mapped[str] = mapped_column(
        String(100),
        nullable=False,
        doc="Specific preference key within the type",
    )

    # Preference value (flexible JSON storage)
    preference_value: Mapped[dict[str, Any]] = mapped_column(
        JSON,
        nullable=False,
        doc="JSON value for the preference (can be boolean, string, number, or object)",
    )

    # Metadata
    description: Mapped[str] = mapped_column(
        Text,
        nullable=True,
        doc="Human-readable description of what this preference controls",
    )

    is_active: Mapped[bool] = mapped_column(
        Boolean,
        nullable=False,
        default=True,
        doc="Whether this preference is currently active",
    )

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        default=datetime.now(UTC),
        doc="When this preference was created",
    )

    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        default=datetime.now(UTC),
        onupdate=datetime.now(UTC),
        doc="When this preference was last updated",
    )

    # Relationships
    user = relationship("User", back_populates="preferences")

    def __repr__(self) -> str:
        """Return string representation of the user preference."""
        return (
            f"<UserPreferences("
            f"preference_id={self.preference_id}, "
            f"user_id={self.user_id}, "
            f"type={self.preference_type}, "
            f"key={self.preference_key}"
            f")>"
        )

    @classmethod
    def create_notification_preference(
        cls,
        user_id: UUID,
        key: NotificationPreferenceKey,
        value: Any,
        description: str | None = None,
    ) -> "UserPreferences":
        """Create a notification preference.

        Args:
            user_id: The user's unique identifier
            key: The preference key from NotificationPreferenceKey enum
            value: The preference value
            description: Optional description of the preference

        Returns:
            UserPreferences: The created preference instance
        """
        return cls(
            user_id=user_id,
            preference_type=PreferenceType.NOTIFICATION.value,
            preference_key=key.value,
            preference_value=value,
            description=description,
        )

    @classmethod
    def get_default_notification_preferences(cls) -> dict[str, Any]:
        """Get default notification preferences.

        Returns:
            dict: Default notification preference values
        """
        return {
            NotificationPreferenceKey.EMAIL_NOTIFICATIONS.value: True,
            NotificationPreferenceKey.PUSH_NOTIFICATIONS.value: True,
            NotificationPreferenceKey.SMS_NOTIFICATIONS.value: False,
            NotificationPreferenceKey.MARKETING_EMAILS.value: False,
            NotificationPreferenceKey.SECURITY_ALERTS.value: True,
            NotificationPreferenceKey.ACTIVITY_SUMMARIES.value: True,
            NotificationPreferenceKey.RECIPE_RECOMMENDATIONS.value: True,
            NotificationPreferenceKey.SOCIAL_INTERACTIONS.value: True,
        }
