"""Social service for user management."""

from uuid import UUID

from fastapi import HTTPException, status
from sqlalchemy import func, select
from sqlalchemy.exc import DisconnectionError, IntegrityError
from sqlalchemy.exc import TimeoutError as SQLTimeoutError

from app.api.v1.schemas.common.user import User as UserSchema
from app.api.v1.schemas.response.social import FollowResponse, GetFollowedUsersResponse
from app.core.logging import get_logger
from app.db.sql.models.user.user import User
from app.db.sql.models.user.user_follows import UserFollows
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError

_log = get_logger(__name__)


class SocialService:
    """Service for social operations."""

    def __init__(self, db: SqlDatabaseSession) -> None:
        """Initialize social service with database session."""
        self.db = db

    async def get_following(
        self, user_id: UUID, limit: int = 20, offset: int = 0, count_only: bool = False
    ) -> GetFollowedUsersResponse:
        """Get users that a user is following."""
        _log.info(f"Getting following list for user: {user_id}")
        try:
            user = await self.db.get_user_by_id(str(user_id))
            if not user:
                _log.warning(f"User not found: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )

            # Get total count
            count_result = await self.db.execute(
                select(func.count(UserFollows.followee_id)).where(
                    UserFollows.follower_id == user_id
                )
            )
            total_count = count_result.scalar()
            _log.info(f"Following count for user {user_id}: {total_count}")

            if count_only:
                return GetFollowedUsersResponse(
                    total_count=total_count,
                    followed_users=None,
                    limit=None,
                    offset=None,
                )

            # Get following relationships with user details
            following_result = await self.db.execute(
                select(User)
                .join(UserFollows, User.user_id == UserFollows.followee_id)
                .where(UserFollows.follower_id == user_id)
                .order_by(UserFollows.followed_at.desc())
                .limit(limit)
                .offset(offset)
            )
            following_users = following_result.scalars().all()
            user_schemas = [
                UserSchema.model_validate(u).model_copy(update={"email": None})
                for u in following_users
            ]
            _log.info(
                f"Retrieved {len(user_schemas)} following users for user {user_id}"
            )

            return GetFollowedUsersResponse(
                total_count=total_count,
                followed_users=user_schemas,
                limit=limit,
                offset=offset,
            )

        except DatabaseError as e:
            _log.error(f"Database error while getting following list: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error while getting following list: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e

    async def follow_user(
        self, follower_id: UUID, target_user_id: UUID
    ) -> FollowResponse:
        """Follow a user.

        Args:
            follower_id: The user's unique identifier who is following
            target_user_id: The user's unique identifier to follow

        Returns:
            FollowResponse: Confirmation of following action

        Raises:
            HTTPException: If users not found, already following, or DB error
        """
        _log.info(f"User {follower_id} attempting to follow user {target_user_id}")

        try:
            # Verify both users exist
            follower = await self.db.get_user_by_id(str(follower_id))
            if not follower:
                _log.warning(f"Follower user not found: {follower_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Follower user not found",
                )

            target_user = await self.db.get_user_by_id(str(target_user_id))
            if not target_user:
                _log.warning(f"Target user not found: {target_user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Target user not found",
                )

            # Prevent self-following
            if follower_id == target_user_id:
                _log.warning(f"User {follower_id} attempted to follow themselves")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Cannot follow yourself",
                )

            # Check if already following
            existing_follow = await self.db.execute(
                select(UserFollows).where(
                    UserFollows.follower_id == follower_id,
                    UserFollows.followee_id == target_user_id,
                )
            )
            if existing_follow.scalar_one_or_none():
                _log.info(f"User {follower_id} already following user {target_user_id}")
                return FollowResponse(
                    message="Already following this user",
                    is_following=True,
                )

            # Create following relationship
            new_follow = UserFollows(
                follower_id=follower_id,
                followee_id=target_user_id,
            )
            self.db.add(new_follow)
            await self.db.commit()
            await self.db.refresh(new_follow)

            _log.info(f"User {follower_id} successfully followed user {target_user_id}")

            return FollowResponse(
                message="Successfully followed user",
                is_following=True,
            )

        except IntegrityError as e:
            await self.db.rollback()
            _log.error(f"Integrity error while following user: {e}")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Unable to follow user. Please try again.",
            ) from e
        except DatabaseError as e:
            await self.db.rollback()
            _log.error(f"Database error while following user: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            await self.db.rollback()
            _log.error(f"Database connection error while following user: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e

    async def unfollow_user(
        self, follower_id: UUID, target_user_id: UUID
    ) -> FollowResponse:
        """Unfollow a user.

        Args:
            follower_id: The user's unique identifier who is unfollowing
            target_user_id: The user's unique identifier to unfollow

        Returns:
            FollowResponse: Confirmation of unfollowing action

        Raises:
            HTTPException: If users not found, not following, or DB error
        """
        _log.info(f"User {follower_id} attempting to unfollow user {target_user_id}")

        try:
            # Verify both users exist
            follower = await self.db.get_user_by_id(str(follower_id))
            if not follower:
                _log.warning(f"Follower user not found: {follower_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Follower user not found",
                )

            target_user = await self.db.get_user_by_id(str(target_user_id))
            if not target_user:
                _log.warning(f"Target user not found: {target_user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Target user not found",
                )

            # Find existing following relationship
            existing_follow = await self.db.execute(
                select(UserFollows).where(
                    UserFollows.follower_id == follower_id,
                    UserFollows.followee_id == target_user_id,
                )
            )
            follow_relationship = existing_follow.scalar_one_or_none()

            if not follow_relationship:
                _log.info(f"User {follower_id} not following user {target_user_id}")
                return FollowResponse(
                    message="Not following this user",
                    is_following=False,
                )

            # Remove following relationship
            await self.db.delete(follow_relationship)
            await self.db.commit()

            _log.info(
                f"User {follower_id} successfully unfollowed user {target_user_id}"
            )

            return FollowResponse(
                message="Successfully unfollowed user",
                is_following=False,
            )

        except DatabaseError as e:
            await self.db.rollback()
            _log.error(f"Database error while unfollowing user: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            await self.db.rollback()
            _log.error(f"Database connection error while unfollowing user: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e
