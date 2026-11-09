Tables made by `bench/bin/report runs-holistic`. Ranges are min – max over the runs of each build.

## Runs

| Scenario | Runs | OK | Not OK | Wall time |
|---|---|---|---|---|
| smart-planned | 9 | 9 | – | 4.9 min |
| smart-pause | 5 | 5 | – | 3.5 min |
| smart-big-pause | 5 | 5 | – | 5.3 min |
| smart-big-steady | 2 | 2 | – | 1.1 min |
| smart-kill | 2 | 2 | – | 1.3 min |
| smart-big-kill | 1 | 1 | – | 1.0 min |
| steady | 1 | 1 | – | 0.5 min |

## steady

no failure (steady state); ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s7 | 0s / 0s | 1 | 6,166,976 | 0 | 0 | 0 | 1 / 1 | 0 | 0 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Longest success |
|---|---|---|---|
| s7 | 308,339 | 0.54 / 2.59 / 11.65 | 23 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s7-1 | 6,166,976 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 2.59 / 0.00 | – | – |

</details>

## smart-big-kill

`MODE=kill ./fail-primary.sh 30` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| b7 | 0s / 0s | 1 | 13,204,720 | 0 | 0 | 2,044 | 12 / 12 / 0 / 1 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| b7 | 8.31 s | +56 ms | 8.31 s | 0 ms | never | 264,092 | 0.52 / 2.66 / 11.90 | 0.45 / 2.21 / 10.62 | 1.95 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| b7-1 | 13,204,720 | 0 | 0 | 2,044 | 12 / 12 / 0 / 1 | 0 | 0 / 0 | 8.31 s | +56 ms | 2.66 / 2.21 | – | 1,940 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-big-pause

