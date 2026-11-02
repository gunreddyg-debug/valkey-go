package valkey

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/valkey-io/valkey-go/internal/cmds"
)

// Tests for ClusterOption.MinRefreshInterval and ClusterOption.MaxSlotWait, and an opt-in benchmark of a
// mocked failover: VALKEY_FAILOVER_BENCH=1 go test -run TestClusterSlotWaitFailoverBench -v .

// --- shared helpers (begin) ---

// swNode is a node in a CLUSTER SHARDS reply built by swShardsReply.
type swNode struct {
	addr   string // host:port
	role   string // "master" or "replica"
	health string // "online", "fail", ...
}

// swShard is a shard in a CLUSTER SHARDS reply built by swShardsReply.
type swShard struct {
	nodes    []swNode
	from, to int64
}

func swPrimary(addr string) swNode { return swNode{addr: addr, role: "master", health: "online"} }
func swReplica(addr string) swNode { return swNode{addr: addr, role: "replica", health: "online"} }

func swShardsReply(shards ...swShard) ValkeyResult {
	reply := make([]ValkeyMessage, 0, len(shards))
	for _, s := range shards {
		nodes := make([]ValkeyMessage, 0, len(s.nodes))
		for _, n := range s.nodes {
			host, port, _ := net.SplitHostPort(n.addr)
			p, _ := strconv.ParseInt(port, 10, 64)
			nodes = append(nodes, slicemsg(typeMap, []ValkeyMessage{
				strmsg(typeBlobString, "id"), strmsg(typeBlobString, n.addr),
				strmsg(typeBlobString, "port"), {typ: typeInteger, intlen: p},
				strmsg(typeBlobString, "ip"), strmsg(typeBlobString, host),
				strmsg(typeBlobString, "endpoint"), strmsg(typeBlobString, host),
				strmsg(typeBlobString, "role"), strmsg(typeBlobString, n.role),
				strmsg(typeBlobString, "replication-offset"), {typ: typeInteger, intlen: 1},
				strmsg(typeBlobString, "health"), strmsg(typeBlobString, n.health),
			}))
		}
		reply = append(reply, slicemsg(typeMap, []ValkeyMessage{
			strmsg(typeBlobString, "slots"), slicemsg(typeArray, []ValkeyMessage{
				{typ: typeInteger, intlen: s.from}, {typ: typeInteger, intlen: s.to},
			}),
			strmsg(typeBlobString, "nodes"), slicemsg(typeArray, nodes),
		}))
	}
	return NewResult(slicemsg(typeArray, reply), nil)
}

// swKeysInSlots returns n keys whose slots are in [from, to].
func swKeysInSlots(n int, from, to uint16) []string {
	b := cmds.NewBuilder(cmds.InitSlot)
	keys := make([]string, 0, n)
	for i := 0; len(keys) < n; i++ {
		k := "key" + strconv.Itoa(i)
		if cmd := b.Get().Key(k).Build(); cmd.Slot() >= from && cmd.Slot() <= to {
			keys = append(keys, k)
		}
	}
	return keys
}

func swIsClusterShards(cmd Completed) bool {
	c := cmd.Commands()
	return len(c) == 2 && c[0] == "CLUSTER" && c[1] == "SHARDS"
}

// --- shared helpers (end) ---

// swApplyOptions sets the options under test. The copy of the failover benchmark that runs on branches
// without these options replaces it with one that only accepts zero values.
func swApplyOptions(opt *ClusterOption, minRefresh, maxWait time.Duration) bool {
	opt.MinRefreshInterval = minRefresh
	opt.MaxSlotWait = maxWait
	return true
}

const (
	swAddr1 = "10.0.0.1:6379" // primary of slots 0-8191
	swAddr2 = "10.0.0.2:6379" // primary of slots 8192-16383, fails
	swAddr3 = "10.0.0.3:6379" // replica of swAddr2, promoted in its place

	// swSlack bounds how much later than expected a request may return in these tests (scheduling, -race).
	swSlack = 150 * time.Millisecond
)

