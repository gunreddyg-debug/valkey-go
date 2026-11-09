Tables made by `bench/bin/report runs-smart`. Ranges are min – max over the runs of each build.

## Runs

| Scenario | Runs | OK | Not OK | Wall time |
|---|---|---|---|---|
| smart-kill | 28 | 28 | – | 17.6 min |
| smart-hold3 | 18 | 18 | – | 11.2 min |
| smart-lazy | 21 | 21 | – | 13.1 min |
| smart-elect5 | 12 | 12 | – | 8.5 min |
| smart-fleet | 12 | 12 | – | 7.5 min |
| smart-planned | 10 | 10 | – | 5.5 min |
| smart-big-lazy | 14 | 14 | – | 14.7 min |
| smart-big-kill | 10 | 10 | – | 10.4 min |
| smart-pause | 3 | 3 | – | 2.1 min |
| smart-migrate | 2 | 2 | – | 1.0 min |
| smart-idle-kill | 2 | 2 | – | 1.6 min |
| smart-pipe-kill | 2 | 2 | – | 1.2 min |
| smart-steady | 2 | 2 | – | 1.1 min |

## smart-big-kill

`MODE=kill ./fail-primary.sh 30` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| lmain | – | 2 | 6,432,828 – 10,612,528 | 0 | 0 | 2,455 – 6,481 | 1 – 2 / 1 – 2 / 0 / 1 | 0 | 0 / 0 |
| v3big | 0s / 0s | 2 | 12,126,496 – 12,136,355 | 0 | 5 – 22 | 1,443 – 1,696 | 10 – 12 / 9 – 12 / 2 / 3 | 0 | 0 / 0 |
| v4big | 0s / 0s | 2 | 11,446,928 – 11,817,660 | 0 | 0 | 1,716 – 2,023 | 10 / 9 – 10 / 0 / 1 | 0 | 0 / 0 |
| v5big | 0s / 0s | 2 | 11,669,664 – 12,397,765 | 0 | 0 | 1,387 – 2,156 | 8 – 11 / 7 – 10 / 0 / 1 | 0 | 0 / 0 |
| v6big | 0s / 0s | 2 | 11,857,449 – 14,024,641 | 0 | 0 | 1,726 – 1,734 | 9 – 10 / 9 – 10 / 0 – 1 / 1 – 2 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| lmain | 9.63 – 25.86 s | +2.98 – +19.15 s | 9.61 – 25.85 s | 0 ms | never | 128,655–212,247 | 0.57 / 3.42–3.49 / 11.90–12.42 | 0.44–0.47 / 2.72–3.04 / 12.67–16.26 | 1.94 – 1.95 s |
| v3big | 6.40 – 7.45 s | +62 – +212 ms | 6.40 – 7.44 s | 0 ms | never | 242,526–242,723 | 0.54–0.55 / 3.23–3.30 / 12.16–12.42 | 0.52–0.58 / 3.04–3.94 / 13.95–15.49 | 2.00 s |
| v4big | 7.16 – 8.00 s | +421 – +648 ms | 7.15 – 7.99 s | 0 ms | never | 228,934–236,351 | 0.55–0.57 / 3.30–3.42 / 12.16 | 0.50–0.51 / 3.17–3.36 / 11.65–12.93 | 1.94 – 1.95 s |
| v5big | 5.90 – 8.88 s | +204 – +249 ms | 5.90 – 8.87 s | 0 ms | never | 233,387–247,946 | 0.54 / 3.23 / 12.16–12.42 | 0.45–0.48 / 3.23–3.55 / 12.67–17.66 | 1.94 – 2.00 s |
| v6big | 7.27 – 7.36 s | +119 – +250 ms | 7.16 – 7.26 s | 0 ms | never | 237,148–280,491 | 0.50–0.55 / 2.53–3.30 / 11.65–12.42 | 0.43–0.44 / 1.65–4.67 / 12.16–17.15 | 1.95 – 1.97 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| lmain-1 | 6,432,828 | 0 | 0 | 6,481 | 1 / 1 / 0 / 1 | 0 | 0 / 0 | 25.86 s | +19149 ms | 3.42 / 3.04 | – | 6,384 × dial tcp 127.0.0.1:7101: connect: connection refused |
| lmain-2 | 10,612,528 | 0 | 0 | 2,455 | 2 / 2 / 0 / 1 | 0 | 0 / 0 | 9.63 s | +2976 ms | 3.49 / 2.72 | – | 2,344 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v3big-1 | 12,136,355 | 0 | 5 | 1,696 | 12 / 12 / 2 / 3 | 0 | 0 / 0 | 7.45 s | +212 ms | 3.23 / 3.94 | – | 1,608 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v3big-2 | 12,126,496 | 0 | 22 | 1,443 | 10 / 9 / 2 / 3 | 0 | 0 / 0 | 6.40 s | +62 ms | 3.30 / 3.04 | – | 1,346 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v4big-1 | 11,817,660 | 0 | 0 | 2,023 | 10 / 9 / 0 / 1 | 0 | 0 / 0 | 8.00 s | +648 ms | 3.30 / 3.36 | – | 1,934 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v4big-2 | 11,446,928 | 0 | 0 | 1,716 | 10 / 10 / 0 / 1 | 0 | 0 / 0 | 7.16 s | +421 ms | 3.42 / 3.17 | – | 1,626 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v5big-1 | 12,397,765 | 0 | 0 | 1,387 | 8 / 7 / 0 / 1 | 0 | 0 / 0 | 5.90 s | +204 ms | 3.23 / 3.55 | – | 1,274 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v5big-2 | 11,669,664 | 0 | 0 | 2,156 | 11 / 10 / 0 / 1 | 0 | 0 / 0 | 8.88 s | +249 ms | 3.23 / 3.23 | – | 2,056 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v6big-1 | 11,857,449 | 0 | 0 | 1,734 | 9 / 9 / 0 / 1 | 0 | 0 / 0 | 7.27 s | +250 ms | 3.30 / 4.67 | – | 1,634 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v6big-2 | 14,024,641 | 0 | 0 | 1,726 | 10 / 10 / 1 / 2 | 0 | 0 / 0 | 7.36 s | +119 ms | 2.53 / 1.65 | – | 1,624 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-big-lazy

