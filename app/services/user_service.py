"""User service for profile management operations."""

import secrets
from datetime import UTC, datetime, timedelta
from uuid import UUID

from fastapi import HTTPException, status
from sqlalchemy import or_, select
from sqlalchemy.orm import selectinload

from app.api.v1.schemas.request.user.user_account_delete_request import (
    UserAccountDeleteRequest,
)
from app.api.v1.schemas.request.user.user_confirm_account_delete_response import (
    UserConfirmAccountDeleteResponse,
)
from app.api.v1.schemas.request.user.user_profile_update_request import (
    UserProfileUpdateRequest,
)
from app.api.v1.schemas.response.user.user_account_delete_response import (
    UserAccountDeleteRequestResponse,
)
from app.api.v1.schemas.response.user.user_profile_response import UserProfileResponse
from app.api.v1.schemas.response.user.user_search_response import (
    UserSearchResponse,
    UserSearchResult,
)
from app.core.logging import get_logger
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.models.user.user import User
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.enums.preferences.profile_visibility_enum import ProfileVisibilityEnum
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError
from app.utils.privacy import PrivacyChecker

_log = get_logger(__name__)


class UserService:
    """Service for user profile operations."""

    def __init__(self, db: SqlDatabaseSession, redis: RedisDatabaseSession) -> None:
        """Initialize user service with database session."""
        self.db = db
        self.redis = redis
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

    async def request_account_deletion(
        self, user_id: UUID
    ) -> UserAccountDeleteRequestResponse:
        """Request account deletion and generate confirmation token.

        Args:
            user_id: The user's unique identifier

        Returns:
            UserAccountDeleteRequestResponse: Deletion request with confirmation token

        Raises:
            HTTPException: If user not found or database error
        """
        _log.info(f"Requesting account deletion for user: {user_id}")

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

            # Check if user is already inactive
            if not user.is_active:
                _log.warning(f"Account already inactive: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Account is already inactive",
                )

            # Generate confirmation token
            confirmation_token = secrets.token_urlsafe(32)
            expires_at = datetime.now(UTC) + timedelta(hours=24)

            # Store in Redis (refreshes any existing token)
            await self.redis.store_deletion_token(
                str(user_id), confirmation_token, expires_at
            )

            _log.info(f"Account deletion request created for user: {user_id}")

            return UserAccountDeleteRequestResponse(
                user_id=user.user_id,
                confirmation_token=confirmation_token,
                expires_at=expires_at,
            )

        except DatabaseError as e:
            _log.error(f"Database error requesting account deletion: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e

    async def confirm_account_deletion(
        self, user_id: UUID, delete_request: UserAccountDeleteRequest
    ) -> UserConfirmAccountDeleteResponse:
        """Confirm account deletion and deactivate the user.

        Args:
            user_id: The user's unique identifier
            delete_request: The deletion confirmation request

        Returns:
            dict: Simple success response

        Raises:
            HTTPException: If user not found, invalid token, or database error
        """
        _log.info(f"Confirming account deletion for user: {user_id}")

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

            # Check if user is already inactive
            if not user.is_active:
                _log.warning(f"Account already inactive: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Account is already inactive",
                )

            # Retrieve from Redis
            token_data = await self.redis.get_deletion_token(str(user_id))
            if not token_data:
                _log.warning(f"No deletion request found for user: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=(
                        "No deletion request found. Please request deletion first."
                    ),
                )
            if token_data["token"] != delete_request.confirmation_token:
                _log.warning(f"Invalid confirmation token for user: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Invalid confirmation token",
                )
            # Check if token has expired
            expires_at = datetime.fromisoformat(token_data["expires_at"])
            current_time = datetime.now(UTC)
            if current_time > expires_at:
                _log.warning(f"Expired confirmation token for user: {user_id}")
                await self.redis.delete_deletion_token(str(user_id))
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=(
                        "Confirmation token has expired. Please request deletion again."
                    ),
                )
            # Deactivate the user (soft delete)
            user.is_active = False

            # Commit changes
            await self.db.commit()

            # Clean up the token
            await self.redis.delete_deletion_token(str(user_id))

            _log.info(f"Account successfully deactivated for user: {user_id}")

            return UserConfirmAccountDeleteResponse(
                user_id=user.user_id,
                deactivated_at=datetime.now(UTC),
            )
        except DatabaseError as e:
            _log.error(f"Database error confirming account deletion: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e

    async def search_users(
        self,
        requester_user_id: UUID,
        query: str | None = None,
        limit: int = 20,
        offset: int = 0,
        count_only: bool = False,
    ) -> UserSearchResponse:
        """Search users by username or full name, applying privacy checks.

        Args:
            requester_user_id: The user making the request
            query: Search query (username or full_name)
            limit: Max results
            offset: Results to skip
            count_only: If True, only return count

        Returns:
            UserSearchResponse: Paginated user search response
        """
        try:
            _log.info(
                f"User search by {requester_user_id}: query='{query}', limit={limit}, "
                f"offset={offset}, count_only={count_only}"
            )
            # Build base query with eager loading
            stmt = (
                select(User)
                .options(selectinload(User.privacy_preferences))
                .where(User.is_active.is_(True))
            )
            if query:
                stmt = stmt.where(
                    or_(
                        User.username.ilike(f"%{query}%"),
                        User.full_name.ilike(f"%{query}%"),
                    )
                )
            stmt = stmt.offset(offset).limit(limit)
            db_result = await self.db.execute(stmt)
            users = db_result.scalars().all()

            # Count total (without pagination, eager load for consistency)
            count_stmt = (
                select(User)
                .options(selectinload(User.privacy_preferences))
                .where(User.is_active.is_(True))
            )
            if query:
                count_stmt = count_stmt.where(
                    or_(
                        User.username.ilike(f"%{query}%"),
                        User.full_name.ilike(f"%{query}%"),
                    )
                )
            count_result = await self.db.execute(count_stmt)
            total_count = len(count_result.scalars().unique().all())

            # Privacy check
            privacy_checker = PrivacyChecker(self.db)
            visible_users = [
                user
                for user in users
                if await privacy_checker.check_access(user, requester_user_id)
            ]

            results = [
                UserSearchResult(
                    user_id=user.user_id,
                    username=user.username,
                    full_name=user.full_name,
                    is_active=user.is_active,
                    created_at=user.created_at,
                    updated_at=user.updated_at,
                )
                for user in visible_users
            ]
            if count_only:
                _log.info(
                    f"User search by {requester_user_id}: count_only requested, "
                    f"returning only total_count={total_count}"
                )
                return UserSearchResponse(
                    results=[], total_count=total_count, limit=limit, offset=offset
                )
            _log.info(
                f"User search by {requester_user_id}: returned {len(results)} of "
                f"{total_count} total"
            )
            return UserSearchResponse(
                results=results, total_count=total_count, limit=limit, offset=offset
            )
        except DatabaseError as e:
            _log.error(f"Database error during user search: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e

    async def get_user_by_id(self, user_id: str) -> User | None:
        """Get user by ID.

        Args:
            user_id: The user's unique identifier as string

        Returns:
            User | None: User object if found, None otherwise
        """
        try:
            user_uuid = UUID(user_id)
            stmt = select(User).where(User.user_id == user_uuid)
            result = await self.db.execute(stmt)
            return result.scalar_one_or_none()
        except (ValueError, DatabaseError) as e:
            _log.error(f"Error getting user by ID {user_id}: {e}")
            return None

    async def can_view_profile(
        self, requester_user_id: UUID, target_user_id: str
    ) -> bool:
        """Check if requester can view target user's profile.

        Args:
            requester_user_id: The user making the request
            target_user_id: The target user's ID as string

        Returns:
            bool: True if access is allowed, False otherwise
        """
        target_user = await self.get_user_by_id(target_user_id)
        if not target_user:
            return False
        return await self.privacy_checker.check_access(target_user, requester_user_id)

    async def can_social_interact(
        self, requester_user_id: UUID, target_user_id: str
    ) -> bool:
        """Check if requester can socially interact with target user.

        Args:
            requester_user_id: The user making the request
            target_user_id: The target user's ID as string

        Returns:
            bool: True if interaction is allowed, False otherwise
        """
        target_user = await self.get_user_by_id(target_user_id)
        if not target_user:
            return False
        # For now, use the same privacy check - can be extended later
        return await self.privacy_checker.check_access(target_user, requester_user_id)

    async def get_public_user_by_id(
        self, user_id: UUID, requester_user_id: UUID | None = None
    ) -> UserSearchResult:
        """Get public user profile by ID with privacy checks.

        Args:
            user_id: The user's unique identifier
            requester_user_id: The user making the request (optional for anonymous
                access)

        Returns:
            UserSearchResult: Public user profile data

        Raises:
            HTTPException: If user not found, forbidden, or database error
        """
        _log.info(
            f"Getting public profile for user: {user_id} by requester: "
            f"{requester_user_id}"
        )

        try:
            # Get user with privacy preferences
            result = await self.db.execute(
                select(User)
                .options(selectinload(User.privacy_preferences))
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
            # Handle anonymous access (requester_user_id is None)
            if requester_user_id is None:
                # For anonymous users, only allow access to PUBLIC profiles
                if not user.privacy_preferences:
                    # No privacy preferences means default to public
                    pass
                else:
                    profile_visibility = user.privacy_preferences.profile_visibility
                    if profile_visibility != ProfileVisibilityEnum.PUBLIC:
                        _log.warning(
                            f"Anonymous access denied for non-public profile: {user_id}"
                        )
                        raise HTTPException(
                            status_code=status.HTTP_404_NOT_FOUND,
                            detail="User not found",
                        )
            elif not await self.privacy_checker.check_access(user, requester_user_id):
                _log.warning(
                    f"Privacy access denied for user: {user_id} by requester: "
                    f"{requester_user_id}"
                )
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )

            # Return public profile data
            return UserSearchResult(
                user_id=user.user_id,
                username=user.username,
                full_name=user.full_name,
                is_active=user.is_active,
                created_at=user.created_at,
                updated_at=user.updated_at,
            )

        except DatabaseError as e:
            _log.error(f"Database error getting public user profile: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