var (
	swHealthy = []swShard{
		{from: 0, to: 8191, nodes: []swNode{swPrimary(swAddr1)}},
		{from: 8192, to: 16383, nodes: []swNode{swPrimary(swAddr2), swReplica(swAddr3)}},
	}
	// swFailingOver has no online primary for slots 8192-16383, so they have no node in the slot map.
	swFailingOver = []swShard{
		{from: 0, to: 8191, nodes: []swNode{swPrimary(swAddr1)}},
		{from: 8192, to: 16383, nodes: []swNode{{addr: swAddr2, role: "master", health: "fail"}, swReplica(swAddr3)}},
	}
	swPromoted = []swShard{
		{from: 0, to: 8191, nodes: []swNode{swPrimary(swAddr1)}},
		{from: 8192, to: 16383, nodes: []swNode{swPrimary(swAddr3)}},
	}
	swPresentKey = swKeysInSlots(1, 0, 8191)[0]
	swMissingKey = swKeysInSlots(1, 8192, 16383)[0]
)

// swCluster serves CLUSTER SHARDS from a topology that tests change, and answers other commands with the
// address of the node that served them.
type swCluster struct {
	shardsErr error         // CLUSTER SHARDS fails with it when set
	hold      chan struct{} // CLUSTER SHARDS blocks until it is closed when set
	topology  ValkeyResult
	refreshes []time.Time // when each CLUSTER SHARDS was received
	mu        sync.Mutex
}

func newSWCluster(shards []swShard) *swCluster {
	return &swCluster{topology: swShardsReply(shards...)}
}

func (m *swCluster) setTopology(shards []swShard) {
	reply := swShardsReply(shards...)
	m.mu.Lock()
	m.topology = reply
	m.mu.Unlock()
}

func (m *swCluster) setShardsErr(err error) {
	m.mu.Lock()
	m.shardsErr = err
	m.mu.Unlock()
}

// holdShards makes CLUSTER SHARDS block until release is called.
func (m *swCluster) holdShards() (release func()) {
	ch := make(chan struct{})
	m.mu.Lock()
	m.hold = ch
	m.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			m.mu.Lock()
			m.hold = nil
			m.mu.Unlock()
			close(ch)
		})
	}
}

func (m *swCluster) refreshTimes() []time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]time.Time(nil), m.refreshes...)
}

func (m *swCluster) refreshCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.refreshes)
}

func (m *swCluster) do(addr string, cmd Completed) ValkeyResult {
	if !swIsClusterShards(cmd) {
		return NewResult(strmsg(typeBlobString, addr), nil)
	}
	m.mu.Lock()
	m.refreshes = append(m.refreshes, time.Now())
	hold, err, topology := m.hold, m.shardsErr, m.topology
	m.mu.Unlock()
	if hold != nil {
		<-hold
	}
	if err != nil {
		return NewErrorResult(err)
	}
	return topology
}

func (m *swCluster) connFn(dst string, _ *ClientOption) conn {
	return &mockConn{
		AddrFn:    func() string { return dst },
		VersionFn: func() int { return 8 }, // CLUSTER SHARDS
		DoFn:      func(cmd Completed) ValkeyResult { return m.do(dst, cmd) },
		DoCacheFn: func(cmd Cacheable, _ time.Duration) ValkeyResult { return m.do(dst, Completed(cmd)) },
		DoMultiFn: func(multi ...Completed) *valkeyresults {
			resps := make([]ValkeyResult, len(multi))
			for i, cmd := range multi {
				resps[i] = m.do(dst, cmd)
			}
			return &valkeyresults{s: resps}
		},
		DoMultiCacheFn: func(multi ...CacheableTTL) *valkeyresults {
			resps := make([]ValkeyResult, len(multi))
			for i, ct := range multi {
				resps[i] = m.do(dst, Completed(ct.Cmd))
			}
			return &valkeyresults{s: resps}
		},
	}
}

