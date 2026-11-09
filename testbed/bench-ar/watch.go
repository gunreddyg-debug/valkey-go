package main

import (
	"bufio"
	"context"
	"net"
	"sort"
	"sync"
	"time"
)

// watcher polls CLUSTER SHARDS on every node directly (one connection per node, every interval) and
// records whether the shard that owns the watched slot has a node with role master and health online.
// valkey-go drops a shard without one from its slot map, so this is the window in which a client
// that refreshes gets no node for those slots (ErrNoSlot).
type watcher struct {
	t0         time.Time
	nodes      []string
	slot       int64
	oldPrimary string
	every      time.Duration
	fast       map[string]bool // nodes polled every fastEvery instead: the ones that can take over the slot
	fastEvery  time.Duration

	mu  sync.Mutex
	obs []observation
}

type observation struct {
	At               time.Duration
	Node             string
	OK               bool // the node answered
	HasPrimary       bool
	Primary          string
	OldPrimaryFailed bool // the original primary is listed with health fail
}

func (w *watcher) run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, n := range w.nodes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.poll(ctx, n)
		}()
	}
	wg.Wait()
}

func (w *watcher) poll(ctx context.Context, node string) {
	var conn net.Conn
	var rd *bufio.Reader
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	every := w.every
	if w.fast[node] && w.fastEvery > 0 {
		every = w.fastEvery
	}
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		at := time.Since(w.t0)
		o := observation{At: at, Node: node}
		shards, err := func() ([]shardInfo, error) {
			if conn == nil {
				c, err := net.DialTimeout("tcp", node, 300*time.Millisecond)
				if err != nil {
					return nil, err
				}
				conn, rd = c, bufio.NewReader(c)
			}
			_ = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
			if _, err := conn.Write(shardsCmd); err != nil {
				return nil, err
			}
			v, err := readRESP(rd)
			if err != nil {
				return nil, err
			}
			if e, ok := v.(respError); ok {
				return nil, e
			}
			return parseShards(v)
		}()
		if err != nil {
			if conn != nil {
				conn.Close()
				conn = nil
			}
		} else {
			o.OK = true
			if s, ok := shardOf(shards, w.slot); ok {
				o.Primary, o.HasPrimary = s.onlinePrimary()
				for _, n := range s.Nodes {
					if n.Addr == w.oldPrimary && n.Health == "fail" {
						o.OldPrimaryFailed = true
					}
				}
			}
		}
		w.mu.Lock()
		w.obs = append(w.obs, o)
		w.mu.Unlock()
	}
}

// topology summarizes the observations after failAt. Durations are offsets from failAt; -1 means never.
type topology struct {
	FailFlagMs       float64 `json:"fail_flag_seen_ms"`          // first poll listing the old primary with health fail
	NoPrimaryStartMs float64 `json:"no_primary_start_ms"`        // first poll without an online primary for the slot
	NoPrimaryEndMs   float64 `json:"no_primary_end_ms"`          // first poll with one again, after the last poll without
	NoPrimaryMs      float64 `json:"no_primary_window_ms"`       // NoPrimaryEndMs - NoPrimaryStartMs
	NoPrimaryPolls   int     `json:"no_primary_polls"`           // polls (all nodes) without an online primary
	PromotedMs       float64 `json:"promoted_ms"`                // first poll showing a different online primary
	NewPrimary       string  `json:"new_primary"`                //
	Polls            int     `json:"polls"`                      // polls answered, all nodes
	NodesNoPrimary   int     `json:"nodes_reporting_no_primary"` // distinct nodes that reported no online primary
}

func (w *watcher) summarize(failAt time.Duration) topology {
	w.mu.Lock()
	obs := append([]observation(nil), w.obs...)
	w.mu.Unlock()
	sort.Slice(obs, func(i, j int) bool { return obs[i].At < obs[j].At })
	ms := func(d time.Duration) float64 { return float64((d - failAt).Microseconds()) / 1000 }
	t := topology{FailFlagMs: -1, NoPrimaryStartMs: -1, NoPrimaryEndMs: -1, NoPrimaryMs: -1, PromotedMs: -1}
	if failAt <= 0 {
		return t
	}
	var lastNo time.Duration = -1
	nodes := map[string]bool{}
	for _, o := range obs {
		if !o.OK || o.At < failAt {
			continue
		}
		t.Polls++
		if o.OldPrimaryFailed && t.FailFlagMs < 0 {
			t.FailFlagMs = ms(o.At)
		}
		if !o.HasPrimary {
			t.NoPrimaryPolls++
			nodes[o.Node] = true
			if t.NoPrimaryStartMs < 0 {
				t.NoPrimaryStartMs = ms(o.At)
			}
			lastNo = o.At
		}
		if o.HasPrimary && o.Primary != w.oldPrimary && t.PromotedMs < 0 {
			t.PromotedMs = ms(o.At)
			t.NewPrimary = o.Primary
		}
	}
	t.NodesNoPrimary = len(nodes)
	if lastNo >= 0 {
		for _, o := range obs {
			if o.OK && o.At > lastNo && o.HasPrimary {
				t.NoPrimaryEndMs = ms(o.At)
				t.NoPrimaryMs = t.NoPrimaryEndMs - t.NoPrimaryStartMs
				break
			}
		}
	}
	return t
}
