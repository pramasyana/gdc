# GDC Task Management API

A multi-user task management REST API in Go, structured as a **feature-based
(per-module) Clean Architecture**. Each user owns their tasks and can assign them
to other members of their team. The service implements JWT authentication,
idempotent task creation, fully transactional assignment, a uniform error
envelope and structured JSON logging.

## Features

- **JWT auth** — access + rotating refresh tokens, Redis-backed revocation.
- **Idempotent `POST /tasks`** — an `Idempotency-Key` guarantees exactly one
  task, even under concurrent retries.
- **Transactional assignment** — assignee update, audit log and notification
  commit or roll back together.
- **Uniform envelope** — one success/error shape for every response.
- **Structured logging** — one JSON line per request via `log/slog`.
- **DB-free unit tests** — concurrency-safe in-memory fakes; `-race` clean.
- **Swagger UI** and a ready-to-import Postman collection.

## Tech stack

| Concern        | Choice                                   |
| -------------- | ---------------------------------------- |
| Language       | Go 1.23+                                 |
| HTTP           | Gin                                      |
| Persistence    | MySQL 8 via GORM                         |
| Migrations     | golang-migrate (versioned SQL)           |
| Cache / tokens | Redis 7 (`go-redis/v9`)                  |
| Auth           | JWT (HS256), bcrypt password hashing     |
| Logging        | `log/slog` (JSON)                        |
| API docs       | swaggo / Swagger UI                      |
| Testing        | stdlib `testing` + testify, `-race`      |

## Contents

