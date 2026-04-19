# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Install build dependencies for CGO (required for sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy go mod and sum files first
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY src/ .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o cache-server main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
# nix is required for agent-node to perform nix-store operations
RUN apk add --no-cache ca-certificates nix

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /build/cache-server .

# Create directory for local storage
RUN mkdir -p /app/binary-caches

# Expose the default ports and a range for dynamic caches
EXPOSE 12345 54321 10000-10100

# Define entrypoint
ENTRYPOINT ["./cache-server"]

# Default command
CMD ["listen", "-f"]
