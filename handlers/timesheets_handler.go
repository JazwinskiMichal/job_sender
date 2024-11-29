package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"job_sender/core"
	"job_sender/types"
	constants "job_sender/utils/constants"

	"github.com/gorilla/mux"
)

type TimesheetsHandler struct {
	emailService         *core.EmailService
	storageService       *core.StorageService
	errorReporterService *core.ErrorReporterService

	groupsDB      *core.GroupsDatabaseService
	contractorsDB *core.ContractorsDatabaseService
	timesheetsDB  *core.TimesheetsDatabaseService
}

// NewTimesheetsHandler creates a new TimesheetsHandler.
func NewTimesheetsHandler(emailService *core.EmailService, storageService *core.StorageService, errorReporterService *core.ErrorReporterService, groupsDB *core.GroupsDatabaseService, contractorsDB *core.ContractorsDatabaseService, timesheetsDB *core.TimesheetsDatabaseService) *TimesheetsHandler {
	return &TimesheetsHandler{
		emailService:         emailService,
		storageService:       storageService,
		errorReporterService: errorReporterService,

		groupsDB:      groupsDB,
		contractorsDB: contractorsDB,
		timesheetsDB:  timesheetsDB,
	}
}

// RegisterTimesheetsHandlers registers the Timesheets handlers.
func (h *TimesheetsHandler) RegisterTimesheetsHandlers(r *mux.Router) {
	r.Methods("POST").Path("/timesheets/request").HandlerFunc(h.RequestTimesheet)
	r.Methods("POST").Path("/timesheets/aggregate").HandlerFunc(h.AggregateTimesheet)
}

// RequestTimesheet sends a timesheet request email to the contractor.
func (h *TimesheetsHandler) RequestTimesheet(w http.ResponseWriter, r *http.Request) {
	// Get the group ID from the query.
	groupID := r.URL.Query().Get("groupID")
	if groupID == "" {
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	// Get the interval from the Group schedule
	group, err := h.groupsDB.GetGroup(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to get group: %w", err))
		return
	}

	// Check if the current time is within the schedule's start and end dates
	requestID, err := getRequestID(group)
	if err != nil {
		if err.Error() == "current time is not within the schedule's start and end dates" {
			return
		} else if err.Error() == "not the correct request time" {
			return
		} else {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to get requestID ID: %w", err))
			return
		}
	}

	// Get contractors from the database
	contractors, err := h.contractorsDB.GetContractors(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to get contractors: %w", err))
		return
	}

	// Send timesheet request emails to the contractors
	for _, contractor := range contractors {
		var requestExists bool
		parsedRequestID := strings.ReplaceAll(strings.ReplaceAll(requestID, "/", "_"), " ", "-")

		if contractor.LastRequests != nil {
			for _, lastRequest := range contractor.LastRequests {
				requestExists = lastRequest.ID == parsedRequestID && lastRequest.Timestamp != 0
			}
		}

		if requestExists {
			continue
		}

		err = h.emailService.SendTimesheetRequestEmail(contractor, requestID)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to send timesheet request email: %w", err))
			continue
		}

		// Update the contractor's last request
		contractor.LastRequests = append(contractor.LastRequests, types.LastRequest{ID: parsedRequestID, Timestamp: 0}) // TODO: should old requests be deleted when schedule changes?

		// Update the contractor in the database
		err = h.contractorsDB.UpdateContractor(contractor)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to update contractor: %w", err))
			continue
		}
	}
}

// AggregateTimesheet aggregates the timesheets from the email attachments.
func (h *TimesheetsHandler) AggregateTimesheet(w http.ResponseWriter, r *http.Request) {
	// Get the timesheet aggregation model from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to read request body: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Unmarshal the timesheet aggregation model
	var timesheetAggregation types.TimesheetAggregation
	err = json.Unmarshal(body, &timesheetAggregation)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to unmarshal timesheet aggregation model: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	emailSubject := fmt.Sprintf("Timesheet %s [%s]", strings.ReplaceAll(strings.ReplaceAll(timesheetAggregation.RequestID, "_", "/"), "-", " "), timesheetAggregation.Contractor.ID)

	// Get the attachments of the email
	attachments, err := h.emailService.GetEmailAttachments(emailSubject)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to get email attachments: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
		return
	}

	// Parse the attachments and process the timesheets
	for _, attachment := range attachments {
		// Get the timesheet extension
		extension := path.Ext(attachment.Filename)

		// Save the timesheet to the storage
		metadata := map[string]string{
			"RequestID":    timesheetAggregation.RequestID,
			"ContractorID": timesheetAggregation.Contractor.ID,
		}

		objectName := fmt.Sprintf("%s-%s_%s%s", timesheetAggregation.Contractor.Name, timesheetAggregation.Contractor.Surname, timesheetAggregation.RequestID, extension)
		timesheetUrl, err := h.storageService.UploadFile(timesheetAggregation.Contractor.GroupID+"/"+objectName, attachment.Content, metadata)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to upload timesheet to storage: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
			return
		}

		timesheet := &types.Timesheet{
			ContractorID: timesheetAggregation.Contractor.ID,
			RequestID:    timesheetAggregation.RequestID,

			StorageURL: timesheetUrl,
		}

		// Add the timesheet to the database
		h.timesheetsDB.AddTimesheet(timesheet)

		// Get contractor
		contractor, err := h.contractorsDB.GetContractor(timesheetAggregation.Contractor.ID)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to get contractor: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
			return
		}

		// Update contractors last request
		for i, lastRequest := range contractor.LastRequests {
			if lastRequest.ID == timesheetAggregation.RequestID {
				contractor.LastRequests[i].Timestamp = time.Now().Unix()
			}
		}

		// Update the contractor in the database
		err = h.contractorsDB.UpdateContractor(contractor)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to update contractor: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
			return
		}

		// Archive the email
		err = h.emailService.ArchiveEmail(emailSubject)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to archive email: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusSeeOther)
			return
		}
	}
}

// getRequestID returns the request ID for the group based on the schedule.
func getRequestID(group *types.Group) (string, error) {

	// Parse the start and end dates from the Schedule
	layout := "2006-01-02" // the layout string used for parsing
	startDate, err := time.Parse(layout, group.Schedule.StartDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse start date: %w", err)
	}

	endDate, err := time.Parse(layout, group.Schedule.EndDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse end date: %w", err)
	}

	// Get the current time in the schedule's timezone
	loc, err := time.LoadLocation(group.Schedule.Timezone)
	if err != nil {
		return "", fmt.Errorf("failed to load location: %w", err)
	}
	now := time.Now().In(loc)

	// Check if the current time is within the schedule's start and end dates
	if now.Before(startDate) || now.After(endDate) {
		return "", fmt.Errorf("current time is not within the schedule's start and end dates") // TODO: inform owner about it
	}

	// Get the current week and month number
	currentYear := now.Year()
	_, currentWeek := now.ISOWeek()
	currentMonth := int(now.Month())

	if group.Schedule.IntervalType == constants.Weeks {
		if currentWeek%group.Schedule.Interval != 0 {
			return "", fmt.Errorf("not the correct request time")
		}
		return fmt.Sprintf("%d/%d %d", max(0, currentWeek-1), currentWeek, currentYear), nil
	} else if group.Schedule.IntervalType == constants.Months {
		if currentMonth%group.Schedule.Interval != 0 {
			return "", fmt.Errorf("not the correct request time")
		}
		return fmt.Sprintf("%d/%d %d", max(0, currentMonth-1), currentMonth, currentYear), nil
	} else {
		return "", fmt.Errorf("invalid interval type")
	}
}
