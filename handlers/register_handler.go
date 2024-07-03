package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"job_sender/core"
	constants "job_sender/utils/constants"

	"github.com/gorilla/mux"
)

type RegisterHandler struct {
	firebaseService      *core.FirebaseService
	authService          *core.AuthService
	templateService      *core.TemplateService
	emailService         *core.EmailService
	errorReporterService *core.ErrorReporterService
}

func NewRegisterHandler(firebaseService *core.FirebaseService, authService *core.AuthService, templateService *core.TemplateService, emailService *core.EmailService, errorReporterService *core.ErrorReporterService) *RegisterHandler {
	return &RegisterHandler{
		firebaseService:      firebaseService,
		authService:          authService,
		templateService:      templateService,
		emailService:         emailService,
		errorReporterService: errorReporterService,
	}
}

func (h *RegisterHandler) RegisterRegisterHandlers(r *mux.Router) {
	r.Methods("GET").Path("/register").Handler(http.HandlerFunc(h.showRegister))
	r.Methods("GET").Path("/register/confirm").Handler(http.HandlerFunc(h.showRegisterConfirm))
	r.Methods("POST").Path("/register").Handler(http.HandlerFunc(h.register))
}

// showRegister displays the register page.
func (h *RegisterHandler) showRegister(w http.ResponseWriter, r *http.Request) {
	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	registerTmpl, err := h.templateService.ParseTemplate(constants.TemplateRegisterName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse register template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(registerTmpl, w, r, nil, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// showRegisterConfirm displays the register confirm page.
func (h *RegisterHandler) showRegisterConfirm(w http.ResponseWriter, r *http.Request) {
	// Retrieve the cookie
	cookie, err := r.Cookie("emailConfirmationAddress")
	if err != nil {
		// Handle error (e.g., cookie not found)
		h.errorReporterService.ReportError(w, r, fmt.Errorf("email cookie not found: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Decode the email value from the cookie
	decodedEmail, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		// Handle error
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not decode email from cookie: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Parse the template
	confirmRegistrationTmpl, err := h.templateService.ParseTemplate(constants.TemplateConfirmRegisterName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse confirm registration template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Create the data to pass to the template
	data := map[string]string{
		"Email": decodedEmail,
	}

	err = h.templateService.ExecuteTemplate(confirmRegistrationTmpl, w, r, data, nil)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not execute template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Clear the cookie by setting its MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:   "emailConfirmationAddress",
		MaxAge: -1, // Immediately expire the cookie
	})
}

// register processes the register form.
func (h *RegisterHandler) register(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if email == "" {
		h.showError(w, r, "Email missing", "email missing")
		return
	}

	// Check if the email is already registered
	userExists, err := h.firebaseService.CheckIfUserExists(email)
	if err != nil {
		h.showError(w, r, "Could not register", fmt.Sprintf("could not check if user exists: %v", err))
		return
	}

	if userExists {
		h.showError(w, r, "Email already registered", "email already registered")
		return
	}

	if password == "" || confirmPassword == "" {
		h.showError(w, r, "Password or confirm password missing", "Password or confirm password missing")
		return
	}

	if password != confirmPassword {
		h.showError(w, r, "Password and confirm password do not match", "password and confirm password do not match")
		return
	}

	_, err = h.authService.Register(email, password)
	if err != nil {
		h.showError(w, r, "Could not register", fmt.Sprintf("could not register user: %v", err))
		return
	}

	// Send verification email
	ctx := r.Context()
	client, err := h.firebaseService.Auth(ctx)
	if err != nil {
		h.showError(w, r, "Could not send verification email", fmt.Sprintf("could not get auth client: %v", err))
		return
	}

	link, err := client.EmailVerificationLink(ctx, email)
	if err != nil {
		h.showError(w, r, "Could not send verification email", fmt.Sprintf("could not get email verification link: %v", err))
		return
	}

	err = h.emailService.SendVerificationEmail(email, link)
	if err != nil {
		h.showError(w, r, "Could not send verification email", fmt.Sprintf("could not send verification email: %v", err))
		return
	}

	// Set a cookie to pass the email to the confirm page
	http.SetCookie(w, &http.Cookie{
		Name:     "emailConfirmationAddress",
		Value:    url.QueryEscape(email), // Ensure the email is URL-encoded
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to true in production to ensure it's sent over HTTPS
		MaxAge:   300,  // Expires in 5 minutes
	})

	// Redirect to the confirm page
	http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
}

// showError displays an error message to the user.
func (h *RegisterHandler) showError(w http.ResponseWriter, r *http.Request, errorMessage string, debugErrorMessage string) {
	// TODO: nie zawsze wymagane jest reportowanie błędu, wydzielic to do osobnej metody
	h.errorReporterService.ReportError(w, r, fmt.Errorf("login error: %s", debugErrorMessage))

	registerTmpl, err := h.templateService.ParseTemplate(constants.TemplateRegisterName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse register template: %w", err))
	}

	err = h.templateService.ShowError(registerTmpl, w, r, errorMessage)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
	}
}
