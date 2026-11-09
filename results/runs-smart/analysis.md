
## smart-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 3 | 0 of 3 | 0 | 4 (0 – 6) | 1,686 (1,419 – 1,780) | 32 (32 – 37) | 5 (3 – 7) | 5 – 8 | 3.9 (3.9 – 4.0) | 7,390 (6,279 – 7,787) |
| Proposed defaults v3, nothing set | 2 | 0 of 2 | 0 | 1 (0 – 2) | 1,844 (1,716 – 1,972) | 36 (35 – 37) | 4 (3 – 5) | 5 – 7 | 4.0 (3.9 – 4.0) | 8,002 (7,424 – 8,579) |
| Proposed defaults v2, nothing set | 3 | 0 of 3 | 0 | 5 (1 – 5) | 1,772 (1,438 – 1,891) | 35 (31 – 37) | 5 (3 – 6) | 4 – 8 | 3.9 (3.9 – 4.0) | 7,950 (6,140 – 8,395) |
| Proposed defaults v1, nothing set | 4 | 0 of 4 | 0 | 0 (0 – 1) | 1,826 (1,380 – 2,120) | 18 (14 – 19) | 0 (0 – 1) | 2 – 4 | 2.0 (1.9 – 2.2) | 8,081 (6,045 – 9,568) |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 4 | 0 of 4 | 0 | 4 (0 – 11) | 1,812 (1,717 – 1,930) | 148 (135 – 154) | 6 (5 – 7) | 20 | 18.6 (18.5 – 18.8) | 7,955 (7,350 – 8,393) |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 4 | 0 of 4 | 0 | 10 (3 – 17) | 1,758 (1,460 – 2,062) | 75 (66 – 83) | 3 (2 – 4) | 10 | 9.7 (9.6 – 9.8) | 7,704 (6,565 – 8,819) |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 4 | 0 of 4 | 0 | 20 (4 – 29) | 1,914 (1,781 – 1,978) | 34 (33 – 35) | 2 (1 – 2) | 4 | 4.0 (3.9 – 4.0) | 8,322 (7,964 – 8,658) |
| Main today | 4 | 0 of 4 | 0 | 0 (0) | 1,922 (1,695 – 2,060) | 19 (14 – 21) | 0 (0) | 2 – 4 | 2.4 (1.7 – 2.5) | 8,470 (7,251 – 9,157) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 42 (38 – 54) | - | 7.14 | 1,992 | 197,961 | 3.23 / 12.16 |
| Proposed defaults v3, nothing set | 57 (54 – 60) | - | 7.87 | 1,945 | 193,476 | 3.23 / 11.90 |
| Proposed defaults v2, nothing set | 51 (18 – 64) | - | 7.80 | 1,972 | 193,480 | 2.98 / 11.90 |
| Proposed defaults v1, nothing set | 349 (244 – 538; 238 – 572) | - | 7.71 | 1,960 | 194,286 | 3.10 / 11.78 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 60 (47 – 80; 44 – 84) | - | 7.85 | 1,967 | 192,562 | 3.30 / 11.90 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 72 (53 – 105; 49 – 114) | - | 7.59 | 1,973 | 196,295 | 3.14 / 12.03 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 224 (85 – 249; 45 – 251) | - | 8.24 | 1,993 | 198,680 | 3.04 / 11.78 |
| Main today | 260 (38 – 462; 25 – 469) | - | 7.75 | 1,945 | 193,474 | 3.10 / 12.03 |

Promotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):

| Comparison | Shift (ms) | p |
|---|---|---|
| Fixed 50 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v1, nothing set | -285 | 0.0304 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v1, nothing set | -268 | 0.0304 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v1, nothing set | -191 | 0.112 |
| Main today vs Proposed defaults v1, nothing set | -145 | 0.665 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | +17 | 0.377 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | +29 | 0.216 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | +183 | 0.216 |
| Main today vs Proposed defaults v2, nothing set | +219 | 0.216 |
| Proposed defaults v1, nothing set vs Proposed defaults v2, nothing set | +308 | 0.0518 |
| Main today vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +214 | 0.377 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +15 | 0.112 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +26 | 0.112 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +178 | 0.112 |

