#!/usr/bin/env bash
# Master Performance Test Runner (Python Modular Mode)
set -e

# Config
APP_NAME="cache-server"
CODE_DIR="./perf"
cd "$(dirname "$0")"

echo "================================================="
echo "  GO CACHE SERVER PERFORMANCE SUITE (PYTHON)"
echo "================================================="

# --- Resource Monitoring Setup ---
PIDS=$(pgrep -f "$APP_NAME" | grep -v "$$" | paste -sd "," -)
if [ -z "$PIDS" ]; then
    echo "ERROR: cache-server process not found. Please start the server first."
    exit 1
fi

MONITOR_FILE=$(mktemp)
monitor_resources() {
    export LC_NUMERIC=C
    while true; do
        CURRENT_PIDS=$(pgrep -f "$APP_NAME" | grep -v "$$" | paste -sd "," -)
        if [ ! -z "$CURRENT_PIDS" ]; then
            ps -p "$CURRENT_PIDS" -o %cpu,rss --no-headers 2>/dev/null | \
            tr ',' '.' | \
            awk '{cpu+=$1; mem+=$2} END {if(NR>0) print cpu, mem}' >> "$MONITOR_FILE"
        fi
        sleep 0.5
    done
}

echo "Monitoring PIDs: $PIDS"
monitor_resources &
MONITOR_PID=$!

# Ensure cleanup on exit
trap 'kill $MONITOR_PID 2>/dev/null || true; rm -f $MONITOR_FILE' EXIT

echo ""
export PYTHONPATH=$PYTHONPATH:.

python3 "$CODE_DIR/sequential_small.py"
echo ""
python3 "$CODE_DIR/sequential_large.py"
echo ""
python3 "$CODE_DIR/concurrent_heavy.py"

GLOBAL_END=$(date +%s.%N)

# --- Report ---
echo ""
echo "================================================="
echo "  RESOURCE USAGE REPORT (Aggregated across all tests)"
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
echo "Note: Measurements focused on UPLOAD throughput."
