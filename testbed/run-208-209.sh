#!/usr/bin/env bash
# Runs the PR #208/#209 x PR #20 matrix from a queue file: one bench run per line, each on a fresh
# cluster (./down.sh + ./up.sh), and leaves the cluster down at the end. Meant to be started once and
# left alone; it never stops on a failed run.
#
#   ./make-queue-208-209.sh >runs208/queue.txt   # the matrix
#   ./run-208-209.sh                             # OUT=runs208, QUEUE=$OUT/queue.txt
#
# Queue lines: <scenario> <label> <round> <timeout_s> <bench command and flags...>
# '#' comments and blank lines are skipped. The queue is read again before every run, so lines can be
# added while it runs (write a new file and rename it over the old one). The runner stops after the
# last line, or before the next run when the file $OUT/STOP exists.
#
# Output, in $OUT:
#   progress.log            one line per step, with the key numbers of every run
#   results.jsonl           one JSON line per run: scenario, label, round, status, exit code, wall time
#   <scenario>/<label>-<round>.json / .log / .haproxy.log / .valkey.log / .up.log
#   meta.txt                versions, image digests, SHAs, binaries
#   summary.md              bench/bin/report $OUT, at the end
# Do not edit this script while it runs: bash reads it as it goes. The queue file is safe to replace.
set -u
cd "$(dirname "$0")"
source ./lib.sh
set +e +o pipefail
OUT=${OUT:-runs208}
QUEUE=${QUEUE:-$OUT/queue.txt}
UP_TIMEOUT=${UP_TIMEOUT:-180}
mkdir -p "$OUT"
[[ -f "$QUEUE" ]] || { echo "no queue file $QUEUE (./make-queue-208-209.sh >$QUEUE)" >&2; exit 1; }
PROGRESS=$OUT/progress.log
RESULTS=$OUT/results.jsonl