`MODE=kill HOLD_PROMOTION=3 ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 1s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| lmain | ShardsRefresh 1s | 2 | 8,412,108 – 9,154,677 | 0 | 0 | 3,526 – 3,925 | 3 – 5 / 3 – 5 / 0 / 1 | 0 | 0 – 386,720 / 0 |
| v3big | 0s / 0s, ShardsRefresh 1s | 3 | 9,250,506 – 10,486,561 | 0 | 0 – 74 | 2,555 – 3,841 | 1 – 8 / 1 – 8 / 0 – 4 / 1 – 3 | 0 | 0 – 966 / 0 |
| v4big | 0s / 0s, ShardsRefresh 1s | 3 | 9,404,639 – 9,617,104 | 0 | 0 | 3,693 – 3,805 | 2 – 3 / 2 – 3 / 0 / 1 | 0 | 0 – 965 / 0 |
| v5big | 0s / 0s, ShardsRefresh 1s | 3 | 10,785,098 – 10,968,045 | 0 | 225 – 285 | 1,928 – 2,158 | 21 – 23 / 20 – 22 / 11 – 12 / 5 | 0 | 0 / 0 |
| v6big | 0s / 0s, ShardsRefresh 1s | 3 | 10,808,665 – 12,394,470 | 0 | 201 – 261 | 1,695 – 1,981 | 21 – 24 / 20 – 24 / 10 – 12 / 5 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| lmain | 14.24 – 15.46 s | +4.46 – +4.74 s | 14.23 – 15.43 s | 0.00 – 7.24 s | 7.32 s (1× never) | 168,241–183,088 | 0.58–0.62 / 3.62–3.74 / 12.42 | 0.44–0.86 / 3.55–6.72 / 13.44–421.89 | 1.93 – 1.96 s |
| v3big | 11.27 – 15.46 s | +0.11 – +5.33 s | 11.27 – 15.44 s | 0.00 – 5.37 s | 5.24 – 5.48 s (1× never) | 185,008–209,722 | 0.57–0.58 / 3.17–3.36 / 12.16–12.67 | 0.43–0.51 / 3.04–4.06 / 14.72–16.00 | 1.96 – 2.00 s |
| v4big | 14.89 – 15.37 s | +5.34 – +5.62 s | 14.88 – 15.35 s | 0.00 – 5.55 s | 5.46 – 5.56 s (1× never) | 188,090–192,341 | 0.57 / 3.23–3.36 / 12.16–12.42 | 0.44–0.51 / 2.08–3.36 / 12.16–16.64 | 1.90 – 1.99 s |
| v5big | 10.38 – 11.54 s | +45 – +301 ms | 10.36 – 11.54 s | 0 ms | never | 215,700–219,359 | 0.55–0.57 / 3.23–3.36 / 12.16–12.42 | 0.51–0.54 / 2.98–3.17 / 12.42–12.93 | 1.96 – 1.99 s |
| v6big | 9.33 – 11.08 s | +58 – +273 ms | 9.09 – 11.07 s | 0 ms | never | 216,171–247,887 | 0.54–0.58 / 2.53–3.42 / 11.90–12.42 | 0.49–0.60 / 2.59–4.06 / 11.14–13.44 | 1.78 – 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| lmain-1 | 9,154,677 | 0 | 0 | 3,526 | 5 / 5 / 0 / 1 | 0 | 0 / 0 | 14.24 s | +4457 ms | 3.62 / 3.55 | – | 3,426 × dial tcp 127.0.0.1:7101: connect: connection refused |
| lmain-2 | 8,412,108 | 0 | 0 | 3,925 | 3 / 3 / 0 / 1 | 0 | 386,720 / 0 | 15.46 s | +4738 ms | 3.74 / 6.72 | – | 3,824 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v3big-1 | 9,250,506 | 0 | 0 | 3,836 | 3 / 3 / 0 / 1 | 0 | 966 / 0 | 15.46 s | +5167 ms | 3.30 / 3.04 | – | 3,736 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v3big-2 | 9,379,817 | 0 | 0 | 3,841 | 1 / 1 / 0 / 1 | 0 | 960 / 0 | 15.46 s | +5325 ms | 3.17 / 4.06 | CLUSTERDOWN 1 | 3,738 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v3big-3 | 10,486,561 | 0 | 74 | 2,555 | 8 / 8 / 4 / 3 | 0 | 0 / 0 | 11.27 s | +115 ms | 3.36 / 3.10 | – | 2,457 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v4big-1 | 9,518,395 | 0 | 0 | 3,796 | 3 / 2 / 0 / 1 | 0 | 965 / 0 | 15.37 s | +5492 ms | 3.23 / 2.08 | – | 3,692 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v4big-2 | 9,617,104 | 0 | 0 | 3,693 | 3 / 3 / 0 / 1 | 0 | 0 / 0 | 14.89 s | +5616 ms | 3.23 / 3.23 | – | 3,592 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v4big-3 | 9,404,639 | 0 | 0 | 3,805 | 2 / 2 / 0 / 1 | 0 | 960 / 0 | 15.36 s | +5341 ms | 3.36 / 3.36 | – | 3,704 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v5big-1 | 10,871,032 | 0 | 254 | 1,930 | 21 / 20 / 11 / 5 | 0 | 0 / 0 | 10.38 s | +45 ms | 3.30 / 2.98 | – | 1,826 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v5big-2 | 10,968,045 | 0 | 285 | 1,928 | 22 / 21 / 12 / 5 | 0 | 0 / 0 | 10.92 s | +301 ms | 3.23 / 3.10 | – | 1,832 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v5big-3 | 10,785,098 | 0 | 225 | 2,158 | 23 / 22 / 11 / 5 | 0 | 0 / 0 | 11.54 s | +201 ms | 3.36 / 3.17 | – | 2,054 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v6big-1 | 10,808,665 | 0 | 201 | 1,727 | 21 / 21 / 10 / 5 | 0 | 0 / 0 | 9.33 s | +58 ms | 3.42 / 4.06 | – | 1,619 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v6big-2 | 12,394,470 | 0 | 208 | 1,695 | 21 / 20 / 11 / 5 | 0 | 0 / 0 | 9.51 s | +273 ms | 2.78 / 2.59 | – | 1,598 × dial tcp 127.0.0.1:7101: connect: connection refused |
| v6big-3 | 12,249,178 | 0 | 261 | 1,981 | 24 / 24 / 12 / 5 | 0 | 0 / 0 | 11.08 s | +188 ms | 2.53 / 2.66 | – | 1,884 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-elect5

`MODE=kill HOLD_PROMOTION=5 ./fail-primary.sh 20` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 100ms / 5s | 2 | 4,701,808 – 4,821,165 | 0 | 510 – 511 | 1,732 – 1,876 | 124 – 129 / 122 – 128 / 49 – 50 / 10 | 0 | 0 / 0 |
| f250 | 250ms / 5s | 2 | 4,980,935 – 5,058,538 | 0 | 469 – 513 | 1,668 – 1,726 | 51 – 52 / 50 – 51 / 20 / 4 | 0 | 0 / 0 |
| f50 | 50ms / 5s | 2 | 4,971,842 – 5,147,116 | 0 | 494 – 518 | 1,452 – 1,699 | 213 – 236 / 211 – 234 / 95 – 97 / 20 | 0 | 0 / 0 |
| main | – | 2 | 4,709,706 – 4,924,354 | 847,254 – 930,705 | 0 | 1,931 – 2,043 | 4,554 – 4,937 / 4,554 – 4,937 / 4,515 – 4,874 / 990 – 1,165 | 0 | 0 / 0 |
| smart3 | 0s / 0s | 2 | 4,515,520 – 4,572,869 | 0 | 502 – 509 | 1,878 | 94 – 95 / 93 – 94 / 61 – 62 / 19 | 0 | 0 / 0 |
| smart6 | 0s / 0s | 2 | 4,995,731 – 5,306,665 | 0 | 465 – 541 | 1,683 – 1,785 | 92 – 95 / 91 – 94 / 62 – 63 / 19 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 12.36 – 12.95 s | +57 – +109 ms | 12.36 – 12.93 s | 0 ms | never | 156,722–160,704 | 0.55–0.57 / 3.17–3.42 / 11.90–12.42 | 0.52–0.74 / 3.04–5.82 / 12.16–18.69 | 1.99 s |
| f250 | 12.00 – 12.20 s | +197 – +227 ms | 12.00 – 12.19 s | 0 ms | never | 166,027–168,616 | 0.55 / 3.23–3.30 / 12.16 | 0.52–0.54 / 3.04–3.36 / 12.67–13.18 | 1.99 – 2.00 s |
| f50 | 11.22 – 12.24 s | +57 – +90 ms | 11.17 – 12.23 s | 0 ms | never | 165,726–171,568 | 0.55–0.57 / 3.23–3.36 / 12.16–12.67 | 0.52–0.54 / 3.49–3.55 / 12.16–17.66 | 1.97 – 2.00 s |
| main | 12.54 – 12.56 s | +36 – +44 ms | 12.52 – 12.55 s | 0 ms | never | 156,986–164,145 | 0.54–0.57 / 3.23–3.42 / 12.16 | 0.50–0.57 / 3.10–3.36 / 12.16–12.67 | 26 ms |
| smart3 | 12.85 – 12.94 s | +151 – +200 ms | 12.84 – 12.93 s | 0 ms | never | 150,516–152,428 | 0.55–0.57 / 3.36–3.68 / 12.42–12.67 | 0.52 / 2.98–3.23 / 11.65–12.67 | 1.98 s |
| smart6 | 11.99 – 12.61 s | +221 ms | 11.99 – 12.60 s | 0 ms | never | 166,512–176,884 | 0.52–0.54 / 3.10–3.30 / 11.90–12.42 | 0.50–0.54 / 2.78–3.87 / 12.42–25.34 | 1.99 – 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f100-1 | 4,701,808 | 0 | 511 | 1,732 | 124 / 122 / 50 / 10 | 0 | 0 / 0 | 12.36 s | +109 ms | 3.42 / 5.82 | – | 1,622 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-2 | 4,821,165 | 0 | 510 | 1,876 | 129 / 128 / 49 / 10 | 0 | 0 / 0 | 12.95 s | +57 ms | 3.17 / 3.04 | – | 1,778 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-1 | 4,980,935 | 0 | 513 | 1,726 | 52 / 51 / 20 / 4 | 0 | 0 / 0 | 12.20 s | +197 ms | 3.23 / 3.04 | – | 1,624 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-2 | 5,058,538 | 0 | 469 | 1,668 | 51 / 50 / 20 / 4 | 0 | 0 / 0 | 12.00 s | +227 ms | 3.30 / 3.36 | – | 1,572 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-1 | 4,971,842 | 0 | 494 | 1,699 | 236 / 234 / 97 / 20 | 0 | 0 / 0 | 12.24 s | +90 ms | 3.23 / 3.55 | – | 1,608 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-2 | 5,147,116 | 0 | 518 | 1,452 | 213 / 211 / 95 / 20 | 0 | 0 / 0 | 11.22 s | +57 ms | 3.36 / 3.49 | – | 1,345 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-1 | 4,709,706 | 847,254 | 0 | 1,931 | 4,554 / 4,554 / 4,515 / 990 | 0 | 0 / 0 | 12.54 s | +36 ms | 3.42 / 3.36 | – | 1,836 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-2 | 4,924,354 | 930,705 | 0 | 2,043 | 4,937 / 4,937 / 4,874 / 1,165 | 0 | 0 / 0 | 12.56 s | +44 ms | 3.23 / 3.10 | – | 1,946 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-1 | 4,572,869 | 0 | 502 | 1,878 | 95 / 94 / 62 / 19 | 0 | 0 / 0 | 12.94 s | +151 ms | 3.36 / 2.98 | – | 1,798 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-2 | 4,515,520 | 0 | 509 | 1,878 | 94 / 93 / 61 / 19 | 0 | 0 / 0 | 12.85 s | +200 ms | 3.68 / 3.23 | – | 1,779 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-1 | 5,306,665 | 0 | 465 | 1,683 | 92 / 91 / 62 / 19 | 0 | 0 / 0 | 11.99 s | +221 ms | 3.10 / 2.78 | – | 1,586 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-2 | 4,995,731 | 0 | 541 | 1,785 | 95 / 94 / 63 / 19 | 0 | 0 / 0 | 12.61 s | +221 ms | 3.30 / 3.87 | – | 1,688 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-fleet

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f_f100 | 100ms / 5s | 2 | 2,530,429 – 2,542,352 | 0 | 0 – 9 | 1,748 – 1,763 | 1,437 – 1,458 / 1,397 – 1,418 / 60 – 68 / 191 – 194 | 0 | 0 / 0 |
| f_f250 | 250ms / 5s | 2 | 2,508,757 – 2,727,201 | 0 | 2 – 14 | 1,602 – 1,848 | 580 – 684 / 560 – 664 / 20 – 27 / 80 | 0 | 0 / 0 |
| f_f50 | 50ms / 5s | 2 | 2,468,043 – 2,513,556 | 0 | 5 – 6 | 1,751 – 1,926 | 2,394 – 2,508 / 2,305 – 2,449 / 84 – 148 / 338 – 345 | 0 | 0 / 0 |
| f_main | – | 2 | 2,518,055 – 2,537,216 | 3,111 – 3,488 | 0 – 1 | 1,722 – 1,874 | 1,293 – 1,523 / 1,290 – 1,513 / 785 – 1,260 / 1,012 – 1,269 | 0 | 0 / 0 |
| f_smart3 | 0s / 0s | 2 | 2,560,211 – 2,572,404 | 0 | 0 | 1,707 – 1,735 | 642 – 706 / 622 – 686 / 44 – 86 / 81 – 141 | 0 | 0 / 0 |
| f_smart6 | 0s / 0s | 2 | 2,445,672 – 2,533,262 | 0 | 1 – 4 | 1,713 – 1,921 | 696 – 762 / 677 – 742 / 87 – 97 / 113 – 130 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f_f100 | 7.56 – 7.66 s | -3 – +10 ms | 7.32 – 7.75 s | 0 ms | never | 101,216–101,694 | 1.20 / 5.95–6.08 / 13.18–14.21 | 1.14 / 6.72–6.98 / 958.46–1327.10 | 1.98 – 1.99 s |
| f_f250 | 6.65 – 7.93 s | +8 – +30 ms | 6.73 – 8.10 s | 0 ms | never | 100,350–109,087 | 1.17–1.20 / 5.18–5.44 / 12.93–13.18 | 1.10–1.14 / 4.80–5.70 / 471.04–479.23 | 2.00 s |
| f_f50 | 7.67 – 8.10 s | -2 – +5 ms | 7.77 – 8.13 s | 0 ms | never | 98,720–100,542 | 1.20 / 5.18–5.82 / 13.44–14.21 | 1.14–1.17 / 3.81–8.83 / 503.81–1097.73 | 1.97 – 2.00 s |
| f_main | 6.85 – 7.64 s | -3 – +10 ms | 7.61 – 8.21 s | 0 ms | never | 100,721–101,487 | 1.20 / 5.06–5.70 / 13.44–13.95 | 1.07–1.10 / 4.06–5.18 / 13.70–16.64 | 1.94 – 1.97 s |
| f_smart3 | 7.11 – 7.52 s | +8 – +10 ms | 6.99 – 7.33 s | 0 ms | never | 102,408–102,893 | 1.20 / 5.57–5.70 / 13.70 | 1.07–1.14 / 3.74–5.95 / 925.70–1007.62 | 1.94 – 1.95 s |
| f_smart6 | 7.25 – 7.99 s | +8 – +40 ms | 7.22 – 8.07 s | 0 ms | never | 97,823–101,328 | 1.20 / 5.82–6.21 / 13.44–15.23 | 1.14 / 6.72–8.58 / 495.62–892.93 | 1.96 – 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f_f100-1 | 2,542,352 | 0 | 0 | 1,763 | 1,437 / 1,397 / 60 / 191 | 0 | 0 / 0 | 7.56 s | +10 ms | 5.95 / 6.72 | – | 1,654 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_f100-2 | 2,530,429 | 0 | 9 | 1,748 | 1,458 / 1,418 / 68 / 194 | 0 | 0 / 0 | 7.66 s | -3 ms | 6.08 / 6.98 | – | 1,650 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_f250-1 | 2,727,201 | 0 | 2 | 1,602 | 580 / 560 / 20 / 80 | 0 | 0 / 0 | 6.65 s | +30 ms | 5.44 / 4.80 | – | 1,506 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_f250-2 | 2,508,757 | 0 | 14 | 1,848 | 684 / 664 / 27 / 80 | 0 | 0 / 0 | 7.93 s | +8 ms | 5.18 / 5.70 | – | 1,756 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_f50-1 | 2,513,556 | 0 | 6 | 1,751 | 2,394 / 2,305 / 148 / 338 | 0 | 0 / 0 | 7.67 s | -2 ms | 5.18 / 8.83 | – | 1,652 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_f50-2 | 2,468,043 | 0 | 5 | 1,926 | 2,508 / 2,449 / 84 / 345 | 0 | 0 / 0 | 8.10 s | +5 ms | 5.82 / 3.81 | – | 1,830 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_main-1 | 2,537,216 | 3,111 | 1 | 1,722 | 1,523 / 1,513 / 1,260 / 1,269 | 0 | 0 / 0 | 6.85 s | -3 ms | 5.70 / 5.18 | – | 1,618 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_main-2 | 2,518,055 | 3,488 | 0 | 1,874 | 1,293 / 1,290 / 785 / 1,012 | 0 | 0 / 0 | 7.64 s | +10 ms | 5.06 / 4.06 | – | 1,786 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_smart3-1 | 2,572,404 | 0 | 0 | 1,707 | 642 / 622 / 44 / 81 | 0 | 0 / 0 | 7.11 s | +8 ms | 5.70 / 3.74 | – | 1,600 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_smart3-2 | 2,560,211 | 0 | 0 | 1,735 | 706 / 686 / 86 / 141 | 0 | 0 / 0 | 7.52 s | +10 ms | 5.57 / 5.95 | – | 1,640 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_smart6-1 | 2,533,262 | 0 | 1 | 1,713 | 696 / 677 / 97 / 130 | 0 | 0 / 0 | 7.25 s | +40 ms | 5.82 / 6.72 | – | 1,606 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f_smart6-2 | 2,445,672 | 0 | 4 | 1,921 | 762 / 742 / 87 / 113 | 0 | 0 / 0 | 7.99 s | +8 ms | 6.21 / 8.58 | – | 1,810 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-hold3

`MODE=kill HOLD_PROMOTION=3 ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 100ms / 5s | 3 | 4,337,005 – 4,550,393 | 0 | 241 – 327 | 1,703 – 1,742 | 100 – 104 / 99 – 103 / 30 / 10 | 0 | 0 / 0 |
| f250 | 250ms / 5s | 3 | 4,037,773 – 4,402,981 | 0 | 266 – 344 | 1,655 – 1,963 | 42 – 46 / 41 – 45 / 12 – 13 / 4 | 0 | 0 / 0 |
| f50 | 50ms / 5s | 3 | 4,049,994 – 4,663,930 | 0 | 287 – 291 | 1,530 – 1,930 | 181 – 210 / 178 – 206 / 58 / 20 | 0 | 0 / 0 |
| main | – | 3 | 4,059,219 – 4,257,132 | 468,984 – 555,356 | 0 – 17 | 1,769 – 2,092 | 2,401 – 3,225 / 2,401 – 3,225 / 2,348 – 3,183 / 1,109 – 1,199 | 0 | 0 / 0 |
| smart2 | 0s / 0s | 3 | 4,249,399 – 4,531,097 | 0 | 228 – 312 | 1,472 – 1,802 | 75 – 80 / 74 – 79 / 47 – 48 / 19 – 20 | 0 | 0 / 0 |
| smart3 | 0s / 0s | 1 | 3,991,276 | 0 | 300 | 1,878 | 81 / 80 / 48 / 20 | 0 | 0 / 0 |
| smart6 | 0s / 0s | 2 | 3,818,853 – 4,782,036 | 0 | 308 – 315 | 1,705 – 1,816 | 79 – 80 / 78 / 48 / 19 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 9.95 – 10.40 s | +41 – +104 ms | 9.95 – 10.40 s | 0 ms | never | 173,478–182,013 | 0.52–0.54 / 2.98–3.04 / 11.14–11.65 | 0.49–0.52 / 3.04–3.23 / 10.88–13.95 | 1.99 – 2.00 s |
| f250 | 9.95 – 11.01 s | +132 – +256 ms | 9.95 – 10.97 s | 0 ms | never | 161,507–176,115 | 0.54–0.55 / 2.91–3.17 / 11.39–11.90 | 0.50–0.55 / 2.91–3.36 / 12.16–14.72 | 2.00 s |
| f50 | 9.36 – 10.91 s | +31 – +59 ms | 9.27 – 10.90 s | 0 ms | never | 161,999–186,552 | 0.54 / 2.98–3.10 / 11.65–11.90 | 0.51–0.52 / 3.04–3.17 / 13.18–23.30 | 1.99 – 2.00 s |
| main | 10.09 – 11.37 s | +33 – +41 ms | 10.09 – 11.36 s | 0 ms | never | 162,366–170,284 | 0.52–0.54 / 2.98–3.10 / 11.65–12.16 | 0.50–0.54 / 2.98–3.36 / 10.62–12.16 | 21 – 34 ms |
| smart2 | 9.32 – 10.51 s | +41 – +77 ms | 9.30 – 10.50 s | 0 ms | never | 169,971–181,242 | 0.54–0.55 / 3.04–3.10 / 11.65–12.16 | 0.51–0.54 / 3.10–3.36 / 10.62–15.74 | 1.88 – 2.00 s |
| smart3 | 10.80 s | +85 ms | 10.78 s | 0 ms | never | 159,647 | 0.55 / 3.23 / 12.16 | 0.52 / 2.91 / 14.72 | 1.99 s |
| smart6 | 10.20 – 10.26 s | +42 – +77 ms | 10.19 – 10.25 s | 0 ms | never | 152,751–191,277 | 0.51–0.57 / 2.27–3.55 / 11.65–12.67 | 0.49–0.55 / 2.34–3.68 / 12.42–18.18 | 1.98 – 1.99 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f100-1 | 4,550,393 | 0 | 241 | 1,703 | 100 / 99 / 30 / 10 | 0 | 0 / 0 | 9.95 s | +41 ms | 2.98 / 3.04 | – | 1,588 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-2 | 4,337,005 | 0 | 327 | 1,742 | 104 / 103 / 30 / 10 | 0 | 0 / 0 | 10.40 s | +78 ms | 2.98 / 3.17 | – | 1,638 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-3 | 4,362,602 | 0 | 295 | 1,718 | 102 / 101 / 30 / 10 | 0 | 0 / 0 | 10.16 s | +104 ms | 3.04 / 3.23 | – | 1,612 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-1 | 4,037,773 | 0 | 268 | 1,963 | 46 / 45 / 12 / 4 | 0 | 0 / 0 | 11.01 s | +132 ms | 3.10 / 3.04 | – | 1,856 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-2 | 4,081,192 | 0 | 344 | 1,794 | 45 / 44 / 13 / 4 | 0 | 0 / 0 | 10.74 s | +246 ms | 3.17 / 3.36 | – | 1,686 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-3 | 4,402,981 | 0 | 266 | 1,655 | 42 / 41 / 13 / 4 | 0 | 0 / 0 | 9.95 s | +256 ms | 2.91 / 2.91 | – | 1,566 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-1 | 4,663,930 | 0 | 291 | 1,530 | 181 / 178 / 58 / 20 | 0 | 0 / 0 | 9.36 s | +59 ms | 3.04 / 3.04 | – | 1,422 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-2 | 4,427,717 | 0 | 290 | 1,706 | 194 / 191 / 58 / 20 | 0 | 0 / 0 | 10.12 s | +31 ms | 2.98 / 3.17 | – | 1,608 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-3 | 4,049,994 | 0 | 287 | 1,930 | 210 / 206 / 58 / 20 | 0 | 0 / 0 | 10.91 s | +45 ms | 3.10 / 3.04 | – | 1,826 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-1 | 4,153,809 | 551,654 | 16 | 1,999 | 2,850 / 2,850 / 2,794 / 1,199 | 0 | 0 / 0 | 10.85 s | +38 ms | 2.98 / 2.98 | – | 1,907 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-2 | 4,059,219 | 468,984 | 17 | 2,092 | 2,401 / 2,401 / 2,348 / 1,109 | 0 | 0 / 0 | 11.37 s | +41 ms | 2.98 / 2.98 | – | 1,988 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-3 | 4,257,132 | 555,356 | 0 | 1,769 | 3,225 / 3,225 / 3,183 / 1,159 | 0 | 0 / 0 | 10.09 s | +33 ms | 3.10 / 3.36 | – | 1,658 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-1 | 4,249,399 | 0 | 310 | 1,802 | 80 / 79 / 48 / 19 | 0 | 0 / 0 | 10.51 s | +41 ms | 3.04 / 3.10 | – | 1,696 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-2 | 4,531,097 | 0 | 228 | 1,647 | 75 / 74 / 47 / 20 | 0 | 0 / 0 | 9.40 s | +77 ms | 3.10 / 3.36 | – | 1,544 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-3 | 4,483,561 | 0 | 312 | 1,472 | 75 / 74 / 48 / 19 | 0 | 0 / 0 | 9.32 s | +66 ms | 3.10 / 3.17 | – | 1,376 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-21 | 3,991,276 | 0 | 300 | 1,878 | 81 / 80 / 48 / 20 | 0 | 0 / 0 | 10.80 s | +85 ms | 3.23 / 2.91 | – | 1,770 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-1 | 3,818,853 | 0 | 315 | 1,816 | 80 / 78 / 48 / 19 | 0 | 0 / 0 | 10.26 s | +42 ms | 3.55 / 3.68 | – | 1,716 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-2 | 4,782,036 | 0 | 308 | 1,705 | 79 / 78 / 48 / 19 | 0 | 0 / 0 | 10.20 s | +77 ms | 2.27 / 2.34 | – | 1,616 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-idle-kill

