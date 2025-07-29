# Go Gin Clean Architecture

A complete Go web application implementing Clean Architecture principles with proper separation of concerns and dependency inversion.

## 🏗️ Architecture Overview

This project follows **Clean Architecture** principles, organizing code into clear layers with proper dependency direction:

```
┌─────────────────────────────────────────┐
│            External Interfaces          │
│        (Database, Web, CLI, etc.)       │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│               Adapters                  │
│         (Controllers, Gateways,         │
│          Presenters, etc.)              │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│            Use Cases                    │
│        (Application Business            │
│             Rules)                      │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│              Entities                   │
│         (Enterprise Business            │
│              Rules)                     │
└─────────────────────────────────────────┘
```

## 📁 Project Structure

```
go-gin-clean/
├── cmd/                          # Application entrypoints
│   ├── server/main.go           # HTTP server
│   └── migrate/main.go          # Database migrations
├── internal/                    # Private application code
│   ├── core/                    # Core business logic (innermost layer)
│   │   ├── domain/              # Enterprise business rules
│   │   │   ├── entities/        # Business entities
│   │   │   ├── valueobjects/    # Value objects
│   │   │   ├── enums/           # Enumerations
│   │   │   └── errors/          # Domain errors
│   │   ├── dto/                 # Data Transfer Objects
│   │   ├── ports/               # Interfaces (contracts)
│   │   └── usecases/            # Application business rules
│   ├── adapters/                # Adapters for external interfaces
│   │   ├── primary/http/        # HTTP handlers, routes, middleware
│   │   └── secondary/           # Database, security, email implementations
│   └── infrastructure/          # Infrastructure concerns
└── pkg/                         # Public libraries
    ├── config/                  # Configuration
    └── utils/                   # Utility functions
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL
- Git

### Setup

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd go-gin-clean
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Environment setup**
   Copy `.env.example` to `.env` and configure:

   ```env
   # Server
   SERVER_HOST=localhost
   SERVER_PORT=8080
   ENVIRONMENT=development

   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USERNAME=postgres
   DB_PASSWORD=your_password
   DB_NAME=go_gin_clean

   # JWT
   JWT_ACCESS_SECRET=your-access-secret-key
   JWT_REFRESH_SECRET=your-refresh-secret-key
   JWT_ACCESS_EXPIRY=1h
   JWT_REFRESH_EXPIRY=168h
   ```

4. **Database migration**

   ```bash
   go run cmd/migrate/main.go migrate
   ```

5. **Start the server**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

## 📚 API Documentation

### Health Check

- `GET /health` - Server health status

### Authentication

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh-token` - Refresh access token
- `GET /api/v1/auth/verify-email` - Email verification

### Protected Routes (require authentication)

- `POST /api/v1/logout` - User logout
- `POST /api/v1/change-password` - Change user password

### User Management

- `GET /api/v1/users` - Get all users (paginated)
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## 🔧 Available Commands

### Database Commands

```bash
# Run migrations
go run cmd/migrate/main.go migrate

# Rollback migrations
go run cmd/migrate/main.go rollback

# Fresh migrations (rollback + migrate)
go run cmd/migrate/main.go fresh
```

### Development Commands

```bash
# Start server
go run cmd/server/main.go

# Build server
go build -o bin/server cmd/server/main.go

# Build migration tool
go build -o bin/migrate cmd/migrate/main.go

# Run tests
go test ./...

# Check for issues
go vet ./...
```

## 🏛️ Clean Architecture Benefits

### 1. **Testability**

- Easy to unit test business logic in isolation
- Mock interfaces for testing
- No external dependencies in core logic

### 2. **Maintainability**

- Clear separation of concerns
- Changes in one layer don't affect others
- Easy to understand and modify

### 3. **Flexibility**

- Easy to swap implementations (database, web framework, etc.)
- Can add new interfaces without changing core logic
- Framework-independent business logic

### 4. **Scalability**

- Well-organized code structure
- Easy to add new features
- Clear boundaries between components

## 🧪 Testing

The architecture makes testing straightforward:

```go
// Example: Testing use case with mocked dependencies
func TestUserUseCase_Login(t *testing.T) {
    // Arrange
    mockUserRepo := &mocks.UserRepository{}
    mockJWTService := &mocks.JWTService{}

    useCase := usecases.NewUserUseCase(mockUserRepo, nil, mockJWTService, nil, nil)

    // Act & Assert
    // Test your business logic
}
```

## 🔒 Authentication

The application uses JWT-based authentication:

- **Access Token**: Short-lived token for API requests (1 hour)
- **Refresh Token**: Long-lived token for obtaining new access tokens (7 days)

Include the access token in requests:

```
Authorization: Bearer <access_token>
```

## 🛠️ Development Guidelines

### Adding New Features

1. **Start with Domain**: Define entities, value objects, and domain rules
2. **Define Ports**: Create interfaces for new repositories or services
3. **Implement Use Cases**: Add business logic in use cases
4. **Create Adapters**: Implement interfaces in appropriate adapters
5. **Wire Dependencies**: Update container for dependency injection

### Error Handling

- Domain errors defined in `internal/core/domain/errors/`
- Use meaningful error messages
- Propagate errors up through layers
- Handle errors appropriately in adapters

## 📈 Performance & Production

### Build for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Set production environment
export ENVIRONMENT=production
```

### Docker (Optional)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```
