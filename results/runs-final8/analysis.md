
<!-- steady -->

## steady

| Build | Runs | Errors | CLUSTER SHARDS during the run | ops/s, median (range) | p50 / p99 / p999, median (ms) |
|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 1 | 0 | 1 | 262,793 (262,793) | 0.60 / 3.30 / 12.16 |

## smart-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 5 | 0 of 5 | 0 | 5 (1 – 7) | 1,704 (1,429 – 1,799) | 35 (32 – 38) | 6 (5 – 7) | 7 – 8 | 3.9 (3.9 – 4.0) | 7,540 (5,930 – 7,938) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 71 (50 – 84; 45 – 90) | 45 – 90 | 7.32 | 1,991 | 200,339 | 3.36 / 12.16 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s8: 78, 71, 90, 54, 45

</details>

## smart-hold3

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 1 | 0 of 1 | 0 | 264 (264) | 1,971 (1,971) | 80 (80) | 47 (47) | 19 | 4.0 (4.0) | 8,553 (8,553) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 73 (73) | - | 10.95 | 1,962 | 173,219 | 2.66 / 11.90 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s8: 73

</details>

## smart-lazy

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 1 | 0 of 1 | 0 | 229 (229) | 2,041 (2,041) | 87 (87) | 46 (46) | 18 | 4.0 (4.0) | 8,438 (8,438) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 148 (148) | - | 11.06 | 2,002 | 140,588 | 3.68 / 12.16 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s8: 148

</details>

## smart-elect5

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 1 | 0 of 1 | 0 | 437 (437) | 1,716 (1,716) | 92 (92) | 62 (62) | 19 | 4.0 (4.0) | 7,126 (7,126) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 77 (77) | - | 11.96 | 1,998 | 163,533 | 3.30 / 12.16 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s8: 77

</details>

## smart-fleet

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 20 clients: final commit, nothing set | 1 | 0 of 1 | 0 | 2 (2) | 1,967 (1,967) | 749 (749) | 97 (97) | 122 | 78.1 (78.1) | 20,100 (20,100) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 20 clients: final commit, nothing set | 3 (3) | - | 8.16 | 1,972 | 96,246 | 7.23 / 13.18 |

<details><summary>Per run: promotion → first success (ms)</summary>

- f_s8: 3

</details>

<!-- smart-planned -->

## smart-planned

| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Traffic to the old primary after the promotion, median (range), ms | Its last MOVED after the promotion, median (ms) | CLUSTER SHARDS after failover | ops/s, median |
|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 2 | 0 | 1,060 (1,046 – 1,073) | 0 (0) | 32 | 2 (2) | 255,924 |
| 20 clients: final commit, nothing set | 2 | 0 | 9,773 (9,103 – 10,443) | 188 (186 – 189) | 224 | 58 (56 – 59) | 147,260 |

## smart-pause

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 2 | 0 of 2 | 0 | 1,000 (1,000) | 0 (0) | 8 (8) | 0 (0) | 1 | 0.6 (0.6) | 1 (1) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 2789 (2771 – 2808) | - | 9.98 | 25 | 180,889 | 3.33 / 12.03 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s8: 2771, 2808

</details>

## smart-migrate

Atomic Slot Migration of slots 0-199 (all keys there) to another primary.

| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Client traffic to the source after the handover, median (range), ms | Last MOVED from the source after the handover, median (ms) | CLUSTER SHARDS during the run | Migration errors |
|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 1 | 0 | 315 (315) | 22 (22) | 35 | 2 (2) | 0 |

## smart-idle-kill

The primary of the keys' shard crashes 1 s into an 18 s idle window and stays down past the resume. No ShardsRefreshInterval, so the client resumes with the dead primary in its slot map.

| Build | Runs | Failed after resume, median (range) | ErrNoSlot / deadline after resume | First success after resume, median (range), ms | Last error after resume, median (range), ms | MOVED in the first 5 s, median (range) | CLUSTER SHARDS in the first 5 s | Longest success, median (ms) |
|---|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 1 | 106 (106) | 0 / 0 | 20 (20) | 1.7 (1.7) | 0 (0) | 2 (2) | 35 |

<details><summary>Per run</summary>

- s8-1: failed 106, first success 20.282 ms, last error 1.744 ms, MOVED 0, errors [('dial tcp 127.0.0.1:7101: connect: connection refused', 106)]

</details>

## smart-pipe-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 1 | 0 of 1 | 0 | 8,380 (8,380) | 6,107 (6,107) | 25 (25) | 0 (0) | 4 | 3.1 (3.1) | 1,802 (1,802) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final commit, nothing set (MaxRefreshDelay 250 ms by default) | 420 (420) | - | 7.78 | 2,000 | 428,711 | 13.70 / 22.27 |

<details><summary>Per run: promotion → first success (ms)</summary>

- s8: 420

</details>
