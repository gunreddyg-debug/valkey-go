// Command bench runs a GET/SET workload through a valkey-go cluster client whose only InitAddress is
// the discovery endpoint, makes something happen to the shard of slot 0 during the run, and reports
// what the client saw: successes, ErrNoSlot, context deadlines, other errors, the CLUSTER SHARDS
// commands it sent (and where), the latency of successful requests, the longest one, and the time
// from the failure to the first success. A watcher polls CLUSTER SHARDS on every node directly to
// measure how long the shard had no online primary and when a new owner appeared, and INFO errorstats
// is read on every node to count the MOVED/ASK/... replies the servers sent.
//
// What happens at -fail-at (-fail-kind):
//
//	primary   fail-primary.sh: unplanned failure of slot 0's primary (-fail-mode, -fail-seconds, -fail-hold)
//	planned   planned-failover.sh: CLUSTER FAILOVER on the replica of slot 0's primary
//	migrate   Atomic Slot Migration of -migrate-slots from their owner to another primary; the run
//	          ends -after-migrate after the migration succeeded
//
// and nothing with -no-fail (steady state).
//
// Build: ./build.sh (bin/bench-upstream, bin/bench-pr1) and ./build-208-209.sh (bin/bench-<tree>).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valkey-io/valkey-go"
)

var (
	addr        = flag.String("addr", "127.0.0.1:16379", "discovery endpoint: the client's only InitAddress")
	nodesFlag   = flag.String("nodes", "127.0.0.1:7101,127.0.0.1:7102,127.0.0.1:7103,127.0.0.1:7104,127.0.0.1:7105,127.0.0.1:7106", "cluster nodes, polled directly (not through the client) with CLUSTER SHARDS and INFO errorstats")
	goroutines  = flag.Int("goroutines", 200, "concurrent request loops")
	duration    = flag.Duration("duration", 40*time.Second, "length of the run (with -fail-kind migrate: the upper bound)")
	failAt      = flag.Duration("fail-at", 5*time.Second, "when the failure (or planned failover, or migration) starts, from the start of the run")
	failKind    = flag.String("fail-kind", "primary", "what happens at -fail-at: primary (fail-primary.sh), planned (planned-failover.sh) or migrate (Atomic Slot Migration of -migrate-slots)")
	failScript  = flag.String("fail-script", "", "the fail script of -fail-kind primary/planned (default: fail-primary.sh / planned-failover.sh two directories above this binary)")
	noFail      = flag.Bool("no-fail", false, "do nothing at -fail-at (steady-state baseline)")
	failSeconds = flag.Int("fail-seconds", 12, "argument of fail-primary.sh: how long the primary stays failed")
	failMode    = flag.String("fail-mode", "pause", "MODE of fail-primary.sh: pause (docker pause) or kill (docker kill, then start)")
	failHold    = flag.Int("fail-hold", 0, "HOLD_PROMOTION of fail-primary.sh: seconds to keep the replica from taking over after the primary is flagged fail (0: normal failover)")
	reqTimeout  = flag.Duration("timeout", 2*time.Second, "context timeout of each request")
	numKeys     = flag.Int("keys", 1000, "number of keys, all in the slots of the primary of slot 0 (or in -key-slots)")
	keySlots    = flag.String("key-slots", "", "put all keys in this slot range, FROM-TO (default: anywhere in the shard of slot 0)")
	pollEvery   = flag.Duration("poll", 100*time.Millisecond, "interval of the CLUSTER SHARDS watcher, per node")
	pollFast    = flag.Duration("poll-fast", 20*time.Millisecond, "interval of the CLUSTER SHARDS watcher on the nodes that can take over slot 0: its replicas, and the migration target")
	errorsPoll  = flag.Duration("errorstats-poll", 100*time.Millisecond, "interval of the INFO errorstats reads on every node during the run")
	label       = flag.String("label", "", "label of the run, for example A")
	runID       = flag.Int("run", 0, "number of the run")
	jsonOut     = flag.String("json", "", "write the result as JSON to this file")

	numClients = flag.Int("clients", 1, "number of valkey clients, like separate application instances behind the same discovery endpoint; the goroutines are spread over them")
	pipeline   = flag.Int("pipeline", 0, "send DoMulti batches of this many commands instead of single commands (0: single commands)")

	idleFrom  = flag.Duration("idle-from", 0, "the workers send nothing from this time (from the start of the run) until -idle-until, to see how an idle client resumes")
	idleUntil = flag.Duration("idle-until", 0, "end of the idle window (0: no idle window)")

	shardsRefresh = flag.Duration("shards-refresh", 10*time.Second, "ClusterOption.ShardsRefreshInterval: periodic topology refresh (0: off); 10s is the valkeygcp default")

	migrateSlots   = flag.String("migrate-slots", "0-199", "-fail-kind migrate: the slots to migrate, FROM-TO")
	migrateTo      = flag.String("migrate-to", "", "-fail-kind migrate: address of the target primary (default: the owner of slot 16383, or of 8192)")
	afterMigrate   = flag.Duration("after-migrate", 10*time.Second, "-fail-kind migrate: the run ends this long after the migration succeeded")
	migrateTimeout = flag.Duration("migrate-timeout", 30*time.Second, "-fail-kind migrate: give up on the migration after this long")
	migratePoll    = flag.Duration("migrate-poll", 5*time.Millisecond, "-fail-kind migrate: interval of CLUSTER GETSLOTMIGRATIONS on the source and the target")
	routeBytes     = flag.Int64("route-bytes", 512, "-fail-kind migrate: a 10 ms interval with more bytes than this written by the client to the source counts as workload traffic still routed there")
	windowTail     = flag.Duration("window-tail", time.Second, "the failure window of the latency percentiles ends this long after the recovery")
)

