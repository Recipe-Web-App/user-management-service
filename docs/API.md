# User Management Service API Documentation

## Overview

The User Management Service provides comprehensive user authentication, profile management, social features, and
administrative capabilities for the Recipe Web Application ecosystem.

**Base URL**: `/api/v1`
**Authentication**: OAuth2 Bearer tokens (recommended) or Legacy JWT tokens
**Content Type**: `application/json`

## Authentication Modes

This service supports **dual authentication modes** for flexibility and backward compatibility:

### 1. OAuth2 Integration (Recommended)

**Features**:

- **Authorization Code Flow with PKCE** for web/mobile apps
- **Client Credentials Flow** for service-to-service authentication
- **Token Introspection** for real-time validation
- **Scope-based authorization** with granular permissions
- **JWT shared secret** validation for performance

**Available Scopes**:

- `openid` - Basic user identification
- `profile` - User profile information
- `user:read` - Read user data and profiles
- `user:write` - Create and update user data
- `admin` - Administrative operations

### 2. Legacy JWT Authentication (Backward Compatibility)

**Features**:

- Local JWT token generation and validation
- Role-based authorization (USER/ADMIN)
- Refresh token support
- Session management

**Migration Path**: Enable `OAUTH2_SERVICE_ENABLED=false` for legacy mode.

## OAuth2 Authentication Flow

### Authorization Code Flow with PKCE

#### Step 1: Authorization Request

```http
GET https://oauth2-service/authorize
  ?response_type=code
  &client_id=recipe-web-client
  &redirect_uri=https://app.example.com/callback
  &scope=openid profile user:read user:write
  &state=random_state_value
  &code_challenge=code_challenge_value
  &code_challenge_method=S256
```

#### Step 2: Token Exchange

```http
POST https://oauth2-service/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=authorization_code&
redirect_uri=https://app.example.com/callback&
client_id=recipe-web-client&
code_verifier=code_verifier_value
```

#### Step 3: Use Access Token

```http
GET /api/v1/users/profile
Authorization: Bearer <oauth2_access_token>
```

### Service-to-Service Authentication

**Client Credentials Flow**:

```http
POST https://oauth2-service/token
Content-Type: application/x-www-form-urlencoded
Authorization: Basic <base64(client_id:client_secret)>

grant_type=client_credentials&
scope=user:read user:write
```

## Legacy Authentication Endpoints

### Register User

```http
POST /auth/register
```

Register a new user account.

**Request Body:**

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response (201 Created):**

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "User registered successfully"
}
```

**Error Responses:**

- `400 Bad Request`: Invalid input data or user already exists
- `422 Unprocessable Entity`: Validation errors

### Login

```http
POST /auth/login
```

Authenticate user and receive JWT tokens.

**Request Body:**

```json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK):**

```json
{
  "access_token": "<access_token_string>",
  "refresh_token": "<refresh_token_string>",
  "token_type": "bearer",
  "expires_in": 1800
}
```

**Error Responses:**

- `401 Unauthorized`: Invalid credentials
- `422 Unprocessable Entity`: Validation errors

### Refresh Token

```http
POST /auth/refresh
Authorization: Bearer {refresh_token}
```

Get a new access token using a refresh token.

**Response (200 OK):**

```json
{
  "access_token": "<access_token_string>",
  "token_type": "bearer",
  "expires_in": 1800
}
```

### Logout

```http
POST /auth/logout
Authorization: Bearer {access_token}
```

Invalidate the current session.

**Response (200 OK):**

```json
{
  "message": "Successfully logged out"
}
```

### Password Reset Request

```http
POST /auth/password-reset
```

Request a password reset email.

**Request Body:**

```json
{
  "email": "john@example.com"
}
```

**Response (200 OK):**

```json
{
  "message": "Password reset email sent if account exists"
}
```

### Password Reset Confirm

```http
POST /auth/password-reset/confirm
```

Confirm password reset with token.

**Request Body:**

```json
{
  "token": "reset_token_here",
  "new_password": "NewSecurePass123!"
}
```

**Response (200 OK):**

```json
{
  "message": "Password reset successful"
}
```

## User Management Endpoints

### Get User Profile

```http
GET /users/{user_id}/profile
Authorization: Bearer {access_token}
```

Get a user's profile information.

