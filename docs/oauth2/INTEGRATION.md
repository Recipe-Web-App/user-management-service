# OAuth2 Integration Guide

This guide provides comprehensive instructions for integrating the User Management Service with an external OAuth2
authentication service.

## Overview

The User Management Service supports **dual authentication modes** to provide flexibility during migration and
different use cases:

1. **OAuth2 Integration** (Recommended) - External OAuth2 service authentication
2. **Legacy JWT Authentication** - Local JWT token generation (backward compatibility)

## OAuth2 Architecture

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client App    │    │  OAuth2 Service │    │  User Mgmt Svc  │
│  (Web/Mobile)   │    │   (Auth Svc)    │    │   (This Svc)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
    1. Authorization              │                       │
         Request                  │                       │
         │────────────────────────▶                      │
         │                       │                       │
    2. Authorization              │                       │
         Code                     │                       │
         ◀────────────────────────│                      │
         │                       │                       │
    3. Token Exchange             │                       │
         │────────────────────────▶                      │
         │                       │                       │
    4. Access Token               │                       │
         ◀────────────────────────│                      │
         │                       │                       │
    5. API Request                │                       │
         with Token              │                       │
         │──────────────────────────────────────────────▶│
         │                       │                       │
    6. Token Validation           │                       │
         │                       ◀───────────────────────│
         │                       │                       │
    7. User Data                  │                       │
         ◀──────────────────────────────────────────────│
```

## Quick Setup

### 1. Environment Configuration

**Required environment variables:**

```bash
# OAuth2 Integration
OAUTH2_SERVICE_ENABLED=true
OAUTH2_SERVICE_TO_SERVICE_ENABLED=true
OAUTH2_INTROSPECTION_ENABLED=false
OAUTH2_CLIENT_ID=recipe-service-client
OAUTH2_CLIENT_SECRET=your-oauth2-client-secret

# JWT Configuration (shared with OAuth2 service)
JWT_SECRET=shared-secret-with-oauth2-service
```

### 2. OAuth2 Service Configuration

**Create `config/oauth2.json`:**

```json
{
  "oauth2_service_urls": {
    "authorization_url": "https://auth-service.example.com/authorize",
    "token_url": "https://auth-service.example.com/token",
    "introspection_url": "https://auth-service.example.com/introspect",
    "userinfo_url": "https://auth-service.example.com/userinfo"
  },
  "scopes": {
    "default_scopes": ["openid", "profile"],
    "admin_scopes": ["openid", "profile", "admin"],
    "user_management_scopes": ["openid", "profile", "user:read", "user:write"]
  },
  "cache": {
    "token_cache_ttl_seconds": 300
  },
  "introspection": {
    "timeout_seconds": 5
  }
}
```

### 3. Verify Configuration

```bash
# Check OAuth2 configuration loads
poetry run python -c "from app.core.config import settings; print('OAuth2 enabled:', settings.oauth2_service_enabled)"

# Test OAuth2 client initialization
poetry run python -c "from app.core.oauth2_client import oauth2_client; print('OAuth2 client ready')"
```

## Token Validation Modes

The service supports two token validation approaches:

### Mode 1: JWT Shared Secret (Default)

**Configuration:**

```bash
OAUTH2_INTROSPECTION_ENABLED=false
JWT_SECRET=shared-secret-with-oauth2-service
```

**How it works:**

- OAuth2 service issues JWT tokens signed with shared secret
- User Management Service validates JWT signature locally
- **Pros**: Fast, offline validation
- **Cons**: No real-time token revocation

**Use case**: High-performance scenarios where token revocation is not critical

### Mode 2: Token Introspection

**Configuration:**

```bash
OAUTH2_INTROSPECTION_ENABLED=true
OAUTH2_CLIENT_ID=service-client-id
OAUTH2_CLIENT_SECRET=service-client-secret
```

**How it works:**

- User Management Service calls OAuth2 introspection endpoint for validation
- OAuth2 service responds with token status and user information
- **Pros**: Real-time validation, supports token revocation
- **Cons**: Network call overhead, dependency on OAuth2 service availability

**Use case**: Security-critical scenarios requiring real-time token status

## OAuth2 Flows

### Authorization Code Flow with PKCE

**For web and mobile applications:**

1. **Client initiates authorization:**

```http
GET https://auth-service.example.com/authorize
  ?response_type=code
  &client_id=recipe-web-client
  &redirect_uri=https://app.example.com/callback
  &scope=openid profile user:read user:write
  &state=random_state_value
  &code_challenge=code_challenge_value
  &code_challenge_method=S256
