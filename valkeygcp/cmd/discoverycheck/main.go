// Command discoverycheck verifies a valkeygcp cluster client against a live cluster:
// a local test bed, or Memorystore for Valkey from a VM in the instance's VPC.
//
// It builds the client option with valkeygcp.ResolveClientOption, so the
// WithDiscoveryEndpoint defaults apply. It then reports:
//
//   - the effective options (InitAddress, refresh pinning, MinRefreshInterval, MaxSlotWait, TLS, auth);
//   - a SET/GET smoke test over -keys keys;
//   - with -duration, a per-second workload report: ok, ErrNoSlot, context deadline and other
//     errors, topology refreshes, and the slowest successful request;
//   - with -new-client-every, how many fresh clients failed to bootstrap during the run;
//   - every CLUSTER SHARDS/SLOTS the client sent, per destination address. PASS means all of
//     them went to the discovery endpoint (InitAddress);
//   - whether the client's node set matches the online nodes in CLUSTER SHARDS.
//
// Examples:
//
//	go run ./cmd/discoverycheck -addr 127.0.0.1:16379                     # local Docker test bed
//	go run ./cmd/discoverycheck -addr 10.10.0.5:6379 -duration 30m        # Memorystore, no auth/TLS
//	go run ./cmd/discoverycheck -instance projects/P/locations/R/instances/I -duration 2h  # IAM + TLS
//	go run ./cmd/discoverycheck -addr ... -min-refresh 0 -max-slot-wait 0 # both settings off (A/B)
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeygcp"
)

var (
	addr        = flag.String("addr", "", "discovery endpoint IP:port (no control-plane call); set this or -instance")
	instance    = flag.String("instance", "", "Memorystore instance projects/P/locations/R/instances/I (endpoint, auth and TLS from the API)")
	network     = flag.String("network", "", "VPC network for -instance when it serves several VPCs: projects/P/global/networks/N")
	numKeys     = flag.Int("keys", 10000, "keys for the smoke test and the workload")
	duration    = flag.Duration("duration", 0, "run a workload for this long after the smoke test (0: smoke test only)")
	goroutines  = flag.Int("goroutines", 50, "concurrent workload loops")
	reqTimeout  = flag.Duration("timeout", 2*time.Second, "context timeout per request")
	minRefresh  = flag.Duration("min-refresh", -1, "override ClusterOption.MinRefreshInterval (0 = off; default: keep the valkeygcp value)")
	maxSlotWait = flag.Duration("max-slot-wait", -1, "override ClusterOption.MaxSlotWait (0 = off; default: keep the valkeygcp value)")
	newEvery    = flag.Duration("new-client-every", 0, "during the workload, bootstrap a fresh client this often (0: off)")
	jsonOut     = flag.String("json", "", "also write the summary as JSON to this file")
)

// Refresh commands as valkey-go writes them on the wire (RESP arrays).
var (
	shardsCmd = []byte("*2\r\n$7\r\nCLUSTER\r\n$6\r\nSHARDS\r\n")
	slotsCmd  = []byte("*2\r\n$7\r\nCLUSTER\r\n$5\r\nSLOTS\r\n")
)

// tap wraps every connection the client dials and counts refresh commands per destination.
type tap struct {
	mu       sync.Mutex
	dials    map[string]int
	dialErrs map[string]int
	refresh  map[string]int
	total    atomic.Int64
}

func newTap() *tap {
	return &tap{dials: map[string]int{}, dialErrs: map[string]int{}, refresh: map[string]int{}}
}

// dial is a valkey.ClientOption.DialCtxFn. valkey-go leaves TLS to the hook when one is set.
func (t *tap) dial(ctx context.Context, dst string, d *net.Dialer, cfg *tls.Config) (net.Conn, error) {
	var c net.Conn
	var err error
	if cfg != nil {
		c, err = (&tls.Dialer{NetDialer: d, Config: cfg}).DialContext(ctx, "tcp", dst)
	} else {
		c, err = d.DialContext(ctx, "tcp", dst)
	}
	t.mu.Lock()
	t.dials[dst]++
	if err != nil {
		t.dialErrs[dst]++
	}
	t.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return &countingConn{Conn: c, dst: dst, t: t}, nil
}

