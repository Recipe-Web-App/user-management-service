"""User service for profile management operations."""

from uuid import UUID

from fastapi import HTTPException, status
from sqlalchemy import select
from sqlalchemy.orm import selectinload

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
