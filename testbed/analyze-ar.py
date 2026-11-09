#!/usr/bin/env python3
"""Tables of the adaptive retry comparison (run-208-209.sh output made with make-queue-ar.sh).

    ./analyze-ar.py runs-ar > runs-ar/analysis.md
    ./analyze-ar.py runs-final > runs-final/analysis.md   # SET=final make-queue-vd.sh: migrate and idle tables too

Per scenario and build: medians and ranges of the bench's JSON fields, storms (runs with ErrNoSlot > 0),
CLUSTER SHARDS before / in the no-primary window, and for the recovery times a Mann-Whitney U test
(normal approximation with tie correction, two-sided) plus the Hodges-Lehmann shift between builds.
Standard library only.
"""
import glob
import json
import math
import os
import random
import statistics
import sys
from collections import defaultdict

ORDER = ["r2", "f_r2", "b2", "s8", "f_s8", "f_s8d1s", "f_s8d50", "b8", "b8d1s", "s7", "s7d250", "s7d50", "f_s7", "f_s7d250", "f_s7d50", "b7", "b7d1s", "b7d250", "v3big", "v4big", "v5big", "v6big", "smart6", "f_smart6", "smart5", "f_smart5", "smart3", "f_smart3", "smart2", "f_smart2", "smart", "f_smart", "f50", "f100", "f250", "f_f50", "f_f100", "f_f250", "f_main", "lmain", "l209", "lpr", "lgcp", "lbig", "main", "pr208", "pr209", "both", "fpr208", "fpr209", "fboth", "mainlong", "fmain", "vdasm", "fvdasm", "vdasmlong", "p50long", "pmain", "def", "trdef", "vd50", "fvd50", "p50", "fp50", "trp50", "i50", "p100", "i100", "fi100", "p250", "fp250",
         "gcp", "fgcp", "pgcp", "trgcp", "i500", "d10", "d200", "dauto", "w500", "w1s", "w2s",
         "vd0", "pace", "wait", "full", "un", "unvd", "pn", "pn208", "pnbd", "pnbdr"]
