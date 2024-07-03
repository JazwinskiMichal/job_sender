package core

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"job_sender/interfaces"
	"job_sender/types"
	constants "job_sender/utils/constants"
)

type TemplateService struct{}

// Ensure TemplateService implements the ITemplateService interface.
var _ interfaces.ITemplateService = &TemplateService{}

// NewTemplateService creates a new TemplateService.
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// ParseTemplate creates a template that applies a given file to the body of the base template.
func (s *TemplateService) ParseTemplate(filename string) (*types.AppTemplate, error) {
	tmpl := template.Must(template.ParseFiles(fmt.Sprintf("%s/%s", constants.TemplatesDir, constants.TemplatesBaseName)))

	// Put the named file into a template called "body"
	path := filepath.Join(constants.TemplatesDir, filename)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not app template: %w", err)
	}
	template.Must(tmpl.New("body").Parse(string(b)))

	return &types.AppTemplate{
		Tmpl: tmpl,
	}, nil
}

// ExecuteTemplate applies the template to the response writer.
func (s *TemplateService) ExecuteTemplate(tmpl *types.AppTemplate, w http.ResponseWriter, r *http.Request, data interface{}, userInfo *types.LoggedUserInfo) error {
	// Check if userInfo is nil and handle accordingly
	if userInfo == nil {
		userInfo = &types.LoggedUserInfo{
			Email:      "",
			IsLoggedIn: false,
			IsVerified: false,
		}
	}

	d := struct {
		Data       interface{}
		Email      string
		IsLoggedIn bool
		IsVerified bool
	}{
		Data:       data,
		Email:      userInfo.Email,
		IsLoggedIn: userInfo.IsLoggedIn,
		IsVerified: userInfo.IsVerified,
	}

	err := tmpl.Tmpl.Execute(w, d)
	if err != nil {
		// Log the error or handle it as needed
		return fmt.Errorf("could not write template: %w", err)
	}
	return nil
}

// ShowError displays an error message to the user.
func (s *TemplateService) ShowError(tmpl *types.AppTemplate, w http.ResponseWriter, r *http.Request, errorMessage string) error {
	data := map[string]interface{}{
		"ErrorMessage": errorMessage,
	}
	err := s.ExecuteTemplate(tmpl, w, r, data, nil)
	if err != nil {
		return fmt.Errorf("could not show error message: %w", err)
	}

	return nil
}
