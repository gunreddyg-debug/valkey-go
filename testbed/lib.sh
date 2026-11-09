# Shared helpers for the local-failover scripts. Source it; do not run it.
set -euo pipefail

LF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOST=127.0.0.1
PORTS=(7101 7102 7103 7104 7105 7106)
LB_PORT=16379
COMPOSE=(docker compose -f "$LF_DIR/docker-compose.yml")

# redis-cli on the host by default; valkey-cli on the host, or the one in a node container, as fallbacks.
# CLI overrides it, e.g. CLI="docker exec lf-valkey-7102 valkey-cli" (no -i: the CLI never reads stdin).
if [[ -n "${CLI:-}" ]]; then
  read -r -a CLI_CMD <<<"$CLI"
elif command -v redis-cli >/dev/null 2>&1; then
  CLI_CMD=(redis-cli)
elif command -v valkey-cli >/dev/null 2>&1; then
  CLI_CMD=(valkey-cli)
else
  CLI_CMD=(docker exec lf-valkey-7102 valkey-cli)
fi

log() { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*" >&2; }
die() { log "ERROR: $*"; exit 1; }

# cli PORT ARGS... runs one command against a node, giving up after 2s so a paused node cannot hang the caller.
cli() {
  local port=$1
  shift
  timeout 2 "${CLI_CMD[@]}" -h "$HOST" -p "$port" "$@" </dev/null
}

container_of() { echo "lf-valkey-$1"; }

is_paused() { [[ "$(docker inspect -f '{{.State.Paused}}' "$1" 2>/dev/null)" == "true" ]]; }

unpause_all() {
  local port c
  for port in "${PORTS[@]}"; do
    c=$(container_of "$port")
    if is_paused "$c"; then
      log "unpausing $c"
      docker unpause "$c" >/dev/null
    fi
  done
}

# cluster_nodes prints CLUSTER NODES from the first node that answers.
cluster_nodes() {
  local port out
  for port in "${PORTS[@]}"; do
    if out=$(cli "$port" CLUSTER NODES 2>/dev/null) && [[ -n "$out" && "$out" != ERR* ]]; then
      printf '%s\n' "$out"
      return 0
    fi
  done
  return 1
}

# slot_owner_port SLOT prints the port of the primary that owns SLOT (ignoring primaries flagged fail).
slot_owner_port() {
  local slot=$1 port
  port=$(cluster_nodes | awk -v slot="$slot" '
    $3 ~ /master/ && $3 !~ /(^|,)fail(,|$)/ {
      for (i = 9; i <= NF; i++) {
        if ($i !~ /^[0-9]+(-[0-9]+)?$/) continue
        n = split($i, r, "-"); lo = r[1] + 0; hi = (n == 2 ? r[2] : r[1]) + 0
        if (slot + 0 >= lo && slot + 0 <= hi) { split($2, a, "@"); m = split(a[1], b, ":"); print b[m]; exit }
      }
    }')
  [[ -n "$port" ]] || return 1
  echo "$port"
}

# replica_of_port PRIMARY_PORT prints the port of a replica of that primary.
replica_of_port() {
  local primary=$1 port
  port=$(cluster_nodes | awk -v p="$primary" '
    { split($2, a, "@"); m = split(a[1], b, ":"); port[$1] = b[m]; flags[$1] = $3; master[$1] = $4 }
    END {
      for (id in port) if (port[id] == p) pid = id
      for (id in port) if (master[id] == pid && flags[id] ~ /slave|replica/ && flags[id] !~ /fail/) { print port[id]; exit }
    }')
  [[ -n "$port" ]] || return 1
  echo "$port"
}

info_field() { # info_field PORT SECTION FIELD
  cli "$1" INFO "$2" 2>/dev/null | tr -d '\r' | awk -F: -v f="$3" '$1 == f { print $2 }'
}

cluster_info_field() { # cluster_info_field PORT FIELD
  cli "$1" CLUSTER INFO 2>/dev/null | tr -d '\r' | awk -F: -v f="$2" '$1 == f { print $2 }'
}

# wait_until TIMEOUT_SECONDS DESCRIPTION COMMAND... retries COMMAND every 0.5s.
wait_until() {
  local timeout=$1 what=$2
  shift 2
  local deadline=$((SECONDS + timeout))
  until "$@"; do
    ((SECONDS < deadline)) || die "timed out after ${timeout}s waiting for: $what"
    sleep 0.5
  done
}

all_ping() {
  local port
  for port in "${PORTS[@]}"; do
    [[ "$(cli "$port" PING 2>/dev/null)" == "PONG" ]] || return 1
  done
}

# cluster_healthy: every node reports cluster_state:ok, knows all 6 nodes and all 16384 slots, there are
# 3 primaries and 3 replicas, and every replica's replication link is up.
cluster_healthy() {
  local port role primaries=0 replicas=0
  for port in "${PORTS[@]}"; do
    [[ "$(cluster_info_field "$port" cluster_state)" == "ok" ]] || return 1
    [[ "$(cluster_info_field "$port" cluster_known_nodes)" == "6" ]] || return 1
    [[ "$(cluster_info_field "$port" cluster_slots_ok)" == "16384" ]] || return 1
    role=$(info_field "$port" replication role)
    case "$role" in
      master) primaries=$((primaries + 1)) ;;
      slave | replica)
        replicas=$((replicas + 1))
        [[ "$(info_field "$port" replication master_link_status)" == "up" ]] || return 1
        ;;
      *) return 1 ;;
    esac
  done
  ((primaries == 3 && replicas == 3))
}

lb_ok() { [[ "$(timeout 2 "${CLI_CMD[@]}" -h "$HOST" -p "$LB_PORT" PING 2>/dev/null </dev/null)" == "PONG" ]]; }

show_topology() {
  cluster_nodes | awk '{ split($2, a, "@"); s = ""; for (i = 9; i <= NF; i++) s = s " " $i; printf "  %-16s %-22s%s\n", a[1], $3, s }' | sort
}

