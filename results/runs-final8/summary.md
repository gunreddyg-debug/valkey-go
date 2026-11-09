Tables made by `bench/bin/report runs-final8`. Ranges are min – max over the runs of each build.

## Runs

| Scenario | Runs | OK | Not OK | Wall time |
|---|---|---|---|---|
| smart-planned | 4 | 4 | – | 2.1 min |
| smart-kill | 5 | 5 | – | 3.1 min |
| smart-hold3 | 1 | 1 | – | 0.6 min |
| smart-pause | 2 | 2 | – | 1.4 min |
| smart-fleet | 1 | 1 | – | 0.6 min |
| smart-lazy | 1 | 1 | – | 0.6 min |
| steady | 1 | 1 | – | 0.6 min |
| smart-elect5 | 1 | 1 | – | 0.7 min |
| smart-migrate | 1 | 1 | – | 0.5 min |
| smart-idle-kill | 1 | 1 | – | 0.8 min |
| smart-pipe-kill | 1 | 1 | – | 0.6 min |

## steady

no failure (steady state); ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s | 1 | 5,255,890 | 0 | 0 | 0 | 1 / 1 | 0 | 0 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Longest success |
|---|---|---|---|
| s8 | 262,793 | 0.60 / 3.30 / 12.16 | 29 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 5,255,890 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 3.30 / 0.00 | – | – |

</details>

## smart-elect5

`MODE=kill HOLD_PROMOTION=5 ./fail-primary.sh 20` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s | 1 | 4,906,126 | 0 | 437 | 1,716 | 93 / 92 / 62 / 19 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 11.96 s | +77 ms | 11.96 s | 0 ms | never | 163,533 | 0.57 / 3.30 / 12.16 | 0.52 / 3.30 / 13.18 | 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 4,906,126 | 0 | 437 | 1,716 | 93 / 92 / 62 / 19 | 0 | 0 / 0 | 11.96 s | +77 ms | 3.30 / 3.30 | – | 1,606 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-fleet

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f_s8 | 0s / 0s | 1 | 2,406,243 | 0 | 2 | 1,967 | 769 / 749 / 97 / 122 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f_s8 | 8.16 s | +3 ms | 8.21 s | 0 ms | never | 96,246 | 1.23 / 7.23 / 13.18 | 1.20 / 6.21 / 495.62 | 1.97 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f_s8-1 | 2,406,243 | 0 | 2 | 1,967 | 769 / 749 / 97 / 122 | 0 | 0 / 0 | 8.16 s | +3 ms | 7.23 / 6.21 | – | 1,870 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-hold3

`MODE=kill HOLD_PROMOTION=3 ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s | 1 | 4,330,564 | 0 | 264 | 1,971 | 81 / 80 / 47 / 19 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 10.95 s | +73 ms | 10.94 s | 0 ms | never | 173,219 | 0.52 / 2.66 / 11.90 | 0.50 / 2.85 / 11.90 | 1.96 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 4,330,564 | 0 | 264 | 1,971 | 81 / 80 / 47 / 19 | 0 | 0 / 0 | 10.95 s | +73 ms | 2.66 / 2.85 | – | 1,874 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-idle-kill

`MODE=kill ./fail-primary.sh 30` at t = 5 s; ShardsRefreshInterval 0s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s, ShardsRefresh 0s | 1 | 3,393,857 | 0 | 0 | 106 | 3 / 2 / 0 / 2 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 16.77 s | +10557 ms | 16.75 s | 0 ms | never | 113,125 | 0.55 / 3.17 / 11.90 | 0.52 / 2.53 / 12.67 | 35 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 3,393,857 | 0 | 0 | 106 | 3 / 2 / 0 / 2 | 0 | 0 / 0 | 16.77 s | +10557 ms | 3.17 / 2.53 | – | 106 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-kill

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s | 5 | 4,337,692 – 5,361,602 | 0 | 1 – 7 | 1,429 – 1,799 | 33 – 39 / 32 – 38 / 5 – 7 / 7 – 8 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 6.28 – 7.84 s | +45 – +90 ms | 6.28 – 7.81 s | 0 ms | never | 173,506–214,455 | 0.52–0.58 / 2.85–3.55 / 11.90–12.16 | 0.47–0.58 / 2.00–3.23 / 12.42–13.70 | 1.94 – 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 5,361,602 | 0 | 1 | 1,704 | 36 / 35 / 5 / 7 | 0 | 0 / 0 | 7.32 s | +78 ms | 2.85 / 2.00 | – | 1,612 × dial tcp 127.0.0.1:7101: connect: connection refused |
| s8-2 | 5,008,520 | 0 | 5 | 1,557 | 33 / 33 / 5 / 8 | 0 | 0 / 0 | 6.64 s | +71 ms | 3.36 / 3.17 | – | 1,454 × dial tcp 127.0.0.1:7101: connect: connection refused |
| s8-3 | 5,111,159 | 0 | 7 | 1,429 | 33 / 32 / 6 / 8 | 0 | 0 / 0 | 6.28 s | +90 ms | 3.49 / 3.23 | – | 1,324 × dial tcp 127.0.0.1:7101: connect: connection refused |
| s8-4 | 4,732,656 | 0 | 5 | 1,771 | 38 / 37 / 6 / 8 | 0 | 0 / 0 | 7.79 s | +54 ms | 3.36 / 2.66 | – | 1,686 × dial tcp 127.0.0.1:7101: connect: connection refused |
| s8-5 | 4,337,692 | 0 | 5 | 1,799 | 39 / 38 / 7 / 8 | 0 | 0 / 0 | 7.84 s | +45 ms | 3.55 / 3.23 | – | 1,704 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-lazy

`MODE=kill HOLD_PROMOTION=3 ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 1s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s, ShardsRefresh 1s | 1 | 3,514,788 | 0 | 229 | 2,041 | 93 / 87 / 46 / 18 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 11.06 s | +148 ms | 11.00 s | 0 ms | never | 140,588 | 0.58 / 3.68 / 12.16 | 0.52 / 3.30 / 12.93 | 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 3,514,788 | 0 | 229 | 2,041 | 93 / 87 / 46 / 18 | 0 | 0 / 0 | 11.06 s | +148 ms | 3.68 / 3.30 | – | 1,941 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-migrate

