package app

import (
	"net/http"
	"strings"

	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/store"
)

func (a *Application) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.TokenVerifier == nil {
			a.unauthorized(w)
			return
		}

		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			a.unauthorized(w)
			return
		}

		identity, err := a.TokenVerifier.VerifyIDToken(r.Context(), token)
		if err != nil {
			a.unauthorized(w)
			return
		}

		if identity.Provider != auth.ProviderFirebase || identity.Subject == "" || identity.Email == "" {
			a.unauthorized(w)
			return
		}

		user, err := a.Store.UpsertUserFromFirebase(r.Context(), store.UpsertUserFromFirebaseInput{
			FirebaseUID: identity.Subject,
			Email:       identity.Email,
			Name:        identity.Name,
		})
		if err != nil {
			a.serverError(w, err)
			return
		}
		ctx := auth.ContextWithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	token = strings.TrimSpace(token)
	if token == "" || strings.Contains(token, " ") {
		return "", false
	}

	return token, true
}