NAMES = {
    "smart": "Proposed defaults v1, nothing set",
    "smart2": "Proposed defaults v2, nothing set",
    "r2": "PR after the second review: no deadline refresh, no slot wait",
    "f_r2": "20 clients: PR after the second review",
    "b2": "1,000 nodes: PR after the second review",
    "s8": "Final commit, nothing set (MaxRefreshDelay 250 ms by default)",
    "f_s8d1s": "20 clients: final commit, MaxRefreshDelay 1 s",
    "f_s8d50": "20 clients: final commit, MaxRefreshDelay 50 ms",
    "b8": "1,000 nodes: final commit, nothing set",
    "b8d1s": "1,000 nodes: final commit, MaxRefreshDelay 1 s",
    "f_s8": "20 clients: final commit, nothing set",
    "s7": "Final: MaxRefreshDelay at its default (1s, 30s at 1,000 nodes)",
    "s7d250": "Final: MaxRefreshDelay 250 ms",
    "s7d50": "Final: MaxRefreshDelay 50 ms",
    "f_s7": "20 clients: final, MaxRefreshDelay default",
    "f_s7d250": "20 clients: final, MaxRefreshDelay 250 ms",
    "f_s7d50": "20 clients: final, MaxRefreshDelay 50 ms",
    "b7": "1,000 nodes: final, MaxRefreshDelay default (30 s)",
    "b7d1s": "1,000 nodes: final, MaxRefreshDelay 1 s",
    "b7d250": "1,000 nodes: final, MaxRefreshDelay 250 ms",
    "smart6": "Final candidate v6 (v3 + Expedite within the error jitter), nothing set",
    "smart5": "v5 (v3 + Expedite at once), nothing set",
    "smart3": "Proposed defaults v3, nothing set",
    "v3big": "1,000 nodes: proposed defaults v3",
    "v4big": "1,000 nodes: v3 + waiting requests expedite a delayed refresh",
    "v5big": "1,000 nodes: v4 + errors also expedite a delayed refresh",
    "v6big": "1,000 nodes: final candidate v6 (Expedite within the error jitter)",
    "f_smart6": "20 clients: final candidate v6",
    "f_smart5": "20 clients: v5",
    "f_smart3": "20 clients: proposed defaults v3",
    "f_smart2": "20 clients: proposed defaults v2",
    "f50": "Fixed 50 ms interval (delay 50 ms, wait 5 s)",
    "f100": "Fixed 100 ms interval (delay 50 ms, wait 5 s)",
    "f250": "Fixed 250 ms interval (delay 50 ms, wait 5 s)",
    "f_smart": "20 clients: proposed defaults",
    "f_f50": "20 clients: fixed 50 ms",
    "f_f100": "20 clients: fixed 100 ms",
    "f_f250": "20 clients: fixed 250 ms",
    "f_main": "20 clients: main",
    "pr208": "Main + #208 (MaxRefreshDelay 50 ms)",
    "pr209": "Main + #209 (MinRefreshInterval 50 ms, MaxSlotWait 5 s, default delay)",
    "both": "Main + #208 + #209 (50 ms / 50 ms / 5 s)",
    "fpr208": "20 clients: main + #208",
    "fpr209": "20 clients: main + #209",
    "fboth": "20 clients: both PRs",
    "l209": "1,000+ nodes: main + #209 (delay up to 30 s)",
    "vd50": "Vardhaman's PR #9 alone, MaxRefreshDelay 50 ms",
    "vdasm": "PR #9 + ASM changes: the PR with only MaxRefreshDelay 50 ms",
    "mainlong": "main, 10 s request deadlines",
    "vdasmlong": "PR #9 + ASM changes, 10 s request deadlines",
    "p50long": "PR, 50 ms / 50 ms / 8 s, 10 s request deadlines",
    "fvdasm": "20 clients: PR #9 + ASM changes",
    "lmain": "1,000+ nodes: main (delay up to 30 s)",
    "lpr": "1,000+ nodes: PR defaults (delay up to 30 s)",
    "lgcp": "1,000+ nodes: PR, 50 ms / 50 ms / 5 s",
    "lbig": "1,000+ nodes: PR, 1 s / 1 s / 5 s",
    "fvd50": "20 clients: Vardhaman's PR #9 alone, 50 ms",
    "p50": "PR, 50 ms / 50 ms / 5 s",
    "fp50": "20 clients: PR, 50 ms / 50 ms / 5 s",
    "trp50": "PR + refresh after deadlines, 50 ms / 50 ms / 5 s",
    "p100": "PR, 50 ms / 100 ms / 5 s",
    "p250": "PR, 50 ms / 250 ms / 5 s",
    "fp250": "20 clients: PR, 50 ms / 250 ms / 5 s",
    "fmain": "20 clients: main",
    "pmain": "Pipelines of 10: main",
    "pgcp": "Pipelines of 10: PR, 50 ms / 250 ms / 5 s",
    "fgcp": "20 clients: PR, 50 ms / 250 ms / 5 s",
    "fi100": "20 clients: PR, 50 ms / 100 ms / 5 s",
    "trdef": "PR defaults + refresh after deadlines",
    "trgcp": "PR + refresh after deadlines, 50 ms / 250 ms / 5 s",
    "i50": "MinRefreshInterval 50 ms", "i100": "MinRefreshInterval 100 ms", "i500": "MinRefreshInterval 500 ms",
    "d10": "MaxRefreshDelay 10 ms", "d200": "MaxRefreshDelay 200 ms", "dauto": "MaxRefreshDelay default (up to 1 s)",
    "w500": "MaxSlotWait 500 ms", "w1s": "MaxSlotWait 1 s", "w2s": "MaxSlotWait 2 s",
    "main": "Main today",
    "def": "PR, defaults (MOVED learning, no refresh on ASK)",
    "gcp": "PR, 50 ms / 250 ms / 5 s",
    "vd50": "Main + delay 50 ms (no 2.3)",
    "vd0": "Main + no delay (no 2.3)",
    "pace": "2.3 cap only + delay 50 ms",
    "wait": "2.3 wait only + delay 50 ms",
    "full": "Full 2.3 + delay 50 ms",
    "un": "Main today (no core change)",
    "unvd": "Main + MaxRefreshDelay 50 ms, no pacing (Vardhaman's draft)",
    "pn": "Baseline: core only",
    "pn208": "Option 1: PR #208 as is",
    "pnbd": "Option 2a: delay 50 ms",
    "pnbdr": "Option 2b: delay 50 ms + gated retry refresh",
}
SCENARIOS = ["idle-kill", "idle-planned", "kill-hold3-lazy", "kill", "kill-hold3", "kill-ab", "kill-hold3-ab", "planned", "migrate", "pause", "steady",
             "tune-kill", "tune-hold3", "tune-lazy", "tune-planned", "fleet-kill", "pipe-kill",
             "h2h-kill", "h2h-hold3", "h2h-lazy", "h2h-planned", "h2h-fleet", "h2h-pause", "h2h-idle-kill", "h2h-steady",
             "large-lazy", "large-kill", "large-pause", "large-idle-kill",
             "asm-kill", "asm-hold3", "asm-planned", "asm-migrate", "asm-fleet",
             "elect-2s", "elect-3s", "elect-5s",
             "split-kill", "split-hold3", "split-lazy", "split-planned", "split-migrate", "split-fleet",
             "split-large-kill", "split-large-lazy",
             "smart-kill", "smart-hold3", "smart-lazy", "smart-elect5", "smart-fleet", "smart-planned", "smart-big-lazy", "smart-big-kill",
             "smart-pause", "smart-migrate", "smart-idle-kill", "smart-pipe-kill", "smart-steady",
             "smart-big-pause", "smart-big-steady"]