type timelineRow struct {
	T        int   `json:"t"` // start of the second, from t0
	OK       int64 `json:"ok"`
	NoSlot   int64 `json:"no_slot"`
	Deadline int64 `json:"deadline"`
	Other    int64 `json:"other"`
	Shards   int64 `json:"cluster_shards"`
	Moved    int64 `json:"server_moved"` // MOVED replies of all nodes, by INFO errorstats read time
	Ask      int64 `json:"server_ask"`
}

// resumeResult describes the requests started after the idle window (-idle-from, -idle-until).
type resumeResult struct {
	IdleFromS            float64 `json:"idle_from_s"`
	IdleUntilS           float64 `json:"idle_until_s"`
	OK                   int64   `json:"ok"`
	NoSlot               int64   `json:"err_no_slot"`
	Deadline             int64   `json:"err_deadline"`
	Other                int64   `json:"err_other"`
	FirstOKAfterResumeMs float64 `json:"first_ok_after_resume_ms"`  // from the end of the idle window; -1: none
	FirstErrAfterResume  float64 `json:"first_err_after_resume_ms"` // -1: none
	LastErrAfterResumeMs float64 `json:"last_err_after_resume_ms"`  // -1: none
	MovedFirst5s         int64   `json:"server_moved_first_5s"`     // MOVED replies of all nodes in the 5 s after the window, by INFO errorstats read time (1 s resolution)
	ShardsFirst5s        int64   `json:"cluster_shards_first_5s"`   // CLUSTER SHARDS the client sent in those 5 s (1 s resolution)
}

type result struct {
	Label       string            `json:"label"`
	Run         int               `json:"run"`
	Build       string            `json:"build"`
	Options     map[string]string `json:"options"`
	Args        []string          `json:"args"`
	FailKind    string            `json:"fail_kind"`
	FailMode    string            `json:"fail_mode"`
	FailSeconds int               `json:"fail_seconds"`
	FailHold    int               `json:"fail_hold_s"`
	Goroutines  int               `json:"goroutines"`
	DurationS   float64           `json:"duration_s"`
	RunLengthS  float64           `json:"run_length_s"` // from t0 to the end of the workload (with -fail-kind migrate it ends early)
	TimeoutMs   float64           `json:"timeout_ms"`
	Addr        string            `json:"addr"`
	OldPrimary  string            `json:"old_primary"`
	Slots       [][2]int64        `json:"slots"`
	KeySlots    string            `json:"key_slots,omitempty"`
	FailAtS     float64           `json:"fail_at_s"` // from t0, when the script reported the failure (or the migration started); -1: none
	RestoredAtS float64           `json:"restored_at_s"`

	Requests    int64        `json:"requests"`
	OK          int64        `json:"ok"`
	NoSlot      int64        `json:"err_no_slot"`
	Deadline    int64        `json:"err_deadline"`
	Other       int64        `json:"err_other"`
	OtherErrors []countEntry `json:"other_errors"`
	OpsPerSec   float64      `json:"ops_per_s"` // successes per second of RunLengthS

	LongestOKMs        float64 `json:"longest_ok_ms"`
	LongestOKStartS    float64 `json:"longest_ok_started_at_s"`
	FirstOKAfterFailMs float64 `json:"first_ok_after_fail_ms"` // first success of a request started after the failure; -1: none
	FirstErrAtS        float64 `json:"first_err_at_s"`         // -1: no errors
	LastErrAfterFailMs float64 `json:"last_err_after_fail_ms"` // -1: no errors
	NoSlotFromMs       float64 `json:"no_slot_from_ms"`        // first ErrNoSlot, from the failure; -1: none
	NoSlotToMs         float64 `json:"no_slot_to_ms"`          // last ErrNoSlot, from the failure; -1: none

	// Script lines and the promotion, from the failure; -1: not seen.
	FlaggedFailLineMs float64 `json:"flagged_fail_line_ms"` // FLAGGED_FAIL (HOLD_PROMOTION)
	PromoteLineMs     float64 `json:"promote_line_ms"`      // PROMOTE: CLUSTER FAILOVER [FORCE] answered
	PromotedLineMs    float64 `json:"promoted_line_ms"`     // PROMOTED (planned-failover.sh)
	FailedOverLineMs  float64 `json:"failed_over_line_ms"`  // FAILED_OVER (planned-failover.sh)
	PromotionMs       float64 `json:"promotion_ms"`         // the new owner of slot 0 is online: watcher, else PROMOTE line
	PromotionSource   string  `json:"promotion_source"`     // watcher, promote-line or ""
	// First success after the failure minus PromotionMs; can be negative by the watcher's resolution. null: n/a.
	PromotionToFirstOKMs *float64 `json:"promotion_to_first_ok_ms"`
	// From the promotion: end of the last 10 ms in which the client still wrote workload traffic (> -route-bytes)
	// to the old primary, i.e. until its slot map pointed elsewhere; 0: none after the promotion; -1: no promotion.
	StaleRouteMs float64 `json:"stale_route_ms"`
	// From the promotion: last INFO errorstats read in which the old primary's MOVED count grew; -1: none.
	LastMovedOldPrimaryMs float64 `json:"last_moved_old_primary_ms"`

	LatencyAll         latencySummary `json:"latency_ok_all"`            // successes of the whole run
	LatencyWindow      latencySummary `json:"latency_ok_failure_window"` // successes that ended in the failure window
	LatencyWindowFromS float64        `json:"latency_window_from_s"`     // from t0; -1: no failure
	LatencyWindowToS   float64        `json:"latency_window_to_s"`       // from t0

	ShardsInit                 int64            `json:"cluster_shards_init"`       // sent before t0, while creating the client
	ShardsRun                  int64            `json:"cluster_shards_run"`        // sent from t0 to the end
	ShardsAfterFail            int64            `json:"cluster_shards_after_fail"` // sent after the failure
	ShardsInWindow             int64            `json:"cluster_shards_in_no_primary_window"`
	ShardsMaxPerSec            int64            `json:"cluster_shards_max_per_s"`
	ShardsByAddr               map[string]int64 `json:"cluster_shards_by_addr"`               // whole run, including init
	ShardsTimesMs              []float64        `json:"cluster_shards_times_ms"`              // after the failure, from the failure (first 300)
	ShardsOffEndpoint          int64            `json:"cluster_shards_off_endpoint"`          // whole run, including init, to addresses other than -addr
	ShardsOffEndpointRun       int64            `json:"cluster_shards_off_endpoint_run"`      // from t0
	ShardsOffEndpointTimesMs   []float64        `json:"cluster_shards_off_endpoint_times_ms"` // from the failure (first 100)
	GetSlotMigrationsByAddr    map[string]int64 `json:"getslotmigrations_by_addr"`            // whole run, including init
	GetSlotMigrationsRun       int64            `json:"getslotmigrations_run"`                // from t0
	GetSlotMigrationsAfterFail int64            `json:"getslotmigrations_after_fail"`         //
	Dials                      map[string]int64 `json:"dials"`
	DialErrs                   map[string]int64 `json:"dial_errors"`
	BytesByAddr                map[string]int64 `json:"bytes_by_addr_run"` // bytes the client wrote per destination, from t0

	ServerErrors serverErrors `json:"server_errors"` // INFO errorstats of all nodes, before to after the run
	ServerMoved  int64        `json:"server_moved"`
	ServerAsk    int64        `json:"server_ask"`

	Migration *migrationResult `json:"migration,omitempty"`

	Resume *resumeResult `json:"resume,omitempty"` // with -idle-until

	Topology     topology      `json:"topology"`
	Timeline     []timelineRow `json:"timeline"`
	ScriptOutput []string      `json:"script_output"`
	ScriptError  string        `json:"script_error,omitempty"`
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "bench:", err)
		os.Exit(1)
	}
}

