"""User privacy preferences model."""

from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy import Boolean, DateTime, Enum, ForeignKey
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID  # noqa: N811
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.db.sql.models.base_sql_model import BaseSqlModel
from app.enums.preferences.profile_visibility_enum import ProfileVisibilityEnum


class UserPrivacyPreferences(BaseSqlModel):
    """SQLAlchemy model for user privacy preferences."""

    __tablename__ = "user_privacy_preferences"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    id: Mapped[UUID] = mapped_column(PostgresUUID(as_uuid=True), primary_key=True)
    user_id: Mapped[UUID] = mapped_column(
        PostgresUUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        nullable=False,
        unique=True,
    )
    profile_visibility: Mapped[str] = mapped_column(
        Enum(
            ProfileVisibilityEnum,
            name="profile_visibility_enum",
            create_constraint=False,
        ),
        nullable=False,
        default=ProfileVisibilityEnum.PUBLIC,
    )
    recipe_visibility: Mapped[str] = mapped_column(
        Enum(
            ProfileVisibilityEnum,
            name="profile_visibility_enum",
            create_constraint=False,
        ),
        nullable=False,
        default=ProfileVisibilityEnum.PUBLIC,
    )
    activity_visibility: Mapped[str] = mapped_column(
        Enum(
            ProfileVisibilityEnum,
            name="profile_visibility_enum",
            create_constraint=False,
        ),
        nullable=False,
        default=ProfileVisibilityEnum.PUBLIC,
    )
    contact_info_visibility: Mapped[str] = mapped_column(
        Enum(
            ProfileVisibilityEnum,
            name="profile_visibility_enum",
            create_constraint=False,
        ),
        nullable=False,
        default=ProfileVisibilityEnum.PRIVATE,
    )
    data_sharing: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    analytics_tracking: Mapped[bool] = mapped_column(
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

    user = relationship("User", back_populates="privacy_preferences")
