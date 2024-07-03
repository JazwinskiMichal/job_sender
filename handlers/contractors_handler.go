package handlers

import (
	"fmt"
	"job_sender/core"
	"job_sender/types"
	"net/http"

	constants "job_sender/utils/constants"

	"github.com/gorilla/mux"
)

type ContractorHandler struct {
	authService   *core.AuthService
	contractorsDB *core.ContractorsDatabaseService
	//storageService       *core.StorageService
	templateService      *core.TemplateService
	errorReporterService *core.ErrorReporterService
}

// NewContractorHandler creates a new ContractorHandler.
func NewContractorHandler(authService *core.AuthService, contractorsDB *core.ContractorsDatabaseService, templateService *core.TemplateService, errorReporterService *core.ErrorReporterService) *ContractorHandler {
	return &ContractorHandler{
		authService:   authService,
		contractorsDB: contractorsDB,
		//storageService:       storageService,
		templateService:      templateService,
		errorReporterService: errorReporterService,
	}
}

// Register contractor handlers.
func (h *ContractorHandler) RegisterContractorHandlers(r *mux.Router) {
	r.Methods("GET").Path("/contractors").HandlerFunc(h.ListContractors)
	r.Methods("GET").Path("/contractors/add").HandlerFunc(h.ShowAddContractor)
	r.Methods("GET").Path("/contractors/{ID}").HandlerFunc(h.GetContractor)

	r.Methods("POST").Path("/contractors{groupID}").HandlerFunc(h.AddContractor)
	r.Methods("PUT").Path("/contractors{ID}").HandlerFunc(h.UpdateContractor)

	r.Methods("DELETE").Path("/contractors{ID}").HandlerFunc(h.DeleteContractor)
}

// ListContractors lists all contractors for a group.
func (h *ContractorHandler) ListContractors(w http.ResponseWriter, r *http.Request) {
	// Get the group ID from the request.
	groupID := mux.Vars(r)["groupID"]
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}
	// TODO: Dodać group database service i w tym miejscu gdy nie ma żadnej grupy do odpalic widok tworzenia nowej grupy

	// Get the contractors.
	contractors, err := h.contractorsDB.ListContractors(groupID)
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

	listTmpl, err := h.templateService.ParseTemplate(constants.TemplateListName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse list template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(listTmpl, w, r, contractors, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not execute template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// ShowAddContractor shows the form to add a contractor.
func (h *ContractorHandler) ShowAddContractor(w http.ResponseWriter, r *http.Request) {
	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	addTmpl, err := h.templateService.ParseTemplate(constants.TemplateEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse add template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(addTmpl, w, r, nil, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not execute template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// GetContractor gets a contractor by ID.
func (h *ContractorHandler) GetContractor(w http.ResponseWriter, r *http.Request) {
	// Get the contractor ID from the request.
	id := mux.Vars(r)["ID"]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Get the contractor.
	contractor, err := h.contractorsDB.GetContractor(id)
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

	editTmpl, err := h.templateService.ParseTemplate(constants.TemplateEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse edit template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(editTmpl, w, r, contractor, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not execute template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// AddContractor adds a contractor to a group.
func (h *ContractorHandler) AddContractor(w http.ResponseWriter, r *http.Request) {
	// Get the group ID from the request.
	groupID := mux.Vars(r)["groupID"]
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	// Get the contractor from the form.
	contractor, err := h.contractorFromForm(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get contractor from form: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Add the contractor.
	err = h.contractorsDB.AddContractor(groupID, contractor)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors", http.StatusFound)
}

// UpdateContractor updates a contractor.
func (h *ContractorHandler) UpdateContractor(w http.ResponseWriter, r *http.Request) {
	// Get the contractor ID from the request.
	id := mux.Vars(r)["ID"]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Get the contractor from the form.
	contractor, err := h.contractorFromForm(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get contractor from form: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Update the contractor.
	contractor.ID = id
	err = h.contractorsDB.UpdateContractor(contractor)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors", http.StatusFound)
}

// DeleteContractor deletes a contractor.
func (h *ContractorHandler) DeleteContractor(w http.ResponseWriter, r *http.Request) {
	// Get the contractor ID from the request.
	id := mux.Vars(r)["ID"]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Delete the contractor.
	err := h.contractorsDB.DeleteContractor(id)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors", http.StatusFound)
}

// contractorFromForm creates a contractor from a form.
func (h *ContractorHandler) contractorFromForm(r *http.Request) (*types.Contractor, error) {
	// ctx := r.Context()

	// imageUrl, err := h.uploadFileFromForm(ctx, r)
	// if err != nil {
	// 	return nil, fmt.Errorf("could not upload file: %w", err)
	// }
	// if imageUrl == "" {
	// 	imageUrl = r.FormValue("imageURL")
	// }

	contractor := &types.Contractor{
		Name:     r.FormValue("name"),
		Surname:  r.FormValue("surname"),
		Email:    r.FormValue("email"),
		Phone:    r.FormValue("phone"),
		PhotoURL: "",
	}

	return contractor, nil
}

// func (h *ContractorHandler) uploadFileFromForm(ctx context.Context, r *http.Request) (url string, err error) {
// 	f, fh, err := r.FormFile("image")
// 	if err == http.ErrMissingFile {
// 		return "", nil
// 	}
// 	if err != nil {
// 		return "", err
// 	}

// 	// random filename, retaining existing extension.
// 	name := uuid.Must(uuid.NewV4()).String() + path.Ext(fh.Filename)

// 	url, err = h.storageService.UploadFileFromForm(ctx, f, fh, "image", name)
// 	if err != nil {
// 		return "", err
// 	}

// 	return url, nil
// }
