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

type ContractorsHandler struct {
	authService   *core.AuthService
	groupsDB      *core.GroupsDatabaseService
	contractorsDB *core.ContractorsDatabaseService
	//storageService       *core.StorageService
	cloudTaskService     *core.CloudTasksService
	templateService      *core.TemplateService
	errorReporterService *core.ErrorReporterService

	envVariables *types.EnvVariables
}

// NewContractorsHandler creates a new ContractorsHandler.
func NewContractorsHandler(authService *core.AuthService, groupsDB *core.GroupsDatabaseService, contractorsDB *core.ContractorsDatabaseService, cloudTaskService *core.CloudTasksService, templateService *core.TemplateService, errorReporterService *core.ErrorReporterService, envVariables *types.EnvVariables) *ContractorsHandler {
	return &ContractorsHandler{
		authService:   authService,
		groupsDB:      groupsDB,
		contractorsDB: contractorsDB,
		//storageService:       storageService,
		cloudTaskService:     cloudTaskService,
		templateService:      templateService,
		errorReporterService: errorReporterService,

		envVariables: envVariables,
	}
}

// Register contractor handlers.
func (h *ContractorsHandler) RegisterContractorsHandler(r *mux.Router) {
	r.Methods("GET").Path("/contractors").HandlerFunc(h.ListContractors)
	r.Methods("GET").Path("/contractors/add").HandlerFunc(h.ShowAddContractor)
	r.Methods("GET").Path("/contractors/{ID}").HandlerFunc(h.GetContractor)
	r.Methods("GET").Path("/contractors/{ID}/timeSheet").HandlerFunc(h.GetContractorsTimeSheet)

	r.Methods("POST").Path("/contractors").HandlerFunc(h.AddContractor)
	r.Methods("POST").Path("/contractors/{ID}").HandlerFunc(h.EditContractor)
	r.Methods("POST").Path("/contractors/{ID}/timeSheet").HandlerFunc(h.AddContractorsTimeSheet)

	r.Methods("DELETE").Path("/contractors/{ID}").HandlerFunc(h.DeleteContractor)
}

// ListContractors lists all contractors for a group.
func (h *ContractorsHandler) ListContractors(w http.ResponseWriter, r *http.Request) {
	// Get the group ID from the query.
	groupID := r.URL.Query().Get("groupID")
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Get the group.
	_, err = h.groupsDB.GetGroup(groupID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Display the add group page.
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
		} else {
			h.errorReporterService.ReportError(w, r, err)
			http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
			return
		}
	}

	// Get the contractors.
	contractors, err := h.contractorsDB.ListContractors(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Start timeheet aggregation for each contractor
	for _, contractor := range contractors {
		taskCreated, err := h.cloudTaskService.CreateTimesheetAggregatorTask(h.envVariables.ProjectID, h.envVariables.ProjectLocationID, h.envVariables.EmailAggregatorQueueName, contractor, types.TimesheetAggregation{
			GroupID:      groupID,
			ContractorID: contractor.ID,
		})
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("could not create timesheet aggregator task: %w", err))
			continue
		}

		if !taskCreated {
			continue
		}

		// Update the contractor
		err = h.contractorsDB.UpdateContractor(contractor)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("could not update contractor: %w", err))
			continue
		}
	}

	data := map[string]interface{}{
		"GroupID":     groupID,
		"Contractors": contractors,
	}

	// Execute the template
	listTmpl, err := h.templateService.ParseTemplate(constants.TemplateContractorsListName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse list template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	err = h.templateService.ExecuteTemplate(listTmpl, w, r, data, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not execute template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// ShowAddContractor shows the form to add a contractor.
func (h *ContractorsHandler) ShowAddContractor(w http.ResponseWriter, r *http.Request) {
	// Get the group ID from the query.
	groupID := r.URL.Query().Get("groupID")
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	addTmpl, err := h.templateService.ParseTemplate(constants.TemplateContractorsAddName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse add template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Create the data to pass to the template
	data := map[string]string{
		"Name":    "",
		"Surname": "",
		"Email":   "",
		"GroupID": groupID,
	}

	err = h.templateService.ExecuteTemplate(addTmpl, w, r, data, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not execute template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// GetContractor gets a contractor by ID.
func (h *ContractorsHandler) GetContractor(w http.ResponseWriter, r *http.Request) {
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

	editTmpl, err := h.templateService.ParseTemplate(constants.TemplateContractorsEditName)
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

// GetContractorsTimeSheet gets a contractor's timesheet by ID.
func (h *ContractorsHandler) GetContractorsTimeSheet(w http.ResponseWriter, r *http.Request) {
	// Get the contractor ID from the request.
	id := mux.Vars(r)["ID"]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Get the contractor's timesheet.
	_, err := h.contractorsDB.GetContractorsTimesheet(id)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}
}

// AddContractor adds a contractor to a group.
func (h *ContractorsHandler) AddContractor(w http.ResponseWriter, r *http.Request) {
	// Get the group ID from the query.
	groupID := r.URL.Query().Get("groupID")
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

	http.Redirect(w, r, "/auth/contractors?groupID="+groupID, http.StatusFound)
}

// EditContractor updates a contractor.
func (h *ContractorsHandler) EditContractor(w http.ResponseWriter, r *http.Request) {
	// Get the contractor id from the request.
	id := mux.Vars(r)["ID"]
	if id == "" {
		http.Error(w, "ID of the contractor is required", http.StatusBadRequest)
		return
	}

	// Get the group ID from the query.
	groupID := r.URL.Query().Get("groupID")
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

	// Update the contractor.
	contractor.ID = id
	contractor.GroupID = groupID
	err = h.contractorsDB.UpdateContractor(contractor)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/auth/contractors?groupID="+groupID, http.StatusFound)
}

// AddContractorsTimeSheet adds a timesheet to a contractor.
func (h *ContractorsHandler) AddContractorsTimeSheet(w http.ResponseWriter, r *http.Request) {
	// TODO: implement this
}

// DeleteContractor deletes a contractor.
func (h *ContractorsHandler) DeleteContractor(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, "/auth/contractors", http.StatusFound) // TODO: group id is needed
}

// contractorFromForm creates a contractor from a form.
func (h *ContractorsHandler) contractorFromForm(r *http.Request) (*types.Contractor, error) {
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
		PhotoURL: r.FormValue("photoURL"),
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
