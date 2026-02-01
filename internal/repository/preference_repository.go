package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// ErrPreferencesNotFound is returned when preferences don't exist for a user.
var ErrPreferencesNotFound = errors.New("preferences not found")

// UserExistsChecker checks if a user exists.
type UserExistsChecker interface {
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
}

// NotificationPreferenceRepo handles notification preferences.
type NotificationPreferenceRepo interface {
	GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*dto.NotificationPreferences, error)
	UpdateNotificationPreferences(
		ctx context.Context, userID uuid.UUID, u *dto.NotificationPreferencesUpdate,
	) (*dto.NotificationPreferences, error)
}

// DisplayPreferenceRepo handles display preferences.
type DisplayPreferenceRepo interface {
	GetDisplayPreferences(ctx context.Context, userID uuid.UUID) (*dto.DisplayPreferences, error)
	UpdateDisplayPreferences(
		ctx context.Context, userID uuid.UUID, u *dto.DisplayPreferencesUpdate,
	) (*dto.DisplayPreferences, error)
}

// PrivacyPreferenceRepo handles privacy preferences.
type PrivacyPreferenceRepo interface {
	GetPrivacyPreferencesData(ctx context.Context, userID uuid.UUID) (*dto.UserPrivacyPreferences, error)
	UpdatePrivacyPreferencesData(
		ctx context.Context, userID uuid.UUID, u *dto.PrivacyPreferencesUpdate,
	) (*dto.UserPrivacyPreferences, error)
}

// AccessibilityPreferenceRepo handles accessibility preferences.
type AccessibilityPreferenceRepo interface {
	GetAccessibilityPreferences(ctx context.Context, userID uuid.UUID) (*dto.AccessibilityPreferences, error)
	UpdateAccessibilityPreferences(
		ctx context.Context, userID uuid.UUID, u *dto.AccessibilityPreferencesUpdate,
	) (*dto.AccessibilityPreferences, error)
}

// LanguagePreferenceRepo handles language preferences.
type LanguagePreferenceRepo interface {
	GetLanguagePreferences(ctx context.Context, userID uuid.UUID) (*dto.LanguagePreferences, error)
	UpdateLanguagePreferences(
		ctx context.Context, userID uuid.UUID, u *dto.LanguagePreferencesUpdate,
	) (*dto.LanguagePreferences, error)
}

// SecurityPreferenceRepo handles security preferences.
type SecurityPreferenceRepo interface {
	GetSecurityPreferences(ctx context.Context, userID uuid.UUID) (*dto.SecurityPreferences, error)
	UpdateSecurityPreferences(
		ctx context.Context, userID uuid.UUID, u *dto.SecurityPreferencesUpdate,
	) (*dto.SecurityPreferences, error)
}

// SocialPreferenceRepo handles social preferences.
type SocialPreferenceRepo interface {
	GetSocialPreferences(ctx context.Context, userID uuid.UUID) (*dto.SocialPreferences, error)
	UpdateSocialPreferences(
		ctx context.Context, userID uuid.UUID, u *dto.SocialPreferencesUpdate,
	) (*dto.SocialPreferences, error)
}

// SoundPreferenceRepo handles sound preferences.
type SoundPreferenceRepo interface {
	GetSoundPreferences(ctx context.Context, userID uuid.UUID) (*dto.SoundPreferences, error)
	UpdateSoundPreferences(
		ctx context.Context, userID uuid.UUID, u *dto.SoundPreferencesUpdate,
	) (*dto.SoundPreferences, error)
}

// ThemePreferenceRepo handles theme preferences.
type ThemePreferenceRepo interface {
	GetThemePreferences(ctx context.Context, userID uuid.UUID) (*dto.ThemePreferences, error)
	UpdateThemePreferences(
		ctx context.Context, userID uuid.UUID, u *dto.ThemePreferencesUpdate,
	) (*dto.ThemePreferences, error)
}

