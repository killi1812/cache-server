#!/usr/bin/env bash
# Master Performance Test Runner
set -e

# Config
APP_NAME="cache-server"
CODE_DIR="./perf"
cd "$(dirname "$0")"

echo "================================================="
echo "  GO CACHE SERVER PERFORMANCE SUITE"
echo "================================================="

# --- Resource Monitoring Setup ---
APP_PID=$(pgrep -f "$APP_NAME" | head -n 1)
if [ -z "$APP_PID" ]; then
    echo "ERROR: $APP_NAME process not found. Please start the server first."
    exit 1
fi

MONITOR_FILE=$(mktemp)
monitor_resources() {
    while true; do
        ps -p "$APP_PID" -o %cpu,rss --no-headers >> "$MONITOR_FILE" 2>/dev/null || break
        sleep 0.5
    done
}

echo "Monitoring PID: $APP_PID"
monitor_resources &
MONITOR_PID=$!

# Ensure cleanup on exit
trap 'kill $MONITOR_PID 2>/dev/null || true; rm -f $MONITOR_FILE; rm -rf test-data' EXIT

# --- Run Scenarios ---
bash "$CODE_DIR/sequential-small.sh"
echo ""
bash "$CODE_DIR/sequential-large.sh"
echo ""
bash "$CODE_DIR/concurrent-small.sh"

# --- Report ---
echo ""
echo "================================================="
echo "  RESOURCE USAGE REPORT"
echo "================================================="
STATS=$(awk '{ 
    cpu+=$1; mem+=$2; count++; 
    if($1>max_cpu) max_cpu=$1; 
    if($2>max_mem) max_mem=$2 
} END { 
    if(count>0) printf "Avg CPU: %.2f%% | Peak CPU: %.2f%%\nAvg Mem: %.2f MB | Peak Mem: %.2f MB", cpu/count, max_cpu, mem/count/1024, max_mem/1024 
}' "$MONITOR_FILE")
echo "$STATS"
echo "================================================="
