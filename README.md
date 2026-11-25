# Requisition System

Backend API for managing requisitions: staff invitations, departments and department
memberships, and item categories, with JWT authentication and Redis-backed
refresh-token sessions.

Built with **Go + Gin + PostgreSQL + Redis**, structured as a modular monolith using
hexagonal architecture (ports & adapters) per feature module.

## Features

- **Authentication** – login, refresh, logout, `me`; access tokens (JWT) plus rotating
  refresh sessions stored in Redis; role-based access (`admin`, member).
- **Password reset** – self-service reset links emailed through
  [Resend](https://resend.com); single-use tokens with a 30-minute lifetime,
  generic responses that do not reveal whether an account exists, and automatic
  revocation of all refresh sessions on success.
- **Invitations** – admins invite staff by email; the invitee accepts via a single-use,
  24h link delivered through [Resend](https://resend.com).
- **Departments** – create and list departments with a unique 4-letter code, and manage
  department memberships (job titles).
- **Item categories** – admins create and update categories; authenticated users list
  and read them. Default categories are provided as a seed.
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
| Observability | OpenTelemetry → Tempo (traces) + Prometheus (metrics) + Grafana |

## Architecture

```
internal/
├── identity/       # auth, users, roles
├── department/     # departments & memberships
├── invitation/     # invitations lifecycle
├── admin/          # admin-only actions (invite members, manage users)
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
- A [Resend](https://resend.com) API key (for invitation and password-reset emails)

## Quick start

```bash
# 1. Configure your environment
cp .env.example .env
#    then edit .env – see "Configuration" below

# 2. Create the database
createdb requisition_system

# 3. Apply migrations
go run ./migrations -up

# 4. Seed the default departments and item categories
go run ./migrations -seed

# 5. Run the API (serves on :8080 by default)
go run ./cmd/api

# 6. In a second terminal, run the worker (drains background tasks,
#    e.g. sending invitation/password-reset emails and cleaning up spent tokens)
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

Migration `00014` cannot be safely rolled back in general: the previous schema has
only one requisition line and no status-history table. The migration CLI refuses this
rollback unless a verified backup has been taken and
`-allow-destructive-down` is supplied. A data-preservation procedure is still
required; the flag is an explicit operator acknowledgement, not a backup.

Item names are globally unique regardless of casing. Categories classify items but do
not define item identity. Before applying migration 00014 to an existing deployment,
audit and resolve any case-insensitive duplicate names first.

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
| POST   | `/api/v1/auth/forgot-password` | –            | Send a password-reset link (always generic, gated by `PASSWORD_RESET_ENABLED`) |
| POST   | `/api/v1/auth/reset-password`  | –            | Set a new password with a single-use token; revokes all refresh sessions |
| GET    | `/api/v1/auth/me`           | access token | Current user |
| GET    | `/api/v1/invitations/:token/validate` | – | Check an invitation link |
| POST   | `/api/v1/invitations/:token/accept`   | – | Accept an invitation and register |
| POST   | `/api/v1/admin/invitations` | admin        | Invite a staff member by email |
| GET    | `/api/v1/admin/invitations` | admin        | List/search/filter invitations (paginated, with total) |
| GET    | `/api/v1/admin/invitations/:id` | admin  | Get invitation detail |
| POST   | `/api/v1/admin/invitations/:id/resend` | admin | Rotate the token and queue a replacement email |
| POST   | `/api/v1/admin/invitations/:id/revoke` | admin | Invalidate the invitation link immediately |
| GET    | `/api/v1/admin/users`       | admin        | List/search/filter users (paginated, with total) |
| GET    | `/api/v1/admin/users/:id`   | admin        | Get a user with department and requisition activity |
| PATCH  | `/api/v1/admin/users/:id/status` | admin  | Activate, suspend or deactivate a user |
| PATCH  | `/api/v1/admin/users/:id/role`    | admin  | Promote or demote a user |
| POST   | `/api/v1/admin/users/:id/sessions/revoke` | admin | Force logout (revoke refresh sessions) |
| POST   | `/api/v1/departments`       | access token | Create a department |
| GET    | `/api/v1/departments`       | access token | List departments (ordered by name) |
| GET    | `/api/v1/departments/:id`   | access token | Get a department |
| POST   | `/api/v1/categories`        | admin        | Create an item category |
| GET    | `/api/v1/categories`        | access token | List item categories (ordered by name) |
| GET    | `/api/v1/categories/:id`    | access token | Get an item category |
| PATCH  | `/api/v1/categories/:id`    | admin        | Partially update an item category's name/description |
| PATCH  | `/api/v1/categories/:id/activate`   | admin | Activate an item category |
| PATCH  | `/api/v1/categories/:id/deactivate` | admin | Deactivate an item category |
| POST   | `/api/v1/items`             | admin        | Create an item |
| GET    | `/api/v1/items`             | access token | List items (ordered by name) |
| GET    | `/api/v1/items/:id`         | access token | Get an item |
| PATCH  | `/api/v1/items/:id`         | admin        | Partially update an item's name/description/category/unit cost |
| PATCH  | `/api/v1/items/:id/activate`   | admin | Activate an item |
| PATCH  | `/api/v1/items/:id/deactivate` | admin | Deactivate an item |
| PATCH  | `/api/v1/requisitions/:id` | access token | Update an owned pending requisition |

Authenticated requests need `Authorization: Bearer <access token>`.

### Admin operations

User and invitation management are admin-only. The list and detail endpoints
never return password hashes, refresh tokens, raw invitation tokens or token
hashes.

- Status transitions: `active → suspended|inactive` and
  `suspended|inactive → active`. A reason is required when suspending or
  deactivating a user.
- Role changes accept only `admin` or `user`; promotion requires an active
  account.
- Self-protection: an admin cannot change their own status or role (`403`).
- Last active admin: suspending, deactivating or demoting the last active admin
  is rejected (`409`). The other active-admin rows are locked with
  `FOR UPDATE`, so two concurrent changes cannot both remove the last admin.
- Every mutation writes an append-only row to `admin_audit_log` with the actor,
  before/after values, reason and request metadata, in the same transaction as
  the change.
- A suspension and any role change revoke all refresh sessions through a
  per-user epoch in Redis. Force logout (`POST .../sessions/revoke`) does the
  same on its own; an optional `?reason=` is recorded in the audit log.

Invitation management follows the same rules:

- The list accepts `q`, `status`, `department_id`, `invited_by`, `from`, `to`,
  `limit` and `offset`. Dates are RFC3339 timestamps or `YYYY-MM-DD` dates in
  UTC; `from` is inclusive and `to` is exclusive.
- Resend rotates the raw token and replaces the email in the queue. Accepted
  and revoked invitations cannot be resent (`409`), and resending fails if
  another active invitation already exists for the address. The response
  carries only the invitation ID and new expiry; the token is delivered by
  email.
- Revoking is immediate because the accept flow checks the current invitation
  status. Revoking an already revoked invitation is a successful no-op; an
  accepted invitation cannot be revoked (`409`). An optional `?reason=` is
  recorded in the audit log.
- Resend and revoke write `invitation.resend` and `invitation.revoke` audit
  entries in the same transaction as the mutation.

### Security

Authorization reads the user's current role and status from PostgreSQL on
every request, not the role claim in the access token. A stale token therefore
cannot keep admin access after demotion or suspension: suspended or inactive
users are refused with `403` even while their access token is still valid, and
a demoted admin loses admin routes immediately. Revoking sessions invalidates
refresh tokens; existing access tokens expire within their 15-minute TTL.

### Password reset

The two self-service endpoints work as follows:

1. `POST /auth/forgot-password` takes an email address. Unknown, inactive and
   throttled addresses all receive the same `200` response and the same message;
   no account information, token or email address is returned. A valid active
   account gets a hashed token row in `password_reset_tokens` and an
   `identity:send-password-reset-email` task on the queue. The raw token is
   never stored — only its SHA-256 hash is.
2. The worker sends a link built from `FRONTEND_BASE_URL` and
   `FRONTEND_PASSWORD_RESET_PATH`, with the raw token in the `?token=` query
   parameter. Each request invalidates the previous active token, so the newest
   email is always the one that works.
3. `POST /auth/reset-password` takes `{ "token": "…", "password": "…" }` in the
   JSON body. The token is consumed atomically, so it can only be used once;
   unknown, expired, spent, superseded or wrong-user tokens all return the same
   generic `400` ("invalid or expired password reset token"). A weak password
   is rejected by the existing password policy *before* the token is consumed,
   leaving it usable.
4. A successful reset writes the new password hash, invalidates any other
   active tokens and revokes all refresh sessions in the same transaction. The
   cached user row is invalidated too, and no access or refresh token is
   issued — the user signs in normally with the new password. If session
   revocation fails, the transaction rolls back and the token stays usable.

Reset requests are limited by client IP (`RATE_LIMIT_FORGOT_PASSWORD_*` and
`RATE_LIMIT_RESET_PASSWORD_*`) and per account (`PASSWORD_RESET_REQUEST_COOLDOWN`
and `PASSWORD_RESET_MAX_PER_HOUR`). Throttled forgot-password requests still
return the generic success response. Set `PASSWORD_RESET_ENABLED=false` to hide
both routes until the frontend page is ready.

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
| `DB_SSL_MODE` | – | `disable` | PostgreSQL TLS mode; use `verify-full` in production |
| `DB_CA_CERT` | – | – | Absolute path to the PostgreSQL CA certificate/bundle |
| `DB_MAX_OPEN_CONNECTIONS` | – | `25` | Pool: max open connections |
| `DB_MAX_IDLE_CONNECTIONS` | – | `10` | Pool: max idle connections |
| `DB_MAX_LIFETIME_MINUTES` | – | `30` | Pool: connection max lifetime |
| `REDIS_HOST` | ✅ | – | Redis host |
| `REDIS_PORT` | ✅ | – | Redis port |
| `REDIS_USERNAME` | – | – | Redis username (ACL) |
| `REDIS_PASSWORD` | – | – | Redis password |
| `REDIS_DB` | – | `0` | Redis database index |
| `REDIS_TLS` | – | `false` | Enable Redis TLS |
| `REDIS_CA_CERT` | – | – | Absolute path to the Redis CA certificate/bundle |
| `REDIS_TLS_SERVER_NAME` | – | `REDIS_HOST` | TLS server name used for certificate verification |
| `JWT_SECRET` | ✅ | – | Signing secret, **min 32 characters** (`openssl rand -hex 32`) |
| `JWT_ISSUER` | – | `requisition-system` | Token issuer |
| `JWT_AUDIENCE` | – | `requisition-system` | Token audience |
| `JWT_ACCESS_TOKEN_TTL` | – | `15m` | Access-token lifetime |
| `JWT_REFRESH_TOKEN_TTL` | – | `168h` | Refresh-token lifetime; must exceed the access TTL |
| `JWT_LEEWAY` | – | `30s` | Clock-skew leeway for token validation |
| `PASSWORD_RESET_ENABLED` | – | `true` | Register the forgot/reset routes |
| `PASSWORD_RESET_TOKEN_TTL` | – | `30m` | Reset-link lifetime; must be positive |
| `PASSWORD_RESET_REQUEST_COOLDOWN` | – | `1m` | Minimum gap between reset emails for one account |
| `PASSWORD_RESET_MAX_PER_HOUR` | – | `5` | Maximum reset requests per account per rolling hour |
| `PASSWORD_RESET_CLEANUP_RETENTION` | – | `168h` | How long spent token rows are kept; at least the token TTL |
| `RATE_LIMIT_LOGIN_MAX` | – | `5` | Login attempts per client IP per window |
| `RATE_LIMIT_LOGIN_WINDOW` | – | `15m` | Login rate-limit window |
| `RATE_LIMIT_REFRESH_MAX` | – | `10` | Refresh attempts per client IP per window |
| `RATE_LIMIT_REFRESH_WINDOW` | – | `1m` | Refresh rate-limit window |
| `RATE_LIMIT_LOGOUT_MAX` | – | `20` | Logout attempts per client IP per window |
| `RATE_LIMIT_LOGOUT_WINDOW` | – | `1m` | Logout rate-limit window |
| `RATE_LIMIT_FORGOT_PASSWORD_MAX` | – | `5` | Forgot-password requests per client IP per window |
| `RATE_LIMIT_FORGOT_PASSWORD_WINDOW` | – | `15m` | Forgot-password rate-limit window |
| `RATE_LIMIT_RESET_PASSWORD_MAX` | – | `10` | Reset-password attempts per client IP per window |
| `RATE_LIMIT_RESET_PASSWORD_WINDOW` | – | `15m` | Reset-password rate-limit window |
| `RATE_LIMIT_TRUSTED_PROXIES` | – | – | Comma-separated IPs/CIDRs allowed to set `X-Forwarded-For` |
| `RESEND_API_KEY` | ✅ | – | Resend API key |
| `EMAIL_FROM` | – | `onboarding@resend.dev` | Sender address (must be verified in Resend) |
| `EMAIL_FROM_NAME` | – | `Requisition System` | Sender display name |
| `FRONTEND_BASE_URL` | – | `http://localhost:3000` | Base URL for invitation and reset links in emails |
| `FRONTEND_PASSWORD_RESET_PATH` | – | `/reset-password` | Reset page path; must start with `/` |
| `WORKER_CONCURRENCY` | – | `10` | Parallel task handlers |
| `WORKER_QUEUE` | – | `default` | Queue the worker drains and tasks are enqueued to |
| `WORKER_TASK_MAX_RETRY` | – | `5` | Delivery attempts per task before it is given up on |
| `WORKER_TASK_TIMEOUT` | – | `30s` | Per-attempt processing timeout |
| `WORKER_SHUTDOWN_TIMEOUT` | – | `8s` | Grace period for in-flight tasks on shutdown |

## Worker (background tasks)

Work that does not need to run inside the request cycle runs in a separate
binary, `cmd/worker`, so slow integrations never add latency to a response.
Invitation and password-reset emails work this way: the API persists the
record, enqueues a task on Redis (via [asynq](https://github.com/hibiken/asynq))
and returns immediately; the worker delivers the mail, retrying with
exponential backoff up to `WORKER_TASK_MAX_RETRY` times.

| Task | Trigger | Purpose |
| ---- | ------- | ------- |
| `admin:send-invitation-email` | API | Deliver an invitation link |
| `identity:send-password-reset-email` | API | Deliver a reset link (deduplicated by reset ID) |
| `identity:cleanup-password-reset-tokens` | Every hour | Delete token rows older than `PASSWORD_RESET_CLEANUP_RETENTION` |
| `invitation:expire` | Every 15 minutes | Expire stale invitations |

```bash
go run ./cmd/worker    # run alongside ./cmd/api
```

Both binaries share the same `.env`. Each domain registers its own tasks
through a `build<Domain>Worker` function in `internal/` — see
`internal/admin_worker.go` and `internal/identity_worker.go` for the pattern.

## Email delivery (Resend)

Invitation and password-reset emails are sent through the
[Resend](https://resend.com) API:

1. Create an API key at <https://resend.com/api-keys> and set it as `RESEND_API_KEY`.
2. For local development you can keep the sandbox sender
   `EMAIL_FROM=onboarding@resend.dev` — it works without domain verification but only
   delivers to your own account email.
3. For production, verify your domain in Resend (add the DNS records it gives you) and
   point `EMAIL_FROM` / `EMAIL_FROM_NAME` at it, e.g. `EMAIL_FROM=noreply@yourdomain.com`.

The sender lives in `internal/shared/infrastructure/email/sender.go`; message
content is built per domain — invitation content in
`internal/admin/adapters/out/email/invitation_sender.go`, reset links in
`internal/identity/adapters/out/email/password_reset_sender.go`.

## Observability

The API and worker processes export OpenTelemetry traces and metrics for their
business operations. The use cases ask for an operation to be
measured through an outbound port (`ports/out/telemetry.go`), so the domain
never depends on OpenTelemetry; the adapter lives in
`internal/shared/infrastructure/telemetry` and is wired in `internal/app.go`
and `internal/worker.go`.

The exporter defaults to `stdout`, so nothing changes until you ask for it to
leave the process.

### The bundled backend

`deploy/observability/` is a self-contained stack the telemetry can be exported
to — the collector, Tempo for traces, Prometheus for metrics and Grafana for
the dashboard. It runs beside the application and never instead of it: with the
directory absent, the application starts and logs to stdout exactly as it did
before.

```bash
docker compose -f deploy/observability/docker-compose.observability.yml up -d
```

The standalone binary takes the same arguments, for a machine that has no
Compose plugin: `docker-compose -f deploy/observability/docker-compose.observability.yml up -d`.

Then export to it:

```bash
TELEMETRY_EXPORTER=otlp
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

| | |
| --- | ----------------------------------------------- |
| http://localhost:3000 | the dashboard (`Requisition System`), login `admin` / `admin` |
| http://localhost:9090 | Prometheus |
| http://localhost:3200 | Tempo |
| http://localhost:13133 | the collector's health endpoint |

The dashboard is provisioned, not imported by hand, so it is there the moment
the stack starts. Two variables, `service` and `env`, filter every panel;
`service` separates the API from the worker. It covers request rate, latency
and the 5xx ratio; the four business operations by attempted, succeeded and
failed; the worker's email delivery and expirations; and the Postgres and
Redis query durations, with a TraceQL search underneath.

Two things the stack deliberately does not do, so nobody goes looking for
them: the Redis instrumentation reports no query duration as a metric, so the
Redis duration is read out of the traces with TraceQL metrics, while the
Postgres one comes from the histogram `otelsql` records; and Tempo's metrics
generator is not enabled, so there is no service map and no
`traces_spanmetrics_*` in Prometheus. Trace relationships within a single trace
still work.

`deploy/observability/README.md` has the details, including how to revalidate
the configs after editing them.

### Where the telemetry goes

`TELEMETRY_EXPORTER` selects the exporter:

| Value | Behaviour |
| ----- | --------- |
| `stdout` (default) | Prints the traces and the metrics on the terminal. No collector needed, so the local loop is unchanged. |
| `otlp` | Ships them to an OpenTelemetry collector over gRPC or HTTP. |
| `none` | Drops them. The application keeps running. |

The rest of the configuration follows the OpenTelemetry specification
(`OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`,
`OTEL_EXPORTER_OTLP_INSECURE`, `OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG`,
`OTEL_METRIC_EXPORT_INTERVAL`, `OTEL_RESOURCE_ATTRIBUTES`), so the process can
be pointed at any OTLP backend without a code change. They are documented in
`.env.example`. The intervals the specification defines are counted in
milliseconds, so `OTEL_METRIC_EXPORT_INTERVAL=10000` is ten seconds — the SDK
reads the very same variable, and the two have to agree on the unit. Pointing
the exporter at a collector that is not listening never stops the application:
the connection is established in the background.

### Automatic instrumentation

Beyond the two business operations, the infrastructure is instrumented
automatically, so a request is traced all the way down:

| Component | Instrumentation | What it reports |
| --------- | --------------- | ----------------- |
| Gin | `otelgin` | `http.server.request.duration`, one span per request named after the route |
| PostgreSQL | `otelsql` | one span per statement, plus the query and the table it touched, and the `db.client.operation.duration` histogram |
| Redis | `redisotel` | one span per command, plus the pool metrics |
| Resend | `otelhttp` | `http.client.request.duration`, one span per call |
| asynq | middleware | one span per task, with the task id, the queue and the retry bookkeeping |

The tracing middleware is registered first, so the span of a request covers
the whole handler, and it rewrites `c.Request` with the span context: every
handler picks it up through `c.Request.Context()` without any change. Because
the instrumented clients capture the providers when they are built, both
composition roots set the telemetry up **before** connecting to the database
and to Redis.

One known limitation: `resend-go` builds its request without the caller's
context, so a call to Resend is reported as its own trace instead of hanging
under the invitation that triggered it. The outcome is still recorded by the
counters of the operation.

### Business operations

| Operation | Span | Metrics |
| --------- | ---- | ------- |
| Admin creates an invitation | `admin.invitation.create` | `requisition.admin.invitations.{attempted,succeeded,failed,duration}` |
| Admin updates a user's status | `admin.user.status.update` | `requisition.admin.user_status_updates.{attempted,succeeded,failed,duration}` |
| Admin changes a user's role | `admin.user.role.update` | `requisition.admin.user_role_updates.{attempted,succeeded,failed,duration}` |
| Admin revokes user sessions | `admin.user.sessions.revoke` | `requisition.admin.session_revocations.{attempted,succeeded,failed,duration}` |
| Admin resends an invitation | `admin.invitation.resend` | `requisition.admin.invitation_resends.{attempted,succeeded,failed,duration}` |
| Admin revokes an invitation | `admin.invitation.revoke` | `requisition.admin.invitation_revocations.{attempted,succeeded,failed,duration}` |
| Invitee accepts an invitation | `invitation.accept` | `requisition.invitee.acceptances.{attempted,succeeded,failed,duration}` |
| Worker delivers the invitation email | `admin.email.delivery` | `requisition.admin.email_delivery.{attempted,succeeded,failed,duration}` |
| Worker expires the stale invitations | `invitation.expire` | `requisition.invitation.expirations.{attempted,succeeded,failed,duration}` |
| Request a password reset | `auth.password_reset.request` | `requisition.auth.password_reset_requests.{attempted,succeeded,failed,duration}` |
| Complete a password reset | `auth.password_reset.complete` | `requisition.auth.password_reset_completions.{attempted,succeeded,failed,duration}` |
| Worker delivers the reset email | `auth.password_reset.email` | `requisition.auth.password_reset_emails.{attempted,succeeded,failed,duration}` |

The span covers the whole operation: the admin span covers the durable
invitation creation and the enqueueing of the delivery task, the acceptance
span covers the transaction that registers the user, creates the membership and
accepts the invitation. The counters carry the unit of what they count and the
histogram reports durations in seconds, so a Prometheus backend names them
`..._attempted_total` and `..._duration_seconds_bucket`.

The names are stable and low cardinality on purpose: no email, token, user or
invitation identifier is ever reported as a metric attribute. The measurement
never changes the outcome of the operation — the business error is always the
one returned to the caller.

### Worker trace continuity

An invitation spans two processes: the API queues the email task and the
worker delivers it. The two are the same trace, so the whole story of an
invitation is one thing to look at:

```
POST /api/admin/invitations        (API process)
└── asynq admin:send-invitation-email   (worker process)
    └── admin.email.delivery
        └── the statement that marks the invitation delivered
```

The traceparent of the request travels in the task payload, under the
`trace_parent` field. asynq hands its handler a context with the retry
bookkeeping but without the producer's span, so the tracing middleware reads
the field back and starts the consumer span as a child of it. Any task type
gets this by carrying the field; a task that carries nothing, the scheduled
`expire_invitations` sweep for instance, starts a trace of its own.

The reset-password email follows the same pattern: the forgot-password request
carries the traceparent in the `identity:send-password-reset-email` payload, and
the delivery span is a child of it. The scheduled cleanup task carries no
traceparent and starts a trace of its own.

The call to Resend is the one gap: `resend-go` builds its request without the
caller's context, so it stays a trace of its own, as described above. The
counters of `admin.email.delivery` and `auth.password_reset.email` still record
whether it worked.

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
go test ./internal/...                  # unit tests only (fast)
go test ./internal/department/...       # a single module
```

The `tests/integrations` suites start PostgreSQL and Redis containers through
testcontainers, so the full run needs Docker and takes longer than the unit
tests. They cover the password-reset repository and the forgot/reset HTTP
flows end to end.

CI (`.github/workflows/ci-cd.yml`) runs the full suite on every pull request and
pushes the image to GHCR on `master`.

## Regenerating API docs

Docs are generated from swag annotations in each `*_docs.go` file:

```bash
go run github.com/swaggo/swag/cmd/swag init -g main.go -d ./cmd/api,./internal -o docs
go run ./cmd/specgen      # converts docs/swagger.json -> docs/openapi.yaml
```

Only `docs/openapi.yaml` and `docs/embed.go` are tracked; the rest are build artifacts.
