package middlewares

import (
	"fmt"
	"net/http"

	"job_sender/core"
)

type authMiddleware struct {
	authService          *core.AuthService
	errorReporterService *core.ErrorReporterService
}

func NewAuthMiddleware(authService *core.AuthService, errorReporterService *core.ErrorReporterService) *authMiddleware {
	return &authMiddleware{
		authService:          authService,
		errorReporterService: errorReporterService,
	}
}

func (h *authMiddleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := h.authService.CheckUser(r)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
			return
		}

		if !userInfo.IsLoggedIn {
			// Redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