// PreferenceRepository combines all preference repository interfaces.
type PreferenceRepository interface {
	UserExistsChecker
	NotificationPreferenceRepo
	DisplayPreferenceRepo
	PrivacyPreferenceRepo
	AccessibilityPreferenceRepo
	LanguagePreferenceRepo
	SecurityPreferenceRepo
	SocialPreferenceRepo
	SoundPreferenceRepo
	ThemePreferenceRepo
}

// SQLPreferenceRepository implements PreferenceRepository using SQL.
type SQLPreferenceRepository struct {
	db *sql.DB
}

// NewPreferenceRepository creates a new SQLPreferenceRepository.
func NewPreferenceRepository(db *sql.DB) *SQLPreferenceRepository {
	return &SQLPreferenceRepository{db: db}
}

// UserExists checks if a user exists.
func (r *SQLPreferenceRepository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM recipe_manager.users WHERE user_id = $1)`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// GetNotificationPreferences retrieves notification preferences for a user.
func (r *SQLPreferenceRepository) GetNotificationPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.NotificationPreferences, error) {
	query := `
		SELECT email_notifications, push_notifications, sms_notifications,
		       marketing_emails, security_alerts, activity_summaries,
		       recipe_recommendations, social_interactions, updated_at
		FROM recipe_manager.user_notification_preferences
		WHERE user_id = $1
	`

	prefs := &dto.NotificationPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.EmailNotifications,
		&prefs.PushNotifications,
		&prefs.SMSNotifications,
		&prefs.MarketingEmails,
		&prefs.SecurityAlerts,
		&prefs.ActivitySummaries,
		&prefs.RecipeRecommendations,
		&prefs.SocialInteractions,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultNotificationPreferences(), nil
		}

		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}

	return prefs, nil
}

func defaultNotificationPreferences() *dto.NotificationPreferences {
	return &dto.NotificationPreferences{
		EmailNotifications:    true,
		PushNotifications:     true,
		SMSNotifications:      false,
		MarketingEmails:       false,
		SecurityAlerts:        true,
		ActivitySummaries:     true,
		RecipeRecommendations: true,
		SocialInteractions:    true,
		UpdatedAt:             time.Now(),
	}
}

// UpdateNotificationPreferences updates notification preferences using upsert.
func (r *SQLPreferenceRepository) UpdateNotificationPreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.NotificationPreferencesUpdate,
) (*dto.NotificationPreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_notification_preferences (
			user_id, email_notifications, push_notifications, sms_notifications,
			marketing_emails, security_alerts, activity_summaries,
			recipe_recommendations, social_interactions, updated_at
		)
		VALUES ($1,
			COALESCE($2, true), COALESCE($3, true), COALESCE($4, false),
			COALESCE($5, false), COALESCE($6, true), COALESCE($7, true),
			COALESCE($8, true), COALESCE($9, true), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			email_notifications = COALESCE($2, user_notification_preferences.email_notifications),
			push_notifications = COALESCE($3, user_notification_preferences.push_notifications),
			sms_notifications = COALESCE($4, user_notification_preferences.sms_notifications),
			marketing_emails = COALESCE($5, user_notification_preferences.marketing_emails),
			security_alerts = COALESCE($6, user_notification_preferences.security_alerts),
			activity_summaries = COALESCE($7, user_notification_preferences.activity_summaries),
			recipe_recommendations = COALESCE($8, user_notification_preferences.recipe_recommendations),
			social_interactions = COALESCE($9, user_notification_preferences.social_interactions),
			updated_at = NOW()
		RETURNING email_notifications, push_notifications, sms_notifications,
		          marketing_emails, security_alerts, activity_summaries,
		          recipe_recommendations, social_interactions, updated_at
	`

	prefs := &dto.NotificationPreferences{}

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.EmailNotifications,
		update.PushNotifications,
		update.SMSNotifications,
		update.MarketingEmails,
		update.SecurityAlerts,
		update.ActivitySummaries,
		update.RecipeRecommendations,
		update.SocialInteractions,
	).Scan(
		&prefs.EmailNotifications,
		&prefs.PushNotifications,
		&prefs.SMSNotifications,
		&prefs.MarketingEmails,
		&prefs.SecurityAlerts,
		&prefs.ActivitySummaries,
		&prefs.RecipeRecommendations,
		&prefs.SocialInteractions,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update notification preferences: %w", err)
	}

	return prefs, nil
}

