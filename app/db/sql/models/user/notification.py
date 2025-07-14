"""Notification model for the database."""

from uuid import uuid4

from sqlalchemy import Boolean, Column, DateTime, ForeignKey, String, Text
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from app.db.sql.models.base_sql_model import BaseSqlModel


class Notification(BaseSqlModel):
    """Notification model representing the notifications table."""

    __tablename__ = "notifications"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    notification_id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    user_id = Column(
        UUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id"),
        nullable=False,
        index=True,
    )
    title = Column(String(255), nullable=False)
    message = Column(Text, nullable=False)
    notification_type = Column(String(50), nullable=False, index=True)
    is_read = Column(Boolean, default=False, nullable=False)
    is_deleted = Column(Boolean, default=False, nullable=False)
    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )
    updated_at = Column(
        DateTime(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )

    # Relationship to User model
    user = relationship("User", back_populates="notifications")

    def __repr__(self) -> str:
        """Return string representation of the Notification model."""
        return (
            f"<Notification(notification_id={self.notification_id}, "
            f"user_id={self.user_id}, title='{self.title}', "
            f"is_read={self.is_read})>"
        )
