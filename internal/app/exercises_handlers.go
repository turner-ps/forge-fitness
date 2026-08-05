package app

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/httpjson"
)

func (a *Application) GetExercises(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 50
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			a.badRequest(w, "invalid exercise limit")
			return
		}
		limit = parsedLimit
	}

	exercises, err := a.Store.GetExercises(r.Context(), search, limit)
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"exercises": exercises})
	if err != nil {
		a.Logger.Printf("write exercises response: %v", err)
	}
}

func (a *Application) GetExerciseByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		a.badRequest(w, "invalid exercise id")
		return
	}

	exercise, err := a.Store.GetExerciseByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "exercise not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"exercise": exercise})
	if err != nil {
		a.Logger.Printf("write exercise response: %v", err)
	}
}
