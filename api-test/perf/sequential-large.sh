#!/usr/bin/env bash
# Sequential Large: 1x 1000MB
set -e
source "$(dirname "$0")/common.sh"

echo "--- Running Sequential Large (1x 1000MB) ---"
mkdir -p test-data
[ -f test-data/1000MB_base.bin ] || dd if=/dev/urandom of=test-data/1000MB_base.bin bs=1M count=1000 2>/dev/null

START=$(get_time)
U=$(init_upload)
# Use timestamp to ensure universal uniqueness
echo -n "$(date +%s%N)_large_single" > "test-data/1000MB_single.bin"
cat test-data/1000MB_base.bin >> "test-data/1000MB_single.bin"
H=$(sha256sum "test-data/1000MB_single.bin" | awk '{print $1}')

curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary "@test-data/1000MB_single.bin" "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null
complete_upload "$U" "$H" 1048576000 "large_single"
END=$(get_time)
echo "Upload/Complete finished in $(calc_duration $START $END)s"

# Download test
START=$(get_time)
curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${H}.nar.xz" -o /dev/null
END=$(get_time)
echo "Download finished in $(calc_duration $START $END)s"

rm "test-data/1000MB_single.bin"
