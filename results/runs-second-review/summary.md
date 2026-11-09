Tables made by `bench/bin/report runs-r2`. Ranges are min – max over the runs of each build.

## Runs

| Scenario | Runs | OK | Not OK | Wall time |
|---|---|---|---|---|
| smart-kill | 6 | 6 | – | 3.7 min |
| smart-hold3 | 6 | 6 | – | 3.7 min |
| smart-planned | 8 | 8 | – | 4.2 min |
| smart-pause | 6 | 6 | – | 4.1 min |
| smart-migrate | 4 | 4 | – | 1.8 min |
| smart-idle-kill | 4 | 4 | – | 3.1 min |
| smart-fleet | 4 | 4 | – | 2.5 min |
| smart-elect5 | 2 | 2 | – | 1.4 min |
| steady | 4 | 4 | – | 2.1 min |
| smart-big-kill | 3 | 3 | – | 3.1 min |
| smart-big-steady | 2 | 2 | – | 1.1 min |
| smart-big-pause | 2 | 2 | – | 2.1 min |

## steady

no failure (steady state); ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 2 | 6,024,602 – 6,571,754 | 0 | 0 | 0 | 1 / 1 | 0 | 0 / 0 |
| r2 | 0s /  | 2 | 6,398,952 – 6,602,307 | 0 | 0 | 0 | 1 / 1 | 0 | 0 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Longest success |
|---|---|---|---|
| main | 301,222–328,557 | 0.55–0.58 / 1.74–1.97 / 2.53–2.98 | 5 ms |
| r2 | 319,939–330,106 | 0.55 / 1.74–1.84 / 2.59–2.85 | 4 – 6 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 6,571,754 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 1.74 / 0.00 | – | – |
| main-2 | 6,024,602 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 1.97 / 0.00 | – | – |
| r2-1 | 6,602,307 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 1.74 / 0.00 | – | – |
| r2-2 | 6,398,952 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 1.84 / 0.00 | – | – |

</details>

## smart-big-kill

`MODE=kill ./fail-primary.sh 30` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| b2 | 0s /  | 2 | 14,833,451 – 14,858,844 | 0 – 13,688 | 0 | 1,756 – 2,075 | 13 – 321 / 13 – 321 / 0 – 311 / 1 – 311 | 0 | 0 / 0 |
| lmain | – | 1 | 11,254,099 | 0 | 0 | 4,527 | 1 / 1 / 0 / 1 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| b2 | 7.55 – 8.38 s | +31 – +673 ms | 7.54 – 8.37 s | 0 ms | never | 296,667–297,174 | 0.50–0.51 / 1.65–1.74 / 2.46–2.66 | 0.42–0.44 / 1.52–1.62 / 2.72–2.85 | 1.93 – 1.94 s |
| lmain | 18.12 s | +9235 ms | 18.11 s | 0 ms | never | 225,078 | 0.50 / 1.68 / 2.40 | 0.40 / 1.49 / 2.27 | 1.98 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| b2-1 | 14,833,451 | 0 | 0 | 2,075 | 13 / 13 / 0 / 1 | 0 | 0 / 0 | 8.38 s | +673 ms | 1.65 / 1.52 | – | 1,980 × dial tcp 127.0.0.1:7101: connect: connection refused |
| b2-2 | 14,858,844 | 13,688 | 0 | 1,756 | 321 / 321 / 311 / 311 | 0 | 0 / 0 | 7.55 s | +31 ms | 1.74 / 1.62 | – | 1,648 × dial tcp 127.0.0.1:7101: connect: connection refused |
| lmain-1 | 11,254,099 | 0 | 0 | 4,527 | 1 / 1 / 0 / 1 | 0 | 0 / 0 | 18.12 s | +9235 ms | 1.68 / 1.49 | – | 4,432 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-big-pause