def load(d):
    out = defaultdict(list)
    for f in glob.glob(os.path.join(d, "*.json")):
        try:
            r = json.load(open(f))
        except Exception as e:  # noqa: BLE001
            print(f"<!-- {f}: {e} -->")
            continue
        out[r["label"]].append(r)
    for rs in out.values():
        rs.sort(key=lambda r: r["run"])
    return out


def pre_window(r):
    t = r.get("topology") or {}
    start = t.get("no_primary_start_ms", -1)
    if start is None or start < 0 or r.get("fail_at_s", -1) < 0:
        return None, None
    times = r.get("cluster_shards_times_ms") or []
    n = sum(1 for x in times if 0 <= x < start)
    complete = len(times) < 300 or (times and times[-1] >= start)
    if not complete:
        fa = r["fail_at_s"]
        a, b = int(fa) + 1, int(fa + start / 1000)
        n = sum(row["cluster_shards"] for row in r["timeline"] if a <= row["t"] < b)
        return n, max(b - a, 0.001)
    return n, start / 1000


def vals(rs, f):
    out = []
    for r in rs:
        try:
            v = f(r)
        except (KeyError, TypeError):
            v = None
        if v is not None:
            out.append(v)
    return out


def med(v, fmt="{:,.0f}"):
    return fmt.format(statistics.median(v)) if v else "-"


def rng(v, fmt="{:,.0f}"):
    if not v:
        return "-"
    lo, hi = fmt.format(min(v)), fmt.format(max(v))
    return lo if lo == hi else f"{lo} – {hi}"


def medrng(v, fmt="{:,.0f}"):
    if not v:
        return "-"
    return f"{med(v, fmt)} ({rng(v, fmt)})"


def quart(v):
    if len(v) < 4:
        return None
    q = statistics.quantiles(v, n=4)
    return q[0], q[2]


def mann_whitney(x, y):
    """Two-sided p of the Mann-Whitney U test (normal approximation, tie-corrected), and U for x > y."""
    n1, n2 = len(x), len(y)
    if n1 < 3 or n2 < 3:
        return None, None
    allv = sorted([(v, 0) for v in x] + [(v, 1) for v in y])
    ranks = [0.0] * len(allv)
    i = 0
    ties = 0.0
    while i < len(allv):
        j = i
        while j + 1 < len(allv) and allv[j + 1][0] == allv[i][0]:
            j += 1
        r = (i + j) / 2 + 1
        for k in range(i, j + 1):
            ranks[k] = r
        t = j - i + 1
        ties += t ** 3 - t
        i = j + 1
    r1 = sum(rk for rk, (_, g) in zip(ranks, allv) if g == 0)
    u1 = r1 - n1 * (n1 + 1) / 2
    mu = n1 * n2 / 2
    n = n1 + n2
    sigma = math.sqrt(n1 * n2 / 12 * ((n + 1) - ties / (n * (n - 1))))
    if sigma == 0:
        return 1.0, u1
    z = (abs(u1 - mu) - 0.5) / sigma
    p = math.erfc(z / math.sqrt(2))
    return min(1.0, p), u1


def hodges_lehmann(x, y):
    d = sorted(a - b for a in x for b in y)
    return statistics.median(d) if d else None


