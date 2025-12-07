"""User notification preferences model."""

from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy import Boolean, DateTime, ForeignKey
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID  # noqa: N811
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.db.sql.models.base_sql_model import BaseSqlModel


class UserNotificationPreferences(BaseSqlModel):
    """SQLAlchemy model for user notification preferences."""

    __tablename__ = "user_notification_preferences"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    id: Mapped[UUID] = mapped_column(PostgresUUID(as_uuid=True), primary_key=True)
    user_id: Mapped[UUID] = mapped_column(
        PostgresUUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        nullable=False,
        unique=True,
    )
    email_notifications: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=True
    )
    push_notifications: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=True
    )
    sms_notifications: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=False
    )
    marketing_emails: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=False
    )
    security_alerts: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    activity_summaries: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=True
    )
    recipe_recommendations: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=True
    )
    social_interactions: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=True
    )
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=datetime.now(UTC)
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        default=datetime.now(UTC),
        onupdate=datetime.now(UTC),
    )

    user = relationship("User", back_populates="notification_preferences")
