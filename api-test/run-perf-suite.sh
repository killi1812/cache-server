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
MONITOR_FILE=$(mktemp)

monitor_resources() {
    export LC_NUMERIC=C
    while true; do
        # Find all cache-server PIDs, excluding this script ($$)
        PIDS=$(pgrep -f "$APP_NAME" | grep -v "$$" | paste -sd "," -)
        
        if [ ! -z "$PIDS" ]; then
            # Sum CPU and RSS for all PIDs in this sample
            ps -p "$PIDS" -o %cpu,rss --no-headers 2>/dev/null | \
            tr ',' '.' | \
            awk '{cpu+=$1; mem+=$2} END {if(NR>0) print cpu, mem}' >> "$MONITOR_FILE"
        fi
        sleep 0.5
    done
}

# Verify at least one process exists
if ! pgrep -f "$APP_NAME" | grep -v "$$" > /dev/null; then
    echo "ERROR: cache-server failed to start. Check server.log"
    exit 1
fi

echo "Monitoring all cache-server processes..."
monitor_resources &
MONITOR_PID=$!

# Ensure cleanup on exit
trap 'kill $MONITOR_PID 2>/dev/null || true; killall cache-server 2>/dev/null || true; rm -f $MONITOR_FILE; rm -rf test-data' EXIT

# --- Run Scenarios ---
bash "$CODE_DIR/sequential-small.sh"
echo ""
bash "$CODE_DIR/sequential-large.sh"
echo ""
python3 "$CODE_DIR/concurrent_heavy.py"

# --- Report ---
echo ""
echo "================================================="
echo "  TOTAL RESOURCE USAGE REPORT (ALL PROCESSES)"
echo "================================================="
STATS=$(awk '{ 
    cpu+=$1; mem+=$2; count++; 
    if($1>max_cpu) max_cpu=$1; 
    if($2>max_mem) max_mem=$2 
} END { 
    if(count>0) printf "Avg Total CPU: %.2f%% | Peak Total CPU: %.2f%%\nAvg Total Mem: %.2f MB | Peak Total Mem: %.2f MB", cpu/count, max_cpu, mem/count/1024, max_mem/1024 
}' "$MONITOR_FILE")
echo "$STATS"
echo "================================================="
