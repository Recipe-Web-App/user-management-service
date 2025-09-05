# User Management API - Postman Collection

This directory contains Postman collection and environment files for
comprehensive API testing of the User Management Service.

## Files Overview

### Collection Files

- **`User-Management-API.postman_collection.json`** - Complete API endpoints
  including user management, authentication, social features, notifications,
  admin operations, and system monitoring

### Environment Files

- **`User-Management-Development.postman_environment.json`** - Development
  environment variables (passwords as placeholders)
- **`User-Management-Local.postman_environment.json`** - Local development
  environment variables (passwords as placeholders)
- **`*-Private.postman_environment.json`** - Local-only files with real
  passwords (gitignored)

### Setup Instructions

#### 1. Import Collections and Environments

1. **Import Collection:**
   - Open Postman
   - Click "Import" button
   - Select `User-Management-API.postman_collection.json`
   - Collection will appear in your workspace

2. **Import Environment Templates:**
   - Import both environment files:
     `User-Management-Development.postman_environment.json` and
     `User-Management-Local.postman_environment.json`

#### 2. Set Up Private Environment with Passwords

The environment files in Git have placeholder values for passwords. To use them locally:

1. **Create Private Environment Files:**

   ```bash
   # Copy the environment files and add '-Private' suffix
   cp User-Management-Development.postman_environment.json \
      User-Management-Development-Private.postman_environment.json
   cp User-Management-Local.postman_environment.json \
      User-Management-Local-Private.postman_environment.json
   ```

2. **Add Real Passwords:**
   Edit your `-Private` files and replace these placeholder values:
   - `REPLACE_WITH_YOUR_TEST_USER_PASSWORD` → Your actual test user password
   - `REPLACE_WITH_YOUR_ADMIN_USER_PASSWORD` → Your actual admin user password

3. **Import Private Environments:**
   - Import your `-Private.postman_environment.json` files into Postman
   - Use these private environments for actual testing
   - The `-Private` files are automatically gitignored

4. **Select Environment:**
   - Choose the appropriate private environment from the dropdown in
     Postman's top-right corner

## Collection Structure

### 1. User Management

Core user operations and profile management:

- **Get Users** - Retrieve paginated list of users with search and filtering
- **Get User by ID** - Retrieve specific user profile (respects privacy settings)
- **Update User Profile** - Update user information and settings
- **Delete User Account** - Account deletion with confirmation workflow
- **Search Users** - Advanced user search with filters and pagination

### 2. Social Features

User relationship and social interaction management:

#### Following/Followers

- **Follow User** - Create following relationship
- **Unfollow User** - Remove following relationship
- **Get Followers** - Retrieve user's followers list
- **Get Following** - Retrieve users that current user follows
- **Get Follow Stats** - Get follower/following counts

#### Social Analytics

- **Get User Feed** - Activity feed from followed users
- **Get Social Recommendations** - User discovery suggestions

### 3. Notifications

Comprehensive notification system:

#### Notification Management

- **Get Notifications** - Retrieve user notifications with pagination
- **Mark Notification as Read** - Mark individual notification as read
- **Mark All as Read** - Mark all notifications as read
- **Delete Notification** - Remove specific notification
- **Delete All Notifications** - Clear all notifications

#### Notification Preferences

- **Get Notification Preferences** - Retrieve user notification settings
- **Update Notification Preferences** - Configure notification types and delivery methods

#### Bulk Operations

- **Bulk Mark as Read** - Mark multiple notifications as read
- **Bulk Delete** - Delete multiple notifications

### 4. Admin Operations

Administrative functions requiring admin privileges:

#### User Administration

- **Get All Users (Admin)** - Administrative user listing with sensitive data
- **Update User (Admin)** - Administrative user updates including role changes
- **Delete User (Admin)** - Administrative user deletion
- **Get User Activity Logs** - Audit trail and activity monitoring

#### System Monitoring

- **Get System Stats** - User counts, activity metrics, and system overview
- **Get Activity Logs** - System-wide activity monitoring and audit trails

### 5. Health & Monitoring

System health checks and operational monitoring:

#### Health Checks

- **Basic Health Check** - Simple service availability check
- **Detailed Health Check** - Comprehensive service health including dependencies

#### System Metrics

- **Performance Metrics** - Response times, request counts, error rates, database stats
- **Cache Metrics** - Redis cache statistics and performance data
- **System Metrics** - CPU, memory, disk usage, and process information
- **Clear Cache** - Administrative cache clearing functionality

## Environment Variables

### Base URLs

- **`baseUrl`** - User Management service base URL

### User Credentials (Test User)