`MODE=kill ./fail-primary.sh 30` at t = 5 s; ShardsRefreshInterval 0s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | ShardsRefresh 0s | 1 | 3,055,143 | 0 | 0 | 100 | 2 / 2 / 0 / 1 | 0 | 0 / 0 |
| smart6 | 0s / 0s, ShardsRefresh 0s | 1 | 3,846,715 | 0 | 0 | 116 | 3 / 2 / 0 / 2 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 17.39 s | +11165 ms | 16.81 s | 0 ms | never | 101,834 | 0.55 / 3.49 / 11.90 | 0.50 / 3.94 / 12.67 | 1.01 s |
| smart6 | 16.79 s | +9819 ms | 16.73 s | 0 ms | never | 128,220 | 0.51 / 2.78 / 11.90 | 0.48 / 3.17 / 15.23 | 121 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 3,055,143 | 0 | 0 | 100 | 2 / 2 / 0 / 1 | 0 | 0 / 0 | 17.39 s | +11165 ms | 3.49 / 3.94 | – | 100 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-1 | 3,846,715 | 0 | 0 | 116 | 3 / 2 / 0 / 2 | 0 | 0 / 0 | 16.79 s | +9819 ms | 2.78 / 3.17 | – | 116 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-kill

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 100ms / 5s | 4 | 4,812,769 – 5,077,380 | 0 | 3 – 17 | 1,460 – 2,062 | 67 – 86 / 66 – 83 / 2 – 4 / 10 | 0 | 0 / 0 |
| f250 | 250ms / 5s | 4 | 4,785,126 – 5,253,700 | 0 | 4 – 29 | 1,781 – 1,978 | 34 – 36 / 33 – 35 / 1 – 2 / 4 | 0 | 0 / 0 |
| f50 | 50ms / 5s | 4 | 4,321,349 – 5,076,462 | 0 | 0 – 11 | 1,717 – 1,930 | 143 – 158 / 135 – 154 / 5 – 7 / 20 | 0 | 0 / 0 |
| main | – | 4 | 4,690,416 – 5,152,056 | 0 | 0 | 1,695 – 2,060 | 14 – 21 / 14 – 21 / 0 / 2 – 4 | 0 | 0 / 0 |
| smart | 0s / 0s | 4 | 4,618,860 – 5,937,237 | 0 | 0 – 1 | 1,380 – 2,120 | 14 – 20 / 14 – 19 / 0 – 1 / 2 – 4 | 0 | 0 / 0 |
| smart2 | 0s / 0s | 3 | 4,777,415 – 5,239,110 | 0 | 1 – 5 | 1,438 – 1,891 | 32 – 38 / 31 – 37 / 3 – 6 / 4 – 8 | 0 | 0 / 0 |
| smart3 | 0s / 0s | 2 | 4,724,081 – 4,950,040 | 0 | 0 – 2 | 1,716 – 1,972 | 36 – 38 / 35 – 37 / 3 – 5 / 5 – 7 | 0 | 0 / 0 |
| smart6 | 0s / 0s | 3 | 4,930,060 – 5,979,085 | 0 | 0 – 6 | 1,419 – 1,780 | 33 – 38 / 32 – 37 / 3 – 7 / 5 – 8 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 6.63 – 8.42 s | +49 – +114 ms | 6.62 – 8.42 s | 0 ms | never | 192,506–203,090 | 0.54–0.57 / 3.10–3.30 / 11.90–12.16 | 0.48–0.54 / 2.78–3.49 / 10.88–15.74 | 1.96 – 2.00 s |
| f250 | 7.94 – 8.38 s | +45 – +251 ms | 7.91 – 8.38 s | 0 ms | never | 191,385–210,145 | 0.52–0.54 / 2.72–3.23 / 11.39–12.16 | 0.48–0.52 / 2.66–3.36 / 11.65–16.00 | 1.99 – 2.00 s |
| f50 | 7.17 – 8.17 s | +44 – +84 ms | 6.85 – 8.16 s | 0 ms | never | 172,852–203,055 | 0.54–0.60 / 3.10–3.49 / 11.65–12.16 | 0.47–0.57 / 2.66–3.30 / 12.16–14.98 | 1.93 – 2.00 s |
| main | 6.83 – 8.37 s | +25 – +469 ms | 6.81 – 8.35 s | 0 ms | never | 187,592–206,074 | 0.54–0.55 / 3.10–3.30 / 11.65–12.16 | 0.41–0.48 / 2.21–3.04 / 10.37–13.18 | 1.94 – 1.96 s |
| smart | 6.23 – 8.78 s | +238 – +572 ms | 5.97 – 8.77 s | 0 ms | never | 184,752–237,487 | 0.52–0.57 / 2.53–3.36 / 11.39–12.16 | 0.43–0.48 / 2.53–3.30 / 11.39–13.95 | 1.94 – 2.00 s |
| smart2 | 6.34 – 7.87 s | +18 – +64 ms | 6.34 – 7.87 s | 0 ms | never | 191,091–209,549 | 0.55 / 2.98–3.23 / 11.65–11.90 | 0.48–0.51 / 2.78–3.04 / 11.90–15.49 | 1.95 – 2.00 s |
| smart3 | 7.52 – 8.21 s | +54 – +60 ms | 7.24 – 8.21 s | 0 ms | never | 188,956–197,997 | 0.54–0.55 / 3.17–3.30 / 11.90 | 0.48–0.49 / 2.85–3.30 / 11.65–11.90 | 1.92 – 1.97 s |
| smart6 | 6.36 – 7.72 s | +38 – +54 ms | 6.34 – 7.72 s | 0 ms | never | 197,198–239,158 | 0.51–0.54 / 2.66–3.23 / 11.65–12.42 | 0.47–0.49 / 2.78–3.10 / 12.42–18.69 | 1.97 – 2.00 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f100-1 | 5,000,280 | 0 | 13 | 1,839 | 82 / 78 / 4 / 10 | 0 | 0 / 0 | 7.93 s | +79 ms | 3.10 / 3.49 | – | 1,744 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-2 | 4,814,763 | 0 | 6 | 2,062 | 86 / 83 / 2 / 10 | 0 | 0 / 0 | 8.42 s | +66 ms | 3.17 / 3.30 | – | 1,952 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-3 | 4,812,769 | 0 | 3 | 1,676 | 73 / 72 / 3 / 10 | 0 | 0 / 0 | 7.25 s | +114 ms | 3.30 / 2.78 | – | 1,590 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-4 | 5,077,380 | 0 | 17 | 1,460 | 67 / 66 / 3 / 10 | 0 | 0 / 0 | 6.63 s | +49 ms | 3.10 / 3.10 | – | 1,364 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-1 | 4,938,342 | 0 | 26 | 1,967 | 36 / 35 / 2 / 4 | 0 | 0 / 0 | 8.38 s | +245 ms | 3.17 / 2.66 | – | 1,874 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-2 | 5,253,700 | 0 | 14 | 1,781 | 34 / 33 / 2 / 4 | 0 | 0 / 0 | 7.94 s | +203 ms | 2.72 / 2.66 | – | 1,686 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-3 | 4,996,019 | 0 | 4 | 1,978 | 35 / 34 / 1 / 4 | 0 | 0 / 0 | 8.25 s | +45 ms | 2.91 / 2.91 | – | 1,870 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-4 | 4,785,126 | 0 | 29 | 1,862 | 35 / 34 / 2 / 4 | 0 | 0 / 0 | 8.23 s | +251 ms | 3.23 / 3.36 | – | 1,762 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-1 | 4,780,499 | 0 | 6 | 1,930 | 158 / 154 / 6 / 20 | 0 | 0 / 0 | 8.17 s | +44 ms | 3.42 / 2.91 | – | 1,824 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-2 | 5,076,462 | 0 | 2 | 1,783 | 150 / 146 / 5 / 20 | 0 | 0 / 0 | 7.66 s | +84 ms | 3.17 / 2.66 | – | 1,676 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-3 | 4,847,995 | 0 | 11 | 1,840 | 152 / 150 / 7 / 20 | 0 | 0 / 0 | 8.03 s | +55 ms | 3.10 / 3.23 | – | 1,738 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-4 | 4,321,349 | 0 | 0 | 1,717 | 143 / 135 / 6 / 20 | 0 | 0 / 0 | 7.17 s | +65 ms | 3.49 / 3.30 | – | 1,614 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-1 | 4,974,282 | 0 | 0 | 1,835 | 14 / 14 / 0 / 2 | 0 | 0 / 0 | 7.50 s | +469 ms | 3.30 / 2.21 | – | 1,728 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-2 | 4,690,416 | 0 | 0 | 2,060 | 20 / 20 / 0 / 4 | 0 | 0 / 0 | 8.37 s | +441 ms | 3.10 / 2.85 | – | 1,968 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-3 | 4,699,737 | 0 | 0 | 2,008 | 21 / 21 / 0 / 4 | 0 | 0 / 0 | 8.01 s | +25 ms | 3.10 / 3.04 | – | 1,904 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-4 | 5,152,056 | 0 | 0 | 1,695 | 18 / 18 / 0 / 3 | 0 | 0 / 0 | 6.83 s | +78 ms | 3.10 / 2.78 | – | 1,596 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart-1 | 5,076,433 | 0 | 1 | 1,765 | 20 / 19 / 1 / 4 | 0 | 0 / 0 | 7.66 s | +238 ms | 3.36 / 2.85 | – | 1,666 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart-2 | 5,937,237 | 0 | 0 | 1,380 | 14 / 14 / 0 / 2 | 0 | 0 / 0 | 6.23 s | +263 ms | 2.53 / 2.59 | – | 1,284 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart-3 | 4,638,115 | 0 | 0 | 2,120 | 20 / 19 / 0 / 3 | 0 | 0 / 0 | 8.78 s | +572 ms | 2.98 / 2.53 | – | 2,028 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart-4 | 4,618,860 | 0 | 0 | 1,887 | 16 / 16 / 0 / 4 | 0 | 0 / 0 | 7.76 s | +435 ms | 3.23 / 3.30 | – | 1,786 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-11 | 5,239,110 | 0 | 5 | 1,438 | 32 / 31 / 6 / 8 | 0 | 0 / 0 | 6.34 s | +51 ms | 2.98 / 3.04 | – | 1,330 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-12 | 4,837,107 | 0 | 5 | 1,772 | 38 / 37 / 5 / 8 | 0 | 0 / 0 | 7.80 s | +64 ms | 2.98 / 2.78 | – | 1,678 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-13 | 4,777,415 | 0 | 1 | 1,891 | 36 / 35 / 3 / 4 | 0 | 0 / 0 | 7.87 s | +18 ms | 3.23 / 2.91 | – | 1,796 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-21 | 4,950,040 | 0 | 0 | 1,716 | 36 / 35 / 5 / 7 | 0 | 0 / 0 | 7.52 s | +54 ms | 3.17 / 2.85 | – | 1,614 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-22 | 4,724,081 | 0 | 2 | 1,972 | 38 / 37 / 3 / 5 | 0 | 0 / 0 | 8.21 s | +60 ms | 3.30 / 3.30 | – | 1,876 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-1 | 4,930,060 | 0 | 4 | 1,780 | 38 / 37 / 5 / 8 | 0 | 0 / 0 | 7.72 s | +42 ms | 3.23 / 2.78 | – | 1,668 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-2 | 5,979,085 | 0 | 6 | 1,419 | 33 / 32 / 7 / 8 | 0 | 0 / 0 | 6.36 s | +38 ms | 2.66 / 3.10 | – | 1,310 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-3 | 4,949,251 | 0 | 0 | 1,686 | 33 / 32 / 3 / 5 | 0 | 0 / 0 | 7.14 s | +54 ms | 3.23 / 2.98 | – | 1,596 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-lazy

