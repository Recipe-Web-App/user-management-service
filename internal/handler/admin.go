//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"net/http"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// AdminHandler handles admin HTTP endpoints.
type AdminHandler struct {
	userService service.UserService
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(userService service.UserService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
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
