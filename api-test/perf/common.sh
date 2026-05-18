#!/usr/bin/env bash
# Common configuration and helpers for performance tests

HOST="localhost"
CACHE="${CACHE:-test}"
PROTOCOL="https"
KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.QlmOBM7imQkVauXII7Hd9rYAFgW6NKMuvZ4GmVSTgpM"

CACHE_DOMAIN="${CACHE}.${HOST}"
MGMT_DOMAIN="${HOST}"

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
    local SUFFIX=$4
    
    local JSON_BODY="{
        \"narInfoCreate\": {
            \"cFileHash\": \"$HASH\",
            \"cFileSize\": $SIZE,
            \"cStoreHash\": \"perf${SUFFIX}\",
            \"cStoreSuffix\": \"file${SUFFIX}\",
            \"cNarHash\": \"$HASH\",
            \"cNarSize\": $SIZE,
            \"cReferences\": [],
            \"cDeriver\": \"perf${SUFFIX}.drv\",
            \"cSig\": \"perfsig\"
        }
    }"

    curl -s -k -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
         -X POST "$PROTOCOL://$MGMT_DOMAIN/api/v1/cache/${CACHE}/multipart-nar/${UUID}/complete" \
         -d "$JSON_BODY" > /dev/null
}

# Helper for timing
get_time() {
    date +%s.%N
}

calc_duration() {
    awk "BEGIN {print $2 - $1}"
}
