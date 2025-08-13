# User Management Service API Documentation

## Overview

The User Management Service provides comprehensive user authentication, profile management, social features, and
administrative capabilities for the Recipe Web Application ecosystem.

**Base URL**: `/api/v1`
**Authentication**: Bearer JWT tokens
**Content Type**: `application/json`

## Authentication Endpoints

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

## Error Responses

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

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the access token in the Authorization header:

```http
Authorization: Bearer {access_token}
```

### Token Lifecycle

- **Access tokens**: Valid for 30 minutes
- **Refresh tokens**: Valid for 7 days
- **Password reset tokens**: Valid for 15 minutes

### Security Notes

- Always use HTTPS in production
- Store tokens securely (not in localStorage for web apps)
- Implement proper token refresh logic
- Handle token expiration gracefully

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
