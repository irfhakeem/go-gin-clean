<p align="center">
  <img width="1280" height="640" alt="Image" src="https://github.com/user-attachments/assets/5503c203-2869-4d83-95df-c5e905996d75" />
</p>

<p align="center">
  <strong>
    A Mini-Intermediate Scale Go Gin with Clean Architecture Boilterplate
  </strong>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/gin-gonic/gin"><img src="https://img.shields.io/badge/Gin-web--framework-00ADD8?logo=go&logoColor=white" alt="Gin"></a>
  <a href="https://gorm.io/"><img src="https://img.shields.io/badge/GORM-ORM-00ADD8" alt="GORM"></a>
  <a href="https://www.postgresql.org/"><img src="https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white" alt="PostgreSQL"></a>
  <a href="https://redis.io/"><img src="https://img.shields.io/badge/Redis-cache-DC382D?logo=redis&logoColor=white" alt="Redis"></a>
  <a href="https://www.rabbitmq.com/"><img src="https://img.shields.io/badge/RabbitMQ-AMQP-FF6600?logo=rabbitmq&logoColor=white" alt="RabbitMQ"></a>
  <a href="https://zap.go.dev/"><img src="https://img.shields.io/badge/Logging-Zap-00ADD8" alt="Zap"></a>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/iranhakim281/go-gin-clean"><img src="https://img.shields.io/badge/status-developing-green" alt="Status: developing"></a>
</p>

<p align="center">
  <a href="#-features">Features</a> &nbsp;•&nbsp;
  <a href="#-getting-started">Getting Started</a> &nbsp;•&nbsp;
  <a href="#-system-design">System Design</a> &nbsp;•&nbsp;
  <a href="#-api">API</a> &nbsp;•&nbsp;
  <a href="#-deployment">Deployment</a> &nbsp;•&nbsp;
  <a href="#-contributing">Contributing</a>
</p>

---

> **Access Control:** Includes two-tier authorization: route-level **RBAC** (`RequireRole` middleware) and domain-level **Data Scoping** (`UserPolicy` filtering).

## Features

| Area               | Details                                                                                                |
| ------------------ | ------------------------------------------------------------------------------------------------------ |
| Authentication     | JWT access + rotating refresh tokens (HttpOnly cookie, revocable), bcrypt password hashing             |
| Authorization      | Two-tier: RBAC middleware (`RequireRole`) + Domain Policy data scoping (`FullAccess` / `FilteredScope`)  |
| Event-driven email | Transactional **outbox pattern** → RabbitMQ with **retry queue &amp; dead-letter queue** → SMTP worker |
| Caching            | Redis (per-key TTL + pattern invalidation) with cache-aside and graceful fallback on cache errors      |
| API                | Gin REST API with unified JSON envelope, **i18n (EN/ID)** per `Accept-Language`, request validation    |
| Storage            | Cloudinary avatar upload with extension allow-list                                                     |
| Operations         | Structured logging (Zap), graceful shutdown, multi-stage non-root Docker image, plain-SQL migrations   |

## Getting Started

### Requirements

- Go 1.24+
- PostgreSQL, Redis, RabbitMQ (or `docker compose up -d`)

### Steps

```bash
git clone https://github.com/iranhakim281/go-gin-clean.git
cd go-gin-clean
go mod download

cp .env.example .env        # fill in your values

docker compose up -d        # start postgres, redis, rabbitmq
go run cmd/migrate up       # run migrations
go run cmd/server           # start the server
```

The server listens on `http://localhost:3000` by default
(`SERVER_HOST` / `SERVER_PORT` in `.env`).

<details>
<summary><b>Troubleshooting</b></summary>

- **Port already in use** — change `SERVER_PORT` in `.env`.
- **Outbox worker stuck-reset on startup** is logged as a restart recovery, not an error.
- `.env.example` documents every variable (DB, Redis, RabbitMQ, SMTP, JWT,
  AES, Cloudinary).
</details>

## System Design

