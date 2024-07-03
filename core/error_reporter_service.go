package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"job_sender/interfaces"
	constants "job_sender/utils/constants"

	"cloud.google.com/go/errorreporting"
)

type ErrorReporterService struct {
	LogWriter   *os.File
	errorClient *errorreporting.Client
}

// Ensure ErrorReporterService implements IErrorReporterService.
var _ interfaces.IErrorReporterService = &ErrorReporterService{}

func NewErrorReporterService(projectID string) *ErrorReporterService {
	logWriter := os.Stderr
	ctx := context.Background()

	errorClient, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName: constants.AppName,
		OnError: func(err error) {
			fmt.Fprintf(logWriter, "Could not log error: %v", err)
		},
	})
	if err != nil {
		fmt.Fprintf(logWriter, "errorreporting.NewClient: %v", err)
		return nil
	}

	return &ErrorReporterService{
		LogWriter:   logWriter,
		errorClient: errorClient,
	}
}

// ReportError logs the error and reports it to Error Reporting.
func (h *ErrorReporterService) ReportError(w http.ResponseWriter, r *http.Request, err error) {
	// Log the error and report it to Error Reporting.
	fmt.Fprintf(h.LogWriter, "Error reported: message: %s, underlying err: %+v\n", err.Error(), err)
	h.errorClient.Report(errorreporting.Entry{
		Error: err,
		Req:   r,
		Stack: debug.Stack(),
	})
	h.errorClient.Flush()
}