// newSWClient returns a client that refreshes from one discovery endpoint, like the clients this is for.
func newSWClient(t *testing.T, m *swCluster, opt ClusterOption) *clusterClient {
	t.Helper()
	opt.PreferInitAddressRefresh = true
	client, err := newClusterClient(
		&ClientOption{InitAddress: []string{"discovery:6379"}, ClusterOption: opt},
		m.connFn, newRetryer(defaultRetryDelayFn),
	)
	if err != nil {
		t.Fatalf("newClusterClient: %v", err)
	}
	t.Cleanup(client.Close)
	return client
}

func swGet(ctx context.Context, client *clusterClient, key string) (string, time.Duration, error) {
	start := time.Now()
	v, err := client.Do(ctx, client.B().Get().Key(key).Build()).ToString()
	return v, time.Since(start), err
}

func swCheckWithin(t *testing.T, what string, got, from, to time.Duration) {
	t.Helper()
	if got < from || got > to {
		t.Errorf("%s took %v, want between %v and %v", what, got, from, to)
	}
}

// swCheckSpacing checks that refreshes were received at least interval apart.
func swCheckSpacing(t *testing.T, times []time.Time, interval time.Duration) {
	t.Helper()
	for i := 1; i < len(times); i++ {
		if gap := times[i].Sub(times[i-1]); gap < interval {
			t.Errorf("refresh %d started %v after the previous one, want at least %v", i, gap, interval)
		}
	}
}

func TestClusterSlotWaitRejectsNegativeOptions(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	for _, tc := range []struct {
		want error
		name string
		opt  ClusterOption
	}{
		{name: "MinRefreshInterval", opt: ClusterOption{MinRefreshInterval: -time.Millisecond}, want: ErrInvalidMinRefreshInterval},
		{name: "MaxSlotWait", opt: ClusterOption{MaxSlotWait: -time.Millisecond}, want: ErrInvalidMaxSlotWait},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newSWCluster(swHealthy)
			client, err := newClusterClient(&ClientOption{InitAddress: []string{"discovery:6379"}, ClusterOption: tc.opt}, m.connFn, newRetryer(defaultRetryDelayFn))
			if err != tc.want || client != nil {
				t.Fatalf("got (%v, %v), want (nil, %v)", client, err, tc.want)
			}
			if n := m.refreshCount(); n != 0 {
				t.Fatalf("the options were checked after %d refreshes, want before connecting", n)
			}
		})
	}
}

// With the zero values, a request for a slot without a node refreshes once and fails with ErrNoSlot, as before.
func TestClusterSlotWaitDefaultsFailFast(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	m := newSWCluster(swFailingOver)
	client := newSWClient(t, m, ClusterOption{})
	ctx := context.Background()
	b := client.B()
	for _, tc := range []struct {
		do   func() error
		name string
	}{
		{name: "Do", do: func() error { return client.Do(ctx, b.Get().Key(swMissingKey).Build()).Error() }},
		{name: "DoCache", do: func() error { return client.DoCache(ctx, b.Get().Key(swMissingKey).Cache(), time.Minute).Error() }},
		{name: "DoMulti", do: func() error {
			return client.DoMulti(ctx, b.Get().Key(swPresentKey).Build(), b.Get().Key(swMissingKey).Build())[1].Error()
		}},
		{name: "DoMultiCache", do: func() error {
			return client.DoMultiCache(ctx, CT(b.Get().Key(swPresentKey).Cache(), time.Minute), CT(b.Get().Key(swMissingKey).Cache(), time.Minute))[1].Error()
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 3; i++ {
				before := m.refreshCount()
				start := time.Now()
				if err := tc.do(); err != ErrNoSlot {
					t.Fatalf("got %v, want ErrNoSlot", err)
				}
				swCheckWithin(t, "failing", time.Since(start), 0, swSlack)
				if n := m.refreshCount() - before; n != 1 {
					t.Fatalf("got %d refreshes, want 1", n)
				}
			}
		})
	}
	if v, _, err := swGet(ctx, client, swPresentKey); err != nil || v != swAddr1 {
		t.Fatalf("got (%q, %v) for a slot with a node, want (%q, nil)", v, err, swAddr1)
	}
}