<details><summary>Per run: promotion → first success (ms)</summary>

- smart6: 42, 38, 54
- smart3: 54, 60
- smart2: 51, 64, 18
- smart: 238, 263, 572, 435
- f50: 44, 84, 55, 65
- f100: 79, 66, 114, 49
- f250: 245, 203, 45, 251
- main: 469, 441, 25, 78

</details>

## smart-hold3

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 2 | 0 of 2 | 0 | 312 (308 – 315) | 1,760 (1,705 – 1,816) | 78 (78) | 48 (48) | 19 | 3.9 (3.9 – 4.0) | 7,551 (7,447 – 7,655) |
| Proposed defaults v3, nothing set | 1 | 0 of 1 | 0 | 300 (300) | 1,878 (1,878) | 80 (80) | 48 (48) | 20 | 3.9 (3.9) | 7,911 (7,911) |
| Proposed defaults v2, nothing set | 3 | 0 of 3 | 0 | 310 (228 – 312) | 1,647 (1,472 – 1,802) | 74 (74 – 79) | 48 (47 – 48) | 19 – 20 | 3.9 (3.9 – 4.0) | 6,778 (6,229 – 8,045) |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 3 | 0 of 3 | 0 | 290 (287 – 291) | 1,706 (1,530 – 1,930) | 191 (178 – 206) | 58 (58) | 20 | 18.7 (18.6 – 18.7) | 7,438 (6,581 – 8,416) |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 3 | 0 of 3 | 0 | 295 (241 – 327) | 1,718 (1,703 – 1,742) | 101 (99 – 103) | 30 (30) | 10 | 9.8 (9.8 – 9.9) | 7,526 (7,437 – 7,823) |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 3 | 0 of 3 | 0 | 268 (266 – 344) | 1,794 (1,655 – 1,963) | 44 (41 – 45) | 13 (12 – 13) | 4 | 3.9 (3.9 – 4.0) | 7,983 (7,099 – 8,694) |
| Main today | 3 | 3 of 3 | 468,984 – 555,356 | 16 (0 – 17) | 1,999 (1,769 – 2,092) | 2,850 (2,401 – 3,225) | 2,794 (2,348 – 3,183) | 1,109 – 1,199 | 1.9 (1.5 – 2.5) | 8,862 (7,796 – 9,304) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 60 (42 – 77) | - | 10.23 | 1,983 | 172,014 | 2.91 / 12.16 |
| Proposed defaults v3, nothing set | 85 (85) | - | 10.80 | 1,986 | 159,647 | 3.23 / 12.16 |
| Proposed defaults v2, nothing set | 66 (41 – 77) | - | 9.40 | 1,930 | 179,340 | 3.10 / 11.65 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 45 (31 – 59) | - | 10.12 | 1,996 | 177,104 | 3.04 / 11.65 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 78 (41 – 104) | - | 10.16 | 1,996 | 174,502 | 2.98 / 11.65 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 246 (132 – 256) | - | 10.74 | 1,998 | 163,245 | 3.10 / 11.65 |
| Main today | 38 (33 – 41) | - | 10.85 | 23 | 166,151 | 2.98 / 11.65 |

Promotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):

| Comparison | Shift (ms) | p |
|---|---|---|
| Fixed 50 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | -17 | 0.383 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | +12 | 0.663 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | +179 | 0.0809 |
| Main today vs Proposed defaults v2, nothing set | -27 | 0.0809 |

<details><summary>Per run: promotion → first success (ms)</summary>

- smart6: 42, 77
- smart3: 85
- smart2: 41, 77, 66
- f50: 59, 31, 45
- f100: 41, 78, 104
- f250: 132, 246, 256
- main: 38, 41, 33

</details>

