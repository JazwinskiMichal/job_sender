package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "time/tzdata"

	"job_sender/core"
	"job_sender/types"
	constants "job_sender/utils/constants"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupsHandler struct {
	authService           *core.AuthService
	schedulerService      *core.SchedulerService
	sessionManagerService *core.SessionManagerService
	storageService        *core.StorageService
	templateService       *core.TemplateService
	errorReporterService  *core.ErrorReporterService

	ownersDB *core.OwnerDatabaseService
	groupsDB *core.GroupsDatabaseService
}

// NewGroupsHandler creates a new GroupsHandler.
func NewGroupsHandler(authService *core.AuthService, schedulerService *core.SchedulerService, sessionManagerService *core.SessionManagerService, storageService *core.StorageService, templateService *core.TemplateService, errorReporterService *core.ErrorReporterService, ownersDB *core.OwnerDatabaseService, groupsDB *core.GroupsDatabaseService) *GroupsHandler {
	return &GroupsHandler{
		authService:           authService,
		schedulerService:      schedulerService,
		sessionManagerService: sessionManagerService,
		storageService:        storageService,
		templateService:       templateService,
		errorReporterService:  errorReporterService,

		ownersDB: ownersDB,
		groupsDB: groupsDB,
	}
}

// RegisterGroupsHandlers registers group handlers.
func (h *GroupsHandler) RegisterGroupsHandlers(r *mux.Router) {
	r.Methods("GET").Path("/groups/add").HandlerFunc(h.ShowAddGroup)
	r.Methods("GET").Path("/groups/{ID}").HandlerFunc(h.GetGroup)
	r.Methods("GET").Path("/groups/{ID}/edit").HandlerFunc(h.ShowEditGroup)
	r.Methods("GET").Path("/groups/{ID}/delete").HandlerFunc(h.DeleteGroup)

	r.Methods("POST").Path("/groups").HandlerFunc(h.AddGroup)
	r.Methods("POST").Path("/groups/{ID}").HandlerFunc(h.EditGroup)
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
			http.Redirect(w, r, "/auth/groups/add", http.StatusSeeOther)
			return
		} else {
			h.errorReporterService.ReportError(w, r, err)
			http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
			return
		}
	}

	http.Redirect(w, r, "/auth/contractors?groupID="+group.ID, http.StatusSeeOther)
}

