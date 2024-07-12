package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"job_sender/core"
	"job_sender/types"

	"github.com/gorilla/mux"
)

type TimesheetsHandler struct {
	emailService         *core.EmailService
	constractorsDB       *core.ContractorsDatabaseService
	timesheetsDB         *core.TimesheetsDatabaseService
	storageService       *core.StorageService
	errorReporterService *core.ErrorReporterService
}

// NewTimesheetsHandler creates a new TimesheetsHandler.
func NewTimesheetsHandler(emailService *core.EmailService, contractorsDB *core.ContractorsDatabaseService, timesheetsDB *core.TimesheetsDatabaseService, storageService *core.StorageService, errorReporterService *core.ErrorReporterService) *TimesheetsHandler {
	return &TimesheetsHandler{
		emailService:         emailService,
		constractorsDB:       contractorsDB,
		timesheetsDB:         timesheetsDB,
		storageService:       storageService,
		errorReporterService: errorReporterService,
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

	// TODO: This check for enen week should also be aligned for settings for the schedule
	// Get the current year and week number
	_, week := time.Now().ISOWeek()

	// Check if it's an even week
	if week%2 != 0 {
		return
	}

	// Get contractors from the database
	contractors, err := h.constractorsDB.ListContractors(groupID)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to get contractors: %w", err))
		return
	}

	// Send timesheet request emails to the contractors
	for _, contractor := range contractors {
		err = h.emailService.SendTimesheetRequestEmail(contractor.Email, contractor.ID, fmt.Sprintf("%02d/%02d", week-1, week))
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to send timesheet request email: %w", err))
			return
		}
	}
}

// AggregateTimesheet aggregates the timesheets from the email attachments.
func (h *TimesheetsHandler) AggregateTimesheet(w http.ResponseWriter, r *http.Request) {
	// Get the timesheet aggregation model from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to read request body: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Unmarshal the timesheet aggregation model
	var timesheetAggregation types.TimesheetAggregation
	err = json.Unmarshal(body, &timesheetAggregation)
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to unmarshal timesheet aggregation model: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Get the attachments of the email
	attachments, err := h.emailService.GetEmailAttachments(fmt.Sprintf("[Contractor ID: %s]", timesheetAggregation.ContractorID))
	if err != nil {
		h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to get email attachments: %w", err))
		http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
		return
	}

	// Parse the attachments and process the timesheets
	for _, attachment := range attachments {
		// Get the timesheet extension
		extension := path.Ext(attachment.Filename)

		// Save the timesheet to the storage
		timesheetUrl, err := h.storageService.UploadFile(fmt.Sprintf("%s_%s_timesheet%s", timesheetAggregation.GroupID, timesheetAggregation.ContractorID, extension), attachment.Content)
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to upload timesheet to storage: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
			return
		}

		timesheet := &types.Timesheet{
			ID:           attachment.Filename,
			ContractorID: timesheetAggregation.ContractorID,
			GroupID:      timesheetAggregation.GroupID,
			StorageURL:   timesheetUrl,
		}

		// Add the timesheet to the database
		h.timesheetsDB.AddTimesheet(timesheet) // TODO: Tutaj tez zapisywac trzeba mądrze, bo aktualnie za kazdym odswieżeniem Contractors list jest wyzwalany task do pobierania timesheet z maila

		// Archive the email
		err = h.emailService.ArchiveEmail("Timesheet")
		if err != nil {
			h.errorReporterService.ReportError(w, r, fmt.Errorf("failed to archive email: %w", err))
			http.Redirect(w, r, "/somethingWentWrong", http.StatusFound)
			return
		}
	}
}