// applyBuildOptions sets the ClusterOption fields of the flags and returns them for the result.
func applyBuildOptions(opt *valkey.ClusterOption) map[string]string {
	m := map[string]string{}
	opt.ShardsRefreshInterval = *shardsRefresh
	m["ShardsRefreshInterval"] = opt.ShardsRefreshInterval.String()
	applyPR1Options(opt, m)
	applyASMOptions(opt, m)
	applyBDOptions(opt, m)
	applyVDOptions(opt, m)
	return m
}

func run() error {
	nodes := strings.Split(*nodesFlag, ",")
	kind := *failKind
	if *noFail {
		kind = "none"
	}
	switch kind {
	case "none", "primary", "planned", "migrate":
	default:
		return fmt.Errorf("unknown -fail-kind %q", kind)
	}
	script := *failScript
	if script == "" && (kind == "primary" || kind == "planned") {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		name := "fail-primary.sh"
		if kind == "planned" {
			name = "planned-failover.sh"
		}
		script = filepath.Join(filepath.Dir(exe), "..", "..", name)
	}

	// The shard of slot 0 (the one fail-primary.sh fails), asked from the nodes directly.
	var shards []shardInfo
	var err error
	for _, n := range nodes {
		if shards, err = fetchShards(n, time.Second); err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("CLUSTER SHARDS: %w (is the cluster up? ./up.sh)", err)
	}
	target, ok := shardOf(shards, 0)
	if !ok {
		return errors.New("no shard owns slot 0")
	}
	oldPrimary, ok := target.onlinePrimary()
	if !ok {
		return errors.New("the shard of slot 0 has no online primary; is the cluster healthy? (./up.sh)")
	}
	var keys []string
	if *keySlots != "" {
		lo, hi, err := parseSlotRange(*keySlots)
		if err != nil {
			return fmt.Errorf("-key-slots: %w", err)
		}
		keys, err = keysInSlots(*numKeys, lo, hi)
		if err != nil {
			return err
		}
	} else if keys, err = keysInShard(*numKeys, target); err != nil {
		return err
	}
	for _, k := range keys[:3] {
		v, err := roundTrip(oldPrimary, time.Second, encodeCommand("CLUSTER", "KEYSLOT", k))
		if err != nil {
			return fmt.Errorf("CLUSTER KEYSLOT: %w", err)
		}
		if s, _ := v.(int64); s != keySlot(k) {
			return fmt.Errorf("slot of %q: server says %v, bench computed %d", k, v, keySlot(k))
		}
	}

	// The nodes that can take over slot 0: polled faster by the watcher.
	fast := map[string]bool{}
	for _, n := range target.Nodes {
		if n.Addr != oldPrimary {
			fast[n.Addr] = true
		}
	}
	var migSrc, migDst shardNode
	var migLo, migHi int64
	watchSlot := int64(0)
	watchOld := oldPrimary
	if kind == "migrate" {
		if migLo, migHi, err = parseSlotRange(*migrateSlots); err != nil {
			return fmt.Errorf("-migrate-slots: %w", err)
		}
		src, ok := shardOf(shards, migLo)
		if !ok {
			return fmt.Errorf("no shard owns slot %d", migLo)
		}
		if migSrc, ok = src.primaryNode(); !ok {
			return fmt.Errorf("the shard of slot %d has no online primary", migLo)
		}
		for _, s := range shards {
			for _, r := range s.Ranges {
				if r[0] <= migHi && migLo <= r[1] && (s.Ranges[0] != src.Ranges[0]) {
					return fmt.Errorf("slots %d-%d are not all owned by %s", migLo, migHi, migSrc.Addr)
				}
			}
		}
		pick := func(slot int64) (shardNode, bool) {
			if s, ok := shardOf(shards, slot); ok {
				if p, ok := s.primaryNode(); ok && p.Addr != migSrc.Addr {
					return p, true
				}
			}
			return shardNode{}, false
		}
		if *migrateTo != "" {
			for _, s := range shards {
				if p, ok := s.primaryNode(); ok && p.Addr == *migrateTo {
					migDst = p
				}
			}
			if migDst.Addr == "" || migDst.Addr == migSrc.Addr {
				return fmt.Errorf("-migrate-to %s is not a primary other than %s", *migrateTo, migSrc.Addr)
			}
		} else if migDst, ok = pick(16383); !ok {
			if migDst, ok = pick(8192); !ok {
				return errors.New("no other primary to migrate to")
			}
		}
		if migDst.ID == "" {
			return fmt.Errorf("no node id for %s in CLUSTER SHARDS", migDst.Addr)
		}
		fast = map[string]bool{migDst.Addr: true}
		watchSlot, watchOld = migLo, migSrc.Addr
	}

	// The client under test.
	dc := newDialCounter()
	opt := valkey.ClientOption{
		InitAddress: []string{*addr},
		DialCtxFn:   dc.dial,
		ClusterOption: valkey.ClusterOption{
			PreferInitAddressRefresh: true,
			PreferClusterShards:      true,
		},
	}
	options := applyBuildOptions(&opt.ClusterOption)
	if *numClients > 1 {
		options["clients"] = strconv.Itoa(*numClients)
	}
	if *pipeline > 0 {
		options["pipeline"] = strconv.Itoa(*pipeline)
	}
	clients := make([]valkey.Client, 0, max(1, *numClients))
	closeAll := func() {
		for _, c := range clients {
			c.Close()
		}
	}
	for len(clients) < cap(clients) {
		c, err := valkey.NewClient(opt)
		if err != nil {
			closeAll()
			return fmt.Errorf("valkey.NewClient: %w", err)
		}
		clients = append(clients, c)
	}
	client := clients[0]
	closed := false
	defer func() {
		if !closed {
			closeAll()
		}
	}()

	value := strings.Repeat("v", 32)
	wctx, wcancel := context.WithTimeout(context.Background(), 10*time.Second)
	sets := make(valkey.Commands, 0, len(keys))
	for _, k := range keys {
		sets = append(sets, client.B().Set().Key(k).Value(value).Build())
	}
	for _, r := range client.DoMulti(wctx, sets...) {
		if err := r.Error(); err != nil {
			wcancel()
			return fmt.Errorf("warm-up SET: %w", err)
		}
	}
	wcancel()

	keyDesc := fmt.Sprintf("in slots %v of primary %s", target.Ranges, oldPrimary)
	if *keySlots != "" {
		keyDesc = fmt.Sprintf("in slots %s", *keySlots)
	}
	event := ""
	switch kind {
	case "none":
		event = "no failure"
	case "primary":
		event = fmt.Sprintf("%s at t=%v for %ds, promotion held %ds", *failMode, *failAt, *failSeconds, *failHold)
	case "planned":
		event = fmt.Sprintf("planned failover (CLUSTER FAILOVER) at t=%v", *failAt)
	case "migrate":
		event = fmt.Sprintf("migration of slots %d-%d from %s to %s at t=%v, run ends %v after its success", migLo, migHi, migSrc.Addr, migDst.Addr, *failAt, *afterMigrate)
	}
	fmt.Printf("bench: label=%s run=%d build=%s options=%v\n", *label, *runID, buildName, options)
	fmt.Printf("bench: %d goroutines, %v, %v per request, %d keys %s; %s\n",
		*goroutines, *duration, *reqTimeout, len(keys), keyDesc, event)

	// Server-side error counters, before the run.
	es := newErrStats(nodes, *errorsPoll)
	es.baseline()

	// The run.
	t0 := time.Now()
	dc.start(t0)
	es.t0 = t0
	st := newStats(t0, *duration+*reqTimeout+5*time.Second)
	if *idleUntil > *idleFrom {
		st.resumeAt = int64(*idleUntil)
	}
	w := &watcher{t0: t0, nodes: nodes, slot: watchSlot, oldPrimary: watchOld, every: *pollEvery, fast: fast, fastEvery: *pollFast}
	watchCtx, stopWatch := context.WithCancel(context.Background())
	watchDone := make(chan struct{})
	go func() {
		w.run(watchCtx)
		close(watchDone)
	}()
	esDone := make(chan struct{})
	go func() {
		es.run(watchCtx)
		close(esDone)
	}()

	runCtx, stopRun := context.WithCancel(context.Background())
	endTimer := time.AfterFunc(*duration, stopRun)
	defer endTimer.Stop()

	var scriptMu sync.Mutex
	var scriptOut []string
	keep := func(line string) {
		scriptMu.Lock()
		scriptOut = append(scriptOut, line)
		scriptMu.Unlock()
	}
	scriptDone := make(chan error, 1)
	var mig *migrationResult
	sink := &migrationSink{t0: t0}
	switch kind {
	case "none":
		scriptDone <- nil
	case "primary":
		time.AfterFunc(*failAt, func() {
			scriptDone <- runFailScript(script, []string{strconv.Itoa(*failSeconds)},
				[]string{"MODE=" + *failMode, "HOLD_PROMOTION=" + strconv.Itoa(*failHold)}, t0, st, keep)
		})
	case "planned":
		time.AfterFunc(*failAt, func() {
			scriptDone <- runFailScript(script, []string{"0"}, nil, t0, st, keep)
		})
	case "migrate":
		time.AfterFunc(*failAt, func() {
			r := runMigration(t0, migSrc, migDst, migLo, migHi, *migrateTimeout, *migratePoll,
				func(at time.Time) { st.failAt.CompareAndSwap(0, int64(at.Sub(t0))) },
				func(successAt time.Time) {
					if !successAt.IsZero() {
						st.restoredAt.CompareAndSwap(0, int64(successAt.Sub(t0)))
					}
					endTimer.Stop()
					time.AfterFunc(*afterMigrate, stopRun) // also when it failed: end the run soon
				}, sink.logf)
			mig = &r
			if r.Error != "" {
				scriptDone <- errors.New(r.Error)
			} else {
				scriptDone <- nil
			}
		})
	}

	var wg sync.WaitGroup
	for i := 0; i < *goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(runCtx, i, clients[i%len(clients)], keys, value, st)
		}()
	}
	<-runCtx.Done()
	runLen := time.Since(t0)
	wg.Wait()
	stopRun()
	stopWatch()
	<-watchDone
	<-esDone
	var scriptErr error
	select {
	case scriptErr = <-scriptDone:
	case <-time.After(time.Duration(*failSeconds)*time.Second + 60*time.Second):
		scriptErr = errors.New("the fail script did not finish")
	}
	snap := dc.snapshot()
	es.final(5 * time.Second)
	closeAll()
	closed = true
	if kind == "migrate" {
		scriptOut = sink.lines
	}

	// Results.
	failAtD := time.Duration(st.failAt.Load())
	ms := func(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }
	sec := func(d time.Duration) float64 { return d.Seconds() }
	fromFail := func(v int64) float64 {
		if v <= 0 || failAtD <= 0 {
			return -1
		}
		return ms(time.Duration(v) - failAtD)
	}
	r := result{
		Label: *label, Run: *runID, Build: buildName, Options: options, Args: os.Args[1:], FailKind: kind,
		FailMode: *failMode, FailSeconds: *failSeconds, Goroutines: *goroutines,
		DurationS: duration.Seconds(), RunLengthS: runLen.Seconds(), TimeoutMs: ms(*reqTimeout), Addr: *addr,
		OldPrimary: oldPrimary, Slots: target.Ranges, KeySlots: *keySlots, FailAtS: -1, RestoredAtS: -1,
		OK: st.ok.Load(), NoSlot: st.noSlot.Load(), Deadline: st.deadline.Load(), Other: st.other.Load(),
		OtherErrors:        st.otherErrors(),
		LongestOKMs:        ms(time.Duration(st.longestOK.Load())),
		LongestOKStartS:    sec(time.Duration(st.longestOKStart.Load())),
		FirstOKAfterFailMs: -1, FirstErrAtS: -1, LastErrAfterFailMs: -1, NoSlotFromMs: -1, NoSlotToMs: -1,
		FlaggedFailLineMs: fromFail(st.flaggedAt.Load()), PromoteLineMs: fromFail(st.promoteAt.Load()),
		PromotedLineMs: fromFail(st.promotedAt.Load()), FailedOverLineMs: fromFail(st.failedOverAt.Load()),
		PromotionMs: -1, LatencyWindowFromS: -1, LatencyWindowToS: -1, StaleRouteMs: -1, LastMovedOldPrimaryMs: -1,
		ShardsByAddr: map[string]int64{}, GetSlotMigrationsByAddr: map[string]int64{}, BytesByAddr: map[string]int64{},
		Dials: snap.Dials, DialErrs: snap.DialErrs, ScriptOutput: scriptOut, FailHold: *failHold, Migration: mig,
	}
	if kind != "primary" {
		r.FailMode, r.FailSeconds, r.FailHold = kind, 0, 0
	}
	if scriptErr != nil {
		r.ScriptError = scriptErr.Error()
	}
	r.Requests = r.OK + r.NoSlot + r.Deadline + r.Other
	if runLen > 0 {
		r.OpsPerSec = float64(r.OK) / runLen.Seconds()
	}
	if failAtD > 0 {
		r.FailAtS = sec(failAtD)
		if v := st.firstOKAfterFail.Load(); v >= 0 {
			r.FirstOKAfterFailMs = ms(time.Duration(v) - failAtD)
		}
		if v := st.lastErr.Load(); v >= 0 {
			r.LastErrAfterFailMs = ms(time.Duration(v) - failAtD)
		}
		if v := st.firstNoSlot.Load(); v >= 0 {
			r.NoSlotFromMs = ms(time.Duration(v) - failAtD)
			r.NoSlotToMs = ms(time.Duration(st.lastNoSlot.Load()) - failAtD)
		}
	}
	if v := st.restoredAt.Load(); v > 0 {
		r.RestoredAtS = sec(time.Duration(v))
	}
	if v := st.firstErr.Load(); v >= 0 {
		r.FirstErrAtS = sec(time.Duration(v))
	}

	seconds := int((runLen+*reqTimeout)/time.Second) + 1
	rows := make([]timelineRow, seconds)
	for i := range rows {
		rows[i].T = i
	}
	perSec := int(time.Second / bucketWidth)
	for i := range st.buckets {
		row := min(i/perSec, len(rows)-1)
		b := &st.buckets[i]
		rows[row].OK += b.ok.Load()
		rows[row].NoSlot += b.noSlot.Load()
		rows[row].Deadline += b.deadline.Load()
		rows[row].Other += b.other.Load()
	}
	for _, e := range snap.Shards {
		r.ShardsByAddr[e.addr]++
		at := e.at.Sub(t0)
		off := e.addr != *addr
		if off {
			r.ShardsOffEndpoint++
		}
		if at < 0 {
			r.ShardsInit++
			continue
		}
		r.ShardsRun++
		if off {
			r.ShardsOffEndpointRun++
			if failAtD > 0 && len(r.ShardsOffEndpointTimesMs) < 100 {
				r.ShardsOffEndpointTimesMs = append(r.ShardsOffEndpointTimesMs, ms(at-failAtD))
			}
		}
		if failAtD > 0 && at >= failAtD {
			r.ShardsAfterFail++
		}
		row := min(int(at/time.Second), len(rows)-1)
		rows[row].Shards++
	}
	for _, e := range snap.GSM {
		r.GetSlotMigrationsByAddr[e.addr]++
		at := e.at.Sub(t0)
		if at >= 0 {
			r.GetSlotMigrationsRun++
			if failAtD > 0 && at >= failAtD {
				r.GetSlotMigrationsAfterFail++
			}
		}
	}
	for a, b := range snap.Bytes {
		var n int64
		for _, v := range b {
			n += v
		}
		r.BytesByAddr[a] = n
	}
	r.ServerErrors = es.summary(seconds)
	r.ServerMoved = r.ServerErrors.Totals["MOVED"]
	r.ServerAsk = r.ServerErrors.Totals["ASK"]
	for i := range rows {
		if m := r.ServerErrors.PerSecond["MOVED"]; i < len(m) {
			rows[i].Moved = m[i]
		}
		if a := r.ServerErrors.PerSecond["ASK"]; i < len(a) {
			rows[i].Ask = a[i]
		}
	}
	for _, row := range rows {
		r.ShardsMaxPerSec = max(r.ShardsMaxPerSec, row.Shards)
	}
	r.Timeline = rows
	if st.resumeAt > 0 {
		at := time.Duration(st.resumeAt)
		fromResume := func(v int64) float64 {
			if v < 0 {
				return -1
			}
			return ms(time.Duration(v) - at)
		}
		rr := &resumeResult{
			IdleFromS: sec(*idleFrom), IdleUntilS: sec(at),
			OK: st.resumeOK.Load(), NoSlot: st.resumeNoSlot.Load(), Deadline: st.resumeDeadline.Load(), Other: st.resumeOther.Load(),
			FirstOKAfterResumeMs: fromResume(st.firstOKAfterResume.Load()),
			FirstErrAfterResume:  fromResume(st.firstErrAfterResume.Load()),
			LastErrAfterResumeMs: fromResume(st.lastErrAfterResume.Load()),
		}
		from := int(at / time.Second)
		for _, row := range rows {
			if row.T >= from && row.T < from+5 {
				rr.MovedFirst5s += row.Moved
				rr.ShardsFirst5s += row.Shards
			}
		}
		r.Resume = rr
	}
	r.Topology = w.summarize(failAtD)
	for _, e := range snap.Shards {
		at := e.at.Sub(t0)
		if failAtD <= 0 || at < failAtD {
			continue
		}
		rel := ms(at - failAtD)
		if len(r.ShardsTimesMs) < 300 {
			r.ShardsTimesMs = append(r.ShardsTimesMs, rel)
		}
		if t := r.Topology; t.NoPrimaryStartMs >= 0 && t.NoPrimaryEndMs >= 0 && rel >= t.NoPrimaryStartMs && rel <= t.NoPrimaryEndMs {
			r.ShardsInWindow++
		}
	}

	// Promotion: when the new owner of slot 0 was online (the watcher polls its candidates every
	// -poll-fast), else when the replica answered CLUSTER FAILOVER [FORCE].
	switch {
	case r.Topology.PromotedMs >= 0:
		r.PromotionMs, r.PromotionSource = r.Topology.PromotedMs, "watcher"
	case r.PromoteLineMs >= 0:
		r.PromotionMs, r.PromotionSource = r.PromoteLineMs, "promote-line"
	}
	if r.PromotionMs >= 0 && r.FirstOKAfterFailMs >= 0 {
		v := r.FirstOKAfterFailMs - r.PromotionMs
		r.PromotionToFirstOKMs = &v
	}
	if r.PromotionMs >= 0 {
		promoted := failAtD + time.Duration(r.PromotionMs*float64(time.Millisecond))
		r.StaleRouteMs = 0
		if b := snap.Bytes[watchOld]; b != nil {
			for i := int(promoted / bytesBucketWidth); i < len(b) && time.Duration(i)*bytesBucketWidth <= runLen; i++ {
				if b[i] > *routeBytes {
					r.StaleRouteMs = ms(time.Duration(i+1)*bytesBucketWidth - promoted)
				}
			}
		}
		if at := es.lastIncrease(watchOld, "MOVED"); at >= promoted {
			r.LastMovedOldPrimaryMs = ms(at - promoted)
		}
	}

	// Latency of successes: whole run, and the failure window: from the failure to the recovery plus
	// -window-tail, where the recovery is the latest of the first success after the failure, the last
	// error, the promotion and the migration's success.
	r.LatencyAll = summarizeHist(st.lat, 0, len(st.lat))
	if failAtD > 0 {
		end := 0.0
		for _, v := range []float64{r.FirstOKAfterFailMs, r.LastErrAfterFailMs, r.PromotionMs} {
			end = max(end, v)
		}
		if mig != nil && mig.SuccessMs >= 0 { // from the start of the migration, which is the failure time
			end = max(end, mig.SuccessMs)
		}
		from := failAtD
		to := min(failAtD+time.Duration(end*float64(time.Millisecond))+*windowTail, runLen+*reqTimeout)
		r.LatencyWindowFromS, r.LatencyWindowToS = sec(from), sec(to)
		r.LatencyWindow = summarizeHist(st.lat, int(from/bucketWidth), int((to+bucketWidth-1)/bucketWidth))
	}

	// Migration: when the client stopped sending workload traffic to the source, and the source its MOVED replies.
	if mig != nil {
		start := time.Duration(mig.StartS * float64(time.Second))
		if r.Topology.PromotedMs >= 0 {
			mig.HandoverSeenMs = r.Topology.PromotedMs
		}
		if mig.HandoverMs >= 0 {
			handover := start + time.Duration(mig.HandoverMs*float64(time.Millisecond))
			mig.RouteConvergedMs = 0
			if b := snap.Bytes[mig.Source]; b != nil {
				for i := int(handover / bytesBucketWidth); i < len(b) && time.Duration(i)*bytesBucketWidth <= runLen; i++ {
					mig.SourceBytesAfterHandover += b[i]
					if b[i] > *routeBytes {
						mig.RouteConvergedMs = ms(time.Duration(i+1)*bytesBucketWidth - handover)
					}
				}
			}
			if at := es.lastIncrease(mig.Source, "MOVED"); at >= 0 {
				mig.LastMovedMs = ms(at - handover)
			}
		}
	}

	printResult(r)
	if *jsonOut != "" {
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(*jsonOut, append(b, '\n'), 0o644); err != nil {
			return err
		}
	}
	if scriptErr != nil {
		return fmt.Errorf("fail script: %w", scriptErr)
	}
	return nil
}

