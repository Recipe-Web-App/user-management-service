# User Management Service - Production Dockerfile
# Multi-stage build for optimal production image size

# Stage 1: Builder
# using alpine based go image for smaller footprint
FROM golang:1.24-alpine AS builder
LABEL stage=builder

# Install required build tools
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
# using -x to verbose output and ensure download
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to strip debug information and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/server ./cmd/api

# Stage 2: Runtime
# using alpine for minimal runtime environment
FROM alpine:3.23 AS runner
LABEL stage=runner \
    service="user-management-service" \
    maintainer="Recipe App Team"

# Install basics
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/server ./server

# Copy configuration files
# The application expects config files in ./config, relative to the working directory
COPY --from=builder /app/config ./config

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose server port
EXPOSE 8080

# Run the server
CMD ["./server"]
