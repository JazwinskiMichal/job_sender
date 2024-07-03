package middlewares

import (
	"fmt"
	"net/http"

	"job_sender/core"
)

type panicRecoverMiddleware struct {
	errorReporterService *core.ErrorReporterService
}

func NewPanicRecoverMiddleware(errorReporterService *core.ErrorReporterService) *panicRecoverMiddleware {
	return &panicRecoverMiddleware{
		errorReporterService: errorReporterService,
	}
}

func (h *panicRecoverMiddleware) PanicRecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				switch x := rec.(type) {
				case string:
					err = fmt.Errorf(x)
				case error:
					err = x
				default:
					err = fmt.Errorf("unknown error")
				}
				// Use the existing ReportError method to log the panic.
				h.errorReporterService.ReportError(w, r, err)

				// Respond to the client.
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
