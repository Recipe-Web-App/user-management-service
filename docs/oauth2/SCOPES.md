# Scope-Based Authorization Guide

This guide explains how to implement and use scope-based authorization in the User Management Service with OAuth2 integration.

## Overview

**Scope-based authorization** provides **granular permission control** by associating specific scopes (permissions) with
OAuth2 access tokens. This is more flexible than traditional role-based access control and aligns with OAuth2 best practices.

## Available Scopes

### Core Scopes

| Scope     | Purpose          | Description                          | Example Use Cases           |
| --------- | ---------------- | ------------------------------------ | --------------------------- |
| `openid`  | **Identity**     | Basic user identification (required) | All authenticated endpoints |
| `profile` | **Profile Data** | Access to user profile information   | Profile views, user cards   |

### User Management Scopes

| Scope        | Purpose          | Description                      | Example Use Cases                  |
| ------------ | ---------------- | -------------------------------- | ---------------------------------- |
| `user:read`  | **Read Users**   | Read user profiles, search users | User directory, profile views      |
| `user:write` | **Modify Users** | Create/update user profiles      | Profile editing, user registration |

### Administrative Scopes

| Scope   | Purpose            | Description                | Example Use Cases                      |
| ------- | ------------------ | -------------------------- | -------------------------------------- |
| `admin` | **Administration** | Full administrative access | User management, system administration |

## Scope Enforcement Architecture

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  OAuth2 Token   │    │  Scope Check    │    │   Endpoint      │
│                 │    │                 │    │                 │
│  scopes:        │───▶│  Required:      │───▶│  /users/profile │
│  - openid       │    │  - user:write   │    │                 │
│  - profile      │    │                 │    │  ❌ 403         │
│  - user:read    │    │  Available:     │    │  Forbidden      │
│                 │    │  - user:read    │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Error Resp    │
                       │                 │
                       │ "Missing scope: │
                       │  user:write"    │
                       │                 │
                       │ Required:       │
                       │ ["user:write"]  │
                       │                 │
                       │ Available:      │
                       │ ["user:read"]   │
                       └─────────────────┘
```

## Implementation Patterns

### 1. Pre-configured Dependencies

**Ready-to-use scope dependencies:**

```python
from app.deps.auth import (
    RequireReadScope,      # Requires user:read
    RequireWriteScope,     # Requires user:write
    RequireAdminScope,     # Requires admin (or falls back to ADMIN role)
    UserContextDep,        # Basic user context (no scope requirements)
)

# Read-only endpoint
@router.get("/users/search")
async def search_users(
    user_context: ReadScopeDep,  # Requires user:read scope
    query: str
):
    # Implementation
    pass

# Write endpoint
@router.put("/users/profile")
async def update_profile(
    user_context: WriteScopeDep,  # Requires user:write scope
    profile_data: UserProfileUpdate
):
    # Implementation
    pass

# Admin endpoint
@router.get("/admin/users/stats")
async def get_user_stats(
    user_context: AdminScopeDep,  # Requires admin scope
):
    # Implementation
    pass
```

### 2. Custom Scope Requirements

**Single scope requirement:**

```python
from app.deps.auth import require_scope

@router.post("/users/custom-action")
async def custom_action(
    user_context: Annotated[UserContext, require_scope("custom:action")]
):
    # Requires exactly "custom:action" scope
    pass
```

**Any of multiple scopes:**

```python
from app.deps.auth import require_any_scope

@router.get("/flexible-endpoint")
async def flexible_endpoint(
    user_context: Annotated[UserContext, require_any_scope(["user:read", "admin"])]
):
    # User needs EITHER user:read OR admin scope
    pass
```

**All of multiple scopes:**

```python
from app.deps.auth import require_all_scopes

@router.post("/sensitive-operation")
async def sensitive_operation(
    user_context: Annotated[UserContext, require_all_scopes(["user:write", "admin"])]
):
    # User needs BOTH user:write AND admin scopes
    pass
```

### 3. Dynamic Scope Checking

**Manual scope verification in endpoint:**

```python
@router.get("/conditional-endpoint")
async def conditional_endpoint(
    user_context: UserContextDep,
    include_sensitive: bool = False
):
    # Basic data always available
    data = get_basic_data()

    # Conditional sensitive data
    if include_sensitive:
        if not user_context.has_scope("admin"):
            raise HTTPException(
                status_code=403,
                detail="Admin scope required for sensitive data"
            )
        data.update(get_sensitive_data())

    return data
```

**Multiple scope options:**

```python
@router.get("/tiered-endpoint")
async def tiered_endpoint(
    user_context: UserContextDep
):
    if user_context.has_scope("admin"):
        # Full admin view
        return get_admin_data()
    elif user_context.has_scope("user:write"):
        # Write user view
        return get_write_user_data()
    elif user_context.has_scope("user:read"):
        # Read-only user view
        return get_read_only_data()
    else:
        # Basic view
        return get_basic_data()
```

## UserContext API

### Properties

```python
class UserContext:
    user_id: str                    # User identifier
    scopes: list[str]               # Available scopes
    client_id: str | None           # OAuth2 client ID
    token_type: str                 # Token type (default: "Bearer")
    authenticated_at: datetime | None  # Authentication timestamp
