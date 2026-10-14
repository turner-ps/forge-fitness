# Forge Fitness Backend — Plan

## Context
`forge-fitness` is a Go shell (chi + pgx + goose, Postgres via docker-compose) with only a `/heartbeat` route. Goal: a scalable, maintainable backend for a **mobile** workout tracking app. Users manage accounts, build routines from an external exercise catalog or their own custom exercises, log workouts (sets/reps/weight/time), and view progress over time.

Decisions made: mobile-first API; JWT access + rotating refresh tokens. Conventions for contributors and agents live in `AGENTS.md`.

## Architecture
Modular monolith, package-per-domain, layered inside each domain: **handler → service → repository (interface) → postgres impl**. Domains depend on interfaces they define, never on each other's internals. Splitting a domain into its own service later stays cheap.

```
cmd/api/main.go            # wiring only: config, deps, server, graceful shutdown
internal/
  config/                  # env -> typed Config (fail fast)
  platform/
    database/              # pgxpool, Ping, goose migrate (from internal/store)
    httpx/                 # WriteJSON, error responses, decode+validate, pagination
    logger/                # slog setup
    middleware/            # request id, slog access log, recover, auth, rate limit, CORS
  user/                    # accounts: register, get, update, delete
  auth/                    # login, logout, refresh, sessions, password hashing, JWT
  catalog/                 # Provider interface for the EXTERNAL exercise DB + fake impl
  exercise/                # user custom exercises CRUD
  routine/                 # routines + routine_exercises CRUD
  workout/                 # logged sessions, session_exercises, sets (+ batch sync)
  progress/                # read-only aggregate queries
  router/                  # chi router, mounts each domain's Routes(r)
migrations/                # goose SQL, embedded
```
Each domain: `model.go`, `service.go`, `repository.go` (interface), `postgres.go`, `handler.go`, `*_test.go`.

Existing code moves: `main.go` → `cmd/api/`; `internal/app`+`routes` → `router` + per-domain handlers; `internal/store` → `platform/database`; `utils.WriteJSON` → `platform/httpx`.

Stack: chi v5, pgx v5 `pgxpool` (replace `database/sql` stdlib wrapper), **sqlc** for typed queries (no ORM), goose, `log/slog`, `golang.org/x/crypto/argon2`, `golang-jwt/jwt/v5`, `go-playground/validator`.

## Cleanup of the existing shell (Phase 0)
- `utils.WriteJSON` calls `log.Fatalf` on marshal error → return the error instead; drop empty-indent `MarshalIndent` for `json.Marshal`.
- `store.Open` requires a `.env` file and never pings → make `.env` optional, `Ping` on startup, replace `fmt.Println` with slog.
- `Heartbeat` panics on write error → log instead. Add `/healthz` (liveness) and `/readyz` (DB ping).
- `go mod tidy` (godotenv wrongly marked indirect); fix invalid `./database/` line in `.gitignore`.
- Add `.env.example`, Makefile, Dockerfile (multi-stage), GitHub Actions CI.

## Data model (Postgres; UUIDs, `timestamptz`, weights stored in kg)
Client-generated UUIDv7 ids for workout data so offline clients can create records and sync idempotently.

| Table | Key columns |
|---|---|
| `users` | id, email (citext unique), password_hash, display_name, unit_system (metric/imperial), created_at, updated_at |
| `sessions` | id, user_id FK cascade, refresh_token_hash, device_label, expires_at, revoked_at, replaced_by, created_at |
| `custom_exercises` | id, user_id FK cascade, name, description, equipment text[], difficulty enum(beginner,intermediate,advanced), primary_muscles text[], secondary_muscles text[], tracking_type enum(weight_reps, reps_only, duration, distance_duration), timestamps |
| `routines` | id, user_id FK cascade, name, description, timestamps |
| `routine_exercises` | id, routine_id FK cascade, position, catalog_exercise_id text NULL, custom_exercise_id uuid NULL, target_sets, target_reps, target_weight_kg, rest_seconds, notes; CHECK exactly one of the two exercise refs |
| `workout_sessions` | id, user_id FK cascade, routine_id FK SET NULL, name (snapshot), started_at, ended_at, notes, timestamps |
| `session_exercises` | id, session_id FK cascade, position, catalog_exercise_id / custom_exercise_id (same CHECK), exercise_name (snapshot, so history survives catalog changes or custom exercise deletion) |
| `sets` | id, session_exercise_id FK cascade, set_number, reps, weight_kg numeric, duration_s, distance_m, rpe, is_warmup, completed_at |