`MODE=pause ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| b2 | 0s /  | 1 | 13,660,662 | 0 | 1,000 | 0 | 5 / 5 / 0 / 1 | 0 | 0 / 0 |
| lmain | – | 1 | 12,276,961 | 0 | 1,200 | 104 | 2 / 2 / 0 / 1 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| b2 | 9.98 s | +2362 ms | 9.98 s | 368 ms | never | 273,209 | 0.52 / 1.71 / 2.46 | 0.48 / 1.78 / 2.46 | 11 ms |
| lmain | 13.88 s | +6300 ms | 13.88 s | 2.41 s | never | 245,537 | 0.54 / 1.71 / 2.53 | 0.48 / 1.74 / 2.72 | 8 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| b2-1 | 13,660,662 | 0 | 1,000 | 0 | 5 / 5 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +2362 ms | 1.71 / 1.78 | – | – |
| lmain-1 | 12,276,961 | 0 | 1,200 | 104 | 2 / 2 / 0 / 1 | 0 | 0 / 0 | 13.88 s | +6300 ms | 1.71 / 1.74 | – | 104 × i/o timeout |

</details>

## smart-big-steady

no failure (steady state); ShardsRefreshInterval 1s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| b2 | 0s / , ShardsRefresh 1s | 1 | 6,519,715 | 0 | 0 | 0 | 19 / 1 | 0 | 0 / 0 |
| lmain | ShardsRefresh 1s | 1 | 6,474,837 | 0 | 0 | 0 | 0 / 0 | 0 | 0 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Longest success |
|---|---|---|---|
| b2 | 325,979 | 0.55 / 1.74 / 2.46 | 5 ms |
| lmain | 323,734 | 0.55 / 1.78 / 2.78 | 12 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| b2-1 | 6,519,715 | 0 | 0 | 0 | 19 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 1.74 / 0.00 | – | – |
| lmain-1 | 6,474,837 | 0 | 0 | 0 | 0 / 0 / 0 / 0 | 0 | 0 / 0 | never | – | 1.78 / 0.00 | – | – |

</details>

## smart-elect5

`MODE=kill HOLD_PROMOTION=5 ./fail-primary.sh 20` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 1 | 6,253,910 | 1,093,099 | 23 | 1,694 | 5,580 / 5,579 / 5,543 / 1,407 | 0 | 0 / 0 |
| r2 | 0s /  | 1 | 6,165,984 | 1,237,295 | 0 | 1,691 | 6,684 / 6,684 / 6,616 / 1,412 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 11.80 s | +28 ms | 11.79 s | 0 ms | never | 208,460 | 0.51 / 1.81 / 2.78 | 0.48 / 1.62 / 2.46 | 10 ms |
| r2 | 11.69 s | +40 ms | 11.69 s | 0 ms | never | 205,523 | 0.52 / 1.81 / 2.78 | 0.49 / 1.74 / 2.59 | 10 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 6,253,910 | 1,093,099 | 23 | 1,694 | 5,580 / 5,579 / 5,543 / 1,407 | 0 | 0 / 0 | 11.80 s | +28 ms | 1.81 / 1.62 | – | 1,572 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-1 | 6,165,984 | 1,237,295 | 0 | 1,691 | 6,684 / 6,684 / 6,616 / 1,412 | 0 | 0 / 0 | 11.69 s | +40 ms | 1.81 / 1.74 | – | 1,588 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-fleet

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f_main | – | 2 | 2,797,847 – 3,086,249 | 8,610 – 12,454 | 0 | 1,594 – 1,977 | 2,450 – 2,969 / 2,444 – 2,963 / 2,183 – 2,650 / 2,206 – 2,540 | 0 | 0 / 0 |
| f_r2 | 0s /  | 2 | 2,771,545 – 2,865,140 | 11,210 – 12,384 | 0 – 1 | 1,803 – 2,006 | 3,375 – 3,694 / 3,355 – 3,675 / 2,346 – 2,870 / 2,728 – 3,123 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f_main | 6.50 – 7.94 s | -5 – -2 ms | 6.94 – 8.30 s | 0 ms | never | 111,913–123,447 | 1.10–1.14 / 2.40–2.46 / 3.23–3.36 | 0.95–1.04 / 2.08–2.27 / 3.17–3.23 | 1.94 – 1.95 s |
| f_r2 | 7.69 – 8.09 s | -4 – +4 ms | 7.72 – 8.22 s | 0 ms | never | 110,860–114,604 | 1.14 / 2.46 / 3.42–3.74 | 1.07–1.10 / 2.34–2.46 / 5.95–10.37 | 1.90 – 1.94 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f_main-1 | 2,797,847 | 12,454 | 0 | 1,977 | 2,969 / 2,963 / 2,650 / 2,540 | 0 | 0 / 0 | 7.94 s | -2 ms | 2.46 / 2.27 | – | 1,874 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_main-2 | 3,086,249 | 8,610 | 0 | 1,594 | 2,450 / 2,444 / 2,183 / 2,206 | 0 | 0 / 0 | 6.50 s | -5 ms | 2.40 / 2.08 | – | 1,502 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_r2-1 | 2,771,545 | 11,210 | 1 | 2,006 | 3,375 / 3,355 / 2,346 / 2,728 | 0 | 0 / 0 | 8.09 s | +4 ms | 2.46 / 2.34 | – | 1,912 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_r2-2 | 2,865,140 | 12,384 | 0 | 1,803 | 3,694 / 3,675 / 2,870 / 3,123 | 0 | 0 / 0 | 7.69 s | -4 ms | 2.46 / 2.46 | – | 1,698 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-hold3

`MODE=kill HOLD_PROMOTION=3 ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 3 | 4,761,055 – 5,040,343 | 607,370 – 726,784 | 0 – 1 | 1,683 – 1,919 | 3,586 – 4,083 / 3,586 – 4,083 / 3,536 – 4,045 / 1,342 – 1,386 | 0 | 0 / 0 |
| r2 | 0s /  | 3 | 4,593,023 – 5,047,752 | 617,774 – 697,476 | 0 – 1 | 1,692 – 2,063 | 3,380 – 4,084 / 3,379 – 4,083 / 3,339 – 4,022 / 1,247 – 1,344 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 9.77 – 10.77 s | +31 – +38 ms | 9.76 – 10.76 s | 0 ms | never | 190,439–201,603 | 0.52 / 1.78–1.87 / 2.72–2.98 | 0.49–0.50 / 1.74–1.87 / 2.40–2.98 | 10 – 13 ms |
| r2 | 9.95 – 11.24 s | +36 – +39 ms | 9.94 – 11.24 s | 0 ms | never | 183,716–201,904 | 0.52 / 1.81–1.84 / 2.72–2.78 | 0.48–0.50 / 1.74–1.87 / 2.40–2.85 | 10 – 12 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 5,040,343 | 726,784 | 1 | 1,683 | 4,083 / 4,083 / 4,045 / 1,386 | 0 | 0 / 0 | 9.77 s | +31 ms | 1.84 / 1.74 | – | 1,575 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-2 | 4,882,696 | 607,370 | 0 | 1,797 | 3,586 / 3,586 / 3,536 / 1,356 | 0 | 0 / 0 | 10.10 s | +38 ms | 1.87 / 1.87 | – | 1,690 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-3 | 4,761,055 | 696,312 | 0 | 1,919 | 3,867 / 3,867 / 3,815 / 1,342 | 0 | 0 / 0 | 10.77 s | +38 ms | 1.78 / 1.81 | – | 1,816 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-1 | 5,047,752 | 617,774 | 0 | 1,707 | 3,380 / 3,379 / 3,339 / 1,247 | 0 | 0 / 0 | 9.97 s | +36 ms | 1.81 / 1.81 | – | 1,604 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-2 | 4,593,023 | 697,476 | 1 | 2,063 | 3,692 / 3,691 / 3,615 / 1,281 | 0 | 0 / 0 | 11.24 s | +39 ms | 1.84 / 1.74 | – | 1,966 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-3 | 5,023,752 | 691,123 | 0 | 1,692 | 4,084 / 4,083 / 4,022 / 1,344 | 0 | 0 / 0 | 9.95 s | +36 ms | 1.81 / 1.87 | – | 1,594 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-idle-kill