func (t *tap) snapshot() (dials, dialErrs, refresh map[string]int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := func(m map[string]int) map[string]int {
		out := make(map[string]int, len(m))
		for k, v := range m {
			out[k] = v
		}
		return out
	}
	return cp(t.dials), cp(t.dialErrs), cp(t.refresh)
}

type countingConn struct {
	net.Conn
	dst string
	t   *tap
}

func (c *countingConn) Write(b []byte) (int, error) {
	if n := bytes.Count(b, shardsCmd) + bytes.Count(b, slotsCmd); n > 0 {
		c.t.mu.Lock()
		c.t.refresh[c.dst] += n
		c.t.mu.Unlock()
		c.t.total.Add(int64(n))
	}
	return c.Conn.Write(b)
}

type counters struct {
	ok, noSlot, deadline, other atomic.Int64
	maxLatNs                    atomic.Int64
}

func (c *counters) observe(err error, lat time.Duration) string {
	switch {
	case err == nil || valkey.IsValkeyNil(err):
		c.ok.Add(1)
		for {
			cur := c.maxLatNs.Load()
			if int64(lat) <= cur || c.maxLatNs.CompareAndSwap(cur, int64(lat)) {
				break
			}
		}
		return ""
	case errors.Is(err, valkey.ErrNoSlot):
		c.noSlot.Add(1)
	case errors.Is(err, context.DeadlineExceeded):
		c.deadline.Add(1)
	default:
		c.other.Add(1)
		return err.Error()
	}
	return ""
}

type summary struct {
	Options          map[string]string `json:"options"`
	SmokeOK          int               `json:"smoke_ok"`
	SmokeFailed      int               `json:"smoke_failed"`
	Requests         map[string]int64  `json:"requests"`
	MaxLatencyMs     float64           `json:"max_success_latency_ms"`
	Bootstraps       int64             `json:"bootstraps"`
	BootstrapFails   int64             `json:"bootstrap_failures"`
	RefreshByAddress map[string]int    `json:"refresh_by_address"`
	DialsByAddress   map[string]int    `json:"dials_by_address"`
	PinningPass      bool              `json:"pinning_pass"`
	TopologyMatch    *bool             `json:"topology_match,omitempty"`
	OtherErrors      map[string]int    `json:"other_errors,omitempty"`
}