| #   | Layer                   | Packages                                                                                    | Responsibility                                                                                                                                    | May import                                  |
| --- | ----------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| 1   | **Domain** (inner)      | `internal/domain/{entity, auth, permission, policy, vo}`                                    | Entities, value objects, invariants, RBAC checker & scoping policies. Zero framework deps.                                                       | stdlib, `pkg/*` only                        |
| 2   | **Application / ports** | `internal/application/port`, `internal/application/usecase`                                 | Use cases (login, register, email verification, …) + all **port interfaces** (repositories, `TokenMaker`, `Mailer`, `Cache`, `Storage`, …) | `domain`, `dto`, `pkg/*`                    |
| 3a  | **Inbound adapters**    | `internal/delivery/http/…` (handlers, routes, middleware, response), `internal/delivery/mq` | Translate HTTP / MQ messages → use-case calls; JSON envelope + i18n errors                                                                        | `application`, `domain`, `dto`, `pkg/*`     |
| 3b  | **Outbound adapters**   | `internal/infrastructure/{repository, messaging, mailer, storage, security}`                | Concrete implementations of the ports (GORM, amqp09, SMTP, Cloudinary, JWT/AES/bcrypt)                                                            | `application/port`, `domain`, `pkg/*`       |
| 4   | **Composition root**    | `internal/app.go`, `cmd/server`, `cmd/migrate`                                              | Manual constructor DI wiring everything together; HTTP server + background workers; migration CLI                                                 | everything                                  |
| —   | **Shared kernel**       | `pkg/{config, connection, logger, message, errors, utils}`, `internal/dto`                  | Leaf utilities: config/env, i18n catalogs, `AppError`/`ConsumerError`, shared DTOs                                                                | `pkg/*` is a leaf: imports nothing internal |

**Design decisions worth calling out**

- **Ports live next to use cases** (`internal/application/port`) so the
  application layer is self-contained: it defines _what_ it needs;
  `infrastructure/` defines _how_. `internal/app.go` is the only place where
  an adapter is injected into a use case — no DI framework.
- **`AppError` vs `ConsumerError`**: the HTTP side wraps failures into
  `AppError{type, i18n-key, cause}` which `delivery/http/response` renders
  into a consistent localized envelope; the MQ side returns
  `ConsumerError{retryable | non-retryable}` so the broker layer — not the
  business code — decides requeue vs. dead-letter.
- **Domain stays pure**: `entity.User` holds invariants (`VerifyEmail`,
  `SetPassword`, …); validation rules for _requests_ live in
  `internal/dto/validator` so domain types never grow HTTP concerns.

### Architecture & Runtime Visualization (`docs/architecture/`)

Interactive multi-level architecture and sequence diagrams (generated with Archify and pinned to source lines) are available in [`docs/architecture/index.html`](docs/architecture/index.html):

