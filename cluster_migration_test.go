package valkey

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------- fixtures ----------

// migEntry builds one CLUSTER GETSLOTMIGRATIONS entry. kv alternates keys and values (string or int64).
// resp3 selects a RESP3 map, otherwise a RESP2 flat array is returned.
func migEntry(resp3 bool, kv ...any) ValkeyMessage {
	values := make([]ValkeyMessage, 0, len(kv))
	for _, v := range kv {
		switch vv := v.(type) {
		case string:
			values = append(values, strmsg(typeBlobString, vv))
		case int64:
			values = append(values, ValkeyMessage{typ: typeInteger, intlen: vv})
		case int:
			values = append(values, ValkeyMessage{typ: typeInteger, intlen: int64(vv)})
		}
	}
	if resp3 {
		return slicemsg(typeMap, values)
	}
	return slicemsg(typeArray, values)
}

func migReply(entries ...ValkeyMessage) ValkeyResult {
	return NewResult(slicemsg(typeArray, entries), nil)
}

func exportEntry(name, state, slots string) ValkeyMessage {
	return migEntry(true,
		"name", name,
		"operation", "EXPORT",
		"slot_ranges", slots,
		"target_node", "4b4f12fdfb58d5e30fef7b9ad3f1651dacbbaba9",
		"source_node", "93941e777e17fcbc92d4398cc957ffea888f472b",
		"state", state,
		"message", "",
		"cow_size", 0,
	)
}

type shardSpec struct {
	ip    string
	port  int64
	slots []int64 // pairs of start, end
}

// testShards builds a CLUSTER SHARDS reply with one primary per shard.
func testShards(specs ...shardSpec) ValkeyResult {
	shards := make([]ValkeyMessage, 0, len(specs))
	for _, s := range specs {
		slots := make([]ValkeyMessage, 0, len(s.slots))
		for _, v := range s.slots {
			slots = append(slots, ValkeyMessage{typ: typeInteger, intlen: v})
		}
		shards = append(shards, slicemsg(typeMap, []ValkeyMessage{
			strmsg(typeBlobString, "slots"),
			slicemsg(typeArray, slots),
			strmsg(typeBlobString, "nodes"),
			slicemsg(typeArray, []ValkeyMessage{
				slicemsg(typeMap, []ValkeyMessage{
					strmsg(typeBlobString, "id"), strmsg(typeBlobString, s.ip),
					strmsg(typeBlobString, "port"), {typ: typeInteger, intlen: s.port},
					strmsg(typeBlobString, "ip"), strmsg(typeBlobString, s.ip),
					strmsg(typeBlobString, "endpoint"), strmsg(typeBlobString, s.ip),
					strmsg(typeBlobString, "role"), strmsg(typeBlobString, "master"),
					strmsg(typeBlobString, "replication-offset"), {typ: typeInteger, intlen: 1},
					strmsg(typeBlobString, "health"), strmsg(typeBlobString, "online"),
				}),
			}),
		}))
	}
	return NewResult(slicemsg(typeArray, shards), nil)
}

func isCmd(cmd Completed, parts ...string) bool {
	c := cmd.Commands()
	if len(c) < len(parts) {
		return false
	}
	for i, p := range parts {
		if !strings.EqualFold(c[i], p) {
			return false
		}
	}
	return true
}

// ---------- parsing ----------

