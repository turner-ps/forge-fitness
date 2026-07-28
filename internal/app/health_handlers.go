package app

import (
	"net/http"

	"github.com/turner-ps/forge-fitness/utils"
)

func (a *Application) Heartbeat(w http.ResponseWriter, r *http.Request) {
	err := utils.WriteJSON(w, http.StatusOK, utils.Envelope{"success": "status is available"})
	if err != nil {
		a.Logger.Printf("write heartbeat response: %v", err)
	}
}