`MODE=kill HOLD_PROMOTION=3 ./fail-primary.sh 15` at t = 5 s; ShardsRefreshInterval 1s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 100ms / 5s, ShardsRefresh 1s | 3 | 4,079,952 – 4,311,207 | 0 | 265 – 325 | 1,568 – 1,769 | 110 – 119 / 103 – 112 / 30 / 10 | 0 | 0 / 0 |
| f250 | 250ms / 5s, ShardsRefresh 1s | 3 | 3,775,574 – 4,342,616 | 0 | 243 – 329 | 1,696 – 2,067 | 55 – 60 / 49 – 54 / 12 – 13 / 4 | 0 | 0 / 0 |
| f50 | 50ms / 5s, ShardsRefresh 1s | 3 | 4,031,380 – 4,408,798 | 0 | 259 – 289 | 1,605 – 1,948 | 197 – 224 / 188 – 216 / 58 – 59 / 20 | 0 | 0 / 0 |
| main | ShardsRefresh 1s | 3 | 3,912,123 – 4,557,527 | 432,354 – 544,561 | 0 – 2 | 1,732 – 2,123 | 2,277 – 3,066 / 2,273 – 3,062 / 2,226 – 3,019 / 1,066 – 1,139 | 0 | 0 / 0 |
| smart2 | 0s / 0s, ShardsRefresh 1s | 3 | 3,761,729 – 4,129,998 | 0 | 238 – 332 | 1,642 – 1,888 | 52 – 66 / 48 – 61 / 10 – 23 / 5 – 15 | 0 | 0 / 0 |
| smart3 | 0s / 0s, ShardsRefresh 1s | 3 | 4,018,086 – 4,843,155 | 0 | 279 – 331 | 1,331 – 1,953 | 85 – 91 / 81 – 87 / 46 – 47 / 19 | 0 | 0 / 0 |
| smart6 | 0s / 0s, ShardsRefresh 1s | 3 | 3,915,167 – 4,579,775 | 0 | 222 – 340 | 1,626 – 1,758 | 88 – 93 / 83 – 88 / 46 – 48 / 18 – 19 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f100 | 9.51 – 10.54 s | +36 – +109 ms | 9.51 – 10.54 s | 0 ms | never | 163,195–172,439 | 0.54–0.55 / 3.10–3.17 / 11.65–11.90 | 0.52–0.57 / 2.78–3.30 / 13.18–13.70 | 1.96 – 2.00 s |
| f250 | 10.06 – 11.33 s | +184 – +254 ms | 10.06 – 11.33 s | 0 ms | never | 151,020–173,695 | 0.54–0.55 / 2.98–3.17 / 11.90–12.67 | 0.51–0.55 / 3.10–3.30 / 12.93–15.49 | 1.99 – 2.00 s |
| f50 | 9.55 – 11.06 s | +44 – +63 ms | 9.51 – 11.05 s | 0 ms | never | 161,252–176,348 | 0.52–0.54 / 2.98–3.17 / 11.39–12.16 | 0.49–0.63 / 2.98–5.18 / 11.90–26.37 | 1.97 – 2.00 s |
| main | 9.61 – 11.36 s | +24 – +35 ms | 9.60 – 11.35 s | 0 ms | never | 156,480–182,297 | 0.54–0.55 / 2.91–3.10 / 12.16–12.42 | 0.51–0.52 / 2.85–3.62 / 7.87–13.18 | 20 – 35 ms |
| smart2 | 9.51 – 11.29 s | +69 – +599 ms | 9.49 – 11.28 s | 0 ms | never | 150,467–165,197 | 0.54–0.58 / 3.04–3.23 / 11.90–12.42 | 0.49–0.78 / 2.53–3.62 / 13.18–22.27 | 1.98 – 2.00 s |
| smart3 | 8.65 – 11.13 s | +96 – +140 ms | 8.62 – 11.11 s | 0 ms | never | 160,719–193,713 | 0.52–0.57 / 3.04–3.17 / 11.90–12.16 | 0.50–0.78 / 2.53–5.95 / 12.16–14.21 | 2.00 – 2.01 s |
| smart6 | 9.53 – 10.54 s | +33 – +104 ms | 9.49 – 10.53 s | 0 ms | never | 156,604–183,187 | 0.52–0.60 / 2.46–3.68 / 11.65–12.67 | 0.50–0.81 / 2.85–4.80 / 12.67–22.27 | 1.94 – 1.98 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f100-1 | 4,079,952 | 0 | 325 | 1,769 | 119 / 112 / 30 / 10 | 0 | 0 / 0 | 10.54 s | +109 ms | 3.10 / 2.78 | – | 1,674 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-2 | 4,283,511 | 0 | 322 | 1,714 | 116 / 109 / 30 / 10 | 0 | 0 / 0 | 10.24 s | +64 ms | 3.10 / 2.85 | – | 1,620 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f100-3 | 4,311,207 | 0 | 265 | 1,568 | 110 / 103 / 30 / 10 | 0 | 0 / 0 | 9.51 s | +36 ms | 3.17 / 3.30 | – | 1,480 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-1 | 4,342,616 | 0 | 329 | 1,696 | 56 / 50 / 13 / 4 | 0 | 0 / 0 | 10.30 s | +254 ms | 2.98 / 3.23 | – | 1,606 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-2 | 4,326,329 | 0 | 272 | 1,710 | 55 / 49 / 12 / 4 | 0 | 0 / 0 | 10.06 s | +185 ms | 3.04 / 3.10 | – | 1,614 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f250-3 | 3,775,574 | 0 | 243 | 2,067 | 60 / 54 / 12 / 4 | 0 | 0 / 0 | 11.33 s | +184 ms | 3.17 / 3.30 | – | 1,972 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-1 | 4,172,160 | 0 | 279 | 1,948 | 224 / 216 / 58 / 20 | 0 | 0 / 0 | 11.06 s | +44 ms | 2.98 / 2.98 | – | 1,854 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-2 | 4,408,798 | 0 | 259 | 1,605 | 197 / 188 / 59 / 20 | 0 | 0 / 0 | 9.55 s | +63 ms | 3.17 / 3.23 | – | 1,506 × dial tcp 127.0.0.1:7101: connect: connection refused |
| f50-3 | 4,031,380 | 0 | 289 | 1,938 | 222 / 213 / 58 / 20 | 0 | 0 / 0 | 10.88 s | +60 ms | 3.17 / 5.18 | – | 1,824 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-1 | 3,912,123 | 488,342 | 0 | 2,123 | 2,992 / 2,988 / 2,954 / 1,139 | 0 | 0 / 0 | 11.36 s | +32 ms | 3.10 / 3.62 | – | 2,020 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-2 | 3,949,360 | 544,561 | 0 | 2,064 | 3,066 / 3,062 / 3,019 / 1,096 | 0 | 0 / 0 | 11.08 s | +24 ms | 3.04 / 2.85 | – | 1,964 × dial tcp 127.0.0.1:7101: connect: connection refused |
| main-3 | 4,557,527 | 432,354 | 2 | 1,732 | 2,277 / 2,273 / 2,226 / 1,066 | 0 | 0 / 0 | 9.61 s | +35 ms | 2.91 / 2.98 | – | 1,621 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-1 | 3,761,729 | 0 | 306 | 1,846 | 66 / 61 / 22 / 15 | 0 | 0 / 0 | 10.53 s | +69 ms | 3.23 / 3.62 | – | 1,744 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-2 | 4,010,359 | 0 | 332 | 1,888 | 52 / 48 / 10 / 5 | 0 | 0 / 0 | 11.29 s | +599 ms | 3.04 / 2.53 | – | 1,778 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart2-3 | 4,129,998 | 0 | 238 | 1,642 | 62 / 58 / 23 / 10 | 0 | 0 / 0 | 9.51 s | +85 ms | 3.23 / 3.36 | – | 1,536 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-11 | 4,843,155 | 0 | 305 | 1,331 | 85 / 81 / 47 / 19 | 0 | 0 / 0 | 8.65 s | +96 ms | 3.04 / 2.53 | – | 1,238 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-12 | 4,035,574 | 0 | 279 | 1,953 | 91 / 87 / 46 / 19 | 0 | 0 / 0 | 11.13 s | +110 ms | 3.10 / 3.04 | – | 1,854 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart3-13 | 4,018,086 | 0 | 331 | 1,746 | 89 / 85 / 47 / 19 | 0 | 0 / 0 | 10.50 s | +140 ms | 3.17 / 5.95 | – | 1,643 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-1 | 4,105,972 | 0 | 222 | 1,669 | 88 / 83 / 46 / 19 | 0 | 0 / 0 | 9.53 s | +34 ms | 3.68 / 4.16 | – | 1,562 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-2 | 4,579,775 | 0 | 340 | 1,758 | 93 / 88 / 48 / 19 | 0 | 0 / 0 | 10.54 s | +104 ms | 2.46 / 2.85 | – | 1,661 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-3 | 3,915,167 | 0 | 243 | 1,626 | 89 / 84 / 48 / 18 | 0 | 0 / 0 | 9.56 s | +33 ms | 3.42 / 4.80 | – | 1,520 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-migrate