func TestParseSlotMigrationRanges(t *testing.T) {
	tests := []struct {
		in   string
		want [][2]int64
	}{
		{"0-10", [][2]int64{{0, 10}}},
		{"100", [][2]int64{{100, 100}}},
		{"0-10 20-30 55", [][2]int64{{0, 10}, {20, 30}, {55, 55}}},
		{"  1-2   3-4 ", [][2]int64{{1, 2}, {3, 4}}},
		{"0-16383", [][2]int64{{0, 16383}}},
		{"", nil},
		{"abc 10-5 0-16384 -1 7-x", nil},
	}
	for _, tt := range tests {
		if got := parseSlotMigrationRanges(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parseSlotMigrationRanges(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseSlotMigrations(t *testing.T) {
	// fixtures follow the examples of https://valkey.io/commands/cluster-getslotmigrations/
	for _, resp3 := range []bool{false, true} {
		reply := slicemsg(typeArray, []ValkeyMessage{
			migEntry(resp3,
				"name", "5371b28997de6fd0bbe813ad8ebdfdf2faadb308",
				"operation", "EXPORT",
				"slot_ranges", "0-10 20-30",
				"target_node", "4b4f12fdfb58d5e30fef7b9ad3f1651dacbbaba9",
				"source_node", "93941e777e17fcbc92d4398cc957ffea888f472b",
				"create_time", 1754870400,
				"last_update_time", 1754870400,
				"last_ack_time", 1754870400,
				"state", "success",
				"message", "",
				"cow_size", 0,
				"remaining_repl_size", 0,
			),
			migEntry(resp3, // reported by the target node
				"name", "aaaa", "operation", "IMPORT", "slot_ranges", "1-1", "state", "receiving-snapshot",
			),
			migEntry(resp3, // unknown operation, ignored
				"name", "cccc", "operation", "OTHER", "slot_ranges", "1-1", "state", "snapshotting",
			),
			migEntry(resp3, // Valkey 9.0 entry without remaining_repl_size
				"name", "bbbb", "operation", "EXPORT", "slot_ranges", "5", "state", "snapshotting",
			),
			migEntry(resp3, // an entry without a name is ignored
				"operation", "EXPORT", "slot_ranges", "5", "state", "snapshotting",
			),
			strmsg(typeBlobString, "garbage"), // not a map, ignored
		})
		got, err := parseSlotMigrations(reply)
		if err != nil {
			t.Fatalf("resp3=%v unexpected err %v", resp3, err)
		}
		want := []slotMigration{
			{name: "5371b28997de6fd0bbe813ad8ebdfdf2faadb308", operation: "EXPORT", state: "success", slots: [][2]int64{{0, 10}, {20, 30}}},
			{name: "aaaa", operation: "IMPORT", state: "receiving-snapshot", slots: [][2]int64{{1, 1}}},
			{name: "bbbb", operation: "EXPORT", state: "snapshotting", slots: [][2]int64{{5, 5}}},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("resp3=%v got %+v, want %+v", resp3, got, want)
		}
	}

	if _, err := parseSlotMigrations(strmsg(typeBlobString, "x")); err == nil {
		t.Fatal("expected an error for a non-array reply")
	}

}

func TestSlotMigrationPhase(t *testing.T) {
	tests := []struct {
		operation, state string
		want             slotMigrationPhase
	}{
		{"EXPORT", "success", slotMigrationPhaseIdle},
		{"EXPORT", "failed", slotMigrationPhaseIdle},
		{"EXPORT", "cancelled", slotMigrationPhaseIdle},
		{"EXPORT", "connecting", slotMigrationPhaseActive},
		{"EXPORT", "snapshotting", slotMigrationPhaseActive},
		{"EXPORT", "replicating", slotMigrationPhaseActive},
		{"EXPORT", "some-future-state", slotMigrationPhaseActive},
		{"EXPORT", "waiting-to-pause", slotMigrationPhaseFinalizing},
		{"EXPORT", "failover-paused", slotMigrationPhaseFinalizing},
		{"EXPORT", "failover-granted", slotMigrationPhaseFinalizing},
		{"EXPORT", "waiting-for-paused", slotMigrationPhaseActive}, // target states only count on the target
		{"IMPORT", "waiting-for-ack", slotMigrationPhaseActive},
		{"IMPORT", "receiving-snapshot", slotMigrationPhaseActive},
		{"IMPORT", "waiting-for-paused", slotMigrationPhaseFinalizing},
		{"IMPORT", "failover-requested", slotMigrationPhaseFinalizing},
		{"IMPORT", "failover-granted", slotMigrationPhaseActive}, // source states only count on the source
		{"IMPORT", "success", slotMigrationPhaseIdle},
		{"IMPORT", "cancelled", slotMigrationPhaseIdle},
	}
	for _, tt := range tests {
		m := slotMigration{operation: tt.operation, state: tt.state}
		if got := m.phase(); got != tt.want {
			t.Fatalf("%s %s phase = %v, want %v", tt.operation, tt.state, got, tt.want)
		}
	}
}

func TestSlotMigrationNodeObserve(t *testing.T) {
	n := &slotMigrationNode{}

	// first poll: finished jobs kept in the server log are only recorded
	done, phase := n.observe([]slotMigration{{name: "old", state: "success"}, {name: "old2", state: "failed"}})
	if len(done) != 0 || phase != slotMigrationPhaseIdle {
		t.Fatalf("baseline should not report completions, got %v %v", done, phase)
	}
	done, phase = n.observe([]slotMigration{{name: "old", state: "success"}, {name: "j1", state: "snapshotting"}})
	if len(done) != 0 || phase != slotMigrationPhaseActive {
		t.Fatalf("unexpected %v %v", done, phase)
	}
	done, phase = n.observe([]slotMigration{{name: "old", state: "success"}, {name: "j1", operation: "EXPORT", state: "failover-granted"}})
	if len(done) != 0 || phase != slotMigrationPhaseFinalizing {
		t.Fatalf("unexpected %v %v", done, phase)
	}
	done, phase = n.observe([]slotMigration{{name: "old", state: "success"}, {name: "j1", state: "success"}})
	if len(done) != 1 || done[0].name != "j1" || phase != slotMigrationPhaseIdle {
		t.Fatalf("expected j1 completion, got %v %v", done, phase)
	}
	// the same finished job is reported only once
	done, _ = n.observe([]slotMigration{{name: "old", state: "success"}, {name: "j1", state: "success"}})
	if len(done) != 0 {
		t.Fatalf("completion reported twice: %v", done)
	}
	// a job that started and finished between two polls is still reported
	done, _ = n.observe([]slotMigration{{name: "j1", state: "success"}, {name: "fast", state: "success"}})
	if len(done) != 1 || done[0].name != "fast" {
		t.Fatalf("expected fast completion, got %v", done)
	}
	// failed and cancelled jobs never trigger a refresh, whether they were seen running or not
	n.observe([]slotMigration{{name: "j2", state: "replicating"}, {name: "j3", state: "snapshotting"}})
	done, _ = n.observe([]slotMigration{{name: "j2", state: "failed"}, {name: "j3", state: "cancelled"}, {name: "j4", state: "failed"}})
	if len(done) != 0 {
		t.Fatalf("unexpected %v", done)
	}
	// jobs that drop out of the log are forgotten
	if _, ok := n.seen["old"]; ok {
		t.Fatal("expected pruned entry")
	}
	// completions are reported for the target side too
	done, _ = n.observe([]slotMigration{{name: "imp", operation: "IMPORT", state: "receiving-snapshot"}})
	if len(done) != 0 {
		t.Fatalf("unexpected %v", done)
	}
	done, _ = n.observe([]slotMigration{{name: "imp", operation: "IMPORT", state: "success"}})
	if len(done) != 1 || done[0].operation != "IMPORT" {
		t.Fatalf("expected an IMPORT completion, got %v", done)
	}
}

// ---------- monitor ----------

type scriptedConn struct {
	mockConn
	mu       sync.Mutex
	polls    int
	reply    ValkeyResult
	topology ValkeyResult
	shards   int
}

func newScriptedConn(addr string, version int) *scriptedConn {
	s := &scriptedConn{reply: migReply()}
	s.AddrFn = func() string { return addr }
	s.VersionFn = func() int { return version }
	s.DoFn = func(cmd Completed) ValkeyResult {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch {
		case isCmd(cmd, "CLUSTER", "GETSLOTMIGRATIONS"):
			s.polls++
			return s.reply
		case isCmd(cmd, "CLUSTER", "SHARDS"), isCmd(cmd, "CLUSTER", "SLOTS"):
			s.shards++
			return s.topology
		}
		return NewResult(strmsg(typeSimpleString, "OK"), nil)
	}
	return s
}

func (s *scriptedConn) set(r ValkeyResult) {
	s.mu.Lock()
	s.reply = r
	s.mu.Unlock()
}

func (s *scriptedConn) setTopology(r ValkeyResult) {
	s.mu.Lock()
	s.topology = r
	s.mu.Unlock()
}

func (s *scriptedConn) counts() (polls, shards int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.polls, s.shards
}

type refreshCall struct {
	completed []slotMigrationCompletion
}

func newTestMonitor(t *testing.T, opt ClusterOption, owners ...conn) (*slotMigrationMonitor, *time.Time, *[]refreshCall) {
	t.Helper()
	c := &clusterClient{
		opt:    &ClientOption{ClusterOption: opt},
		conns:  map[string]connrole{},
		stopCh: make(chan struct{}),
	}
	per := 16384 / len(owners)
	for i := range c.wslots {
		idx := i / per
		if idx >= len(owners) {
			idx = len(owners) - 1
		}
		c.wslots[i] = owners[idx]
	}
	for _, o := range owners {
		c.conns[o.Addr()] = connrole{conn: o}
	}
	m := newSlotMigrationMonitor(c)
	now := time.Unix(1_700_000_000, 0)
	m.now = func() time.Time { return now }
	m.jitter = func(d time.Duration) time.Duration { return d } // deterministic schedule
	var calls []refreshCall
	m.refresh = func(completed []slotMigrationCompletion) {
		calls = append(calls, refreshCall{completed: append([]slotMigrationCompletion(nil), completed...)})
	}
	return m, &now, &calls
}

// TestSlotMigrationAwarenessOptions checks the interval defaults and overrides, that the monitor is opt-in, and
// that it is not started when the client fails to build.
func TestSlotMigrationAwarenessOptions(t *testing.T) {
	m, _, _ := newTestMonitor(t, ClusterOption{}, newScriptedConn("a:1", 9))
	if m.idle != defaultSlotMigrationIdleInterval || m.active != defaultSlotMigrationActiveInterval || m.finalizing != defaultSlotMigrationFinalizingInterval {
		t.Fatalf("unexpected defaults %v %v %v", m.idle, m.active, m.finalizing)
	}
	m, _, _ = newTestMonitor(t, ClusterOption{
		SlotMigrationIdleInterval:       time.Second,
		SlotMigrationActiveInterval:     100 * time.Millisecond,
		SlotMigrationFinalizingInterval: time.Millisecond,
	}, newScriptedConn("a:1", 9))
	if m.idle != time.Second || m.active != 100*time.Millisecond || m.finalizing != time.Millisecond {
		t.Fatalf("unexpected intervals %v %v %v", m.idle, m.active, m.finalizing)
	}

	off := newScriptedConn("127.0.0.1:1", 9)
	dst := newScriptedConn("127.0.0.2:2", 9)
	off.setTopology(topologyBefore)
	dst.setTopology(topologyBefore)
	newRefreshTestClient(t, ClusterOption{}, off, dst)

	failed := newScriptedConn("127.0.0.3:3", 9)
	failed.setTopology(topologyBefore)
	_, err := newClusterClient(
		&ClientOption{InitAddress: []string{failed.Addr()}, ClusterOption: ClusterOption{SlotMigrationAwareness: true, ShardsRefreshInterval: -1}},
		func(addr string, _ *ClientOption) conn { return failed },
		newRetryer(defaultRetryDelayFn),
	)
	if err != ErrInvalidShardsRefreshInterval {
		t.Fatalf("unexpected err %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if p, _ := off.counts(); p != 0 {
		t.Fatalf("the monitor must be opt-in, got %d polls", p)
	}
	if p, _ := failed.counts(); p != 0 {
		t.Fatalf("the monitor must not start when the client fails to build, got %d polls", p)
	}
}

func TestSlotMigrationMonitorLifecycleOfAJob(t *testing.T) {
	src := newScriptedConn("127.0.0.1:1", 9)
	other := newScriptedConn("127.0.0.2:2", 9)
	m, now, calls := newTestMonitor(t, ClusterOption{}, src, other)
	start := *now

	step := func(reply ValkeyResult, wantDelay time.Duration) {
		t.Helper()
		src.set(reply)
		d, ok := m.tick()
		if !ok {
			t.Fatal("monitor stopped unexpectedly")
		}
		if d != wantDelay {
			t.Fatalf("delay = %v, want %v", d, wantDelay)
		}
		*now = now.Add(d)
	}

	step(migReply(exportEntry("old", "success", "0-1")), m.idle) // baseline: old finished job ignored
	step(migReply(exportEntry("old", "success", "0-1"), exportEntry("j1", "snapshotting", "0-99")), m.active)
	step(migReply(exportEntry("j1", "replicating", "0-99")), m.active)
	step(migReply(exportEntry("j1", "waiting-to-pause", "0-99")), m.finalizing)
	step(migReply(exportEntry("j1", "waiting-to-pause", "0-99")), m.finalizing) // a job can stay in this state for several polls
	step(migReply(exportEntry("j1", "failover-granted", "0-99")), m.finalizing)
	if len(*calls) != 0 {
		t.Fatalf("refresh before success: %v", *calls)
	}
	step(migReply(exportEntry("j1", "success", "0-99")), m.active) // cooldown keeps the source at the active cadence
	if len(*calls) != 1 || len((*calls)[0].completed) != 1 {
		t.Fatalf("expected one refresh, got %+v", *calls)
	}
	if got := (*calls)[0].completed[0]; got.reporter != src || got.migration.name != "j1" || got.migration.operation != "EXPORT" ||
		!reflect.DeepEqual(got.migration.slots, [][2]int64{{0, 99}}) {
		t.Fatalf("unexpected completion %+v", got)
	}

	// the finished job stays in the log but must not trigger another refresh
	for i := 0; i < 5; i++ {
		src.set(migReply(exportEntry("j1", "success", "0-99")))
		d, _ := m.tick()
		*now = now.Add(d)
	}
	if len(*calls) != 1 {
		t.Fatalf("refresh repeated for a finished job: %d", len(*calls))
	}

	// after the cooldown, the source returns to the idle cadence
	*now = now.Add(slotMigrationCooldown)
	if d, _ := m.tick(); d != m.idle {
		t.Fatalf("expected the idle cadence after the cooldown, got %v", d)
	}

	// the node without migrations was only polled at the idle cadence
	if polls, maxPolls := func() (int, int) { p, _ := other.counts(); return p, int(now.Sub(start)/m.idle) + 1 }(); polls > maxPolls {
		t.Fatalf("idle node polled too often: %d > %d", polls, maxPolls)
	}
}

func TestSlotMigrationMonitorTransientErrors(t *testing.T) {
	n := newScriptedConn("127.0.0.1:1", 9)
	m, now, _ := newTestMonitor(t, ClusterOption{}, n)
	n.set(migReply(exportEntry("j", "snapshotting", "0")))
	m.tick()
	*now = now.Add(m.active)
	n.set(NewErrorResult(errors.New("i/o timeout")))
	if d, ok := m.tick(); !ok || d != m.active {
		t.Fatalf("network errors keep the cadence, got %v %v", d, ok)
	}
	*now = now.Add(m.active)
	n.set(NewResult(strmsg(typeBlobString, "not an array"), nil))
	if d, ok := m.tick(); !ok || d != m.active {
		t.Fatalf("parse errors keep the cadence, got %v %v", d, ok)
	}
}

func TestSlotMigrationMonitorTracksTopologyChanges(t *testing.T) {
	a := newScriptedConn("127.0.0.1:1", 9)
	b := newScriptedConn("127.0.0.2:2", 9)
	m, now, _ := newTestMonitor(t, ClusterOption{}, a, b)
	m.tick()
	if len(m.nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(m.nodes))
	}
	m.c.mu.Lock()
	for i := range m.c.wslots {
		m.c.wslots[i] = b
	}
	m.c.mu.Unlock()
	*now = now.Add(m.idle)
	m.tick()
	if _, ok := m.nodes[a]; ok || len(m.nodes) != 1 {
		t.Fatalf("node without slots should be dropped, nodes=%d", len(m.nodes))
	}
}

func TestSlotMigrationMonitorReplicaOnlyPollsAllNodes(t *testing.T) {
	a := newScriptedConn("127.0.0.1:1", 9)
	b := newScriptedConn("127.0.0.2:2", 9)
	m, _, _ := newTestMonitor(t, ClusterOption{}, a)
	m.c.opt.ReplicaOnly = true
	m.c.conns[b.Addr()] = connrole{conn: b}
	m.c.conns["hidden:3"] = connrole{conn: newScriptedConn("hidden:3", 9), hidden: true}
	if got := m.primaries(); len(got) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(got))
	}
}

// ---------- refresh path ----------

func newRefreshTestClient(t *testing.T, opt ClusterOption, src, dst *scriptedConn) *clusterClient {
	t.Helper()
	conns := map[string]*scriptedConn{src.Addr(): src, dst.Addr(): dst}
	client, err := newClusterClient(
		&ClientOption{InitAddress: []string{src.Addr()}, ClusterOption: opt},
		func(addr string, _ *ClientOption) conn {
			if c, ok := conns[addr]; ok {
				return c
			}
			return newScriptedConn(addr, 9)
		},
		newRetryer(defaultRetryDelayFn),
	)
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}
	t.Cleanup(client.Close)
	return client
}

var (
	topologyBefore = testShards(
		shardSpec{ip: "127.0.0.1", port: 1, slots: []int64{0, 8191}},
		shardSpec{ip: "127.0.0.2", port: 2, slots: []int64{8192, 16383}},
	)
	topologyAfter = testShards( // slots 0-99 moved from 127.0.0.1:1 to 127.0.0.2:2
		shardSpec{ip: "127.0.0.1", port: 1, slots: []int64{100, 8191}},
		shardSpec{ip: "127.0.0.2", port: 2, slots: []int64{0, 99, 8192, 16383}},
	)
)

// TestSlotMigrationRefreshFromReporter checks that a completion refreshes the topology from the node that reported
// it, which already knows the new owner, without asking the other node, whose view may still be stale, and that the
// refresh is repeated while the reporter's own answer is still stale.
func TestSlotMigrationRefreshFromReporter(t *testing.T) {
	for _, op := range []string{"EXPORT", "IMPORT"} {
		t.Run(op, func(t *testing.T) {
			src := newScriptedConn("127.0.0.1:1", 9)
			dst := newScriptedConn("127.0.0.2:2", 9)
			src.setTopology(topologyBefore)
			dst.setTopology(topologyBefore)
			client := newRefreshTestClient(t, ClusterOption{SlotMigrationFinalizingInterval: time.Millisecond}, src, dst)
			if got := client._pick(0, false); got != src {
				t.Fatalf("slot 0 should start on the source")
			}
			reporter, other := src, dst // the source reports EXPORT, the target IMPORT
			if op == "IMPORT" {
				reporter, other = dst, src
			}

			// the reporter's first answer is still stale, the second one is fresh
			var calls int32
			reporter.DoFn = func(cmd Completed) ValkeyResult {
				if isCmd(cmd, "CLUSTER", "SHARDS") {
					if atomic.AddInt32(&calls, 1) == 1 {
						return topologyBefore
					}
					return topologyAfter
				}
				return migReply()
			}
			_, otherShards := other.counts()
			newSlotMigrationMonitor(client).refreshFromReporters([]slotMigrationCompletion{{
				reporter:  reporter,
				migration: slotMigration{name: "j", operation: op, state: "success", slots: [][2]int64{{0, 99}}},
			}})
			if client._pick(0, false) != dst || client._pick(99, false) != dst {
				t.Fatal("slots 0-99 should move to the target")
			}
			if client._pick(100, false) != src {
				t.Fatal("slot 100 should stay on the source")
			}
			if n := atomic.LoadInt32(&calls); n != 2 {
				t.Fatalf("expected 2 refresh attempts, got %d", n)
			}
			if _, n := other.counts(); n != otherShards {
				t.Fatalf("the refresh should only ask the reporting node, the other node was asked %d times", n-otherShards)
			}
		})
	}
}

func TestSlotMigrationRefreshExpeditesPendingLazyRefresh(t *testing.T) {
	src := newScriptedConn("127.0.0.1:1", 9)
	dst := newScriptedConn("127.0.0.2:2", 9)
	src.setTopology(topologyBefore)
	dst.setTopology(topologyBefore)
	client := newRefreshTestClient(t, ClusterOption{SlotMigrationFinalizingInterval: time.Millisecond}, src, dst)

	// a MOVED reply at cutover has scheduled a jittered refresh far in the future
	client.sc.DelayDo(time.Hour, client._refresh)
	src.setTopology(topologyAfter)
	dst.setTopology(topologyAfter)

	m := newSlotMigrationMonitor(client)
	start := time.Now()
	m.refreshFromReporters([]slotMigrationCompletion{exportDone(src)})
	if time.Since(start) > time.Second {
		t.Fatalf("refresh waited for the delayed flight: %v", time.Since(start))
	}
	if got := client._pick(0, false); got != dst {
		t.Fatal("slot 0 should move to the target")
	}
}

// ---------- compatibility ----------

func TestIsSlotMigrationUnsupported(t *testing.T) {
	for msg, want := range map[string]bool{
		"ERR unknown subcommand 'GETSLOTMIGRATIONS'. Try CLUSTER HELP.":                                  true, // Valkey 8 / Redis 7
		"ERR Unknown subcommand or wrong number of arguments for 'GETSLOTMIGRATIONS'. Try CLUSTER HELP.": true, // Redis 5
		"ERR unknown command 'CLUSTER', with args beginning with: 'GETSLOTMIGRATIONS'":                   true, // renamed / disabled
		"ERR command is not allowed":                      true, // managed services
		"ERR This instance has cluster support disabled":  false,
		"LOADING Valkey is loading the dataset in memory": false,
		"TRYAGAIN": false,
	} {
		if got := isSlotMigrationUnsupported(msg); got != want {
			t.Fatalf("%q: got %v, want %v", msg, got, want)
		}
	}
}

// TestSlotMigrationMonitorServerSupport checks which nodes are polled: only Valkey 9.0+ servers, including over
// RESP2 where the version comes from HELLO; a node without the command, or that blocks it, is polled once; and
// NOPERM stops the monitor, because the ACL user cannot run the command anywhere in the cluster.
func TestSlotMigrationMonitorServerSupport(t *testing.T) {
	running := NewResult(slicemsg(typeArray, []ValkeyMessage{ // RESP2 flat array
		migEntry(false, "name", "j", "operation", "EXPORT", "slot_ranges", "0-10", "state", "snapshotting"),
	}), nil)
	tests := []struct {
		name      string
		version   int    // conn.Version(), which is 5 on every RESP2 connection
		hello     string // server version reported by HELLO, if any
		reply     ValkeyResult
		wantPolls int
		wantDelay time.Duration // after the first tick
		wantStop  bool
	}{
		{name: "pre-9.0 server", version: 8, reply: migReply(), wantDelay: defaultSlotMigrationIdleInterval},
		{name: "unknown version", version: 0, reply: migReply(), wantDelay: defaultSlotMigrationIdleInterval},
		{name: "Valkey 9 over RESP2", version: 5, hello: "9.0.6", reply: running, wantPolls: 3, wantDelay: defaultSlotMigrationActiveInterval},
		{name: "command not available", version: 9, reply: NewResult(strmsg(typeSimpleErr, "ERR unknown subcommand 'GETSLOTMIGRATIONS'. Try CLUSTER HELP."), nil), wantPolls: 1, wantDelay: defaultSlotMigrationIdleInterval},
		{name: "NOPERM", version: 9, reply: NewResult(strmsg(typeSimpleErr, "NOPERM User app has no permissions to run the 'cluster|getslotmigrations' command"), nil), wantPolls: 1, wantStop: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newScriptedConn("127.0.0.1:1", tt.version)
			if tt.hello != "" {
				n.InfoFn = func() map[string]ValkeyMessage {
					return map[string]ValkeyMessage{"version": strmsg(typeBlobString, tt.hello)}
				}
			}
			n.set(tt.reply)
			m, now, _ := newTestMonitor(t, ClusterOption{}, n)
			stopped := false
			for i := 0; i < 3 && !stopped; i++ {
				d, ok := m.tick()
				if stopped = !ok; !stopped && i == 0 && d != tt.wantDelay {
					t.Fatalf("delay after the first poll = %v, want %v", d, tt.wantDelay)
				}
				*now = now.Add(time.Hour)
			}
			if stopped != tt.wantStop {
				t.Fatalf("monitor stopped = %v, want %v", stopped, tt.wantStop)
			}
			if p, _ := n.counts(); p != tt.wantPolls {
				t.Fatalf("polls = %d, want %d", p, tt.wantPolls)
			}
		})
	}
}