`MODE=kill ./fail-primary.sh 30` at t = 5 s; ShardsRefreshInterval 0s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | ShardsRefresh 0s | 2 | 3,774,419 – 4,144,796 | 0 | 0 | 101 – 110 | 3 / 2 – 3 / 0 / 2 | 0 | 0 / 0 |
| r2 | 0s / , ShardsRefresh 0s | 2 | 4,181,826 – 4,203,630 | 0 | 0 | 97 – 98 | 3 / 2 / 0 / 2 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 16.85 – 17.58 s | +9.41 – +10.30 s | 16.75 – 16.76 s | 0 ms | never | 125,813–138,159 | 0.50–0.51 / 1.71–1.74 / 2.53–2.59 | 0.36–0.47 / 1.46–1.71 / 2.40–2.66 | 0.24 – 1.80 s |
| r2 | 16.80 – 16.92 s | +9.40 – +10.45 s | 16.75 – 16.81 s | 0 ms | never | 139,392–140,119 | 0.50–0.51 / 1.68–1.71 / 2.46–2.53 | 0.49 / 1.78–1.90 / 2.66–3.30 | 84 – 230 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 4,144,796 | 0 | 0 | 101 | 3 / 3 / 0 / 2 | 0 | 0 / 0 | 16.85 s | +9410 ms | 1.74 / 1.71 | – | 101 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-2 | 3,774,419 | 0 | 0 | 110 | 3 / 2 / 0 / 2 | 0 | 0 / 0 | 17.58 s | +10302 ms | 1.71 / 1.46 | – | 110 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-1 | 4,181,826 | 0 | 0 | 97 | 3 / 2 / 0 / 2 | 0 | 0 / 0 | 16.80 s | +10447 ms | 1.71 / 1.78 | – | 97 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-2 | 4,203,630 | 0 | 0 | 98 | 3 / 2 / 0 / 2 | 0 | 0 / 0 | 16.92 s | +9403 ms | 1.68 / 1.90 | – | 98 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-kill

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 3 | 5,659,077 – 5,841,584 | 0 – 1,448 | 0 | 1,890 – 1,950 | 17 – 78 / 17 – 78 / 0 – 29 / 3 – 61 | 0 | 0 / 0 |
| r2 | 0s /  | 3 | 5,717,305 – 6,641,835 | 0 – 44,379 | 0 – 1 | 1,669 – 1,905 | 31 – 586 / 30 – 585 / 1 – 508 / 4 – 409 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 7.78 – 7.87 s | +30 – +94 ms | 7.78 – 7.86 s | 0 ms | never | 226,359–233,661 | 0.51–0.52 / 1.78–1.81 / 2.66–2.72 | 0.43–0.44 / 1.68–1.87 / 2.46–3.55 | 1.90 – 1.93 s |
| r2 | 7.03 – 7.96 s | +32 – +104 ms | 7.02 – 7.95 s | 0 ms | never | 228,690–265,669 | 0.48–0.54 / 1.65–1.94 / 2.66–3.17 | 0.41–0.49 / 1.14–1.65 / 1.97–2.78 | 1.92 – 1.99 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 5,687,430 | 0 | 0 | 1,950 | 18 / 18 / 0 / 3 | 0 | 0 / 0 | 7.87 s | +94 ms | 1.78 / 1.68 | – | 1,852 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-2 | 5,841,584 | 0 | 0 | 1,890 | 17 / 17 / 0 / 3 | 0 | 0 / 0 | 7.78 s | +31 ms | 1.81 / 1.68 | – | 1,786 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-3 | 5,659,077 | 1,448 | 0 | 1,918 | 78 / 78 / 29 / 61 | 0 | 0 / 0 | 7.81 s | +30 ms | 1.81 / 1.87 | – | 1,814 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-1 | 6,641,835 | 0 | 1 | 1,694 | 31 / 30 / 1 / 4 | 0 | 0 / 0 | 7.24 s | +104 ms | 1.65 / 1.14 | – | 1,602 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-2 | 5,735,291 | 24,495 | 0 | 1,905 | 359 / 358 / 279 / 280 | 0 | 0 / 0 | 7.96 s | +40 ms | 1.74 / 1.62 | – | 1,800 × dial tcp 127.0.0.1:7101: connect: connection refused |
| r2-3 | 5,717,305 | 44,379 | 0 | 1,669 | 586 / 585 / 508 / 409 | 0 | 0 / 0 | 7.03 s | +32 ms | 1.94 / 1.65 | – | 1,570 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-migrate

