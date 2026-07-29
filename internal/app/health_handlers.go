package app

import (
	"net/http"

	"github.com/turner-ps/forge-fitness/internal/httpjson"
)

func (a *Application) Heartbeat(w http.ResponseWriter, r *http.Request) {
	err := httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"success": "status is available"})
	if err != nil {
		a.Logger.Printf("write heartbeat response: %v", err)
	}
}
