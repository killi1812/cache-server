#!/bin/bash
set -e

# 1. Build local binary
echo ">>> Building local binary..."
task build

# 2. Start stack
echo ">>> Starting Lab Stack..."
docker compose -f deploy.yaml up -d --build

# Wait for healthy DB and Mgmt Server
echo ">>> Waiting for services to be ready..."
sleep 5

# 3. Create Cache & Workspace
echo ">>> Provisioning Lab Cache & Workspace..."
# Get a token for management
MGMT_TOKEN=$(go run src/main.go util auth generate test-admin | grep -oE '[a-zA-Z0-9\._\-]+$')

curl -s -X POST http://localhost:12345/api/v1/cache \
  -H "Authorization: Bearer $MGMT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "lab-cache", "port": 10001}'

curl -s -X POST http://localhost:12345/api/v1/deploy/workspace \
  -H "Authorization: Bearer $MGMT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "lab-ws", "cacheName": "lab-cache"}'

# 4. Build something with Nix
echo ">>> Building test.nix..."
STORE_PATH=$(nix-build --no-out-link api-test/test.nix)
echo ">>> Built: $STORE_PATH"

# 5. Configure Cachix to use local server
echo ">>> Pointing Cachix to localhost..."
cachix config set hostname http://localhost:12345

# 6. Push to local binary cache
# We use the token from the cache we just created
CACHE_TOKEN="secret" # Default from config if not changed
export CACHIX_AUTH_TOKEN=$CACHE_TOKEN
echo ">>> Pushing $STORE_PATH to lab-cache..."
echo $STORE_PATH | cachix push lab-cache

# 7. Register and Start Agent
echo ">>> Registering Agent..."
AGENT_JSON=$(curl -s -X POST http://localhost:12345/api/v1/deploy/agent/lab-ws/agent-1 -H "Authorization: Bearer $MGMT_TOKEN")
AGENT_TOKEN=$(echo $AGENT_JSON | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

echo ">>> Starting Agent Node Listener..."
docker compose -f deploy.yaml exec -d agent-node ./cache-server agent listen agent-1 http://mgmt-server:12345 --token "$AGENT_TOKEN"

sleep 2

# 8. Trigger Deployment via Cachix
echo ">>> Triggering Deployment via Cachix..."
SPEC_FILE="deploy-lab.json"
echo "{\"agents\": {\"agent-1\": \"$STORE_PATH\"}}" > $SPEC_FILE

export CACHIX_ACTIVATE_TOKEN=$MGMT_TOKEN
cachix deploy activate $SPEC_FILE

rm $SPEC_FILE
echo ""
echo ">>> Lab Test with Cachix complete."
echo ">>> Check logs: docker compose -f deploy.yaml logs -f mgmt-server agent-node"