```

1. **OAuth2 service redirects with authorization code:**

```http
HTTP/1.1 302 Found
Location: https://app.example.com/callback
  ?code=authorization_code_here
  &state=random_state_value
```

1. **Client exchanges code for tokens:**

```http
POST https://auth-service.example.com/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=authorization_code_here
&redirect_uri=https://app.example.com/callback
&client_id=recipe-web-client
&code_verifier=code_verifier_value
```

1. **OAuth2 service returns tokens:**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "openid profile user:read user:write"
}
```

1. **Client uses access token:**

```http
GET /api/v1/users/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Client Credentials Flow

**For service-to-service authentication:**

```http
POST https://auth-service.example.com/token
Content-Type: application/x-www-form-urlencoded
Authorization: Basic <base64(client_id:client_secret)>

grant_type=client_credentials
&scope=user:read user:write
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "user:read user:write"
}
```

## Scopes and Permissions

### Available Scopes

| Scope        | Description               | Required For                |
| ------------ | ------------------------- | --------------------------- |
| `openid`     | Basic user identification | All authenticated endpoints |
| `profile`    | User profile information  | Profile-related endpoints   |
| `user:read`  | Read user data            | GET /users/\*, /search      |
| `user:write` | Create/update user data   | POST/PUT /users/\*          |
| `admin`      | Administrative operations | /admin/\* endpoints         |

### Scope Enforcement

**Automatic scope validation:**

```python
# Route automatically requires 'user:write' scope
@router.put("/users/profile")
async def update_profile(
    user_context: WriteScopeDep,  # Requires user:write scope
    profile_data: UserProfileUpdate
):
    # Implementation here
    pass
```

**Custom scope requirements:**

```python
from app.deps.auth import require_scope

@router.get("/custom-endpoint")
async def custom_endpoint(
    user_context: Annotated[UserContext, require_scope("custom:scope")]
):
    # Implementation here
    pass
```

**Multiple scope options:**

```python
from app.deps.auth import require_any_scope

@router.get("/flexible-endpoint")
async def flexible_endpoint(
    user_context: Annotated[UserContext, require_any_scope(["user:read", "admin"])]
):
    # User needs either user:read OR admin scope
    pass
```

## Migration Strategies

### Phase 1: Parallel Authentication

**Configuration:**

```bash
OAUTH2_SERVICE_ENABLED=true
OAUTH2_INTROSPECTION_ENABLED=false
# Legacy JWT settings remain active
```

**Benefits:**

- Both OAuth2 and legacy JWT tokens work
- Gradual client migration possible
- Zero downtime deployment

### Phase 2: OAuth2 Primary

**Configuration:**

```bash
OAUTH2_SERVICE_ENABLED=true
OAUTH2_INTROSPECTION_ENABLED=true  # Optional: Enable real-time validation
# Legacy endpoints still available for backward compatibility
```

**Benefits:**

- OAuth2 becomes primary authentication method
- Real-time token validation (if introspection enabled)
- Legacy clients still supported

### Phase 3: OAuth2 Only (Future)

**Configuration:**

- Remove legacy JWT endpoints entirely
- Update client applications to use OAuth2 flow exclusively

## Error Handling

### OAuth2 Scope Errors

**403 Forbidden - Insufficient Scope:**

```json
{
  "detail": "Missing required scope: user:write",
  "error_code": "INSUFFICIENT_SCOPE",
  "required_scopes": ["user:write"],
  "available_scopes": ["user:read", "profile"]
}
```

**401 Unauthorized - Invalid Token:**

```json
{
  "detail": "OAuth2 token validation failed",
  "error_code": "INVALID_OAUTH2_TOKEN",
  "token_type": "oauth2"
}
```

### Token Validation Failures

**JWT Validation Mode:**

- Invalid signature → 401 Unauthorized
- Expired token → 401 Unauthorized
- Missing required claims → 401 Unauthorized

**Introspection Mode:**

- OAuth2 service unreachable → 503 Service Unavailable
- Token revoked → 401 Unauthorized
- Invalid token format → 401 Unauthorized

## Performance Considerations

### JWT Validation Mode

- **Pros**:
  - No network calls
  - Sub-millisecond validation
  - OAuth2 service downtime doesn't affect validation
- **Cons**:
  - No real-time revocation
  - Token remains valid until expiration

### Introspection Mode

- **Pros**:
  - Real-time token status
  - Immediate revocation support
  - Centralized token management
- **Cons**:
  - Network latency (5-50ms typically)
  - Dependency on OAuth2 service availability
  - Higher resource usage

### Service-to-Service Token Caching

The OAuth2 client automatically caches service-to-service tokens:

```python
# Tokens cached for 5 minutes by default
oauth2_client = OAuth2Client()
token = await oauth2_client.get_service_token(['user:read', 'user:write'])