- [Architecture](#architecture)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [API reference](#api-reference)
- [Idempotency](#idempotency)
- [Transactional assignment](#transactional-assignment)
- [Response envelope](#response-envelope)
- [Logging](#logging)
- [Testing](#testing)
- [Postman](#postman)
- [Make targets](#make-targets)

## Architecture

Dependency direction is strictly one-way. A module's `usecase` depends only on
its own `domain`/`ports` (plus the `ports`/`domain` of the modules it
collaborates with); handlers depend on usecases; repositories implement ports.
Nothing imports a handler. Cross-cutting contracts live in `shared/kernel`
(imported by every layer) and infrastructure lives under `platform`.

```
handler -> usecase -> domain (this module)
               |  \-> other modules' ports
               \---> shared/kernel     # errors, tx, auth, pagination, notification
repository ---------> platform         # db, redis, jwt, logger, notify
```

```
cmd/api                  entrypoint: config, wiring, graceful shutdown
internal/
  http                   composition router + middleware chain
  package/               one package tree per feature module
    auth|user|team|task|idempotency|health
      domain             entities (no frameworks)
      ports              repository/usecase interfaces + input DTOs
      usecase            business rules
      repository         GORM models, mappings, MySQL repositories
      handler            HTTP handlers + dto/{request,response}
  shared/
    kernel               cross-cutting contracts (TxManager, Token*, errors)
    apperr               typed application errors (status + code)
    response             uniform success/error envelope
    reqctx, httpx        request context + bind/context helpers
  platform/
    config               env -> typed configuration
    database             GORM connection, TxManager, error translation
    redis, cache         Redis client + token store
    security             JWT + bcrypt
    logger, notify       slog logger + logging notifier
    middleware           request id, recovery, logging, CORS, auth
migrations               numbered up/down SQL
test/mocks               concurrency-safe in-memory fakes for unit tests
test/domain              test-only aggregation of the module domain types
test/integration         build-tagged tests against real MySQL
docs                     generated Swagger + Postman collection
```

Design patterns: Repository, Unit of Work (`TxManager`), constructor dependency
injection, Factory (router), Chain of Responsibility (middleware), Strategy
(notifier, token store) and explicit DTO mapping.

## Getting started

### Prerequisites

- Go 1.23+
- Docker (for the Docker workflow), or a local MySQL 8 + Redis 7 and the
  `migrate` CLI (`brew install golang-migrate`)

### Docker (everything in one command)

Starts MySQL, Redis, applies migrations and runs the API.

```bash
cp .env.example .env
make docker-up          # docker compose up -d --build
# API:      http://localhost:8080
# Swagger:  http://localhost:8080/swagger/index.html
make logs
make docker-down
```

### Local

Requires a running MySQL and Redis and the `migrate` CLI.

```bash
cp .env.example .env
make migrate-up
make run
```


Aplikasi akan melakukan **auto-seed** saat pertama kali dijalankan jika belum ada user di database.

**Dummy users (tersedia setelah seed):**
- `demo@example.com` / `password123` (Owner - Demo Team)
- `alice@example.com` / `password123` (Member - Demo Team)

Bisa juga dijalankan manual: `make seed` (membutuhkan DB berjalan).

## Configuration

Configuration is loaded from `.env` (via Viper) and the process environment.
See `.env.example`; the notable variables are:

| Variable          | Default         | Purpose                           |
| ----------------- | --------------- | --------------------------------- |
| `APP_ENV`         | `development`   | `production` disables Swagger     |
| `APP_NAME`        | `gdc-task-api`  | Application name                  |
| `APP_PORT`        | `8080`          | HTTP port                         |
| `DB_*`            | localhost / gdc | MySQL connection                  |
| `REDIS_*`         | localhost:6379  | Redis connection                  |
| `JWT_SECRET`      | – (required)    | HMAC signing key                  |
| `JWT_ISSUER`      | `gdc-task-api`  | JWT issuer claim                  |
| `JWT_ACCESS_TTL`  | `15m`           | Access token lifetime             |
| `JWT_REFRESH_TTL` | `168h`          | Refresh token lifetime            |
| `IDEMPOTENCY_TTL` | `24h`           | Idempotency replay window         |

## API reference

Base URL: `http://localhost:8080`. All task and `/users/me` routes require
`Authorization: Bearer <access_token>`.

| Method | Path                 | Auth | Description                                    |
| ------ | -------------------- | :--: | ---------------------------------------------- |
| GET    | `/health`            |  –   | Liveness probe                                 |
| GET    | `/ready`             |  –   | Readiness probe (MySQL + Redis)                |
| POST   | `/auth/register`     |  –   | Register (creates a personal team)             |
| POST   | `/auth/login`        |  –   | Log in, returns access + refresh tokens        |
| POST   | `/auth/refresh`      |  –   | Rotate a refresh token                         |
| POST   | `/auth/logout`       |  ✓   | Revoke access + refresh tokens                 |
| GET    | `/users/me`          |  ✓   | Current user profile                           |
| PUT    | `/users/me`          |  ✓   | Update current user profile                    |
| POST   | `/tasks`             |  ✓   | Create a task (requires `Idempotency-Key`)     |
| GET    | `/tasks`             |  ✓   | List tasks (`status`, `search`, `page`, `limit`) |
| GET    | `/tasks/{id}`        |  ✓   | Task detail                                    |
| PUT    | `/tasks/{id}`        |  ✓   | Update task                                    |
| DELETE | `/tasks/{id}`        |  ✓   | Delete task                                    |
| POST   | `/tasks/{id}/assign` |  ✓   | Assign to a teammate (transactional)           |

Interactive docs: **`http://localhost:8080/swagger/index.html`** (disabled when
`APP_ENV=production`).

### Example flow

```bash
BASE=http://localhost:8080

# 1. Register (returns a token pair)
curl -s $BASE/auth/register -H 'Content-Type: application/json' \
  -d '{"name":"Jane","email":"jane@example.com","password":"password123"}'

# 2. Log in and capture the access token
TOKEN=$(curl -s $BASE/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"jane@example.com","password":"password123"}' | jq -r .data.access_token)

# 3. Create a task (Idempotency-Key must be a UUID)
curl -s $BASE/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $(uuidgen | tr '[:upper:]' '[:lower:]')" \
  -d '{"title":"Write the report","status":"todo"}'

# 4. List tasks with filters and pagination
curl -s "$BASE/tasks?status=todo&search=report&page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

## Idempotency

`POST /tasks` requires an `Idempotency-Key` header containing a UUID.

- The key's uniqueness constraint, the task insert and the stored response are
  written in **one transaction**. The database constraint is the source of
  truth, so concurrent retries create exactly one task.
- A replay within the window returns the **identical stored response** without
  side effects.
- The exact response bytes are captured by a render callback on first
  execution, so replays are never re-serialized.

| Condition                            | Response                                    |
| ------------------------------------ | ------------------------------------------- |
| Missing key                          | `400 IDEMPOTENCY_KEY_REQUIRED`              |
| Non-UUID key                         | `400 IDEMPOTENCY_KEY_INVALID`               |
| Same key, different payload          | `409 IDEMPOTENCY_KEY_MISMATCH`              |
| Same key, still processing           | `409 IDEMPOTENCY_KEY_IN_PROGRESS`           |

## Transactional assignment

`POST /tasks/{id}/assign` runs *update assignee + write `task_logs` + dispatch
notification* inside a single `TxManager` transaction. If any step fails (for
example the notifier), the whole operation is rolled back. Repositories join the
ambient transaction automatically by reading it from the context.

## Response envelope

Every response — success or error — uses one envelope. Successful responses
carry `success: true` with `message` and `data`; errors carry `success: false`
with `message` and, for validation failures, an `errors` map of field →
message. List endpoints return `data` plus a `metadata` block with pagination.

Success:

```json
{
  "success": true,
  "message": "OK",
  "data": { "id": 1, "title": "Write the report", "status": "todo" }
}
```

List (pagination lives in `metadata`):

```json
{
  "success": true,
  "message": "OK",
  "data": [ { "id": 1, "title": "Write the report", "status": "todo" } ],
  "metadata": { "current_page": 1, "per_page": 10, "total": 42, "total_pages": 5 }
}
```

No payload (delete / logout):

```json
{
  "success": true,
  "message": "Task deleted",
  "data": null
}
```

Error:

```json
{
  "status": 400,
  "code": "VALIDATION_ERROR",
  "message": "title is required",
  "timestamp": "2026-01-02T15:04:05Z",
  "errors": {
    "title": "title is required"
  }
}
```

Per-context shape:

| Context         | success | message                       | data   | extra                       |
| --------------- | ------- | ----------------------------- | ------ | --------------------------- |
| Single item     | `true`  | `OK`                          | object | —                           |
| Create          | `true`  | `Created`                     | object | —                           |
| List            | `true`  | `OK`                          | array  | `metadata`                  |
| Delete / logout | `true`  | `Task deleted` / `Logged out` | `null` | —                           |
| Validation      | `false` | joined messages               | —      | `errors` (field → message)  |
| Any error       | —       | error message                 | —      | `{status, code, timestamp, errors?}` |

HTTP status codes distinguish the failure kind (400/401/403/404/409/422/500);
internal details and stack traces are never exposed. A recovery middleware
catches panics and returns a generic `500`.

## Logging

Each request emits one JSON line via `slog` with `request_id`, `method`, `path`,
`status`, `latency_ms`, `client_ip` and (when authenticated) `user_id`. Level is
`INFO` for 2xx/3xx, `WARN` for 4xx and `ERROR` for 5xx.

## Testing

Unit tests need **no database or Redis** — they use concurrency-safe in-memory
fakes (`test/mocks`).

```bash
make test           # go test ./...
make test-race      # go test -race -count=1 ./...
make seed           # seed dummy users and tasks (requires DB)
make cover          # coverage summary
```

The core test is `internal/package/task/usecase/task_usecase_race_test.go`: 64
goroutines submit the same `Idempotency-Key` and exactly one task must be
created, with identical responses. Table-driven tests use maps keyed by case
name.

Integration tests hit real MySQL and are build-tagged:

```bash
make docker-up && make migrate-up
make test-integration   # go test -tags integration ./test/integration/...
```

## Postman

Import `docs/postman/gdc.postman_collection.json` and
`docs/postman/gdc.local.postman_environment.json`, then select the environment.
The Login/Register requests save the access token automatically, and the
`Idempotency-Key` header is generated per request.

## Make targets

```bash
make help            # list all targets with descriptions
```

`run build tidy fmt vet test test-race test-integration cover lint swag
migrate-up migrate-down migrate-create docker-up docker-down logs clean`
