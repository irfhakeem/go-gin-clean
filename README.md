# go-gin-clean

A Go backend template built with **Gin** and **Clean Architecture**.

## Features

- JWT auth + Google OAuth 2.0 (web & mobile)
- Outbox pattern + RabbitMQ (async events, retry queue, dead-letter queue)
- Redis caching
- PostgreSQL (GORM) + SQL migrations
- Email sending (SMTP)
- File storage (local, Cloudinary, or MinIO)
- Structured logging (Zap)

## Requirements

- Go 1.24+
- PostgreSQL, Redis, RabbitMQ (or `docker compose up -d`)

## Getting Started

```bash
git clone <repo-url>
cd go-gin-clean
go mod download

cp .env.example .env   # fill in your values

docker compose up -d           # start postgres, redis, rabbitmq
go run cmd/migrate/main.go up  # run migrations
go run cmd/server/main.go      # start the server
```

Server runs at `http://localhost:3000`.

## Project Layout

```
cmd/server/      entrypoint
cmd/migrate/     migration CLI
internal/
  delivery/      HTTP handlers, routes, middleware, MQ consumers
  usecase/       business logic
  repository/    database access
  gateway/       security, cache, storage, messaging, mailer
  entity/        domain models
  model/         request/response DTOs
  infrastructure/ dependency wiring, background workers
migrations/      SQL migrations
pkg/             config, errors, logger, utils
```

## API

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
| GET | `/api/v1/users/:id` | JWT | Get user by ID |
| POST | `/api/v1/users` | JWT | Create user |
| PUT | `/api/v1/users/:id` | JWT | Update user |
| PUT | `/api/v1/users/:id/change-status` | JWT | Toggle active status |
| DELETE | `/api/v1/users/:id` | JWT | Delete user |

Authenticated requests: `Authorization: Bearer <access_token>`

## Deployment

```bash
go build -ldflags="-s -w" -o bin/server ./cmd/server
go build -ldflags="-s -w" -o bin/migrate ./cmd/migrate

docker build -t go-gin-clean:latest .
docker run -d --name go-gin-clean -p 3000:3000 --env-file .env go-gin-clean:latest

# run migrations before starting the app
docker run --rm --env-file .env go-gin-clean:latest ./migrate up
```

Set `ENVIRONMENT=production` and use strong secrets for `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `AES_KEY`, `AES_IV`, and `OAUTH_STATE_STRING`.

## License

MIT
