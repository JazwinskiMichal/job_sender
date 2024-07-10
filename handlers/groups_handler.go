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

type GroupsHandler struct {
	authService           *core.AuthService
	ownersDB              *core.OwnerDatabaseService
	groupsDB              *core.GroupsDatabaseService
	sessionManagerService *core.SessionManagerService
	templateService       *core.TemplateService
	errorReporterService  *core.ErrorReporterService
}

// NewGroupsHandler creates a new GroupsHandler.
func NewGroupsHandler(authService *core.AuthService, ownersDB *core.OwnerDatabaseService, groupsDB *core.GroupsDatabaseService, sessionManagerService *core.SessionManagerService, templateService *core.TemplateService, errorReporterService *core.ErrorReporterService) *GroupsHandler {
	return &GroupsHandler{
		authService:           authService,
		ownersDB:              ownersDB,
		groupsDB:              groupsDB,
		sessionManagerService: sessionManagerService,
		templateService:       templateService,
		errorReporterService:  errorReporterService,
	}
}

// RegisterGroupsHandlers registers group handlers.
func (h *GroupsHandler) RegisterGroupsHandlers(r *mux.Router) {
	r.Methods("GET").Path("/groups/add").HandlerFunc(h.ShowAddGroup)
	r.Methods("GET").Path("/groups/{ID}").HandlerFunc(h.GetGroup)
	r.Methods("GET").Path("/groups/{ID}/edit").HandlerFunc(h.EditGroup)

	r.Methods("POST").Path("/groups").HandlerFunc(h.AddGroup)
	r.Methods("PUT").Path("/groups/{ID}").HandlerFunc(h.UpdateGroup)

	r.Methods("DELETE").Path("/groups/{ID}").HandlerFunc(h.DeleteGroup)
}

// ShowAddGroup displays the add group page.
func (h *GroupsHandler) ShowAddGroup(w http.ResponseWriter, r *http.Request) {
	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	groupTmpl, err := h.templateService.ParseTemplate(constants.TemplateGroupEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse group template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(groupTmpl, w, r, nil, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// GetGroup gets a group.
func (h *GroupsHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID := mux.Vars(r)["ID"]
	if groupID == "" {
		http.Error(w, "ID of the group required", http.StatusBadRequest)
		return
	}

	group, err := h.groupsDB.GetGroup(groupID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Redirect(w, r, "/auth/groups/add", http.StatusFound)
		} else {
			h.errorReporterService.ReportError(w, r, err)
			http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/auth/contractors?groupID="+group.ID, http.StatusFound)
}

// EditGroup edits a group.
func (h *GroupsHandler) EditGroup(w http.ResponseWriter, r *http.Request) {
	groupID := mux.Vars(r)["ID"]
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	group, err := h.groupsDB.GetGroup(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	groupTmpl, err := h.templateService.ParseTemplate(constants.TemplateGroupEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse group template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(groupTmpl, w, r, group, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// AddGroup adds a group.
func (h *GroupsHandler) AddGroup(w http.ResponseWriter, r *http.Request) {
	// Get the group from the form.
	group, err := h.groupFromForm(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get group from form: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Get the owner id from session.
	ownerID, err := h.sessionManagerService.GetElement(r, constants.UserSessionName, constants.SesstionOwnerIdField)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner id from session: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Append the owner id to the group.
	ownerIDString, ok := ownerID.(string)
	if !ok {
		http.Error(w, "ownerID is required", http.StatusBadRequest)
		return
	}
	group.OwnerID = ownerIDString

	// Add the group.
	_, err = h.groupsDB.AddGroup(group)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not add group: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Get the owner.
	owner, err := h.ownersDB.GetOwnerByID(ownerIDString)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Update the owner with the group id.
	owner.GroupID = group.ID

	err = h.ownersDB.UpdateOwner(owner)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not update owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors?groupID="+group.ID, http.StatusFound)
}

// UpdateGroup updates a group.
func (h *GroupsHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	// Get the group from the form.
	group, err := h.groupFromForm(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get group from form: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Update the group.
	err = h.groupsDB.UpdateGroup(group)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors", http.StatusFound) // TODO: group id is needed
}

// DeleteGroup deletes a group.
func (h *GroupsHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	groupID := mux.Vars(r)["ID"]
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	err := h.groupsDB.DeleteGroup(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors", http.StatusFound) // TODO: group id is needed
}

// groupFromForm creates a group from a form.
func (h *GroupsHandler) groupFromForm(r *http.Request) (*types.Group, error) {
	return &types.Group{
		Name:    r.FormValue("name"),
		OwnerID: r.FormValue("ownerID"),
	}, nil
}
