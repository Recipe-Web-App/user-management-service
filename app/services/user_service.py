"""User service for profile management operations."""

from uuid import UUID

from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from app.api.v1.schemas.request.user.user_profile_update_request import (
    UserProfileUpdateRequest,
)
from app.api.v1.schemas.response.user.user_profile_response import UserProfileResponse
from app.core.logging import get_logger
from app.db.sql.models.user.user import User
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError
from app.utils.privacy import PrivacyChecker

_log = get_logger(__name__)


class UserService:
    """Service for user profile operations."""

    def __init__(self, db: SqlDatabaseSession) -> None:
        """Initialize user service with database session."""
        self.db = db
        self.privacy_checker = PrivacyChecker(db)

    async def get_user_profile(
        self, user_id: UUID, requester_user_id: UUID
    ) -> UserProfileResponse:
        """Get user profile with privacy checks.

        Args:
            user_id: The user's unique identifier
            requester_user_id: The user making the request

        Returns:
            UserProfileResponse: User profile data with optional preferences

        Raises:
            HTTPException: If user not found, forbidden, or database error
        """
        _log.info(
            f"Getting profile for user: {user_id} by requester: {requester_user_id}"
        )

        try:
            # Get user with all preference relationships
            result = await self.db.execute(
                select(User)
                .options(
                    selectinload(User.privacy_preferences),
                    selectinload(User.notification_preferences),
                    selectinload(User.display_preferences),
                    selectinload(User.theme_preferences),
                    selectinload(User.security_preferences),
                    selectinload(User.sound_preferences),
                    selectinload(User.social_preferences),
                    selectinload(User.language_preferences),
                    selectinload(User.accessibility_preferences),
                )
                .where(User.user_id == user_id)
            )
            user = result.scalar_one_or_none()

            if not user:
                _log.warning(f"User not found: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )

            # Check if user is active
            if not user.is_active:
                _log.warning(f"Inactive user profile requested: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )

            # Check privacy preferences using centralized utility
            if not await self.privacy_checker.check_access(user, requester_user_id):
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied due to privacy settings",
                )

            # Check if contact info can be shared
            can_share_email = await self.privacy_checker.can_share_contact_info(
                user, requester_user_id
            )

            # Build response with privacy-controlled email
            return UserProfileResponse(
                user_id=user.user_id,
                username=user.username,
                email=user.email if can_share_email else None,
                full_name=user.full_name,
                bio=user.bio,
                is_active=user.is_active,
                created_at=user.created_at,
                updated_at=user.updated_at,
            )

        except DatabaseError as e:
            _log.error(f"Database error getting user profile: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e

    async def update_user_profile(
        self, user_id: UUID, update_data: UserProfileUpdateRequest
    ) -> UserProfileResponse:
        """Update user profile information.

        Args:
            user_id: The user's unique identifier
            update_data: The profile data to update

        Returns:
            UserProfileResponse: Updated user profile data

        Raises:
            HTTPException: If user not found, validation error, or database error
        """
        _log.info(f"Updating profile for user: {user_id}")

        try:
            # Get user
            result = await self.db.execute(select(User).where(User.user_id == user_id))
            user = result.scalar_one_or_none()

            if not user:
                _log.warning(f"User not found: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )

            # Check if user is active
            if not user.is_active:
                _log.warning(f"Inactive user profile update requested: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )

            # Check for username uniqueness if username is being updated
            if (
                update_data.username is not None
                and update_data.username != user.username
            ):
                existing_user = await self.db.execute(
                    select(User).where(User.username == update_data.username)
                )
                if existing_user.scalar_one_or_none():
                    _log.warning(f"Username already exists: {update_data.username}")
                    raise HTTPException(
                        status_code=status.HTTP_400_BAD_REQUEST,
                        detail="Username already registered",
                    )

            # Check for email uniqueness if email is being updated
            if update_data.email is not None and update_data.email != user.email:
                existing_user = await self.db.execute(
                    select(User).where(User.email == update_data.email)
                )
                if existing_user.scalar_one_or_none():
                    _log.warning(f"Email already exists: {update_data.email}")
                    raise HTTPException(
                        status_code=status.HTTP_400_BAD_REQUEST,
                        detail="Email already registered",
                    )

            # Update user fields
            if update_data.username is not None:
                user.username = update_data.username
                _log.info(
                    f"Updated username for user {user_id}: {update_data.username}"
                )

            if update_data.email is not None:
                user.email = update_data.email
                _log.info(f"Updated email for user {user_id}: {update_data.email}")

            if update_data.full_name is not None:
                user.full_name = update_data.full_name
                _log.info(
                    f"Updated full_name for user {user_id}: {update_data.full_name}"
                )

            if update_data.bio is not None:
                user.bio = update_data.bio
                _log.info(f"Updated bio for user {user_id}: {update_data.bio}")

            # Commit changes
            await self.db.commit()

            # Refresh the user object to get the latest data including updated_at
            await self.db.refresh(user)

            _log.info(f"Successfully updated profile for user: {user_id}")

            # Return updated profile
            return UserProfileResponse(
                user_id=user.user_id,
                username=user.username,
                email=user.email,
                full_name=user.full_name,
                bio=user.bio,
                is_active=user.is_active,
                created_at=user.created_at,
                updated_at=user.updated_at,
            )

        except DatabaseError as e:
            _log.error(f"Database error updating user profile: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