| Level | Diagram | What It Shows | Interactive View & Spec |
| ----- | ------- | ------------- | ----------------------- |
| **Portal** | **4-Level Drill-Down Index** | Overview, semantic color/arrow legend, and drill-down navigation (`Level 1 → Level 2 → Level 3 → Level 4`). | [`index.html`](docs/architecture/index.html) |
| **Level 1** | **System Context** | High-level boundary of **Go Gin Clean API**, external actors (**End User**, **Super Admin**), and external systems (**PostgreSQL**, **Redis**, **RabbitMQ**, **Cloudinary**, **SMTP**). | [`level-1-context.html`](docs/architecture/level-1-context/level-1-context.html) · [`candidate.json`](docs/architecture/level-1-context/candidate.json) |
| **Level 2** | **Containers & Runtime Processes** | Deployable Docker services (`postgres:16-alpine`, `redis:7-alpine`, `rabbitmq:3-management-alpine`), `cmd/migrate` CLI, and concurrent goroutines inside `cmd/server` (`Gin HTTP Server`, `OutboxWorker`, `ConsumerWorker`). | [`level-2-containers.html`](docs/architecture/level-2-containers/level-2-containers.html) · [`candidate.json`](docs/architecture/level-2-containers/candidate.json) |
| **Level 3** | **Internal Components (Clean Architecture)** | Package-level dependency map across Delivery (`HTTP Middleware`, `AuthHandler`, `UserHandler`, `EmailEventHandler`), Application (`AuthUseCase`, `UserUseCase`, `OutboxUseCase`, `EmailUseCase`), Domain (`Checker`, `UserPolicy`, `Entities`), and Infrastructure (`GORM + WithTx`, `JWT/Bcrypt/AES`, `Redis`, `Cloudinary`, `RabbitMQ/SMTP`). | [`level-3-components.html`](docs/architecture/level-3-components/level-3-components.html) · [`candidate.json`](docs/architecture/level-3-components/candidate.json) |
| **Level 4A** | **Register & Outbox Flow** | Step-by-step sequence for `POST /api/v1/auth/register`: atomic `WithTx` write (`userRepo.Create` + `outboxRepo.Create`), 5s `FOR UPDATE SKIP LOCKED` polling by `OutboxWorker`, RabbitMQ publish/retry/DLQ, and SMTP email delivery. | [`level-4-register-flow.html`](docs/architecture/level-4-register-flow/level-4-register-flow.html) · [`candidate.json`](docs/architecture/level-4-register-flow/candidate.json) |
| **Level 4B** | **Login & Atomic Refresh Flow** | Step-by-step sequence for `POST /api/v1/auth/login` and `POST /api/v1/auth/refresh-token`: HS256 JWT + AES-256-GCM token encryption, `HttpOnly` cookie handling, and atomic `WithTx` token rotation (`RevokeByToken` + `Save`). | [`level-4-login-refresh-flow.html`](docs/architecture/level-4-login-refresh-flow/level-4-login-refresh-flow.html) · [`candidate.json`](docs/architecture/level-4-login-refresh-flow/candidate.json) |
| **Level 4C** | **Two-Tier RBAC & Cache-Aside Flow** | Step-by-step sequence for `GET /api/v1/users`: Tier 1 route RBAC (`AuthMiddleware` + `RequireRole(UserRead)`), Tier 2 domain scoping (`UserPolicy.Scope`), scoped Redis cache-aside (`users:all:<role>:<id>:...`), and SQL-injection-safe sorting (`SanitizeSortColumn`). | [`level-4-users-rbac-flow.html`](docs/architecture/level-4-users-rbac-flow/level-4-users-rbac-flow.html) · [`candidate.json`](docs/architecture/level-4-users-rbac-flow/candidate.json) |

```bash
# Open the interactive drill-down portal locally
open docs/architecture/index.html
```

## Project Layout

```
cmd/
  server/            entrypoint (HTTP server + background workers)
  migrate/           migration CLI (up | down | force | version | create)
internal/
  app.go             composition root (manual wiring)
  dto/               shared request/response DTOs + request validation
  domain/            LAYER 1: core business rules (no framework deps)
    entity/          domain objects
    auth/            token claims
    permission/      RBAC permission definitions and role-permission matrix
    policy/          data scoping policies (ScopeAll, ScopeFiltered, ScopeNone)
    vo/              value objects (role, gender, …)
  application/
    port/            outbound port interfaces (repo, mq, mailer, jwt, cache, …)
    usecase/         application use cases
  delivery/
    http/            REST API (handlers, routes, middleware, response envelope)
    mq/              RabbitMQ consumers (email, events)
  infrastructure/    outbound adapters implementing the ports
    repository/      GORM repositories (users, refresh tokens, outbox)
    messaging/       RabbitMQ publisher/consumer (topology, retry, DLX)
    mailer/          SMTP email service + HTML templates
    storage/         Cloudinary file storage
    security/        JWT, AES (GCM + CTR), bcrypt
  worker/            background jobs (outbox dispatcher, email consumer)
migrations/          plain-SQL migrations
pkg/                 leaf utilities: config, connection, logger, message (i18n), errors, utils
```

## Data & Caching Strategy

- **Sessions**: access token (short-lived JWT) in the `Authorization` header;
  refresh token as an **HttpOnly cookie**, rotated on every login/refresh and
  revoked on logout (all sessions of a user are terminated on logout).