// GetDisplayPreferences retrieves display preferences for a user.
func (r *SQLPreferenceRepository) GetDisplayPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.DisplayPreferences, error) {
	query := `
		SELECT font_size, color_scheme, layout_density, show_images, compact_mode, updated_at
		FROM recipe_manager.user_display_preferences
		WHERE user_id = $1
	`

	prefs := &dto.DisplayPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.FontSize,
		&prefs.ColorScheme,
		&prefs.LayoutDensity,
		&prefs.ShowImages,
		&prefs.CompactMode,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultDisplayPreferences(), nil
		}

		return nil, fmt.Errorf("failed to get display preferences: %w", err)
	}

	return prefs, nil
}

func defaultDisplayPreferences() *dto.DisplayPreferences {
	return &dto.DisplayPreferences{
		FontSize:      dto.FontSizeMedium,
		ColorScheme:   dto.ColorSchemeLight,
		LayoutDensity: dto.LayoutDensityComfortable,
		ShowImages:    true,
		CompactMode:   false,
		UpdatedAt:     time.Now(),
	}
}

// UpdateDisplayPreferences updates display preferences using upsert.
func (r *SQLPreferenceRepository) UpdateDisplayPreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.DisplayPreferencesUpdate,
) (*dto.DisplayPreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_display_preferences (
			user_id, font_size, color_scheme, layout_density, show_images, compact_mode, updated_at
		)
		VALUES ($1,
			COALESCE($2, 'MEDIUM'), COALESCE($3, 'LIGHT'), COALESCE($4, 'COMFORTABLE'),
			COALESCE($5, true), COALESCE($6, false), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			font_size = COALESCE($2, user_display_preferences.font_size),
			color_scheme = COALESCE($3, user_display_preferences.color_scheme),
			layout_density = COALESCE($4, user_display_preferences.layout_density),
			show_images = COALESCE($5, user_display_preferences.show_images),
			compact_mode = COALESCE($6, user_display_preferences.compact_mode),
			updated_at = NOW()
		RETURNING font_size, color_scheme, layout_density, show_images, compact_mode, updated_at
	`

	prefs := &dto.DisplayPreferences{}

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.FontSize,
		update.ColorScheme,
		update.LayoutDensity,
		update.ShowImages,
		update.CompactMode,
	).Scan(
		&prefs.FontSize,
		&prefs.ColorScheme,
		&prefs.LayoutDensity,
		&prefs.ShowImages,
		&prefs.CompactMode,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update display preferences: %w", err)
	}

	return prefs, nil
}

// GetPrivacyPreferencesData retrieves privacy preferences for a user.
func (r *SQLPreferenceRepository) GetPrivacyPreferencesData(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.UserPrivacyPreferences, error) {
	query := `
		SELECT profile_visibility, recipe_visibility, activity_visibility,
		       contact_info_visibility, data_sharing, analytics_tracking, updated_at
		FROM recipe_manager.user_privacy_preferences
		WHERE user_id = $1
	`

	prefs := &dto.UserPrivacyPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.ProfileVisibility,
		&prefs.RecipeVisibility,
		&prefs.ActivityVisibility,
		&prefs.ContactInfoVisibility,
		&prefs.DataSharing,
		&prefs.AnalyticsTracking,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultPrivacyPreferences(), nil
		}

		return nil, fmt.Errorf("failed to get privacy preferences: %w", err)
	}

	return prefs, nil
}

func defaultPrivacyPreferences() *dto.UserPrivacyPreferences {
	return &dto.UserPrivacyPreferences{
		ProfileVisibility:     dto.ProfileVisibilityPublic,
		RecipeVisibility:      dto.ProfileVisibilityPublic,
		ActivityVisibility:    dto.ProfileVisibilityPublic,
		ContactInfoVisibility: dto.ProfileVisibilityPrivate,
		DataSharing:           false,
		AnalyticsTracking:     false,
		UpdatedAt:             time.Now(),
	}
}

// UpdatePrivacyPreferencesData updates privacy preferences using upsert.
func (r *SQLPreferenceRepository) UpdatePrivacyPreferencesData(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.PrivacyPreferencesUpdate,
) (*dto.UserPrivacyPreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_privacy_preferences (
			user_id, profile_visibility, recipe_visibility, activity_visibility,
			contact_info_visibility, data_sharing, analytics_tracking, updated_at
		)
		VALUES ($1,
			COALESCE($2, 'PUBLIC'), COALESCE($3, 'PUBLIC'), COALESCE($4, 'PUBLIC'),
			COALESCE($5, 'PRIVATE'), COALESCE($6, false), COALESCE($7, false), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			profile_visibility = COALESCE($2, user_privacy_preferences.profile_visibility),
			recipe_visibility = COALESCE($3, user_privacy_preferences.recipe_visibility),
			activity_visibility = COALESCE($4, user_privacy_preferences.activity_visibility),
			contact_info_visibility = COALESCE($5, user_privacy_preferences.contact_info_visibility),
			data_sharing = COALESCE($6, user_privacy_preferences.data_sharing),
			analytics_tracking = COALESCE($7, user_privacy_preferences.analytics_tracking),
			updated_at = NOW()
		RETURNING profile_visibility, recipe_visibility, activity_visibility,
		          contact_info_visibility, data_sharing, analytics_tracking, updated_at
	`

	prefs := &dto.UserPrivacyPreferences{}

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.ProfileVisibility,
		update.RecipeVisibility,
		update.ActivityVisibility,
		update.ContactInfoVisibility,
		update.DataSharing,
		update.AnalyticsTracking,
	).Scan(
		&prefs.ProfileVisibility,
		&prefs.RecipeVisibility,
		&prefs.ActivityVisibility,
		&prefs.ContactInfoVisibility,
		&prefs.DataSharing,
		&prefs.AnalyticsTracking,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update privacy preferences: %w", err)
	}

	return prefs, nil
}

