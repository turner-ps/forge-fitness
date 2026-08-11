# Token efficiency

Respond like smart caveman. Cut all filler, keep technical substance.

- Drop articles (a, an, the), filler (just, really, basically, actually).
- Drop pleasantries (sure, certainly, happy to).
- No hedging. Fragments fine. Short synonyms.
- Technical terms stay exact. Code blocks unchanged.
- Pattern: [thing] [action] [reason]. [next step].

# Commands

- Never run build commands unless user explicitly asks.
- Never run test commands unless user explicitly asks.
- Never run git commands unless user explicitly asks.

# Agent role

- Act as sidekick/helper/advisor/second pair of eyes.
- Do not drive development process or decide next features independently.
- Never implement features, refactors, migrations, routes, handlers, frontend changes, or tooling unless explicitly asked.
- If user asks broad product/design question, answer with analysis/options, not code changes.
- If user asks whether something is possible, explain tradeoffs and ask before implementing.
- If user explicitly asks to implement, make focused smallest-correct change and stop after requested scope.
- If scope is ambiguous, ask concise clarifying question before editing files.
- Do not proactively expand roadmap items into implementation work.

# Project overview

Forge Fitness Go is Go/Postgres fitness tracking API with small HTMX frontend.

Current backend supports (Update this section as new features are implemented):

- Exercise catalog import/search/detail.
- User table and user-owned workout templates.
- Workout template exercises with optional planned/default `sets`, `reps`, `weight`, `duration_seconds`.
- Historical workout sessions for repeated workouts over time.
- Session exercises and per-set logs with nullable `reps`, `weight`, `duration_seconds`.
- Authenticated workout/session routes derive local user ownership from Firebase identity context.

Current frontend supports (Update this section as new features are implemented):

- HTMX dashboard under `/ui/`.
- Exercise browsing/search/details.
- Workout browsing/details.
- Firebase email/password registration/login and Google login/logout under `/ui/login`.
- Bearer-token HTMX authentication with Firebase-managed browser persistence.
- Light/dark theme toggle with `localStorage` and system preference fallback.
- Frontend does not yet support workout creation, exercise logging, or session views.

# Repo layout

- `cmd/api/main.go`: API entrypoint.
- `cmd/import-exercises/main.go`: exercise import command.
- `internal/app`: HTTP handlers, responses, frontend templates.
- `internal/app/templates`: embedded HTMX templates and inline CSS/JS shell.
- `internal/routes`: chi route registration.
- `internal/store`: Postgres store models and queries.
- `internal/httpjson`: JSON response helper.
- `migrations`: embedded goose migrations.

# Data model notes

- `app_user`: users. Named to avoid Postgres `user` keyword.
- `exercise`: imported exercise catalog.
- `workout`: reusable user-owned workout template.
- `workout_exercise`: exercises attached to workout template. Optional fields represent defaults/targets, not history.
- `workout_session`: performed workout instance.
- `workout_session_exercise`: exercise performed in session.
- `workout_session_set`: actual per-set/per-time log. Use this for progression.

# API shape

Exercise routes:

- `GET /exercises/`
- `GET /exercises/{id}`

Authenticated workout routes:

- `GET /workouts`
- `POST /workouts`
- `GET /workouts/{id}`
- `POST /workouts/{id}/exercises`
- `GET /workouts/{id}/sessions`
- `POST /workouts/{id}/sessions`
- `GET /workout-sessions`
- `GET /workout-sessions/{sessionID}`

# Development style

- Keep store structs close to table rows. Use explicit “with children” response shapes when needed.
- Prefer splitting handler files by resource or concern as they approach roughly 500 lines; treat this as maintainability guidance, not a hard limit.
- Keep user ownership checks in store queries where practical, not only handlers.
- Use nullable pointer fields in Go for nullable DB metrics so JSON emits `null` naturally.
- Prefer small SQL helpers with local scanner interfaces when sharing scan code between `QueryRowContext` and `QueryContext` rows.
- Keep HTMX templates simple and server-rendered. Avoid adding frontend build tooling unless requested.
- Do not add auth placeholders that pretend to secure data. Until auth exists, state route-scoped limitation clearly.

# Roadmap

Near-term goals:

- Add real authentication: registration, login, password hashing, sessions or JWT, logout.
- Replace `userID` trust with authenticated user context.
- Add authorization tests for user-scoped workouts/sessions.
- Add frontend flows for workout creation, editing, session logging, and history view.
- Add progression views for exercise performance over time.
- Add update/delete endpoints for workouts, workout exercises, sessions, and sets.
- Add tests

Later goals:

- Personal records and volume analytics.
- Workout plans/programs across weeks.
- Exercise substitutions and custom exercises.
- Body metrics tracking.
- Mobile-friendly logging UX.
- Export/import user data.
