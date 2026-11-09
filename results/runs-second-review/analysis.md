
<!-- steady -->

## steady

| Build | Runs | Errors | CLUSTER SHARDS during the run | ops/s, median (range) | p50 / p99 / p999, median (ms) |
|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 2 | 0 | 1 | 325,023 (319,939 – 330,106) | 0.55 / 1.79 / 2.72 |
| Main today | 2 | 0 | 1 | 314,889 (301,222 – 328,557) | 0.57 / 1.86 / 2.75 |

## smart-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 3 | 2 of 3 | 0 – 44,379 | 0 (0 – 1) | 1,694 (1,669 – 1,905) | 358 (30 – 585) | 279 (1 – 508) | 4 – 409 | 4.0 (3.9 – 6.4) | 7,930 (7,727 – 8,780) |
| Main today | 3 | 1 of 3 | 0 – 1,448 | 0 (0) | 1,918 (1,890 – 1,950) | 18 (17 – 78) | 0 (0 – 29) | 3 – 61 | 2.1 (2.0 – 2.2) | 8,932 (8,768 – 9,103) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 40 (32 – 104) | - | 7.24 | 1,922 | 229,403 | 1.74 / 2.72 |
| Main today | 31 (30 – 94) | - | 7.81 | 1,911 | 227,496 | 1.81 / 2.66 |

Promotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):

| Comparison | Shift (ms) | p |
|---|---|---|
| Main today vs PR after the second review: no deadline refresh, no slot wait | -8 | 0.383 |

<details><summary>Per run: promotion → first success (ms)</summary>

- r2: 104, 40, 32
- main: 94, 31, 30

</details>

## smart-hold3

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 3 | 3 of 3 | 617,774 – 697,476 | 0 (0 – 1) | 1,707 (1,692 – 2,063) | 3,691 (3,379 – 4,083) | 3,615 (3,339 – 4,022) | 1,247 – 1,344 | 3.9 (3.9) | 7,955 (7,868 – 9,457) |
| Main today | 3 | 3 of 3 | 607,370 – 726,784 | 0 (0 – 1) | 1,797 (1,683 – 1,919) | 3,867 (3,586 – 4,083) | 3,815 (3,536 – 4,045) | 1,342 – 1,386 | 2.0 (1.5 – 2.3) | 8,418 (7,549 – 8,979) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 36 (36 – 39) | - | 9.97 | 11 | 200,945 | 1.81 / 2.78 |
| Main today | 38 (31 – 38) | - | 10.10 | 10 | 195,303 | 1.84 / 2.78 |

Promotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):

| Comparison | Shift (ms) | p |
|---|---|---|
| Main today vs PR after the second review: no deadline refresh, no slot wait | -1 | 1 |

<details><summary>Per run: promotion → first success (ms)</summary>

- r2: 36, 39, 36
- main: 31, 38, 38

</details>

## smart-elect5

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 1 | 1 of 1 | 1,237,295 | 0 (0) | 1,691 (1,691) | 6,684 (6,684) | 6,616 (6,616) | 1,412 | 4.1 (4.1) | 7,811 (7,811) |
| Main today | 1 | 1 of 1 | 1,093,099 | 23 (23) | 1,694 (1,694) | 5,579 (5,579) | 5,543 (5,543) | 1,407 | 1.8 (1.8) | 7,611 (7,611) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 40 (40) | - | 11.69 | 10 | 205,523 | 1.81 / 2.78 |
| Main today | 28 (28) | - | 11.80 | 10 | 208,460 | 1.81 / 2.78 |

<details><summary>Per run: promotion → first success (ms)</summary>

- r2: 40
- main: 28

</details>

## smart-fleet

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 20 clients: PR after the second review | 2 | 2 of 2 | 11,210 – 12,384 | 0 (0 – 1) | 1,904 (1,803 – 2,006) | 3,515 (3,355 – 3,675) | 2,608 (2,346 – 2,870) | 2,728 – 3,123 | 78.3 (78.1 – 78.5) | 19,623 (18,632 – 20,614) |
| 20 clients: main | 2 | 2 of 2 | 8,610 – 12,454 | 0 (0) | 1,786 (1,594 – 1,977) | 2,704 (2,444 – 2,963) | 2,416 (2,183 – 2,650) | 2,206 – 2,540 | 37.2 (37.0 – 37.5) | 18,475 (16,558 – 20,392) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 20 clients: PR after the second review | -0 (-4 – 4) | - | 7.89 | 1,923 | 112,732 | 2.46 / 3.58 |
| 20 clients: main | -3 (-5 – -2) | - | 7.22 | 1,942 | 117,680 | 2.43 / 3.30 |

<details><summary>Per run: promotion → first success (ms)</summary>

- f_r2: 4, -4
- f_main: -2, -5

</details>

<!-- smart-planned -->

## smart-planned

| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Traffic to the old primary after the promotion, median (range), ms | Its last MOVED after the promotion, median (ms) | CLUSTER SHARDS after failover | ops/s, median |
|---|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 3 | 0 | 1,068 (1,065 – 1,075) | 29 (19 – 30) | 101 | 2 (2) | 318,031 |
| 20 clients: PR after the second review | 2 | 0 | 9,738 (9,431 – 10,046) | 205 (180 – 230) | 262 | 50 (49 – 52) | 167,215 |
| Main today | 3 | 0 | 111,965 (58,609 – 167,510) | 590 (276 – 967) | 601 | 3 (3) | 307,806 |

