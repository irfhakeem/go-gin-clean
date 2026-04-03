# go-gin-clean

A Go backend template built with **Gin**, **Clean Architecture**, and production-ready patterns: JWT + OAuth 2.0 (web & mobile), outbox-based async messaging, Redis caching, and PostgreSQL.

---

## Architecture Overview

The codebase follows Clean Architecture with a strict inward dependency rule — outer layers depend on inner layers, never the reverse.

```
HTTP Request
     │
     ▼
┌─────────────┐
│   Delivery  │  Handlers, middleware, routes (Gin)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   UseCase   │  Business logic, orchestration
└──────┬──────┘
       │
  ┌────┴─────┐
  ▼          ▼
┌──────┐  ┌─────────┐
│ Repo │  │ Gateway │  Database / Security / Cache / Messaging
└──────┘  └─────────┘
       │
       ▼
┌─────────────┐
│   Entity    │  Domain models, pure Go, no dependencies
└─────────────┘
```

### Outbox Pattern

Events (e.g. `user.register`, `user.reset_password`) are written to the `outbox_messages` table inside the same DB transaction as the business operation. A background worker polls the table every 5 seconds and publishes them to RabbitMQ — guaranteeing at-least-once delivery with retry and dead-letter tracking.

```
Register/Reset → SaveOutboxMessage (DB) → OutboxWorker polls → PublisherService → RabbitMQ
```

### OAuth 2.0 (Web + Mobile)

```
Client → POST /auth/oauth2/url (provider, app_id, platform)
       ← { auth_url }
       → Redirect to Google
       ← Google redirects to /auth/oauth2/google/callback
       → State validated, tokens issued
       ← Web:    302 → FRONTEND_URL/oauth/callback#access_token=...
       ← Mobile: 302 → deep-link://oauth?access_token=...&refresh_token=...
```

---

## Directory Structure

```
go-gin-clean/
├── cmd/
│   ├── server/main.go          # App entrypoint
│   └── migrate/main.go         # Migration CLI
│
├── internal/
│   ├── delivery/http/          # Handlers, middleware, routes
│   ├── usecase/                # Business logic
│   │   ├── user_usecase.go
│   │   └── outbox_usecase.go
│   ├── repository/             # Database access (GORM)
│   │   ├── user_repository.go
│   │   ├── refresh_token_repository.go
│   │   └── outbox_repository.go
│   ├── gateway/
│   │   ├── security/           # JWT, Bcrypt, AES, OAuth
│   │   ├── cache/              # Redis
│   │   ├── media/              # Cloudinary / local storage
│   │   └── messaging/          # RabbitMQ publisher
│   ├── entity/                 # Domain models
│   ├── model/                  # DTOs, request/response structs
│   └── infrastructure/
│       ├── container.go        # Dependency wiring
│       └── worker/
│           └── outbox_worker.go
│
├── migrations/                 # SQL migration files (golang-migrate)
├── pkg/
│   ├── config/                 # Env-based configuration
│   ├── errors/                 # Centralized error definitions
│   └── utils/
├── assets/                     # Local file storage
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Setup & Running

### Prerequisites

- Go 1.24+
- PostgreSQL 16+
- Redis 7+
- RabbitMQ 3+
- Docker & Docker Compose (for local infra)

### 1. Clone & install dependencies

```bash
git clone <repo-url>
cd go-gin-clean
go mod download
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values. Key variables:

```env
# Server
SERVER_PORT=3000
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=go_clean_architecture

# JWT
JWT_ACCESS_SECRET=your-access-secret
JWT_REFRESH_SECRET=your-refresh-secret
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# AES (32-char key, 16-char IV)
AES_KEY=your-32-character-aes-key-here
AES_IV=your-16-char-iv

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:3000/api/v1/auth/oauth2/google/callback
GOOGLE_ALLOWED_ORIGINS=http://localhost:3120

# Multi-app OAuth (web)
# Format: app_id=url;app_id2=url2
FRONTEND_URLS=default=http://localhost:3120

# Mobile deep links
# Format: app_id=scheme://path;app_id2=scheme://path
MOBILE_DEEP_LINKS=android=com.example.app://oauth/callback

DEFAULT_APP_ID=default

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_EXCHANGE=main_event_bus
```

### 3. Start infrastructure

```bash
docker compose up -d
```

This starts PostgreSQL, Redis, and RabbitMQ. RabbitMQ management UI is available at `http://localhost:15672` (guest/guest).

### 4. Run migrations

```bash
go run cmd/migrate/main.go up
```

Other migration commands:

```bash
go run cmd/migrate/main.go down            # Rollback last
go run cmd/migrate/main.go version         # Current version
go run cmd/migrate/main.go create <name>   # New migration pair
```

### 5. Run the server

```bash
go run cmd/server/main.go
```

**With hot reload (Air):**

```bash
go install github.com/cosmtrek/air@latest
air
```

Server starts at `http://localhost:3000`.

---

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | — | Register |
| POST | `/api/v1/auth/login` | — | Login |
| POST | `/api/v1/auth/refresh-token` | — | Refresh access token |
| POST | `/api/v1/auth/verify-email` | — | Verify email |
| POST | `/api/v1/auth/send-reset-password` | — | Request password reset |
| POST | `/api/v1/auth/reset-password` | — | Reset password |
| POST | `/api/v1/auth/resend-verification` | — | Resend verification email |
| POST | `/api/v1/auth/oauth2/url` | — | Get OAuth login URL |
| GET | `/api/v1/auth/oauth2/:provider/callback` | — | OAuth callback |
| GET | `/api/v1/profile` | JWT | Get own profile |
| PUT | `/api/v1/profile` | JWT | Update profile |
| PUT | `/api/v1/profile/change-password` | JWT | Change password |
| POST | `/api/v1/profile/logout` | JWT | Logout |
| GET | `/api/v1/users` | JWT | List users (paginated) |
| GET | `/api/v1/users/:code` | JWT | Get user by code |
| POST | `/api/v1/users` | JWT | Create user |
| PUT | `/api/v1/users/:code` | JWT | Update user |
| PUT | `/api/v1/users/:code/change-status` | JWT | Toggle active status |
| DELETE | `/api/v1/users/:code` | JWT | Delete user |

**Authorization header:** `Authorization: Bearer <access_token>`

---

## Deployment

### Build binary

```bash
go build -ldflags="-s -w" -o bin/server ./cmd/server
go build -ldflags="-s -w" -o bin/migrate ./cmd/migrate
```

### Docker

```bash
# Build
docker build -t go-gin-clean:latest .

# Run (pass env file or individual -e flags)
docker run -d \
  --name go-gin-clean \
  -p 3000:3000 \
  --env-file .env \
  go-gin-clean:latest
```

### Run migrations before deploying

The migration binary is included in the Docker image at `/app/migrate`:

```bash
# From inside the container or as an init container
./migrate up
```

Or run it as a separate one-off job before starting the app container:

```bash
docker run --rm --env-file .env go-gin-clean:latest ./migrate up
```

### Production checklist

- [ ] Set `ENVIRONMENT=production`
- [ ] Use strong, randomly generated values for `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `AES_KEY`, `AES_IV`, `OAUTH_STATE_STRING`
- [ ] Point `GOOGLE_REDIRECT_URL` to your production domain
- [ ] Run migrations before deploying a new version
- [ ] Ensure RabbitMQ exchange (`RABBITMQ_EXCHANGE`) exists and is durable

---

## License

MIT