func TestClusterSlotWaitMinRefreshInterval(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	const interval = 100 * time.Millisecond

	// hammer sends GETs for a slot without a node from 200 goroutines for d, and counts the results.
	hammer := func(client *clusterClient, d time.Duration) (noSlot, other int64) {
		var wg sync.WaitGroup
		end := time.Now().Add(d)
		for i := 0; i < 200; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for time.Now().Before(end) {
					if _, _, err := swGet(context.Background(), client, swMissingKey); err == ErrNoSlot {
						atomic.AddInt64(&noSlot, 1)
					} else {
						atomic.AddInt64(&other, 1)
					}
				}
			}()
		}
		wg.Wait()
		return noSlot, other
	}

	t.Run("concurrent requests refresh at most once per interval and still fail fast", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MinRefreshInterval: interval})
		noSlot, other := hammer(client, 550*time.Millisecond)
		if noSlot == 0 || other != 0 {
			t.Fatalf("got %d ErrNoSlot and %d other results, want only ErrNoSlot", noSlot, other)
		}
		times := m.refreshTimes()
		swCheckSpacing(t, times, interval)
		if len(times) < 4 { // the initial refresh and at least 3 more in 550ms
			t.Fatalf("got %d refreshes, want at least 4", len(times))
		}
		t.Logf("%d requests, %d refreshes", noSlot, len(times))
	})

	t.Run("without it the same requests refresh far more often", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{})
		noSlot, _ := hammer(client, 300*time.Millisecond)
		if n := m.refreshCount(); n < 50 {
			t.Fatalf("got %d refreshes, want many more than one per %v", n, interval)
		}
		t.Logf("%d requests, %d refreshes", noSlot, m.refreshCount())
	})

	t.Run("concurrent waiting requests refresh at most once per interval", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MinRefreshInterval: interval, MaxSlotWait: 450 * time.Millisecond})
		var wg sync.WaitGroup
		for i := 0; i < 200; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, _, err := swGet(context.Background(), client, swMissingKey); err != ErrNoSlot {
					t.Errorf("got %v, want ErrNoSlot after MaxSlotWait", err)
				}
			}()
		}
		wg.Wait()
		times := m.refreshTimes()
		swCheckSpacing(t, times, interval)
		if len(times) < 4 {
			t.Fatalf("got %d refreshes, want at least 4", len(times))
		}
	})

	t.Run("refresh within the interval returns at once and does not join a pending lazy refresh", func(t *testing.T) {
		m := newSWCluster(swHealthy)
		client := newSWClient(t, m, ClusterOption{MinRefreshInterval: 300 * time.Millisecond})
		client.lazyRefresh() // pending for at least the rest of the interval
		start := time.Now()
		if err := client.refresh(context.Background()); err != nil {
			t.Fatalf("refresh: %v", err)
		}
		swCheckWithin(t, "refresh", time.Since(start), 0, 50*time.Millisecond)
		if n := m.refreshCount(); n != 1 {
			t.Fatalf("got %d refreshes, want only the initial one", n)
		}
		// lazyRefresh waits for the interval.
		for deadline := time.Now().Add(3 * time.Second); m.refreshCount() < 2 && time.Now().Before(deadline); {
			time.Sleep(5 * time.Millisecond)
		}
		times := m.refreshTimes()
		if len(times) != 2 {
			t.Fatalf("got %d refreshes, want the lazy refresh too", len(times))
		}
		swCheckSpacing(t, times, 300*time.Millisecond)
		for client.sc.suppressing() != 0 { // let the lazy refresh finish
			time.Sleep(time.Millisecond)
		}
	})

	t.Run("refresh after the interval refreshes", func(t *testing.T) {
		m := newSWCluster(swHealthy)
		client := newSWClient(t, m, ClusterOption{MinRefreshInterval: interval})
		time.Sleep(interval)
		if err := client.refresh(context.Background()); err != nil {
			t.Fatalf("refresh: %v", err)
		}
		if n := m.refreshCount(); n != 2 {
			t.Fatalf("got %d refreshes, want 2", n)
		}
	})
}