// ---------- target side, MOVED wake, deduplication, concurrency and shutdown ----------

func importEntry(name, state, slots string) ValkeyMessage {
	return migEntry(true, "name", name, "operation", "IMPORT", "slot_ranges", slots, "state", state)
}

func exportDone(src conn) slotMigrationCompletion {
	return slotMigrationCompletion{reporter: src, migration: slotMigration{name: "j", operation: "EXPORT", state: "success", slots: [][2]int64{{0, 99}}}}
}

func TestSlotMigrationMonitorMovedWake(t *testing.T) {
	src := newScriptedConn("127.0.0.1:1", 9)
	dst := newScriptedConn("127.0.0.2:2", 9)
	m, now, calls := newTestMonitor(t, ClusterOption{}, src, dst)
	m.tick() // baseline; both nodes idle until now+idle
	if d := m.delay(*now); d != m.idle {
		t.Fatalf("delay = %v, want %v", d, m.idle)
	}
	if d := m.delay(now.Add(2 * m.idle)); d != 0 {
		t.Fatalf("an overdue poll should be due immediately, delay = %v", d)
	}

	// a job ran between two idle polls; the first MOVED names the target, which already reports success
	*now = now.Add(time.Second)
	dst.set(migReply(importEntry("j", "success", "0-99")))
	if !m.wake("127.0.0.2:2") {
		t.Fatal("the MOVED target should be scheduled")
	}
	if d := m.delay(*now); d != 0 {
		t.Fatalf("the MOVED target should be due now, delay = %v", d)
	}
	m.tick()
	if len(*calls) != 1 || (*calls)[0].completed[0].reporter != dst {
		t.Fatalf("expected a refresh from the target, got %+v", *calls)
	}
	if p, _ := src.counts(); p != 1 {
		t.Fatalf("only the MOVED target should be polled early, source polled %d times", p)
	}

	// nodes that are not monitored primaries are ignored
	if m.wake("10.0.0.9:9") {
		t.Fatal("unknown address should be ignored")
	}
	m.nodes[src].unsupported = true
	if m.wake("127.0.0.1:1") {
		t.Fatal("unsupported node should be ignored")
	}
}