## smart-big-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 1,000 nodes: PR after the second review | 2 | 1 of 2 | 0 – 13,688 | 0 (0) | 1,916 (1,756 – 2,075) | 167 (13 – 321) | 156 (0 – 311) | 1 – 311 | 1.0 (1.0 – 1.1) | 8,983 (8,262 – 9,704) |
| 1,000+ nodes: main (delay up to 30 s) | 1 | 0 of 1 | 0 | 0 (0) | 4,527 (4,527) | 1 (1) | 0 (0) | 1 | 0.0 (0.0) | 22,739 (22,739) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 1,000 nodes: PR after the second review | 352 (31 – 673) | - | 7.97 | 1,937 | 296,921 | 1.70 / 2.56 |
| 1,000+ nodes: main (delay up to 30 s) | 9235 (9235) | - | 18.12 | 1,978 | 225,078 | 1.68 / 2.40 |

<details><summary>Per run: promotion → first success (ms)</summary>

- b2: 673, 31
- lmain: 9235

</details>

## smart-pause

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 3 | 0 of 3 | 0 | 1,098 (1,096 – 1,107) | 102 (93 – 104) | 4 (4) | 0 (0) | 2 | 0.1 (0.1 – 0.2) | 2 (2 – 3) |
| Main today | 3 | 0 of 3 | 0 | 1,104 (1,097 – 1,110) | 96 (90 – 103) | 4 (4) | 0 (0) | 1 – 2 | 0.1 (0.1) | 3 (3) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 4502 (4331 – 5898) | - | 11.99 | 117 | 191,627 | 1.87 / 2.78 |
| Main today | 4382 (3677 – 4553) | - | 12.09 | 215 | 199,125 | 1.78 / 2.66 |

Promotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):

| Comparison | Shift (ms) | p |
|---|---|---|
| Main today vs PR after the second review: no deadline refresh, no slot wait | -654 | 0.663 |

<details><summary>Per run: promotion → first success (ms)</summary>

- r2: 5898, 4331, 4502
- main: 4382, 3677, 4553

</details>

## smart-migrate

Atomic Slot Migration of slots 0-199 (all keys there) to another primary.

| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Client traffic to the source after the handover, median (range), ms | Last MOVED from the source after the handover, median (ms) | CLUSTER SHARDS during the run | Migration errors |
|---|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 2 | 0 | 256 (200 – 313) | 18 (18) | 23 | 2 (2) | 0 |
| Main today | 2 | 0 | 66,728 (20,999 – 112,456) | 340 (129 – 552) | 338 | 3 (3) | 0 |

## smart-idle-kill

The primary of the keys' shard crashes 1 s into an 18 s idle window and stays down past the resume. No ShardsRefreshInterval, so the client resumes with the dead primary in its slot map.

| Build | Runs | Failed after resume, median (range) | ErrNoSlot / deadline after resume | First success after resume, median (range), ms | Last error after resume, median (range), ms | MOVED in the first 5 s, median (range) | CLUSTER SHARDS in the first 5 s | Longest success, median (ms) |
|---|---|---|---|---|---|---|---|---|
| PR after the second review: no deadline refresh, no slot wait | 2 | 98 (97 – 98) | 0 / 0 | 77 (48 – 106) | 2.9 (1.5 – 4.3) | 0 (0) | 2 (2) | 157 |
| Main today | 2 | 106 (101 – 110) | 0 / 0 | 467 (111 – 823) | 2.9 (2.7 – 3.0) | 0 (0) | 2 (2) | 1,018 |

<details><summary>Per run</summary>

- r2-1: failed 97, first success 47.734 ms, last error 4.343 ms, MOVED 0, errors [('dial tcp 127.0.0.1:7101: connect: connection refused', 97)]
- r2-2: failed 98, first success 106.115 ms, last error 1.483 ms, MOVED 0, errors [('dial tcp 127.0.0.1:7101: connect: connection refused', 98)]
- main-1: failed 101, first success 110.994 ms, last error 3.026 ms, MOVED 0, errors [('dial tcp 127.0.0.1:7101: connect: connection refused', 101)]
- main-2: failed 110, first success 823.262 ms, last error 2.702 ms, MOVED 0, errors [('dial tcp 127.0.0.1:7101: connect: connection refused', 110)]

</details>

## smart-big-pause

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 1,000 nodes: PR after the second review | 1 | 0 of 1 | 0 | 1,000 (1,000) | 0 (0) | 5 (5) | 0 (0) | 1 | 0.0 (0.0) | 1 (1) |
| 1,000+ nodes: main (delay up to 30 s) | 1 | 0 of 1 | 0 | 1,200 (1,200) | 104 (104) | 2 (2) | 0 (0) | 1 | 0.0 (0.0) | 2 (2) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 1,000 nodes: PR after the second review | 2362 (2362) | - | 9.98 | 11 | 273,209 | 1.71 / 2.46 |
| 1,000+ nodes: main (delay up to 30 s) | 6300 (6300) | - | 13.88 | 8 | 245,537 | 1.71 / 2.53 |

<details><summary>Per run: promotion → first success (ms)</summary>

- b2: 2362
- lmain: 6300

</details>

<!-- smart-big-steady -->

## steady

| Build | Runs | Errors | CLUSTER SHARDS during the run | ops/s, median (range) | p50 / p99 / p999, median (ms) |
|---|---|---|---|---|---|
| 1,000 nodes: PR after the second review | 1 | 0 | 19 | 325,979 (325,979) | 0.55 / 1.74 / 2.46 |
| 1,000+ nodes: main (delay up to 30 s) | 1 | 0 | 0 | 323,734 (323,734) | 0.55 / 1.78 / 2.78 |
