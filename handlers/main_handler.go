package handlers

import (
	"fmt"
	"net/http"

	"job_sender/core"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MainHandler struct {
	authService          *core.AuthService
	errorReporterService *core.ErrorReporterService

	ownersDB *core.OwnerDatabaseService
}

func NewMainHandler(authService *core.AuthService, errorReporterService *core.ErrorReporterService, ownersDB *core.OwnerDatabaseService) *MainHandler {
	return &MainHandler{
		authService:          authService,
		errorReporterService: errorReporterService,

		ownersDB: ownersDB,
	}
}

func (h *MainHandler) CreateRouter() *mux.Router {
	r := mux.NewRouter()

	r.Handle("/", http.RedirectHandler("/main", http.StatusSeeOther))

	r.Methods("GET").Path("/main").Handler(http.HandlerFunc(h.showMain))

	// Delegate all of the HTTP routing and serving to the gorilla/mux router.
	// Log all requests using the standard Apache format.
	http.Handle("/", handlers.CombinedLoggingHandler(h.errorReporterService.LogWriter, r))
	return r
}

func (h *MainHandler) showMain(w http.ResponseWriter, r *http.Request) {
	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	if userInfo.IsLoggedIn {
		// Get owner by email
		owner, err := h.ownersDB.GetOwnerByEmail(userInfo.Email)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				http.Redirect(w, r, "/auth/owners/add", http.StatusSeeOther)
				return
			}

			h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner by email: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/auth/owners/"+owner.ID, http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
}
