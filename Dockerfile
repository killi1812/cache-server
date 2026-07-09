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

# Define build arguments
ARG BUILD=prod
ARG VERSION=0.0.0
ARG COMMIT_HASH=n/a
ARG BUILD_TIMESTAMP=n/a
ARG PACKAGE="github.com/killi1812/go-cache-server"

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-X '${PACKAGE}/app.Build=${BUILD}' -X '${PACKAGE}/app.Version=${VERSION}' -X '${PACKAGE}/app.CommitHash=${COMMIT_HASH}' -X '${PACKAGE}/app.BuildTimestamp=${BUILD_TIMESTAMP}'" \
    -o cache-server main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /build/cache-server .

# Create directory for local storage
RUN mkdir -p /app/binary-caches

# Expose the default ports and a range for dynamic caches
EXPOSE 12345 54321 10000-10100

# Re-declare build arguments to make them available in runtime stage ENV
ARG BUILD=prod
ARG VERSION=0.0.0
ARG COMMIT_HASH=n/a
ARG BUILD_TIMESTAMP=n/a

ENV APP_BUILD=${BUILD}
ENV APP_VERSION=${VERSION}
ENV APP_COMMIT_HASH=${COMMIT_HASH}
ENV APP_BUILD_TIMESTAMP=${BUILD_TIMESTAMP}

# Define entrypoint
ENTRYPOINT ["./cache-server"]

# Default command
CMD ["listen", "-f"]
