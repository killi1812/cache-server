#!/usr/bin/env bash
# Concurrent Load Test: 50 parallel 100MB uploads
set -e
source "$(dirname "$0")/common.sh"

echo "--- Running Concurrent Load Test (50x 100MB) ---"
mkdir -p test-data
[ -f test-data/100MB_base.bin ] || dd if=/dev/urandom of=test-data/100MB_base.bin bs=1M count=100 2>/dev/null

run_single_concurrent() {
    local i=$1
    local U=$(init_upload)
    local FILE="test-data/100MB_para_${i}.bin"
    echo -n "$i" > "$FILE"
    cat test-data/100MB_base.bin >> "$FILE"
    local H=$(sha256sum "$FILE" | awk '{print $1}')
    
    curl -s -k -H "Authorization: Bearer $KEY" -X PUT --data-binary "@$FILE" "$PROTOCOL://${CACHE_DOMAIN}/${U}" > /dev/null
    complete_upload "$U" "$H" 52428800 "para${i}"
    
    # Store hash for download check (tricky in parallel, we'll just download one later or all)
    echo "$H" >> test-data/concurrent_hashes.txt
    rm "$FILE"
}

rm -f test-data/concurrent_hashes.txt
START=$(get_time)
for i in {1..50}; do
    run_single_concurrent "$i" &
done
wait
END=$(get_time)
echo "Concurrent Upload/Complete finished in $(calc_duration $START $END)s"

# Concurrent Download
HASHES=($(cat test-data/concurrent_hashes.txt))
START=$(get_time)
for H in "${HASHES[@]}"; do
    curl -s -k -H "Authorization: Bearer $KEY" "$PROTOCOL://${CACHE_DOMAIN}/nar/${H}.nar.xz" -o /dev/null &
done
wait
END=$(get_time)
echo "Concurrent Download finished in $(calc_duration $START $END)s"

rm test-data/100MB_base.bin test-data/concurrent_hashes.txt
