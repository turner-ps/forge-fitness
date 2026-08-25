package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/internal/testutil"
)

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{}})

	rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsInvalidToken(t *testing.T) {
	_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{}})

	rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "not-a-real-token", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsNonFirebaseIdentity(t *testing.T) {
	identity := &auth.Identity{Provider: "other-provider", Subject: "uid-1", Email: "a@example.com"}
	_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{"token": identity}})

	rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "token", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("non-firebase identity status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsMissingSubjectOrEmail(t *testing.T) {
	for name, identity := range map[string]*auth.Identity{
		"no subject": {Provider: auth.ProviderFirebase, Subject: "", Email: "a@example.com"},
		"no email":   {Provider: auth.ProviderFirebase, Subject: "uid-1", Email: ""},
	} {
		t.Run(name, func(t *testing.T) {
			_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{"token": identity}})
			rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "token", "")
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestRequireAuthFailsWhenVerifierErrors(t *testing.T) {
	_, handler := newTestApp(t, &testutil.FakeVerifier{Reject: errors.New("verifier down")})

	rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "token", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("verifier error status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthFailsWhenNoVerifierConfigured(t *testing.T) {
	_, handler := newTestApp(t, nil)

	rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "token", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("nil verifier status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticatedMeReturnsCurrentUser(t *testing.T) {
	application, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token": testutil.FirebaseIdentity("uid-1", "a@example.com"),
	}})

	rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "token", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response struct {
		User struct {
			ID    int64  `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.User.Email != "a@example.com" {
		t.Fatalf("email = %q, want %q", response.User.Email, "a@example.com")
	}
	if response.User.ID < 1 {
		t.Fatalf("user id = %d, want >= 1", response.User.ID)
	}

	user, err := application.Store.GetUserByFirebaseUID(t.Context(), "uid-1")
	if err != nil {
		t.Fatalf("fetch auto-provisioned user: %v", err)
	}
	if user.Email != "a@example.com" {
		t.Fatalf("provisioned email = %q, want %q", user.Email, "a@example.com")
	}
}

func TestAuthAutoProvisionCreatesLocalUser(t *testing.T) {
	application, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token": testutil.FirebaseIdentity("uid-bravo", "bravo@example.com"),
	}})

	// No user exists yet; a first authenticated call should create the row.
	rec := testutil.DoRequest(t, handler, http.MethodGet, "/me", "token", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	user, err := application.Store.GetUserByFirebaseUID(t.Context(), "uid-bravo")
	if err != nil {
		t.Fatalf("provisioned user missing: %v", err)
	}
	if user.Email != "bravo@example.com" {
		t.Fatalf("email = %q, want %q", user.Email, "bravo@example.com")
	}

	// Second authenticated call should not create a duplicate.
	rec = testutil.DoRequest(t, handler, http.MethodGet, "/me", "token", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", rec.Code, http.StatusOK)
	}
	count := 0
	if err := application.Store.DB.QueryRow(`SELECT count(*) FROM app_user WHERE firebase_uid = 'uid-bravo'`).Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 1 {
		t.Fatalf("user row count = %d, want 1", count)
	}
}

func TestUserCannotAccessAnotherUsersWorkout(t *testing.T) {
	_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token-a": testutil.FirebaseIdentity("uid-a", "a@example.com"),
		"token-b": testutil.FirebaseIdentity("uid-b", "b@example.com"),
	}})

	// User A creates a workout.
	createRec := testutil.DoRequest(t, handler, http.MethodPost, "/workouts", "token-a", `{"name":"A Workout"}`)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create workout status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created struct {
		Workout struct {
			ID int64 `json:"id"`
		} `json:"workout"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created workout: %v", err)
	}

	// User B attempts to fetch A's workout.
	path := "/workouts/" + testutil.Itoa(created.Workout.ID)
	rec := testutil.DoRequest(t, handler, http.MethodGet, path, "token-b", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user workout status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	// User A can fetch their own workout.
	rec = testutil.DoRequest(t, handler, http.MethodGet, path, "token-a", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("owner workout status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListWorkoutsIsUserScoped(t *testing.T) {
	_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token-a": testutil.FirebaseIdentity("uid-a", "a@example.com"),
		"token-b": testutil.FirebaseIdentity("uid-b", "b@example.com"),
	}})

	for _, item := range []struct {
		token string
		name  string
	}{
		{"token-a", "Alpha"},
		{"token-b", "Beta"},
	} {
		rec := testutil.DoRequest(t, handler, http.MethodPost, "/workouts", item.token, `{"name":"`+item.name+`"}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %s: status = %d", item.name, rec.Code)
		}
	}

	rec := testutil.DoRequest(t, handler, http.MethodGet, "/workouts", "token-a", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d", rec.Code)
	}
	var list struct {
		Workouts []store.Workout `json:"workouts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list.Workouts) != 1 || list.Workouts[0].Name != "Alpha" {
		t.Fatalf("user A workouts = %+v, want only Alpha", list.Workouts)
	}

	rec = testutil.DoRequest(t, handler, http.MethodGet, "/workouts", "token-b", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list B status = %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list B: %v", err)
	}
	if len(list.Workouts) != 1 || list.Workouts[0].Name != "Beta" {
		t.Fatalf("user B workouts = %+v, want only Beta", list.Workouts)
	}
}

func TestUserCannotAccessAnotherUsersSession(t *testing.T) {
	application, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token-a": testutil.FirebaseIdentity("uid-a", "a@example.com"),
		"token-b": testutil.FirebaseIdentity("uid-b", "b@example.com"),
	}})

	userA, err := application.Store.UpsertUserFromFirebase(t.Context(), store.UpsertUserFromFirebaseInput{
		FirebaseUID: "uid-a", Email: "a@example.com",
	})
	if err != nil {
		t.Fatalf("seed user A: %v", err)
	}

	workout, err := application.Store.CreateWorkout(t.Context(), store.CreateWorkoutInput{
		UserID: userA.ID, Name: "A Workout",
	})
	if err != nil {
		t.Fatalf("seed workout: %v", err)
	}

	session, err := application.Store.CreateWorkoutSession(t.Context(), store.CreateWorkoutSessionInput{
		UserID: userA.ID, WorkoutID: workout.ID, PerformedAt: &time.Time{},
	})
	if err != nil {
		t.Fatalf("seed session: %v", err)
	}

	path := "/workout-sessions/" + testutil.Itoa(session.ID)
	rec := testutil.DoRequest(t, handler, http.MethodGet, path, "token-b", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user session status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	rec = testutil.DoRequest(t, handler, http.MethodGet, path, "token-a", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("owner session status = %d, want %d", rec.Code, http.StatusOK)
	}
}