// ShowAddGroup displays the add group page.
func (h *GroupsHandler) ShowAddGroup(w http.ResponseWriter, r *http.Request) {
	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	groupTmpl, err := h.templateService.ParseTemplate(constants.TemplateGroupAddName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse group template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	err = h.templateService.ExecuteTemplate(groupTmpl, w, r, nil, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}
}

// EditGroup edits a group.
func (h *GroupsHandler) ShowEditGroup(w http.ResponseWriter, r *http.Request) {
	groupID := mux.Vars(r)["ID"]
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	group, err := h.groupsDB.GetGroup(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	userInfo, err := h.authService.CheckUser(r)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not check user: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Add group info to the user info.
	userInfo.GroupID = group.ID
	userInfo.GroupName = group.Name

	groupTmpl, err := h.templateService.ParseTemplate(constants.TemplateGroupEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse group template: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	err = h.templateService.ExecuteTemplate(groupTmpl, w, r, group, userInfo)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}
}

// AddGroup adds a group.
func (h *GroupsHandler) AddGroup(w http.ResponseWriter, r *http.Request) {
	// Get the group from the form.
	group, err := h.groupFromForm(r)
	if err != nil {
		h.showError(w, r, fmt.Sprintf("could not get group from form: %v", err))
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get group from form: %w", err))
		return
	}

	// Get the owner id from session.
	ownerID, err := h.sessionManagerService.GetElement(r, constants.UserSessionName, constants.SesstionOwnerIdField)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner id from session: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
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
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Get the owner.
	owner, err := h.ownersDB.GetOwnerByID(ownerIDString)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Update the owner with the group id.
	owner.GroupID = group.ID

	err = h.ownersDB.UpdateOwner(owner)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not update owner: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Create the timesheet request schedule job for the group.
	err = h.schedulerService.CreateTimesheetRequestJob(group.ID, &group.Schedule)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not create timesheet request job: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/auth/contractors?groupID="+group.ID, http.StatusSeeOther)
}

// EditGroup updates a group.
func (h *GroupsHandler) EditGroup(w http.ResponseWriter, r *http.Request) {
	groupID := mux.Vars(r)["ID"]
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	// Get the group from the form.
	group, err := h.groupFromForm(r)
	if err != nil {
		h.showError(w, r, fmt.Sprintf("could not get group from form: %v", err))
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get group from form: %w", err))
		return
	}

	group.ID = groupID

	// Get the owner id from session.
	ownerID, err := h.sessionManagerService.GetElement(r, constants.UserSessionName, constants.SesstionOwnerIdField)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not get owner id from session: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Append the owner id to the group.
	ownerIDString, ok := ownerID.(string)
	if !ok {
		http.Error(w, "ownerID is required", http.StatusBadRequest)
		return
	}
	group.OwnerID = ownerIDString

	// Update the group.
	err = h.groupsDB.UpdateGroup(group)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Update the timesheet request schedule job for the group.
	err = h.schedulerService.EditTimesheetRequestJob(group.ID, &group.Schedule)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not edit timesheet request job: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/auth/contractors?groupID="+group.ID, http.StatusSeeOther)
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
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Delete the timsheet request schedule jobs for the group.
	err = h.schedulerService.DeleteTimesheetRequestJob(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not delete timesheet request job: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Delete the timesheets from the storage.
	err = h.storageService.DeleteFiles(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not delete timesheets: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/auth/groups/add", http.StatusSeeOther)
}

// groupFromForm creates a group from a form.
func (h *GroupsHandler) groupFromForm(r *http.Request) (*types.Group, error) {
	name := r.FormValue("name")
	weekday := r.FormValue("weekday")
	monthday := r.FormValue("monthday")
	timezoneStr := r.FormValue("timezone")
	timeStr := r.FormValue("time")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	intervalTypeStr := r.FormValue("interval_type")
	intervalStr := r.FormValue("interval")

	if intervalTypeStr == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	var intervalType constants.IntervalTypes
	switch intervalTypeStr {
	case "weeks":
		intervalType = constants.Weeks
	case "months":
		intervalType = constants.Months
	default:
		return nil, fmt.Errorf("invalid interval type: %s", intervalTypeStr)
	}

	if intervalType == constants.Months && monthday == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	if intervalType == constants.Weeks && weekday == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	if name == "" || timezoneStr == "" || timeStr == "" || intervalStr == "" || startDateStr == "" || endDateStr == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	timezoneParsed, err := time.LoadLocation(timezoneStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse timezone: %w", err)
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse start date: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse end date: %w", err)
	}

	timeParsed, err := time.Parse("15:04", timeStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse time: %w", err)
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("could not convert interval to number: %w", err)
	}

	// Round the monthday to the nearest valid day of the month.
	if intervalType == constants.Months {
		monthdayInt, err := strconv.Atoi(monthday)
		if err != nil {
			return nil, fmt.Errorf("could not convert monthday to number: %w", err)
		}

		// Get current month and check if the monthday is valid.
		now := time.Now()
		year, month, _ := now.Date()

		// Get the last day of current month.
		lastDay := time.Date(year, month, 0, 0, 0, 0, 0, time.UTC).Day()

		// Round the monthday to the nearest valid day of the month.
		if monthdayInt > lastDay {
			monthday = fmt.Sprintf("%d", lastDay)
		}
	}

	return &types.Group{
		OwnerID: r.FormValue("ownerID"),
		Name:    name,

		Schedule: types.Schedule{
			Weekday:      weekday,
			Monthday:     monthday,
			Timezone:     timezoneParsed.String(),
			Time:         timeParsed.Format("15:04"),
			StartDate:    startDate.Format("2006-01-02"),
			EndDate:      endDate.Format("2006-01-02"),
			IntervalType: intervalType,
			Interval:     interval,
		},
	}, nil
}

// showError renders the login page with an error message.
func (h *GroupsHandler) showError(w http.ResponseWriter, r *http.Request, errorMessage string) {
	groupTmpl, err := h.templateService.ParseTemplate(constants.TemplateGroupEditName)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("could not parse group template: %w", err))
	}

	err = h.templateService.ShowError(groupTmpl, w, r, errorMessage)
	if err != nil {
		h.errorReporterService.ReportError(w, r, err)
	}
}