**Response (200 OK):**

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "johndoe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Food enthusiast and home cook",
  "profile_picture_url": "https://example.com/profiles/johndoe.jpg",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-20T15:45:00Z",
  "privacy_settings": {
    "profile_visibility": "public",
    "show_email": false,
    "show_recipes": true
  }
}
```

### Update User Profile

```http
PUT /users/profile
Authorization: Bearer {access_token}
```

Update the current user's profile.

**Request Body:**

```json
{
  "first_name": "John",
  "last_name": "Smith",
  "bio": "Updated bio text",
  "profile_picture_url": "https://example.com/new-profile.jpg"
}
```

**Response (200 OK):**

```json
{
  "message": "Profile updated successfully",
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Smith",
    "bio": "Updated bio text",
    "updated_at": "2024-01-20T16:00:00Z"
  }
}
```

### Search Users

```http
GET /users/search?query=john&limit=10&offset=0
Authorization: Bearer {access_token}
```

Search for users by username, first name, or last name.

**Query Parameters:**

- `query` (string): Search term
- `limit` (integer): Number of results to return (1-100, default: 20)
- `offset` (integer): Number of results to skip (default: 0)

**Response (200 OK):**

```json
{
  "users": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "johndoe",
      "first_name": "John",
      "last_name": "Doe",
      "profile_picture_url": "https://example.com/profiles/johndoe.jpg",
      "bio": "Food enthusiast"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Delete User Account

```http
DELETE /users/account
Authorization: Bearer {access_token}
```

Delete the current user's account permanently.

**Response (200 OK):**

```json
{
  "message": "Account deleted successfully"
}
```

## Social Features Endpoints

### Follow User

```http
POST /social/follow
Authorization: Bearer {access_token}
```

Follow another user.

**Request Body:**

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Response (200 OK):**

```json
{
  "message": "Successfully followed user",
  "following": true
}
```

### Unfollow User

```http
DELETE /social/follow/{user_id}
Authorization: Bearer {access_token}
```

Unfollow a user.

**Response (200 OK):**

```json
{
  "message": "Successfully unfollowed user",
  "following": false
}
```

### Get Followers

```http
GET /social/followers?limit=20&offset=0
Authorization: Bearer {access_token}
```

Get the current user's followers.

**Response (200 OK):**

```json
{
  "followers": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440002",
      "username": "janedoe",
      "first_name": "Jane",
      "last_name": "Doe",
      "profile_picture_url": "https://example.com/profiles/janedoe.jpg",
      "followed_at": "2024-01-18T12:00:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### Get Following

```http
GET /social/following?limit=20&offset=0
Authorization: Bearer {access_token}
```

Get users that the current user is following.

**Response (200 OK):**

```json
{
  "following": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440003",
      "username": "chefmike",
      "first_name": "Mike",
      "last_name": "Chef",
      "profile_picture_url": "https://example.com/profiles/chefmike.jpg",
      "followed_at": "2024-01-17T09:30:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

## Notifications Endpoints

### Get Notifications

```http
GET /notifications?limit=20&offset=0&unread_only=false
Authorization: Bearer {access_token}
```

Get user notifications.

**Query Parameters:**

- `limit` (integer): Number of notifications to return (1-100, default: 20)
- `offset` (integer): Number of notifications to skip (default: 0)
- `unread_only` (boolean): Only return unread notifications (default: false)

**Response (200 OK):**

```json
{
  "notifications": [
    {
      "notification_id": "550e8400-e29b-41d4-a716-446655440010",
      "type": "follow",
      "title": "New Follower",
      "message": "Jane Doe started following you",
      "is_read": false,
      "created_at": "2024-01-20T14:30:00Z",
      "data": {
        "user_id": "550e8400-e29b-41d4-a716-446655440002",
        "username": "janedoe"
      }
    }
  ],
  "total": 1,
  "unread_count": 1,
  "limit": 20,
  "offset": 0
}
```

### Mark Notification as Read

```http
PUT /notifications/{notification_id}/read
Authorization: Bearer {access_token}
```

Mark a specific notification as read.

**Response (200 OK):**

```json
{
  "message": "Notification marked as read"
}
```

### Mark All Notifications as Read

```http
PUT /notifications/read-all
Authorization: Bearer {access_token}
```

Mark all notifications as read.

**Response (200 OK):**

```json
{
  "message": "All notifications marked as read",
  "updated_count": 5
}
```

### Delete Notification

```http
DELETE /notifications/{notification_id}
Authorization: Bearer {access_token}
```

Delete a specific notification.

**Response (200 OK):**

```json
{
  "message": "Notification deleted successfully"
}
```

## Admin Endpoints

**Note**: All admin endpoints require admin role authentication.

### Get User Statistics

```http
GET /admin/users/stats
Authorization: Bearer {admin_access_token}
```

Get user registration and activity statistics.

**Response (200 OK):**

```json
{
  "total_users": 1250,
  "active_users": 1100,
  "new_users_today": 15,
  "new_users_this_week": 85,
  "new_users_this_month": 320,
  "user_growth_rate": 12.5
}
```

### Get System Health

```http
GET /admin/system/health
Authorization: Bearer {admin_access_token}
```

Get detailed system health information.

**Response (200 OK):**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T16:00:00Z",
  "services": {
    "database": {
      "status": "healthy",
      "response_time_ms": 15,
      "connections": {
        "active": 8,
        "max": 20
      }
    },
    "redis": {
      "status": "healthy",
      "response_time_ms": 2,
      "memory_usage": "45MB",
      "connected_clients": 12
    }
  },
  "performance": {
    "uptime_seconds": 86400,
    "memory_usage_mb": 256,
    "cpu_usage_percent": 15.2
  }
}
```

### Force User Logout

```http
POST /admin/users/{user_id}/force-logout
Authorization: Bearer {admin_access_token}
```

Force logout a specific user from all sessions.

**Response (200 OK):**

```json
{
  "message": "User successfully logged out from all sessions",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Clear All Sessions

```http
POST /admin/sessions/clear-all
Authorization: Bearer {admin_access_token}
```

Clear all active user sessions.

**Response (200 OK):**

```json
{
  "message": "All sessions cleared successfully",
  "sessions_cleared": 150
}
```

## Health Check Endpoints

### Basic Health Check

```http
GET /health
```

Basic health check endpoint.

**Response (200 OK):**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T16:00:00Z"
}
```

### Detailed Health Check

```http
GET /health/detailed
```

Detailed health check with dependency status.

**Response (200 OK):**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T16:00:00Z",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy"
  },
  "version": "1.0.0",
  "uptime_seconds": 86400
}
```

## OAuth2 Error Responses

### 403 Forbidden (Insufficient Scopes)

```json
{
  "detail": "Missing required scope: user:write",
  "error_code": "INSUFFICIENT_SCOPE",
  "required_scopes": ["user:write"],
  "available_scopes": ["user:read", "profile"]
}
```

### 401 Unauthorized (Invalid Token)

```json
{
  "detail": "OAuth2 token validation failed",
  "error_code": "INVALID_OAUTH2_TOKEN",
  "token_type": "oauth2"
}
```

## Standard Error Responses

All endpoints may return the following error responses:

### 400 Bad Request

```json
{
  "detail": "Invalid request data",
  "error_code": "INVALID_REQUEST"
}
```

### 401 Unauthorized

```json
{
  "detail": "Authentication required",
  "error_code": "AUTHENTICATION_REQUIRED"
}
```

### 403 Forbidden

```json
{
  "detail": "Insufficient permissions",
  "error_code": "INSUFFICIENT_PERMISSIONS"
}
```

### 404 Not Found

```json
{
  "detail": "Resource not found",
  "error_code": "RESOURCE_NOT_FOUND"
}
```

### 422 Unprocessable Entity

```json
{
  "detail": [
    {
      "loc": ["body", "email"],
      "msg": "field required",
      "type": "value_error.missing"
    }
  ]
}
```

### 429 Too Many Requests

```json
{
  "detail": "Rate limit exceeded",
  "error_code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60
}
```

### 500 Internal Server Error

```json
{
  "detail": "Internal server error",
  "error_code": "INTERNAL_ERROR"
}
```

## Rate Limiting

Most endpoints are rate limited to prevent abuse:

- **Authentication endpoints**: 5 requests per minute per IP
- **User management endpoints**: 30 requests per minute per user
- **Social endpoints**: 20 requests per minute per user
- **Admin endpoints**: 100 requests per minute per admin user

Rate limit headers are included in responses:

- `X-RateLimit-Limit`: Request limit per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when rate limit resets (Unix timestamp)

## Authentication Headers

### OAuth2 Authentication

```http
Authorization: Bearer {oauth2_access_token}
```

### Legacy JWT Authentication

```http
Authorization: Bearer {jwt_access_token}
```

## Token Management

### OAuth2 Tokens

- **Access tokens**: Configured by OAuth2 service (typically 1-24 hours)
- **Refresh tokens**: Configured by OAuth2 service (typically 7-30 days)
- **Token introspection**: Real-time validation available
- **Token caching**: Service-to-service tokens cached for performance

### Legacy JWT Tokens

- **Access tokens**: Valid for 30 minutes
- **Refresh tokens**: Valid for 7 days
- **Password reset tokens**: Valid for 15 minutes

## Security Considerations

### OAuth2 Security

- **PKCE (Proof Key for Code Exchange)** prevents authorization code interception
- **Scope-based authorization** provides granular access control
- **Token introspection** enables real-time revocation
- **Client authentication** secures service-to-service communication

### General Security Notes

- Always use HTTPS in production
- Store tokens securely (not in localStorage for web apps)
- Implement proper token refresh logic
- Handle token expiration gracefully
- Validate token scopes for each operation

## Scope-Based Authorization

### Available Scopes

| Scope        | Description               | Endpoints Affected          |
| ------------ | ------------------------- | --------------------------- |
| `openid`     | Basic user identification | All authenticated endpoints |
| `profile`    | User profile information  | Profile endpoints           |
| `user:read`  | Read user data            | GET /users/\*, /search      |
| `user:write` | Modify user data          | POST/PUT /users/\*          |
| `admin`      | Administrative operations | /admin/\* endpoints         |

### Scope Validation

Endpoints automatically validate required scopes:

```http
# This request requires 'user:write' scope
PUT /api/v1/users/profile
Authorization: Bearer {token_with_user_write_scope}
```

If the token lacks required scopes, a 403 Forbidden response is returned with details about missing scopes.

### Fallback Authorization

When OAuth2 is disabled (`OAUTH2_SERVICE_ENABLED=false`), the service falls back to role-based authorization:

- `admin` scope → `ADMIN` role required
- `user:*` scopes → `USER` role required

## Pagination

List endpoints support pagination with the following parameters:

- `limit`: Number of items to return (1-100, default: 20)
- `offset`: Number of items to skip (default: 0)

Pagination response format:

```json
{
  "data": [...],
  "total": 100,
  "limit": 20,
  "offset": 0,
  "has_next": true,
  "has_previous": false
}
```

## Webhooks

The service supports webhooks for real-time notifications:

### Supported Events

- `user.registered`: New user registration
- `user.updated`: User profile updated
- `user.deleted`: User account deleted
- `social.followed`: User followed another user
- `social.unfollowed`: User unfollowed another user

### Webhook Payload Format

```json
{
  "event": "user.registered",
  "timestamp": "2024-01-20T16:00:00Z",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com"
  }
}
```

For webhook configuration and management, contact system administrators.

## OAuth2 Configuration

### Environment Variables

Required OAuth2 configuration:

```bash
# OAuth2 Integration
OAUTH2_SERVICE_ENABLED=true
OAUTH2_SERVICE_TO_SERVICE_ENABLED=true
OAUTH2_INTROSPECTION_ENABLED=false  # JWT validation (faster) vs introspection
OAUTH2_CLIENT_ID=your-client-id
OAUTH2_CLIENT_SECRET=your-client-secret
JWT_SECRET=shared-secret-with-oauth2-service
```

### Service URLs Configuration

OAuth2 service URLs are configured in `config/oauth2.json`:

```json
{
  "oauth2_service_urls": {
    "authorization_url": "https://oauth2-service/authorize",
    "token_url": "https://oauth2-service/token",
    "introspection_url": "https://oauth2-service/introspect",
    "userinfo_url": "https://oauth2-service/userinfo"
  },
  "scopes": {
    "default_scopes": ["openid", "profile"],
    "admin_scopes": ["openid", "profile", "admin"],
    "user_management_scopes": ["openid", "profile", "user:read", "user:write"]
  }
}
```

### Token Validation Modes

**1. JWT Shared Secret (Default - Faster)**:

- `OAUTH2_INTROSPECTION_ENABLED=false`
- Validates JWT signature using shared secret
- Offline validation, better performance
- Requires synchronized JWT secrets between services

**2. Token Introspection (Real-time)**:

- `OAUTH2_INTROSPECTION_ENABLED=true`
- Validates tokens by calling OAuth2 introspection endpoint
- Real-time validation, supports token revocation
- Requires network call to OAuth2 service

### Migration and Compatibility

**Phase 1 - Enable OAuth2**:

```bash
OAUTH2_SERVICE_ENABLED=true
OAUTH2_INTROSPECTION_ENABLED=false
```

**Phase 2 - Enable Introspection** (optional):

```bash
OAUTH2_INTROSPECTION_ENABLED=true
```

**Phase 3 - Disable Legacy JWT** (optional):

- Remove legacy JWT endpoints
- Update client applications to use OAuth2 flow

The service maintains full backward compatibility - existing JWT tokens continue to work when OAuth2 is enabled.
