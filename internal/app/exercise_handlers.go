package app

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/utils"
)

func (a *Application) GetExerciseByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		a.badRequest(w, "invalid exercise id")
		return
	}

	exercise, err := store.GetExerciseByID(r.Context(), a.DB, id)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "exercise not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"exercise": exercise})
	if err != nil {
		a.Logger.Printf("write exercise response: %v", err)
	}
}
