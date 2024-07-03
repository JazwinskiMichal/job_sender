package handlers

import (
	"net/http"

	"job_sender/core"
	constants "job_sender/utils/constants"

	"github.com/gorilla/mux"
)

type SomethingWentWrongHandler struct {
	templateService *core.TemplateService
}

func NewSomethingWentWrongHandler(templateService *core.TemplateService) *SomethingWentWrongHandler {
	return &SomethingWentWrongHandler{
		templateService: templateService,
	}
}

func (h *SomethingWentWrongHandler) RegisterSomethingWentWrongHandlers(r *mux.Router) {
	r.Methods("GET").Path("/somethingWentWrong").Handler(http.HandlerFunc(h.showSomethingWentWrong))
}

// showSomethingWentWrong displays the something went wrong page.
func (h *SomethingWentWrongHandler) showSomethingWentWrong(w http.ResponseWriter, r *http.Request) {
	somethingWentWrongTmpl, err := h.templateService.ParseTemplate(constants.TemplateSomethingWentWrong)
	if err != nil {
		http.Error(w, "could not parse something went wrong template", http.StatusInternalServerError)
		return
	}

	err = h.templateService.ExecuteTemplate(somethingWentWrongTmpl, w, r, nil, nil)
	if err != nil {
		http.Error(w, "could not execute something went wrong template", http.StatusInternalServerError)
		return
	}
}