- **Cache-aside**: hot paths (`/users`, `/users/:id`) read Redis first with a
  5-minute TTL; mutations delete both the single-user key and the paginated
  list pattern (`users:all:*`). A cache hit/miss or Redis outage degrades
  gracefully to a direct DB read — it never fails the request.
- **Passwords**: bcrypt. **Token material**: AES-encrypted when stored.
- **Consistency**: email fan-out is driven from the `outbox_messages` table
  written in the same transaction as the state change, so no event is lost on
  a crash before publish.

## API

| Method | Path                                     | Auth  | Description               |
| ------ | ---------------------------------------- | ----- | ------------------------- |
| POST   | `/api/v1/auth/register`                  | —     | Register                  |
| POST   | `/api/v1/auth/login`                     | —     | Login                     |
| POST   | `/api/v1/auth/refresh-token`             | —     | Refresh access token      |
| POST   | `/api/v1/auth/verify-email`              | —     | Verify email              |
| POST   | `/api/v1/auth/send-reset-password`       | —     | Request password reset    |
| POST   | `/api/v1/auth/reset-password`            | —     | Reset password            |
| POST   | `/api/v1/auth/resend-verification`       | —     | Resend verification email |
| GET    | `/api/v1/profile`                        | JWT   | Get own profile           |
| PUT    | `/api/v1/profile`                        | JWT   | Update profile            |
| PUT    | `/api/v1/profile/change-password`        | JWT   | Change password           |
| POST   | `/api/v1/profile/logout`                 | JWT   | Logout (revoke all)       |
| GET    | `/api/v1/users`                          | JWT\* | List users (paginated, scoped)    |
| GET    | `/api/v1/users/:id`                      | JWT\* | Get user by ID (scoped)           |
| POST   | `/api/v1/users`                          | JWT\* | Create user (super_admin)         |
| PUT    | `/api/v1/users/:id`                      | JWT\* | Update user (scoped)              |
| PUT    | `/api/v1/users/:id/change-status`        | JWT\* | Toggle active status (scoped)     |
| DELETE | `/api/v1/users/:id`                      | JWT\* | Delete user (super_admin)         |

`GET /health` returns a plain liveness probe.

\* Protected by `RequireRole` (RBAC) and filtered by `UserPolicy` data scoping.

**Auth**: `Authorization: Bearer <access_token>`; the refresh token is an
HttpOnly `refresh_token` cookie. **i18n**: `Accept-Language: en | id`
(default `en`).

### Response envelope

```jsonc
// success
{ "status": true, "message": "Login successful", "data": { "...": "..." }, "meta": { "page": 1, "per_page": 10, "total": 42, "total_pages": 5 } }

// error
{ "status": false, "code": "TOKEN_INVALID", "message": "Token is invalid", "errors": { "email": "..." } }
```

## Deployment

Multi-stage Docker build (builder → alpine, non-root `appuser`):

```bash
docker build -t go-gin-clean:latest .
docker run -d --name go-gin-clean -p 3000:3000 --env-file .env go-gin-clean:latest

# run migrations before starting the app
docker run --rm --env-file .env go-gin-clean:latest ./migrate up
```

Production checklist: set `ENVIRONMENT=production` and use strong secrets for
`JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `AES_KEY`, and `AES_IV`. Put the API
behind TLS (CORS + `Secure` cookies are configured accordingly).

## Development

- [.air](https://github.com/air-verse/air) is configured (`.air.toml`) for
  live-reload: `air` — rebuilds `cmd/server` on changes.
- Migrations are plain SQL in `migrations/`; scaffold one with
  `go run cmd/migrate create your_migration_name`.
- Ports and adapters: to swap an implementation (e.g. Redis → in-memory cache),
  implement the matching interface in `internal/application/port` and change
  the single wiring line in `internal/app.go`.

## Contributing

Contributions are welcome!

1. Fork the repo and create a branch: `git checkout -b feat/my-feature`
2. Keep the dependency rule intact — new infrastructure must implement a port;
   use cases must not import infrastructure.
3. Run `gofmt -s` and `go vet ./...` before pushing.
4. Open a PR with a clear description of the change.

Please open an issue first for major features or behavior changes.

## License

Distributed under the [MIT License](LICENSE).
