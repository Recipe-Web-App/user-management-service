package dto

import "time"

// PreferenceCategory represents valid preference category names.
type PreferenceCategory string

const (
	PreferenceCategoryNotification  PreferenceCategory = "notification"
	PreferenceCategoryDisplay       PreferenceCategory = "display"
	PreferenceCategoryPrivacy       PreferenceCategory = "privacy"
	PreferenceCategoryAccessibility PreferenceCategory = "accessibility"
	PreferenceCategoryLanguage      PreferenceCategory = "language"
	PreferenceCategorySecurity      PreferenceCategory = "security"
	PreferenceCategorySocial        PreferenceCategory = "social"
	PreferenceCategorySound         PreferenceCategory = "sound"
	PreferenceCategoryTheme         PreferenceCategory = "theme"
)

// ValidPreferenceCategories lists all valid category names.
var ValidPreferenceCategories = []PreferenceCategory{
	PreferenceCategoryNotification,
	PreferenceCategoryDisplay,
	PreferenceCategoryPrivacy,
	PreferenceCategoryAccessibility,
	PreferenceCategoryLanguage,
	PreferenceCategorySecurity,
	PreferenceCategorySocial,
	PreferenceCategorySound,
	PreferenceCategoryTheme,
}

// IsValidPreferenceCategory checks if a category string is valid.
func IsValidPreferenceCategory(cat string) bool {
	for _, valid := range ValidPreferenceCategories {
		if string(valid) == cat {
			return true
		}
	}

	return false
}

// FontSize represents font size preference values.
type FontSize string

const (
	FontSizeSmall      FontSize = "SMALL"
	FontSizeMedium     FontSize = "MEDIUM"
	FontSizeLarge      FontSize = "LARGE"
	FontSizeExtraLarge FontSize = "EXTRA_LARGE"
)

// ColorScheme represents color scheme preference values.
type ColorScheme string

const (
	ColorSchemeLight        ColorScheme = "LIGHT"
	ColorSchemeDark         ColorScheme = "DARK"
	ColorSchemeAuto         ColorScheme = "AUTO"
	ColorSchemeHighContrast ColorScheme = "HIGH_CONTRAST"
)

// LayoutDensity represents layout density preference values.
type LayoutDensity string

const (
	LayoutDensityCompact     LayoutDensity = "COMPACT"
	LayoutDensityComfortable LayoutDensity = "COMFORTABLE"
	LayoutDensitySpacious    LayoutDensity = "SPACIOUS"
)

// ProfileVisibility represents profile visibility preference values.
type ProfileVisibility string

const (
	ProfileVisibilityPublic      ProfileVisibility = "PUBLIC"
	ProfileVisibilityFriendsOnly ProfileVisibility = "FRIENDS_ONLY"
	ProfileVisibilityPrivate     ProfileVisibility = "PRIVATE"
)

// Language represents language preference values.
type Language string

const (
	LanguageEN Language = "EN"
	LanguageES Language = "ES"
	LanguageFR Language = "FR"
	LanguageDE Language = "DE"
	LanguageIT Language = "IT"
	LanguagePT Language = "PT"
	LanguageZH Language = "ZH"
	LanguageJA Language = "JA"
	LanguageKO Language = "KO"
	LanguageRU Language = "RU"
)

// Theme represents theme preference values.
type Theme string

const (
	ThemeLight  Theme = "LIGHT"
	ThemeDark   Theme = "DARK"
	ThemeAuto   Theme = "AUTO"
	ThemeCustom Theme = "CUSTOM"
)

// VolumeLevel represents volume level preference values.
type VolumeLevel string

const (
	VolumeLevelMuted  VolumeLevel = "MUTED"
	VolumeLevelLow    VolumeLevel = "LOW"
	VolumeLevelMedium VolumeLevel = "MEDIUM"
	VolumeLevelHigh   VolumeLevel = "HIGH"
)

