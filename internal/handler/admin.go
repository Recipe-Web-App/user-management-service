//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// AdminHandler handles admin HTTP endpoints.
type AdminHandler struct {
	userService  service.UserService
	adminService service.AdminService
	binder       *RequestBinder
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(userService service.UserService, adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		userService:  userService,
		adminService: adminService,
		binder:       NewRequestBinder(),
	}
}

// GetUserStats handles GET /admin/users/stats.
func (h *AdminHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userService.GetUserStats(r.Context())
	if err != nil {
		InternalErrorResponse(w)
		return
	}

	SuccessResponse(w, http.StatusOK, stats)
}

// ClearCache handles POST /admin/cache/clear.
func (h *AdminHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	var req dto.CacheClearRequest

	err := h.binder.BindAndValidate(r, &req)
	if err != nil {
		if !errors.Is(err, ErrEmptyBody) {
			h.handleBindError(w, err)
			return
		}
		// If body is empty, we continue with empty req which will be defaulted below
	}

	// Default to * if empty
	if req.KeyPattern == "" {
		req.KeyPattern = "*"
	}

	resp, err := h.adminService.ClearCache(r.Context(), req.KeyPattern)
	if err != nil {
		InternalErrorResponse(w)
		return
	}

	SuccessResponse(w, http.StatusOK, resp)
}

func (h *AdminHandler) handleBindError(w http.ResponseWriter, err error) {
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