Atomic Slot Migration of the workload's slots from 127.0.0.1:7101 to 127.0.0.1:7103 at t = 5 s; the run ends 10 s after the migration succeeded; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 2 | 5,008,703 – 5,143,267 | 0 | 0 | 0 | 3 / 3 / 0 / 1 | 0 | 20,999 – 112,456 / 0 |
| r2 | 0s /  | 2 | 5,108,757 – 5,191,726 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 200 – 313 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, migration window (ms) | Longest success |
|---|---|---|---|---|
| main | 330,127–337,293 | 0.54 / 1.74–1.81 / 2.59–2.72 | 0.55–0.63 / 1.81–2.14 / 3.17–3.42 | 59 – 66 ms |
| r2 | 335,565–340,064 | 0.54 / 1.71–1.74 / 2.46–2.53 | 0.55 / 1.78–1.81 / 2.46–2.66 | 62 – 66 ms |

Migration. Times from MIGRATESLOTS: source (EXPORT) and target (IMPORT) report success (CLUSTER GETSLOTMIGRATIONS polled every 5 ms), the watcher sees the target own the slot. The handover is the first success report. From the handover: "traffic to the source until" is the end of the last 10 ms in which the client still wrote workload traffic (> 512 bytes) to the source, i.e. when all migrated slots routed to the new owner; "last MOVED" is the last INFO errorstats read of the source (every 20 ms) in which its MOVED count grew.