## smart-lazy

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 3 | 0 of 3 | 0 | 243 (222 – 340) | 1,669 (1,626 – 1,758) | 84 (83 – 88) | 48 (46 – 48) | 18 – 19 | 4.0 (3.9 – 4.1) | 7,062 (6,706 – 7,681) |
| Proposed defaults v3, nothing set | 3 | 0 of 3 | 0 | 305 (279 – 331) | 1,746 (1,331 – 1,953) | 85 (81 – 87) | 47 (46 – 47) | 19 | 3.9 (3.7 – 4.0) | 7,642 (5,675 – 8,726) |
| Proposed defaults v2, nothing set | 3 | 0 of 3 | 0 | 306 (238 – 332) | 1,846 (1,642 – 1,888) | 58 (48 – 61) | 22 (10 – 23) | 5 – 15 | 3.8 (3.8 – 3.9) | 7,954 (6,931 – 8,273) |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 3 | 0 of 3 | 0 | 279 (259 – 289) | 1,938 (1,605 – 1,948) | 213 (188 – 216) | 58 (58 – 59) | 20 | 18.7 (18.5 – 18.7) | 8,554 (6,904 – 8,601) |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 3 | 0 of 3 | 0 | 322 (265 – 325) | 1,714 (1,568 – 1,769) | 109 (103 – 112) | 30 (30) | 10 | 9.7 (9.7) | 7,479 (6,535 – 7,883) |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 3 | 0 of 3 | 0 | 272 (243 – 329) | 1,710 (1,696 – 2,067) | 50 (49 – 54) | 12 (12 – 13) | 4 | 4.0 (3.9 – 4.0) | 7,561 (7,540 – 8,914) |
| Main today | 3 | 3 of 3 | 432,354 – 544,561 | 0 (0 – 2) | 2,064 (1,732 – 2,123) | 2,988 (2,273 – 3,062) | 2,954 (2,226 – 3,019) | 1,066 – 1,139 | 2.4 (2.2 – 2.6) | 9,100 (7,578 – 9,602) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 34 (33 – 104) | - | 9.56 | 1,973 | 164,235 | 3.42 / 12.16 |
| Proposed defaults v3, nothing set | 110 (96 – 140) | - | 10.50 | 1,999 | 161,419 | 3.10 / 12.16 |
| Proposed defaults v2, nothing set | 85 (69 – 599) | - | 10.53 | 1,991 | 160,407 | 3.23 / 12.16 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 60 (44 – 63) | - | 10.88 | 1,990 | 166,884 | 3.17 / 11.90 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 64 (36 – 109) | - | 10.24 | 1,992 | 171,339 | 3.10 / 11.65 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 185 (184 – 254) | - | 10.30 | 1,991 | 173,047 | 3.04 / 12.16 |
| Main today | 32 (24 – 35) | - | 11.08 | 23 | 157,974 | 3.04 / 12.16 |

Promotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):

| Comparison | Shift (ms) | p |
|---|---|---|
| Fixed 50 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | -26 | 0.0809 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | -33 | 0.383 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v2, nothing set | +100 | 0.663 |
| Main today vs Proposed defaults v2, nothing set | -53 | 0.0809 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v3, nothing set | -52 | 0.0809 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v3, nothing set | -46 | 0.19 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) vs Proposed defaults v3, nothing set | +88 | 0.0809 |
| Main today vs Proposed defaults v3, nothing set | -78 | 0.0809 |
| Proposed defaults v2, nothing set vs Proposed defaults v3, nothing set | -25 | 0.663 |
| Proposed defaults v3, nothing set vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +63 | 0.19 |
| Main today vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | -9 | 0.383 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +11 | 0.663 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +6 | 0.383 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) vs Final candidate v6 (v3 + Expedite within the error jitter), nothing set | +150 | 0.0809 |

<details><summary>Per run: promotion → first success (ms)</summary>

- smart6: 34, 104, 33
- smart3: 96, 110, 140
- smart2: 69, 599, 85
- f50: 44, 63, 60
- f100: 109, 64, 36
- f250: 254, 185, 184
- main: 32, 24, 35

</details>

