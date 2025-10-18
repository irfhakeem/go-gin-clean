# Go Gin Clean Architecture

A complete Go web application implementing Clean Architecture principles with proper separation of concerns, dependency inversion, and framework-independent core business logic.

## 🏗️ Architecture Overview

This project follows **Clean Architecture** principles with proper separation of concerns and framework independence. The core business logic is completely isolated from external dependencies:

```
┌─────────────────────────────────────────────────────────────┐
│                     Adapters Layer                          │
├─────────────────────────┬───────────────────────────────────┤
│    Primary (HTTP)       │       Secondary (Infrastructure)  │
│                         │                                   │
│  ┌─────────────────┐   │   ┌─────────────────────────────┐ │
│  │ DTOs (Framework │   │   │ Database, SMTP, JWT,        │ │
│  │ Specific)       │   │   │ Media Services              │ │
│  └─────────────────┘   │   └─────────────────────────────┘ │
│           │             │                                   │
│  ┌─────────────────┐   │                                   │
│  │ Mappers         │   │                                   │
│  └─────────────────┘   │                                   │
└─────────┬───────────────┴───────────────────────────────────┘
          │ (Contracts)
┌─────────▼───────────────────────────────────────────────────┐
│                     Core Layer                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Contracts       │  │ Use Cases       │  │ Entities    │ │
│  │ (Clean DTOs)    │  │ (Business Logic)│  │ (Domain)    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**Key Architectural Benefits:**
- **Framework Independence**: Core layer has zero dependencies on HTTP frameworks, databases, or external libraries
- **Testability**: Business logic can be tested in isolation with pure domain objects
- **Flexibility**: Easy to swap HTTP frameworks (Gin → Fiber/Echo) or databases without affecting business logic
- **Maintainability**: Clear boundaries and single responsibility for each layer

## 📁 Project Structure

```
go-gin-clean/
├── cmd/                          # Application entrypoints
│   ├── server/main.go           # HTTP server
│   └── migrate/main.go          # Database migrations
├── internal/                    # Private application code
│   ├── core/                    # Core business logic (framework-independent)
│   │   ├── contracts/           # Clean, framework-agnostic DTOs
│   │   │   ├── user_contracts.go      # User-related contracts
│   │   │   └── pagination_contracts.go # Pagination contracts
│   │   ├── domain/              # Enterprise business rules
│   │   │   ├── entities/        # Business entities (User, RefreshToken, Audit)
│   │   │   ├── enums/           # Enumerations (Gender)
│   │   │   └── errors/          # Domain errors
│   │   ├── ports/               # Interfaces (use contracts, not DTOs)
│   │   │   ├── repositories.go  # Repository interfaces
│   │   │   ├── services.go      # Service interfaces
│   │   │   └── usecases.go      # Use case interfaces
│   │   └── usecases/            # Application business rules
│   │       ├── user_usecase.go  # User business logic
│   │       └── email_usecase.go # Email business logic
│   ├── adapters/                # Adapters for external interfaces
│   │   ├── primary/http/        # HTTP layer (framework-specific)
│   │   │   ├── dto/             # HTTP DTOs with framework bindings
│   │   │   │   ├── user_dto.go     # Gin-specific user DTOs
│   │   │   │   └── pagination_dto.go # Gin-specific pagination DTOs
│   │   │   ├── mappers/         # Convert DTOs ↔ Contracts
│   │   │   │   ├── interfaces.go    # Mapper interfaces
│   │   │   │   ├── user_mapper.go   # User mapping implementation
│   │   │   │   └── pagination_mapper.go # Pagination mapping
│   │   │   ├── handlers/        # HTTP handlers (use mappers)
│   │   │   ├── messages/        # Response messages
│   │   │   ├── response/        # Response utilities
│   │   │   ├── middleware.go    # Authentication middleware
│   │   │   └── routes.go        # Route definitions
│   │   └── secondary/           # External service implementations
│   │       ├── database/        # Database repositories
│   │       ├── security/        # JWT, Bcrypt, AES services (use contracts)
│   │       ├── mailer/          # SMTP email service
│   │       └── media/           # Local storage service (framework-independent)
│   └── infrastructure/          # Infrastructure concerns
│       └── container.go         # Dependency injection
└── pkg/                         # Public libraries
    ├── config/                  # Configuration management
    └── utils/                   # Utility functions
