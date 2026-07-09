#!/bin/sh
set -e

# Gather build information
BUILD="prod"
VERSION=$(git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null | sed 's/^.//')
if [ -z "$VERSION" ]; then
  VERSION="0.0.0"
fi
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
if [ -z "$COMMIT_HASH" ]; then
  COMMIT_HASH="n/a"
fi
BUILD_TIMESTAMP=$(date '+%Y-%m-%dT%H:%M:%S')

echo "Building all docker images with build args:"
echo "  BUILD=$BUILD"
echo "  VERSION=$VERSION"
echo "  COMMIT_HASH=$COMMIT_HASH"
echo "  BUILD_TIMESTAMP=$BUILD_TIMESTAMP"

docker compose -f deploy.yaml build \
  --build-arg BUILD="$BUILD" \
  --build-arg VERSION="$VERSION" \
  --build-arg COMMIT_HASH="$COMMIT_HASH" \
  --build-arg BUILD_TIMESTAMP="$BUILD_TIMESTAMP"

echo "Starting database and minio..."
docker compose -f deploy.yaml up -d db minio

echo "Starting cache-server..."
docker compose -f deploy.yaml up -d cache-server

# Check if cache 'test' already exists in DB
if ! docker compose -f deploy.yaml exec -T cache-server ./cache-server -m -c cache-server.conf cache info test > /dev/null 2>&1; then
  echo "Initializing cache 'test' and related workspace/agent..."
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf cache create test 23455 -r 2
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf workspace create w1 test
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf agent add a1 w1
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf agent info a1
else
  echo "Cache 'test' already exists. Skipping initialization."
fi

echo "Starting agent-node and caddy proxy..."
docker compose -f deploy.yaml up -d cache-test-node caddy

echo "Deployment complete!"