// GetAccessibilityPreferences retrieves accessibility preferences for a user.
func (r *SQLPreferenceRepository) GetAccessibilityPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.AccessibilityPreferences, error) {
	query := `
		SELECT screen_reader, high_contrast, reduced_motion, large_text, keyboard_navigation, updated_at
		FROM recipe_manager.user_accessibility_preferences
		WHERE user_id = $1
	`

	prefs := &dto.AccessibilityPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.ScreenReader,
		&prefs.HighContrast,
		&prefs.ReducedMotion,
		&prefs.LargeText,
		&prefs.KeyboardNavigation,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultAccessibilityPreferences(), nil
		}

		return nil, fmt.Errorf("failed to get accessibility preferences: %w", err)
	}

	return prefs, nil
}

func defaultAccessibilityPreferences() *dto.AccessibilityPreferences {
	return &dto.AccessibilityPreferences{
		ScreenReader:       false,
		HighContrast:       false,
		ReducedMotion:      false,
		LargeText:          false,
		KeyboardNavigation: false,
		UpdatedAt:          time.Now(),
	}
}

// UpdateAccessibilityPreferences updates accessibility preferences using upsert.
func (r *SQLPreferenceRepository) UpdateAccessibilityPreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.AccessibilityPreferencesUpdate,
) (*dto.AccessibilityPreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_accessibility_preferences (
			user_id, screen_reader, high_contrast, reduced_motion, large_text, keyboard_navigation, updated_at
		)
		VALUES ($1,
			COALESCE($2, false), COALESCE($3, false), COALESCE($4, false),
			COALESCE($5, false), COALESCE($6, false), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			screen_reader = COALESCE($2, user_accessibility_preferences.screen_reader),
			high_contrast = COALESCE($3, user_accessibility_preferences.high_contrast),
			reduced_motion = COALESCE($4, user_accessibility_preferences.reduced_motion),
			large_text = COALESCE($5, user_accessibility_preferences.large_text),
			keyboard_navigation = COALESCE($6, user_accessibility_preferences.keyboard_navigation),
			updated_at = NOW()
		RETURNING screen_reader, high_contrast, reduced_motion, large_text, keyboard_navigation, updated_at
	`

	prefs := &dto.AccessibilityPreferences{}

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.ScreenReader,
		update.HighContrast,
		update.ReducedMotion,
		update.LargeText,
		update.KeyboardNavigation,
	).Scan(
		&prefs.ScreenReader,
		&prefs.HighContrast,
		&prefs.ReducedMotion,
		&prefs.LargeText,
		&prefs.KeyboardNavigation,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update accessibility preferences: %w", err)
	}

	return prefs, nil
}

// GetLanguagePreferences retrieves language preferences for a user.
func (r *SQLPreferenceRepository) GetLanguagePreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.LanguagePreferences, error) {
	query := `
		SELECT primary_language, secondary_language, translation_enabled, updated_at
		FROM recipe_manager.user_language_preferences
		WHERE user_id = $1
	`

	prefs := &dto.LanguagePreferences{}

	var secondaryLang sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.PrimaryLanguage,
		&secondaryLang,
		&prefs.TranslationEnabled,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultLanguagePreferences(), nil
		}

		return nil, fmt.Errorf("failed to get language preferences: %w", err)
	}

	if secondaryLang.Valid {
		lang := dto.Language(secondaryLang.String)
		prefs.SecondaryLanguage = &lang
	}

	return prefs, nil
}

func defaultLanguagePreferences() *dto.LanguagePreferences {
	return &dto.LanguagePreferences{
		PrimaryLanguage:    dto.LanguageEN,
		SecondaryLanguage:  nil,
		TranslationEnabled: false,
		UpdatedAt:          time.Now(),
	}
}

// UpdateLanguagePreferences updates language preferences using upsert.
func (r *SQLPreferenceRepository) UpdateLanguagePreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.LanguagePreferencesUpdate,
) (*dto.LanguagePreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_language_preferences (
			user_id, primary_language, secondary_language, translation_enabled, updated_at
		)
		VALUES ($1, COALESCE($2, 'EN'), $3, COALESCE($4, false), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			primary_language = COALESCE($2, user_language_preferences.primary_language),
			secondary_language = COALESCE($3, user_language_preferences.secondary_language),
			translation_enabled = COALESCE($4, user_language_preferences.translation_enabled),
			updated_at = NOW()
		RETURNING primary_language, secondary_language, translation_enabled, updated_at
	`

	prefs := &dto.LanguagePreferences{}

	var secondaryLang sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.PrimaryLanguage,
		update.SecondaryLanguage,
		update.TranslationEnabled,
	).Scan(
		&prefs.PrimaryLanguage,
		&secondaryLang,
		&prefs.TranslationEnabled,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update language preferences: %w", err)
	}

	if secondaryLang.Valid {
		lang := dto.Language(secondaryLang.String)
		prefs.SecondaryLanguage = &lang
	}

	return prefs, nil
}