- **`testUsername`** - Test user username
- **`testEmail`** - Test user email
- **`testFullName`** - Test user full name
- **`testPassword`** - Test user password (secret type)

### User Credentials (Admin User)

- **`adminUsername`** - Admin user username
- **`adminEmail`** - Admin user email
- **`adminPassword`** - Admin user password (secret type)

### Authentication Tokens (Auto-managed)

These variables are automatically set by test scripts:

- **`accessToken`** - OAuth2 Bearer token for authenticated requests (secret type)
- **`adminAccessToken`** - Admin OAuth2 Bearer token for administrative requests (secret type)
- **`userId`** - Current authenticated user's ID

### Test Data Variables

- **`targetUserId`** - Target user ID for operations on other users
- **`notificationId`** - Sample notification ID for testing
- **`notificationId1`** - First notification ID for bulk operations
- **`notificationId2`** - Second notification ID for bulk operations
- **`deletionToken`** - Account deletion confirmation token

## Authentication Model

**Important:** This service uses **external OAuth2 authentication**. The collection assumes:

- Users authenticate through an external OAuth2 provider
- Access tokens are obtained externally and stored in environment variables
- The service validates Bearer tokens but does not provide login/registration endpoints
- Admin privileges are determined by the OAuth2 provider's role claims

### Token Management

- Access tokens must be manually obtained from your OAuth2 provider
- Store tokens in the `accessToken` and `adminAccessToken` environment variables
- Tokens are used automatically in requests requiring authentication
- Update tokens manually when they expire

## Automatic Response Field Extraction

The collection includes test scripts that automatically extract important response fields:

### User Operations

- Create/Update operations extract user IDs and relevant data
- Search operations validate pagination and result structure

### Social Features

- Follow operations update relationship status
- Social stats are extracted for verification

### Notifications

- Notification operations extract notification IDs for subsequent operations
- Bulk operations track affected notification counts

### Admin Operations

- Administrative operations extract audit information
- System metrics are validated for proper structure

## Environment Switching

**Development Environment:**

- Service: `http://user-management.dev.local/api/v1`

**Local Environment:**

- Service: `http://localhost:8000/api/v1`

Switch between environments using the environment selector dropdown in
Postman's top-right corner.

## Security Features

- **Password Protection**: Sensitive passwords are excluded from Git repository
- **Private Environment Pattern**: Use local `-Private` files for
  credentials (automatically gitignored)
- **Secret Variables**: Passwords and tokens are marked as secret type in
  Postman
- **External Authentication**: OAuth2 integration with external providers
- **Role-Based Access**: Admin operations require appropriate OAuth2 role claims
- **Environment Isolation**: Separate environments prevent accidental
  cross-environment requests

### Security Model

- **Git Repository**: Contains collections and environment templates with
  placeholder passwords
- **Local Development**: Uses private environment files with real
  credentials and OAuth2 tokens
- **Team Collaboration**: Secure sharing of API structure without
  exposing credentials
- **OAuth2 Integration**: Leverages external authentication providers for security

## Usage Workflow

### Getting Started

1. Import collection file and environment templates
2. Set up private environment files with real OAuth2 tokens (see setup
   instructions above)
3. Select appropriate private environment (Development-Private or
   Local-Private)
4. Ensure OAuth2 tokens are valid and stored in environment variables
5. Start with basic user operations to verify authentication
6. Use specialized folders for targeted testing of specific functionality

### Typical Testing Flow

1. **Verify Authentication** - Use basic user operations to confirm OAuth2 tokens work
2. **Test Core User Operations** - Profile management, user search, and CRUD operations
3. **Test Social Features** - Following/followers, social feed, and recommendations
4. **Test Notifications** - Notification management and preferences
5. **Test Admin Features** - Administrative operations (requires admin token)
6. **Monitor System Health** - Use health checks and metrics endpoints
7. **Token Refresh** - Manually update OAuth2 tokens when they expire

## Test Script Features

All requests include comprehensive test scripts that:

- Validate HTTP status codes and response structure
- Check authentication and authorization
- Extract and store important response data as environment variables
- Provide clear test result feedback
- Enable request chaining through automatic variable management
- Validate OAuth2 Bearer token usage

## Privacy and Security Considerations

- **Privacy Settings**: User profile endpoints respect privacy preferences
- **Data Protection**: Sensitive user data is protected based on access levels
- **Audit Trail**: Administrative operations are logged for security monitoring
- **Rate Limiting**: API endpoints implement appropriate rate limiting
- **OAuth2 Security**: External authentication provides enterprise-grade security

This collection provides comprehensive API testing capabilities for the User
Management Service with OAuth2 integration, social features, notifications,
and administrative functions.
