"""Permission-based dependency providers."""

from typing import TYPE_CHECKING, Annotated, cast

from fastapi import Depends, HTTPException, status

from app.db.sql.models.user.user import User
from app.deps.auth import CurrentActiveUser
from app.deps.services import UserServiceDep
from app.enums.user_role import UserRole

if TYPE_CHECKING:
    from uuid import UUID


async def require_admin_permission(
    current_user: CurrentActiveUser,
) -> User:
    """Require admin permissions."""
    if current_user.role != UserRole.ADMIN:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Admin permissions required"
        )
    return current_user


async def require_user_or_admin_permission(
    current_user: CurrentActiveUser,
    target_user_id: str,
) -> User:
    """Require user to be accessing their own data or be an admin."""
    if current_user.role == UserRole.ADMIN:
        return current_user

    if str(current_user.user_id) != target_user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Access denied: can only access your own data",
        )

    return current_user


async def require_profile_access_permission(
    current_user: CurrentActiveUser,
    target_user_id: str,
    user_service: UserServiceDep,
) -> tuple[User, User]:
    """Require permission to access a user's profile based on privacy settings."""
    # Get target user
    target_user = await user_service.get_user_by_id(target_user_id)
    if not target_user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )

    # Admin can access all profiles
    if current_user.role == UserRole.ADMIN:
        return current_user, target_user

    # Users can always access their own profile
    if str(current_user.user_id) == target_user_id:
        return current_user, target_user

    # Check privacy settings for other users
    if not await user_service.can_view_profile(
        cast("UUID", current_user.user_id), target_user_id
    ):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Profile access denied due to privacy settings",
        )

    return current_user, target_user


async def require_social_interaction_permission(
    current_user: CurrentActiveUser,
    target_user_id: str,
    user_service: UserServiceDep,
) -> tuple[User, User]:
    """Require permission for social interactions (follow, message, etc.)."""
    # Get target user
    target_user = await user_service.get_user_by_id(target_user_id)
    if not target_user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )

    # Can't interact with self
    if str(current_user.user_id) == target_user_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Cannot perform social interactions with yourself",
        )

    # Check if social interactions are allowed
    if not await user_service.can_social_interact(
        cast("UUID", current_user.user_id), target_user_id
    ):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Social interaction denied due to privacy settings",
        )

    return current_user, target_user


# Type aliases for dependency injection
AdminUser = Annotated[User, Depends(require_admin_permission)]