func TestSlotMigrationMonitorMovedWakeRateLimit(t *testing.T) {
	src := newScriptedConn("127.0.0.1:1", 9)
	dst := newScriptedConn("127.0.0.2:2", 9)
	m, now, _ := newTestMonitor(t, ClusterOption{}, src, dst)
	m.tick() // baseline
	polled := *now

	// no known migration: a burst of MOVED replies polls the node at most once per active interval
	*now = now.Add(10 * time.Millisecond)
	if !m.wake("127.0.0.2:2") {
		t.Fatal("expected the node to be scheduled")
	}
	if got := m.nodes[dst].next; !got.Equal(polled.Add(m.active)) {
		t.Fatalf("next poll = %v after the previous one, want %v", got.Sub(polled), m.active)
	}
	if m.wake("127.0.0.2:2") {
		t.Fatal("a second MOVED must not move the poll earlier")
	}

	// while a migration is known to run, the finalizing interval applies
	m, now, _ = newTestMonitor(t, ClusterOption{}, src, dst)
	src.set(migReply(exportEntry("j", "snapshotting", "0-99")))
	m.tick() // first poll of both nodes; the source reports a running job
	polled = *now
	*now = now.Add(5 * time.Millisecond)
	if !m.wake("127.0.0.2:2") {
		t.Fatal("expected the node to be scheduled")
	}
	if got := m.nodes[dst].next; !got.Equal(polled.Add(m.finalizing)) {
		t.Fatalf("next poll = %v after the previous one, want %v", got.Sub(polled), m.finalizing)
	}
}