// GetSecurityPreferences retrieves security preferences for a user.
func (r *SQLPreferenceRepository) GetSecurityPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.SecurityPreferences, error) {
	query := `
		SELECT two_factor_auth, login_notifications, session_timeout, password_requirements, updated_at
		FROM recipe_manager.user_security_preferences
		WHERE user_id = $1
	`

	prefs := &dto.SecurityPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.TwoFactorAuth,
		&prefs.LoginNotifications,
		&prefs.SessionTimeout,
		&prefs.PasswordRequirements,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultSecurityPreferences(), nil
		}

		return nil, fmt.Errorf("failed to get security preferences: %w", err)
	}

	return prefs, nil
}

func defaultSecurityPreferences() *dto.SecurityPreferences {
	return &dto.SecurityPreferences{
		TwoFactorAuth:        false,
		LoginNotifications:   true,
		SessionTimeout:       false,
		PasswordRequirements: true,
		UpdatedAt:            time.Now(),
	}
}

// UpdateSecurityPreferences updates security preferences using upsert.
func (r *SQLPreferenceRepository) UpdateSecurityPreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.SecurityPreferencesUpdate,
) (*dto.SecurityPreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_security_preferences (
			user_id, two_factor_auth, login_notifications, session_timeout, password_requirements, updated_at
		)
		VALUES ($1,
			COALESCE($2, false), COALESCE($3, true), COALESCE($4, false), COALESCE($5, true), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			two_factor_auth = COALESCE($2, user_security_preferences.two_factor_auth),
			login_notifications = COALESCE($3, user_security_preferences.login_notifications),
			session_timeout = COALESCE($4, user_security_preferences.session_timeout),
			password_requirements = COALESCE($5, user_security_preferences.password_requirements),
			updated_at = NOW()
		RETURNING two_factor_auth, login_notifications, session_timeout, password_requirements, updated_at
	`

	prefs := &dto.SecurityPreferences{}

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.TwoFactorAuth,
		update.LoginNotifications,
		update.SessionTimeout,
		update.PasswordRequirements,
	).Scan(
		&prefs.TwoFactorAuth,
		&prefs.LoginNotifications,
		&prefs.SessionTimeout,
		&prefs.PasswordRequirements,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update security preferences: %w", err)
	}

	return prefs, nil
}

// GetSocialPreferences retrieves social preferences for a user.
func (r *SQLPreferenceRepository) GetSocialPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.SocialPreferences, error) {
	query := `
		SELECT friend_requests, message_notifications, group_invites, share_activity, updated_at
		FROM recipe_manager.user_social_preferences
		WHERE user_id = $1
	`

	prefs := &dto.SocialPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.FriendRequests,
		&prefs.MessageNotifications,
		&prefs.GroupInvites,
		&prefs.ShareActivity,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultSocialPreferences(), nil
		}

		return nil, fmt.Errorf("failed to get social preferences: %w", err)
	}

	return prefs, nil
}

func defaultSocialPreferences() *dto.SocialPreferences {
	return &dto.SocialPreferences{
		FriendRequests:       true,
		MessageNotifications: true,
		GroupInvites:         true,
		ShareActivity:        true,
		UpdatedAt:            time.Now(),
	}
}

// UpdateSocialPreferences updates social preferences using upsert.
func (r *SQLPreferenceRepository) UpdateSocialPreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.SocialPreferencesUpdate,
) (*dto.SocialPreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_social_preferences (
			user_id, friend_requests, message_notifications, group_invites, share_activity, updated_at
		)
		VALUES ($1,
			COALESCE($2, true), COALESCE($3, true), COALESCE($4, true), COALESCE($5, true), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			friend_requests = COALESCE($2, user_social_preferences.friend_requests),
			message_notifications = COALESCE($3, user_social_preferences.message_notifications),
			group_invites = COALESCE($4, user_social_preferences.group_invites),
			share_activity = COALESCE($5, user_social_preferences.share_activity),
			updated_at = NOW()
		RETURNING friend_requests, message_notifications, group_invites, share_activity, updated_at
	`

	prefs := &dto.SocialPreferences{}

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.FriendRequests,
		update.MessageNotifications,
		update.GroupInvites,
		update.ShareActivity,
	).Scan(
		&prefs.FriendRequests,
		&prefs.MessageNotifications,
		&prefs.GroupInvites,
		&prefs.ShareActivity,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update social preferences: %w", err)
	}

	return prefs, nil
}

