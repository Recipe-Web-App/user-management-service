package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

// Service-level errors for preferences.
var (
	ErrUnauthorizedAccess = errors.New("unauthorized access to preferences")
	ErrInvalidCategory    = errors.New("invalid preference category")
	ErrInvalidUpdateType  = errors.New("invalid update type for preference category")
)

// PreferenceService defines business logic for preference operations.
type PreferenceService interface {
	// GetAllPreferences retrieves all or filtered preferences for a user.
	GetAllPreferences(
		ctx context.Context,
		requesterID, targetUserID uuid.UUID,
		categories []dto.PreferenceCategory,
		isAdmin bool,
		hasServiceScope bool,
	) (*dto.UserPreferencesResponse, error)

	// GetCategoryPreferences retrieves a single preference category.
	GetCategoryPreferences(
		ctx context.Context,
		requesterID, targetUserID uuid.UUID,
		category dto.PreferenceCategory,
		isAdmin bool,
		hasServiceScope bool,
	) (*dto.PreferenceCategoryResponse, error)

	// UpdateAllPreferences updates multiple preference categories.
	UpdateAllPreferences(
		ctx context.Context,
		requesterID, targetUserID uuid.UUID,
		update *dto.UserPreferencesUpdateRequest,
		isAdmin bool,
		hasServiceScope bool,
	) (*dto.UserPreferencesResponse, error)

	// UpdateCategoryPreferences updates a single preference category.
	UpdateCategoryPreferences(
		ctx context.Context,
		requesterID, targetUserID uuid.UUID,
		category dto.PreferenceCategory,
		update any,
		isAdmin bool,
		hasServiceScope bool,
	) (*dto.PreferenceCategoryResponse, error)
}

// PreferenceServiceImpl implements PreferenceService.
type PreferenceServiceImpl struct {
	repo repository.PreferenceRepository
}

// NewPreferenceService creates a new PreferenceService.
func NewPreferenceService(repo repository.PreferenceRepository) *PreferenceServiceImpl {
	return &PreferenceServiceImpl{repo: repo}
}

