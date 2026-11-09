package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// dialCounter is the client's ClientOption.DialCtxFn. It counts dials per destination address, and
// wraps every connection to count the CLUSTER SHARDS and CLUSTER GETSLOTMIGRATIONS commands the client
// writes to it, and the bytes it writes per 10 ms. The client writes commands as RESP arrays of bulk
// strings, so each command is always the same byte sequence (shardsCmd, getSlotMigrationsCmd). It is
// matched across Write calls, in case one is split.
type dialCounter struct {
	mu       sync.Mutex
	dials    map[string]int64
	dialErrs map[string]int64
	shards   []shardsEvent
	gsm      []shardsEvent // CLUSTER GETSLOTMIGRATIONS (the slot migration monitor of PR #209)
	addrs    map[string]*addrStats

	t0 atomic.Int64 // unix ns of the start of the run; 0 before: bytes are only bucketed from then on
}

// addrStats are the bytes the client wrote to one destination, in total and per bytesBucketWidth from t0.
type addrStats struct {
	total   atomic.Int64
	buckets [bytesBuckets]atomic.Int64
}

const (
	bytesBucketWidth = 10 * time.Millisecond
	bytesBuckets     = 15000 // 150 s
)

type shardsEvent struct {
	at   time.Time
	addr string
}

func newDialCounter() *dialCounter {
	return &dialCounter{dials: map[string]int64{}, dialErrs: map[string]int64{}, addrs: map[string]*addrStats{}}
}

func (d *dialCounter) start(t0 time.Time) { d.t0.Store(t0.UnixNano()) }

func (d *dialCounter) dial(ctx context.Context, dst string, dialer *net.Dialer, _ *tls.Config) (net.Conn, error) {
	conn, err := dialer.DialContext(ctx, "tcp", dst)
	d.mu.Lock()
	d.dials[dst]++
	if err != nil {
		d.dialErrs[dst]++
	}
	as := d.addrs[dst]
	if as == nil {
		as = &addrStats{}
		d.addrs[dst] = as
	}
	d.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return &countingConn{Conn: conn, dst: dst, d: d, as: as}, nil
}

func (d *dialCounter) add(list *[]shardsEvent, dst string, n int) {
	now := time.Now()
	d.mu.Lock()
	for i := 0; i < n; i++ {
		*list = append(*list, shardsEvent{at: now, addr: dst})
	}
	d.mu.Unlock()
}

type dialSnapshot struct {
	Dials    map[string]int64
	DialErrs map[string]int64
	Shards   []shardsEvent
	GSM      []shardsEvent
	Bytes    map[string][]int64 // per destination, per bytesBucketWidth from t0
}

func (d *dialCounter) snapshot() dialSnapshot {
	d.mu.Lock()
	defer d.mu.Unlock()
	s := dialSnapshot{Dials: map[string]int64{}, DialErrs: map[string]int64{}, Bytes: map[string][]int64{},
		Shards: append([]shardsEvent(nil), d.shards...), GSM: append([]shardsEvent(nil), d.gsm...)}
	for k, v := range d.dials {
		s.Dials[k] = v
	}
	for k, v := range d.dialErrs {
		s.DialErrs[k] = v
	}
	for k, as := range d.addrs {
		b := make([]int64, bytesBuckets)
		for i := range b {
			b[i] = as.buckets[i].Load()
		}
		s.Bytes[k] = b
	}
	sort.Slice(s.Shards, func(i, j int) bool { return s.Shards[i].at.Before(s.Shards[j].at) })
	sort.Slice(s.GSM, func(i, j int) bool { return s.GSM[i].at.Before(s.GSM[j].at) })
	return s
}

type countingConn struct {
	net.Conn
	d   *dialCounter
	as  *addrStats
	dst string

	mu     sync.Mutex
	shards patternCounter
	gsm    patternCounter
}

func (c *countingConn) Write(b []byte) (int, error) {
	ns, ng := c.countAll(b)
	if c.d != nil {
		if ns > 0 {
			c.d.add(&c.d.shards, c.dst, ns)
		}
		if ng > 0 {
			c.d.add(&c.d.gsm, c.dst, ng)
		}
		if c.as != nil {
			c.as.total.Add(int64(len(b)))
			if t0 := c.d.t0.Load(); t0 != 0 {
				if i := int((time.Now().UnixNano() - t0) / int64(bytesBucketWidth)); i >= 0 && i < bytesBuckets {
					c.as.buckets[i].Add(int64(len(b)))
				}
			}
		}
	}
	return c.Conn.Write(b)
}

// count returns how many CLUSTER SHARDS commands end in b, including one that started in an earlier write.
func (c *countingConn) count(b []byte) int {
	n, _ := c.countAll(b)
	return n
}

// countAll returns how many CLUSTER SHARDS and CLUSTER GETSLOTMIGRATIONS commands end in b, including
// ones that started in an earlier write.
func (c *countingConn) countAll(b []byte) (shards, gsm int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shards.pat == nil {
		c.shards.pat = shardsCmd
	}
	if c.gsm.pat == nil {
		c.gsm.pat = getSlotMigrationsCmd
	}
	return c.shards.feed(b), c.gsm.feed(b)
}

// patternCounter counts the occurrences of pat in a stream of writes, including those split across writes.
type patternCounter struct {
	pat  []byte
	tail []byte // the last len(pat)-1 bytes written
}

func (p *patternCounter) feed(b []byte) int {
	n := bytes.Count(b, p.pat)
	keep := len(p.pat) - 1
	if len(p.tail) > 0 {
		head := b
		if len(head) > keep {
			head = head[:keep]
		}
		joined := append(append(make([]byte, 0, len(p.tail)+len(head)), p.tail...), head...)
		for i := 0; i < len(p.tail); i++ { // only matches that start in the tail: the others are in b
			if bytes.HasPrefix(joined[i:], p.pat) {
				n++
			}
		}
	}
	if len(b) >= keep {
		p.tail = append(p.tail[:0], b[len(b)-keep:]...)
	} else {
		p.tail = append(p.tail, b...)
		if len(p.tail) > keep {
			p.tail = p.tail[len(p.tail)-keep:]
		}
	}
	return n
}
