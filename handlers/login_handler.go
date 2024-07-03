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
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(loginTmpl, w, r, nil, nil)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// login processes the login form.
func (h *LoginHandler) login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.showError(w, r, "Email or password missing", "email or password missing")
		return
	}

	// Login the user
	responseBody, err := h.authService.Login(email, password)
	if err != nil {
		h.showError(w, r, "Invalid email or password", fmt.Sprintf("could not login user: %v", err))
		return
	}

	if responseBody.IdToken == "" {
		h.showError(w, r, "Invalid email or password", "id token is empty")
		return
	}

	// Extract the expires in time
	expiresIn, err := strconv.ParseInt(responseBody.ExpiresIn, 10, 64)
	if err != nil {
		h.showError(w, r, "Could not parse expires in time", fmt.Sprintf("could not parse expires in time: %v", err))
		return
	}

	// Convert expiresIn to a timestamp
	expirationTimestamp := time.Now().Add(time.Second * time.Duration(expiresIn))

	// Check if the user is verified
	isVerified, err := h.firebaseService.CheckIsUserVerified(email)
	if err != nil {
		h.showError(w, r, "Could not check if user is verified", fmt.Sprintf("could not check if user is verified: %v", err))
		return
	}

	// Create the data
	data := map[string]interface{}{
		constants.UserSessionTokenField:     responseBody.IdToken,
		constants.UserSessionEmailField:     responseBody.Email,
		constants.UserSessionIsVerfiedField: isVerified,
	}

	// Create the session
	err = h.sessionManagerService.CreateSession(w, r, constants.UserSessionName, expirationTimestamp, data)
	if err != nil {
		h.showError(w, r, "Could not create session", fmt.Sprintf("could not create session: %v", err))
		return
	}

	http.Redirect(w, r, "/auth/hashTags", http.StatusFound)
}

// logout logs out the user.
func (h *LoginHandler) logout(w http.ResponseWriter, r *http.Request) {
	err := h.sessionManagerService.DeleteSession(w, r, constants.UserSessionName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not delete session: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/main", http.StatusFound)
}

// showError renders the login page with an error message.
func (h *LoginHandler) showError(w http.ResponseWriter, r *http.Request, errorMessage string, debugErrorMessage string) {
	h.errorReporterService.ReportError(w, r, fmt.Errorf("login error: %s", debugErrorMessage))

	loginTmpl, err := h.templateService.ParseTemplate(constants.TemplateLoginName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse login template: %w", err))
	}

	err = h.templateService.ShowError(loginTmpl, w, r, errorMessage)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
	}
}