func worker(ctx context.Context, id int, client valkey.Client, keys []string, value string, st *stats) {
	rnd := rand.New(rand.NewPCG(uint64(id)+1, 0x5eed))
	for i := 0; ctx.Err() == nil; i++ {
		if *idleUntil > *idleFrom {
			if now := time.Since(st.t0); now >= *idleFrom && now < *idleUntil {
				select {
				case <-time.After(*idleUntil - now):
				case <-ctx.Done():
					return
				}
			}
		}
		if *pipeline > 0 {
			batch := make(valkey.Commands, 0, *pipeline)
			for j := 0; j < *pipeline; j++ {
				k := keys[rnd.IntN(len(keys))]
				if i%2 == 0 {
					batch = append(batch, client.B().Get().Key(k).Build())
				} else {
					batch = append(batch, client.B().Set().Key(k).Value(value).Build())
				}
			}
			rctx, cancel := context.WithTimeout(context.Background(), *reqTimeout)
			start := time.Now()
			resps := client.DoMulti(rctx, batch...)
			end := time.Now()
			cancel()
			for _, r := range resps {
				st.record(start, end, r.Error())
			}
			continue
		}
		k := keys[rnd.IntN(len(keys))]
		var cmd valkey.Completed
		if i%2 == 0 {
			cmd = client.B().Get().Key(k).Build()
		} else {
			cmd = client.B().Set().Key(k).Value(value).Build()
		}
		rctx, cancel := context.WithTimeout(context.Background(), *reqTimeout)
		start := time.Now()
		err := client.Do(rctx, cmd).Error()
		end := time.Now()
		cancel()
		st.record(start, end, err)
	}
}

