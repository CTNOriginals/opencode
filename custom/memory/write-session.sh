#!/bin/bash
set -euo pipefail

BINARY="/home/ctn/.config/opencode/custom/memory/build/memory-server"
LOG_FILE="/home/ctn/.config/opencode/custom/memory/memory-hook.log"
NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

log() {
    echo "[$NOW] $1" >> "$LOG_FILE"
}

if [ ! -x "$BINARY" ]; then
    log "binary not found at $BINARY, attempting build"
    cd /home/ctn/.config/opencode/custom/memory
    go build -o build/memory-server . 2>>"$LOG_FILE" || {
        log "build failed"
        exit 0
    }
fi

INPUT=$(cat)

if [ -z "$INPUT" ]; then
    log "empty input, nothing to process"
    exit 0
fi

echo "$INPUT" | "$BINARY" process-session 2>>"$LOG_FILE" || {
    log "process-session failed"
    exit 0
}

log "session processed through pipeline"