```

**Layer Responsibilities:**
- **Core/Contracts**: Clean data structures for inter-layer communication
- **Core/Domain**: Pure business entities and rules (no external dependencies)
- **Core/Ports**: Interfaces defining contracts between layers
- **Core/UseCases**: Business logic using contracts for communication
- **Adapters/Primary/HTTP**: Web layer with framework-specific DTOs and mappers
- **Adapters/Secondary**: Infrastructure implementations (database, services)

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
   APP_FE_URL=
   TIMEOUT=30

   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USERNAME=postgres
   DB_PASSWORD=your_password
   DB_NAME=go_gin_clean
   DB_MAX_IDLE_CONNS=25
   DB_MAX_OPEN_CONNS=5

   # JWT
   JWT_ISSUER=go-gin-clean
   JWT_ACCESS_SECRET=your-access-secret-key
   JWT_REFRESH_SECRET=your-refresh-secret-key
   JWT_ACCESS_EXPIRY=1h
   JWT_REFRESH_EXPIRY=168h

   # AES Encryption
   AES_KEY=your-32-character-encryption-key
   AES_IV=your-16-character-iv-key

   # SMTP Email (optional)
   MAILER_HOST=smtp.gmail.com
   MAILER_PORT=587
   MAILER_SENDER=your-email@gmail.com
   MAILER_AUTH=your-email@gmail.com
   MAILER_PASSWORD=your-app-password
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

### Authentication (Public Routes)

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login (sets refresh token in cookie)
- `POST /api/v1/auth/refresh-token` - Refresh access token
- `POST /api/v1/auth/verify-email` - Email verification
- `POST /api/v1/auth/send-verify-email` - Send verification email
- `POST /api/v1/auth/send-reset-password` - Send reset password email
- `POST /api/v1/auth/reset-password` - Reset password with token

### Profile Management (Protected Routes)

- `GET /api/v1/profile` - Get current user profile
- `PUT /api/v1/profile` - Update current user profile
- `POST /api/v1/profile/change-password` - Change user password
- `POST /api/v1/profile/logout` - User logout

### User Management (Protected Routes)

- `GET /api/v1/users` - Get all users (paginated)
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Static Assets

- `GET /assets/*` - Serve static files from assets directory

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

### 1. **Framework Independence**

- Core business logic has **zero dependencies** on Gin, HTTP, or external frameworks
- Easy to switch from Gin to Fiber, Echo, or any other HTTP framework
- Business rules remain unchanged when external dependencies change
- **Contracts layer** ensures clean communication between layers

### 2. **Testability**

- Use cases can be tested with pure domain objects (no mocking of framework types)
- **Mappers** can be unit tested independently
- Business logic isolated from HTTP concerns and database specifics
- Mock interfaces at the ports level for comprehensive testing

### 3. **Maintainability**

- **Clear separation**: DTOs (HTTP layer) vs Contracts (domain layer)
- **Single Responsibility**: Each layer has a specific, well-defined purpose
- **Dependency Rule**: Inner layers never depend on outer layers
- Easy to understand, modify, and extend

### 4. **Flexibility & Scalability**

- **Plug-and-play architecture**: Swap implementations without affecting business logic
- **Mapper pattern**: Clean conversion between external data formats and domain contracts
- Add new delivery mechanisms (GraphQL, gRPC) without changing use cases
- Horizontal scaling through clear component boundaries

### 5. **Domain-Driven Design**

- **Pure domain entities** with no external dependencies
- **Contracts** represent the true business data structures
- Business rules concentrated in the use case layer
- Framework concerns isolated in adapter layers

## 🧪 Testing

The Clean Architecture with contracts makes testing straightforward and framework-independent:

### Use Case Testing (Pure Business Logic)
```go
// Test use cases with contracts - no framework dependencies
func TestUserUseCase_Login(t *testing.T) {
    // Arrange
    mockUserRepo := &mocks.UserRepository{}
    mockJWTService := &mocks.JWTService{}
    mockBcryptService := &mocks.BcryptService{}

    useCase := usecases.NewUserUseCase(mockUserRepo, mockJWTService, mockBcryptService)

    loginReq := &contracts.LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    }

    // Act
    result, err := useCase.Login(context.Background(), loginReq)

    // Assert - pure domain testing
    assert.NoError(t, err)
    assert.NotEmpty(t, result.AccessToken)
}
```

### Mapper Testing (Conversion Logic)
```go
// Test mappers independently
func TestUserMapper_LoginRequestToContract(t *testing.T) {
    mapper := mappers.NewUserMapper()

    dtoReq := &dto.LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    }

    contractReq := mapper.LoginRequestToContract(dtoReq)

    assert.Equal(t, dtoReq.Email, contractReq.Email)
    assert.Equal(t, dtoReq.Password, contractReq.Password)
}
```

### Handler Testing (HTTP Layer)
```go
// Test handlers with mocked mappers and use cases
func TestUserHandler_Login(t *testing.T) {
    mockUseCase := &mocks.UserUseCase{}
    mockMapper := &mocks.UserMapper{}

    handler := handlers.NewUserHandler(mockUseCase, mockMapper)

    // Test HTTP concerns separately from business logic
}
```

## 🔐 Security Features

### Password Security

- **Bcrypt Hashing**: Industry-standard password hashing
- **Salt Generation**: Automatic salt generation for each password
- **Cost Factor**: Configurable cost factor for security vs performance

### JWT Security

- **HMAC Signing**: Secure token signing with secret keys
- **Token Expiration**: Configurable expiration times
- **Refresh Rotation**: Secure refresh token rotation

### Data Encryption

- **AES Encryption**: Additional data encryption capabilities
- **Configurable Keys**: Environment-based encryption keys
- **PKCS7 Padding**: Standard padding for block cipher

## 🔒 Authentication

The application uses JWT-based authentication with the following features:

- **Access Token**: Short-lived token for API requests (1 hour)
- **Refresh Token**: Long-lived token stored in HTTP-only cookie (7 days)
- **Password Hashing**: Bcrypt for secure password storage
- **AES Encryption**: Additional data encryption capabilities

### Authentication Flow

1. **Login**: User provides email/password, receives access token and refresh token (in cookie)
2. **API Requests**: Include access token in Authorization header
3. **Token Refresh**: Automatic refresh using HTTP-only cookie
4. **Logout**: Clears refresh token and invalidates session

Include the access token in requests:

```
Authorization: Bearer <access_token>
```

## 📧 Email Features

The application includes comprehensive email functionality:

- **Email Verification**: Send verification emails to new users
- **Password Reset**: Send password reset emails with secure tokens
- **Template System**: HTML email templates with dynamic data
- **SMTP Integration**: Configurable SMTP service for email delivery

## 📁 File Upload

Local file storage implementation:

- **Avatar Upload**: Users can upload profile pictures
- **Local Storage**: Files stored in local filesystem
- **File Validation**: Type and size validation
- **Secure Paths**: Protected file path handling

## 🛠️ Development Guidelines

### Adding New Features (Clean Architecture Flow)

1. **Start with Domain**: Define entities, value objects, and domain rules in `internal/core/domain/`
2. **Create Contracts**: Define clean data structures in `internal/core/contracts/`
3. **Define Ports**: Create interfaces in `internal/core/ports/` using contracts (not DTOs)
4. **Implement Use Cases**: Add business logic in `internal/core/usecases/` using contracts
5. **Create Secondary Adapters**: Implement infrastructure services in `internal/adapters/secondary/`
6. **Add HTTP DTOs**: Create framework-specific DTOs in `internal/adapters/primary/http/dto/`
7. **Create Mappers**: Build mappers in `internal/adapters/primary/http/mappers/` to convert DTOs ↔ Contracts
8. **Add HTTP Handlers**: Create handlers in `internal/adapters/primary/http/handlers/` using mappers
9. **Wire Dependencies**: Update `internal/infrastructure/container.go`
10. **Update Routes**: Add new routes in `internal/adapters/primary/http/routes.go`

### Architecture Rules

- **Dependency Rule**: Core layer NEVER imports from adapters layer
- **Use Contracts**: Use cases communicate via contracts, never DTOs
- **Map at Boundaries**: Convert DTOs to contracts at the HTTP boundary using mappers
- **Framework Isolation**: Keep framework-specific code (Gin, GORM) in adapters layer only

### Error Handling

- **Domain errors** defined in `internal/core/domain/errors/`
- **Contract-based** error handling in use cases
- **HTTP-specific** error responses in `internal/adapters/primary/http/messages/`
- **Consistent format** across all API endpoints

### Layer Communication

```go
// ❌ Wrong - Use case importing DTO
func (uc *UserUseCase) Login(req *dto.LoginRequest) error

// ✅ Correct - Use case using contracts
func (uc *UserUseCase) Login(req *contracts.LoginRequest) error

// ❌ Wrong - Handler calling use case directly with DTO
result, err := h.userUseCase.Login(&req)

// ✅ Correct - Handler using mapper
contractReq := h.userMapper.LoginRequestToContract(&req)
contractResult, err := h.userUseCase.Login(contractReq)
result := h.userMapper.LoginResponseToDTO(contractResult)
```

### Current Service Implementations

**Core Services (Framework-Independent):**
- **UserUseCase**: Business logic using contracts for all operations
- **EmailUseCase**: Email verification and password reset workflows
- **Contracts**: Clean data structures (LoginRequest, UserInfo, etc.)

**Infrastructure Services (Framework-Specific):**
- **JWTService**: Token generation/validation using contracts
- **BcryptService**: Password hashing
- **EncryptionService**: AES encryption/decryption
- **MailerService**: SMTP email with HTML templates
- **MediaService**: File storage with framework-independent interface

**HTTP Services:**
- **Mappers**: Convert between HTTP DTOs and domain contracts
- **Handlers**: HTTP request/response handling using mappers
- **DTOs**: Gin-specific data structures with binding tags

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
COPY --from=builder /app/templates ./templates
CMD ["./server"]
```

## 🌟 Features

### Core Features

- ✅ User Registration and Authentication
- ✅ JWT Token-based Authentication with Refresh Tokens
- ✅ Email Verification System
- ✅ Password Reset Functionality
- ✅ User Profile Management
- ✅ File Upload (Avatar)
- ✅ Pagination Support
- ✅ CORS Configuration
- ✅ Middleware Authentication

### Security Features

- ✅ Bcrypt Password Hashing
- ✅ JWT Token Security
- ✅ AES Data Encryption
- ✅ HTTP-Only Cookie for Refresh Tokens
- ✅ Input Validation and Sanitization

### Infrastructure Features

- ✅ Database Migration System
- ✅ Configuration Management
- ✅ SMTP Email Integration
- ✅ Local File Storage
- ✅ Structured Logging
- ✅ Graceful Shutdown
