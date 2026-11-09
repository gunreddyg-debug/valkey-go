package main

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valkey-io/valkey-go"
)

const bucketWidth = 100 * time.Millisecond

type bucket struct {
	ok, noSlot, deadline, other atomic.Int64
}

// stats counts the results of the workload's requests. Times are offsets from t0, the start of the run.
type stats struct {
	t0      time.Time
	buckets []bucket // by request end time
	lat     []hist   // latency of successful requests, per bucket (by request end time)

	ok, noSlot, deadline, other atomic.Int64
	otherMsgs                   sync.Map // error message -> *msgCount

	failAt     atomic.Int64 // when the fail script reported the failure; 0 until then
	restoredAt atomic.Int64
	// Other lines of the fail scripts; 0 until seen.
	flaggedAt    atomic.Int64 // FLAGGED_FAIL: the replica sees the primary flagged fail (HOLD_PROMOTION only)
	promoteAt    atomic.Int64 // PROMOTE: CLUSTER FAILOVER [FORCE] answered by the replica
	promotedAt   atomic.Int64 // PROMOTED: the replica owns the slot (planned-failover.sh)
	failedOverAt atomic.Int64 // FAILED_OVER: planned failover done, cluster healthy (planned-failover.sh)

	longestMu      sync.Mutex
	longestOK      atomic.Int64 // ns
	longestOKStart atomic.Int64

	firstOKAfterFail atomic.Int64 // end of the first success of a request started after failAt
	firstErr         atomic.Int64 // end of the first failed request
	lastErr          atomic.Int64 // end of the last failed request
	firstNoSlot      atomic.Int64 // end of the first request that failed with ErrNoSlot
	lastNoSlot       atomic.Int64

	// Requests started at or after resumeAt (-idle-until), the end of the idle window; resumeAt 0: no window.
	resumeAt                                                    int64
	resumeOK, resumeNoSlot, resumeDeadline, resumeOther         atomic.Int64
	firstOKAfterResume, firstErrAfterResume, lastErrAfterResume atomic.Int64 // request end times; -1: none
}

type msgCount struct {
	n     atomic.Int64
	first time.Duration // end of the first request that failed with the message
}

func newStats(t0 time.Time, length time.Duration) *stats {
	n := int(length/bucketWidth) + 1
	s := &stats{t0: t0, buckets: make([]bucket, n), lat: make([]hist, n)}
	s.firstOKAfterFail.Store(-1)
	s.firstOKAfterResume.Store(-1)
	s.firstErrAfterResume.Store(-1)
	s.lastErrAfterResume.Store(-1)
	s.firstErr.Store(-1)
	s.lastErr.Store(-1)
	s.firstNoSlot.Store(-1)
	s.lastNoSlot.Store(-1)
	return s
}

func (s *stats) bucketIndex(at time.Duration) int {
	i := int(at / bucketWidth)
	if i < 0 {
		i = 0
	}
	if i >= len(s.buckets) {
		i = len(s.buckets) - 1
	}
	return i
}

func (s *stats) bucket(at time.Duration) *bucket {
	return &s.buckets[s.bucketIndex(at)]
}

func storeMin(a *atomic.Int64, v int64) {
	for {
		cur := a.Load()
		if cur >= 0 && cur <= v {
			return
		}
		if a.CompareAndSwap(cur, v) {
			return
		}
	}
}

func storeMax(a *atomic.Int64, v int64) {
	for {
		cur := a.Load()
		if cur >= v {
			return
		}
		if a.CompareAndSwap(cur, v) {
			return
		}
	}
}

func (s *stats) record(start, end time.Time, err error) {
	ts, te := start.Sub(s.t0), end.Sub(s.t0)
	if s.resumeAt > 0 && int64(ts) >= s.resumeAt {
		s.recordResume(te, err)
	}
	bi := s.bucketIndex(te)
	b := &s.buckets[bi]
	switch {
	case err == nil || valkey.IsValkeyNil(err):
		s.ok.Add(1)
		b.ok.Add(1)
		s.lat[bi][histIndex(te-ts)].Add(1)
		if lat := int64(te - ts); lat > s.longestOK.Load() {
			s.longestMu.Lock()
			if lat > s.longestOK.Load() {
				s.longestOK.Store(lat)
				s.longestOKStart.Store(int64(ts))
			}
			s.longestMu.Unlock()
		}
		if f := s.failAt.Load(); f > 0 && int64(ts) > f {
			storeMin(&s.firstOKAfterFail, int64(te))
		}
		return
	case errors.Is(err, valkey.ErrNoSlot):
		s.noSlot.Add(1)
		b.noSlot.Add(1)
		storeMin(&s.firstNoSlot, int64(te))
		storeMax(&s.lastNoSlot, int64(te))
	case errors.Is(err, context.DeadlineExceeded):
		s.deadline.Add(1)
		b.deadline.Add(1)
	default:
		s.other.Add(1)
		b.other.Add(1)
		msg := err.Error()
		if len(msg) > 120 {
			msg = msg[:120]
		}
		c, ok := s.otherMsgs.Load(msg)
		if !ok {
			c, _ = s.otherMsgs.LoadOrStore(msg, &msgCount{first: te})
		}
		c.(*msgCount).n.Add(1)
	}
	storeMin(&s.firstErr, int64(te))
	storeMax(&s.lastErr, int64(te))
}

type countEntry struct {
	Msg     string  `json:"msg"`
	Count   int64   `json:"count"`
	FirstAt float64 `json:"first_at_s"` // from t0
}

func (s *stats) otherErrors() []countEntry {
	var out []countEntry
	s.otherMsgs.Range(func(k, v any) bool {
		c := v.(*msgCount)
		out = append(out, countEntry{Msg: k.(string), Count: c.n.Load(), FirstAt: c.first.Seconds()})
		return true
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}

// recordResume counts a request started after the idle window.
func (s *stats) recordResume(te time.Duration, err error) {
	switch {
	case err == nil || valkey.IsValkeyNil(err):
		s.resumeOK.Add(1)
		if s.firstOKAfterResume.Load() < 0 {
			s.firstOKAfterResume.CompareAndSwap(-1, int64(te))
		}
		return
	case errors.Is(err, valkey.ErrNoSlot):
		s.resumeNoSlot.Add(1)
	case errors.Is(err, context.DeadlineExceeded):
		s.resumeDeadline.Add(1)
	default:
		s.resumeOther.Add(1)
	}
	if s.firstErrAfterResume.Load() < 0 {
		s.firstErrAfterResume.CompareAndSwap(-1, int64(te))
	}
	storeMax(&s.lastErrAfterResume, int64(te))
}
