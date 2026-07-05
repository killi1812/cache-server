#!/bin/sh
set -e

echo "Starting database and minio..."
docker compose -f deploy.yaml up -d db minio

echo "Starting cache-server..."
docker compose -f deploy.yaml up -d cache-server

echo "Waiting for cache-server to be ready..."
until docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf cache list > /dev/null 2>&1; do
  sleep 1
done

# Check if cache 'test' already exists in DB
if ! docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf cache info test > /dev/null 2>&1; then
  echo "Initializing cache 'test' and related workspace/agent..."
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf cache create test 23455 -r 2
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf workspace create w1 test
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf agent add a1 w1
  docker compose -f deploy.yaml exec -T cache-server ./cache-server -c cache-server.conf agent info a1
else
  echo "Cache 'test' already exists. Skipping initialization."
fi

echo "Starting agent-node..."
docker compose -f deploy.yaml up -d cache-test-node

echo "Deployment complete!"
