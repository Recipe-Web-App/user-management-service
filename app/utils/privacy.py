"""Privacy utility functions for user data access control."""

from uuid import UUID

from sqlalchemy import select

from app.core.logging import get_logger
from app.db.sql.models.user.user import User
from app.db.sql.models.user.user_follows import UserFollows
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.enums.preferences.profile_visibility_enum import ProfileVisibilityEnum

_log = get_logger(__name__)


class PrivacyChecker:
    """Centralized privacy checking utility."""

    def __init__(self, db: SqlDatabaseSession | None = None) -> None:
        """Initialize privacy checker with optional database session.

        Args:
            db: Database session for friendship checks
        """
        self.db = db

    async def _is_friend(self, user: User, requester_user_id: UUID) -> bool:
        """Check if user follows the requester (friendship check).

        Args:
            user: The user to check
            requester_user_id: The user making the request

        Returns:
            bool: True if user follows requester, False otherwise
        """
        if not self.db:
            _log.debug(
                f"No database session available for friendship check: "
                f"user {user.user_id} -> requester {requester_user_id}"
            )
            return False

        try:
            result = await self.db.execute(
                select(UserFollows).where(
                    UserFollows.follower_id == user.user_id,
                    UserFollows.followee_id == requester_user_id,
                )
            )
            is_friend = result.scalar_one_or_none() is not None
            _log.debug(
                f"Friendship check: user {user.user_id} follows requester "
                f"{requester_user_id} = {is_friend}"
            )
        except Exception as e:
            _log.error(f"Error checking friendship status: {e}")
            return False
        else:
            return is_friend

    async def check_access(self, user: User, requester_user_id: UUID) -> bool:
        """Check if requester can access user's profile.

        Args:
            user: The user whose data is being accessed (with loaded preferences)
            requester_user_id: The user making the request

        Returns:
            bool: True if access is allowed, False otherwise
        """
        _log.debug(
            f"Checking access: user {user.user_id} -> requester {requester_user_id}"
        )

        # If requester is the same user, always allow access
        if requester_user_id == user.user_id:
            _log.debug(f"Access granted: same user {requester_user_id}")
            return True

        if not user.privacy_preferences:
            _log.debug(
                f"No privacy preferences for user {user.user_id}, "
                f"defaulting to PUBLIC"
            )
            return True

        # Get the profile visibility setting
        visibility = user.privacy_preferences.profile_visibility
        _log.debug(f"Profile visibility for user {user.user_id}: {visibility}")

        if visibility == ProfileVisibilityEnum.PUBLIC:
            _log.debug(
                f"Access allowed: user {user.user_id} has PUBLIC profile visibility"
            )
            return True

        if visibility == ProfileVisibilityEnum.PRIVATE:
            _log.debug(
                f"Access denied: user {user.user_id} has PRIVATE profile visibility"
            )
            return False

        if visibility == ProfileVisibilityEnum.FRIENDS_ONLY:
            # Check if user follows requester (friendship check)
            is_friend = await self._is_friend(user, requester_user_id)
            _log.debug(
                f"FRIENDS_ONLY profile visibility: user {user.user_id} -> "
                f"requester {requester_user_id} = {is_friend}"
            )
            return is_friend

        _log.debug(
            f"Access denied: user {user.user_id} -> requester {requester_user_id} "
            f"(no matching conditions)"
        )
        return False

    async def can_share_contact_info(self, user: User, requester_user_id: UUID) -> bool:
        """Check if contact info (email) can be shared with the requester.

        Args:
            user: The user whose contact info is being accessed (with loaded
                preferences)
            requester_user_id: The user making the request

        Returns:
            bool: True if contact info can be shared, False otherwise
        """
        _log.debug(
            f"Checking contact info sharing: user {user.user_id} -> "
            f"requester {requester_user_id}"
        )

        # If requester is the same user, always allow access
        if requester_user_id == user.user_id:
            _log.debug(
                f"Contact info sharing allowed: user {user.user_id} "
                f"accessing their own contact info"
            )
            return True

        if not user.privacy_preferences:
            _log.debug(
                f"No privacy preferences for user {user.user_id}, "
                f"defaulting to PUBLIC"
            )
            return True

        # Get contact info visibility from the loaded user object
        contact_visibility = user.privacy_preferences.contact_info_visibility
        _log.debug(
            f"Contact info visibility for user {user.user_id}: {contact_visibility}"
        )

        if contact_visibility == ProfileVisibilityEnum.PRIVATE:
            _log.debug(
                f"Contact info sharing denied: user {user.user_id} has "
                f"PRIVATE contact info visibility"
            )
            return False

        if contact_visibility == ProfileVisibilityEnum.PUBLIC:
            _log.debug(
                f"Contact info sharing allowed: user {user.user_id} has "
                f"PUBLIC contact info visibility"
            )
            return True

        # For FRIENDS_ONLY, check if user follows requester
        if contact_visibility == ProfileVisibilityEnum.FRIENDS_ONLY:
            is_friend = await self._is_friend(user, requester_user_id)
            _log.debug(
                f"Contact info sharing for FRIENDS_ONLY: user {user.user_id} -> "
                f"requester {requester_user_id} = {is_friend}"
            )
            return is_friend

        _log.debug(
            f"Contact info sharing denied: user {user.user_id} -> "
            f"requester {requester_user_id} (no matching conditions)"
        )
        return False