Atomic Slot Migration of the workload's slots from 127.0.0.1:7101 to 127.0.0.1:7103 at t = 5 s; the run ends 10 s after the migration succeeded; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s | 1 | 4,222,734 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 315 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, migration window (ms) | Longest success |
|---|---|---|---|---|
| s8 | 277,017 | 0.57 / 3.23 / 11.90 | 0.58 / 3.04 / 12.93 | 67 ms |

Migration. Times from MIGRATESLOTS: source (EXPORT) and target (IMPORT) report success (CLUSTER GETSLOTMIGRATIONS polled every 5 ms), the watcher sees the target own the slot. The handover is the first success report. From the handover: "traffic to the source until" is the end of the last 10 ms in which the client still wrote workload traffic (> 512 bytes) to the source, i.e. when all migrated slots routed to the new owner; "last MOVED" is the last INFO errorstats read of the source (every 20 ms) in which its MOVED count grew.

| Build | Source success | Target success | Handover seen (watcher) | Traffic to the source until (after handover) | Last MOVED from the source (after handover) | CLUSTER GETSLOTMIGRATIONS sent (run) | Off-endpoint CLUSTER SHARDS (by address, all runs) | Errors |
|---|---|---|---|---|---|---|---|---|
| s8 | 234 ms | 224 ms | 239 ms | 22 ms | 35 ms | 0 | 0 | – |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 4,222,734 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 315 / 0 | 1 ms | -238 ms | 3.23 / 3.04 | – | – |

</details>

## smart-pause

