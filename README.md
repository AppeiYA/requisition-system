# Requisition System

Backend API for managing requisitions: staff invitations, departments and department
memberships, with JWT authentication and Redis-backed refresh-token sessions.

Built with **Go + Gin + PostgreSQL + Redis**, structured as a modular monolith using
hexagonal architecture (ports & adapters) per feature module.

## Features

- **Authentication** – login, refresh, logout, `me`; access tokens (JWT) plus rotating
  refresh sessions stored in Redis; role-based access (`admin`, member).
- **Invitations** – admins invite staff by email; the invitee accepts via a single-use,
  24h link delivered through [Resend](https://resend.com).
- **Departments** – create and list departments with a unique 4-letter code, and manage
  department memberships (job titles).
- **API documentation** – OpenAPI 3 spec served by Scalar at `/api/v1/docs`.

## Tech stack

| Concern      | Choice |
| ------------ | ------ |
| Language     | Go 1.26 |
| HTTP         | [Gin](https://github.com/gin-gonic/gin) |
| Database     | PostgreSQL (lib/pq + sqlx) |
| Sessions     | Redis |
| Auth         | JWT (access) + Redis refresh sessions |
| Email        | [Resend](https://resend.com) Go SDK |
| Migrations   | golang-migrate (plain SQL) + custom seed runner |
| Docs         | swaggo → OpenAPI 3 → Scalar |
| Telemetry    | OpenTelemetry |

## Architecture

```
internal/
├── identity/       # auth, users, roles
├── department/     # departments & memberships
├── invitation/     # invitations lifecycle
├── admin/          # admin-only actions (invite members)
├── shared/         # config, db, redis, email, jwt, response, telemetry
├── app.go          # composition root: wires modules together
└── router.go       # route table + docs endpoints
```

Each module is split into `ports/in` (use-case interfaces + read models),
`ports/out` (repository interfaces), `usecase`, and `adapters/{in,out}` — so HTTP,
Postgres and email concerns stay outside the domain. `internal/app.go` is the only
place that knows the concrete implementations.

## Prerequisites

- Go 1.26+
- PostgreSQL 14+
- Redis 6+
- A [Resend](https://resend.com) API key (for invitation emails)

## Quick start

```bash
# 1. Configure your environment
cp .env.example .env
#    then edit .env – see "Configuration" below

# 2. Create the database
createdb requisition_system

# 3. Apply migrations
go run ./migrations -up

# 4. Seed the default departments (Engineering, Design, Administrator)
go run ./migrations -seed

# 5. Run the API (serves on :8080 by default)
go run ./cmd/api

# 6. In a second terminal, run the worker (drains background tasks,
#    e.g. sending invitation emails)
go run ./cmd/worker
```

The seed step also creates a default admin:

| Field    | Value |
| -------- | ----- |
| Email    | `admin@example.com` |
| Password | `Admin123!` |

Change or delete this account before deploying anywhere real.

## Migrations & seeds

The CLI in `migrations/` wraps golang-migrate plus a small seed runner:

```bash
go run ./migrations -up                       # apply all migrations
go run ./migrations -down                    # roll back one step
go run ./migrations -reset                   # reset migration state (force to -1)
go run ./migrations -force 3                 # force a specific version (clears dirty state)
go run ./migrations -seed                    # run every file in ./seeds (sorted)
go run ./migrations -seed-file seeds/001_admin.sql   # run a single seed file
```

Migrations run in `migrations/*.up.sql` / `*.down.sql` pairs; seed files run inside a
transaction and are idempotent (they use `ON CONFLICT ... DO UPDATE`).

## API

Base URL: `http://localhost:8080/api/v1` — interactive docs at
[`/api/v1/docs`](http://localhost:8080/api/v1/docs), raw spec at `/api/v1/openapi.yaml`.

| Method | Path                        | Auth         | Description |
| ------ | --------------------------- | ------------ | ----------- |
| GET    | `/api/healthz`              | –            | Liveness check |
| GET    | `/api/v1/docs`              | –            | Scalar API reference |
| GET    | `/api/v1/openapi.yaml`      | –            | OpenAPI 3 spec |
| POST   | `/api/v1/auth/login`        | –            | Login, returns access token + refresh token |
| POST   | `/api/v1/auth/refresh`      | –            | Rotate the refresh session |
| POST   | `/api/v1/auth/logout`       | –            | Revoke the refresh session |
| GET    | `/api/v1/auth/me`           | access token | Current user |
| GET    | `/api/v1/invitations/:token/validate` | – | Check an invitation link |
| POST   | `/api/v1/invitations/:token/accept`   | – | Accept an invitation and register |
| POST   | `/api/v1/admin/invitations` | admin        | Invite a staff member by email |
| POST   | `/api/v1/departments`       | access token | Create a department |
| GET    | `/api/v1/departments`       | access token | List departments (ordered by name) |
| GET    | `/api/v1/departments/:id`   | access token | Get a department |

Authenticated requests need `Authorization: Bearer <access token>`.

## Configuration

All configuration comes from environment variables (a local `.env` is loaded
automatically). Required variables have no default and stop the app at boot.

| Variable | Required | Default | Description |
| -------- | -------- | ------- | ----------- |
| `APP_ENV` | – | `development` | `development`, `staging` or `production` |
| `API_BASE_URL` | – | `/api/v1` | Base path the API is served under |
| `SERVER_PORT` | – | `:8080` | Listen address |
| `SERVER_READ_TIMEOUT` | – | `5s` | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | – | `10s` | HTTP write timeout |
| `SERVER_IDLE_TIMEOUT` | – | `120s` | Keep-alive idle timeout |
| `DB_HOST` | ✅ | – | PostgreSQL host |
| `DB_PORT` | – | `5432` | PostgreSQL port |
| `DB_USER` | ✅ | – | PostgreSQL user |
| `DB_PASSWORD` | ✅ | – | PostgreSQL password |
| `DB_NAME` | ✅ | – | Database name |
| `DB_SSL_MODE` | – | `disable` | `disable`, `require`, … |
| `DB_MAX_OPEN_CONNECTIONS` | – | `25` | Pool: max open connections |
| `DB_MAX_IDLE_CONNECTIONS` | – | `10` | Pool: max idle connections |
| `DB_MAX_LIFETIME_MINUTES` | – | `30` | Pool: connection max lifetime |
| `REDIS_HOST` | ✅ | – | Redis host |
| `REDIS_PORT` | ✅ | – | Redis port |
| `REDIS_USERNAME` | – | – | Redis username (ACL) |
| `REDIS_PASSWORD` | – | – | Redis password |
| `REDIS_DB` | – | `0` | Redis database index |
| `JWT_SECRET` | ✅ | – | Signing secret, **min 32 characters** (`openssl rand -hex 32`) |
| `JWT_ISSUER` | – | `requisition-system` | Token issuer |
| `JWT_AUDIENCE` | – | `requisition-system` | Token audience |
| `JWT_ACCESS_TOKEN_TTL` | – | `15m` | Access-token lifetime |
| `JWT_REFRESH_TOKEN_TTL` | – | `168h` | Refresh-token lifetime; must exceed the access TTL |
| `JWT_LEEWAY` | – | `30s` | Clock-skew leeway for token validation |
| `RESEND_API_KEY` | ✅ | – | Resend API key |
| `EMAIL_FROM` | – | `onboarding@resend.dev` | Sender address (must be verified in Resend) |
| `EMAIL_FROM_NAME` | – | `Requisition System` | Sender display name |
| `FRONTEND_BASE_URL` | – | `http://localhost:3000` | Base URL for invitation links in emails |
| `WORKER_CONCURRENCY` | – | `10` | Parallel task handlers |
| `WORKER_QUEUE` | – | `default` | Queue the worker drains and tasks are enqueued to |
| `WORKER_TASK_MAX_RETRY` | – | `5` | Delivery attempts per task before it is given up on |
| `WORKER_TASK_TIMEOUT` | – | `30s` | Per-attempt processing timeout |
| `WORKER_SHUTDOWN_TIMEOUT` | – | `8s` | Grace period for in-flight tasks on shutdown |

## Worker (background tasks)

Work that does not need to run inside the request cycle runs in a separate
binary, `cmd/worker`, so slow integrations never add latency to a response.
Invitation emails work this way: the API persists the invitation, enqueues an
`admin:send-invitation-email` task on Redis (via [asynq](https://github.com/hibiken/asynq))
and returns immediately; the worker delivers the mail, retrying with
exponential backoff up to `WORKER_TASK_MAX_RETRY` times.

```bash
go run ./cmd/worker    # run alongside ./cmd/api
```

Both binaries share the same `.env`. Each domain registers its own tasks
through a `build<Domain>Worker` function in `internal/` — see
`internal/admin_worker.go` for the pattern.

## Email delivery (Resend)

Invitation emails are sent through the [Resend](https://resend.com) API:

1. Create an API key at <https://resend.com/api-keys> and set it as `RESEND_API_KEY`.
2. For local development you can keep the sandbox sender
   `EMAIL_FROM=onboarding@resend.dev` — it works without domain verification but only
   delivers to your own account email.
3. For production, verify your domain in Resend (add the DNS records it gives you) and
   point `EMAIL_FROM` / `EMAIL_FROM_NAME` at it, e.g. `EMAIL_FROM=noreply@yourdomain.com`.

The sender lives in `internal/shared/infrastructure/email/sender.go`; invitation
content is built in `internal/admin/adapters/out/email/invitation_sender.go`.

## Telemetry

The API process exports OpenTelemetry traces and metrics for the two
request-driven business operations. The use cases ask for an operation to be
measured through an outbound port (`ports/out/telemetry.go`), so the domain
never depends on OpenTelemetry; the adapter lives in
`internal/shared/infrastructure/telemetry` and is wired in `internal/app.go`.

| Operation | Span | Metrics |
| --------- | ---- | ------- |
| Admin creates an invitation | `admin.invitation.create` | `requisition.admin.invitations.{attempted,succeeded,failed,duration}` |
| Invitee accepts an invitation | `invitation.accept` | `requisition.invitee.acceptances.{attempted,succeeded,failed,duration}` |

The span covers the whole operation: the admin span covers the durable
invitation creation and the enqueueing of the delivery task, the acceptance
span covers the transaction that registers the user, creates the membership and
accepts the invitation. Durations are reported in seconds.

The names are stable and low cardinality on purpose: no email, token, user or
invitation identifier is ever reported as a metric attribute. The measurement
never changes the outcome of the operation — the business error is always the
one returned to the caller.

## Docker

```bash
# API image (default target)
docker build -f dockerfile -t requisition-system .

# Worker image
docker build -f dockerfile --target worker -t requisition-system-worker .

docker run --rm -p 8080:8080 --env-file .env requisition-system
docker run --rm --env-file .env requisition-system-worker
```

The dockerfile builds both binaries in one stage and ships them as two image
targets: the API image runs `./app`, the worker image runs `./worker`. Both
expect the PostgreSQL and Redis addresses in the environment to be reachable
from the container. On push to `master`, CI publishes both images to GHCR
(`…/requisition-system` and `…/requisition-system-worker`).

## Testing

```bash
go test ./...        # everything
go test ./internal/department/...   # a single module
```

CI (`.github/workflows/ci-cd.yml`) runs the full suite on every pull request and
pushes the image to GHCR on `master`.

## Regenerating API docs

Docs are generated from swag annotations in each `*_docs.go` file:

```bash
go run github.com/swaggo/swag/cmd/swag init -g main.go -d ./cmd/api,./internal -o docs
go run ./cmd/specgen      # converts docs/swagger.json -> docs/openapi.yaml
```

Only `docs/openapi.yaml` and `docs/embed.go` are tracked; the rest are build artifacts.