def boot_ci(v, reps=4000, seed=1):
    if len(v) < 5:
        return None
    rnd = random.Random(seed)
    ms = sorted(statistics.median(rnd.choices(v, k=len(v))) for _ in range(reps))
    return ms[int(0.025 * reps)], ms[int(0.975 * reps) - 1]


def p2ok(r):
    return r.get("promotion_to_first_ok_ms")


def f2ok(r):
    v = r.get("first_ok_after_fail_ms", -1)
    return v if v is not None and v >= 0 else None


def old_dials(r):
    return (r.get("dials") or {}).get(r.get("old_primary"), 0)


def table_failure(sc, groups, labels):
    print(f"\n## {sc}\n")
    print("| Build | Runs | Storms | ErrNoSlot | Ctx deadlines | Other errors | CLUSTER SHARDS after failure | In no-primary window | Peak /s | Refreshes/s before window | Dials to the old primary |")
    print("|---|---|---|---|---|---|---|---|---|---|---|")
    for l in labels:
        rs = groups[l]
        rates = []
        for r in rs:
            n, s = pre_window(r)
            if n is not None and s:
                rates.append(n / s)
        storms = sum(1 for r in rs if r["err_no_slot"] > 0)
        print(f"| {NAMES.get(l, l)} | {len(rs)} | {storms} of {len(rs)} | {rng(vals(rs, lambda r: r['err_no_slot']))} "
              f"| {medrng(vals(rs, lambda r: r['err_deadline']))} | {medrng(vals(rs, lambda r: r['err_other']))} "
              f"| {medrng(vals(rs, lambda r: r['cluster_shards_after_fail']))} "
              f"| {medrng(vals(rs, lambda r: r['cluster_shards_in_no_primary_window']))} "
              f"| {rng(vals(rs, lambda r: r['cluster_shards_max_per_s']))} | {medrng(rates, '{:.1f}')} "
              f"| {medrng(vals(rs, old_dials))} |")
    print()
    print("| Build | Promotion → first success, median (IQR; min – max), ms | 95% CI of the median | Failure → first success, median (s) | Longest success, median (ms) | ops/s, median | p99 / p999, whole run, median (ms) |")
    print("|---|---|---|---|---|---|---|")
    for l in labels:
        rs = groups[l]
        v = vals(rs, p2ok)
        q = quart(v)
        ci = boot_ci(v)
        iq = f"{q[0]:.0f} – {q[1]:.0f}; " if q else ""
        print(f"| {NAMES.get(l, l)} | {med(v, '{:.0f}')} ({iq}{rng(v, '{:.0f}')}) | {f'{ci[0]:.0f} – {ci[1]:.0f}' if ci else '-'} "
              f"| {med([x / 1000 for x in vals(rs, f2ok)], '{:.2f}')} | {med(vals(rs, lambda r: r['longest_ok_ms']), '{:,.0f}')} "
              f"| {med(vals(rs, lambda r: r['ops_per_s']))} "
              f"| {med(vals(rs, lambda r: r['latency_ok_all']['p99_ms']), '{:.2f}')} / {med(vals(rs, lambda r: r['latency_ok_all']['p999_ms']), '{:.2f}')} |")
    pairs = [("main", "r2"), ("f_main", "f_r2"), ("lmain", "b2"), ("pn", "pn208"), ("pn", "pnbd"), ("pn", "pnbdr"), ("pn208", "pnbd"), ("pn208", "pnbdr"), ("pnbd", "pnbdr"), ("full", "gcp"),
             ("main", "vd50"), ("vd50", "p50"), ("main", "p50"), ("p50", "p100"), ("p100", "p250"), ("vd50", "p250"),
             ("vdasm", "p50"), ("vd50", "vdasm"), ("lmain", "lpr"), ("lpr", "lgcp"), ("lgcp", "lbig"), ("i50", "i100"), ("i100", "gcp"), ("gcp", "i500"), ("d10", "gcp"), ("gcp", "dauto"), ("main", "trdef"), ("gcp", "trgcp"),
             ("main", "pr208"), ("main", "pr209"), ("main", "both"), ("pr208", "pr209"), ("pr209", "both"), ("pr208", "both"),
             ("fmain", "fpr209"), ("fpr209", "fboth"), ("lmain", "l209"),
             ("f50", "smart"), ("f100", "smart"), ("f250", "smart"), ("main", "smart"), ("f_f50", "f_smart"), ("f_f250", "f_smart"), ("f50", "smart2"), ("f100", "smart2"), ("f250", "smart2"), ("main", "smart2"), ("smart", "smart2"), ("f_f50", "f_smart2"), ("f_f250", "f_smart2"), ("f50", "smart3"), ("f100", "smart3"), ("f250", "smart3"), ("main", "smart3"), ("smart2", "smart3"), ("f_f50", "f_smart3"), ("f_f250", "f_smart3"), ("v3big", "v4big"), ("lmain", "v3big"), ("lmain", "v4big"), ("v3big", "v5big"), ("v4big", "v5big"), ("lmain", "v5big"),
             ("smart3", "smart5"), ("main", "smart5"), ("f50", "smart5"), ("f100", "smart5"), ("f250", "smart5"), ("f_smart3", "f_smart5"), ("f_main", "f_smart5"), ("f_f50", "f_smart5"), ("f_f250", "f_smart5"),
             ("smart3", "smart6"), ("main", "smart6"), ("f50", "smart6"), ("f100", "smart6"), ("f250", "smart6"), ("f_smart3", "f_smart6"), ("f_main", "f_smart6"), ("f_f50", "f_smart6"), ("f_f250", "f_smart6"), ("v3big", "v6big"), ("v5big", "v6big"), ("lmain", "v6big")]
    rows = []
    for a, b in pairs:
        if a in groups and b in groups:
            x, y = vals(groups[a], p2ok), vals(groups[b], p2ok)
            p, _ = mann_whitney(x, y)
            hl = hodges_lehmann(x, y)
            if p is not None:
                rows.append(f"| {NAMES.get(a, a)} vs {NAMES.get(b, b)} | {hl:+.0f} | {p:.3g} |")
    if rows:
        print("\nPromotion → first success, pairwise (Hodges-Lehmann shift = typical difference first minus second; Mann-Whitney two-sided p):\n")
        print("| Comparison | Shift (ms) | p |")
        print("|---|---|---|")
        print("\n".join(rows))
    print("\n<details><summary>Per run: promotion → first success (ms)</summary>\n")
    for l in labels:
        print(f"- {l}: " + ", ".join(f"{x:.0f}" for x in vals(groups[l], p2ok)))
    print("\n</details>")


