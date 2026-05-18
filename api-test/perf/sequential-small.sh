#!/usr/bin/env bash
# Sequential Burst: 50x 100MB
set -e
source "$(dirname "$0")/common.sh"

echo "--- Running Sequential Burst (20x 100MB) ---"
mkdir -p test-data
[ -f test-data/100MB_base.bin ] || dd if=/dev/urandom of=test-data/100MB_base.bin bs=1M count=100 2>/dev/null

START=$(get_time)
HASHES=()
for i in {1..20}; do
    U=$(init_upload)
    # Use timestamp + index to ensure universal uniqueness
    echo -n "$(date +%s%N)_burst_${i}" > "test-data/100MB_${i}.bin"
    cat test-data/100MB_base.bin >> "test-data/100MB_${i}.bin"
    H=$(sha256sum "test-data/100MB_${i}.bin" | awk '{print $1}')
    HASHES+=("$H")
    
    curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary "@test-data/100MB_${i}.bin" "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null
    complete_upload "$U" "$H" 104857600 "burst${i}"
    rm "test-data/100MB_${i}.bin"
done
END=$(get_time)
echo "Upload/Complete finished in $(calc_duration $START $END)s"

# Download test
START=$(get_time)
for H in "${HASHES[@]}"; do
    curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${H}.nar.xz" -o /dev/null
done
END=$(get_time)
echo "Download finished in $(calc_duration $START $END)s"
