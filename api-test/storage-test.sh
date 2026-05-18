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

echo "================================================="
echo "  Storage Performance Test Suite (Robust Mode)"
echo "  Target: $PROTOCOL://$CACHE_DOMAIN"
echo "================================================="

# Create test data
echo "Generating test data..."
mkdir -p test-data
dd if=/dev/urandom of=test-data/1MB.bin bs=1M count=1 2>/dev/null
dd if=/dev/urandom of=test-data/100MB.bin bs=1M count=100 2>/dev/null

# Helper to initialize a multipart upload and return the UUID
init_upload() {
    local RES=$(curl -s -k -H "Authorization: Bearer $KEY" -X POST "$PROTOCOL://$MGMT_DOMAIN/api/v1/cache/${CACHE}/multipart-nar?compression=xz")
    
    # Try to extract uploadId. If it fails, show the raw response for debugging.
    local UUID=$(echo "$RES" | python3 -c "import sys, json; 
try:
    print(json.load(sys.stdin)['uploadId'])
except Exception:
    sys.exit(1)" 2>/dev/null)

    if [ -z "$UUID" ]; then
        echo "FAILED to initialize upload. Server returned:" >&2
        echo "$RES" >&2
        exit 1
    fi
    echo "$UUID"
}

echo ""
echo "--- 1. Small File Burst (Sequential Upload 50x 1MB) ---"
START_TIME=$(date +%s.%N)
UUIDS_SMALL=()
for i in {1..50}; do
  U=$(init_upload)
  UUIDS_SMALL+=("$U")
  # PUT to /UUID (Go server strips leading slash in c.Param("filename"))
  RESPONSE=$(curl -v -k -X PUT --data-binary @test-data/1MB.bin "$PROTOCOL://${CACHE_DOMAIN}/${U}" 2>&1) || {
    echo "Request $i failed!"
    echo "$RESPONSE"
    exit 1
  }
done
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 2. Small File Burst (Sequential Download 50x 1MB) ---"
START_TIME=$(date +%s.%N)
for U in "${UUIDS_SMALL[@]}"; do
  curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${U}.nar.xz" -o /dev/null
done
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 3. Large File Throughput (Upload 1x 100MB) ---"
U_LARGE=$(init_upload)
START_TIME=$(date +%s.%N)
curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary @test-data/100MB.bin "$PROTOCOL://${CACHE_DOMAIN}/${U_LARGE}" > /dev/null
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 4. Large File Throughput (Download 1x 100MB) ---"
START_TIME=$(date +%s.%N)
curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${U_LARGE}.nar.xz" -o /dev/null
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 5. Concurrent Load (Upload 10x 100MB) ---"
UUIDS_PARA=()
START_TIME=$(date +%s.%N)
for i in {1..10}; do
  U=$(init_upload)
  UUIDS_PARA+=("$U")
  curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary @test-data/100MB.bin "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null &
done
wait
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 6. Concurrent Load (Download 10x 100MB) ---"
START_TIME=$(date +%s.%N)
for U in "${UUIDS_PARA[@]}"; do
  curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${U}.nar.xz" -o /dev/null &
done
wait
END_TIME=$(date +%s.%N)
DURATION=$(awk "BEGIN {print $END_TIME - $START_TIME}")
echo "Completed in $DURATION seconds"

echo ""
echo "--- 7. Data Integrity Validation ---"
ORIG_HASH=$(sha256sum test-data/100MB.bin | awk '{print $1}')
curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${U_LARGE}.nar.xz" -o test-data/100MB-downloaded.bin
DOWN_HASH=$(sha256sum test-data/100MB-downloaded.bin | awk '{print $1}')

if [ "$ORIG_HASH" == "$DOWN_HASH" ]; then
    echo "Integrity Check: PASSED (Hashes match)"
else
    echo "Integrity Check: FAILED (Expected $ORIG_HASH, got $DOWN_HASH)"
fi

echo ""
echo "Cleaning up..."
rm -rf test-data
echo "Storage tests complete."
