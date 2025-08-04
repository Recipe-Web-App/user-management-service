"""Social service for user social interactions."""

from uuid import UUID

from fastapi import HTTPException, status
from sqlalchemy import desc, func, select
from sqlalchemy.exc import DisconnectionError
from sqlalchemy.exc import TimeoutError as SQLTimeoutError

from app.api.v1.schemas.common.user import User as UserSchema
from app.api.v1.schemas.response.social import FollowResponse, GetFollowedUsersResponse
from app.api.v1.schemas.response.user_activity.favorite_summary import FavoriteSummary
from app.api.v1.schemas.response.user_activity.recipe_summary import RecipeSummary
from app.api.v1.schemas.response.user_activity.review_summary import ReviewSummary
from app.api.v1.schemas.response.user_activity.user_activity_response import (
    UserActivityResponse,
)
from app.api.v1.schemas.response.user_activity.user_summary import UserSummary
from app.core.logging import get_logger
from app.db.sql.models.recipe.recipe import Recipe
from app.db.sql.models.recipe.recipe_favorite import RecipeFavorite
from app.db.sql.models.recipe.recipe_review import RecipeReview
from app.db.sql.models.user.user import User
from app.db.sql.models.user.user_follows import UserFollows
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError
from app.utils.privacy import PrivacyChecker

_log = get_logger(__name__)


