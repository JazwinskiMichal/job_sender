package interfaces

import (
	"net/http"

	"job_sender/types"
)

type ITemplateService interface {

	// ParseTemplate creates a template that applies a given file to the body of the base template.
	ParseTemplate(filename string) (*types.AppTemplate, error)

	// ExecuteTemplate applies the template to the response writer.
	ExecuteTemplate(tmpl *types.AppTemplate, w http.ResponseWriter, r *http.Request, data interface{}, userInfo *types.LoggedUserInfo) error

	// ShowError displays an error message to the user.
	ShowError(tmpl *types.AppTemplate, w http.ResponseWriter, r *http.Request, errorMessage string) error
}