plog() { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*" | tee -a "$PROGRESS"; }

child=0
stopping=0
on_signal() {
  stopping=1
  plog "runner: signal received, stopping"
  ((child > 0)) && kill -TERM "$child" 2>/dev/null
}
cleanup() {
  plog "runner: taking the cluster down"
  timeout 90 ./down.sh >>"$OUT/down.log" 2>&1
  if [[ -x bench/bin/report ]]; then
    bench/bin/report "$OUT" >"$OUT/summary.md" 2>>"$OUT/report.err" && plog "runner: tables in $OUT/summary.md"
  fi
  local ok bad
  ok=$(grep -c '"status":"ok"' "$RESULTS" 2>/dev/null)
  bad=$(grep -vc '"status":"ok"' "$RESULTS" 2>/dev/null)
  plog "runner: finished (${ok:-0} runs ok, ${bad:-0} not ok)"
}
trap on_signal INT TERM
trap cleanup EXIT

# bg_run LOG TIMEOUT CMD...: runs CMD with a timeout, in the background so that signals reach the
# runner at once; sets rc.
bg_run() {
  local logf=$1 tmo=$2
  shift 2
  timeout -k 15 "$tmo" "$@" >"$logf" 2>&1 &
  child=$!
  wait "$child"
  rc=$?
  while kill -0 "$child" 2>/dev/null; do # wait was interrupted by a signal
    wait "$child"
    rc=$?
  done
  child=0
}

json_str() { printf '"%s"' "$(printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g')"; }

write_meta() {
  local image
  image=$(awk '$1 == "image:" && $2 ~ /^valkey/ { print $2; exit }' docker-compose.yml)
  {
    echo "date:            $(date -Is)"
    echo "host:            $(uname -sr), $(nproc) CPUs, $(docker --version)"
    echo "valkey image:    $image $(docker image inspect "$image" --format '{{index .RepoDigests 0}}' 2>/dev/null)"
    echo "valkey server:   $(timeout 60 docker run --rm "$image" valkey-server --version 2>/dev/null)"
    echo "haproxy image:   $(docker image inspect haproxy:3.2-alpine --format '{{index .RepoDigests 0}}' 2>/dev/null)"
    echo "go:              $(go version 2>/dev/null || echo '?')"
    echo
    cat src/SHAS-208-209.txt 2>/dev/null
    echo
    for b in bench/bin/bench-*; do
      echo "$b: $(go version -m "$b" 2>/dev/null | awk '$1 == "=>" { print $2 }') $(go version -m "$b" 2>/dev/null | awk '$1 == "build" && $2 ~ /^(-tags|-ldflags)/ { printf "%s ", $2 }')"
    done
    echo
    echo "queue at start ($QUEUE):"
    cat "$QUEUE"
  } >"$OUT/meta.txt" 2>&1
}

queue_lines() { grep -c '' "$QUEUE" 2>/dev/null || echo 0; }

# est_seconds LINE: expected wall time of one queue line (bench duration + about 13 s of cluster setup).
est_seconds() {
  local d
  d=$(sed -n 's/.*-duration \([0-9]*\)s.*/\1/p' <<<"$1")
  [[ "$1" == *"-fail-kind migrate"* ]] && d=18
  echo $((${d:-40} + 13))
}

remaining_estimate() { # remaining_estimate FROM_LINE: minutes for the lines from FROM_LINE on
  local n i total=0 line
  n=$(queue_lines)
  for ((i = $1; i <= n; i++)); do
    line=$(sed -n "${i}p" "$QUEUE")
    [[ -z "${line// /}" || "$line" == \#* ]] && continue
    total=$((total + $(est_seconds "$line")))
  done
  echo $(((total + 59) / 60))
}

write_meta
plog "runner: start, queue $QUEUE ($(grep -cv -E '^\s*(#|$)' "$QUEUE") runs, about $(remaining_estimate 1) min), output $OUT"

i=0
while ((!stopping)); do
  i=$((i + 1))
  if [[ -e "$OUT/STOP" ]]; then
    plog "runner: $OUT/STOP exists, stopping before line $i"
    break
  fi
  ((i <= $(queue_lines))) || break
  line=$(sed -n "${i}p" "$QUEUE")
  [[ -z "${line// /}" || "$line" == \#* ]] && continue
  read -r scenario label round tmo rest <<<"$line"
  if [[ -z "${rest:-}" || ! "$round" =~ ^[0-9]+$ || ! "$tmo" =~ ^[0-9]+$ ]]; then
    plog "runner: line $i is malformed, skipped: $line"
    continue
  fi
  read -r -a cmd <<<"$rest"
  dir=$OUT/$scenario
  name=$label-$round
  mkdir -p "$dir"
  rm -f "$dir/$name.json" "$dir/$name.log"
  started=$(date +%s)
  plog "=== line $i: $scenario $name (about $(remaining_estimate "$i") min left) ==="

  status=ok
  up_ok=0
  for attempt in 1 2; do
    timeout 90 ./down.sh >"$dir/$name.up.log" 2>&1
    if timeout "$UP_TIMEOUT" ./up.sh >>"$dir/$name.up.log" 2>&1; then
      up_ok=1
      break
    fi
    plog "runner: up.sh failed (attempt $attempt), see $dir/$name.up.log"
    ((stopping)) && break
  done
  rc=0
  if ((!up_ok)); then
    status=up_failed
  elif ((!stopping)); then
    if [[ ! -x "${cmd[0]}" ]]; then
      status=no_binary
      rc=127
    else
      bg_run "$dir/$name.log" "$tmo" "${cmd[@]}" -label "$label" -run "$round" -json "$dir/$name.json"
      if ((rc == 124 || rc == 137)); then
        status=timeout
      elif ((rc != 0)); then
        status=bench_failed
      fi
      [[ -s "$dir/$name.json" ]] || { [[ $status == ok ]] && status=no_json; }
    fi
    docker logs lf-haproxy >"$dir/$name.haproxy.log" 2>&1
    for p in "${PORTS[@]}"; do
      docker logs "$(container_of "$p")" 2>&1 | sed "s/^/$p| /"
    done >"$dir/$name.valkey.log"
  fi
  ((stopping)) && [[ $status == ok ]] && status=interrupted
  elapsed=$(($(date +%s) - started))
  printf '{"line":%d,"scenario":%s,"label":%s,"round":%d,"status":%s,"exit":%d,"elapsed_s":%d,"started":%s,"json":%s}\n' \
    "$i" "$(json_str "$scenario")" "$(json_str "$label")" "$round" "$(json_str "$status")" "$rc" "$elapsed" \
    "$(json_str "$(date -d "@$started" -Is)")" "$(json_str "$dir/$name.json")" >>"$RESULTS"
  summary=$(grep -E '^(requests |CLUSTER SHARDS sent|longest success|promotion |server errors|latency of successes: run|migration )' "$dir/$name.log" 2>/dev/null |
    sed -E 's/; by address.*//; s/ \(started at[^)]*\)//' | tr '\n' ' ' | cut -c1-900)
  plog "$scenario $name: $status (exit $rc, ${elapsed}s) $summary"
  if [[ $status != ok ]]; then
    tail -5 "$dir/$name.log" 2>/dev/null | sed 's/^/    /' | tee -a "$PROGRESS"
  fi
done
