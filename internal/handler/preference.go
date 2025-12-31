package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/middleware"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// Scope constants for preference access.
const (
	scopeUserRead  = "user:read"
	scopeUserWrite = "user:write"
	scopeAdmin     = "admin"
)

// PreferenceHandler handles preference-related HTTP endpoints.
type PreferenceHandler struct {
	preferenceService service.PreferenceService
	binder            *RequestBinder
}

// NewPreferenceHandler creates a new preference handler.
func NewPreferenceHandler(preferenceService service.PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{
		preferenceService: preferenceService,
		binder:            NewRequestBinder(),
	}
}

// GetAllPreferences handles GET /users/{user_id}/preferences.
func (h *PreferenceHandler) GetAllPreferences(w http.ResponseWriter, r *http.Request) {
	// 1. Extract target user ID from path
	targetUserID, ok := h.parseUserID(w, r)
	if !ok {
		return
	}

	// 2. Get authenticated user and authorization info
	requesterID, isAdmin, hasServiceScope, ok := h.extractAuthInfo(w, r)
	if !ok {
		return
	}

	// 3. Parse optional categories filter
	categories := h.parseCategoriesParam(r)

	// 4. Call service
	response, err := h.preferenceService.GetAllPreferences(
		r.Context(),
		requesterID,
		targetUserID,
		categories,
		isAdmin,
		hasServiceScope,
	)
	if err != nil {
		h.handleServiceError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// GetCategoryPreferences handles GET /users/{user_id}/preferences/{category}.
func (h *PreferenceHandler) GetCategoryPreferences(w http.ResponseWriter, r *http.Request) {
	// 1. Extract target user ID from path
	targetUserID, ok := h.parseUserID(w, r)
	if !ok {
		return
	}

	// 2. Extract and validate category from path
	category := chi.URLParam(r, "category")
	if !dto.IsValidPreferenceCategory(category) {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_CATEGORY", "Invalid preference category")

		return
	}

	// 3. Get authenticated user and authorization info
	requesterID, isAdmin, hasServiceScope, ok := h.extractAuthInfo(w, r)
	if !ok {
		return
	}

	// 4. Call service
	response, err := h.preferenceService.GetCategoryPreferences(
		r.Context(),
		requesterID,
		targetUserID,
		dto.PreferenceCategory(category),
		isAdmin,
		hasServiceScope,
	)
	if err != nil {
		h.handleServiceError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// UpdateAllPreferences handles PUT /users/{user_id}/preferences.
func (h *PreferenceHandler) UpdateAllPreferences(w http.ResponseWriter, r *http.Request) {
	// 1. Extract target user ID from path
	targetUserID, ok := h.parseUserID(w, r)
	if !ok {
		return
	}

	// 2. Get authenticated user and authorization info
	requesterID, isAdmin, hasServiceScope, ok := h.extractAuthInfo(w, r)
	if !ok {
		return
	}

	// 3. Bind and validate request body
	var req dto.UserPreferencesUpdateRequest

	bindErr := h.binder.BindAndValidate(r, &req)
	if bindErr != nil {
		h.handleBindError(w, bindErr)

		return
	}

	// 4. Call service
	response, err := h.preferenceService.UpdateAllPreferences(
		r.Context(),
		requesterID,
		targetUserID,
		&req,
		isAdmin,
		hasServiceScope,
	)
	if err != nil {
		h.handleServiceError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// UpdateCategoryPreferences handles PUT /users/{user_id}/preferences/{category}.
func (h *PreferenceHandler) UpdateCategoryPreferences(w http.ResponseWriter, r *http.Request) {
	// 1. Extract target user ID from path
	targetUserID, ok := h.parseUserID(w, r)
	if !ok {
		return
	}

	// 2. Extract and validate category from path
	category := chi.URLParam(r, "category")
	if !dto.IsValidPreferenceCategory(category) {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_CATEGORY", "Invalid preference category")

		return
	}

	// 3. Get authenticated user and authorization info
	requesterID, isAdmin, hasServiceScope, ok := h.extractAuthInfo(w, r)
	if !ok {
		return
	}

	// 4. Parse request body based on category
	update, err := h.parseUpdateRequest(r, category)
	if err != nil {
		h.handleBindError(w, err)

		return
	}

	// 5. Call service
	response, err := h.preferenceService.UpdateCategoryPreferences(
		r.Context(),
		requesterID,
		targetUserID,
		dto.PreferenceCategory(category),
		update,
		isAdmin,
		hasServiceScope,
	)
	if err != nil {
		h.handleServiceError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

func (h *PreferenceHandler) parseUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userIDStr := chi.URLParam(r, "user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")

		return uuid.Nil, false
	}

	return userID, true
}

func (h *PreferenceHandler) extractAuthInfo(
	w http.ResponseWriter,
	r *http.Request,
) (uuid.UUID, bool, bool, bool) {
	authUser, ok := middleware.GetAuthenticatedUser(r.Context())
	if !ok {
		UnauthorizedResponse(w, "Authentication required")

		return uuid.Nil, false, false, false
	}

	// For service accounts, use a nil UUID as requester ID
	requesterID := authUser.UserID

	// Check admin scope
	isAdmin := middleware.HasScope(r.Context(), scopeAdmin)

	// Check if service account with proper scopes
	hasServiceScope := authUser.IsService &&
		(middleware.HasScope(r.Context(), scopeUserRead) || middleware.HasScope(r.Context(), scopeUserWrite))

	return requesterID, isAdmin, hasServiceScope, true
}

func (h *PreferenceHandler) parseCategoriesParam(r *http.Request) []dto.PreferenceCategory {
	categoriesParam := r.URL.Query().Get("categories")
	if categoriesParam == "" {
		return nil
	}

	parts := strings.Split(categoriesParam, ",")
	categories := make([]dto.PreferenceCategory, 0, len(parts))

	for _, part := range parts {
		cat := strings.TrimSpace(part)
		if dto.IsValidPreferenceCategory(cat) {
			categories = append(categories, dto.PreferenceCategory(cat))
		}
	}

	return categories
}

//nolint:cyclop,funlen // Switch over 9 categories is inherent to domain design.
func (h *PreferenceHandler) parseUpdateRequest(r *http.Request, category string) (any, error) {
	switch dto.PreferenceCategory(category) {
	case dto.PreferenceCategoryNotification:
		var update dto.NotificationPreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategoryDisplay:
		var update dto.DisplayPreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategoryPrivacy:
		var update dto.PrivacyPreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategoryAccessibility:
		var update dto.AccessibilityPreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategoryLanguage:
		var update dto.LanguagePreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategorySecurity:
		var update dto.SecurityPreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategorySocial:
		var update dto.SocialPreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategorySound:
		var update dto.SoundPreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	case dto.PreferenceCategoryTheme:
		var update dto.ThemePreferencesUpdate

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			return nil, ErrInvalidJSON
		}

		return &update, nil
	default:
		return nil, service.ErrInvalidCategory
	}
}

func (h *PreferenceHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrUnauthorizedAccess):
		ForbiddenResponse(w, "Not authorized to access these preferences")
	case errors.Is(err, service.ErrInvalidCategory):
		ErrorResponse(w, http.StatusBadRequest, "INVALID_CATEGORY", "Invalid preference category")
	default:
		slog.Error("preference service error", "error", err)
		InternalErrorResponse(w)
	}
}

func (h *PreferenceHandler) handleBindError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmptyBody):
		ErrorResponse(w, http.StatusBadRequest, "EMPTY_BODY", "Request body is required")
	case errors.Is(err, ErrInvalidJSON), errors.Is(err, ErrInvalidFieldType):
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
	case errors.Is(err, ErrValidationFailed):
		ValidationErrorResponse(w, err)
	default:
		slog.Error("failed to bind request body", "error", err)
		ErrorResponse(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
}