| Build | Source success | Target success | Handover seen (watcher) | Traffic to the source until (after handover) | Last MOVED from the source (after handover) | CLUSTER GETSLOTMIGRATIONS sent (run) | Off-endpoint CLUSTER SHARDS (by address, all runs) | Errors |
|---|---|---|---|---|---|---|---|---|
| main | 167 – 241 ms | 156 – 229 ms | 162 – 240 ms | 129 – 552 ms | 130 – 546 ms | 0 | 0 | – |
| r2 | 218 – 261 ms | 208 – 249 ms | 221 – 259 ms | 0 – 18 ms | 14 – 31 ms | 0 | 0 | – |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 5,143,267 | 0 | 0 | 0 | 3 / 3 / 0 / 1 | 0 | 20,999 / 0 | 1 ms | -239 ms | 1.74 / 1.81 | – | – |
| main-2 | 5,008,703 | 0 | 0 | 0 | 3 / 3 / 0 / 1 | 0 | 112,456 / 0 | 1 ms | -161 ms | 1.81 / 2.14 | – | – |
| r2-1 | 5,108,757 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 200 / 0 | 1 ms | -220 ms | 1.74 / 1.81 | – | – |
| r2-2 | 5,191,726 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 313 / 0 | 1 ms | -258 ms | 1.71 / 1.78 | – | – |

</details>

## smart-pause