func main() {
	flag.Parse()
	if (*addr == "") == (*instance == "") {
		log.Fatal("set exactly one of -addr or -instance")
	}

	cfg := valkeygcp.Config{Instance: *instance, Network: *network, DiscoveryEndpoint: *addr}
	if *instance == "" {
		cfg.Topology = valkeygcp.TopologyCluster
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	opt, err := valkeygcp.ResolveClientOption(ctx, cfg)
	cancel()
	if err != nil {
		log.Fatalf("ResolveClientOption: %v", err)
	}
	if *minRefresh >= 0 {
		opt.ClusterOption.MinRefreshInterval = *minRefresh
	}
	if *maxSlotWait >= 0 {
		opt.ClusterOption.MaxSlotWait = *maxSlotWait
	}
	sum := summary{Options: describe(opt), Requests: map[string]int64{}, OtherErrors: map[string]int{}}
	fmt.Println("== effective options")
	keys := make([]string, 0, len(sum.Options))
	for k := range sum.Options {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("   %-26s %s\n", k, sum.Options[k])
	}

	base := opt // fresh bootstrap clients use the same option, without the tap
	t := newTap()
	opt.DialCtxFn = t.dial
	start := time.Now()
	client, err := valkey.NewClient(opt)
	if err != nil {
		log.Fatalf("valkey.NewClient: %v", err)
	}
	defer client.Close()
	fmt.Printf("== client ready in %v; %d nodes\n", time.Since(start).Round(time.Millisecond), len(client.Nodes()))

	// Smoke test.
	keyNames := make([]string, *numKeys)
	for i := range keyNames {
		keyNames[i] = "discoverycheck:" + strconv.Itoa(i)
	}
	for _, k := range keyNames {
		c, cancel := context.WithTimeout(context.Background(), *reqTimeout)
		err := client.Do(c, client.B().Set().Key(k).Value(k).Build()).Error()
		if err == nil {
			err = client.Do(c, client.B().Get().Key(k).Build()).Error()
		}
		cancel()
		if err != nil {
			sum.SmokeFailed++
			if sum.SmokeFailed <= 5 {
				fmt.Printf("   smoke error on %s: %v\n", k, err)
			}
		} else {
			sum.SmokeOK++
		}
	}
	fmt.Printf("== smoke: %d SET+GET ok, %d failed\n", sum.SmokeOK, sum.SmokeFailed)

	// Workload. Ctrl-C (or SIGTERM) ends it early; the report below still runs.
	var cnt counters
	var bootOK, bootFail atomic.Int64
	if *duration > 0 {
		fmt.Printf("== workload: %d goroutines for %v (Ctrl-C ends it early and prints the report)\n", *goroutines, *duration)
		stop, stopCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stopCancel()
		deadline := time.Now().Add(*duration)
		running := func() bool { return stop.Err() == nil && time.Now().Before(deadline) }
		var errMu sync.Mutex
		var wg sync.WaitGroup
		for g := 0; g < *goroutines; g++ {
			wg.Add(1)
			go func(seed int64) {
				defer wg.Done()
				r := rand.New(rand.NewSource(seed))
				for running() {
					k := keyNames[r.Intn(len(keyNames))]
					c, cancel := context.WithTimeout(context.Background(), *reqTimeout)
					t0 := time.Now()
					var err error
					if r.Intn(5) == 0 {
						err = client.Do(c, client.B().Set().Key(k).Value(k).Build()).Error()
					} else {
						err = client.Do(c, client.B().Get().Key(k).Build()).Error()
					}
					cancel()
					if msg := cnt.observe(err, time.Since(t0)); msg != "" {
						errMu.Lock()
						sum.OtherErrors[msg]++
						errMu.Unlock()
					}
				}
			}(int64(g) + time.Now().UnixNano())
		}
		if *newEvery > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for running() {
					select {
					case <-time.After(*newEvery):
					case <-stop.Done():
						return
					}
					c, err := valkey.NewClient(base)
					if err == nil {
						cctx, cancel := context.WithTimeout(context.Background(), *reqTimeout)
						err = c.Do(cctx, c.B().Set().Key("discoverycheck:bootstrap").Value("1").Build()).Error()
						cancel()
						c.Close()
					}
					if err != nil {
						bootFail.Add(1)
						fmt.Printf("   %s bootstrap FAILED: %v\n", time.Now().Format("15:04:05"), err)
					} else {
						bootOK.Add(1)
					}
				}
			}()
		}
		var lastOK, lastNoSlot, lastDeadline, lastOther, lastRefresh int64
		tick := time.NewTicker(time.Second)
		for running() {
			var now time.Time
			select {
			case now = <-tick.C:
			case <-stop.Done():
				now = time.Now()
				fmt.Println("   interrupted: ending the workload")
			}
			ok, ns, dl, ot, rf := cnt.ok.Load(), cnt.noSlot.Load(), cnt.deadline.Load(), cnt.other.Load(), t.total.Load()
			fmt.Printf("   %s ok %7d  ErrNoSlot %6d  deadline %5d  other %5d  refreshes %3d  slowest-ok %v\n",
				now.Format("15:04:05"), ok-lastOK, ns-lastNoSlot, dl-lastDeadline, ot-lastOther, rf-lastRefresh,
				time.Duration(cnt.maxLatNs.Load()).Round(time.Millisecond))
			lastOK, lastNoSlot, lastDeadline, lastOther, lastRefresh = ok, ns, dl, ot, rf
		}
		tick.Stop()
		wg.Wait()
		stopCancel() // a second Ctrl-C now kills the process as usual
	}
	sum.Requests = map[string]int64{"ok": cnt.ok.Load(), "ErrNoSlot": cnt.noSlot.Load(), "deadline": cnt.deadline.Load(), "other": cnt.other.Load()}
	sum.MaxLatencyMs = float64(cnt.maxLatNs.Load()) / 1e6
	sum.Bootstraps, sum.BootstrapFails = bootOK.Load()+bootFail.Load(), bootFail.Load()

	// Refresh pinning: snapshot before our own CLUSTER SHARDS below.
	dials, dialErrs, refresh := t.snapshot()
	sum.RefreshByAddress, sum.DialsByAddress = refresh, dials
	pinned := map[string]bool{}
	for _, a := range opt.InitAddress {
		pinned[a] = true
	}
	elsewhere := 0
	for a, n := range refresh {
		if !pinned[a] {
			elsewhere += n
		}
	}
	sum.PinningPass = elsewhere == 0 && len(refresh) > 0
	fmt.Printf("== requests: %v; slowest successful request %.1fms\n", sum.Requests, sum.MaxLatencyMs)
	if *newEvery > 0 {
		fmt.Printf("== bootstrap: %d fresh clients, %d failed\n", sum.Bootstraps, sum.BootstrapFails)
	}
	fmt.Printf("== dials by address: %v (failed: %v)\n", dials, dialErrs)
	fmt.Printf("== CLUSTER SHARDS/SLOTS by address: %v\n", refresh)
	if sum.PinningPass {
		fmt.Printf("== PINNING PASS: all refreshes went to the discovery endpoint %v\n", opt.InitAddress)
	} else {
		fmt.Printf("== PINNING FAIL: %d refreshes went to addresses other than %v\n", elsewhere, opt.InitAddress)
	}
	for msg, n := range sum.OtherErrors {
		fmt.Printf("   other error x%d: %s\n", n, msg)
	}

	// Topology: the client's nodes vs the online nodes in CLUSTER SHARDS.
	if match, err := checkTopology(client, opt.TLSConfig != nil); err != nil {
		fmt.Printf("== topology check skipped: %v\n", err)
	} else {
		sum.TopologyMatch = &match
	}

	if *jsonOut != "" {
		b, _ := json.MarshalIndent(sum, "", "  ")
		if err := os.WriteFile(*jsonOut, b, 0o644); err != nil {
			log.Printf("write %s: %v", *jsonOut, err)
		}
	}
	if !sum.PinningPass || sum.SmokeFailed > 0 || sum.BootstrapFails > 0 {
		os.Exit(1)
	}
}

