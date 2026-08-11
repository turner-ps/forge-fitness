package app

import (
	"net/http"

	"github.com/turner-ps/forge-fitness/internal/httpjson"
)

func (a *Application) FrontendLogin(w http.ResponseWriter, r *http.Request) {
	a.renderFrontend(w, r, "login-page", "login-content", frontendData{
		Title:  "Sign in | Forge Fitness",
		Active: "login",
	})
}

func (a *Application) FrontendAuthConfig(w http.ResponseWriter, _ *http.Request) {
	if a.FirebaseWebConfig.APIKey == "" {
		a.notFound(w, "authentication is not configured")
		return
	}

	if err := httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"firebase": a.FirebaseWebConfig}); err != nil {
		a.Logger.Printf("write frontend auth config: %v", err)
	}
}
