package handlers

import (
	"fmt"
	"net/http"

	"job_sender/core"
	"job_sender/types"
	constants "job_sender/utils/constants"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OwnersHandler struct {
	authService           *core.AuthService
	ownersDB              *core.OwnerDatabaseService
	sessionManagerService *core.SessionManagerService
	templateService       *core.TemplateService
	errorReporterService  *core.ErrorReporterService
}

// NewOwnersHandler creates a new OwnersHandler.
func NewOwnersHandler(authService *core.AuthService, ownersDB *core.OwnerDatabaseService, sessionManagerService *core.SessionManagerService, templateService *core.TemplateService, errorReporterService *core.ErrorReporterService) *OwnersHandler {
	return &OwnersHandler{
		authService:           authService,
		ownersDB:              ownersDB,
		sessionManagerService: sessionManagerService,
		templateService:       templateService,
		errorReporterService:  errorReporterService,
	}
}

// RegisterOwnersHandlers registers owners handlers.
func (h *OwnersHandler) RegisterOwnersHandlers(r *mux.Router) {
	r.Methods("GET").Path("/owners/add").HandlerFunc(h.ShowAddOwner)
	r.Methods("GET").Path("/owners/{ID}").HandlerFunc(h.GetOwner)
	r.Methods("GET").Path("/owners/{ID}/edit").HandlerFunc(h.EditOwner)

	r.Methods("POST").Path("/owners").HandlerFunc(h.AddOwner)
	r.Methods("PUT").Path("/owners/{ID}").HandlerFunc(h.UpdateOwner)

	r.Methods("DELETE").Path("/owners/{ID}").HandlerFunc(h.DeleteOwner)
}

// ShowAddOwner displays the add owner page.
func (h *OwnersHandler) ShowAddOwner(w http.ResponseWriter, r *http.Request) {
	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	ownerTmpl, err := h.templateService.ParseTemplate(constants.TemplateOwnerEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse owner template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Create data for email
	data := map[string]interface{}{
		"Name":    "",
		"Surname": "",
		"Email":   userInfo.Email,
	}

	err = h.templateService.ExecuteTemplate(ownerTmpl, w, r, data, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
	}
}

// GetOwner gets an owner by ID.
func (h *OwnersHandler) GetOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := mux.Vars(r)["ID"]
	if ownerID == "" {
		http.Error(w, "ownerID is required", http.StatusBadRequest)
		return
	}

	owner, err := h.ownersDB.GetOwnerByID(ownerID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Redirect(w, r, "/auth/owners/add", http.StatusFound)
			return
		} else {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/auth/groups/"+owner.GroupID, http.StatusFound)
}

// EditOwner displays the edit owner page.
func (h *OwnersHandler) EditOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := mux.Vars(r)["ID"]
	if ownerID == "" {
		http.Error(w, "ownerID is required", http.StatusBadRequest)
		return
	}

	owner, err := h.ownersDB.GetOwnerByID(ownerID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	ownerTmpl, err := h.templateService.ParseTemplate(constants.TemplateOwnerEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse owner template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(ownerTmpl, w, r, owner, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
	}
}

// AddOwner adds an owner.
func (h *OwnersHandler) AddOwner(w http.ResponseWriter, r *http.Request) {
	// Get the owner id from session.
	ownerID, err := h.sessionManagerService.GetElement(r, constants.UserSessionName, constants.UserSesstionOwnerIdField)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner id from session: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Get the owner from the form.
	owner, err := ownerFromForm(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner from form: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	if ownerID == nil {
		http.Error(w, "ownerID is required", http.StatusBadRequest)
		return
	}

	// Set the owner id.
	ownerIDString, ok := ownerID.(string)
	if !ok {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not convert owner id to string: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
	owner.ID = string(ownerIDString)

	// Add the owner.
	err = h.ownersDB.AddOwner(owner)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not add owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/groups/add", http.StatusFound)
}

// UpdateOwner updates an owner.
func (h *OwnersHandler) UpdateOwner(w http.ResponseWriter, r *http.Request) {
	// Get the owner ID from the request.
	ownerID := mux.Vars(r)["ID"]
	if ownerID == "" {
		http.Error(w, "ownerID is required", http.StatusBadRequest)
		return
	}

	// Get the owner from the form.
	owner, err := ownerFromForm(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner from form: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Update the owner.
	owner.ID = ownerID
	err = h.ownersDB.UpdateOwner(owner)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not update owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors", http.StatusFound) // TODO: group id is needed
}

// DeleteOwner deletes an owner.
func (h *OwnersHandler) DeleteOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := mux.Vars(r)["ID"]
	if ownerID == "" {
		http.Error(w, "ownerID is required", http.StatusBadRequest)
		return
	}

	err := h.ownersDB.DeleteOwner(ownerID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not delete owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

func ownerFromForm(r *http.Request) (*types.Owner, error) {
	owner := &types.Owner{
		Name:     r.FormValue("name"),
		Surname:  r.FormValue("surname"),
		Email:    r.FormValue("email"),
		Phone:    r.FormValue("phone"),
		PhotoURL: r.FormValue("photoURL"),
	}

	return owner, nil
}