`MODE=pause ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 3 | 5,892,207 – 6,070,465 | 0 | 1,097 – 1,110 | 90 – 103 | 4 / 4 / 0 / 1 – 2 | 0 | 25,378 – 101,132 / 0 |
| r2 | 0s /  | 3 | 5,723,648 – 6,005,906 | 0 | 1,096 – 1,107 | 93 – 104 | 4 / 4 / 0 / 2 | 0 | 0 – 198 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 12.08 – 12.09 s | +3.68 – +4.55 s | 12.08 – 12.09 s | 4.16 – 4.67 s | 4.16 – 4.74 s | 196,403–202,345 | 0.54 / 1.74–1.81 / 2.59–2.78 | 0.50–0.57 / 1.74–2.14 / 3.04–3.62 | 207 – 219 ms |
| r2 | 11.98 – 12.09 s | +4.33 – +5.90 s | 11.98 – 12.08 s | 2.33 – 4.50 s | 4.50 s (2× never) | 190,786–200,195 | 0.54–0.55 / 1.74–1.87 / 2.66–2.85 | 0.49–0.51 / 1.68–1.84 / 2.72–3.23 | 116 – 205 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 6,070,465 | 0 | 1,097 | 103 | 4 / 4 / 0 / 2 | 0 | 25,378 / 0 | 12.08 s | +4382 ms | 1.74 / 1.81 | – | 103 × i/o timeout |
| main-2 | 5,892,207 | 0 | 1,110 | 90 | 4 / 4 / 0 / 1 | 0 | 101,132 / 0 | 12.09 s | +3677 ms | 1.81 / 2.14 | – | 90 × i/o timeout |
| main-3 | 5,973,889 | 0 | 1,104 | 96 | 4 / 4 / 0 / 2 | 0 | 25,592 / 0 | 12.09 s | +4553 ms | 1.78 / 1.74 | – | 96 × i/o timeout |
| r2-1 | 5,723,648 | 0 | 1,096 | 104 | 4 / 4 / 0 / 2 | 0 | 0 / 0 | 11.98 s | +5898 ms | 1.87 / 1.84 | – | 104 × i/o timeout |
| r2-2 | 5,748,935 | 0 | 1,107 | 93 | 4 / 4 / 0 / 2 | 0 | 0 / 0 | 11.99 s | +4331 ms | 1.87 / 1.81 | – | 93 × i/o timeout |
| r2-3 | 6,005,906 | 0 | 1,098 | 102 | 4 / 4 / 0 / 2 | 0 | 198 / 0 | 12.09 s | +4502 ms | 1.74 / 1.68 | – | 102 × i/o timeout |

</details>

## smart-planned

planned failover (`CLUSTER FAILOVER` on the replica of slot 0's primary) at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f_r2 | 0s /  | 2 | 3,326,485 – 3,362,196 | 0 | 0 | 0 | 49 – 52 / 49 – 52 / 0 / 26 – 29 | 0 | 9,431 – 10,046 / 0 |
| main | – | 3 | 5,796,574 – 6,529,411 | 0 | 0 | 0 | 3 / 3 / 0 / 1 – 2 | 0 | 58,609 – 167,510 / 0 |
| r2 | 0s /  | 3 | 6,340,833 – 6,586,206 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,065 – 1,075 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f_r2 | 1 ms | -122 – -41 ms | never | 180 – 230 ms | 221 – 302 ms | 166,323–168,106 | 1.10–1.14 / 2.40–2.46 / 3.17–3.23 | 1.20–1.23 / 2.85–2.98 / 27.39–53.76 | 35 – 64 ms |
| main | 0 – 1 ms | -115 – -53 ms | never | 276 – 967 ms | 338 – 982 ms | 289,826–326,464 | 0.55–0.60 / 1.78–2.08 / 2.78–3.17 | 0.70–0.92 / 2.27–2.98 / 3.81–7.10 | 32 – 62 ms |
| r2 | 1 ms | -67 – -27 ms | never | 19 – 30 ms | 80 – 122 ms | 317,039–329,306 | 0.54–0.55 / 1.74–1.87 / 2.72–2.78 | 0.57–0.60 / 1.84–2.14 / 3.10–3.49 | 26 – 64 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f_r2-1 | 3,326,485 | 0 | 0 | 0 | 52 / 52 / 0 / 29 | 0 | 9,431 / 0 | 1 ms | -41 ms | 2.46 / 2.98 | – | – |
| f_r2-2 | 3,362,196 | 0 | 0 | 0 | 49 / 49 / 0 / 26 | 0 | 10,046 / 0 | 1 ms | -122 ms | 2.40 / 2.85 | – | – |
| main-1 | 6,156,237 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 167,510 / 0 | 0 ms | -69 ms | 1.97 / 2.98 | – | – |
| main-2 | 5,796,574 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 58,609 / 0 | 1 ms | -115 ms | 2.08 / 2.27 | – | – |
| main-3 | 6,529,411 | 0 | 0 | 0 | 3 / 3 / 0 / 1 | 0 | 111,965 / 0 | 1 ms | -53 ms | 1.78 / 2.40 | – | – |
| r2-1 | 6,340,833 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,075 / 0 | 1 ms | -67 ms | 1.87 / 2.14 | – | – |
| r2-2 | 6,360,851 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,065 / 0 | 1 ms | -50 ms | 1.87 / 2.08 | – | – |
| r2-3 | 6,586,206 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,068 / 0 | 1 ms | -27 ms | 1.74 / 1.84 | – | – |

</details>

## No-primary window and promotion, per run

From the watcher (CLUSTER SHARDS on every node every 100 ms, on the replica every 20 ms) and the fail script's lines; times from the failure. "Discovery node": where HAProxy sent the client's long-lived discovery-endpoint connection(s).

| Scenario | Run | Old primary flagged fail | No online primary from | No-primary window | New primary online (watcher) | PROMOTE line | Client CLUSTER SHARDS in window | Discovery node |
|---|---|---|---|---|---|---|---|---|
| smart-big-kill | b2-1 | 7.43 s | 7.43 s | 300 ms | 7.71 s | never | 0 | 7102 |
| smart-big-kill | b2-2 | 7.18 s | 7.18 s | 382 ms | 7.52 s | never | 311 | 7102 |
| smart-big-kill | lmain-1 | 8.48 s | 8.48 s | 400 ms | 8.88 s | never | 0 | 7102 |
| smart-big-pause | b2-1 | 7.41 s | 7.41 s | 201 ms | 7.62 s | never | 0 | 7102 |
| smart-big-pause | lmain-1 | 7.30 s | 7.30 s | 300 ms | 7.58 s | never | 0 | 7102 |
| smart-elect5 | main-1 | 6.71 s | 6.71 s | 5.06 s | 11.77 s | 11.75 s | 5,543 | 7102 |
| smart-elect5 | r2-1 | 6.61 s | 6.61 s | 5.04 s | 11.65 s | 11.64 s | 6,616 | 7102 |
| smart-fleet | f_main-1 | 7.60 s | 7.60 s | 379 ms | 7.94 s | never | 2,650 | 7101+7101+7101+7102+7103+7106+7103+7106+7102+7104+7102+7104+7106+7105+7106+7102+7103+7105+7105+7105 |
| smart-fleet | f_main-2 | 6.19 s | 6.19 s | 360 ms | 6.51 s | never | 2,183 | 7101+7101+7101+7104+7105+7104+7105+7102+7104+7103+7105+7105+7104+7102+7102+7106+7103+7106+7106+7106 |
| smart-fleet | f_r2-1 | 7.79 s | 7.79 s | 300 ms | 8.09 s | never | 2,346 | 7101+7101+7101+7102+7105+7106+7104+7103+7105+7103+7102+7106+7104+7104+7103+7105+7102+7104+7106+7103+7102+7106 |
| smart-fleet | f_r2-2 | 7.37 s | 7.37 s | 323 ms | 7.69 s | never | 2,870 | 7101+7101+7101+7105+7106+7102+7106+7102+7104+7104+7104+7104 |
| smart-hold3 | main-1 | 6.68 s | 6.68 s | 3.06 s | 9.74 s | 9.72 s | 4,045 | 7102 |
| smart-hold3 | main-2 | 7.02 s | 7.02 s | 3.04 s | 10.06 s | 10.05 s | 3,536 | 7102 |
| smart-hold3 | main-3 | 7.67 s | 7.67 s | 3.06 s | 10.73 s | 10.72 s | 3,815 | 7102 |
| smart-hold3 | r2-1 | 6.87 s | 6.87 s | 3.08 s | 9.93 s | 9.92 s | 3,339 | 7102 |
| smart-hold3 | r2-2 | 8.15 s | 8.15 s | 3.06 s | 11.21 s | 11.20 s | 3,615 | 7102 |
| smart-hold3 | r2-3 | 6.89 s | 6.89 s | 3.02 s | 9.91 s | 9.90 s | 4,022 | 7102 |
| smart-idle-kill | main-1 | 7.16 s | 7.16 s | 280 ms | 7.44 s | never | 0 | 7102 |
| smart-idle-kill | main-2 | 6.99 s | 6.99 s | 280 ms | 7.27 s | never | 0 | 7102 |
| smart-idle-kill | r2-1 | 6.13 s | 6.13 s | 220 ms | 6.35 s | never | 0 | 7102 |
| smart-idle-kill | r2-2 | 7.09 s | 7.09 s | 421 ms | 7.51 s | never | 0 | 7102 |
| smart-kill | main-1 | 7.38 s | 7.38 s | 400 ms | 7.78 s | never | 0 | 7102 |
| smart-kill | main-2 | 7.37 s | 7.37 s | 379 ms | 7.75 s | never | 0 | 7102 |
| smart-kill | main-3 | 7.48 s | 7.48 s | 300 ms | 7.78 s | never | 29 | 7102 |
| smart-kill | r2-1 | 6.81 s | 6.81 s | 320 ms | 7.13 s | never | 1 | 7102 |
| smart-kill | r2-2 | 7.61 s | 7.61 s | 319 ms | 7.92 s | never | 279 | 7102 |
| smart-kill | r2-3 | 6.68 s | 6.68 s | 320 ms | 7.00 s | never | 508 | 7102 |
| smart-pause | main-1 | 7.40 s | 7.40 s | 300 ms | 7.70 s | never | 0 | 7102 |
| smart-pause | main-2 | 8.00 s | 8.00 s | 420 ms | 8.42 s | never | 0 | 7102 |
| smart-pause | main-3 | 7.26 s | 7.26 s | 280 ms | 7.54 s | never | 0 | 7102 |
| smart-pause | r2-1 | 5.77 s | 5.77 s | 316 ms | 6.08 s | never | 0 | 7102 |
| smart-pause | r2-2 | 7.28 s | 7.28 s | 420 ms | 7.66 s | never | 0 | 7102 |
| smart-pause | r2-3 | 7.27 s | 7.27 s | 340 ms | 7.59 s | never | 0 | 7102 |
| smart-planned | f_r2-1 | never | never | never | 41 ms | 14 ms | 0 | 7105+7103+7103+7103+7102+7104+7102+7105+7102+7104+7104+7102+7106+7101+7101+7106+7101 |
| smart-planned | f_r2-2 | never | never | never | 123 ms | 12 ms | 0 | 7106+7102+7103+7103+7103+7106+7102+7104+7106+7103+7105+7105+7102+7105+7102+7104+7101+7104+7101+7101 |
| smart-planned | main-1 | never | never | never | 70 ms | 12 ms | 0 | 7102 |
| smart-planned | main-2 | never | never | never | 115 ms | 13 ms | 0 | 7102 |
| smart-planned | main-3 | never | never | never | 54 ms | 14 ms | 0 | 7102 |
| smart-planned | r2-1 | never | never | never | 68 ms | 16 ms | 0 | 7102 |
| smart-planned | r2-2 | never | never | never | 50 ms | 13 ms | 0 | 7102 |
| smart-planned | r2-3 | never | never | never | 28 ms | 10 ms | 0 | 7102 |

