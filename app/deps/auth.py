"""Authentication and authorization dependency providers."""

from typing import Annotated, Any, NoReturn

from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from app.api.v1.schemas.downstream.auth import UserContext
from app.core.config import settings
from app.db.sql.models.user.user import User
from app.deps.database import DatabaseSession, RedisDbSession
from app.enums.user_role import UserRole
from app.middleware.auth_middleware import get_current_user_id_and_scopes
from app.services.user_service import UserService

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
    # Use the OAuth2-aware middleware to get user ID and scopes
    authorization = f"Bearer {credentials.credentials}"
    user_id, scopes, client_id = await get_current_user_id_and_scopes(authorization)

    # Get user from database
    user = await user_service.get_user_by_id(user_id)
    if not user:
        _raise_user_not_found_error()

    if not getattr(user, "is_active", False):
        _raise_inactive_user_error()

    # Store OAuth2 context as runtime attributes
    # Note: Using setattr to avoid mypy attr-defined errors
    user.oauth2_scopes = scopes  # type: ignore[attr-defined]
    user.oauth2_client_id = client_id  # type: ignore[attr-defined]

    return user


async def get_current_active_user(
    current_user: Annotated[User, Depends(get_current_user)],
) -> User:
    """Get current active user."""
    if not getattr(current_user, "is_active", False):
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


# OAuth2 Context Management
def get_user_context(
    current_user: Annotated[User, Depends(get_current_user)],
) -> UserContext:
    """Get user context with OAuth2 scopes.

    Args:
        current_user: Current authenticated user

    Returns:
        UserContext: User context with OAuth2 information
    """
    scopes = getattr(current_user, "oauth2_scopes", [])
    client_id = getattr(current_user, "oauth2_client_id", None)

    return UserContext(
        user_id=str(current_user.user_id),
        scopes=scopes,
        client_id=client_id,
        token_type="Bearer",
        authenticated_at=None,
    )


# Scope-based Authorization
def require_scope(required_scope: str) -> Any:
    """Dependency factory to require specific OAuth2 scope.

    Args:
        required_scope: The scope that must be present in the token

    Returns:
        Dependency function that validates the required scope
    """

    async def scope_dependency(
        user_context: Annotated[UserContext, Depends(get_user_context)],
    ) -> UserContext:
        if not settings.oauth2_service_enabled:
            # If OAuth2 is disabled, fall back to role-based authorization
            return user_context

        if not user_context.has_scope(required_scope):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Missing required scope: {required_scope}",
            )
        return user_context

    return Depends(scope_dependency)


def require_any_scope(required_scopes: list[str]) -> Any:
    """Dependency factory to require any of the specified OAuth2 scopes.

    Args:
        required_scopes: List of scopes, user must have at least one

    Returns:
        Dependency function that validates the required scopes
    """

    async def scope_dependency(
        user_context: Annotated[UserContext, Depends(get_user_context)],
    ) -> UserContext:
        if not settings.oauth2_service_enabled:
            # If OAuth2 is disabled, fall back to role-based authorization
            return user_context

        if not user_context.has_any_scope(required_scopes):
            scopes_str = ", ".join(required_scopes)
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Missing required scopes (need one of): {scopes_str}",
            )
        return user_context

    return Depends(scope_dependency)


def require_all_scopes(required_scopes: list[str]) -> Any:
    """Dependency factory to require all of the specified OAuth2 scopes.

    Args:
        required_scopes: List of scopes, user must have all of them

    Returns:
        Dependency function that validates the required scopes
    """

    async def scope_dependency(
        user_context: Annotated[UserContext, Depends(get_user_context)],
    ) -> UserContext:
        if not settings.oauth2_service_enabled:
            # If OAuth2 is disabled, fall back to role-based authorization
            return user_context

        if not user_context.has_all_scopes(required_scopes):
            scopes_str = ", ".join(required_scopes)
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Missing required scopes (need all of): {scopes_str}",
            )
        return user_context

    return Depends(scope_dependency)


def require_admin_scope() -> Any:
    """Dependency to require admin scope or admin role.

    Returns:
        Dependency function that validates admin access
    """

    async def admin_dependency(
        current_user: Annotated[User, Depends(get_current_user)],
        user_context: Annotated[UserContext, Depends(get_user_context)],
    ) -> UserContext:
        if settings.oauth2_service_enabled:
            # Check OAuth2 admin scopes
            admin_scopes = settings.oauth2_admin_scopes
            if not user_context.has_any_scope(admin_scopes):
                scopes_str = ", ".join(admin_scopes)
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail=f"Missing admin scopes (need one of): {scopes_str}",
                )
        # Fall back to role-based authorization
        elif current_user.role != UserRole.ADMIN:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN, detail="Admin role required"
            )

        return user_context

    return Depends(admin_dependency)


def require_user_management_scope() -> Any:
    """Dependency to require user management scopes.

    Returns:
        Dependency function that validates user management access
    """
    user_mgmt_scopes = settings.oauth2_user_management_scopes
    return require_any_scope(user_mgmt_scopes)


# Type aliases for dependency injection
CurrentUser = Annotated[User, Depends(get_current_user)]
CurrentActiveUser = Annotated[User, Depends(get_current_active_user)]
CurrentAdminUser = Annotated[User, Depends(get_current_admin_user)]

# OAuth2 Authorization Dependencies
RequireReadScope = require_scope("user:read")
RequireWriteScope = require_scope("user:write")
RequireAdminScope = require_admin_scope()
RequireUserManagementScope = require_user_management_scope()

# Annotated types for OAuth2 context
UserContextDep = Annotated[UserContext, Depends(get_user_context)]
ReadScopeDep = Annotated[UserContext, RequireReadScope]
WriteScopeDep = Annotated[UserContext, RequireWriteScope]
AdminScopeDep = Annotated[UserContext, RequireAdminScope]
UserManagementScopeDep = Annotated[UserContext, RequireUserManagementScope]