# Clear cache if needed
oauth2_client.clear_cache()
```

## Security Best Practices

### Client Security

1. **Use HTTPS only** in production
2. **Implement PKCE** for authorization code flow
3. **Store tokens securely** (not in localStorage for web apps)
4. **Implement token refresh** logic
5. **Handle token expiration** gracefully

### Service Security

1. **Rotate client secrets** regularly
2. **Use environment variables** for secrets
3. **Monitor token usage** and failures
4. **Implement rate limiting** on token endpoints
5. **Use minimal scopes** principle

### OAuth2 Service Requirements

1. **Support PKCE** for authorization code flow
2. **Provide token introspection** endpoint (RFC 7662)
3. **Issue JWT tokens** with required claims
4. **Support client credentials** flow
5. **Implement proper CORS** policies

## Troubleshooting

### Common Issues

**OAuth2 service connectivity:**

```bash
# Test OAuth2 service health
curl -v https://auth-service.example.com/health

# Check configuration
cat config/oauth2.json

# Verify environment variables
env | grep OAUTH2
```

**Token validation failures:**

```bash
# Test JWT validation mode
export DEBUG_TOKEN="your_jwt_token"
poetry run python -c "
from app.middleware.auth_middleware import _validate_token
import asyncio
import os
token = os.getenv('DEBUG_TOKEN')
if token:
    result = asyncio.run(_validate_token(token))
    print('Validation result:', result)
"
```

**Scope authorization issues:**

```bash
# Check available scopes in token
# Decode JWT token at https://jwt.io
# Or use introspection endpoint:
curl -X POST https://auth-service.example.com/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "client_id:client_secret" \
  -d "token=your_access_token"
```

### Health Monitoring

**OAuth2 integration health:**

```bash
# Service health includes OAuth2 status
curl http://localhost:8000/api/v1/user-management/health

# Test OAuth2 client functionality
poetry run python -c "
from app.core.oauth2_client import oauth2_client
print('OAuth2 client status:', 'ready' if oauth2_client else 'not configured')
"
```

## Development Testing

### Mock OAuth2 Service

For development, you can use a mock OAuth2 service or configure the service to use JWT validation mode with a
development OAuth2 service.

**Development configuration:**

```bash
# Use JWT validation for faster development
OAUTH2_INTROSPECTION_ENABLED=false
JWT_SECRET=development-shared-secret

# Point to development OAuth2 service
# Update config/oauth2.json accordingly
```

### Testing Scopes

**Create test tokens with different scopes:**

```python
# Example test helper (not production code)
import jwt

def create_test_token(scopes):
    payload = {
        "sub": "test-user-123",
        "scopes": scopes,
        "exp": int(time.time()) + 3600
    }
    return jwt.encode(payload, "development-shared-secret", algorithm="HS256")

# Test tokens
read_token = create_test_token(["openid", "profile", "user:read"])
write_token = create_test_token(["openid", "profile", "user:read", "user:write"])
admin_token = create_test_token(["openid", "profile", "admin"])
```

This guide should provide everything needed to integrate the User Management Service with your OAuth2 authentication
service. For additional support, refer to the API documentation and development guide.