// NotificationPreferences represents notification preference settings.
type NotificationPreferences struct {
	EmailNotifications    bool      `json:"emailNotifications"`
	PushNotifications     bool      `json:"pushNotifications"`
	SMSNotifications      bool      `json:"smsNotifications"`
	MarketingEmails       bool      `json:"marketingEmails"`
	SecurityAlerts        bool      `json:"securityAlerts"`
	ActivitySummaries     bool      `json:"activitySummaries"`
	RecipeRecommendations bool      `json:"recipeRecommendations"`
	SocialInteractions    bool      `json:"socialInteractions"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

// DisplayPreferences represents display preference settings.
type DisplayPreferences struct {
	FontSize      FontSize      `json:"fontSize"`
	ColorScheme   ColorScheme   `json:"colorScheme"`
	LayoutDensity LayoutDensity `json:"layoutDensity"`
	ShowImages    bool          `json:"showImages"`
	CompactMode   bool          `json:"compactMode"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

// UserPrivacyPreferences represents the full privacy preference settings from the database.
type UserPrivacyPreferences struct {
	ProfileVisibility     ProfileVisibility `json:"profileVisibility"`
	RecipeVisibility      ProfileVisibility `json:"recipeVisibility"`
	ActivityVisibility    ProfileVisibility `json:"activityVisibility"`
	ContactInfoVisibility ProfileVisibility `json:"contactInfoVisibility"`
	DataSharing           bool              `json:"dataSharing"`
	AnalyticsTracking     bool              `json:"analyticsTracking"`
	UpdatedAt             time.Time         `json:"updatedAt"`
}

// AccessibilityPreferences represents accessibility preference settings.
type AccessibilityPreferences struct {
	ScreenReader       bool      `json:"screenReader"`
	HighContrast       bool      `json:"highContrast"`
	ReducedMotion      bool      `json:"reducedMotion"`
	LargeText          bool      `json:"largeText"`
	KeyboardNavigation bool      `json:"keyboardNavigation"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// LanguagePreferences represents language preference settings.
type LanguagePreferences struct {
	PrimaryLanguage    Language  `json:"primaryLanguage"`
	SecondaryLanguage  *Language `json:"secondaryLanguage,omitempty"`
	TranslationEnabled bool      `json:"translationEnabled"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// SecurityPreferences represents security preference settings.
type SecurityPreferences struct {
	TwoFactorAuth        bool      `json:"twoFactorAuth"`
	LoginNotifications   bool      `json:"loginNotifications"`
	SessionTimeout       bool      `json:"sessionTimeout"`
	PasswordRequirements bool      `json:"passwordRequirements"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// SocialPreferences represents social preference settings.
type SocialPreferences struct {
	FriendRequests       bool      `json:"friendRequests"`
	MessageNotifications bool      `json:"messageNotifications"`
	GroupInvites         bool      `json:"groupInvites"`
	ShareActivity        bool      `json:"shareActivity"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// SoundPreferences represents sound preference settings.
type SoundPreferences struct {
	NotificationSounds bool        `json:"notificationSounds"`
	SystemSounds       bool        `json:"systemSounds"`
	VolumeLevel        VolumeLevel `json:"volumeLevel"`
	MuteNotifications  bool        `json:"muteNotifications"`
	UpdatedAt          time.Time   `json:"updatedAt"`
}

// ThemePreferences represents theme preference settings.
type ThemePreferences struct {
	DarkMode    bool      `json:"darkMode"`
	LightMode   bool      `json:"lightMode"`
	AutoTheme   bool      `json:"autoTheme"`
	CustomTheme *Theme    `json:"customTheme,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NotificationPreferencesUpdate represents update request for notification preferences.
type NotificationPreferencesUpdate struct {
	EmailNotifications    *bool `json:"emailNotifications,omitempty"`
	PushNotifications     *bool `json:"pushNotifications,omitempty"`
	SMSNotifications      *bool `json:"smsNotifications,omitempty"`
	MarketingEmails       *bool `json:"marketingEmails,omitempty"`
	SecurityAlerts        *bool `json:"securityAlerts,omitempty"`
	ActivitySummaries     *bool `json:"activitySummaries,omitempty"`
	RecipeRecommendations *bool `json:"recipeRecommendations,omitempty"`
	SocialInteractions    *bool `json:"socialInteractions,omitempty"`
}

// DisplayPreferencesUpdate represents update request for display preferences.
type DisplayPreferencesUpdate struct {
	FontSize      *FontSize      `json:"fontSize,omitempty"`
	ColorScheme   *ColorScheme   `json:"colorScheme,omitempty"`
	LayoutDensity *LayoutDensity `json:"layoutDensity,omitempty"`
	ShowImages    *bool          `json:"showImages,omitempty"`
	CompactMode   *bool          `json:"compactMode,omitempty"`
}

// PrivacyPreferencesUpdate represents update request for privacy preferences.
type PrivacyPreferencesUpdate struct {
	ProfileVisibility     *ProfileVisibility `json:"profileVisibility,omitempty"`
	RecipeVisibility      *ProfileVisibility `json:"recipeVisibility,omitempty"`
	ActivityVisibility    *ProfileVisibility `json:"activityVisibility,omitempty"`
	ContactInfoVisibility *ProfileVisibility `json:"contactInfoVisibility,omitempty"`
	DataSharing           *bool              `json:"dataSharing,omitempty"`
	AnalyticsTracking     *bool              `json:"analyticsTracking,omitempty"`
}

// AccessibilityPreferencesUpdate represents update request for accessibility preferences.
type AccessibilityPreferencesUpdate struct {
	ScreenReader       *bool `json:"screenReader,omitempty"`
	HighContrast       *bool `json:"highContrast,omitempty"`
	ReducedMotion      *bool `json:"reducedMotion,omitempty"`
	LargeText          *bool `json:"largeText,omitempty"`
	KeyboardNavigation *bool `json:"keyboardNavigation,omitempty"`
}

// LanguagePreferencesUpdate represents update request for language preferences.
type LanguagePreferencesUpdate struct {
	PrimaryLanguage    *Language `json:"primaryLanguage,omitempty"`
	SecondaryLanguage  *Language `json:"secondaryLanguage,omitempty"`
	TranslationEnabled *bool     `json:"translationEnabled,omitempty"`
}

// SecurityPreferencesUpdate represents update request for security preferences.
type SecurityPreferencesUpdate struct {
	TwoFactorAuth        *bool `json:"twoFactorAuth,omitempty"`
	LoginNotifications   *bool `json:"loginNotifications,omitempty"`
	SessionTimeout       *bool `json:"sessionTimeout,omitempty"`
	PasswordRequirements *bool `json:"passwordRequirements,omitempty"`
}

// SocialPreferencesUpdate represents update request for social preferences.
type SocialPreferencesUpdate struct {
	FriendRequests       *bool `json:"friendRequests,omitempty"`
	MessageNotifications *bool `json:"messageNotifications,omitempty"`
	GroupInvites         *bool `json:"groupInvites,omitempty"`
	ShareActivity        *bool `json:"shareActivity,omitempty"`
}

// SoundPreferencesUpdate represents update request for sound preferences.
type SoundPreferencesUpdate struct {
	NotificationSounds *bool        `json:"notificationSounds,omitempty"`
	SystemSounds       *bool        `json:"systemSounds,omitempty"`
	VolumeLevel        *VolumeLevel `json:"volumeLevel,omitempty"`
	MuteNotifications  *bool        `json:"muteNotifications,omitempty"`
}

// ThemePreferencesUpdate represents update request for theme preferences.
type ThemePreferencesUpdate struct {
	DarkMode    *bool  `json:"darkMode,omitempty"`
	LightMode   *bool  `json:"lightMode,omitempty"`
	AutoTheme   *bool  `json:"autoTheme,omitempty"`
	CustomTheme *Theme `json:"customTheme,omitempty"`
}

// UserPreferencesUpdateRequest represents a request to update multiple preference categories.
type UserPreferencesUpdateRequest struct {
	Notification  *NotificationPreferencesUpdate  `json:"notification,omitempty"`
	Display       *DisplayPreferencesUpdate       `json:"display,omitempty"`
	Privacy       *PrivacyPreferencesUpdate       `json:"privacy,omitempty"`
	Accessibility *AccessibilityPreferencesUpdate `json:"accessibility,omitempty"`
	Language      *LanguagePreferencesUpdate      `json:"language,omitempty"`
	Security      *SecurityPreferencesUpdate      `json:"security,omitempty"`
	Social        *SocialPreferencesUpdate        `json:"social,omitempty"`
	Sound         *SoundPreferencesUpdate         `json:"sound,omitempty"`
	Theme         *ThemePreferencesUpdate         `json:"theme,omitempty"`
}

// UserPreferencesResponse represents the response containing user preferences.
type UserPreferencesResponse struct {
	UserID        string                    `json:"userId"`
	Notification  *NotificationPreferences  `json:"notification,omitempty"`
	Display       *DisplayPreferences       `json:"display,omitempty"`
	Privacy       *UserPrivacyPreferences   `json:"privacy,omitempty"`
	Accessibility *AccessibilityPreferences `json:"accessibility,omitempty"`
	Language      *LanguagePreferences      `json:"language,omitempty"`
	Security      *SecurityPreferences      `json:"security,omitempty"`
	Social        *SocialPreferences        `json:"social,omitempty"`
	Sound         *SoundPreferences         `json:"sound,omitempty"`
	Theme         *ThemePreferences         `json:"theme,omitempty"`
}

// PreferenceCategoryResponse represents the response for a single preference category.
type PreferenceCategoryResponse struct {
	UserID      string    `json:"userId"`
	Category    string    `json:"category"`
	Preferences any       `json:"preferences"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
