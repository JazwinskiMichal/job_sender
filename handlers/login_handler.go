package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"job_sender/core"
	constants "job_sender/utils/constants"

	"github.com/gorilla/mux"
)

type LoginHandler struct {
	authService           *core.AuthService
	firebaseService       *core.FirebaseService
	templateService       *core.TemplateService
	sessionManagerService *core.SessionManagerService
	errorReporterService  *core.ErrorReporterService
}

func NewLoginHandler(authService *core.AuthService, firebaseService *core.FirebaseService, templateService *core.TemplateService, sessionManagerService *core.SessionManagerService, errorReporterService *core.ErrorReporterService) *LoginHandler {
	return &LoginHandler{
		authService:           authService,
		firebaseService:       firebaseService,
		templateService:       templateService,
		sessionManagerService: sessionManagerService,
		errorReporterService:  errorReporterService,
	}
}

func (h *LoginHandler) RegisterLoginHandlers(r *mux.Router) {
	r.Methods("GET").Path("/login").Handler(http.HandlerFunc(h.showLogin))

	r.Methods("POST").Path("/login").Handler(http.HandlerFunc(h.login))
	r.Methods("POST").Path("/logout").Handler(http.HandlerFunc(h.logout))
}

// showLogin displays the login page.
func (h *LoginHandler) showLogin(w http.ResponseWriter, r *http.Request) {
	loginTmpl, err := h.templateService.ParseTemplate(constants.TemplateLoginName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse login template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	err = h.templateService.ExecuteTemplate(loginTmpl, w, r, nil, nil)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}
}

// login processes the login form.
func (h *LoginHandler) login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.showError(w, r, "Email or password missing")
		return
	}

	// Login the user
	responseBody, err := h.authService.Login(email, password)
	if err != nil {
		h.showError(w, r, "Invalid email or password")
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not login user: %w", err))
		return
	}

	if responseBody.IdToken == "" {
		h.showError(w, r, "Invalid email or password")
		return
	}

	// Extract the expires in time
	expiresIn, err := strconv.ParseInt(responseBody.ExpiresIn, 10, 64)
	if err != nil {
		h.showError(w, r, "Could not parse expires in time")
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse expires in time: %w", err))
		return
	}

	// Convert expiresIn to a timestamp
	expirationTimestamp := time.Now().Add(time.Second * time.Duration(expiresIn))

	// Check if the user is verified
	isVerified, err := h.firebaseService.CheckIsUserVerified(email)
	if err != nil {
		h.showError(w, r, "Could not check if user is verified")
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check if user is verified: %w", err))
		return
	}

	// Create the data
	data := map[string]interface{}{
		constants.SessionTokenField:     responseBody.IdToken,
		constants.SessionEmailField:     responseBody.Email,
		constants.SessionIsVerfiedField: isVerified,
		constants.SesstionOwnerIdField:  responseBody.LocalId,
	}

	// Create the session
	_, err = h.sessionManagerService.CreateSession(w, r, constants.UserSessionName, expirationTimestamp, data)
	if err != nil {
		h.showError(w, r, "Could not create session")
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not create session: %w", err))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/auth/owners/%s", responseBody.LocalId), http.StatusSeeOther)
}

// logout logs out the user.
func (h *LoginHandler) logout(w http.ResponseWriter, r *http.Request) {
	// Delete the user session
	err := h.sessionManagerService.DeleteSession(w, r, constants.UserSessionName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not delete session: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Delete the timesheet aggregation session
	err = h.sessionManagerService.DeleteSession(w, r, constants.TimesheetAggegationSessionName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not delete session: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/main", http.StatusSeeOther)
}

// showError renders the login page with an error message.
func (h *LoginHandler) showError(w http.ResponseWriter, r *http.Request, errorMessage string) {
	loginTmpl, err := h.templateService.ParseTemplate(constants.TemplateLoginName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse login template: %w", err))
	}

	err = h.templateService.ShowError(loginTmpl, w, r, errorMessage)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
	}
}
