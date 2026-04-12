# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Install build dependencies for CGO (required for sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod and sum files first to leverage Docker cache
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY src/ .

# Build the application
# CGO_ENABLED=1 is required for the sqlite driver
RUN CGO_ENABLED=1 GOOS=linux go build -o cache-server main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates musl

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/cache-server .

# Create directory for local storage
RUN mkdir -p /app/binary-caches

# Expose the default ports and a range for dynamic caches
EXPOSE 12345 10000-10100

# Define entrypoint
ENTRYPOINT ["./cache-server"]

# Default command
CMD ["listen", "-f"]