func TestClusterSlotWaitResumesWhenSlotAppears(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	const promoteAfter = 300 * time.Millisecond
	for _, tc := range []struct {
		name     string
		interval time.Duration
	}{
		{name: "every poll refreshes", interval: 0},
		{name: "refreshes limited by MinRefreshInterval", interval: 100 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newSWCluster(swFailingOver)
			client := newSWClient(t, m, ClusterOption{MinRefreshInterval: tc.interval, MaxSlotWait: 3 * time.Second})
			promote := time.AfterFunc(promoteAfter, func() { m.setTopology(swPromoted) })
			defer promote.Stop()
			v, took, err := swGet(context.Background(), client, swMissingKey)
			if err != nil || v != swAddr3 {
				t.Fatalf("got (%q, %v), want (%q, nil) from the promoted replica", v, err, swAddr3)
			}
			// Found by the first refresh after the promotion: within one poll, or one MinRefreshInterval.
			swCheckWithin(t, "the request", took, promoteAfter, promoteAfter+tc.interval+slotWaitPoll+swSlack)
			if tc.interval > 0 {
				swCheckSpacing(t, m.refreshTimes(), tc.interval)
			}
		})
	}
}

func TestClusterSlotWaitGivesUpAfterMaxSlotWait(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	const maxWait = 200 * time.Millisecond

	t.Run("ErrNoSlot", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MaxSlotWait: maxWait})
		_, took, err := swGet(context.Background(), client, swMissingKey)
		if err != ErrNoSlot {
			t.Fatalf("got %v, want ErrNoSlot", err)
		}
		swCheckWithin(t, "the request", took, maxWait, maxWait+swSlack)
	})

	t.Run("last refresh error", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MaxSlotWait: maxWait})
		refreshErr := errors.New("refresh failed")
		m.setShardsErr(refreshErr)
		_, took, err := swGet(context.Background(), client, swMissingKey)
		if err != refreshErr {
			t.Fatalf("got %v, want %v", err, refreshErr)
		}
		swCheckWithin(t, "the request", took, maxWait, maxWait+swSlack)
	})

	t.Run("last refresh error survives rate-limited polls", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MinRefreshInterval: time.Second, MaxSlotWait: maxWait})
		time.Sleep(time.Second) // the next refresh is allowed
		refreshErr := errors.New("refresh failed")
		m.setShardsErr(refreshErr)
		_, took, err := swGet(context.Background(), client, swMissingKey)
		if err != refreshErr {
			t.Fatalf("got %v, want %v", err, refreshErr)
		}
		swCheckWithin(t, "the request", took, maxWait, maxWait+swSlack)
		if n := m.refreshCount(); n != 2 {
			t.Fatalf("got %d refreshes, want 2", n)
		}
	})
}

