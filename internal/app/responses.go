package app

import (
	"net/http"

	"github.com/turner-ps/forge-fitness/utils"
)

func (a *Application) badRequest(w http.ResponseWriter, message string) {
	a.errorResponse(w, http.StatusBadRequest, message)
}

func (a *Application) notFound(w http.ResponseWriter, message string) {
	a.errorResponse(w, http.StatusNotFound, message)
}

func (a *Application) serverError(w http.ResponseWriter, err error) {
	a.Logger.Printf("server error: %v", err)
	a.errorResponse(w, http.StatusInternalServerError, "internal server error")
}

func (a *Application) errorResponse(w http.ResponseWriter, status int, message string) {
	err := utils.WriteJSON(w, status, utils.Envelope{"error": message})
	if err != nil {
		a.Logger.Printf("write error response: %v", err)
	}
}