`MODE=pause ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| b7 | 0s / 0s | 2 | 9,373,744 – 9,596,807 | 0 | 1,193 – 1,301 | 99 – 106 | 5 – 6 / 5 – 6 / 0 / 1 | 0 | 0 / 0 |
| b7d1s | 0s / 0s | 2 | 7,372,338 – 12,186,515 | 0 | 800 – 1,000 | 0 | 9 – 10 / 9 – 10 / 0 – 1 / 1 | 0 | 0 / 0 |
| b7d250 | 0s / 0s | 1 | 11,014,648 | 0 | 1,000 | 0 | 11 / 11 / 1 / 1 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| b7 | 13.87 – 13.88 s | +6.58 – +7.10 s | 13.87 s | 2.70 – 3.21 s | never | 187,473–191,933 | 0.58 / 3.36–3.49 / 12.16–12.67 | 0.50–0.52 / 2.72–3.30 / 12.67 | 0.03 – 2.00 s |
| b7d1s | 7.99 – 9.98 s | +1.01 – +1.98 s | 7.99 – 9.99 s | 0 ms | never | 147,446–243,724 | 0.54–0.86 / 3.17–4.93 / 12.16–12.67 | 0.49–0.74 / 3.10–3.87 / 11.39–12.42 | 30 ms |
| b7d250 | 9.98 s | +1977 ms | 9.98 s | 0 ms | never | 220,291 | 0.57 / 3.42 / 12.16 | 0.52 / 3.10 / 11.90 | 26 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| b7-1 | 9,373,744 | 0 | 1,193 | 106 | 5 / 5 / 0 / 1 | 0 | 0 / 0 | 13.87 s | +7101 ms | 3.49 / 3.30 | – | 106 × i/o timeout |
| b7-2 | 9,596,807 | 0 | 1,301 | 99 | 6 / 6 / 0 / 1 | 0 | 0 / 0 | 13.88 s | +6583 ms | 3.36 / 2.72 | – | 99 × i/o timeout |
| b7d1s-1 | 7,372,338 | 0 | 1,000 | 0 | 10 / 10 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +1982 ms | 4.93 / 3.87 | – | – |
| b7d1s-2 | 12,186,515 | 0 | 800 | 0 | 9 / 9 / 1 / 1 | 0 | 0 / 0 | 7.99 s | +1008 ms | 3.17 / 3.10 | – | – |
| b7d250-1 | 11,014,648 | 0 | 1,000 | 0 | 11 / 11 / 1 / 1 | 0 | 0 / 0 | 9.98 s | +1977 ms | 3.42 / 3.10 | – | – |

</details>

## smart-big-steady

no failure (steady state); ShardsRefreshInterval 1s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| b7 | 0s / 0s, ShardsRefresh 1s | 1 | 3,481,571 | 0 | 0 | 0 | 19 / 1 | 0 | 0 / 0 |
| lmain | ShardsRefresh 1s | 1 | 3,561,555 | 0 | 0 | 0 | 1 / 1 | 0 | 0 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Longest success |
|---|---|---|---|
| b7 | 174,076 | 0.89 / 4.93 / 13.70 | 25 ms |
| lmain | 178,072 | 0.89 / 4.93 / 13.18 | 28 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| b7-1 | 3,481,571 | 0 | 0 | 0 | 19 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 4.93 / 0.00 | – | – |
| lmain-1 | 3,561,555 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 4.93 / 0.00 | – | – |

</details>

## smart-kill

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s7 | 0s / 0s | 1 | 3,191,371 | 0 | 1 | 1,792 | 36 / 35 / 4 / 8 | 0 | 0 / 0 |
| s7d50 | 0s / 0s | 1 | 3,125,065 | 0 | 1 | 1,818 | 36 / 35 / 4 / 7 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s7 | 7.67 s | +51 ms | 7.66 s | 0 ms | never | 127,650 | 0.82 / 4.67 / 13.18 | 0.68 / 3.42 / 13.44 | 1.99 s |
| s7d50 | 7.68 s | +40 ms | 7.63 s | 0 ms | never | 124,999 | 0.84 / 5.06 / 13.18 | 0.78 / 4.80 / 430.08 | 1.99 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s7-1 | 3,191,371 | 0 | 1 | 1,792 | 36 / 35 / 4 / 8 | 0 | 0 / 0 | 7.67 s | +51 ms | 4.67 / 3.42 | – | 1,696 × dial tcp 127.0.0.1:7101: connect: connection refused |
| s7d50-1 | 3,125,065 | 0 | 1 | 1,818 | 36 / 35 / 4 / 7 | 0 | 0 / 0 | 7.68 s | +40 ms | 5.06 / 4.80 | – | 1,708 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-pause

`MODE=pause ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| s7 | 0s / 0s | 2 | 5,468,443 – 5,634,448 | 0 | 800 – 1,000 | 0 | 6 – 7 / 6 – 7 / 0 / 1 | 0 | 0 / 0 |
| s7d250 | 0s / 0s | 2 | 5,479,778 – 5,521,121 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 |
| s7d50 | 0s / 0s | 1 | 5,642,971 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| s7 | 7.97 – 9.99 s | +1.44 – +2.35 s | 7.97 – 9.99 s | 0 – 348 ms | never | 182,265–187,812 | 0.57–0.60 / 3.30–3.42 / 12.16–12.42 | 0.57–0.63 / 3.49–3.62 / 11.65–12.42 | 26 ms |
| s7d250 | 9.98 – 9.99 s | +2.51 – +2.59 s | 9.98 – 9.99 s | 517 – 598 ms | never | 182,651–184,028 | 0.57 / 3.30–3.36 / 11.90–12.16 | 0.54–0.55 / 3.04–3.36 / 12.16–12.93 | 17 – 27 ms |
| s7d50 | 9.98 s | +2842 ms | 9.98 s | 847 ms | never | 188,092 | 0.55 / 3.23 / 11.90 | 0.51 / 2.98 / 8.58 | 36 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| s7-1 | 5,468,443 | 0 | 1,000 | 0 | 7 / 7 / 0 / 1 | 0 | 0 / 0 | 9.99 s | +2348 ms | 3.30 / 3.49 | – | – |
| s7-2 | 5,634,448 | 0 | 800 | 0 | 6 / 6 / 0 / 1 | 0 | 0 / 0 | 7.97 s | +1443 ms | 3.42 / 3.62 | – | – |
| s7d250-1 | 5,521,121 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 | 9.99 s | +2591 ms | 3.30 / 3.36 | – | – |
| s7d250-2 | 5,479,778 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +2511 ms | 3.36 / 3.04 | – | – |
| s7d50-1 | 5,642,971 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +2842 ms | 3.23 / 2.98 | – | – |