func TestClusterSlotWaitHonorsContext(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	const timeout = 100 * time.Millisecond

	t.Run("deadline", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MinRefreshInterval: 50 * time.Millisecond, MaxSlotWait: 5 * time.Second})
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		_, took, err := swGet(ctx, client, swMissingKey)
		if err != context.DeadlineExceeded {
			t.Fatalf("got %v, want context.DeadlineExceeded", err)
		}
		swCheckWithin(t, "the request", took, timeout, timeout+swSlack)
	})

	t.Run("cancel", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MaxSlotWait: 5 * time.Second})
		ctx, cancel := context.WithCancel(context.Background())
		stop := time.AfterFunc(timeout, cancel)
		defer stop.Stop()
		_, took, err := swGet(ctx, client, swMissingKey)
		if err != context.Canceled {
			t.Fatalf("got %v, want context.Canceled", err)
		}
		swCheckWithin(t, "the request", took, timeout, timeout+swSlack)
	})

	t.Run("deadline while the refresh it started hangs", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MaxSlotWait: 5 * time.Second})
		release := m.holdShards()
		defer release()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		_, took, err := swGet(ctx, client, swMissingKey)
		if err != context.DeadlineExceeded {
			t.Fatalf("got %v, want context.DeadlineExceeded", err)
		}
		swCheckWithin(t, "the request", took, timeout, timeout+swSlack)
		release()
		for client.sc.suppressing() != 0 { // let the refresh finish
			time.Sleep(time.Millisecond)
		}
	})

	t.Run("MaxSlotWait while the refresh hangs", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MaxSlotWait: timeout})
		release := m.holdShards()
		defer release()
		_, took, err := swGet(context.Background(), client, swMissingKey)
		if err != ErrNoSlot {
			t.Fatalf("got %v, want ErrNoSlot", err)
		}
		swCheckWithin(t, "the request", took, timeout, timeout+swSlack)
		release()
		for client.sc.suppressing() != 0 {
			time.Sleep(time.Millisecond)
		}
	})

	t.Run("client closed", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MaxSlotWait: 5 * time.Second})
		closing := time.AfterFunc(timeout, client.Close)
		defer closing.Stop()
		_, took, err := swGet(context.Background(), client, swMissingKey)
		if err != ErrClosing {
			t.Fatalf("got %v, want ErrClosing", err)
		}
		swCheckWithin(t, "the request", took, timeout, timeout+swSlack)
	})

	t.Run("client closed while the refresh hangs", func(t *testing.T) {
		m := newSWCluster(swFailingOver)
		client := newSWClient(t, m, ClusterOption{MaxSlotWait: timeout})
		release := m.holdShards()
		defer release()
		closing := time.AfterFunc(timeout/2, client.Close)
		defer closing.Stop()
		_, took, err := swGet(context.Background(), client, swMissingKey)
		if err != ErrClosing {
			t.Fatalf("got %v, want ErrClosing", err)
		}
		swCheckWithin(t, "the request", took, timeout, timeout+swSlack)
		release()
		for client.sc.suppressing() != 0 {
			time.Sleep(time.Millisecond)
		}
	})
}

// pickMulti and pickMultiCache wait the same way as pick.
func TestClusterSlotWaitMultiCommandPaths(t *testing.T) {
	defer ShouldNotLeak(SetupLeakDetection())
	const promoteAfter = 200 * time.Millisecond
	for _, tc := range []struct {
		do   func(client *clusterClient) ([]string, error)
		name string
	}{
		{name: "DoMulti", do: func(client *clusterClient) ([]string, error) {
			b := client.B()
			return swStrings(client.DoMulti(context.Background(), b.Get().Key(swPresentKey).Build(), b.Get().Key(swMissingKey).Build()))
		}},
		{name: "DoMultiCache", do: func(client *clusterClient) ([]string, error) {
			b := client.B()
			return swStrings(client.DoMultiCache(context.Background(),
				CT(b.Get().Key(swPresentKey).Cache(), time.Minute), CT(b.Get().Key(swMissingKey).Cache(), time.Minute)))
		}},
		{name: "DoCache", do: func(client *clusterClient) ([]string, error) {
			b := client.B()
			v, err := client.DoCache(context.Background(), b.Get().Key(swMissingKey).Cache(), time.Minute).ToString()
			return []string{swAddr1, v}, err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newSWCluster(swFailingOver)
			client := newSWClient(t, m, ClusterOption{MinRefreshInterval: 50 * time.Millisecond, MaxSlotWait: 3 * time.Second})
			promote := time.AfterFunc(promoteAfter, func() { m.setTopology(swPromoted) })
			defer promote.Stop()
			start := time.Now()
			got, err := tc.do(client)
			if err != nil || strings.Join(got, ",") != swAddr1+","+swAddr3 {
				t.Fatalf("got (%v, %v), want ([%s %s], nil)", got, err, swAddr1, swAddr3)
			}
			swCheckWithin(t, tc.name, time.Since(start), promoteAfter, promoteAfter+50*time.Millisecond+slotWaitPoll+swSlack)
		})
	}
}