Atomic Slot Migration of the workload's slots from 127.0.0.1:7101 to 127.0.0.1:7103 at t = 5 s; the run ends 10 s after the migration succeeded; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 1 | 4,053,502 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 49,663 / 0 |
| smart6 | 0s / 0s | 1 | 4,836,985 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 344 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, migration window (ms) | Longest success |
|---|---|---|---|---|
| main | 266,562 | 0.57 / 3.42 / 12.16 | 0.62 / 3.49 / 11.90 | 23 ms |
| smart6 | 318,040 | 0.52 / 2.53 / 11.90 | 0.54 / 2.59 / 12.93 | 63 ms |

Migration. Times from MIGRATESLOTS: source (EXPORT) and target (IMPORT) report success (CLUSTER GETSLOTMIGRATIONS polled every 5 ms), the watcher sees the target own the slot. The handover is the first success report. From the handover: "traffic to the source until" is the end of the last 10 ms in which the client still wrote workload traffic (> 512 bytes) to the source, i.e. when all migrated slots routed to the new owner; "last MOVED" is the last INFO errorstats read of the source (every 20 ms) in which its MOVED count grew.

| Build | Source success | Target success | Handover seen (watcher) | Traffic to the source until (after handover) | Last MOVED from the source (after handover) | CLUSTER GETSLOTMIGRATIONS sent (run) | Off-endpoint CLUSTER SHARDS (by address, all runs) | Errors |
|---|---|---|---|---|---|---|---|---|
| main | 198 ms | 188 ms | 198 ms | 320 ms | 334 ms | 0 | 0 | – |
| smart6 | 202 ms | 194 ms | 199 ms | 23 ms | 25 ms | 0 | 0 | – |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 4,053,502 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 49,663 / 0 | 1 ms | -198 ms | 3.42 / 3.49 | – | – |
| smart6-1 | 4,836,985 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 344 / 0 | 1 ms | -198 ms | 2.53 / 2.59 | – | – |

