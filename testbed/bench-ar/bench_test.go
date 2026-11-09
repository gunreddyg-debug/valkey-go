package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCRC16(t *testing.T) {
	if got := crc16([]byte("123456789")); got != 0x31C3 {
		t.Fatalf("crc16(123456789) = %#x, want 0x31c3", got)
	}
	// From CLUSTER KEYSLOT.
	for key, want := range map[string]int64{"foo": 12182, "bar": 5061, "hello": 866, "{user1000}.following": 3443, "{user1000}.followers": 3443} {
		if got := keySlot(key); got != want {
			t.Errorf("keySlot(%q) = %d, want %d", key, got, want)
		}
	}
}

func TestCountingConnMatchesSplitCommands(t *testing.T) {
	stream := append(append(append([]byte("*1\r\n$4\r\nPING\r\n"), shardsCmd...), []byte("*2\r\n$3\r\nGET\r\n$1\r\nk\r\n")...), shardsCmd...)
	stream = append(stream, shardsCmd...)
	for split := 1; split <= len(stream); split++ {
		c := &countingConn{}
		n := 0
		for i := 0; i < len(stream); i += split {
			n += c.count(stream[i:min(i+split, len(stream))])
		}
		if n != 3 {
			t.Fatalf("chunks of %d bytes: counted %d CLUSTER SHARDS, want 3", split, n)
		}
	}
	if n := (&countingConn{}).count(bytes.Repeat([]byte("x"), 100)); n != 0 {
		t.Fatalf("counted %d in unrelated bytes", n)
	}
}

// A CLUSTER SHARDS reply (RESP2) in the shape valkey returns it, written by hand to show a failover:
// the primary 7101 is flagged fail and its replica 7106 is not promoted yet. The node fields are the
// ones in testdata/cluster_shards_healthy.resp, which was captured from the test bed (valkey 9.0.6).
const shardsReplyFailingOver = "*2\r\n" +
	"*4\r\n$5\r\nslots\r\n*2\r\n:0\r\n:5460\r\n$5\r\nnodes\r\n*2\r\n" +
	"*14\r\n$2\r\nid\r\n$3\r\naaa\r\n$4\r\nport\r\n:7101\r\n$2\r\nip\r\n$9\r\n127.0.0.1\r\n$8\r\nendpoint\r\n$9\r\n127.0.0.1\r\n$4\r\nrole\r\n$6\r\nmaster\r\n$18\r\nreplication-offset\r\n:100\r\n$6\r\nhealth\r\n$4\r\nfail\r\n" +
	"*14\r\n$2\r\nid\r\n$3\r\nbbb\r\n$4\r\nport\r\n:7106\r\n$2\r\nip\r\n$9\r\n127.0.0.1\r\n$8\r\nendpoint\r\n$9\r\n127.0.0.1\r\n$4\r\nrole\r\n$7\r\nreplica\r\n$18\r\nreplication-offset\r\n:100\r\n$6\r\nhealth\r\n$6\r\nonline\r\n" +
	"*4\r\n$5\r\nslots\r\n*4\r\n:5461\r\n:10922\r\n:10923\r\n:16383\r\n$5\r\nnodes\r\n*1\r\n" +
	"*14\r\n$2\r\nid\r\n$3\r\nccc\r\n$4\r\nport\r\n:7102\r\n$2\r\nip\r\n$9\r\n127.0.0.1\r\n$8\r\nendpoint\r\n$9\r\n127.0.0.1\r\n$4\r\nrole\r\n$6\r\nmaster\r\n$18\r\nreplication-offset\r\n:100\r\n$6\r\nhealth\r\n$6\r\nonline\r\n"