def table_planned(sc, groups, labels):
    print(f"\n## {sc}\n")
    print("| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Traffic to the old primary after the promotion, median (range), ms | Its last MOVED after the promotion, median (ms) | CLUSTER SHARDS after failover | ops/s, median |")
    print("|---|---|---|---|---|---|---|---|")
    for l in labels:
        rs = groups[l]
        errs = vals(rs, lambda r: r["err_no_slot"] + r["err_deadline"] + r["err_other"])
        lm = [x for x in vals(rs, lambda r: r.get("last_moved_old_primary_ms")) if x >= 0]
        print(f"| {NAMES.get(l, l)} | {len(rs)} | {rng(errs)} | {medrng(vals(rs, lambda r: r['server_moved']))} "
              f"| {medrng([x for x in vals(rs, lambda r: r.get('stale_route_ms')) if x >= 0])} | {med(lm)} "
              f"| {medrng(vals(rs, lambda r: r['cluster_shards_after_fail']))} | {med(vals(rs, lambda r: r['ops_per_s']))} |")


def table_steady(groups, labels):
    print("\n## steady\n")
    print("| Build | Runs | Errors | CLUSTER SHARDS during the run | ops/s, median (range) | p50 / p99 / p999, median (ms) |")
    print("|---|---|---|---|---|---|")
    for l in labels:
        rs = groups[l]
        errs = vals(rs, lambda r: r["err_no_slot"] + r["err_deadline"] + r["err_other"])
        print(f"| {NAMES.get(l, l)} | {len(rs)} | {rng(errs)} | {rng(vals(rs, lambda r: r['cluster_shards_run']))} "
              f"| {medrng(vals(rs, lambda r: r['ops_per_s']))} "
              f"| {med(vals(rs, lambda r: r['latency_ok_all']['p50_ms']), '{:.2f}')} / {med(vals(rs, lambda r: r['latency_ok_all']['p99_ms']), '{:.2f}')} / {med(vals(rs, lambda r: r['latency_ok_all']['p999_ms']), '{:.2f}')} |")


