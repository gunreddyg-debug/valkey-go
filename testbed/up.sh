#!/usr/bin/env bash
# Starts the 6 Valkey nodes and the discovery endpoint (HAProxy on 127.0.0.1:16379), creates the
# cluster (3 primaries + 3 replicas) if it does not exist yet, and waits until it is healthy:
# cluster_state:ok on every node, all replicas in sync, and the discovery endpoint answering PING.
#
# Idempotent: on a running, healthy cluster it only checks health. It also unpauses or restarts a node
# left paused or stopped by an interrupted fail-primary.sh.
source "$(dirname "$0")/lib.sh"

unpause_all
log "starting containers"
"${COMPOSE[@]}" up -d --quiet-pull 2>&1 | grep -v -E '^\s*$' >&2 || true

wait_until 30 "all 6 nodes answering PING" all_ping

known=()
for port in "${PORTS[@]}"; do
  known+=("$(cluster_info_field "$port" cluster_known_nodes)")
done
fresh=1
for k in "${known[@]}"; do
  [[ "$k" == "1" ]] || fresh=0
done

if ((fresh)); then
  log "creating the cluster"
  addrs=()
  for port in "${PORTS[@]}"; do addrs+=("$HOST:$port"); done
  "${CLI_CMD[@]}" --cluster create "${addrs[@]}" --cluster-replicas 1 --cluster-yes </dev/null >&2
else
  log "cluster already created (known nodes per node: ${known[*]})"
fi

log "waiting for cluster_state:ok and replica sync"
deadline=$((SECONDS + 90))
until cluster_healthy; do
  if ((SECONDS >= deadline)); then
    show_topology >&2 || true
    die "cluster not healthy after 90s; reset it with ./down.sh && ./up.sh"
  fi
  sleep 0.5
done

wait_until 30 "discovery endpoint $HOST:$LB_PORT answering PING" lb_ok

log "cluster healthy; discovery endpoint $HOST:$LB_PORT"
show_topology >&2