func TestParseShards(t *testing.T) {
	v, err := readRESP(bufio.NewReader(strings.NewReader(shardsReplyFailingOver)))
	if err != nil {
		t.Fatal(err)
	}
	shards, err := parseShards(v)
	if err != nil {
		t.Fatal(err)
	}
	s0, ok := shardOf(shards, 0)
	if !ok || len(s0.Nodes) != 2 || s0.Nodes[0].Addr != "127.0.0.1:7101" || s0.Nodes[0].Health != "fail" {
		t.Fatalf("shard of slot 0: %+v %v", s0, ok)
	}
	if p, ok := s0.onlinePrimary(); ok {
		t.Fatalf("slot 0 has online primary %s, want none", p)
	}
	s1, ok := shardOf(shards, 16000)
	if p, _ := s1.onlinePrimary(); !ok || p != "127.0.0.1:7102" {
		t.Fatalf("primary of slot 16000 = %q, want 127.0.0.1:7102", p)
	}
	keys, err := keysInShard(50, s0)
	if err != nil || len(keys) != 50 {
		t.Fatalf("keysInShard: %v %v", keys, err)
	}
	for _, k := range keys {
		if s := keySlot(k); s > 5460 {
			t.Fatalf("key %q has slot %d", k, s)
		}
	}
}

func TestParseCapturedShards(t *testing.T) {
	raw, err := os.ReadFile("testdata/cluster_shards_healthy.resp")
	if err != nil {
		t.Fatal(err)
	}
	v, err := readRESP(bufio.NewReader(bytes.NewReader(raw)))
	if err != nil {
		t.Fatal(err)
	}
	shards, err := parseShards(v)
	if err != nil || len(shards) != 3 {
		t.Fatalf("parseShards: %d shards, %v", len(shards), err)
	}
	covered := 0
	for _, s := range shards {
		if _, ok := s.onlinePrimary(); !ok || len(s.Nodes) != 2 {
			t.Fatalf("shard %+v: want an online primary and 2 nodes", s)
		}
		for _, r := range s.Ranges {
			covered += int(r[1] - r[0] + 1)
		}
	}
	if covered != 16384 {
		t.Fatalf("shards cover %d slots", covered)
	}
	if s, ok := shardOf(shards, 0); !ok || s.Ranges[0] != [2]int64{0, 5460} {
		t.Fatalf("shard of slot 0: %+v", s)
	}
}

func TestHistBins(t *testing.T) {
	prev := -1
	for _, us := range []int64{0, 1, 2, 63, 64, 65, 100, 127, 128, 1000, 4095, 4096, 99999, 1 << 20, 1<<25 - 1} {
		i := histIndex(time.Duration(us) * time.Microsecond)
		if i < prev {
			t.Fatalf("histIndex(%dµs) = %d, below the previous %d", us, i, prev)
		}
		prev = i
		if lo, hi := histLow(i), histHigh(i); float64(us) < lo || float64(us) >= hi {
			t.Fatalf("%dµs is in bin %d = [%v, %v)", us, i, lo, hi)
		}
		if us >= histLinear {
			if w := histHigh(i) - histLow(i); w/histLow(i) > 1.0/histSub+1e-9 {
				t.Fatalf("bin %d of %dµs is %v wide", i, us, w)
			}
		}
	}
	if i := histIndex(time.Hour); i != histBins-1 {
		t.Fatalf("1h in bin %d, want the last (%d)", i, histBins-1)
	}
	if histHigh(histBins-1) != 1<<25 {
		t.Fatalf("last bin ends at %v µs", histHigh(histBins-1))
	}
}

func TestSummarizeHist(t *testing.T) {
	hs := make([]hist, 3)
	for i := 0; i < 989; i++ {
		hs[0][histIndex(time.Millisecond)].Add(1)
	}
	for i := 0; i < 10; i++ {
		hs[1][histIndex(10*time.Millisecond)].Add(1)
	}
	hs[2][histIndex(2*time.Second)].Add(1)
	s := summarizeHist(hs, 0, 3)
	near := func(got, want float64) bool { return got >= want*0.97 && got <= want*1.03 }
	if s.Count != 1000 || !near(s.P50, 1) || !near(s.P99, 10) || !near(s.P999, 10) || !near(s.Max, 2000) {
		t.Fatalf("summary %+v", s)
	}
	if s := summarizeHist(hs, 1, 2); s.Count != 10 || !near(s.P50, 10) {
		t.Fatalf("window summary %+v", s)
	}
	if s := summarizeHist(hs, 5, 9); s.Count != 0 {
		t.Fatalf("empty window %+v", s)
	}
}