def table_migrate(sc, groups, labels):
    print(f"\n## {sc}\n")
    print("Atomic Slot Migration of slots 0-199 (all keys there) to another primary.\n")
    print("| Build | Runs | Errors (all kinds) | MOVED replies, median (range) | Client traffic to the source after the handover, median (range), ms | Last MOVED from the source after the handover, median (ms) | CLUSTER SHARDS during the run | Migration errors |")
    print("|---|---|---|---|---|---|---|---|")
    for l in labels:
        rs = groups[l]
        errs = vals(rs, lambda r: r["err_no_slot"] + r["err_deadline"] + r["err_other"])
        mig = [r.get("migration") or {} for r in rs]
        conv = [m["route_converged_ms"] for m in mig if (m.get("route_converged_ms") or -1) >= 0]
        last = [m["last_moved_ms"] for m in mig if (m.get("last_moved_ms") or -1) >= 0]
        print(f"| {NAMES.get(l, l)} | {len(rs)} | {rng(errs)} | {medrng(vals(rs, lambda r: r['server_moved']))} "
              f"| {medrng(conv)} | {med(last)} | {medrng(vals(rs, lambda r: r['cluster_shards_run']))} "
              f"| {sum(1 for m in mig if m.get('error'))} |")


def table_idle(sc, groups, labels):
    print(f"\n## {sc}\n")
    if sc.endswith("idle-kill"):
        print("The primary of the keys' shard crashes 1 s into an 18 s idle window and stays down past the resume. "
              "No ShardsRefreshInterval, so the client resumes with the dead primary in its slot map.\n")
    else:
        print("A planned failover of the keys' shard happens 2 s into a 12 s idle window. No ShardsRefreshInterval.\n")
    print("| Build | Runs | Failed after resume, median (range) | ErrNoSlot / deadline after resume | First success after resume, median (range), ms | Last error after resume, median (range), ms | MOVED in the first 5 s, median (range) | CLUSTER SHARDS in the first 5 s | Longest success, median (ms) |")
    print("|---|---|---|---|---|---|---|---|---|")
    for l in labels:
        rs = [r for r in groups[l] if r.get("resume")]
        if not rs:
            continue
        rr = [r["resume"] for r in rs]
        fok = [x["first_ok_after_resume_ms"] for x in rr if x["first_ok_after_resume_ms"] >= 0]
        lerr = [x["last_err_after_resume_ms"] for x in rr if x["last_err_after_resume_ms"] >= 0]
        print(f"| {NAMES.get(l, l)} | {len(rs)} | {medrng([x['err_other'] for x in rr])} "
              f"| {sum(x['err_no_slot'] for x in rr)} / {sum(x['err_deadline'] for x in rr)} "
              f"| {medrng(fok)} | {medrng(lerr, '{:,.1f}') if lerr else 'none'} "
              f"| {medrng([x['server_moved_first_5s'] for x in rr])} | {medrng([x['cluster_shards_first_5s'] for x in rr])} "
              f"| {med(vals(rs, lambda r: r['longest_ok_ms']), '{:,.0f}')} |")
    print("\n<details><summary>Per run</summary>\n")
    for l in labels:
        for r in groups[l]:
            x = r.get("resume") or {}
            print(f"- {l}-{r['run']}: failed {x.get('err_other')}, first success {x.get('first_ok_after_resume_ms')} ms, "
                  f"last error {x.get('last_err_after_resume_ms')} ms, MOVED {x.get('server_moved_first_5s')}, "
                  f"errors {[(e.get('msg'), e.get('count')) for e in (r.get('other_errors') or [])][:2]}")
    print("\n</details>")


def main():
    root = sys.argv[1]
    for sc in SCENARIOS:
        groups = load(os.path.join(root, sc))
        if not groups:
            continue
        labels = sorted(groups, key=lambda x: ORDER.index(x) if x in ORDER else 99)
        if sc.endswith("planned") and "idle" not in sc:
            print(f"\n<!-- {sc} -->")
            table_planned(sc, groups, labels)
        elif sc.endswith("steady"):
            print(f"\n<!-- {sc} -->")
            table_steady(groups, labels)
        elif sc.endswith("migrate"):
            table_migrate(sc, groups, labels)
        elif "idle-" in sc:
            table_idle(sc, groups, labels)
        else:
            table_failure(sc, groups, labels)


if __name__ == "__main__":
    main()

