#!/usr/bin/env bash
set -e

# Configuration
HOST="localhost"
CACHE="${CACHE:-test}"
PROTOCOL="https"
KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.QlmOBM7imQkVauXII7Hd9rYAFgW6NKMuvZ4GmVSTgpM"

CACHE_DOMAIN="${CACHE}.${HOST}"
# Management API uses a different domain (the root host)
MGMT_DOMAIN="${HOST}"

# --- Resource Monitoring Setup ---
APP_PID=$(pgrep -f "cache-server" | head -n 1)
MONITOR_FILE=$(mktemp)
monitor_resources() {
    while true; do
        if [ ! -z "$APP_PID" ]; then
            ps -p "$APP_PID" -o %cpu,rss --no-headers >> "$MONITOR_FILE" 2>/dev/null || break
        fi
        sleep 0.5
    done
}

if [ ! -z "$APP_PID" ]; then
    echo "Monitoring PID: $APP_PID"
    monitor_resources &
    MONITOR_PID=$!
fi

echo "================================================="
echo "  Storage Performance Test Suite (Unique Files)"
echo "  Target: $PROTOCOL://$CACHE_DOMAIN"
echo "================================================="

# Create base test data
mkdir -p test-data
[ -f test-data/base.bin ] || dd if=/dev/urandom of=test-data/base.bin bs=1M count=100 2>/dev/null

# Helper to initialize a multipart upload and return the UUID
init_upload() {
    local RES=$(curl -s -k -H "Authorization: Bearer $KEY" -X POST "$PROTOCOL://$MGMT_DOMAIN/api/v1/cache/${CACHE}/multipart-nar?compression=xz")
    echo "$RES" | python3 -c "import sys, json; print(json.load(sys.stdin)['uploadId'])"
}

# Helper to complete an upload
complete_upload() {
    local UUID=$1
    local HASH=$2
    local SIZE=$3
    local INDEX=$4
    
    local JSON_BODY="{
        \"narInfoCreate\": {
            \"cFileHash\": \"$HASH\",
            \"cFileSize\": $SIZE,
            \"cStoreHash\": \"perf${INDEX}\",
            \"cStoreSuffix\": \"file${INDEX}\",
            \"cNarHash\": \"$HASH\",
            \"cNarSize\": $SIZE,
            \"cReferences\": [],
            \"cDeriver\": \"perf${INDEX}.drv\",
            \"cSig\": \"perfsig\"
        }
    }"

    curl -s -k -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
         -X POST "$PROTOCOL://$MGMT_DOMAIN/api/v1/cache/${CACHE}/multipart-nar/${UUID}/complete" \
         -d "$JSON_BODY" > /dev/null
}

echo ""
echo "--- 1. Unique Upload & Complete (20x 100MB) ---"
# We reduce count to 20 for faster cycle, but each is UNIQUE
START_TIME=$(date +%s.%N)
HASHES=()
for i in {1..20}; do
  U=$(init_upload)
  # Modify 1 byte to ensure unique hash
  echo -n "$i" > "test-data/unique-${i}.bin"
  cat test-data/base.bin >> "test-data/unique-${i}.bin"
  
  H=$(sha256sum "test-data/unique-${i}.bin" | awk '{print $1}')
  HASHES+=("$H")
  
  curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary "@test-data/unique-${i}.bin" "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null
  complete_upload "$U" "$H" 104857600 "$i"
  rm "test-data/unique-${i}.bin"
done
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed 20 unique 100MB uploads in $DURATION seconds"

echo ""
echo "--- 2. Unique Download by Hash (20x 100MB) ---"
START_TIME=$(date +%s.%N)
for H in "${HASHES[@]}"; do
  curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${H}.nar.xz" -o /dev/null
done
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed 20 unique downloads in $DURATION seconds"

echo ""
echo "Cleaning up..."
rm -rf test-data

if [ ! -z "$MONITOR_PID" ]; then
    kill "$MONITOR_PID" 2>/dev/null
    echo "--- Resource Usage Report ---"
    STATS=$(awk '{ 
        cpu+=$1; mem+=$2; count++; 
        if($1>max_cpu) max_cpu=$1; 
        if($2>max_mem) max_mem=$2 
    } END { 
        if(count>0) printf "Avg CPU: %.2f%% | Peak CPU: %.2f%%\nAvg Mem: %.2f MB | Peak Mem: %.2f MB", cpu/count, max_cpu, mem/count/1024, max_mem/1024 
    }' "$MONITOR_FILE")
    echo "$STATS"
    rm "$MONITOR_FILE"
fi

echo ""
echo "Storage tests complete."
