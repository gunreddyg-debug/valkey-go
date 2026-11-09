package main

import (
	"bufio"
	"context"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// errStats follows the error replies the nodes send, from INFO errorstats (errorstat_MOVED, _ASK,
// _CLUSTERDOWN, ...), read on every node directly (not through the client or the discovery endpoint):
// once before the run (the baseline), every interval during it, and once after it. The counts are
// summed as increments between reads, so a node that is restarted during the run (MODE=kill: its
// counters start again from 0; detected by a new run_id) is counted correctly; a node that does not
// answer (paused, killed) is read again when it is back.
type errStats struct {
	t0    time.Time
	nodes []string
	every time.Duration

	mu       sync.Mutex
	last     map[string]errRead                  // node -> last successful read
	inc      map[string]map[string]int64         // node -> error type -> increments since the baseline
	buckets  map[string][]int64                  // error type -> increments per 100 ms (by read time, from t0)
	lastInc  map[string]map[string]time.Duration // node -> error type -> time of the last read that saw an increment
	restarts map[string]int
	readErrs map[string]int
	missing  map[string]string // node -> error of the baseline read, if it failed
}

type errRead struct {
	runID  string
	counts map[string]int64
}

func newErrStats(nodes []string, every time.Duration) *errStats {
	return &errStats{nodes: nodes, every: every, last: map[string]errRead{}, inc: map[string]map[string]int64{},
		buckets: map[string][]int64{}, lastInc: map[string]map[string]time.Duration{}, restarts: map[string]int{},
		readErrs: map[string]int{}, missing: map[string]string{}}
}

var infoErrorstatsCmd = encodeCommand("INFO", "server", "errorstats")

// parseErrorstats returns the run_id and the errorstat_<TYPE>:count=N values of an INFO reply.
func parseErrorstats(info string) errRead {
	r := errRead{counts: map[string]int64{}}
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "run_id:"); ok {
			r.runID = v
			continue
		}
		rest, ok := strings.CutPrefix(line, "errorstat_")
		if !ok {
			continue
		}
		typ, fields, ok := strings.Cut(rest, ":")
		if !ok {
			continue
		}
		for _, f := range strings.Split(fields, ",") {
			if v, ok := strings.CutPrefix(f, "count="); ok {
				if n, err := strconv.ParseInt(v, 10, 64); err == nil {
					r.counts[typ] = n
				}
			}
		}
	}
	return r
}

func readErrorstats(conn net.Conn, rd *bufio.Reader, timeout time.Duration) (errRead, error) {
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write(infoErrorstatsCmd); err != nil {
		return errRead{}, err
	}
	v, err := readRESP(rd)
	if err != nil {
		return errRead{}, err
	}
	if e, ok := v.(respError); ok {
		return errRead{}, e
	}
	s, _ := v.(string)
	return parseErrorstats(s), nil
}

// observe records a read of node at time at (from t0); baseline marks the read before the run.
func (e *errStats) observe(node string, r errRead, at time.Duration, baseline bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	prev, ok := e.last[node]
	e.last[node] = r
	if baseline {
		return
	}
	if !ok {
		// No baseline for this node (it did not answer before the run): all it reports counts. That
		// includes errors from before the run; summary lists the node under no_baseline.
		prev = errRead{counts: map[string]int64{}}
	}
	restarted := prev.runID != "" && r.runID != "" && prev.runID != r.runID
	if restarted {
		e.restarts[node]++
	}
	for typ, n := range r.counts {
		d := n - prev.counts[typ]
		if restarted || d < 0 {
			d = n // the counters started again from 0
		}
		if d <= 0 {
			continue
		}
		if e.inc[node] == nil {
			e.inc[node] = map[string]int64{}
			e.lastInc[node] = map[string]time.Duration{}
		}
		e.inc[node][typ] += d
		e.lastInc[node][typ] = at
		b := e.buckets[typ]
		i := max(0, int(at/bucketWidth))
		for len(b) <= i {
			b = append(b, 0)
		}
		b[i] += d
		e.buckets[typ] = b
	}
}

func (e *errStats) readOnce(node string, timeout time.Duration) (errRead, error) {
	conn, err := net.DialTimeout("tcp", node, timeout)
	if err != nil {
		return errRead{}, err
	}
	defer conn.Close()
	return readErrorstats(conn, bufio.NewReader(conn), timeout)
}