func printResult(r result) {
	f := func(v float64) string {
		if v < 0 {
			return "never"
		}
		return fmt.Sprintf("+%.3fs", v/1000)
	}
	fmt.Printf("\n=== result label=%s run=%d build=%s options=%v mode=%s hold=%ds ===\n", r.Label, r.Run, r.Build, r.Options, r.FailMode, r.FailHold)
	fmt.Printf("failed primary %s (slots %v): failure at t=%.3fs, restored at t=%.3fs\n", r.OldPrimary, r.Slots, r.FailAtS, r.RestoredAtS)
	fmt.Printf("requests %d: ok %d, ErrNoSlot %d, context deadline %d, other %d\n", r.Requests, r.OK, r.NoSlot, r.Deadline, r.Other)
	for i, e := range r.OtherErrors {
		if i == 8 {
			fmt.Printf("  ... %d more messages\n", len(r.OtherErrors)-i)
			break
		}
		fmt.Printf("  other %8d  first at t=%.3fs  %s\n", e.Count, e.FirstAt, e.Msg)
	}
	fmt.Printf("longest success %.1fms (started at t=%.3fs); first success after the failure %s; errors from t=%.3fs to %s after the failure; ErrNoSlot from %s to %s\n",
		r.LongestOKMs, r.LongestOKStartS, f(r.FirstOKAfterFailMs), r.FirstErrAtS, f(r.LastErrAfterFailMs), f(r.NoSlotFromMs), f(r.NoSlotToMs))
	fmt.Printf("CLUSTER SHARDS sent: %d while creating the client, %d during the run, %d after the failure, %d in the no-primary window, max %d/s; by address %v\n",
		r.ShardsInit, r.ShardsRun, r.ShardsAfterFail, r.ShardsInWindow, r.ShardsMaxPerSec, r.ShardsByAddr)
	if len(r.ShardsTimesMs) > 0 {
		var b strings.Builder
		for i, v := range r.ShardsTimesMs {
			if i == 40 {
				fmt.Fprintf(&b, " ... (%d more)", r.ShardsAfterFail-40)
				break
			}
			fmt.Fprintf(&b, " +%.3f", v/1000)
		}
		fmt.Printf("CLUSTER SHARDS times after the failure (s):%s\n", b.String())
	}
	fmt.Printf("CLUSTER SHARDS to other addresses than the discovery endpoint %s: %d (%d during the run); CLUSTER GETSLOTMIGRATIONS: %d during the run, by address %v\n",
		r.Addr, r.ShardsOffEndpoint, r.ShardsOffEndpointRun, r.GetSlotMigrationsRun, r.GetSlotMigrationsByAddr)
	keys := make([]string, 0, len(r.Dials))
	for k := range r.Dials {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("dials:")
	for _, k := range keys {
		fmt.Printf(" %s=%d", k, r.Dials[k])
		if e := r.DialErrs[k]; e > 0 {
			fmt.Printf("(%d failed)", e)
		}
	}
	fmt.Println()
	t := r.Topology
	fmt.Printf("topology (CLUSTER SHARDS on every node every %v): old primary flagged fail %s; no online primary from %s to %s (%.0fms, %d polls on %d nodes); new primary %s %s\n",
		*pollEvery, f(t.FailFlagMs), f(t.NoPrimaryStartMs), f(t.NoPrimaryEndMs), t.NoPrimaryMs, t.NoPrimaryPolls, t.NodesNoPrimary, t.NewPrimary, f(t.PromotedMs))
	pf := "n/a"
	if r.PromotionToFirstOKMs != nil {
		pf = fmt.Sprintf("%+.3fs", *r.PromotionToFirstOKMs/1000)
	}
	fmt.Printf("promotion %s (%s; PROMOTE line %s, FLAGGED_FAIL line %s); promotion → first success %s; client traffic to the old primary until %s after the promotion, its last MOVED %s after it\n",
		f(r.PromotionMs), r.PromotionSource, f(r.PromoteLineMs), f(r.FlaggedFailLineMs), pf, f(r.StaleRouteMs), f(r.LastMovedOldPrimaryMs))
	other := map[string]int64{}
	for k, v := range r.ServerErrors.Totals {
		if k != "MOVED" && k != "ASK" {
			other[k] = v
		}
	}
	fmt.Printf("server errors (INFO errorstats, all nodes, before → after the run): MOVED %d, ASK %d, other %v; restarts %v\n",
		r.ServerMoved, r.ServerAsk, other, r.ServerErrors.Restarts)
	l := func(s latencySummary) string {
		return fmt.Sprintf("n=%d p50 %.3fms p99 %.3fms p999 %.3fms max %.1fms", s.Count, s.P50, s.P99, s.P999, s.Max)
	}
	fmt.Printf("latency of successes: run %s; %.0f ops/s over %.1fs\n", l(r.LatencyAll), r.OpsPerSec, r.RunLengthS)
	if r.LatencyWindowFromS >= 0 {
		fmt.Printf("latency of successes in the failure window t=%.1fs..%.1fs: %s\n", r.LatencyWindowFromS, r.LatencyWindowToS, l(r.LatencyWindow))
	}
	if m := r.Migration; m != nil {
		fmt.Printf("migration %s of slots %d-%d %s → %s: MIGRATESLOTS at t=%.3fs, source success %s, target success %s, handover seen by the watcher %s; client traffic to the source until %s after the handover (%d bytes from it on), last MOVED from the source %s after it; error %q\n",
			m.Job, m.Slots[0], m.Slots[1], m.Source, m.Target, m.StartS, f(m.SuccessMs), f(m.TargetSuccessMs), f(m.HandoverSeenMs), f(m.RouteConvergedMs), m.SourceBytesAfterHandover, f(m.LastMovedMs), m.Error)
	}
	if rr := r.Resume; rr != nil {
		fmt.Printf("idle from t=%.1fs to t=%.1fs; after it: ok %d, ErrNoSlot %d, deadline %d, other %d; first success %s, first error %s, last error %s after it; in its first 5s: MOVED (srv) %d, CLUSTER SHARDS %d\n",
			rr.IdleFromS, rr.IdleUntilS, rr.OK, rr.NoSlot, rr.Deadline, rr.Other, f(rr.FirstOKAfterResumeMs), f(rr.FirstErrAfterResume), f(rr.LastErrAfterResumeMs), rr.MovedFirst5s, rr.ShardsFirst5s)
	}
	fmt.Printf("timeline (by request end time):\n     t        ok  ErrNoSlot   deadline      other  CLUSTER_SHARDS  MOVED(srv)  ASK(srv)\n")
	for _, row := range r.Timeline {
		fmt.Printf("  %3ds %9d %10d %10d %10d %10d %11d %9d\n", row.T, row.OK, row.NoSlot, row.Deadline, row.Other, row.Shards, row.Moved, row.Ask)
	}
}
