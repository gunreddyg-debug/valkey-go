#!/usr/bin/env bash
# Stops and removes the test bed: the 6 Valkey containers, the discovery endpoint, and the data volumes
# (so the next ./up.sh creates a fresh cluster). Idempotent: does nothing if nothing is running.
source "$(dirname "$0")/lib.sh"

unpause_all 2>/dev/null || true
log "removing containers and volumes"
"${COMPOSE[@]}" down --volumes --remove-orphans --timeout 5 2>&1 | grep -v -E '^\s*$' >&2 || true
log "down"