// baseline reads every node before the run.
func (e *errStats) baseline() {
	var wg sync.WaitGroup
	for _, n := range e.nodes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := e.readOnce(n, time.Second)
			if err != nil {
				e.mu.Lock()
				e.missing[n] = err.Error()
				e.mu.Unlock()
				return
			}
			e.observe(n, r, 0, true)
		}()
	}
	wg.Wait()
}

// final reads every node after the run, retrying a node that does not answer for up to wait.
func (e *errStats) final(wait time.Duration) {
	var wg sync.WaitGroup
	deadline := time.Now().Add(wait)
	for _, n := range e.nodes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				r, err := e.readOnce(n, time.Second)
				if err == nil {
					e.observe(n, r, time.Since(e.t0), false)
					return
				}
				e.mu.Lock()
				e.readErrs[n]++
				e.mu.Unlock()
				if time.Now().After(deadline) {
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
}

// run polls every node every interval until ctx is done.
func (e *errStats) run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, n := range e.nodes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.poll(ctx, n)
		}()
	}
	wg.Wait()
}

func (e *errStats) poll(ctx context.Context, node string) {
	var conn net.Conn
	var rd *bufio.Reader
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	ticker := time.NewTicker(e.every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if conn == nil {
			c, err := net.DialTimeout("tcp", node, 300*time.Millisecond)
			if err != nil {
				e.mu.Lock()
				e.readErrs[node]++
				e.mu.Unlock()
				continue
			}
			conn, rd = c, bufio.NewReader(c)
		}
		r, err := readErrorstats(conn, rd, 500*time.Millisecond)
		if err != nil {
			conn.Close()
			conn = nil
			e.mu.Lock()
			e.readErrs[node]++
			e.mu.Unlock()
			continue
		}
		e.observe(node, r, time.Since(e.t0), false)
	}
}

// serverErrors is the summary in the result.
type serverErrors struct {
	Totals       map[string]int64              `json:"totals"`    // error type -> replies, all nodes, from before to after the run
	ByNode       map[string]map[string]int64   `json:"by_node"`   // node -> error type -> replies
	LastAtS      map[string]map[string]float64 `json:"last_at_s"` // node -> error type -> last read that saw new ones (from t0)
	Restarts     map[string]int                `json:"restarts,omitempty"`
	ReadErrors   map[string]int                `json:"read_errors,omitempty"`
	NoBaseline   map[string]string             `json:"no_baseline,omitempty"`
	PerSecond    map[string][]int64            `json:"per_second"` // error type -> replies per second (by read time)
	PollInterval float64                       `json:"poll_interval_ms"`
}

func (e *errStats) summary(seconds int) serverErrors {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := serverErrors{Totals: map[string]int64{}, ByNode: map[string]map[string]int64{}, LastAtS: map[string]map[string]float64{},
		Restarts: e.restarts, ReadErrors: e.readErrs, NoBaseline: e.missing, PerSecond: map[string][]int64{},
		PollInterval: float64(e.every.Microseconds()) / 1000}
	for node, m := range e.inc {
		s.ByNode[node] = map[string]int64{}
		s.LastAtS[node] = map[string]float64{}
		for typ, n := range m {
			s.Totals[typ] += n
			s.ByNode[node][typ] = n
			s.LastAtS[node][typ] = e.lastInc[node][typ].Seconds()
		}
	}
	perSec := int(time.Second / bucketWidth)
	types := make([]string, 0, len(e.buckets))
	for typ := range e.buckets {
		types = append(types, typ)
	}
	sort.Strings(types)
	for _, typ := range types {
		rows := make([]int64, seconds)
		for i, v := range e.buckets[typ] {
			r := min(i/perSec, seconds-1)
			rows[r] += v
		}
		s.PerSecond[typ] = rows
	}
	return s
}

// lastIncrease returns when node last reported new errors of type typ (from t0), or -1.
func (e *errStats) lastIncrease(node, typ string) time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	if m := e.lastInc[node]; m != nil {
		if at, ok := m[typ]; ok {
			return at
		}
	}
	return -1
}