Indexes: `(user_id, started_at desc)` on sessions; `(session_exercise_id)`; exercise-ref columns on `session_exercises` for per-exercise progress; GIN on muscle arrays.
Custom exercises referenced by history are soft-deleted (`deleted_at`) so logged data keeps its meaning. Account deletion is a hard delete via cascade.

## External exercise catalog
`catalog.Provider` interface: `Get(ctx, id)`, `List(ctx, filter{muscle, equipment, difficulty, q}, page)`. Ship an in-memory fake now; real implementation (read-only Postgres connection or HTTP client) plugs in later without touching other domains. Catalog ids are opaque strings stored by reference; names snapshotted at log time.

## Auth
- Argon2id password hashes.
- Access JWT (15 min, HS256 secret from config, `sub`=user id, `sid`=session id).
- Opaque 256-bit refresh token, stored hashed (SHA-256), 60-day expiry, **rotated on every refresh**; replay of a rotated token revokes the whole session chain.
- Sign out = revoke that session; `POST /auth/logout-all` revokes all.
- Delete account requires current password; cascades all data.
- Rate-limit `/auth/*`. Uniform "invalid credentials" errors.
- `user_identities` table deferred (room for Sign in with Apple/Google).

## API (REST, JSON, prefix `/api/v1`)
- Auth: `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/logout-all`
- Me: `GET|PATCH|DELETE /me`
- Catalog: `GET /catalog/exercises`, `GET /catalog/exercises/{id}`
- Custom exercises: `GET|POST /exercises`, `GET|PATCH|DELETE /exercises/{id}`
- Routines: `GET|POST /routines`, `GET|PATCH|DELETE /routines/{id}`, `PUT /routines/{id}/exercises` (replace ordered list)
- Workouts: `GET|POST /workouts`, `GET|PATCH|DELETE /workouts/{id}`, `PUT /workouts/{id}/exercises/{eid}/sets/{sid}` (idempotent upsert), `DELETE` set, `POST /workouts/sync` (batch upsert for offline clients)
- Progress: `GET /progress/routines/{id}?from&to&bucket=day|week|month` (total volume, sessions, duration); `GET /progress/exercises/{ref}?...` (max weight, est. 1RM via Epley, volume, reps, PRs)
- Conventions: cursor pagination, ISO-8601 UTC times, `Idempotency-Key` honored on POSTs, uniform error body `{error:{code,message,details}}`, ownership enforced in every repo query (`WHERE user_id = $1`), warmup sets excluded from volume.

Progress is computed with SQL aggregates over indexed data first; add rollup tables only when measurements justify it.

## Cross-cutting
- Stateless API, config via env, graceful shutdown → horizontally scalable behind a load balancer.
- slog JSON logs with request id; timeouts on server and DB calls via `context`.
- Testing: service unit tests with fake repos; repository integration tests against `test_db` (docker-compose, port 5433) with per-test transactions; handler tests via `httptest`; table-driven style. `go vet`, `staticcheck`/`golangci-lint`, `go test -race ./...` in CI.

## Phases
0. Cleanup + restructure + config/logging/health (above)
1. `user` + `auth` (migrations 1–2, tokens, middleware)
2. `catalog` interface/fake + `exercise` custom CRUD
3. `routine` CRUD
4. `workout` logging + idempotent upsert + batch sync
5. `progress` queries
6. Hardening: rate limits, OpenAPI spec (`api/openapi.yaml`), Dockerfile, CI, load smoke test

## Verification
- `docker compose up -d db test_db`, `make migrate`, `make run`; `curl /readyz` → 200.
- `make test` (unit + integration), `make lint`.
- End-to-end curl script (`scripts/smoke.sh`): register → login → create custom exercise → create routine → log workout → fetch progress → refresh → logout → delete account; each step asserts status codes and that data is gone after deletion.
