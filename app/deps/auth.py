"""Authentication dependency providers."""

from typing import Annotated, NoReturn

from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from jose import JWTError

from app.db.sql.models.user.user import User
from app.deps.database import DatabaseSession, RedisDbSession
from app.enums.user_role import UserRole
from app.services.user_service import UserService
from app.utils.security import decode_jwt_token

security = HTTPBearer()


def _raise_invalid_token_error() -> NoReturn:
    """Raise HTTPException for invalid authentication token."""
    raise HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Invalid authentication token",
        headers={"WWW-Authenticate": "Bearer"},
    )


def _raise_user_not_found_error() -> NoReturn:
    """Raise HTTPException for user not found."""
    raise HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="User not found",
        headers={"WWW-Authenticate": "Bearer"},
    )


def _raise_inactive_user_error() -> NoReturn:
    """Raise HTTPException for inactive user account."""
    raise HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="User account is inactive",
        headers={"WWW-Authenticate": "Bearer"},
    )


async def get_user_service_for_auth(
    db_session: DatabaseSession,
    redis_session: RedisDbSession,
) -> UserService:
    """Get user service dependency for authentication."""
    return UserService(db_session, redis_session)


async def get_current_user(
    credentials: Annotated[HTTPAuthorizationCredentials, Depends(security)],
    user_service: Annotated[UserService, Depends(get_user_service_for_auth)],
) -> User:
    """Get current authenticated user."""
    # Decode and validate the JWT token
    token = credentials.credentials

    try:
        payload = decode_jwt_token(token)
    except JWTError:
        _raise_invalid_token_error()

    user_id = payload.get("sub")
    if not user_id or not isinstance(user_id, str):
        _raise_invalid_token_error()

    # Get user from database
    user = await user_service.get_user_by_id(user_id)
    if not user:
        _raise_user_not_found_error()

    if not user.is_active:
        _raise_inactive_user_error()

    return user


async def get_current_active_user(
    current_user: Annotated[User, Depends(get_current_user)],
) -> User:
    """Get current active user."""
    if not current_user.is_active:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST, detail="Inactive user"
        )
    return current_user


async def get_current_admin_user(
    current_user: Annotated[User, Depends(get_current_active_user)],
) -> User:
    """Get current admin user."""
    if current_user.role != UserRole.ADMIN:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Not enough permissions"
        )
    return current_user


# Type aliases for dependency injection
CurrentUser = Annotated[User, Depends(get_current_user)]
CurrentActiveUser = Annotated[User, Depends(get_current_active_user)]
CurrentAdminUser = Annotated[User, Depends(get_current_admin_user)]