// GetAllPreferences retrieves all or filtered preferences for a user.
func (s *PreferenceServiceImpl) GetAllPreferences(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	categories []dto.PreferenceCategory,
	isAdmin bool,
	hasServiceScope bool,
) (*dto.UserPreferencesResponse, error) {
	if !s.canAccessPreferences(requesterID, targetUserID, isAdmin, hasServiceScope) {
		return nil, ErrUnauthorizedAccess
	}

	exists, err := s.repo.UserExists(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	if !exists {
		return nil, ErrUserNotFound
	}

	categoriesToFetch := dto.ValidPreferenceCategories
	if len(categories) > 0 {
		categoriesToFetch = categories
	}

	response := &dto.UserPreferencesResponse{UserID: targetUserID.String()}

	for _, category := range categoriesToFetch {
		err = s.fetchCategory(ctx, targetUserID, category, response)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

// GetCategoryPreferences retrieves a single preference category.
func (s *PreferenceServiceImpl) GetCategoryPreferences(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	category dto.PreferenceCategory,
	isAdmin bool,
	hasServiceScope bool,
) (*dto.PreferenceCategoryResponse, error) {
	if !s.canAccessPreferences(requesterID, targetUserID, isAdmin, hasServiceScope) {
		return nil, ErrUnauthorizedAccess
	}

	exists, err := s.repo.UserExists(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	if !exists {
		return nil, ErrUserNotFound
	}

	if !dto.IsValidPreferenceCategory(string(category)) {
		return nil, ErrInvalidCategory
	}

	prefs, updatedAt, err := s.fetchSingleCategory(ctx, targetUserID, category)
	if err != nil {
		return nil, err
	}

	return &dto.PreferenceCategoryResponse{
		UserID:      targetUserID.String(),
		Category:    string(category),
		Preferences: prefs,
		UpdatedAt:   updatedAt,
	}, nil
}

// UpdateAllPreferences updates multiple preference categories.
//
//nolint:cyclop,funlen // Updating 9 categories sequentially is inherent to domain design.
func (s *PreferenceServiceImpl) UpdateAllPreferences(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	update *dto.UserPreferencesUpdateRequest,
	isAdmin bool,
	hasServiceScope bool,
) (*dto.UserPreferencesResponse, error) {
	if !s.canAccessPreferences(requesterID, targetUserID, isAdmin, hasServiceScope) {
		return nil, ErrUnauthorizedAccess
	}

	exists, err := s.repo.UserExists(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	if !exists {
		return nil, ErrUserNotFound
	}

	response := &dto.UserPreferencesResponse{UserID: targetUserID.String()}

	err = s.updateNotificationIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updateDisplayIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updatePrivacyIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updateAccessibilityIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updateLanguageIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updateSecurityIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updateSocialIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updateSoundIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	err = s.updateThemeIfPresent(ctx, targetUserID, update, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// UpdateCategoryPreferences updates a single preference category.
func (s *PreferenceServiceImpl) UpdateCategoryPreferences(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	category dto.PreferenceCategory,
	update any,
	isAdmin bool,
	hasServiceScope bool,
) (*dto.PreferenceCategoryResponse, error) {
	if !s.canAccessPreferences(requesterID, targetUserID, isAdmin, hasServiceScope) {
		return nil, ErrUnauthorizedAccess
	}

	exists, err := s.repo.UserExists(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	if !exists {
		return nil, ErrUserNotFound
	}

	if !dto.IsValidPreferenceCategory(string(category)) {
		return nil, ErrInvalidCategory
	}

	prefs, updatedAt, err := s.updateSingleCategory(ctx, targetUserID, category, update)
	if err != nil {
		return nil, err
	}

	return &dto.PreferenceCategoryResponse{
		UserID:      targetUserID.String(),
		Category:    string(category),
		Preferences: prefs,
		UpdatedAt:   updatedAt,
	}, nil
}

// --- Private methods below ---

func (s *PreferenceServiceImpl) canAccessPreferences(
	requesterID, targetUserID uuid.UUID,
	isAdmin bool,
	hasServiceScope bool,
) bool {
	if requesterID == targetUserID {
		return true
	}

	if isAdmin {
		return true
	}

	if hasServiceScope {
		return true
	}

	return false
}

//nolint:cyclop // Switch over 9 categories is inherent to domain design.
func (s *PreferenceServiceImpl) fetchCategory(
	ctx context.Context,
	userID uuid.UUID,
	category dto.PreferenceCategory,
	response *dto.UserPreferencesResponse,
) error {
	var err error

	switch category {
	case dto.PreferenceCategoryNotification:
		response.Notification, err = s.repo.GetNotificationPreferences(ctx, userID)
	case dto.PreferenceCategoryDisplay:
		response.Display, err = s.repo.GetDisplayPreferences(ctx, userID)
	case dto.PreferenceCategoryPrivacy:
		response.Privacy, err = s.repo.GetPrivacyPreferencesData(ctx, userID)
	case dto.PreferenceCategoryAccessibility:
		response.Accessibility, err = s.repo.GetAccessibilityPreferences(ctx, userID)
	case dto.PreferenceCategoryLanguage:
		response.Language, err = s.repo.GetLanguagePreferences(ctx, userID)
	case dto.PreferenceCategorySecurity:
		response.Security, err = s.repo.GetSecurityPreferences(ctx, userID)
	case dto.PreferenceCategorySocial:
		response.Social, err = s.repo.GetSocialPreferences(ctx, userID)
	case dto.PreferenceCategorySound:
		response.Sound, err = s.repo.GetSoundPreferences(ctx, userID)
	case dto.PreferenceCategoryTheme:
		response.Theme, err = s.repo.GetThemePreferences(ctx, userID)
	default:
		return ErrInvalidCategory
	}

	if err != nil {
		return fmt.Errorf("failed to fetch %s preferences: %w", category, err)
	}

	return nil
}

//nolint:cyclop // Switch over 9 categories is inherent to domain design.
func (s *PreferenceServiceImpl) fetchSingleCategory(
	ctx context.Context,
	userID uuid.UUID,
	category dto.PreferenceCategory,
) (any, time.Time, error) {
	var (
		prefs     any
		updatedAt time.Time
		err       error
	)

	switch category {
	case dto.PreferenceCategoryNotification:
		p, e := s.repo.GetNotificationPreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryDisplay:
		p, e := s.repo.GetDisplayPreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryPrivacy:
		p, e := s.repo.GetPrivacyPreferencesData(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryAccessibility:
		p, e := s.repo.GetAccessibilityPreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryLanguage:
		p, e := s.repo.GetLanguagePreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategorySecurity:
		p, e := s.repo.GetSecurityPreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategorySocial:
		p, e := s.repo.GetSocialPreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategorySound:
		p, e := s.repo.GetSoundPreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryTheme:
		p, e := s.repo.GetThemePreferences(ctx, userID)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	default:
		return nil, time.Time{}, ErrInvalidCategory
	}

	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to fetch %s preferences: %w", category, err)
	}

	return prefs, updatedAt, nil
}

//nolint:cyclop,funlen // Switch over 9 categories with type assertions is inherent to domain design.
func (s *PreferenceServiceImpl) updateSingleCategory(
	ctx context.Context,
	userID uuid.UUID,
	category dto.PreferenceCategory,
	update any,
) (any, time.Time, error) {
	var (
		prefs     any
		updatedAt time.Time
		err       error
	)

	switch category {
	case dto.PreferenceCategoryNotification:
		u, ok := update.(*dto.NotificationPreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateNotificationPreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryDisplay:
		u, ok := update.(*dto.DisplayPreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateDisplayPreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryPrivacy:
		u, ok := update.(*dto.PrivacyPreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdatePrivacyPreferencesData(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryAccessibility:
		u, ok := update.(*dto.AccessibilityPreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateAccessibilityPreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryLanguage:
		u, ok := update.(*dto.LanguagePreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateLanguagePreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategorySecurity:
		u, ok := update.(*dto.SecurityPreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateSecurityPreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategorySocial:
		u, ok := update.(*dto.SocialPreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateSocialPreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategorySound:
		u, ok := update.(*dto.SoundPreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateSoundPreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	case dto.PreferenceCategoryTheme:
		u, ok := update.(*dto.ThemePreferencesUpdate)
		if !ok {
			return nil, time.Time{}, ErrInvalidUpdateType
		}

		p, e := s.repo.UpdateThemePreferences(ctx, userID, u)
		prefs, updatedAt, err = p, p.UpdatedAt, e
	default:
		return nil, time.Time{}, ErrInvalidCategory
	}

	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to update %s preferences: %w", category, err)
	}

	return prefs, updatedAt, nil
}

func (s *PreferenceServiceImpl) updateNotificationIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Notification == nil {
		return nil
	}

	prefs, err := s.repo.UpdateNotificationPreferences(ctx, userID, update.Notification)
	if err != nil {
		return fmt.Errorf("failed to update notification preferences: %w", err)
	}

	response.Notification = prefs

	return nil
}

func (s *PreferenceServiceImpl) updateDisplayIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Display == nil {
		return nil
	}

	prefs, err := s.repo.UpdateDisplayPreferences(ctx, userID, update.Display)
	if err != nil {
		return fmt.Errorf("failed to update display preferences: %w", err)
	}

	response.Display = prefs

	return nil
}

func (s *PreferenceServiceImpl) updatePrivacyIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Privacy == nil {
		return nil
	}

	prefs, err := s.repo.UpdatePrivacyPreferencesData(ctx, userID, update.Privacy)
	if err != nil {
		return fmt.Errorf("failed to update privacy preferences: %w", err)
	}

	response.Privacy = prefs

	return nil
}

func (s *PreferenceServiceImpl) updateAccessibilityIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Accessibility == nil {
		return nil
	}

	prefs, err := s.repo.UpdateAccessibilityPreferences(ctx, userID, update.Accessibility)
	if err != nil {
		return fmt.Errorf("failed to update accessibility preferences: %w", err)
	}

	response.Accessibility = prefs

	return nil
}

func (s *PreferenceServiceImpl) updateLanguageIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Language == nil {
		return nil
	}

	prefs, err := s.repo.UpdateLanguagePreferences(ctx, userID, update.Language)
	if err != nil {
		return fmt.Errorf("failed to update language preferences: %w", err)
	}

	response.Language = prefs

	return nil
}

func (s *PreferenceServiceImpl) updateSecurityIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Security == nil {
		return nil
	}

	prefs, err := s.repo.UpdateSecurityPreferences(ctx, userID, update.Security)
	if err != nil {
		return fmt.Errorf("failed to update security preferences: %w", err)
	}

	response.Security = prefs

	return nil
}

func (s *PreferenceServiceImpl) updateSocialIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Social == nil {
		return nil
	}

	prefs, err := s.repo.UpdateSocialPreferences(ctx, userID, update.Social)
	if err != nil {
		return fmt.Errorf("failed to update social preferences: %w", err)
	}

	response.Social = prefs

	return nil
}

func (s *PreferenceServiceImpl) updateSoundIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Sound == nil {
		return nil
	}

	prefs, err := s.repo.UpdateSoundPreferences(ctx, userID, update.Sound)
	if err != nil {
		return fmt.Errorf("failed to update sound preferences: %w", err)
	}

	response.Sound = prefs

	return nil
}

func (s *PreferenceServiceImpl) updateThemeIfPresent(
	ctx context.Context, userID uuid.UUID, update *dto.UserPreferencesUpdateRequest,
	response *dto.UserPreferencesResponse,
) error {
	if update.Theme == nil {
		return nil
	}

	prefs, err := s.repo.UpdateThemePreferences(ctx, userID, update.Theme)
	if err != nil {
		return fmt.Errorf("failed to update theme preferences: %w", err)
	}

	response.Theme = prefs

	return nil
}
