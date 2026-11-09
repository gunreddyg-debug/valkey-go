package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Atomic Slot Migration (Valkey 9.0+) during the run, done the way the integration tests of PR #209
// do it (cluster_migration_e2e_test.go, asmHarness.migrate): CLUSTER MIGRATESLOTS SLOTSRANGE <from>
// <to> NODE <target node id>, sent to the source primary, then CLUSTER GETSLOTMIGRATIONS on the source
// until the new job reports success. The bench also polls the target, to see when the IMPORT side
// reports success. Both go to the nodes directly, never through the client or the discovery endpoint.

type migState struct {
	Ms    float64 `json:"ms"` // from the start of the migration
	Node  string  `json:"node"`
	State string  `json:"state"`
}

type migrationResult struct {
	Source          string     `json:"source"`
	Target          string     `json:"target"`
	TargetID        string     `json:"target_id"`
	Slots           [2]int64   `json:"slots"`
	StartS          float64    `json:"start_s"`     // from t0: MIGRATESLOTS sent (the failure time of this kind)
	AcceptedMs      float64    `json:"accepted_ms"` // its reply, from the start
	Reply           string     `json:"reply"`
	Job             string     `json:"job"`
	SuccessMs       float64    `json:"success_ms"`        // the source (EXPORT) reports success, from the start; -1: never
	TargetSuccessMs float64    `json:"target_success_ms"` // the target (IMPORT) reports success, from the start; -1: never
	HandoverMs      float64    `json:"handover_ms"`       // the first of the two, from the start: the reference of the times below; -1: never
	States          []migState `json:"states"`            // state changes seen, both sides (times: when the poll that saw them was sent)
	Error           string     `json:"error,omitempty"`
	Polls           int        `json:"polls"`

	// Filled in after the run.
	HandoverSeenMs           float64 `json:"handover_seen_ms"`            // from the start: the watcher first saw the slot owned by the target; -1: never
	RouteConvergedMs         float64 `json:"route_converged_ms"`          // from the handover: end of the last 10 ms bucket in which the client still wrote workload traffic (> -route-bytes) to the source; 0: none after the handover; -1: no handover
	SourceBytesAfterHandover int64   `json:"source_bytes_after_handover"` // bytes the client wrote to the source from the handover bucket on
	LastMovedMs              float64 `json:"last_moved_ms"`               // from the handover: last INFO errorstats read of the source that saw new MOVED replies; -1: none
}

type slotMigrationEntry struct {
	Name, Operation, State, Message, Slots string
}

func parseSlotMigrationEntries(v any) []slotMigrationEntry {
	arr, _ := v.([]any)
	out := make([]slotMigrationEntry, 0, len(arr))
	for _, e := range arr {
		m := flatMap(e)
		str := func(k string) string {
			switch x := m[k].(type) {
			case string:
				return x
			case int64:
				return strconv.FormatInt(x, 10)
			}
			return ""
		}
		out = append(out, slotMigrationEntry{Name: str("name"), Operation: str("operation"), State: str("state"), Message: str("message"), Slots: str("slot_ranges")})
	}
	return out
}

// rconn is a persistent raw connection to one node.
type rconn struct {
	addr string
	conn net.Conn
	rd   *bufio.Reader
}

func (c *rconn) do(timeout time.Duration, cmd []byte) (any, error) {
	if c.conn == nil {
		conn, err := net.DialTimeout("tcp", c.addr, timeout)
		if err != nil {
			return nil, err
		}
		c.conn, c.rd = conn, bufio.NewReader(conn)
	}
	_ = c.conn.SetDeadline(time.Now().Add(timeout))
	if _, err := c.conn.Write(cmd); err != nil {
		c.close()
		return nil, err
	}
	v, err := readRESP(c.rd)
	if err != nil {
		c.close()
		return nil, err
	}
	if e, ok := v.(respError); ok {
		return nil, e
	}
	return v, nil
}

func (c *rconn) close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

var getSlotMigrationsRaw = encodeCommand("CLUSTER", "GETSLOTMIGRATIONS")