func describe(o valkey.ClientOption) map[string]string {
	auth := "none"
	switch {
	case o.AuthCredentialsFn != nil:
		auth = "AuthCredentialsFn (IAM or token)"
	case o.Password != "":
		auth = "password"
	}
	return map[string]string{
		"InitAddress":              strings.Join(o.InitAddress, ","),
		"PreferInitAddressRefresh": strconv.FormatBool(o.ClusterOption.PreferInitAddressRefresh),
		"PreferClusterShards":      strconv.FormatBool(o.ClusterOption.PreferClusterShards),
		"ShardsRefreshInterval":    o.ClusterOption.ShardsRefreshInterval.String(),
		"MinRefreshInterval":       o.ClusterOption.MinRefreshInterval.String(),
		"MaxSlotWait":              o.ClusterOption.MaxSlotWait.String(),
		"TLS":                      strconv.FormatBool(o.TLSConfig != nil),
		"Auth":                     auth,
	}
}

func checkTopology(client valkey.Client, useTLS bool) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shards, err := client.Do(ctx, client.B().ClusterShards().Build()).ToArray()
	if err != nil {
		return false, err
	}
	online := map[string]bool{}
	field := func(m map[string]valkey.ValkeyMessage, k string) *valkey.ValkeyMessage {
		v := m[k]
		return &v
	}
	for _, s := range shards {
		m, err := s.AsMap()
		if err != nil {
			return false, err
		}
		nodes, _ := field(m, "nodes").ToArray()
		for _, n := range nodes {
			nm, err := n.AsMap()
			if err != nil {
				continue
			}
			if h, _ := field(nm, "health").ToString(); h != "online" {
				continue
			}
			host, _ := field(nm, "endpoint").ToString()
			if host == "" || host == "?" {
				host, _ = field(nm, "ip").ToString()
			}
			port, _ := field(nm, "port").AsInt64()
			if tp, err := field(nm, "tls-port").AsInt64(); useTLS && err == nil && tp > 0 {
				port = tp
			}
			online[net.JoinHostPort(host, strconv.FormatInt(port, 10))] = true
		}
	}
	clientNodes := client.Nodes()
	var missing, extra []string
	for a := range online {
		if _, ok := clientNodes[a]; !ok {
			missing = append(missing, a)
		}
	}
	for a := range clientNodes {
		if !online[a] {
			extra = append(extra, a)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	fmt.Printf("== topology: %d online nodes in CLUSTER SHARDS, client knows %d", len(online), len(clientNodes))
	if len(missing)+len(extra) == 0 {
		fmt.Println(" -> MATCH")
		return true, nil
	}
	fmt.Printf(" -> MISMATCH (missing in client: %v; client-only: %v)\n", missing, extra)
	return false, nil
}