`MODE=pause ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s | 2 | 5,116,808 – 5,736,654 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 9.98 s | +2.77 – +2.81 s | 9.98 s | 780 – 807 ms | never | 170,559–191,219 | 0.55–0.60 / 3.17–3.49 / 11.90–12.16 | 0.51–0.60 / 2.34–3.49 / 11.39–12.42 | 23 – 28 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 5,736,654 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +2771 ms | 3.17 / 2.34 | – | – |
| s8-2 | 5,116,808 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +2808 ms | 3.49 / 3.49 | – | – |

</details>

## smart-pipe-kill

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 0s / 0s | 1 | 10,717,963 | 0 | 8,380 | 6,107 | 26 / 25 / 0 / 4 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s8 | 7.78 s | +420 ms | 7.83 s | 0 ms | never | 428,711 | 2.78 / 13.70 / 22.27 | 2.78 / 13.95 / 18.18 | 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s8-1 | 10,717,963 | 0 | 8,380 | 6,107 | 26 / 25 / 0 / 4 | 0 | 0 / 0 | 7.78 s | +420 ms | 13.70 / 13.95 | – | 5,170 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-planned

planned failover (`CLUSTER FAILOVER` on the replica of slot 0's primary) at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f_s8 | 0s / 0s | 2 | 2,836,554 – 3,053,964 | 0 | 0 | 0 | 56 – 59 / 56 – 59 / 0 / 31 – 35 | 0 | 9,103 – 10,443 / 0 |
| s8 | 0s / 0s | 2 | 5,095,481 – 5,141,695 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,046 – 1,073 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f_s8 | 1 – 3 ms | -120 – -118 ms | never | 186 – 189 ms | 220 – 227 ms | 141,826–152,694 | 1.17–1.23 / 6.08–7.10 / 12.93–13.18 | 1.23–1.36 / 7.23–8.32 / 58.88–59.90 | 63 ms |
| s8 | 0 ms | -124 – -113 ms | never | 0 ms | 25 – 40 ms | 254,766–257,083 | 0.60 / 3.42–3.68 / 12.16–12.42 | 0.60–0.62 / 3.42 / 12.42–13.44 | 57 – 59 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f_s8-1 | 3,053,964 | 0 | 0 | 0 | 59 / 59 / 0 / 35 | 0 | 10,443 / 0 | 3 ms | -120 ms | 6.08 / 7.23 | – | – |
| f_s8-2 | 2,836,554 | 0 | 0 | 0 | 56 / 56 / 0 / 31 | 0 | 9,103 / 0 | 1 ms | -118 ms | 7.10 / 8.32 | – | – |
| s8-1 | 5,141,695 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,046 / 0 | 0 ms | -124 ms | 3.68 / 3.42 | – | – |
| s8-2 | 5,095,481 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,073 / 0 | 0 ms | -113 ms | 3.42 / 3.42 | – | – |

</details>

## No-primary window and promotion, per run

From the watcher (CLUSTER SHARDS on every node every 100 ms, on the replica every 20 ms) and the fail script's lines; times from the failure. "Discovery node": where HAProxy sent the client's long-lived discovery-endpoint connection(s).

| Scenario | Run | Old primary flagged fail | No online primary from | No-primary window | New primary online (watcher) | PROMOTE line | Client CLUSTER SHARDS in window | Discovery node |
|---|---|---|---|---|---|---|---|---|
| smart-elect5 | s8-1 | 6.81 s | 6.81 s | 5.12 s | 11.89 s | 11.87 s | 62 | 7102 |
| smart-fleet | f_s8-1 | 7.81 s | 7.81 s | 359 ms | 8.15 s | never | 97 | 7101+7101+7101+7104+7105+7103+7102+7105+7103+7104+7105+7102+7106+7104+7102+7103+7102+7105+7104+7106+7106+7106 |
| smart-hold3 | s8-1 | 7.82 s | 7.82 s | 3.08 s | 10.88 s | 10.86 s | 47 | 7102 |
| smart-idle-kill | s8-1 | 5.95 s | 5.95 s | 261 ms | 6.21 s | never | 0 | 7102 |
| smart-kill | s8-1 | 6.90 s | 6.90 s | 340 ms | 7.24 s | never | 5 | 7102 |
| smart-kill | s8-2 | 6.23 s | 6.23 s | 337 ms | 6.57 s | never | 5 | 7102 |
| smart-kill | s8-3 | 5.85 s | 5.85 s | 359 ms | 6.19 s | never | 6 | 7102 |
| smart-kill | s8-4 | 7.47 s | 7.47 s | 301 ms | 7.73 s | never | 6 | 7102 |
| smart-kill | s8-5 | 7.39 s | 7.39 s | 420 ms | 7.79 s | never | 7 | 7102 |
| smart-lazy | s8-1 | 7.83 s | 7.83 s | 3.08 s | 10.91 s | 10.91 s | 46 | 7102 |
| smart-pause | s8-1 | 6.87 s | 6.87 s | 340 ms | 7.21 s | never | 0 | 7102 |
| smart-pause | s8-2 | 6.87 s | 6.87 s | 321 ms | 7.17 s | never | 0 | 7102 |
| smart-pipe-kill | s8-1 | 7.10 s | 7.10 s | 259 ms | 7.36 s | never | 0 | 7102 |
| smart-planned | f_s8-1 | never | never | never | 123 ms | 14 ms | 0 | 7106+7102+7103+7104+7106+7102+7103+7103+7106+7102+7104+7103+7102+7101+7101+7104+7101+7105+7105+7105 |
| smart-planned | f_s8-2 | never | never | never | 119 ms | 14 ms | 0 | 7102+7103+7106+7102+7106+7103+7102+7103+7104+7104+7101+7104+7102+7101+7105+7105 |
| smart-planned | s8-1 | never | never | never | 125 ms | 12 ms | 0 | 7102 |
| smart-planned | s8-2 | never | never | never | 113 ms | 14 ms | 0 | 7102 |

