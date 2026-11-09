# Test bed and results for valkey-go PR #209

This is the failover test bed we used for the numbers in [valkey-io/valkey-go#209](https://github.com/valkey-io/valkey-go/pull/209), together with every run behind them. Everything runs on one Linux machine with Docker. No cloud account is needed.

```
                         127.0.0.1:16379  (discovery endpoint: HAProxy, TCP, round robin, PING health check)
 bench ──InitAddress──►  HAProxy ──► one node per TCP connection
   │                         ▼
   └── direct to nodes ► 7101 7102 7103 primaries, 7104 7105 7106 replicas
                         Valkey 9.0.6, cluster-node-timeout 5000, one replica per shard
```

The bench is a Go program built against a chosen valkey-go tree. It runs 200 goroutines (alternating GET and SET on 1,000 keys of the failing shard, 2 s deadline per request, spread over `-clients` clients) through `PreferInitAddressRefresh` against the discovery endpoint, injects one failure, and records every request, every CLUSTER SHARDS the client sends (timestamps), MOVED/ASK counts from `INFO errorstats`, and the cluster's own view of the failover (a watcher polls CLUSTER SHARDS on every node every 100 ms, and every 20 ms on the replicas that can take over). One run writes one JSON file; `analyze-ar.py` turns a directory of runs into the tables.

## Requirements

Docker with the compose plugin, Go 1.25 or later, git, Python 3, and `redis-cli` or `valkey-cli`. The nodes and HAProxy use host networking on 127.0.0.1 ports 7101–7106, 17101–17106 and 16379.

## Reproduce

The PR after its second review (9 November), against main:

```bash
git clone --branch pr209-testbed --single-branch https://github.com/gunreddyg-debug/valkey-go pr209-testbed
cd pr209-testbed/testbed
./setup-trees.sh                                   # valkey-go trees under src/ (main, the PR, baselines; see the script)
cd bench-ar && TREES="m2 m2vd r2 r2big" ./build-ar.sh && cd ..
OUT=runs-second-review QUEUE=queues/runs-second-review.txt ./run-208-209.sh   # 51 runs, about 35 minutes
./analyze-ar.py runs-second-review > runs-second-review/analysis.md
```

The version reviewed on 8 November, which still had the wait for a slot without a node and the refresh after a
request deadline:

```bash
cd bench-ar && TREES="m2 m2vd final8 final8big finalboth" ./build-ar.sh && cd ..
OUT=runs-reproduce QUEUE=queues/reproduce.txt ./run-208-209.sh   # about 60 runs, 1 hour; fresh cluster for every run
./analyze-ar.py runs-reproduce > runs-reproduce/analysis.md
```

`run-208-209.sh` brings a fresh cluster up for every run (`up.sh`), runs one line of the queue, saves `<scenario>/<label>-<round>.json` plus the bench, HAProxy and Valkey logs, and takes the cluster down at the end (`down.sh`). The queue file is read again before every run, so lines can be appended while it runs. `touch $OUT/STOP` stops it after the current run.

One bench run by hand, for example a crash with the PR and nothing set:

```bash
./up.sh
bench-ar/bin/bench-final8 -fail-mode kill -fail-seconds 12 -fail-hold 0 -duration 25s
./down.sh
```

## Builds

| Tree (`src/`) | Binary | What | Bench flags |
| --- | --- | --- | --- |
| `m2` | `bench-m2` | upstream main `a23e130`, the base of the PR | none |
| `m2vd` | `bench-m2vd` | main + `MaxRefreshDelay` from Vardhaman's draft; only used with `-max-refresh-delay 30s` to emulate main at 1,000 nodes | `-max-refresh-delay` |
| `r2` | `bench-r2` | the PR after its second review (`pull/209/head`): no refresh after a request deadline, no wait for a slot without a node | `-min-refresh`, `-max-refresh-delay` (0: the defaults) |
| `r2big` | `bench-r2big` | `r2` with `nodeCount()` fixed at 1,000 (`patches/emulate-1000-nodes.patch`) | same |
| `final8` | `bench-final8` | the PR as reviewed on 8 November (fork branch `pr209-before-second-review`) | `-min-refresh`, `-max-slot-wait`, `-max-refresh-delay` (0: the defaults) |
| `final8big` | `bench-final8big` | the PR with `nodeCount()` fixed at 1,000 (`patches/emulate-1000-nodes.patch`), so every default that depends on the cluster size takes its 1,000-node value: 1 s interval, 200 ms gap, delays up to 30 s | same |
| `finalboth` | `bench-finalboth` | the earlier version of the PR with fixed settings (`patches/fixed-settings-split-both.mbox`), for the fixed 50 ms / 250 ms comparison | same |

## Scenarios

| Scenario | Flags | What happens |
| --- | --- | --- |
| `smart-kill` | `-fail-mode kill -fail-seconds 12` | The primary of slot 0 is killed at t = 5 s and restarted 12 s later. The cluster flags it failed after about 5 s and promotes its replica a few hundred ms later. |
| `smart-hold3` | `-fail-hold 3` | Same, but the promotion is held 3 s after the failure flag (`cluster-replica-no-failover`, then `CLUSTER FAILOVER FORCE`) |
| `smart-lazy` | `-fail-hold 3 -shards-refresh 1s` | As `smart-hold3`, with a 1 s `ShardsRefreshInterval`, so that a lazy refresh is often pending at the promotion |
| `smart-elect5` | `-fail-hold 5` | A 5 s election |
| `smart-fleet` | `-clients 20` | A crash with 20 clients sharing the discovery endpoint |
| `smart-planned` | `-fail-kind planned` | `CLUSTER FAILOVER` on the replica of slot 0's primary |
| `smart-pause` | `-fail-mode pause` | A hung primary: `docker pause` for 12 s |
| `smart-migrate` | `-fail-kind migrate` | Atomic slot migration of slots 0–199 to another primary under load |
| `smart-idle-kill` | `-idle-from 4s -idle-until 22s -shards-refresh 0s` | The client is idle while the primary crashes and resumes with the dead node in its slot map |
| `smart-pipe-kill` | `-pipeline 10` | A crash with `DoMulti` pipelines of 10 commands |
| `smart-big-*` | `bench-final8big`, `bench-m2vd -max-refresh-delay 30s` | The 1,000-node emulation: crash, crash with a 1 s periodic refresh, hung primary, periodic refresh without failure |
| `steady` | `-no-fail` | No failure, for throughput and latency |

The main metric is the time from the promotion (as the watcher sees it) to the first successful request for the failed shard. "Refreshes" are the CLUSTER SHARDS the client sends after the failure.

## Results in this package

| Directory | What | Runs |
| --- | --- | --- |
| `results/runs-second-review` | The PR after its second review against main, every scenario, alternating builds | 51 |
| `results/runs-final8` | The version reviewed on 8 November with nothing set, every scenario | 19 |
| `results/runs-holistic` | `MaxRefreshDelay` samples (1 s, 250 ms, 50 ms; 30 s, 1 s, 250 ms at 1,000 nodes) and the periodic refresh at 1,000 nodes | 25 |
| `results/runs-smart` | The development comparison: main, the fixed settings (50, 100, 250 ms), and the intermediate versions of the PR, including the 1,000-node stall and its fix | 136 |

Each directory has `analysis.md` (the tables), `summary.md`, `results.jsonl` (one line per run), `meta.txt` (image digests and commands), the queue it ran, and per run the JSON and the logs. The labels are explained at the top of each table. The intermediate versions in `runs-smart` (labels `smart`, `smart2`, `smart3`, `smart5`, `smart6`, `v3big` to `v6big`) are local commits that led to the PR and are not needed to reproduce its numbers.

The product code measured in `runs-second-review` (`r2`, the PR's code on main a23e130, before the rebase) is identical to the PR after its second review: `cluster.go`, `valkey.go` and `singleflight.go` are the same. The product code measured in `runs-final8` is that of the version reviewed on 8 November.

## Limits

- One machine, 3 shards, Docker networking, HAProxy as the discovery endpoint. Cloud networks add latency and jitter.
- 1 to 5 runs per case and build. We report medians and ranges, not confidence intervals, except where the tables say so.
- The 1,000-node emulation changes only what the client derives from the node count. The CLUSTER SHARDS replies stay small, so it does not show the cost of refreshing a real 1,000-node cluster, nor a fleet of thousands of clients.
- The crash timing has a few tens of ms of noise: a refresh right after the promotion sometimes reaches a node that has not heard of the new primary yet, and the next poll finds it.

## Large-cluster simulation

`simulation/zz_largesim_test.go.txt` drives the real cluster client against the mocked cluster of the PR's tests (`swCluster`) with topologies of 60, 200, 1,000 and 1,500 nodes: a primary crashes, the cluster flags it failed 1 s later and promotes its replica 300–800 ms after that, while 50 goroutines read. It reports the time from the promotion to the first success, `ErrNoSlot`, and refreshes, for the defaults, the defaults with a 1 s `ShardsRefreshInterval`, everything at -1, and fixed 50 ms and 250 ms intervals.

```bash
cp simulation/zz_largesim_test.go.txt <valkey-go>/zz_largesim_test.go
cd <valkey-go> && LARGESIM=1 GOMAXPROCS=4 go test -run TestLargeClusterSim -v -timeout 60m .
rm zz_largesim_test.go
```

`LARGESIM_SIZES=60,1000` and `LARGESIM_TRIALS=10` change the sizes and the trials.

## The PR's own tests

```bash
# unit tests of the PR, mocked connections, no Docker (about 3 minutes)
go test -race -run 'TestCluster|TestSingleFlight|TestCall|TestSlot' -skip Integration .

# the whole suite as CI runs it, against the Docker services of the repository
docker compose up -d
go test -race -timeout 30m ./...
for d in valkeycompat valkeyaside om valkeyhook valkeyotel valkeylimiter valkeyprob mock valkeycompatmock valkeycompatrate; do (cd $d && go test -race -timeout 30m ./...); done
docker compose down -v
```
