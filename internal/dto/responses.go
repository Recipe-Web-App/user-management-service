package dto

import "time"

// ============================================================================
// User Management Responses
// ============================================================================

// UserProfileResponse represents a user profile.
type UserProfileResponse struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Email     *string   `json:"email,omitempty"`
	FullName  *string   `json:"fullName,omitempty"`
	Bio       *string   `json:"bio,omitempty"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserSearchResult represents a user in search results.
type UserSearchResult struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	FullName  *string   `json:"fullName,omitempty"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserSearchResponse represents search results.
type UserSearchResponse struct {
	Results    []UserSearchResult `json:"results"`
	TotalCount int                `json:"totalCount"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// UserAccountDeleteRequestResponse represents the response for account deletion request.
type UserAccountDeleteRequestResponse struct {
	UserID            string    `json:"userId"`
	ConfirmationToken string    `json:"confirmationToken"`
	ExpiresAt         time.Time `json:"expiresAt"`
}

// UserConfirmAccountDeleteResponse represents the response for confirmed account deletion.
type UserConfirmAccountDeleteResponse struct {
	UserID        string    `json:"userId"`
	DeactivatedAt time.Time `json:"deactivatedAt"`
}

// ============================================================================
// Social Feature Responses
// ============================================================================

// User represents a user in social contexts.
type User struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Email     *string   `json:"email,omitempty"`
	FullName  *string   `json:"fullName,omitempty"`
	Bio       *string   `json:"bio,omitempty"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GetFollowedUsersResponse represents the response for following/followers list.
type GetFollowedUsersResponse struct {
	TotalCount    int    `json:"totalCount"`
	FollowedUsers []User `json:"followedUsers,omitempty"`
	Limit         *int   `json:"limit,omitempty"`
	Offset        *int   `json:"offset,omitempty"`
}

// FollowResponse represents the response for follow/unfollow actions.
type FollowResponse struct {
	Message     string `json:"message"`
	IsFollowing bool   `json:"isFollowing"`
}

// RecipeSummary represents a recipe in activity.
type RecipeSummary struct {
	RecipeID  int       `json:"recipeId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserSummary represents a followed user in activity.
type UserSummary struct {
	UserID     string    `json:"userId"`
	Username   string    `json:"username"`
	FollowedAt time.Time `json:"followedAt"`
}

// ReviewSummary represents a review in activity.
type ReviewSummary struct {
	ReviewID  int       `json:"reviewId"`
	RecipeID  int       `json:"recipeId"`
	Rating    float64   `json:"rating"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// FavoriteSummary represents a favorite in activity.
type FavoriteSummary struct {
	RecipeID    int       `json:"recipeId"`
	Title       string    `json:"title"`
	FavoritedAt time.Time `json:"favoritedAt"`
}

// UserActivityResponse represents user activity data.
type UserActivityResponse struct {
	UserID          string            `json:"userId"`
	RecentRecipes   []RecipeSummary   `json:"recentRecipes"`
	RecentFollows   []UserSummary     `json:"recentFollows"`
	RecentReviews   []ReviewSummary   `json:"recentReviews"`
	RecentFavorites []FavoriteSummary `json:"recentFavorites"`
}

// ============================================================================
// Notification Responses
// ============================================================================

// Notification represents a notification.
type Notification struct {
	NotificationID   string    `json:"notificationId"`
	UserID           string    `json:"userId"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	NotificationType string    `json:"notificationType"`
	IsRead           bool      `json:"isRead"`
	IsDeleted        bool      `json:"isDeleted"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// NotificationListResponse represents a list of notifications.
type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	TotalCount    int            `json:"totalCount"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
}

// NotificationCountResponse represents notification count.
type NotificationCountResponse struct {
	TotalCount int `json:"totalCount"`
}

// NotificationReadResponse represents the response for marking a notification read.
type NotificationReadResponse struct {
	Message string `json:"message"`
}

// NotificationReadAllResponse represents the response for marking all notifications read.
type NotificationReadAllResponse struct {
	Message             string   `json:"message"`
	ReadNotificationIDs []string `json:"readNotificationIds"`
}

// NotificationDeleteResponse represents the response for deleting notifications.
type NotificationDeleteResponse struct {
	Message                string   `json:"message"`
	DeletedNotificationIDs []string `json:"deletedNotificationIds"`
}

// ============================================================================
// User Preferences Responses
// ============================================================================

// NotificationPreferences represents notification settings.
type NotificationPreferences struct {
	EmailNotifications   bool `json:"emailNotifications"`
	PushNotifications    bool `json:"pushNotifications"`
	FollowNotifications  bool `json:"followNotifications"`
	LikeNotifications    bool `json:"likeNotifications"`
	CommentNotifications bool `json:"commentNotifications"`
	RecipeNotifications  bool `json:"recipeNotifications"`
	SystemNotifications  bool `json:"systemNotifications"`
}

// PrivacyPreferences represents privacy settings.
type PrivacyPreferences struct {
	ProfileVisibility string `json:"profileVisibility"`
	ShowEmail         bool   `json:"showEmail"`
	ShowFullName      bool   `json:"showFullName"`
	AllowFollows      bool   `json:"allowFollows"`
	AllowMessages     bool   `json:"allowMessages"`
}

// DisplayPreferences represents display settings.
type DisplayPreferences struct {
	Theme    string `json:"theme"`
	Language string `json:"language"`
	Timezone string `json:"timezone"`
}

// UserPreferences represents all user preferences.
type UserPreferences struct {
	NotificationPreferences *NotificationPreferences `json:"notificationPreferences,omitempty"`
	PrivacyPreferences      *PrivacyPreferences      `json:"privacyPreferences,omitempty"`
	DisplayPreferences      *DisplayPreferences      `json:"displayPreferences,omitempty"`
}

// UserPreferenceResponse represents the response for user preferences.
type UserPreferenceResponse struct {
	Preferences UserPreferences `json:"preferences"`
}

// ============================================================================
// Admin Responses
// ============================================================================

// RedisSessionStatsResponse represents Redis session statistics.
type RedisSessionStatsResponse struct {
	TotalSessions  int            `json:"totalSessions"`
	ActiveSessions int            `json:"activeSessions"`
	MemoryUsage    string         `json:"memoryUsage"`
	TTLInfo        map[string]any `json:"ttlInfo"`
}

// UserStatsResponse represents user statistics.
type UserStatsResponse struct {
	TotalUsers        int `json:"totalUsers"`
	ActiveUsers       int `json:"activeUsers"`
	InactiveUsers     int `json:"inactiveUsers"`
	NewUsersToday     int `json:"newUsersToday"`
	NewUsersThisWeek  int `json:"newUsersThisWeek"`
	NewUsersThisMonth int `json:"newUsersThisMonth"`
}

// SystemHealthResponse represents system health status.
type SystemHealthResponse struct {
	Status         string `json:"status"`
	DatabaseStatus string `json:"databaseStatus"`
	RedisStatus    string `json:"redisStatus"`
	UptimeSeconds  int    `json:"uptimeSeconds"`
	Version        string `json:"version"`
}

// ForceLogoutResponse represents force logout response.
type ForceLogoutResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	SessionsCleared int    `json:"sessionsCleared"`
}

// ClearSessionsResponse represents clear sessions response.
type ClearSessionsResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	SessionsCleared int    `json:"sessionsCleared"`
}

// ============================================================================
// Metrics Responses
// ============================================================================

// ResponseTimes represents response time metrics.
type ResponseTimes struct {
	AverageMs float64 `json:"averageMs"`
	P50Ms     float64 `json:"p50Ms"`
	P95Ms     float64 `json:"p95Ms"`
	P99Ms     float64 `json:"p99Ms"`
}

// RequestCounts represents request count metrics.
type RequestCounts struct {
	TotalRequests     int `json:"totalRequests"`
	RequestsPerMinute int `json:"requestsPerMinute"`
	ActiveSessions    int `json:"activeSessions"`
}

// ErrorRates represents error rate metrics.
type ErrorRates struct {
	TotalErrors      int     `json:"totalErrors"`
	ErrorRatePercent float64 `json:"errorRatePercent"`
	Errors4xx        int     `json:"errors4xx"`
	Errors5xx        int     `json:"errors5xx"`
}

// DatabaseMetrics represents database metrics.
type DatabaseMetrics struct {
	ActiveConnections int     `json:"activeConnections"`
	MaxConnections    int     `json:"maxConnections"`
	AvgQueryTimeMs    float64 `json:"avgQueryTimeMs"`
	SlowQueriesCount  int     `json:"slowQueriesCount"`
}

// PerformanceMetricsResponse represents performance metrics.
type PerformanceMetricsResponse struct {
	ResponseTimes ResponseTimes   `json:"responseTimes"`
	RequestCounts RequestCounts   `json:"requestCounts"`
	ErrorRates    ErrorRates      `json:"errorRates"`
	Database      DatabaseMetrics `json:"database"`
}

// CacheMetricsResponse represents cache metrics.
type CacheMetricsResponse struct {
	MemoryUsage      string  `json:"memoryUsage"`
	MemoryUsageHuman string  `json:"memoryUsageHuman"`
	KeysCount        int     `json:"keysCount"`
	HitRate          float64 `json:"hitRate"`
	ConnectedClients int     `json:"connectedClients"`
	EvictedKeys      int     `json:"evictedKeys"`
	ExpiredKeys      int     `json:"expiredKeys"`
}

// CacheClearResponse represents cache clear response.
type CacheClearResponse struct {
	Message      string `json:"message"`
	Pattern      string `json:"pattern"`
	ClearedCount int    `json:"clearedCount"`
}

// SystemInfo represents system resource information.
type SystemInfo struct {
	CPUUsagePercent    float64 `json:"cpuUsagePercent"`
	MemoryTotalGB      float64 `json:"memoryTotalGb"`
	MemoryUsedGB       float64 `json:"memoryUsedGb"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent"`
	DiskTotalGB        float64 `json:"diskTotalGb"`
	DiskUsedGB         float64 `json:"diskUsedGb"`
	DiskUsagePercent   float64 `json:"diskUsagePercent"`
}

// ProcessInfo represents process information.
type ProcessInfo struct {
	MemoryRSSMB float64 `json:"memoryRssMb"`
	MemoryVMSMB float64 `json:"memoryVmsMb"`
	CPUPercent  float64 `json:"cpuPercent"`
	NumThreads  int     `json:"numThreads"`
	OpenFiles   int     `json:"openFiles"`
}

// SystemMetricsResponse represents system metrics.
type SystemMetricsResponse struct {
	Timestamp     time.Time   `json:"timestamp"`
	System        SystemInfo  `json:"system"`
	Process       ProcessInfo `json:"process"`
	UptimeSeconds int         `json:"uptimeSeconds"`
}

// RedisHealth represents Redis health details.
type RedisHealth struct {
	Status           string  `json:"status"`
	ResponseTimeMs   float64 `json:"responseTimeMs"`
	MemoryUsage      string  `json:"memoryUsage"`
	ConnectedClients int     `json:"connectedClients"`
	HitRatePercent   float64 `json:"hitRatePercent"`
}

// DatabaseHealth represents database health details.
type DatabaseHealth struct {
	Status            string  `json:"status"`
	ResponseTimeMs    float64 `json:"responseTimeMs"`
	ActiveConnections int     `json:"activeConnections"`
	MaxConnections    int     `json:"maxConnections"`
}

// ServicesHealth represents all services health.
type ServicesHealth struct {
	Redis    RedisHealth    `json:"redis"`
	Database DatabaseHealth `json:"database"`
}

// ApplicationFeatures represents application feature status.
type ApplicationFeatures struct {
	Authentication  string `json:"authentication"`
	Caching         string `json:"caching"`
	Monitoring      string `json:"monitoring"`
	SecurityHeaders string `json:"securityHeaders"`
}

// ApplicationInfo represents application information.
type ApplicationInfo struct {
	Version     string              `json:"version"`
	Environment string              `json:"environment"`
	Features    ApplicationFeatures `json:"features"`
}

// DetailedHealthMetricsResponse represents detailed health metrics.
type DetailedHealthMetricsResponse struct {
	Timestamp     time.Time       `json:"timestamp"`
	OverallStatus string          `json:"overallStatus"`
	Services      ServicesHealth  `json:"services"`
	Application   ApplicationInfo `json:"application"`
}
