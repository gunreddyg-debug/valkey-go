#!/usr/bin/env bash
# Unplanned failure of the primary that owns slot 0.
#
#   ./fail-primary.sh [SECONDS]        # default 12
#
# MODE=pause (default): `docker pause` the primary for SECONDS, then unpause it. The process is frozen:
#   its TCP connections stay open and the kernel still accepts new ones, but nothing is answered, like a
#   hung node or a network partition. SECONDS (default 12) is longer than cluster-node-timeout (5s),
#   so the cluster flags the primary as failed and promotes its replica; once unpaused, the old primary
#   rejoins as a replica.
# MODE=kill: `docker kill` the primary (connections reset, new ones refused, like a crash), then
#   `docker start` it after SECONDS.
# HOLD_PROMOTION=N (default 0): stretch the window in which the shard has no online primary. Before the
#   failure, the replica gets cluster-replica-no-failover yes; once the primary is flagged fail, the
#   script waits N more seconds and then runs CLUSTER FAILOVER FORCE on the replica. SECONDS must be
#   larger than the time to flag the primary (6-8s here) plus N.
#
# Prints machine-readable lines; the bench parses the first two:
#   FAILED port=<port> mode=<mode> container=<name>
#   RESTORED port=<port>
#   FLAGGED_FAIL port=<port> / PROMOTE replica=<port>        (HOLD_PROMOTION only)
# The node (and the replica's setting) is restored on exit even if the script is interrupted.
source "$(dirname "$0")/lib.sh"

secs=${1:-12}
mode=${MODE:-pause}
hold=${HOLD_PROMOTION:-0}
[[ "$secs" =~ ^[0-9]+$ ]] || die "usage: $0 [SECONDS]"
[[ "$hold" =~ ^[0-9]+$ ]] || die "HOLD_PROMOTION must be a number of seconds"
[[ "$mode" == pause || "$mode" == kill ]] || die "MODE must be pause or kill"

port=$(slot_owner_port 0) || die "no healthy primary owns slot 0 (is the cluster up? ./up.sh)"
c=$(container_of "$port")
replica=""
if ((hold > 0)); then
  replica=$(replica_of_port "$port") || die "primary $port has no healthy replica"
fi

restored=0
nap_pid=""
restore() {
  ((restored)) && return 0
  restored=1
  [[ -z "$nap_pid" ]] || kill "$nap_pid" 2>/dev/null || true
  case "$mode" in
    pause) is_paused "$c" && docker unpause "$c" >/dev/null ;;
    kill) [[ "$(docker inspect -f '{{.State.Running}}' "$c" 2>/dev/null)" == "true" ]] || docker start "$c" >/dev/null ;;
  esac
  [[ -z "$replica" ]] || cli "$replica" CONFIG SET cluster-replica-no-failover no >/dev/null || true
  echo "RESTORED port=$port"
}
trap restore EXIT
trap 'exit 130' INT TERM

# nap SECONDS sleeps in the background and waits for it, so INT/TERM sent to this script alone (not to
# its process group, as Ctrl-C does) run the trap at once instead of after the sleep.
nap() {
  sleep "$1" &
  nap_pid=$!
  wait "$nap_pid" || true
  nap_pid=""
}

flagged_fail() { # the replica sees the primary with the fail flag (not just fail?)
  cli "$replica" CLUSTER NODES 2>/dev/null | awk -v p=":$port@" 'index($2, p) && $3 ~ /(^|,)fail(,|$)/ { found = 1 } END { exit !found }'
}

[[ -z "$replica" ]] || cli "$replica" CONFIG SET cluster-replica-no-failover yes >/dev/null
case "$mode" in
  pause) is_paused "$c" || docker pause "$c" >/dev/null ;;
  kill) docker kill --signal KILL "$c" >/dev/null ;;
esac
start_ns=$(date +%s%N)
echo "FAILED port=$port mode=$mode container=$c"
log "$mode $c (primary of slot 0) for ${secs}s${replica:+, promotion of replica $replica held ${hold}s after the fail flag}"
if [[ -z "$replica" ]]; then
  nap "$secs"
else
  elapsed_ms() { echo $((($(date +%s%N) - start_ns) / 1000000)); }
  until flagged_fail; do
    (($(elapsed_ms) < secs * 1000)) || die "primary $port not flagged fail within ${secs}s"
    sleep 0.05
  done
  echo "FLAGGED_FAIL port=$port"
  nap "$hold"
  reply=$(cli "$replica" CLUSTER FAILOVER FORCE)
  echo "PROMOTE replica=$replica reply=$reply"
  left=$((secs * 1000 - $(elapsed_ms)))
  ((left <= 0)) || nap "$(printf '%d.%03d' $((left / 1000)) $((left % 1000)))"
fi
restore
trap - EXIT
log "restored $c; slot 0 is now owned by port $(slot_owner_port 0 || echo '?')"
