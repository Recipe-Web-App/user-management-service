"""User sound preferences model."""

from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy import Boolean, DateTime, ForeignKey
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID  # noqa: N811
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.db.sql.models.base_sql_model import BaseSqlModel


class UserSoundPreferences(BaseSqlModel):
    """SQLAlchemy model for user sound preferences."""

    __tablename__ = "user_sound_preferences"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    id: Mapped[UUID] = mapped_column(PostgresUUID(as_uuid=True), primary_key=True)
    user_id: Mapped[UUID] = mapped_column(
        PostgresUUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        nullable=False,
        unique=True,
    )
    notification_sounds: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=True
    )
    system_sounds: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    volume_level: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    mute_notifications: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=False
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

    user = relationship("User", back_populates="sound_preferences")