</details>

## smart-pause

`MODE=pause ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 1 | 4,926,132 | 0 | 1,099 | 101 | 3 / 3 / 0 / 1 | 0 | 201 / 0 |
| smart6 | 0s / 0s | 2 | 5,494,848 – 6,393,418 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 12.09 s | +4883 ms | 12.09 s | 4.89 s | 4.97 s | 164,204 | 0.57 / 3.23 / 12.16 | 0.51 / 3.23 / 11.90 | 224 ms |
| smart6 | 9.98 s | +2.64 – +2.66 s | 9.98 s | 639 – 667 ms | never | 183,156–213,110 | 0.52–0.57 / 2.59–3.36 / 11.65–12.67 | 0.48–0.51 / 2.40–2.66 / 11.65–11.90 | 26 – 29 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 4,926,132 | 0 | 1,099 | 101 | 3 / 3 / 0 / 1 | 0 | 201 / 0 | 12.09 s | +4883 ms | 3.23 / 3.23 | – | 101 × i/o timeout |
| smart6-1 | 5,494,848 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +2662 ms | 3.36 / 2.66 | – | – |
| smart6-2 | 6,393,418 | 0 | 1,000 | 0 | 8 / 8 / 0 / 1 | 0 | 0 / 0 | 9.98 s | +2636 ms | 2.59 / 2.40 | – | – |

</details>

## smart-pipe-kill

`MODE=kill ./fail-primary.sh 12` at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 1 | 10,433,729 | 0 | 8,730 | 6,261 | 17 / 17 / 1 / 3 | 0 | 0 / 0 |
| smart6 | 0s / 0s | 1 | 11,499,169 | 0 | 8,590 | 6,061 | 27 / 26 / 0 / 4 | 0 | 0 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| main | 8.28 s | +394 ms | 7.86 s | 0 ms | never | 417,340 | 2.78 / 13.70 / 21.76 | 2.59 / 13.70 / 430.08 | 438 ms |
| smart6 | 7.70 s | +91 ms | 7.77 s | 0 ms | never | 459,960 | 2.59 / 13.44 / 22.27 | 2.59 / 14.46 / 19.20 | 2.01 s |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 10,433,729 | 0 | 8,730 | 6,261 | 17 / 17 / 1 / 3 | 0 | 0 / 0 | 8.28 s | +394 ms | 13.70 / 13.70 | – | 5,270 × dial tcp 127.0.0.1:7101: connect: connection refused |
| smart6-1 | 11,499,169 | 0 | 8,590 | 6,061 | 27 / 26 / 0 / 4 | 0 | 0 / 0 | 7.70 s | +91 ms | 13.44 / 14.46 | – | 5,110 × dial tcp 127.0.0.1:7101: connect: connection refused |

</details>

## smart-planned

planned failover (`CLUSTER FAILOVER` on the replica of slot 0's primary) at t = 5 s; ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / after failure / in no-primary window / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| f250 | 250ms / 5s | 2 | 5,494,207 – 5,533,012 | 0 | 0 | 0 | 2 – 3 / 2 – 3 / 0 / 1 – 2 | 0 | 926 – 992 / 0 |
| f50 | 50ms / 5s | 2 | 5,351,721 – 5,467,593 | 0 | 0 | 0 | 3 – 4 / 3 – 4 / 0 / 2 | 0 | 873 – 1,000 / 0 |
| main | – | 2 | 5,290,559 – 5,473,050 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 28,436 – 44,702 / 0 |
| smart3 | 0s / 0s | 2 | 5,358,550 – 5,387,569 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,066 – 1,086 / 0 |
| smart6 | 0s / 0s | 2 | 5,280,984 – 6,018,890 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,067 – 1,072 / 0 |

| Build | Failure → first success | Promotion → first success | Failure → last error | Traffic to the old primary until (after promotion) | Its last MOVED (after promotion) | ops/s | Latency p50 / p99 / p999, whole run (ms) | Latency p50 / p99 / p999, failure window (ms) | Longest success |
|---|---|---|---|---|---|---|---|---|---|
| f250 | 0 – 1 ms | -127 – -108 ms | never | 0 ms | 22 – 43 ms | 274,694–276,648 | 0.57 / 3.17–3.23 / 11.90–12.16 | 0.55–0.60 / 3.49–3.55 / 12.93–14.21 | 56 – 58 ms |
| f50 | 0 ms | -52 – -39 ms | never | 26 – 30 ms | 99 – 104 ms | 267,582–273,379 | 0.57–0.58 / 3.23–3.30 / 12.16–12.42 | 0.60–0.62 / 3.17–3.42 / 12.42–12.93 | 37 – 57 ms |
| main | 0 ms | -126 – -67 ms | never | 196 – 328 ms | 278 – 425 ms | 264,523–273,651 | 0.57–0.58 / 3.23–3.30 / 12.16–12.42 | 0.66–0.68 / 3.81–4.67 / 12.42–15.23 | 60 – 68 ms |
| smart3 | 0 – 2 ms | -62 – -47 ms | never | 18 – 29 ms | 82 – 100 ms | 267,923–269,377 | 0.58 / 3.30 / 12.42 | 0.60–0.62 / 3.30–3.49 / 13.44–16.00 | 37 – 61 ms |
| smart6 | 0 – 1 ms | -120 – -72 ms | never | 0 – 19 ms | 20 – 81 ms | 264,043–300,906 | 0.55–0.58 / 2.59–3.36 / 11.90–12.67 | 0.58–0.60 / 2.40–3.42 / 16.64–22.27 | 55 – 63 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| f250-1 | 5,494,207 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 992 / 0 | 1 ms | -127 ms | 3.23 / 3.55 | – | – |
| f250-2 | 5,533,012 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 926 / 0 | 0 ms | -108 ms | 3.17 / 3.49 | – | – |
| f50-1 | 5,351,721 | 0 | 0 | 0 | 4 / 4 / 0 / 2 | 0 | 1,000 / 0 | 0 ms | -39 ms | 3.30 / 3.42 | – | – |
| f50-2 | 5,467,593 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 873 / 0 | 0 ms | -52 ms | 3.23 / 3.17 | – | – |
| main-1 | 5,290,559 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 28,436 / 0 | 0 ms | -67 ms | 3.30 / 3.81 | – | – |
| main-2 | 5,473,050 | 0 | 0 | 0 | 3 / 3 / 0 / 2 | 0 | 44,702 / 0 | 0 ms | -126 ms | 3.23 / 4.67 | – | – |
| smart3-1 | 5,387,569 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,066 / 0 | 0 ms | -62 ms | 3.30 / 3.30 | – | – |
| smart3-2 | 5,358,550 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,086 / 0 | 2 ms | -47 ms | 3.30 / 3.49 | – | – |
| smart6-1 | 6,018,890 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,072 / 0 | 0 ms | -72 ms | 2.59 / 2.40 | – | – |
| smart6-2 | 5,280,984 | 0 | 0 | 0 | 2 / 2 / 0 / 1 | 0 | 1,067 / 0 | 1 ms | -120 ms | 3.36 / 3.42 | – | – |

</details>

## smart-steady

no failure (steady state); ShardsRefreshInterval 10s. CLUSTER SHARDS columns count what the client wrote (all addresses); "off-endpoint" is the part sent to other addresses than the discovery endpoint. MOVED / ASK are the error replies of all nodes (INFO errorstats, before → after the run).

| Build | Settings | Runs | Successes | ErrNoSlot | Ctx deadline | Other errors | CLUSTER SHARDS total / peak per s | Off-endpoint CLUSTER SHARDS | MOVED / ASK |
|---|---|---|---|---|---|---|---|---|---|
| main | – | 1 | 5,222,186 | 0 | 0 | 0 | 1 / 1 | 0 | 0 / 0 |
| smart6 | 0s / 0s | 1 | 6,209,971 | 0 | 0 | 0 | 1 / 1 | 0 | 0 / 0 |

| Build | ops/s | Latency p50 / p99 / p999, whole run (ms) | Longest success |
|---|---|---|---|
| main | 261,105 | 0.58 / 3.49 / 12.16 | 26 ms |
| smart6 | 310,459 | 0.54 / 2.53 / 11.90 | 26 ms |

<details><summary>Per run</summary>

| Run | Successes | ErrNoSlot | Ctx deadline | Other | CLUSTER SHARDS run / after failure / in window / peak per s | Off-endpoint | MOVED / ASK | Failure → first success | Promotion → first success | p99 run / window (ms) | Other server errors | Top other error |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| main-1 | 5,222,186 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 3.49 / 0.00 | – | – |
| smart6-1 | 6,209,971 | 0 | 0 | 0 | 1 / 0 / 0 / 1 | 0 | 0 / 0 | never | – | 2.53 / 0.00 | – | – |

</details>

## No-primary window and promotion, per run

From the watcher (CLUSTER SHARDS on every node every 100 ms, on the replica every 20 ms) and the fail script's lines; times from the failure. "Discovery node": where HAProxy sent the client's long-lived discovery-endpoint connection(s).

| Scenario | Run | Old primary flagged fail | No online primary from | No-primary window | New primary online (watcher) | PROMOTE line | Client CLUSTER SHARDS in window | Discovery node |
|---|---|---|---|---|---|---|---|---|
| smart-big-kill | lmain-1 | 6.41 s | 6.41 s | 300 ms | 6.71 s | never | 0 | 7102 |
| smart-big-kill | lmain-2 | 6.28 s | 6.28 s | 378 ms | 6.66 s | never | 0 | 7102 |
| smart-big-kill | v3big-1 | 6.90 s | 6.90 s | 339 ms | 7.24 s | never | 2 | 7102 |
| smart-big-kill | v3big-2 | 5.88 s | 5.88 s | 461 ms | 6.34 s | never | 2 | 7102 |
| smart-big-kill | v4big-1 | 7.07 s | 7.07 s | 280 ms | 7.35 s | never | 0 | 7102 |
| smart-big-kill | v4big-2 | 6.38 s | 6.38 s | 381 ms | 6.74 s | never | 0 | 7102 |
| smart-big-kill | v5big-1 | 5.36 s | 5.36 s | 380 ms | 5.70 s | never | 0 | 7102 |
| smart-big-kill | v5big-2 | 8.37 s | 8.37 s | 261 ms | 8.63 s | never | 0 | 7102 |
| smart-big-kill | v6big-1 | 6.82 s | 6.82 s | 200 ms | 7.02 s | never | 0 | 7102 |
| smart-big-kill | v6big-2 | 7.03 s | 7.03 s | 219 ms | 7.25 s | never | 1 | 7102 |
| smart-big-lazy | lmain-1 | 6.76 s | 6.76 s | 3.02 s | 9.78 s | 9.77 s | 0 | 7102 |
| smart-big-lazy | lmain-2 | 7.64 s | 7.64 s | 3.08 s | 10.72 s | 10.69 s | 0 | 7102 |
| smart-big-lazy | v3big-1 | 7.15 s | 7.15 s | 3.14 s | 10.29 s | 10.28 s | 0 | 7102 |
| smart-big-lazy | v3big-2 | 7.05 s | 7.05 s | 3.08 s | 10.13 s | 10.12 s | 0 | 7102 |
| smart-big-lazy | v3big-3 | 8.05 s | 8.05 s | 3.10 s | 11.15 s | 11.14 s | 4 | 7102 |
| smart-big-lazy | v4big-1 | 6.74 s | 6.74 s | 3.14 s | 9.88 s | 9.86 s | 0 | 7102 |
| smart-big-lazy | v4big-2 | 6.20 s | 6.20 s | 3.08 s | 9.28 s | 9.26 s | 0 | 7102 |
| smart-big-lazy | v4big-3 | 6.98 s | 6.98 s | 3.04 s | 10.02 s | 10.01 s | 0 | 7102 |
| smart-big-lazy | v5big-1 | 7.29 s | 7.29 s | 3.04 s | 10.33 s | 10.31 s | 11 | 7102 |
| smart-big-lazy | v5big-2 | 7.56 s | 7.56 s | 3.06 s | 10.62 s | 10.60 s | 12 | 7102 |
| smart-big-lazy | v5big-3 | 8.20 s | 8.20 s | 3.14 s | 11.34 s | 11.31 s | 11 | 7102 |
| smart-big-lazy | v6big-1 | 6.11 s | 6.11 s | 3.16 s | 9.27 s | 9.25 s | 10 | 7102 |
| smart-big-lazy | v6big-2 | 6.18 s | 6.18 s | 3.06 s | 9.24 s | 9.21 s | 11 | 7102 |
| smart-big-lazy | v6big-3 | 7.83 s | 7.83 s | 3.06 s | 10.89 s | 10.88 s | 12 | 7102 |
| smart-elect5 | f100-1 | 7.15 s | 7.15 s | 5.12 s | 12.25 s | 12.24 s | 50 | 7102 |
| smart-elect5 | f100-2 | 7.83 s | 7.83 s | 5.08 s | 12.89 s | 12.88 s | 49 | 7102 |
| smart-elect5 | f250-1 | 6.95 s | 6.95 s | 5.06 s | 12.01 s | 12.00 s | 20 | 7102 |
| smart-elect5 | f250-2 | 6.71 s | 6.71 s | 5.06 s | 11.77 s | 11.75 s | 20 | 7102 |
| smart-elect5 | f50-1 | 7.09 s | 7.09 s | 5.10 s | 12.15 s | 12.14 s | 97 | 7102 |
| smart-elect5 | f50-2 | 6.14 s | 6.14 s | 5.06 s | 11.16 s | 11.15 s | 95 | 7102 |
| smart-elect5 | main-1 | 7.46 s | 7.46 s | 5.04 s | 12.50 s | 12.48 s | 4,515 | 7102 |
| smart-elect5 | main-2 | 7.46 s | 7.46 s | 5.06 s | 12.52 s | 12.51 s | 4,874 | 7102 |
| smart-elect5 | smart3-1 | 7.74 s | 7.74 s | 5.04 s | 12.79 s | 12.77 s | 62 | 7102 |
| smart-elect5 | smart3-2 | 7.63 s | 7.63 s | 5.06 s | 12.65 s | 12.65 s | 61 | 7102 |
| smart-elect5 | smart6-1 | 6.73 s | 6.73 s | 5.04 s | 11.77 s | 11.77 s | 62 | 7102 |
| smart-elect5 | smart6-2 | 7.37 s | 7.37 s | 5.04 s | 12.39 s | 12.39 s | 63 | 7102 |
| smart-fleet | f_f100-1 | 7.21 s | 7.21 s | 340 ms | 7.55 s | never | 60 | 7101+7101+7101+7103+7102+7104+7103+7102+7105+7105+7102+7105+7103+7106+7106+7106+7104+7104+7104 |
| smart-fleet | f_f100-2 | 7.31 s | 7.31 s | 359 ms | 7.67 s | never | 68 | 7101+7101+7101+7105+7105+7104+7102+7105+7103+7103+7104+7104+7103+7104+7102+7103+7102+7106+7106+7106 |
| smart-fleet | f_f250-1 | 6.34 s | 6.34 s | 280 ms | 6.62 s | never | 20 | 7101+7101+7101+7102+7104+7105+7103+7102+7105+7105+7104+7103+7105+7103+7102+7106+7106+7106+7106+7103 |
| smart-fleet | f_f250-2 | 7.56 s | 7.56 s | 361 ms | 7.92 s | never | 27 | 7101+7101+7101+7104+7106+7103+7106+7102+7103+7102+7103+7106+7102+7104+7102+7104+7106+7105+7105+7105+7105 |
| smart-fleet | f_f50-1 | 7.28 s | 7.28 s | 400 ms | 7.68 s | never | 148 | 7101+7101+7101+7103+7102+7105+7106+7106+7106+7105+7102+7104+7102+7103+7106+7102+7103+7105+7105+7104+7104+7104 |
| smart-fleet | f_f50-2 | 7.87 s | 7.87 s | 221 ms | 8.09 s | never | 84 | 7101+7101+7101+7104+7103+7105+7102+7103+7105+7104+7106+7106+7105+7103+7106+7106+7102 |
| smart-fleet | f_main-1 | 6.55 s | 6.55 s | 320 ms | 6.85 s | never | 1,260 | 7101+7101+7101+7103+7103+7105+7102+7105+7104+7103+7102+7104+7104+7104+7105+7105+7103+7106+7106+7106+7106 |
| smart-fleet | f_main-2 | 7.39 s | 7.39 s | 239 ms | 7.63 s | never | 785 | 7101+7101+7101+7103+7104+7102+7104+7106+7104+7104+7105+7102+7106+7103+7103+7106+7102+7106+7102+7105+7105+7105 |
| smart-fleet | f_smart3-1 | 6.87 s | 6.87 s | 240 ms | 7.11 s | never | 44 | 7101+7101+7101+7104+7106+7102+7104+7103+7102+7105+7104+7102+7102+7106+7103+7103+7106+7106+7105+7105+7105 |
| smart-fleet | f_smart3-2 | 7.15 s | 7.15 s | 360 ms | 7.51 s | never | 86 | 7101+7101+7101+7105+7102+7103+7102+7106+7105+7106+7106+7102+7103+7105+7105+7106+7103+7103+7104+7104+7104+7104 |
| smart-fleet | f_smart6-1 | 6.83 s | 6.83 s | 379 ms | 7.21 s | never | 97 | 7101+7101+7101+7104+7102+7103+7103+7104+7104+7102+7102+7105+7103+7103+7106+7105+7102+7105+7106+7105 |
| smart-fleet | f_smart6-2 | 7.68 s | 7.68 s | 300 ms | 7.98 s | never | 87 | 7101+7101+7101+7104+7106+7106+7104+7103+7102+7102+7102+7105+7105+7105+7105+7104 |
| smart-hold3 | f100-1 | 6.85 s | 6.85 s | 3.06 s | 9.91 s | 9.90 s | 30 | 7102 |
| smart-hold3 | f100-2 | 7.25 s | 7.25 s | 3.08 s | 10.33 s | 10.30 s | 30 | 7102 |
| smart-hold3 | f100-3 | 7.00 s | 7.00 s | 3.08 s | 10.06 s | 10.04 s | 30 | 7102 |
| smart-hold3 | f250-1 | 7.81 s | 7.81 s | 3.08 s | 10.87 s | 10.86 s | 12 | 7102 |
| smart-hold3 | f250-2 | 7.46 s | 7.46 s | 3.04 s | 10.50 s | 10.49 s | 13 | 7102 |
| smart-hold3 | f250-3 | 6.64 s | 6.64 s | 3.06 s | 9.70 s | 9.69 s | 13 | 7102 |
| smart-hold3 | f50-1 | 6.24 s | 6.24 s | 3.06 s | 9.30 s | 9.28 s | 58 | 7102 |
| smart-hold3 | f50-2 | 7.02 s | 7.02 s | 3.06 s | 10.08 s | 10.07 s | 58 | 7102 |
| smart-hold3 | f50-3 | 7.83 s | 7.83 s | 3.08 s | 10.87 s | 10.85 s | 58 | 7102 |
| smart-hold3 | main-1 | 7.71 s | 7.71 s | 3.10 s | 10.81 s | 10.80 s | 2,794 | 7102 |
| smart-hold3 | main-2 | 8.25 s | 8.25 s | 3.08 s | 11.33 s | 11.32 s | 2,348 | 7102 |
| smart-hold3 | main-3 | 6.96 s | 6.96 s | 3.10 s | 10.06 s | 10.04 s | 3,183 | 7102 |
| smart-hold3 | smart2-1 | 7.43 s | 7.43 s | 3.04 s | 10.47 s | 10.45 s | 48 | 7102 |
| smart-hold3 | smart2-2 | 6.24 s | 6.24 s | 3.08 s | 9.32 s | 9.31 s | 47 | 7102 |
| smart-hold3 | smart2-3 | 6.15 s | 6.15 s | 3.10 s | 9.25 s | 9.24 s | 48 | 7102 |
| smart-hold3 | smart3-21 | 7.64 s | 7.64 s | 3.08 s | 10.72 s | 10.70 s | 48 | 7102 |
| smart-hold3 | smart6-1 | 7.14 s | 7.14 s | 3.08 s | 10.22 s | 10.21 s | 48 | 7102 |
| smart-hold3 | smart6-2 | 7.04 s | 7.04 s | 3.08 s | 10.12 s | 10.10 s | 48 | 7102 |
| smart-idle-kill | main-1 | 5.89 s | 5.89 s | 340 ms | 6.23 s | never | 0 | 7102 |
| smart-idle-kill | smart6-1 | 6.63 s | 6.63 s | 340 ms | 6.97 s | never | 0 | 7102 |
| smart-kill | f100-1 | 7.49 s | 7.49 s | 379 ms | 7.85 s | never | 4 | 7102 |
| smart-kill | f100-2 | 8.11 s | 8.11 s | 239 ms | 8.35 s | never | 2 | 7102 |
| smart-kill | f100-3 | 6.81 s | 6.81 s | 320 ms | 7.13 s | never | 3 | 7102 |
| smart-kill | f100-4 | 6.22 s | 6.22 s | 360 ms | 6.58 s | never | 3 | 7102 |
| smart-kill | f250-1 | 7.76 s | 7.76 s | 380 ms | 8.14 s | never | 2 | 7102 |
| smart-kill | f250-2 | 7.38 s | 7.38 s | 380 ms | 7.74 s | never | 2 | 7102 |
| smart-kill | f250-3 | 7.84 s | 7.84 s | 380 ms | 8.20 s | never | 1 | 7102 |
| smart-kill | f250-4 | 7.60 s | 7.60 s | 381 ms | 7.98 s | never | 2 | 7102 |
| smart-kill | f50-1 | 7.81 s | 7.81 s | 320 ms | 8.13 s | never | 6 | 7102 |
| smart-kill | f50-2 | 7.34 s | 7.34 s | 240 ms | 7.58 s | never | 5 | 7102 |
| smart-kill | f50-3 | 7.64 s | 7.64 s | 340 ms | 7.98 s | never | 7 | 7102 |
| smart-kill | f50-4 | 6.84 s | 6.84 s | 280 ms | 7.10 s | never | 6 | 7102 |
| smart-kill | main-1 | 6.66 s | 6.66 s | 361 ms | 7.03 s | never | 0 | 7102 |
| smart-kill | main-2 | 7.58 s | 7.58 s | 341 ms | 7.92 s | never | 0 | 7102 |
| smart-kill | main-3 | 7.63 s | 7.63 s | 380 ms | 7.99 s | never | 0 | 7102 |
| smart-kill | main-4 | 6.33 s | 6.33 s | 420 ms | 6.75 s | never | 0 | 7102 |
| smart-kill | smart-1 | 7.12 s | 7.12 s | 300 ms | 7.42 s | never | 1 | 7102 |
| smart-kill | smart-2 | 5.77 s | 5.77 s | 199 ms | 5.97 s | never | 0 | 7102 |
| smart-kill | smart-3 | 7.90 s | 7.90 s | 301 ms | 8.20 s | never | 0 | 7102 |
| smart-kill | smart-4 | 7.09 s | 7.09 s | 240 ms | 7.33 s | never | 0 | 7102 |
| smart-kill | smart2-11 | 5.95 s | 5.95 s | 339 ms | 6.29 s | never | 6 | 7102 |
| smart-kill | smart2-12 | 7.43 s | 7.43 s | 300 ms | 7.73 s | never | 5 | 7102 |
| smart-kill | smart2-13 | 7.54 s | 7.54 s | 320 ms | 7.86 s | never | 3 | 7102 |
| smart-kill | smart3-21 | 7.11 s | 7.11 s | 360 ms | 7.47 s | never | 5 | 7102 |
| smart-kill | smart3-22 | 7.81 s | 7.81 s | 339 ms | 8.15 s | never | 3 | 7102 |
| smart-kill | smart6-1 | 7.41 s | 7.41 s | 260 ms | 7.67 s | never | 5 | 7102 |
| smart-kill | smart6-2 | 5.94 s | 5.94 s | 381 ms | 6.32 s | never | 7 | 7102 |
| smart-kill | smart6-3 | 6.77 s | 6.77 s | 320 ms | 7.09 s | never | 3 | 7102 |
| smart-lazy | f100-1 | 7.33 s | 7.33 s | 3.10 s | 10.44 s | 10.41 s | 30 | 7102 |
| smart-lazy | f100-2 | 7.10 s | 7.10 s | 3.12 s | 10.18 s | 10.16 s | 30 | 7102 |
| smart-lazy | f100-3 | 6.40 s | 6.40 s | 3.08 s | 9.48 s | 9.46 s | 30 | 7102 |
| smart-lazy | f250-1 | 7.00 s | 7.00 s | 3.04 s | 10.04 s | 10.04 s | 13 | 7102 |
| smart-lazy | f250-2 | 6.80 s | 6.80 s | 3.12 s | 9.88 s | 9.86 s | 12 | 7102 |
| smart-lazy | f250-3 | 8.04 s | 8.04 s | 3.10 s | 11.14 s | 11.12 s | 12 | 7102 |
| smart-lazy | f50-1 | 7.95 s | 7.95 s | 3.06 s | 11.01 s | 11.01 s | 58 | 7102 |
| smart-lazy | f50-2 | 6.38 s | 6.38 s | 3.14 s | 9.48 s | 9.48 s | 59 | 7102 |
| smart-lazy | f50-3 | 7.76 s | 7.76 s | 3.06 s | 10.82 s | 10.80 s | 58 | 7102 |
| smart-lazy | main-1 | 8.25 s | 8.25 s | 3.10 s | 11.33 s | 11.31 s | 2,954 | 7102 |
| smart-lazy | main-2 | 7.97 s | 7.97 s | 3.08 s | 11.05 s | 11.03 s | 3,019 | 7102 |
| smart-lazy | main-3 | 6.50 s | 6.50 s | 3.08 s | 9.58 s | 9.56 s | 2,226 | 7102 |
| smart-lazy | smart2-1 | 7.40 s | 7.40 s | 3.08 s | 10.46 s | 10.44 s | 22 | 7102 |
| smart-lazy | smart2-2 | 7.66 s | 7.66 s | 3.06 s | 10.70 s | 10.67 s | 10 | 7102 |
| smart-lazy | smart2-3 | 6.36 s | 6.36 s | 3.11 s | 9.43 s | 9.41 s | 23 | 7102 |
| smart-lazy | smart3-11 | 5.51 s | 5.51 s | 3.04 s | 8.56 s | 8.53 s | 47 | 7102 |
| smart-lazy | smart3-12 | 7.92 s | 7.92 s | 3.10 s | 11.02 s | 10.99 s | 46 | 7102 |
| smart-lazy | smart3-13 | 7.32 s | 7.32 s | 3.04 s | 10.36 s | 10.34 s | 47 | 7102 |
| smart-lazy | smart6-1 | 6.43 s | 6.43 s | 3.06 s | 9.49 s | 9.46 s | 46 | 7102 |
| smart-lazy | smart6-2 | 7.35 s | 7.35 s | 3.08 s | 10.43 s | 10.42 s | 48 | 7102 |
| smart-lazy | smart6-3 | 6.43 s | 6.43 s | 3.12 s | 9.53 s | 9.50 s | 48 | 7102 |
| smart-pause | main-1 | 6.93 s | 6.93 s | 279 ms | 7.21 s | never | 0 | 7102 |
| smart-pause | smart6-1 | 7.00 s | 7.00 s | 320 ms | 7.32 s | never | 0 | 7102 |
| smart-pause | smart6-2 | 7.05 s | 7.05 s | 300 ms | 7.35 s | never | 0 | 7102 |
| smart-pipe-kill | main-1 | 7.56 s | 7.56 s | 322 ms | 7.88 s | never | 1 | 7102 |
| smart-pipe-kill | smart6-1 | 7.32 s | 7.32 s | 280 ms | 7.60 s | never | 0 | 7102 |
| smart-planned | f250-1 | never | never | never | 128 ms | 12 ms | 0 | 7102 |
| smart-planned | f250-2 | never | never | never | 109 ms | 13 ms | 0 | 7102 |
| smart-planned | f50-1 | never | never | never | 40 ms | 12 ms | 0 | 7102 |
| smart-planned | f50-2 | never | never | never | 53 ms | 13 ms | 0 | 7102 |
| smart-planned | main-1 | never | never | never | 68 ms | 14 ms | 0 | 7102 |
| smart-planned | main-2 | never | never | never | 126 ms | 14 ms | 0 | 7102 |
| smart-planned | smart3-1 | never | never | never | 62 ms | 11 ms | 0 | 7102 |
| smart-planned | smart3-2 | never | never | never | 48 ms | 12 ms | 0 | 7102 |
| smart-planned | smart6-1 | never | never | never | 73 ms | 13 ms | 0 | 7102 |
| smart-planned | smart6-2 | never | never | never | 120 ms | 13 ms | 0 | 7102 |

