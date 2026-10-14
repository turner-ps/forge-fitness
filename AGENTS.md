# AGENTS.md

Guidance for AI agents and contributors working on forge-fitness.

## Project
Go backend (REST/JSON) for a mobile workout tracking app. Users manage accounts, build routines from an external exercise catalog or their own custom exercises, log workouts (sets, reps, weight, time), and view progress per routine and per exercise. Full design: [`plan.md`](plan.md). Read it before starting a feature.

Module: `github.com/turner-ps/forge-fitness`. Stack: Go 1.26, chi v5, pgx v5 (`pgxpool`), goose migrations, sqlc, `log/slog`, Postgres 18.

## Status
The repo is currently a shell: `/heartbeat` only. The target layout below is built in phases (see `plan.md`). Check what actually exists before assuming a package is there.

## Layout (target)
```
cmd/api/            main.go: wiring only (config, deps, server, graceful shutdown)
internal/config/    env -> typed Config
internal/platform/  database, httpx, logger, middleware (shared infrastructure)
internal/<domain>/  user, auth, catalog, exercise, routine, workout, progress
internal/router/    chi router; mounts each domain's routes
migrations/         goose SQL, embedded via migrations/fs.go
```
Each domain package: `model.go`, `service.go`, `repository.go` (interface), `postgres.go`, `handler.go`, tests.

## Commands
Targets below are the intended Makefile; add them if missing.
```
docker compose up -d db test_db   # Postgres on :5432, test DB on :5433
make run                          # go run ./cmd/api
make test                         # go test -race ./...
make lint                         # go vet + golangci-lint
make migrate                      # goose up
make sqlc                         # regenerate typed queries
```
Config comes from env vars (`.env` optional locally; see `.env.example`). Never commit `.env`.

## Architecture rules
- Layering: handler -> service -> repository interface -> postgres impl. Handlers do HTTP only (decode, validate, call service, encode). Services hold business logic and know nothing about HTTP or SQL. Repositories do SQL only.
- Interfaces are defined by the consumer (the service package), not the implementer.
- Domains do not import each other's internals; talk through small interfaces.
- The exercise catalog is an external, read-only DB. Access it only through `catalog.Provider`. Never write to it, never FK to it. Store catalog ids as opaque strings plus a name snapshot.
- Every user-owned query filters by `user_id`. Ownership is enforced in the repository, not trusted from the handler.
- Pass `context.Context` as the first parameter through every layer, including DB calls.
- Stateless API: no in-process session state or local files, so instances scale horizontally.

## Conventions
- Errors: return them, wrap with `fmt.Errorf("pkg:action %w", err)` (matches existing style, e.g. `"db:open %w"`). Use sentinel/typed domain errors (`ErrNotFound`, `ErrConflict`, ...) and map them to HTTP codes in one place (`platform/httpx`). No `panic` or `log.Fatal` outside `main`.
- Responses: use `httpx` helpers. Error body is `{"error":{"code","message","details"}}`. Never leak internal errors or SQL to clients.
- Logging: `slog` with request id. Never log passwords, tokens, or full request bodies.
- IDs are UUIDs (client-generated UUIDv7 for workout data so offline mobile clients can sync idempotently). Times are UTC `timestamptz`, ISO-8601 on the wire. Weights are stored in kg; convert per user unit preference at the edge.
- Auth: Argon2id passwords; 15-min JWT access tokens; hashed, rotating refresh tokens in `sessions`. Never store or log raw refresh tokens.
- Migrations are append-only goose files (`NNNNN_name.sql`, with Up and Down). Never edit a migration that has been merged; add a new one.
- Style: `gofmt`, doc comment on every package (`// Package x ...`), small functions, table-driven tests.
- API is versioned under `/api/v1`. Additive changes only within a version.

## Testing
- Services: unit tests with fake repositories.
- Repositories: integration tests against `test_db` (port 5433), one transaction per test, rolled back.
- Handlers: `httptest` with a real router and fake services.
- Add or update tests with every behavior change. Run `go test -race ./...` before finishing.

## Git
- Branch from `develop`; PRs target `main` through `develop`.
- Commit message format: `(type): description` (types: feat, fix, task, docs, test, refactor), optional bullet body.
- Do not commit `.env`, `database/`, or secrets.

## Do / Don't
- Do read `plan.md` and the neighboring package before adding code; match its patterns.
- Do keep changes scoped to the task; ask before changing the data model or API shape in `plan.md`.
- Don't add an ORM, global state, or `init()` side effects.
- Don't add dependencies without a clear need; run `go mod tidy` after.
- Don't hard-delete custom exercises that logged history references; soft-delete them.