func TestClusterClientMovedNotifiesSlotMigrationMonitor(t *testing.T) {
	src := newScriptedConn("127.0.0.1:1", 9)
	dst := newScriptedConn("127.0.0.2:2", 9)
	src.setTopology(topologyBefore)
	dst.setTopology(topologyBefore)
	client := newRefreshTestClient(t, ClusterOption{}, src, dst)
	client.migMoved = make(chan string, 1) // as if the monitor were enabled, without its goroutine

	moved := NewResult(strmsg('-', "MOVED 0 127.0.0.2:2"), nil).Error()
	ask := NewResult(strmsg('-', "ASK 0 127.0.0.2:2"), nil).Error()
	if _, mode := client.shouldRefreshRetry(ask, context.Background()); mode != RedirectAsk || len(client.migMoved) != 0 {
		t.Fatalf("ASK must not wake the monitor, mode %v queued %d", mode, len(client.migMoved))
	}
	if _, mode := client.shouldRefreshRetry(moved, context.Background()); mode != RedirectMove {
		t.Fatalf("unexpected mode %v", mode)
	}
	if got := <-client.migMoved; got != "127.0.0.2:2" {
		t.Fatalf("monitor got %q, want the MOVED target", got)
	}
	// a full queue never blocks the command path
	done := make(chan struct{})
	go func() {
		for i := 0; i < 3; i++ {
			client.shouldRefreshRetry(moved, context.Background())
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("shouldRefreshRetry blocked on a full monitor queue")
	}
}

func TestSlotMigrationMonitorOneRefreshPerJob(t *testing.T) {
	src := newScriptedConn("127.0.0.1:1", 9)
	dst := newScriptedConn("127.0.0.2:2", 9)
	m, now, calls := newTestMonitor(t, ClusterOption{}, src, dst)
	m.tick() // baseline

	// both sides report the same job finishing in the same poll
	*now = now.Add(m.idle)
	src.set(migReply(exportEntry("j", "success", "0-99")))
	dst.set(migReply(importEntry("j", "success", "0-99")))
	m.tick()
	if len(*calls) != 1 || len((*calls)[0].completed) != 1 {
		t.Fatalf("expected one completion for the job, got %+v", *calls)
	}

	// the source reports a second job first, the target only on a later poll
	*now = now.Add(m.idle)
	src.set(migReply(exportEntry("j", "success", "0-99"), exportEntry("k", "success", "100")))
	m.tick()
	*now = now.Add(m.idle)
	dst.set(migReply(importEntry("j", "success", "0-99"), importEntry("k", "success", "100")))
	m.tick()
	if len(*calls) != 2 {
		t.Fatalf("expected exactly one refresh per job, got %d", len(*calls))
	}
}

func TestJitterSlotMigrationInterval(t *testing.T) {
	for _, d := range []time.Duration{100 * time.Millisecond, time.Second, 30 * time.Second} {
		lo, hi := d-d/5, d+d/5
		for i := 0; i < 1000; i++ {
			if got := jitterSlotMigrationInterval(d); got < lo || got > hi {
				t.Fatalf("jitter(%v) = %v, want within [%v, %v]", d, got, lo, hi)
			}
		}
	}
	if got := jitterSlotMigrationInterval(3); got != 3 {
		t.Fatalf("tiny intervals are returned unchanged, got %v", got)
	}
}

func TestSlotMigrationMonitorBoundedConcurrency(t *testing.T) {
	var running, peak int32
	owners := make([]conn, 40)
	for i := range owners {
		c := newScriptedConn(fmt.Sprintf("127.0.%d.1:1", i), 9)
		c.DoFn = func(cmd Completed) ValkeyResult {
			n := atomic.AddInt32(&running, 1)
			for {
				p := atomic.LoadInt32(&peak)
				if n <= p || atomic.CompareAndSwapInt32(&peak, p, n) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&running, -1)
			return migReply()
		}
		owners[i] = c
	}
	m, _, _ := newTestMonitor(t, ClusterOption{}, owners...)
	m.tick()
	if p := atomic.LoadInt32(&peak); p > slotMigrationMaxConcurrentPolls || p < 2 {
		t.Fatalf("expected concurrent polls capped at %d, peak %d", slotMigrationMaxConcurrentPolls, p)
	}
	if len(m.nodes) != 40 {
		t.Fatalf("expected all 40 primaries to be tracked, got %d", len(m.nodes))
	}
}

func TestSlotMigrationMonitorPrunesHandledJobs(t *testing.T) {
	src := newScriptedConn("127.0.0.1:1", 9)
	m, now, _ := newTestMonitor(t, ClusterOption{}, src)
	m.handled["gone"] = *now
	m.handled["kept"] = *now
	src.set(migReply(exportEntry("kept", "success", "0")))
	*now = now.Add(slotMigrationHandledRetention + time.Second)
	m.tick()
	if _, ok := m.handled["gone"]; ok {
		t.Fatal("a job no longer in any log should be forgotten after the retention period")
	}
	if _, ok := m.handled["kept"]; !ok {
		t.Fatal("a job still in a log must stay handled")
	}
}

// ctxConn blocks CLUSTER GETSLOTMIGRATIONS until its context is done.
type ctxConn struct {
	*scriptedConn
	started chan struct{}
	once    sync.Once
}

func (c *ctxConn) Do(ctx context.Context, cmd Completed) ValkeyResult {
	if isCmd(cmd, "CLUSTER", "GETSLOTMIGRATIONS") {
		c.once.Do(func() { close(c.started) })
		<-ctx.Done()
		return NewErrorResult(ctx.Err())
	}
	return c.scriptedConn.Do(ctx, cmd)
}

func TestSlotMigrationCloseCancelsInflightPoll(t *testing.T) {
	src := &ctxConn{scriptedConn: newScriptedConn("127.0.0.1:1", 9), started: make(chan struct{})}
	dst := newScriptedConn("127.0.0.2:2", 9)
	src.setTopology(topologyBefore)
	dst.setTopology(topologyBefore)
	client, err := newClusterClient(
		&ClientOption{InitAddress: []string{src.Addr()}, ClusterOption: ClusterOption{SlotMigrationAwareness: true}},
		func(addr string, _ *ClientOption) conn {
			if addr == src.Addr() {
				return src
			}
			return dst
		},
		newRetryer(defaultRetryDelayFn),
	)
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}
	select {
	case <-src.started:
	case <-time.After(5 * time.Second):
		t.Fatal("the monitor did not poll")
	}
	start := time.Now()
	client.Close()
	if d := time.Since(start); d > slotMigrationPollTimeout/2 {
		t.Fatalf("Close waited %v for an in-flight poll instead of cancelling it", d)
	}
	select {
	case <-client.migDone:
	default:
		t.Fatal("the monitor goroutine must have exited when Close returns")
	}
	client.Close() // a second Close must not block or panic
}