func swStrings(resps []ValkeyResult) ([]string, error) {
	out := make([]string, len(resps))
	for i, r := range resps {
		v, err := r.ToString()
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// --- failover benchmark (begin) ---

// TestClusterSlotWaitFailoverBench runs one client through a mocked failover of one shard, for each variant of
// the options, and prints one FAILOVER_BENCH line per variant. It is opt-in:
//
//	VALKEY_FAILOVER_BENCH=1 [VALKEY_FAILOVER_BENCH_VARIANTS=a,d] go test -run TestClusterSlotWaitFailoverBench -v .
//
// The client refreshes from one discovery endpoint (PreferInitAddressRefresh) of a version 8 cluster
// (CLUSTER SHARDS) with 3 shards. CLUSTER SHARDS and GET take about 1ms each. 1,000 goroutines send GETs
// for keys of shard A, each with a 5s timeout, back to back. At t=0 shard A loses its primary: CLUSTER SHARDS
// shows no online primary for it and the old primary's connection fails every command. A refresh is
// triggered with lazyRefresh, like the close hook of the old primary's connections does. At t=3s a replica
// is promoted at a new address. The goroutines stop sending at t=4s.
func TestClusterSlotWaitFailoverBench(t *testing.T) {
	if os.Getenv("VALKEY_FAILOVER_BENCH") != "1" {
		t.Skip("set VALKEY_FAILOVER_BENCH=1 to run the mock failover benchmark")
	}
	selected := map[string]bool{}
	for _, v := range strings.Split(os.Getenv("VALKEY_FAILOVER_BENCH_VARIANTS"), ",") {
		if v = strings.TrimSpace(v); v != "" {
			selected[v] = true
		}
	}
	for _, v := range []struct {
		name       string
		minRefresh time.Duration
		maxWait    time.Duration
	}{
		{name: "a", minRefresh: 0, maxWait: 0},
		{name: "b", minRefresh: 250 * time.Millisecond, maxWait: 0},
		{name: "c", minRefresh: 0, maxWait: 5 * time.Second},
		{name: "d", minRefresh: 250 * time.Millisecond, maxWait: 5 * time.Second},
	} {
		if len(selected) > 0 && !selected[v.name] {
			continue
		}
		var opt ClusterOption
		if !swApplyOptions(&opt, v.minRefresh, v.maxWait) {
			t.Logf("variant %s: the options do not exist on this branch", v.name)
			continue
		}
		r := swRunFailover(t, opt)
		line := fmt.Sprintf("FAILOVER_BENCH variant=%s min_refresh=%v max_slot_wait=%v cluster_shards=%d err_no_slot=%d other_err=%d ok=%d max_ok_latency_ms=%.1f first_refresh_ms=%.1f recovered_ms=%.1f dead_primary_sends=%d other_sample=%q",
			v.name, v.minRefresh, v.maxWait, r.shards, r.noSlot, r.other, r.ok, swMs(r.maxOK), swMs(r.firstRefresh), swMs(r.recovered), r.deadSends, r.otherSample)
		fmt.Println(line)
		t.Log(line)
	}
}

func swMs(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }

type swFailoverResult struct {
	otherSample  string
	shards       int64         // CLUSTER SHARDS sent from t=0 until the last request finished
	noSlot       int64         // requests that failed with ErrNoSlot
	other        int64         // requests that failed with another error
	ok           int64         // requests that succeeded
	deadSends    int64         // GETs sent to the failed primary after t=0
	maxOK        time.Duration // longest latency of a successful request
	firstRefresh time.Duration // when the first CLUSTER SHARDS after t=0 was sent
	recovered    time.Duration // first success after the promotion, relative to the promotion
}

type swWorker struct {
	otherSample string
	firstOK     time.Time
	noSlot      int64
	other       int64
	ok          int64
	maxOK       time.Duration
	_           [64]byte // keep workers on separate cache lines
}

func swRunFailover(t *testing.T, opt ClusterOption) (r swFailoverResult) {
	const (
		goroutines = 1000
		promoteAt  = 3 * time.Second
		stopAt     = 4 * time.Second
		reqTimeout = 5 * time.Second
		rtt        = time.Millisecond
		discovery  = "discovery:6379"
		oldPrimary = "10.0.1.1:6379"
		newPrimary = "10.0.1.2:6379" // the replica of shard A, promoted
	)
	var (
		phase        atomic.Int32 // 0: healthy, 1: shard A has no online primary, 2: replica promoted
		t0           atomic.Int64 // UnixNano of t=0
		shards       atomic.Int64
		firstRefresh atomic.Int64
		deadSends    atomic.Int64
	)
	firstRefresh.Store(-1)
	shardB := swShard{from: 5461, to: 10922, nodes: []swNode{swPrimary("10.0.2.1:6379"), swReplica("10.0.2.2:6379")}}
	shardC := swShard{from: 10923, to: 16383, nodes: []swNode{swPrimary("10.0.3.1:6379"), swReplica("10.0.3.2:6379")}}
	topologies := []ValkeyResult{
		swShardsReply(swShard{from: 0, to: 5460, nodes: []swNode{swPrimary(oldPrimary), swReplica(newPrimary)}}, shardB, shardC),
		swShardsReply(swShard{from: 0, to: 5460, nodes: []swNode{{addr: oldPrimary, role: "master", health: "fail"}, swReplica(newPrimary)}}, shardB, shardC),
		swShardsReply(swShard{from: 0, to: 5460, nodes: []swNode{swPrimary(newPrimary)}}, shardB, shardC),
	}
	connErr := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connect: connection refused")}
	connFn := func(dst string, _ *ClientOption) conn {
		return &mockConn{
			AddrFn:    func() string { return dst },
			VersionFn: func() int { return 8 },
			DoFn: func(cmd Completed) ValkeyResult {
				if swIsClusterShards(cmd) {
					if start := t0.Load(); start != 0 {
						shards.Add(1)
						firstRefresh.CompareAndSwap(-1, time.Now().UnixNano()-start)
					}
					time.Sleep(rtt)
					return topologies[phase.Load()]
				}
				if dst == oldPrimary && phase.Load() > 0 {
					deadSends.Add(1)
					return NewErrorResult(connErr)
				}
				time.Sleep(rtt)
				return NewResult(strmsg(typeBlobString, "v"), nil)
			},
		}
	}
	opt.PreferInitAddressRefresh = true
	client, err := newClusterClient(&ClientOption{InitAddress: []string{discovery}, ClusterOption: opt}, connFn, newRetryer(defaultRetryDelayFn))
	if err != nil {
		t.Fatalf("newClusterClient: %v", err)
	}
	defer client.Close()
	keys := swKeysInSlots(64, 0, 5460)
	time.Sleep(300 * time.Millisecond) // more than MinRefreshInterval since the initial refresh

	start := time.Now()
	t0.Store(start.UnixNano())
	phase.Store(1)
	client.lazyRefresh()
	promote := time.AfterFunc(promoteAt, func() { phase.Store(2) })
	defer promote.Stop()
	var stop atomic.Bool
	stopper := time.AfterFunc(stopAt, func() { stop.Store(true) })
	defer stopper.Stop()

	workers := make([]swWorker, goroutines)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(w *swWorker, key string) {
			defer wg.Done()
			for !stop.Load() {
				ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
				begin := time.Now()
				err := client.Do(ctx, client.B().Get().Key(key).Build()).Error()
				end := time.Now()
				cancel()
				switch {
				case err == nil:
					w.ok++
					w.maxOK = max(w.maxOK, end.Sub(begin))
					if w.firstOK.IsZero() {
						w.firstOK = end
					}
				case errors.Is(err, ErrNoSlot):
					w.noSlot++
				default:
					w.other++
					if w.otherSample == "" {
						w.otherSample = err.Error()
					}
				}
			}
		}(&workers[i], keys[i%len(keys)])
	}
	wg.Wait()

	r.shards = shards.Load()
	r.deadSends = deadSends.Load()
	r.firstRefresh = time.Duration(firstRefresh.Load())
	var firstOK time.Time
	for _, w := range workers {
		r.noSlot += w.noSlot
		r.other += w.other
		r.ok += w.ok
		r.maxOK = max(r.maxOK, w.maxOK)
		if r.otherSample == "" {
			r.otherSample = w.otherSample
		}
		if !w.firstOK.IsZero() && (firstOK.IsZero() || w.firstOK.Before(firstOK)) {
			firstOK = w.firstOK
		}
	}
	if !firstOK.IsZero() {
		r.recovered = firstOK.Sub(start.Add(promoteAt))
	}
	return r
}

// --- failover benchmark (end) ---