// runMigration migrates slots [lo, hi] from src to dst. It calls started right after sending
// MIGRATESLOTS and finished when the job has ended (successAt is zero if it did not succeed), and logs
// progress lines with logf. It returns when the source and the target both report success (or after
// 2 s more for the target), when the job fails, or after timeout.
func runMigration(t0 time.Time, src, dst shardNode, lo, hi int64, timeout, poll time.Duration,
	started func(time.Time), finished func(successAt time.Time), logf func(string, ...any)) migrationResult {
	r := migrationResult{Source: src.Addr, Target: dst.Addr, TargetID: dst.ID, Slots: [2]int64{lo, hi},
		AcceptedMs: -1, SuccessMs: -1, TargetSuccessMs: -1, HandoverMs: -1, HandoverSeenMs: -1, RouteConvergedMs: -1, LastMovedMs: -1}
	sc, dc := &rconn{addr: src.Addr}, &rconn{addr: dst.Addr}
	defer sc.close()
	defer dc.close()
	before := map[string]bool{}
	if v, err := sc.do(time.Second, getSlotMigrationsRaw); err == nil {
		for _, e := range parseSlotMigrationEntries(v) {
			before[e.Name] = true
		}
	} else {
		r.Error = "GETSLOTMIGRATIONS on the source: " + err.Error()
		finished(time.Time{})
		return r
	}
	dstBefore := map[string]bool{}
	if v, err := dc.do(time.Second, getSlotMigrationsRaw); err == nil {
		for _, e := range parseSlotMigrationEntries(v) {
			dstBefore[e.Name] = true
		}
	}

	cmd := encodeCommand("CLUSTER", "MIGRATESLOTS", "SLOTSRANGE", strconv.FormatInt(lo, 10), strconv.FormatInt(hi, 10), "NODE", dst.ID)
	start := time.Now()
	started(start)
	r.StartS = start.Sub(t0).Seconds()
	ms := func(t time.Time) float64 { return float64(t.Sub(start).Microseconds()) / 1000 }
	v, err := sc.do(2*time.Second, cmd)
	r.AcceptedMs = ms(time.Now())
	if err != nil {
		r.Error = "MIGRATESLOTS: " + err.Error()
		logf("MIGRATESLOTS SLOTSRANGE %d %d NODE %s on %s: %v", lo, hi, dst.ID, src.Addr, err)
		finished(time.Time{})
		return r
	}
	r.Reply = fmt.Sprint(v)
	logf("MIGRATESLOTS SLOTSRANGE %d %d NODE %s (%s) on %s: %v", lo, hi, dst.ID, dst.Addr, src.Addr, v)

	var successAt, targetSuccessAt time.Time
	lastState := map[string]string{}
	deadline := start.Add(timeout)
	for time.Now().Before(deadline) {
		r.Polls++
		if successAt.IsZero() {
			now := time.Now()
			if v, err := sc.do(500*time.Millisecond, getSlotMigrationsRaw); err == nil {
				for _, e := range parseSlotMigrationEntries(v) {
					if before[e.Name] || (r.Job != "" && e.Name != r.Job) {
						continue
					}
					r.Job = e.Name
					if e.State != lastState[src.Addr] {
						lastState[src.Addr] = e.State
						r.States = append(r.States, migState{Ms: ms(now), Node: src.Addr, State: e.State})
						logf("migration %s on the source %s: %s %s", e.Name, src.Addr, e.State, e.Message)
					}
					switch e.State {
					case "success":
						successAt = now
						r.SuccessMs = ms(now)
					case "failed", "cancelled":
						r.Error = fmt.Sprintf("migration %s: %s %s", e.Name, e.State, e.Message)
						finished(time.Time{})
						return r
					}
				}
			}
		}
		if targetSuccessAt.IsZero() {
			now := time.Now()
			if v, err := dc.do(500*time.Millisecond, getSlotMigrationsRaw); err == nil {
				for _, e := range parseSlotMigrationEntries(v) {
					if dstBefore[e.Name] || e.Operation != "IMPORT" {
						continue
					}
					if e.State != lastState[dst.Addr] {
						lastState[dst.Addr] = e.State
						r.States = append(r.States, migState{Ms: ms(now), Node: dst.Addr, State: e.State})
						logf("migration %s on the target %s: %s %s", e.Name, dst.Addr, e.State, e.Message)
					}
					if e.State == "success" {
						targetSuccessAt = now
						r.TargetSuccessMs = ms(now)
					}
				}
			}
		}
		if !successAt.IsZero() && (!targetSuccessAt.IsZero() || time.Since(successAt) > 2*time.Second) {
			break
		}
		time.Sleep(poll)
	}
	if successAt.IsZero() {
		r.Error = fmt.Sprintf("migration did not succeed within %v (job %q)", timeout, r.Job)
	}
	switch {
	case r.SuccessMs >= 0 && r.TargetSuccessMs >= 0:
		r.HandoverMs = min(r.SuccessMs, r.TargetSuccessMs)
	case r.SuccessMs >= 0:
		r.HandoverMs = r.SuccessMs
	case r.TargetSuccessMs >= 0:
		r.HandoverMs = r.TargetSuccessMs
	}
	finished(successAt)
	return r
}

// migrationSink collects the progress lines of the migration like the fail script's output.
type migrationSink struct {
	mu    sync.Mutex
	t0    time.Time
	lines []string
}

func (m *migrationSink) logf(format string, args ...any) {
	line := fmt.Sprintf("[t=%6.3fs] %s", time.Since(m.t0).Seconds(), fmt.Sprintf(format, args...))
	fmt.Println("migrate:", line)
	m.mu.Lock()
	m.lines = append(m.lines, line)
	m.mu.Unlock()
}

func parseSlotRange(s string) (int64, int64, error) {
	a, b, ok := strings.Cut(s, "-")
	if !ok {
		b = a
	}
	lo, err1 := strconv.ParseInt(strings.TrimSpace(a), 10, 64)
	hi, err2 := strconv.ParseInt(strings.TrimSpace(b), 10, 64)
	if err1 != nil || err2 != nil || lo < 0 || hi > 16383 || lo > hi {
		return 0, 0, fmt.Errorf("bad slot range %q (want FROM-TO within 0-16383)", s)
	}
	return lo, hi, nil
}
