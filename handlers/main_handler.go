package handlers

import (
	"fmt"
	"net/http"

	"job_sender/core"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type MainHandler struct {
	authService          *core.AuthService
	errorReporterService *core.ErrorReporterService
}

func NewMainHandler(authService *core.AuthService, errorReporterService *core.ErrorReporterService) *MainHandler {
	return &MainHandler{
		authService:          authService,
		errorReporterService: errorReporterService,
	}
}

func (h *MainHandler) CreateRouter() *mux.Router {
	r := mux.NewRouter()

	r.Handle("/", http.RedirectHandler("/main", http.StatusFound))

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
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	if userInfo.IsLoggedIn {
		http.Redirect(w, r, "/auth/hashTags", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
