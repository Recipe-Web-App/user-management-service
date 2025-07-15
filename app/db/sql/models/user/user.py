"""User model for the database."""

from uuid import uuid4

from sqlalchemy import Boolean, Column, DateTime, String, Text
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from app.db.sql.models.base_sql_model import BaseSqlModel
from app.utils.security import SecurePasswordHash, SensitiveData


class User(BaseSqlModel):
    """User model representing the users table."""

    __tablename__ = "users"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    user_id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    username = Column(String(50), unique=True, nullable=False, index=True)
    email = Column(String(255), unique=True, nullable=False, index=True)
    password_hash: Column[SensitiveData | None] = Column(
        SecurePasswordHash(255), nullable=False
    )
    full_name = Column(String(255), nullable=True)
    bio = Column(Text, nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )
    updated_at = Column(
        DateTime(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )

    notifications = relationship("Notification", back_populates="user")
    preferences = relationship("UserPreferences", back_populates="user")

    def __repr__(self) -> str:
        """Return string representation of the User model."""
        return (
            f"<User(user_id={self.user_id}, "
            f"username='{self.username}', email='{self.email}')>"
        )
