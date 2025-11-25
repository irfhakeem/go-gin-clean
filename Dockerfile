# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o main ./cmd/server

RUN go build -o migrate ./cmd/migrate

# Production stage
FROM alpine:latest

# Install ca-certificates if needed for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

COPY --from=builder /app/.env .

# Copy the binary from builder stage, migrations, and other necessary files
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/assets ./assets

# Change ownership to non-root user
RUN chown appuser:appgroup main

# Switch to non-root user
USER appuser

EXPOSE ${INTERNAL_PORT}

# Run the binary
CMD ["./main"]
