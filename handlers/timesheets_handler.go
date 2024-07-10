package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"job_sender/core"
	"job_sender/types"

	"github.com/gorilla/mux"
)

type TimesheetsHandler struct {
	emailService         *core.EmailService
	timesheetsDB         *core.TimesheetsDatabaseService
	storageService       *core.StorageService
	errorReporterService *core.ErrorReporterService
}

// NewTimesheetsHandler creates a new TimesheetsHandler.
func NewTimesheetsHandler(emailService *core.EmailService, timesheetsDB *core.TimesheetsDatabaseService, storageService *core.StorageService, errorReporterService *core.ErrorReporterService) *TimesheetsHandler {
	return &TimesheetsHandler{
		emailService:         emailService,
		timesheetsDB:         timesheetsDB,
		storageService:       storageService,
		errorReporterService: errorReporterService,
	}
}

// RegisterTimesheetsHandlers registers the Timesheets handlers.
func (h *TimesheetsHandler) RegisterTimesheetsHandlers(r *mux.Router) {
	r.Methods("POST").Path("/timesheets/aggregator").HandlerFunc(h.CreateTimesheetAggregator)
}

// CreateTimesheetAggregator creates a new timesheet aggregator.
func (h *TimesheetsHandler) CreateTimesheetAggregator(w http.ResponseWriter, r *http.Request) {
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
	attachments, err := h.emailService.GetEmailAttachments("Timesheet") // TODO: Przy wysyłaniu emaila o timesheet, w tytule dodać ReferenceID, i tutaj po tym szukać
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