</details>

## smart-planned

planned failover (`CLUSTER FAILOVER` on the replica of slot 0's primary) at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f_s7 | 0s / 0s | 1 | 1,897,497 | 0 | 0 | 0 | 53 / 53 / 0 / 19 | 0 | 13,625 / 0 |
| f_s7d250 | 0s / 0s | 1 | 1,738,800 | 0 | 0 | 0 | 56 / 56 / 0 / 31 | 0 | 4,717 / 0 |
| f_s7d50 | 0s / 0s | 1 | 1,822,679 | 0 | 0 | 0 | 70 / 70 / 0 / 33 | 0 | 1,401 / 0 |
| s7 | 0s / 0s | 2 | 3,560,612 – 5,183,074 | 0 | 0 | 0 | 2 – 3 / 2 – 3 / 0 / 1 | 0 | 1,069 – 1,071 / 0 |
| s7d250 | 0s / 0s | 2 | 4,967,105 – 5,305,865 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,052 – 1,087 / 0 |
| s7d50 | 0s / 0s | 2 | 4,165,299 – 5,272,390 | 0 | 0 | 0 | 2 – 3 / 2 – 3 / 0 / 1 – 2 | 0 | 717 – 1,056 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f_s7 | 2 ms | -116 ms | never | 880 ms | 1.00 s | 94,873 | 1.84 / 10.37 / 15.74 | 2.14 / 13.70 / 68.61 | 76 ms |
| f_s7d250 | 2 ms | -108 ms | never | 199 ms | 202 ms | 86,937 | 2.00 / 10.62 / 14.98 | 2.27 / 11.90 / 65.02 | 72 ms |
| f_s7d50 | 1 ms | -130 ms | never | 0 ms | 100 ms | 91,130 | 1.90 / 10.37 / 16.00 | 2.21 / 13.18 / 72.70 | 88 ms |
| s7 | 0 ms | -123 – -112 ms | never | 0 – 102 ms | 20 – 202 ms | 178,027–259,145 | 0.60–0.89 / 3.49–4.67 / 12.16–13.95 | 0.63–0.95 / 4.06–4.93 / 12.67–137.22 | 60 – 138 ms |
| s7d250 | 0 ms | -108 – -52 ms | never | 0 – 29 ms | 42 – 82 ms | 248,342–265,275 | 0.58–0.62 / 3.36–3.55 / 12.16–12.42 | 0.62–0.63 / 3.74–4.16 / 11.90–13.18 | 48 – 55 ms |
| s7d50 | 0 – 1 ms | -123 – -121 ms | never | 0 – 39 ms | 26 – 101 ms | 208,256–263,613 | 0.58–0.70 / 3.36–4.42 / 12.16–12.42 | 0.62–0.95 / 3.36–6.46 / 12.42–111.62 | 57 – 113 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f_s7-1 | 1,897,497 | 0 | 0 | 0 | 53 / 53 / 0 / 19 | 0 | 13,625 / 0 | 2 ms | -116 ms | 10.37 / 13.70 | – | – |
| f_s7d250-1 | 1,738,800 | 0 | 0 | 0 | 56 / 56 / 0 / 31 | 0 | 4,717 / 0 | 2 ms | -108 ms | 10.62 / 11.90 | – | – |
| f_s7d50-1 | 1,822,679 | 0 | 0 | 0 | 70 / 70 / 0 / 33 | 0 | 1,401 / 0 | 1 ms | -130 ms | 10.37 / 13.18 | – | – |
| s7-1 | 5,183,074 | 0 | 0 | 0 | 3 / 3 / 0 / 1 | 0 | 1,069 / 0 | 0 ms | -123 ms | 3.49 / 4.06 | – | – |
| s7-2 | 3,560,612 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,071 / 0 | 0 ms | -112 ms | 4.67 / 4.93 | – | – |
| s7d250-1 | 5,305,865 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,087 / 0 | 0 ms | -108 ms | 3.36 / 4.16 | – | – |
| s7d250-2 | 4,967,105 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,052 / 0 | 0 ms | -52 ms | 3.55 / 3.74 | – | – |
| s7d50-1 | 5,272,390 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 1,056 / 0 | 0 ms | -123 ms | 3.36 / 3.36 | – | – |
| s7d50-2 | 4,165,299 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 717 / 0 | 1 ms | -121 ms | 4.42 / 6.46 | – | – |

</details>

## No-primary window and promotion, per run

From the watcher (CLUSTER SHARDS on every node every 100 ms, on the replica every 20 ms) and the fail script's lines; times from the failure. "Discovery node": where HAProxy sent the client's long-lived discovery-endpoint connection(s).

| Scenario | Run | Old primary flagged fail | No online primary from | No-primary window | New primary online (watcher) | PROMOTE line | Client CLUSTER SHARDS in window | Discovery node |
|---|---|---|---|---|---|---|---|---|
| smart-big-kill | b7-1 | 7.94 s | 7.94 s | 320 ms | 8.26 s | never | 0 | 7102 |
| smart-big-pause | b7-1 | 6.40 s | 6.40 s | 375 ms | 6.77 s | never | 0 | 7102 |
| smart-big-pause | b7-2 | 6.89 s | 6.89 s | 400 ms | 7.29 s | never | 0 | 7102 |
| smart-big-pause | b7d1s-1 | 7.68 s | 7.68 s | 344 ms | 8.00 s | never | 0 | 7102 |
| smart-big-pause | b7d1s-2 | 6.70 s | 6.70 s | 280 ms | 6.98 s | never | 1 | 7102 |
| smart-big-pause | b7d250-1 | 7.78 s | 7.78 s | 220 ms | 8.00 s | never | 1 | 7102 |
| smart-kill | s7-1 | 7.31 s | 7.31 s | 300 ms | 7.61 s | never | 4 | 7102 |
| smart-kill | s7d50-1 | 7.40 s | 7.40 s | 240 ms | 7.64 s | never | 4 | 7102 |
| smart-pause | s7-1 | 7.34 s | 7.34 s | 300 ms | 7.64 s | never | 0 | 7102 |
| smart-pause | s7-2 | 6.29 s | 6.29 s | 260 ms | 6.53 s | never | 0 | 7102 |
| smart-pause | s7d250-1 | 7.04 s | 7.04 s | 360 ms | 7.40 s | never | 0 | 7102 |
| smart-pause | s7d250-2 | 7.12 s | 7.12 s | 361 ms | 7.47 s | never | 0 | 7102 |
| smart-pause | s7d50-1 | 6.73 s | 6.73 s | 400 ms | 7.13 s | never | 0 | 7102 |
| smart-planned | f_s7-1 | never | never | never | 118 ms | 21 ms | 0 | 7102+7103+7105+7104+7102+7103+7104+7103+7105+7102+7105+7103+7104+7102+7101+7106+7101 |
| smart-planned | f_s7d250-1 | never | never | never | 109 ms | 25 ms | 0 | 7103+7102+7103+7103+7106+7101+7105+7104+7101+7102+7102+7106+7101+7103+7105 |
| smart-planned | f_s7d50-1 | never | never | never | 130 ms | 17 ms | 0 | 7103+7102+7106+7106+7105+7105+7105+7102+7103+7102+7103+7103+7104+7106+7102+7104+7101+7101+7101+7104 |
| smart-planned | s7-1 | never | never | never | 124 ms | 9 ms | 0 | 7102 |
| smart-planned | s7-2 | never | never | never | 112 ms | 20 ms | 0 | 7102 |
| smart-planned | s7d250-1 | never | never | never | 108 ms | 12 ms | 0 | 7102 |
| smart-planned | s7d250-2 | never | never | never | 53 ms | 11 ms | 0 | 7102 |
| smart-planned | s7d50-1 | never | never | never | 124 ms | 12 ms | 0 | 7102 |
| smart-planned | s7d50-2 | never | never | never | 122 ms | 19 ms | 0 | 7102 |

