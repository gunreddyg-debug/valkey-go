#!/usr/bin/env bash
# Planned failover: CLUSTER FAILOVER on the replica of the primary that owns SLOT (default 0), then
# wait until the replica owns the slot and the cluster is healthy again. This is the analog of a
# Memorystore maintenance event.
#
#   ./planned-failover.sh [SLOT]
#
# Idempotent in the sense that every call leaves a healthy cluster: calling it again fails the shard
# back to the previous primary.
#
# Prints machine-readable lines; the bench (-fail-kind planned) parses them:
#   FAILED port=<primary> mode=planned replica=<replica>   right before CLUSTER FAILOVER is sent
#   PROMOTE replica=<replica> reply=OK                       CLUSTER FAILOVER answered
#   PROMOTED port=<replica>                                  the replica owns the slot
#   FAILED_OVER slot=<slot> old=<primary> new=<replica>      the cluster is healthy again
#   RESTORED port=<primary>                                  (same moment; the end of the event)
source "$(dirname "$0")/lib.sh"

slot=${1:-0}
primary=$(slot_owner_port "$slot") || die "no healthy primary owns slot $slot (is the cluster up? ./up.sh)"
replica=$(replica_of_port "$primary") || die "primary $primary has no healthy replica"

log "CLUSTER FAILOVER on replica $replica (primary $primary owns slot $slot)"
echo "FAILED port=$primary mode=planned replica=$replica"
reply=$(cli "$replica" CLUSTER FAILOVER)
[[ "$reply" == "OK" ]] || die "CLUSTER FAILOVER on $replica: $reply"
echo "PROMOTE replica=$replica reply=$reply"

promoted() { [[ "$(slot_owner_port "$slot" 2>/dev/null)" == "$replica" ]]; }
wait_until 30 "port $replica to own slot $slot" promoted
echo "PROMOTED port=$replica"
wait_until 60 "cluster healthy" cluster_healthy
echo "FAILED_OVER slot=$slot old=$primary new=$replica"
echo "RESTORED port=$primary"
show_topology >&2
