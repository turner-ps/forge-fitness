package app

import (
	"net/http"

	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/httpjson"
)

type meResponseUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (a *Application) GetMe(w http.ResponseWriter, r *http.Request) {
	user, err := auth.RequireUser(r.Context())
	if err != nil {
		a.unauthorized(w)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{
		"user": meResponseUser{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	})
	if err != nil {
		a.Logger.Printf("get me response: %v", err)
	}
}
