"""User model for the database."""

from uuid import uuid4

from sqlalchemy import Boolean, Column, DateTime, String, Text
from sqlalchemy.dialects.postgresql import ENUM as SAEnum  # noqa: N811
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.sql import func

from app.db.sql.models.base_sql_model import BaseSqlModel
from app.enums.user_role_enum import UserRoleEnum
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
    role: Mapped[str] = mapped_column(
        SAEnum(
            UserRoleEnum,
            name="user_role_enum",
            schema="recipe_manager",
            create_constraint=False,
        ),
        nullable=False,
        default=UserRoleEnum.USER,
        server_default=UserRoleEnum.USER.value,
        doc="Role of the user: ADMIN or USER.",
    )

    notifications = relationship("Notification", back_populates="user")
    notification_preferences = relationship(
        "UserNotificationPreferences", uselist=False, back_populates="user"
    )
    display_preferences = relationship(
        "UserDisplayPreferences", uselist=False, back_populates="user"
    )
    theme_preferences = relationship(
        "UserThemePreferences", uselist=False, back_populates="user"
    )
    privacy_preferences = relationship(
        "UserPrivacyPreferences", uselist=False, back_populates="user"
    )
    security_preferences = relationship(
        "UserSecurityPreferences", uselist=False, back_populates="user"
    )
    sound_preferences = relationship(
        "UserSoundPreferences", uselist=False, back_populates="user"
    )
    social_preferences = relationship(
        "UserSocialPreferences", uselist=False, back_populates="user"
    )
    language_preferences = relationship(
        "UserLanguagePreferences", uselist=False, back_populates="user"
    )
    accessibility_preferences = relationship(
        "UserAccessibilityPreferences", uselist=False, back_populates="user"
    )

    # Association table relationships for followers/following
    following_associations = relationship(
        "UserFollows",
        foreign_keys="UserFollows.follower_id",
        back_populates="follower",
        cascade="all, delete-orphan",
    )
    follower_associations = relationship(
        "UserFollows",
        foreign_keys="UserFollows.followee_id",
        back_populates="followee",
        cascade="all, delete-orphan",
    )

    # Convenient access to User objects
    following = relationship(
        "User",
        secondary="recipe_manager.user_follows",
        primaryjoin="User.user_id==UserFollows.follower_id",
        secondaryjoin="User.user_id==UserFollows.followee_id",
        viewonly=True,
        backref="followers",
    )

    recipes = relationship(
        "Recipe",
        back_populates="user",
        cascade="all, delete-orphan",
        doc="List of recipes created by the user.",
    )

    recipe_reviews = relationship(
        "RecipeReview",
        back_populates="user",
        cascade="all, delete-orphan",
        doc="List of recipe reviews written by the user.",
    )

    recipe_favorites = relationship(
        "RecipeFavorite",
        back_populates="user",
        cascade="all, delete-orphan",
        doc="List of recipes favorited by the user.",
    )

    def __repr__(self) -> str:
        """Return string representation of the User model."""
        return (
            f"<User(user_id={self.user_id}, "
            f"username='{self.username}', email='{self.email}')>"
        )
