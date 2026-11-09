
<!-- steady -->

## steady

| Build | Runs | Errors | CLUSTER SHARDS during the run | ops/s, median (range) | p50 / p99 / p999, median (ms) |
|---|---|---|---|---|---|
| Final: MaxRefreshDelay at its default (1s, 30s at 1,000 nodes) | 1 | 0 | 1 | 308,339 (308,339) | 0.54 / 2.59 / 11.65 |

## smart-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final: MaxRefreshDelay at its default (1s, 30s at 1,000 nodes) | 1 | 0 of 1 | 0 | 1 (1) | 1,792 (1,792) | 35 (35) | 4 (4) | 8 | 4.0 (4.0) | 7,495 (7,495) |
| Final: MaxRefreshDelay 50 ms | 1 | 0 of 1 | 0 | 1 (1) | 1,818 (1,818) | 35 (35) | 4 (4) | 7 | 3.9 (3.9) | 7,619 (7,619) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final: MaxRefreshDelay at its default (1s, 30s at 1,000 nodes) | 51 (51) | - | 7.67 | 1,985 | 127,650 | 4.67 / 13.18 |
| Final: MaxRefreshDelay 50 ms | 40 (40) | - | 7.68 | 1,988 | 124,999 | 5.06 / 13.18 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s7: 51
- s7d50: 40

</details>

<!-- smart-planned -->

## smart-planned

| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Traffic to the old primary after the promotion, median (range), ms | Its last MOVED after the promotion, median (ms) | CLUSTER SHARDS after failover | ops/s, median |
|---|---|---|---|---|---|---|---|
| Final: MaxRefreshDelay at its default (1s, 30s at 1,000 nodes) | 2 | 0 | 1,070 (1,069 – 1,071) | 51 (0 – 102) | 111 | 2 (2 – 3) | 218,586 |
| Final: MaxRefreshDelay 250 ms | 2 | 0 | 1,070 (1,052 – 1,087) | 14 (0 – 29) | 62 | 2 (2) | 256,809 |
| Final: MaxRefreshDelay 50 ms | 2 | 0 | 886 (717 – 1,056) | 19 (0 – 39) | 64 | 2 (2 – 3) | 235,934 |
| 20 clients: final, MaxRefreshDelay default | 1 | 0 | 13,625 (13,625) | 880 (880) | 1,002 | 53 (53) | 94,873 |
| 20 clients: final, MaxRefreshDelay 250 ms | 1 | 0 | 4,717 (4,717) | 199 (199) | 202 | 56 (56) | 86,937 |
| 20 clients: final, MaxRefreshDelay 50 ms | 1 | 0 | 1,401 (1,401) | 0 (0) | 100 | 70 (70) | 91,130 |

## smart-big-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 1,000 nodes: final, MaxRefreshDelay default (30 s) | 1 | 0 of 1 | 0 | 0 (0) | 2,044 (2,044) | 12 (12) | 0 (0) | 1 | 1.0 (1.0) | 9,086 (9,086) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 1,000 nodes: final, MaxRefreshDelay default (30 s) | 56 (56) | - | 8.31 | 1,950 | 264,092 | 2.66 / 11.90 |

<details><summary>Per run: promotion → first success (ms)</summary>

- b7: 56

</details>

## smart-pause

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final: MaxRefreshDelay at its default (1s, 30s at 1,000 nodes) | 2 | 0 of 2 | 0 | 900 (800 – 1,000) | 0 (0) | 6 (6 – 7) | 0 (0) | 1 | 0.4 (0.4 – 0.5) | 1 (1) |
| Final: MaxRefreshDelay 250 ms | 2 | 0 of 2 | 0 | 1,000 (1,000) | 0 (0) | 8 (8) | 0 (0) | 1 | 0.6 (0.6) | 1 (1) |
| Final: MaxRefreshDelay 50 ms | 1 | 0 of 1 | 0 | 1,000 (1,000) | 0 (0) | 8 (8) | 0 (0) | 1 | 0.6 (0.6) | 1 (1) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final: MaxRefreshDelay at its default (1s, 30s at 1,000 nodes) | 1895 (1443 – 2348) | - | 8.98 | 26 | 185,038 | 3.36 / 12.29 |
| Final: MaxRefreshDelay 250 ms | 2551 (2511 – 2591) | - | 9.98 | 22 | 183,340 | 3.33 / 12.03 |
| Final: MaxRefreshDelay 50 ms | 2842 (2842) | - | 9.98 | 36 | 188,092 | 3.23 / 11.90 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s7: 2348, 1443
- s7d250: 2591, 2511
- s7d50: 2842

</details>

## smart-big-pause

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 1,000 nodes: final, MaxRefreshDelay default (30 s) | 2 | 0 of 2 | 0 | 1,247 (1,193 – 1,301) | 102 (99 – 106) | 6 (5 – 6) | 0 (0) | 1 | 0.1 (0.0 – 0.1) | 2 (2 – 3) |
| 1,000 nodes: final, MaxRefreshDelay 1 s | 2 | 0 of 2 | 0 | 900 (800 – 1,000) | 0 (0) | 10 (9 – 10) | 0 (0 – 1) | 1 | 0.5 (0.4 – 0.5) | 1 (1) |
| 1,000 nodes: final, MaxRefreshDelay 250 ms | 1 | 0 of 1 | 0 | 1,000 (1,000) | 0 (0) | 11 (11) | 1 (1) | 1 | 0.5 (0.5) | 1 (1) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 1,000 nodes: final, MaxRefreshDelay default (30 s) | 6842 (6583 – 7101) | - | 13.87 | 1,013 | 189,703 | 3.42 / 12.42 |
| 1,000 nodes: final, MaxRefreshDelay 1 s | 1495 (1008 – 1982) | - | 8.99 | 30 | 195,585 | 4.05 / 12.42 |
| 1,000 nodes: final, MaxRefreshDelay 250 ms | 1977 (1977) | - | 9.98 | 26 | 220,291 | 3.42 / 12.16 |

<details><summary>Per run: promotion → first success (ms)</summary>

- b7: 7101, 6583
- b7d1s: 1982, 1008
- b7d250: 1977

</details>

<!-- smart-big-steady -->

## steady

| Build | Runs | Errors | CLUSTER SHARDS during the run | ops/s, median (range) | p50 / p99 / p999, median (ms) |
|---|---|---|---|---|---|
| 1,000 nodes: final, MaxRefreshDelay default (30 s) | 1 | 0 | 19 | 174,076 (174,076) | 0.89 / 4.93 / 13.70 |
| 1,000+ nodes: main (delay up to 30 s) | 1 | 0 | 1 | 178,072 (178,072) | 0.89 / 4.93 / 13.18 |