class SocialService:
    """Service for social operations."""

    def __init__(self, db: SqlDatabaseSession) -> None:
        """Initialize social service with database session."""
        self.db = db
        self.privacy_checker = PrivacyChecker(db)

    async def get_followed_users(
        self,
        user_id: UUID,
        requester_user_id: UUID,
        limit: int = 20,
        offset: int = 0,
        count_only: bool = False,
    ) -> GetFollowedUsersResponse:
        """Get users that a user is following, respecting privacy preferences."""
        _log.info(f"Getting following list for user: {user_id}")
        try:
            user = await self.db.get_user_by_id(str(user_id))
            if not user:
                _log.warning(f"User not found: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )
            if (
                requester_user_id is not None
                and not await self.privacy_checker.check_access(user, requester_user_id)
            ):
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied due to privacy settings",
                )

            # Get total count (before filtering)
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

            # Per-user privacy filtering
            filtered_users = [
                u
                for u in following_users
                if await self.privacy_checker.check_access(u, requester_user_id)
            ]
            user_schemas = [
                UserSchema.model_validate(u).model_copy(update={"email": None})
                for u in filtered_users
            ]
            _log.info(
                f"Retrieved {len(user_schemas)} following users for user {user_id} "
                f"(after privacy filtering)"
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

    async def get_followers(
        self,
        user_id: UUID,
        requester_user_id: UUID,
        limit: int = 20,
        offset: int = 0,
        count_only: bool = False,
    ) -> GetFollowedUsersResponse:
        """Get users who follow the given user, respecting privacy preferences."""
        _log.info(f"Getting followers list for user: {user_id}")
        try:
            user = await self.db.get_user_by_id(str(user_id))
            if not user:
                _log.warning(f"User not found: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )
            if (
                requester_user_id is not None
                and not await self.privacy_checker.check_access(user, requester_user_id)
            ):
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied due to privacy settings",
                )

            # Get total count (before filtering)
            count_result = await self.db.execute(
                select(func.count(UserFollows.follower_id)).where(
                    UserFollows.followee_id == user_id
                )
            )
            total_count = count_result.scalar()
            _log.info(f"Followers count for user {user_id}: {total_count}")

            if count_only:
                return GetFollowedUsersResponse(
                    total_count=total_count,
                    followed_users=None,
                    limit=None,
                    offset=None,
                )

            # Get follower relationships with user details
            followers_result = await self.db.execute(
                select(User)
                .join(UserFollows, User.user_id == UserFollows.follower_id)
                .where(UserFollows.followee_id == user_id)
                .order_by(UserFollows.followed_at.desc())
                .limit(limit)
                .offset(offset)
            )
            followers = followers_result.scalars().all()

            # Per-user privacy filtering
            filtered_users = [
                u
                for u in followers
                if await self.privacy_checker.check_access(u, requester_user_id)
            ]
            user_schemas = [
                UserSchema.model_validate(u).model_copy(update={"email": None})
                for u in filtered_users
            ]
            _log.info(
                f"Retrieved {len(user_schemas)} followers for user {user_id} "
                f"(after privacy filtering)"
            )

            return GetFollowedUsersResponse(
                total_count=total_count,
                followed_users=user_schemas,
                limit=limit,
                offset=offset,
            )

        except DatabaseError as e:
            _log.error(f"Database error while getting followers list: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error while getting followers list: {e}")
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
            follower_id: The user who wants to follow
            target_user_id: The user to be followed

        Returns:
            FollowResponse: Confirmation of following action

        Raises:
            HTTPException: If users not found, already following, or database error
        """
        _log.info(f"User {follower_id} attempting to follow user {target_user_id}")

        try:
            # Check if both users exist
            follower = await self.db.get_user_by_id(str(follower_id))
            if not follower:
                _log.warning(f"Follower not found: {follower_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Follower not found",
                )

            target_user = await self.db.get_user_by_id(str(target_user_id))
            if not target_user:
                _log.warning(f"Target user not found: {target_user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Target user not found",
                )

            # Check if users are active
            if not follower.is_active:
                _log.warning(f"Inactive follower: {follower_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Follower account is inactive",
                )

            if not target_user.is_active:
                _log.warning(f"Inactive target user: {target_user_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Target user account is inactive",
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
                _log.info(f"User {follower_id} already following {target_user_id}")
                return FollowResponse(
                    message="Already following this user",
                    is_following=True,
                )

            # Create follow relationship
            follow_relationship = UserFollows(
                follower_id=follower_id,
                followee_id=target_user_id,
            )
            self.db.add(follow_relationship)
            await self.db.commit()

            _log.info(f"User {follower_id} successfully followed {target_user_id}")
            return FollowResponse(
                message="Successfully followed user",
                is_following=True,
            )

        except DatabaseError as e:
            _log.error(f"Database error while following user: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
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
            follower_id: The user who wants to unfollow
            target_user_id: The user to be unfollowed

        Returns:
            FollowResponse: Confirmation of unfollowing action

        Raises:
            HTTPException: If users not found, not following, or database error
        """
        _log.info(f"User {follower_id} attempting to unfollow user {target_user_id}")

        try:
            # Check if both users exist
            follower = await self.db.get_user_by_id(str(follower_id))
            if not follower:
                _log.warning(f"Follower not found: {follower_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Follower not found",
                )

            target_user = await self.db.get_user_by_id(str(target_user_id))
            if not target_user:
                _log.warning(f"Target user not found: {target_user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Target user not found",
                )

            # Check if users are active
            if not follower.is_active:
                _log.warning(f"Inactive follower: {follower_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Follower account is inactive",
                )

            if not target_user.is_active:
                _log.warning(f"Inactive target user: {target_user_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Target user account is inactive",
                )

            # Prevent self-unfollowing
            if follower_id == target_user_id:
                _log.warning(f"User {follower_id} attempted to unfollow themselves")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Cannot unfollow yourself",
                )

            # Check if currently following
            existing_follow = await self.db.execute(
                select(UserFollows).where(
                    UserFollows.follower_id == follower_id,
                    UserFollows.followee_id == target_user_id,
                )
            )
            follow_relationship = existing_follow.scalar_one_or_none()
            if not follow_relationship:
                _log.info(f"User {follower_id} not following {target_user_id}")
                return FollowResponse(
                    message="Not following this user",
                    is_following=False,
                )

            # Remove follow relationship
            await self.db.delete(follow_relationship)
            await self.db.commit()

            _log.info(f"User {follower_id} successfully unfollowed {target_user_id}")
            return FollowResponse(
                message="Successfully unfollowed user",
                is_following=False,
            )

        except DatabaseError as e:
            _log.error(f"Database error while unfollowing user: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error while unfollowing user: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e

    async def get_user_activity(
        self,
        user_id: UUID,
        requester_user_id: UUID,
        per_type_limit: int = 15,
    ) -> UserActivityResponse:
        """Get recent user activity (recipes, follows, reviews, favorites).

        Args:
            user_id: The user's unique identifier
            requester_user_id: The user making the request
            per_type_limit: Number of results per activity type

        Returns:
            UserActivityResponse: Aggregated user activity

        Raises:
            HTTPException: If user not found, forbidden, or DB error
        """
        _log.info(f"Getting activity for user: {user_id}")
        try:
            user = await self.db.get_user_by_id(str(user_id))
            if not user:
                _log.warning(f"User not found: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )
            if not await self.privacy_checker.check_access(user, requester_user_id):
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied due to privacy settings",
                )

            # Recipes
            recipes_q = (
                select(Recipe)
                .where(Recipe.user_id == user_id)
                .order_by(desc(Recipe.created_at))
                .limit(per_type_limit)
            )
            recipes_result = await self.db.execute(recipes_q)
            recipes = recipes_result.scalars().all()

            # Follows (recently followed users)
            follows_q = (
                select(UserFollows)
                .where(UserFollows.follower_id == user_id)
                .order_by(desc(UserFollows.followed_at))
                .limit(per_type_limit)
            )
            follows_result = await self.db.execute(follows_q)
            follows = follows_result.scalars().all()

            # Reviews
            reviews_q = (
                select(RecipeReview)
                .where(RecipeReview.user_id == user_id)
                .order_by(desc(RecipeReview.created_at))
                .limit(per_type_limit)
            )
            reviews_result = await self.db.execute(reviews_q)
            reviews = reviews_result.scalars().all()

            # Favorites
            favorites_q = (
                select(RecipeFavorite)
                .where(RecipeFavorite.user_id == user_id)
                .order_by(desc(RecipeFavorite.favorited_at))
                .limit(per_type_limit)
            )
            favorites_result = await self.db.execute(favorites_q)
            favorites = favorites_result.scalars().all()

            # Transform to response schemas
            recipe_summaries = [
                RecipeSummary.model_validate(recipe) for recipe in recipes
            ]
            review_summaries = [
                ReviewSummary.model_validate(review) for review in reviews
            ]
            favorite_summaries = [
                FavoriteSummary.model_validate(favorite) for favorite in favorites
            ]

            # For follows, we need to get the followed user details
            followed_users = []
            for follow in follows:
                followed_user = await self.db.get_user_by_id(str(follow.followee_id))
                if followed_user:
                    followed_users.append(
                        UserSummary.model_validate(follow).model_copy(
                            update={
                                "user_id": followed_user.user_id,
                                "username": followed_user.username,
                            }
                        )
                    )

            return UserActivityResponse(
                user_id=user_id,
                recent_recipes=recipe_summaries,
                recent_follows=followed_users,
                recent_reviews=review_summaries,
                recent_favorites=favorite_summaries,
            )

        except DatabaseError as e:
            _log.error(f"Database error while getting user activity: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error while getting user activity: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e