## smart-elect5

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 2 | 0 of 2 | 0 | 503 (465 – 541) | 1,734 (1,683 – 1,785) | 92 (91 – 94) | 62 (62 – 63) | 19 | 3.9 (3.9) | 7,631 (7,162 – 8,100) |
| Proposed defaults v3, nothing set | 2 | 0 of 2 | 0 | 506 (502 – 509) | 1,878 (1,878) | 94 (93 – 94) | 62 (61 – 62) | 19 | 3.9 (3.9) | 8,151 (8,093 – 8,209) |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 2 | 0 of 2 | 0 | 506 (494 – 518) | 1,576 (1,452 – 1,699) | 222 (211 – 234) | 96 (95 – 97) | 20 | 18.7 (18.6 – 18.9) | 6,737 (6,091 – 7,383) |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 2 | 0 of 2 | 0 | 510 (510 – 511) | 1,804 (1,732 – 1,876) | 125 (122 – 128) | 50 (49 – 50) | 10 | 9.7 (9.6 – 9.8) | 7,717 (7,455 – 7,979) |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 2 | 0 of 2 | 0 | 491 (469 – 513) | 1,697 (1,668 – 1,726) | 50 (50 – 51) | 20 (20) | 4 | 4.0 (4.0) | 7,126 (6,979 – 7,272) |
| Main today | 2 | 2 of 2 | 847,254 – 930,705 | 0 (0) | 1,987 (1,931 – 2,043) | 4,746 (4,554 – 4,937) | 4,694 (4,515 – 4,874) | 990 – 1,165 | 1.8 (1.6 – 2.0) | 8,684 (8,503 – 8,864) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 221 (221) | - | 12.30 | 1,996 | 171,698 | 3.20 / 12.16 |
| Proposed defaults v3, nothing set | 176 (151 – 200) | - | 12.90 | 1,980 | 151,472 | 3.52 / 12.54 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 73 (57 – 90) | - | 11.73 | 1,984 | 168,647 | 3.30 / 12.42 |
| Fixed 100 ms interval (delay 50 ms, wait 5 s) | 83 (57 – 109) | - | 12.65 | 1,991 | 158,713 | 3.30 / 12.16 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 212 (197 – 227) | - | 12.10 | 1,993 | 167,321 | 3.26 / 12.16 |
| Main today | 40 (36 – 44) | - | 12.55 | 26 | 160,566 | 3.33 / 12.16 |

<details><summary>Per run: promotion → first success (ms)</summary>

- smart6: 221, 221
- smart3: 151, 200
- f50: 90, 57
- f100: 109, 57
- f250: 197, 227
- main: 36, 44

</details>

## smart-fleet

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 20 clients: final candidate v6 | 2 | 0 of 2 | 0 | 2 (1 – 4) | 1,817 (1,713 – 1,921) | 710 (677 – 742) | 92 (87 – 97) | 113 – 130 | 77.9 (77.7 – 78.2) | 18,594 (17,735 – 19,454) |
| 20 clients: proposed defaults v3 | 2 | 0 of 2 | 0 | 0 (0) | 1,721 (1,707 – 1,735) | 654 (622 – 686) | 65 (44 – 86) | 81 – 141 | 77.2 (76.8 – 77.5) | 17,652 (17,331 – 17,973) |
| 20 clients: fixed 50 ms | 2 | 0 of 2 | 0 | 6 (5 – 6) | 1,838 (1,751 – 1,926) | 2,377 (2,305 – 2,449) | 116 (84 – 148) | 338 – 345 | 296.3 (294.2 – 298.4) | 18,880 (18,130 – 19,629) |
| 20 clients: fixed 100 ms | 2 | 0 of 2 | 0 | 4 (0 – 9) | 1,756 (1,748 – 1,763) | 1,408 (1,397 – 1,418) | 64 (60 – 68) | 191 – 194 | 178.4 (178.2 – 178.7) | 17,973 (17,960 – 17,986) |
| 20 clients: fixed 250 ms | 2 | 0 of 2 | 0 | 8 (2 – 14) | 1,725 (1,602 – 1,848) | 612 (560 – 664) | 24 (20 – 27) | 80 | 78.7 (78.2 – 79.2) | 17,714 (16,429 – 18,999) |
| 20 clients: main | 2 | 2 of 2 | 3,111 – 3,488 | 0 (0 – 1) | 1,798 (1,722 – 1,874) | 1,402 (1,290 – 1,513) | 1,022 (785 – 1,260) | 1,012 – 1,269 | 35.8 (34.0 – 37.6) | 18,594 (17,766 – 19,423) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 20 clients: final candidate v6 | 24 (8 – 40) | - | 7.62 | 1,977 | 99,576 | 6.02 / 14.34 |
| 20 clients: proposed defaults v3 | 9 (8 – 10) | - | 7.31 | 1,946 | 102,651 | 5.63 / 13.70 |
| 20 clients: fixed 50 ms | 2 (-2 – 5) | - | 7.89 | 1,985 | 99,631 | 5.50 / 13.82 |
| 20 clients: fixed 100 ms | 3 (-3 – 10) | - | 7.61 | 1,985 | 101,455 | 6.02 / 13.70 |
| 20 clients: fixed 250 ms | 19 (8 – 30) | - | 7.29 | 1,997 | 104,718 | 5.31 / 13.06 |
| 20 clients: main | 4 (-3 – 10) | - | 7.24 | 1,958 | 101,104 | 5.38 / 13.70 |

