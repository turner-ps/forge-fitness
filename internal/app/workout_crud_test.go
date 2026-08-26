package app

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/testutil"
)

func workoutTestApp(t *testing.T, uid, email string) (*Application, http.Handler) {
	return newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token": testutil.FirebaseIdentity(uid, email),
	}})
}

func createTestWorkout(t *testing.T, handler http.Handler, token, name string) int64 {
	t.Helper()
	rec := testutil.DoRequest(t, handler, http.MethodPost, "/workouts", token, `{"name":"`+name+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workout status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created struct {
		Workout struct {
			ID int64 `json:"id"`
		} `json:"workout"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created workout: %v", err)
	}
	return created.Workout.ID
}

func TestUpdateWorkout(t *testing.T) {
	_, handler := workoutTestApp(t, "uid-a", "a@example.com")
	id := createTestWorkout(t, handler, "token", "Old Name")

	rec := testutil.DoRequest(t, handler, http.MethodPatch, "/workouts/"+testutil.Itoa(id), "token", `{"name":"New Name"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var response struct {
		Workout struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"workout"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if response.Workout.Name != "New Name" {
		t.Fatalf("name = %q, want %q", response.Workout.Name, "New Name")
	}
	if response.Workout.ID != id {
		t.Fatalf("id = %d, want %d", response.Workout.ID, id)
	}
}

func TestUpdateWorkoutEmptyNameRejected(t *testing.T) {
	_, handler := workoutTestApp(t, "uid-a", "a@example.com")
	id := createTestWorkout(t, handler, "token", "Old Name")

	rec := testutil.DoRequest(t, handler, http.MethodPatch, "/workouts/"+testutil.Itoa(id), "token", `{"name":"   "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty name status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateWorkoutCrossUserNotFound(t *testing.T) {
	_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token-a": testutil.FirebaseIdentity("uid-a", "a@example.com"),
		"token-b": testutil.FirebaseIdentity("uid-b", "b@example.com"),
	}})
	id := createTestWorkout(t, handler, "token-a", "A Workout")

	rec := testutil.DoRequest(t, handler, http.MethodPatch, "/workouts/"+testutil.Itoa(id), "token-b", `{"name":"Hijack"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user update status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteWorkout(t *testing.T) {
	_, handler := workoutTestApp(t, "uid-a", "a@example.com")
	id := createTestWorkout(t, handler, "token", "Disposable")

	rec := testutil.DoRequest(t, handler, http.MethodDelete, "/workouts/"+testutil.Itoa(id), "token", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	getRec := testutil.DoRequest(t, handler, http.MethodGet, "/workouts/"+testutil.Itoa(id), "token", "")
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want %d", getRec.Code, http.StatusNotFound)
	}
}

func TestDeleteWorkoutCrossUserNotFound(t *testing.T) {
	_, handler := newTestApp(t, &testutil.FakeVerifier{Identities: map[string]*auth.Identity{
		"token-a": testutil.FirebaseIdentity("uid-a", "a@example.com"),
		"token-b": testutil.FirebaseIdentity("uid-b", "b@example.com"),
	}})
	id := createTestWorkout(t, handler, "token-a", "A Workout")

	rec := testutil.DoRequest(t, handler, http.MethodDelete, "/workouts/"+testutil.Itoa(id), "token-b", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user delete status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