// TestSlotMigrationAwarenessNoGoroutineLeak runs the monitor through polls, a MOVED wake and a refresh, and
// checks that closing the client leaves no goroutine behind.
func TestSlotMigrationAwarenessNoGoroutineLeak(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	src := newScriptedConn("127.0.0.1:1", 9)
	dst := newScriptedConn("127.0.0.2:2", 9)
	src.setTopology(topologyBefore)
	dst.setTopology(topologyBefore)
	client := newRefreshTestClient(t, ClusterOption{
		SlotMigrationAwareness:          true,
		SlotMigrationIdleInterval:       5 * time.Millisecond,
		SlotMigrationActiveInterval:     2 * time.Millisecond,
		SlotMigrationFinalizingInterval: time.Millisecond,
	}, src, dst)
	defer client.Close() // runs before the deferred leak check, also when the test fails
	deadline := time.Now().Add(5 * time.Second)
	for p, _ := dst.counts(); p < 2; p, _ = dst.counts() {
		if time.Now().After(deadline) {
			t.Fatal("monitor did not poll")
		}
		time.Sleep(time.Millisecond)
	}
	dst.setTopology(topologyAfter)
	dst.set(migReply(importEntry("j", "success", "0-99")))
	client.shouldRefreshRetry(NewResult(strmsg('-', "MOVED 0 127.0.0.2:2"), nil).Error(), context.Background())
	for client._pick(0, false) != dst {
		if time.Now().After(deadline) {
			t.Fatal("slot map did not follow the migration")
		}
		time.Sleep(time.Millisecond)
	}
}