<details><summary>Per run: promotion → first success (ms)</summary>

- f_smart6: 40, 8
- f_smart3: 8, 10
- f_f50: -2, 5
- f_f100: 10, -3
- f_f250: 30, 8
- f_main: -3, 10

</details>

<!-- smart-planned -->

## smart-planned

| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Traffic to the old primary after the promotion, median (range), ms | Its last MOVED after the promotion, median (ms) | CLUSTER SHARDS after failover | ops/s, median |
|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 2 | 0 | 1,070 (1,067 – 1,072) | 10 (0 – 19) | 50 | 2 (2) | 282,475 |
| Proposed defaults v3, nothing set | 2 | 0 | 1,076 (1,066 – 1,086) | 23 (18 – 29) | 91 | 2 (2) | 268,650 |
| Fixed 50 ms interval (delay 50 ms, wait 5 s) | 2 | 0 | 936 (873 – 1,000) | 28 (26 – 30) | 102 | 4 (3 – 4) | 270,480 |
| Fixed 250 ms interval (delay 50 ms, wait 5 s) | 2 | 0 | 959 (926 – 992) | 0 (0) | 33 | 2 (2 – 3) | 275,671 |
| Main today | 2 | 0 | 36,569 (28,436 – 44,702) | 262 (196 – 328) | 351 | 3 (3) | 269,087 |

## smart-big-lazy

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 1,000 nodes: proposed defaults v3 | 3 | 0 of 3 | 0 | 0 (0 – 74) | 3,836 (2,555 – 3,841) | 3 (1 – 8) | 0 (0 – 4) | 1 – 3 | 0.0 (0.0) | 17,428 (11,339 – 17,649) |
| 1,000 nodes: v3 + waiting requests expedite a delayed refresh | 3 | 0 of 3 | 0 | 0 (0) | 3,796 (3,693 – 3,805) | 2 (2 – 3) | 0 (0) | 1 | 0.0 (0.0) | 17,817 (17,316 – 17,876) |
| 1,000 nodes: v4 + errors also expedite a delayed refresh | 3 | 0 of 3 | 0 | 254 (225 – 285) | 1,930 (1,928 – 2,158) | 21 (20 – 22) | 11 (11 – 12) | 5 | 1.0 (0.9 – 1.0) | 8,544 (8,361 – 9,550) |
| 1,000 nodes: final candidate v6 (Expedite within the error jitter) | 3 | 0 of 3 | 0 | 208 (201 – 261) | 1,727 (1,695 – 1,981) | 21 (20 – 24) | 11 (10 – 12) | 5 | 1.0 (1.0 – 1.1) | 7,395 (7,119 – 8,895) |
| 1,000+ nodes: main (delay up to 30 s) | 2 | 0 of 2 | 0 | 0 (0) | 3,726 (3,526 – 3,925) | 4 (3 – 5) | 0 (0) | 1 | 0.1 (0.0 – 0.1) | 17,176 (16,240 – 18,112) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 1,000 nodes: proposed defaults v3 | 5167 (115 – 5325) | - | 15.46 | 1,967 | 187,586 | 3.30 / 12.16 |
| 1,000 nodes: v3 + waiting requests expedite a delayed refresh | 5492 (5341 – 5616) | - | 15.36 | 1,951 | 190,365 | 3.23 / 12.42 |
| 1,000 nodes: v4 + errors also expedite a delayed refresh | 201 (45 – 301) | - | 10.92 | 1,974 | 217,418 | 3.30 / 12.42 |
| 1,000 nodes: final candidate v6 (Expedite within the error jitter) | 188 (58 – 273) | - | 9.51 | 1,971 | 244,982 | 2.78 / 11.90 |
| 1,000+ nodes: main (delay up to 30 s) | 4598 (4457 – 4738) | - | 14.85 | 1,944 | 175,664 | 3.68 / 12.42 |

Promotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):

| Comparison | Shift (ms) | p |
|---|---|---|
| 1,000 nodes: proposed defaults v3 vs 1,000 nodes: v3 + waiting requests expedite a delayed refresh | -325 | 0.0809 |
| 1,000 nodes: proposed defaults v3 vs 1,000 nodes: v4 + errors also expedite a delayed refresh | +4966 | 0.383 |
| 1,000 nodes: v3 + waiting requests expedite a delayed refresh vs 1,000 nodes: v4 + errors also expedite a delayed refresh | +5296 | 0.0809 |
| 1,000 nodes: proposed defaults v3 vs 1,000 nodes: final candidate v6 (Expedite within the error jitter) | +4979 | 0.383 |
| 1,000 nodes: v4 + errors also expedite a delayed refresh vs 1,000 nodes: final candidate v6 (Expedite within the error jitter) | +13 | 1 |

<details><summary>Per run: promotion → first success (ms)</summary>

- v3big: 5167, 5325, 115
- v4big: 5492, 5616, 5341
- v5big: 45, 301, 201
- v6big: 58, 273, 188
- lmain: 4457, 4738

</details>

## smart-big-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| 1,000 nodes: proposed defaults v3 | 2 | 0 of 2 | 0 | 14 (5 – 22) | 1,570 (1,443 – 1,696) | 10 (9 – 12) | 2 (2) | 3 | 0.9 (0.9 – 1.0) | 6,733 (6,073 – 7,393) |
| 1,000 nodes: v3 + waiting requests expedite a delayed refresh | 2 | 0 of 2 | 0 | 0 (0) | 1,870 (1,716 – 2,023) | 10 (9 – 10) | 0 (0) | 1 | 1.0 (1.0 – 1.1) | 7,928 (7,314 – 8,541) |
| 1,000 nodes: v4 + errors also expedite a delayed refresh | 2 | 0 of 2 | 0 | 0 (0) | 1,772 (1,387 – 2,156) | 8 (7 – 10) | 0 (0) | 1 | 0.9 (0.9 – 1.0) | 7,662 (5,821 – 9,502) |
| 1,000 nodes: final candidate v6 (Expedite within the error jitter) | 2 | 0 of 2 | 0 | 0 (0) | 1,730 (1,726 – 1,734) | 10 (9 – 10) | 0 (0 – 1) | 1 – 2 | 1.0 (1.0) | 7,543 (7,531 – 7,555) |
| 1,000+ nodes: main (delay up to 30 s) | 2 | 0 of 2 | 0 | 0 (0) | 4,468 (2,455 – 6,481) | 2 (1 – 2) | 0 (0) | 1 | 0.0 (0.0) | 20,774 (10,912 – 30,636) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| 1,000 nodes: proposed defaults v3 | 137 (62 – 212) | - | 6.93 | 1,999 | 242,625 | 3.26 / 12.29 |
| 1,000 nodes: v3 + waiting requests expedite a delayed refresh | 535 (421 – 648) | - | 7.58 | 1,945 | 232,643 | 3.36 / 12.16 |
| 1,000 nodes: v4 + errors also expedite a delayed refresh | 226 (204 – 249) | - | 7.39 | 1,968 | 240,666 | 3.23 / 12.29 |
| 1,000 nodes: final candidate v6 (Expedite within the error jitter) | 185 (119 – 250) | - | 7.32 | 1,962 | 258,819 | 2.91 / 12.03 |
| 1,000+ nodes: main (delay up to 30 s) | 11063 (2976 – 19149) | - | 17.75 | 1,944 | 170,451 | 3.46 / 12.16 |