```

### Methods

```python
# Check single scope
user_context.has_scope("user:write") -> bool

# Check any of multiple scopes
user_context.has_any_scope(["user:read", "admin"]) -> bool

# Check all of multiple scopes
user_context.has_all_scopes(["user:write", "profile"]) -> bool
```

### Usage Examples

```python
@router.get("/example")
async def example_endpoint(user_context: UserContextDep):
    # Get user ID
    user_id = user_context.user_id

    # Check individual scopes
    can_read = user_context.has_scope("user:read")
    can_write = user_context.has_scope("user:write")
    is_admin = user_context.has_scope("admin")

    # Check multiple scopes
    has_user_access = user_context.has_any_scope(["user:read", "user:write"])
    has_full_access = user_context.has_all_scopes(["user:read", "user:write", "admin"])

    return {
        "user_id": user_id,
        "permissions": {
            "can_read": can_read,
            "can_write": can_write,
            "is_admin": is_admin,
            "has_user_access": has_user_access,
            "has_full_access": has_full_access
        }
    }
```

## Error Handling

### Scope Error Response

When a user lacks required scopes, the service returns a detailed error:

```json
{
  "detail": "Missing required scope: user:write",
  "error_code": "INSUFFICIENT_SCOPE",
  "required_scopes": ["user:write"],
  "available_scopes": ["user:read", "profile", "openid"]
}
```

### Client Error Handling

**Frontend JavaScript example:**

```javascript
async function callUserAPI(endpoint, options = {}) {
  try {
    const response = await fetch(`/api/v1${endpoint}`, {
      ...options,
      headers: {
        Authorization: `Bearer ${accessToken}`,
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json();

      if (error.error_code === "INSUFFICIENT_SCOPE") {
        // Handle scope error
        console.error("Missing scopes:", error.required_scopes);

        // Redirect to get additional scopes
        requestAdditionalScopes(error.required_scopes);
        return;
      }

      throw new Error(error.detail);
    }

    return await response.json();
  } catch (error) {
    console.error("API call failed:", error);
    throw error;
  }
}

function requestAdditionalScopes(requiredScopes) {
  const currentScopes = "openid profile user:read"; // Current scopes
  const allScopes = currentScopes + " " + requiredScopes.join(" ");

  // Redirect to OAuth2 authorize endpoint with additional scopes
  window.location.href =
    `https://auth-service.example.com/authorize?` +
    `response_type=code&client_id=web-client&` +
    `scope=${encodeURIComponent(allScopes)}&` +
    `redirect_uri=${encodeURIComponent(window.location.href)}`;
}
```

## Backward Compatibility

### Legacy JWT Mode

When OAuth2 is disabled (`OAUTH2_SERVICE_ENABLED=false`), scope-based authorization falls back to role-based authorization:

```python
# OAuth2 enabled: Checks for "admin" scope
# OAuth2 disabled: Checks for UserRole.ADMIN

@router.get("/admin/stats")
async def admin_stats(user_context: AdminScopeDep):
    # Works in both modes
    pass
```

**Fallback mapping:**

- `admin` scope → `ADMIN` role required
- `user:*` scopes → `USER` role required (any authenticated user)

### Migration Strategy

**Phase 1 - Enable OAuth2 with fallback:**

```python
# Both OAuth2 scopes and legacy roles work
@router.get("/endpoint")
async def endpoint(user_context: AdminScopeDep):
    # OAuth2: Requires admin scope
    # Legacy: Requires ADMIN role
    pass
```

**Phase 2 - OAuth2 primary:**

```python
# Most clients use OAuth2, some still use legacy JWT
# Same code works for both
```

**Phase 3 - OAuth2 only:**

```python
# Remove legacy JWT endpoints
# All authorization uses scopes
```

## Scope Design Patterns

### 1. Resource-Based Scopes

**Pattern**: `resource:action`

```text
user:read       # Read user data
user:write      # Write user data
recipe:read     # Read recipes
recipe:write    # Write recipes
```

### 2. Hierarchical Scopes

**Pattern**: Broader scopes include narrower ones

```text
admin           # Includes all permissions
user:write      # Includes user:read
user:read       # Basic read permission
```

**Implementation:**

```python
def has_permission(user_scopes: list[str], required: str) -> bool:
    if "admin" in user_scopes:
        return True  # Admin can do everything

    if required == "user:read" and "user:write" in user_scopes:
        return True  # Write implies read

    return required in user_scopes
```

### 3. Contextual Scopes

**Pattern**: Scopes that depend on context

```python
@router.get("/users/{user_id}/profile")
async def get_profile(
    user_id: str,
    user_context: Annotated[UserContext, require_scope("user:read")]
):
    # Additional check: users can always read their own profile
    if user_context.user_id == user_id:
        return get_user_profile(user_id)

    # Others need explicit read permission
    if not user_context.has_scope("user:read"):
        raise HTTPException(403, "Cannot read other user's profile")

    # Apply privacy filters for other users
    return get_filtered_profile(user_id, user_context.user_id)
```

## Testing Scope-Based Authorization

### Unit Tests

```python
import pytest
from app.api.v1.schemas.downstream.auth import UserContext

def test_user_context_scopes():
    # Test scope checking
    context = UserContext(
        user_id="test-user",
        scopes=["openid", "profile", "user:read"],
        client_id="test-client",
        token_type="Bearer",
        authenticated_at=None
    )

    assert context.has_scope("user:read")
    assert not context.has_scope("user:write")
    assert not context.has_scope("admin")

    assert context.has_any_scope(["user:read", "admin"])
    assert not context.has_any_scope(["user:write", "admin"])

    assert context.has_all_scopes(["openid", "profile"])
    assert not context.has_all_scopes(["openid", "user:write"])
```

### Integration Tests

```python
import pytest
from fastapi.testclient import TestClient

def test_scope_enforcement(client: TestClient, oauth2_token):
    # Token with user:read scope only
    headers = {"Authorization": f"Bearer {oauth2_token(['user:read'])}"}

    # Should succeed - requires user:read
    response = client.get("/api/v1/users/search", headers=headers)
    assert response.status_code == 200

    # Should fail - requires user:write
    response = client.put("/api/v1/users/profile", headers=headers, json={})
    assert response.status_code == 403

    error = response.json()
    assert error["error_code"] == "INSUFFICIENT_SCOPE"
    assert "user:write" in error["required_scopes"]
```

### Mock OAuth2 Tokens

```python
import jwt
from datetime import datetime, timedelta

def create_test_token(scopes: list[str], user_id: str = "test-user") -> str:
    """Create a test OAuth2 JWT token with specified scopes."""
    payload = {
        "sub": user_id,
        "scopes": scopes,
        "client_id": "test-client",
        "exp": datetime.utcnow() + timedelta(hours=1),
        "iat": datetime.utcnow(),
        "iss": "test-oauth2-service"
    }
    return jwt.encode(payload, "test-secret", algorithm="HS256")

# Usage in tests
read_token = create_test_token(["openid", "profile", "user:read"])
write_token = create_test_token(["openid", "profile", "user:read", "user:write"])
admin_token = create_test_token(["openid", "profile", "admin"])
```

## Best Practices

### 1. Principle of Least Privilege

**Request only necessary scopes:**

```javascript
// Good: Request specific scopes
const scopes = "openid profile user:read";

// Bad: Request excessive scopes
const scopes = "openid profile user:read user:write admin";
```

### 2. Graceful Degradation

**Provide different functionality based on available scopes:**

```python
@router.get("/users/dashboard")
async def user_dashboard(user_context: UserContextDep):
    dashboard = {"user_id": user_context.user_id}

    # Add features based on available scopes
    if user_context.has_scope("user:read"):
        dashboard["profile"] = get_profile_data(user_context.user_id)

    if user_context.has_scope("user:write"):
        dashboard["edit_profile_url"] = "/profile/edit"

    if user_context.has_scope("admin"):
        dashboard["admin_panel_url"] = "/admin"

    return dashboard
```

### 3. Clear Error Messages

**Provide helpful error information:**

```python
def require_custom_scope(required_scope: str):
    async def scope_dependency(user_context: UserContextDep):
        if not user_context.has_scope(required_scope):
            raise HTTPException(
                status_code=403,
                detail={
                    "message": f"This action requires the '{required_scope}' scope",
                    "error_code": "INSUFFICIENT_SCOPE",
                    "required_scopes": [required_scope],
                    "available_scopes": user_context.scopes,
                    "help_url": "https://docs.example.com/oauth2/scopes"
                }
            )
        return user_context

    return Depends(scope_dependency)
```

### 4. Scope Documentation

**Document scope requirements in API documentation:**

```python
@router.put("/users/profile")
async def update_profile(
    user_context: WriteScopeDep,
    profile_data: UserProfileUpdate
):
    """
    Update user profile.

    **Required OAuth2 Scopes**: `user:write`

    **Fallback Authorization**: USER role (when OAuth2 disabled)
    """
    pass
```

## Advanced Usage

### Custom Scope Logic

```python
from app.deps.auth import get_user_context

def require_profile_access(target_user_id: str = None):
    """Custom dependency that allows access to own profile or requires admin scope."""

    async def profile_access_dependency(
        user_context: Annotated[UserContext, Depends(get_user_context)]
    ):
        # Users can always access their own profile
        if target_user_id and user_context.user_id == target_user_id:
            return user_context

        # Others need admin scope
        if not user_context.has_scope("admin"):
            raise HTTPException(
                status_code=403,
                detail="Admin scope required to access other user's profile"
            )

        return user_context

    return Depends(profile_access_dependency)

@router.get("/users/{user_id}/private-profile")
async def get_private_profile(
    user_id: str,
    user_context: Annotated[UserContext, require_profile_access(user_id)]
):
    # Implementation
    pass
```

This guide provides comprehensive coverage of scope-based authorization patterns and best practices for the User Management
Service. Use these patterns to implement fine-grained access control in your OAuth2-integrated application.
