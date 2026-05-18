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
# Find the PID (handles both 'cache-server' and '.cache-server-wrapped')
APP_PID=$(pgrep -f "cache-server" | head -n 1)
if [ -z "$APP_PID" ]; then
    echo "WARNING: Could not find running cache-server process. Monitoring disabled."
fi

MONITOR_FILE=$(mktemp)
monitor_resources() {
    while true; do
        if [ ! -z "$APP_PID" ]; then
            # Get CPU (%) and RSS (KB)
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
echo "  Storage Performance Test Suite (Full Lifecycle)"
echo "  Target: $PROTOCOL://$CACHE_DOMAIN"
echo "================================================="

# Create test data
echo "Generating test data..."
mkdir -p test-data
[ -f test-data/100MB.bin ] || dd if=/dev/urandom of=test-data/100MB.bin bs=1M count=100 2>/dev/null
[ -f test-data/1000MB.bin ] || dd if=/dev/urandom of=test-data/1000MB.bin bs=1M count=1000 2>/dev/null

# Pre-calculate hashes
HASH_100MB=$(sha256sum test-data/100MB.bin | awk '{print $1}')
HASH_1000MB=$(sha256sum test-data/1000MB.bin | awk '{print $1}')

# Helper to initialize a multipart upload and return the UUID
init_upload() {
    local RES=$(curl -s -k -H "Authorization: Bearer $KEY" -X POST "$PROTOCOL://$MGMT_DOMAIN/api/v1/cache/${CACHE}/multipart-nar?compression=xz")
    local UUID=$(echo "$RES" | python3 -c "import sys, json; 
try:
    print(json.load(sys.stdin)['uploadId'])
except Exception:
    sys.exit(1)" 2>/dev/null)

    if [ -z "$UUID" ]; then
        echo "FAILED to initialize upload. Server returned: $RES" >&2
        exit 1
    fi
    echo "$UUID"
}

# Helper to complete an upload (renames UUID to HASH)
complete_upload() {
    local UUID=$1
    local HASH=$2
    local SIZE=$3
    
    local JSON_BODY="{
        \"narInfoCreate\": {
            \"cFileHash\": \"$HASH\",
            \"cFileSize\": $SIZE,
            \"cStoreHash\": \"perfhash\",
            \"cStoreSuffix\": \"perfsuffix\",
            \"cNarHash\": \"sha256:$HASH\",
            \"cNarSize\": $SIZE,
            \"cReferences\": [],
            \"cDeriver\": \"perf.drv\",
            \"cSig\": \"perfsig\"
        }
    }"

    curl -s -k -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
         -X POST "$PROTOCOL://$MGMT_DOMAIN/api/v1/cache/${CACHE}/multipart-nar/${UUID}/complete" \
         -d "$JSON_BODY" > /dev/null
}

echo ""
echo "--- 1. Sequential Upload & Complete (50x 100MB) ---"
START_TIME=$(date +%s.%N)
for i in {1..50}; do
  U=$(init_upload)
  curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary @test-data/100MB.bin "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null
  complete_upload "$U" "$HASH_100MB" 104857600
done
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 2. Sequential Download by Hash (50x 100MB) ---"
START_TIME=$(date +%s.%N)
for i in {1..50}; do
  curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${HASH_100MB}.nar.xz" -o /dev/null
done
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 3. Large File Lifecycle (Upload 1x 1000MB) ---"
START_TIME=$(date +%s.%N)
U_LARGE=$(init_upload)
curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary @test-data/1000MB.bin "$PROTOCOL://${CACHE_DOMAIN}/${U_LARGE}" > /dev/null
complete_upload "$U_LARGE" "$HASH_1000MB" 1048576000
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 4. Large File Download (Download 1x 1000MB) ---"
START_TIME=$(date +%s.%N)
curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${HASH_1000MB}.nar.xz" -o /dev/null
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 7. Data Integrity Validation ---"
curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${HASH_1000MB}.nar.xz" -o test-data/1000MB-downloaded.bin
DOWN_HASH=$(sha256sum test-data/1000MB-downloaded.bin | awk '{print $1}')

if [ "$HASH_1000MB" == "$DOWN_HASH" ]; then
    echo "Integrity Check: PASSED (Hashes match)"
else
    echo "Integrity Check: FAILED (Expected $HASH_1000MB, got $DOWN_HASH)"
fi

echo ""
echo "--- 8. Sequential Upload, Complete and Download (50x 100MB) ---"
START_TIME=$(date +%s.%N)
for i in {1..50}; do
  U=$(init_upload)
  curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary @test-data/100MB.bin "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null
  complete_upload "$U" "$HASH_100MB" 104857600
  curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${HASH_100MB}.nar.xz" -o /dev/null
done
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "Cleaning up..."
rm -rf test-data

if [ ! -z "$MONITOR_PID" ]; then
    kill "$MONITOR_PID" 2>/dev/null
    echo "--- Resource Usage Report ---"
    # CPU is first col, RSS is second
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