<details><summary>Per run: promotion → first success (ms)</summary>

- v3big: 212, 62
- v4big: 648, 421
- v5big: 204, 249
- v6big: 250, 119
- lmain: 19149, 2976

</details>

## smart-pause

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 2 | 0 of 2 | 0 | 1,000 (1,000) | 0 (0) | 8 (8) | 0 (0) | 1 | 0.6 (0.6) | 1 (1) |
| Main today | 1 | 0 of 1 | 0 | 1,099 (1,099) | 101 (101) | 3 (3) | 0 (0) | 1 | 0.1 (0.1) | 3 (3) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 2649 (2636 – 2662) | - | 9.98 | 28 | 198,133 | 2.98 / 12.16 |
| Main today | 4883 (4883) | - | 12.09 | 224 | 164,204 | 3.23 / 12.16 |

<details><summary>Per run: promotion → first success (ms)</summary>

- smart6: 2662, 2636
- main: 4883

</details>

## smart-migrate

Atomic Slot Migration of slots 0-199 (all keys there) to another primary.

| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Client traffic to the source after the handover, median (range), ms | Last MOVED from the source after the handover, median (ms) | CLUSTER SHARDS during the run | Migration errors |
|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 1 | 0 | 344 (344) | 23 (23) | 25 | 2 (2) | 0 |
| Main today | 1 | 0 | 49,663 (49,663) | 320 (320) | 334 | 3 (3) | 0 |

## smart-idle-kill

The primary of the keys' shard crashes 1 s into an 18 s idle window and stays down past the resume. No ShardsRefreshInterval, so the client resumes with the dead primary in its slot map.

| Build | Runs | Failed after resume, median (range) | ErrNoSlot / deadline after resume | First success after resume, median (range), ms | Last error after resume, median (range), ms | MOVED in the first 5 s, median (range) | CLUSTER SHARDS in the first 5 s | Longest success, median (ms) |
|---|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 1 | 116 (116) | 0 / 0 | 61 (61) | 1.8 (1.8) | 0 (0) | 2 (2) | 121 |
| Main today | 1 | 100 (100) | 0 / 0 | 586 (586) | 3.3 (3.3) | 0 (0) | 1 (1) | 1,014 |

<details><summary>Per run</summary>

- smart6-1: failed 116, first success 60.644 ms, last error 1.8 ms, MOVED 0, errors [('dial tcp 127.0.0.1:7101: connect: connection refused', 116)]
- main-1: failed 100, first success 586.043 ms, last error 3.275 ms, MOVED 0, errors [('dial tcp 127.0.0.1:7101: connect: connection refused', 100)]

</details>

## smart-pipe-kill

| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |
|---|---|---|---|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 1 | 0 of 1 | 0 | 8,590 (8,590) | 6,061 (6,061) | 26 (26) | 0 (0) | 4 | 3.1 (3.1) | 1,495 (1,495) |
| Main today | 1 | 0 of 1 | 0 | 8,730 (8,730) | 6,261 (6,261) | 17 (17) | 1 (1) | 3 | 1.9 (1.9) | 1,676 (1,676) |

| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |
|---|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 91 (91) | - | 7.70 | 2,009 | 459,960 | 13.44 / 22.27 |
| Main today | 394 (394) | - | 8.28 | 438 | 417,340 | 13.70 / 21.76 |

<details><summary>Per run: promotion → first success (ms)</summary>

- smart6: 91
- main: 394

</details>

<!-- smart-steady -->

## steady

| Build | Runs | Errors | CLUSTER SHARDS during the run | ops/s, median (range) | p50 / p99 / p999, median (ms) |
|---|---|---|---|---|---|
| Final candidate v6 (v3 + Expedite within the error jitter), nothing set | 1 | 0 | 1 | 310,459 (310,459) | 0.54 / 2.53 / 11.90 |
| Main today | 1 | 0 | 1 | 261,105 (261,105) | 0.58 / 3.49 / 12.16 |