// GetSoundPreferences retrieves sound preferences for a user.
func (r *SQLPreferenceRepository) GetSoundPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.SoundPreferences, error) {
	query := `
		SELECT notification_sounds, system_sounds, volume_level, mute_notifications, updated_at
		FROM recipe_manager.user_sound_preferences
		WHERE user_id = $1
	`

	prefs := &dto.SoundPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.NotificationSounds,
		&prefs.SystemSounds,
		&prefs.VolumeLevel,
		&prefs.MuteNotifications,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultSoundPreferences(), nil
		}

		return nil, fmt.Errorf("failed to get sound preferences: %w", err)
	}

	return prefs, nil
}

func defaultSoundPreferences() *dto.SoundPreferences {
	return &dto.SoundPreferences{
		NotificationSounds: true,
		SystemSounds:       true,
		VolumeLevel:        dto.VolumeLevelMedium,
		MuteNotifications:  false,
		UpdatedAt:          time.Now(),
	}
}

// UpdateSoundPreferences updates sound preferences using upsert.
func (r *SQLPreferenceRepository) UpdateSoundPreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.SoundPreferencesUpdate,
) (*dto.SoundPreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_sound_preferences (
			user_id, notification_sounds, system_sounds, volume_level, mute_notifications, updated_at
		)
		VALUES ($1,
			COALESCE($2, true), COALESCE($3, true), COALESCE($4, 'MEDIUM'), COALESCE($5, false), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			notification_sounds = COALESCE($2, user_sound_preferences.notification_sounds),
			system_sounds = COALESCE($3, user_sound_preferences.system_sounds),
			volume_level = COALESCE($4, user_sound_preferences.volume_level),
			mute_notifications = COALESCE($5, user_sound_preferences.mute_notifications),
			updated_at = NOW()
		RETURNING notification_sounds, system_sounds, volume_level, mute_notifications, updated_at
	`

	prefs := &dto.SoundPreferences{}

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.NotificationSounds,
		update.SystemSounds,
		update.VolumeLevel,
		update.MuteNotifications,
	).Scan(
		&prefs.NotificationSounds,
		&prefs.SystemSounds,
		&prefs.VolumeLevel,
		&prefs.MuteNotifications,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update sound preferences: %w", err)
	}

	return prefs, nil
}

// GetThemePreferences retrieves theme preferences for a user.
func (r *SQLPreferenceRepository) GetThemePreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.ThemePreferences, error) {
	query := `
		SELECT dark_mode, light_mode, auto_theme, custom_theme, updated_at
		FROM recipe_manager.user_theme_preferences
		WHERE user_id = $1
	`

	prefs := &dto.ThemePreferences{}

	var customTheme sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.DarkMode,
		&prefs.LightMode,
		&prefs.AutoTheme,
		&customTheme,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultThemePreferences(), nil
		}

		return nil, fmt.Errorf("failed to get theme preferences: %w", err)
	}

	if customTheme.Valid {
		theme := dto.Theme(customTheme.String)
		prefs.CustomTheme = &theme
	}

	return prefs, nil
}

func defaultThemePreferences() *dto.ThemePreferences {
	return &dto.ThemePreferences{
		DarkMode:    false,
		LightMode:   true,
		AutoTheme:   false,
		CustomTheme: nil,
		UpdatedAt:   time.Now(),
	}
}

// UpdateThemePreferences updates theme preferences using upsert.
func (r *SQLPreferenceRepository) UpdateThemePreferences(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.ThemePreferencesUpdate,
) (*dto.ThemePreferences, error) {
	query := `
		INSERT INTO recipe_manager.user_theme_preferences (
			user_id, dark_mode, light_mode, auto_theme, custom_theme, updated_at
		)
		VALUES ($1, COALESCE($2, false), COALESCE($3, true), COALESCE($4, false), $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			dark_mode = COALESCE($2, user_theme_preferences.dark_mode),
			light_mode = COALESCE($3, user_theme_preferences.light_mode),
			auto_theme = COALESCE($4, user_theme_preferences.auto_theme),
			custom_theme = COALESCE($5, user_theme_preferences.custom_theme),
			updated_at = NOW()
		RETURNING dark_mode, light_mode, auto_theme, custom_theme, updated_at
	`

	prefs := &dto.ThemePreferences{}

	var customTheme sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		userID,
		update.DarkMode,
		update.LightMode,
		update.AutoTheme,
		update.CustomTheme,
	).Scan(
		&prefs.DarkMode,
		&prefs.LightMode,
		&prefs.AutoTheme,
		&customTheme,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update theme preferences: %w", err)
	}

	if customTheme.Valid {
		theme := dto.Theme(customTheme.String)
		prefs.CustomTheme = &theme
	}

	return prefs, nil
}
