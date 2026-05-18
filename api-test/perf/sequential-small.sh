#!/usr/bin/env bash
# Sequential Small File Burst: 100x 1MB
set -e
source "$(dirname "$0")/common.sh"

echo "--- Running Sequential Small File Burst (100x 1MB) ---"
mkdir -p test-data
[ -f test-data/1000MB_base.bin ] || dd if=/dev/urandom of=test-data/1MB_base.bin bs=1M count=1 2>/dev/null

START=$(get_time)
HASHES=()
for i in {1..100}; do
    U=$(init_upload)
    echo -n "$i" > "test-data/1MB_${i}.bin"
    cat test-data/1MB_base.bin >> "test-data/1MB_${i}.bin"
    H=$(sha256sum "test-data/1MB_${i}.bin" | awk '{print $1}')
    HASHES+=("$H")
    
    curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary "@test-data/1MB_${i}.bin" "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null
    complete_upload "$U" "$H" 1048576 "small${i}"
    rm "test-data/1MB_${i}.bin"
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
rm test-data/1MB_base.bin
