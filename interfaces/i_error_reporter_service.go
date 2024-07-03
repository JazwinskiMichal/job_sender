package interfaces

import "net/http"

type IErrorReporterService interface {

	// ReportError sends an error report to the error reporting service.
	ReportError(w http.ResponseWriter, r *http.Request, err error)
}