func TestCountingConnCountsGetSlotMigrations(t *testing.T) {
	stream := append(append(append([]byte{}, getSlotMigrationsCmd...), shardsCmd...), getSlotMigrationsCmd...)
	for split := 1; split <= len(stream); split++ {
		c := &countingConn{}
		ns, ng := 0, 0
		for i := 0; i < len(stream); i += split {
			a, b := c.countAll(stream[i:min(i+split, len(stream))])
			ns, ng = ns+a, ng+b
		}
		if ns != 1 || ng != 2 {
			t.Fatalf("chunks of %d bytes: %d CLUSTER SHARDS, %d GETSLOTMIGRATIONS; want 1, 2", split, ns, ng)
		}
	}
}

const infoSample = "# Server\r\nvalkey_version:9.0.6\r\nrun_id:aaaa\r\ntcp_port:7101\r\n\r\n# Errorstats\r\nerrorstat_ERR:count=3\r\nerrorstat_MOVED:count=120\r\nerrorstat_ASK:count=7\r\n"

func TestParseErrorstats(t *testing.T) {
	r := parseErrorstats(infoSample)
	if r.runID != "aaaa" || r.counts["MOVED"] != 120 || r.counts["ASK"] != 7 || r.counts["ERR"] != 3 || len(r.counts) != 3 {
		t.Fatalf("parsed %+v", r)
	}
}

func TestErrStatsObserve(t *testing.T) {
	e := newErrStats([]string{"n1", "n2"}, 100*time.Millisecond)
	e.observe("n1", errRead{runID: "a", counts: map[string]int64{"MOVED": 10}}, 0, true)
	e.observe("n1", errRead{runID: "a", counts: map[string]int64{"MOVED": 15}}, 1500*time.Millisecond, false)
	e.observe("n1", errRead{runID: "b", counts: map[string]int64{"MOVED": 4}}, 2500*time.Millisecond, false) // restarted
	e.observe("n1", errRead{runID: "b", counts: map[string]int64{"MOVED": 9, "ASK": 1}}, 2600*time.Millisecond, false)
	e.observe("n2", errRead{runID: "c", counts: map[string]int64{"ASK": 2}}, 3*time.Second, false) // no baseline
	s := e.summary(4)
	if s.Totals["MOVED"] != 5+4+5 || s.Totals["ASK"] != 1+2 || s.Restarts["n1"] != 1 {
		t.Fatalf("summary %+v", s)
	}
	if m := s.PerSecond["MOVED"]; len(m) != 4 || m[1] != 5 || m[2] != 9 {
		t.Fatalf("MOVED per second %v", m)
	}
	if at := e.lastIncrease("n1", "MOVED"); at != 2600*time.Millisecond {
		t.Fatalf("last MOVED increase at %v", at)
	}
}

func TestKeysInSlots(t *testing.T) {
	keys, err := keysInSlots(1000, 0, 199)
	if err != nil || len(keys) != 1000 {
		t.Fatalf("keysInSlots: %d keys, %v", len(keys), err)
	}
	slots := map[int64]bool{}
	for _, k := range keys {
		s := keySlot(k)
		if s > 199 {
			t.Fatalf("key %q in slot %d", k, s)
		}
		slots[s] = true
	}
	if len(slots) < 190 {
		t.Fatalf("keys cover only %d of 200 slots", len(slots))
	}
	if lo, hi, err := parseSlotRange("0-199"); err != nil || lo != 0 || hi != 199 {
		t.Fatalf("parseSlotRange: %d %d %v", lo, hi, err)
	}
	if _, _, err := parseSlotRange("5-2"); err == nil {
		t.Fatal("parseSlotRange accepted 5-2")
	}
}

func TestParseSlotMigrationEntries(t *testing.T) {
	reply := "*1\r\n*8\r\n$4\r\nname\r\n$3\r\nabc\r\n$9\r\noperation\r\n$6\r\nEXPORT\r\n$5\r\nstate\r\n$7\r\nsuccess\r\n$11\r\nslot_ranges\r\n$5\r\n0-199\r\n"
	v, err := readRESP(bufio.NewReader(strings.NewReader(reply)))
	if err != nil {
		t.Fatal(err)
	}
	es := parseSlotMigrationEntries(v)
	if len(es) != 1 || es[0].Name != "abc" || es[0].Operation != "EXPORT" || es[0].State != "success" || es[0].Slots != "0-199" {
		t.Fatalf("entries %+v", es)
	}
}
