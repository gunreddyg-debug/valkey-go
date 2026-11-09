package main

import (
	"math/bits"
	"sync/atomic"
	"time"
)

// Latency histogram of successful requests: log-linear bins over microseconds, exact below 64 µs and
// 32 bins per power of two above (at most 1/32 ≈ 3% relative error), up to 2^25 µs ≈ 33 s; longer
// requests go to the last bin. stats keeps one histogram per 100 ms bucket (by request end time), so
// percentiles can be computed for any window after the run.

const (
	histLinear = 64 // exact bins: 0 .. 63 µs
	histSubLog = 5  // 2^5 = 32 bins per power of two
	histSub    = 1 << histSubLog
	histMaxExp = 24 // last power of two: [2^24, 2^25) µs
	histBins   = histLinear + (histMaxExp-5)*histSub
)

type hist [histBins]atomic.Int64

func histIndex(d time.Duration) int {
	us := d.Microseconds()
	if us < histLinear {
		if us < 0 {
			return 0
		}
		return int(us)
	}
	e := bits.Len64(uint64(us)) - 1 // >= 6
	if e > histMaxExp {
		return histBins - 1
	}
	sub := int(us>>(e-histSubLog)) & (histSub - 1)
	return histLinear + (e-6)*histSub + sub
}

// histLow and histHigh return the bounds of bin i in µs: [low, high).
func histLow(i int) float64 {
	if i < histLinear {
		return float64(i)
	}
	e := (i-histLinear)/histSub + 6
	sub := (i - histLinear) % histSub
	return float64(int64(histSub+sub) << (e - histSubLog))
}

func histHigh(i int) float64 {
	if i < histLinear {
		return float64(i + 1)
	}
	e := (i-histLinear)/histSub + 6
	return histLow(i) + float64(int64(1)<<(e-histSubLog))
}

// latencySummary is what the result reports for a set of successful requests. Values are in
// milliseconds; a percentile is the middle of the bin that holds it.
type latencySummary struct {
	Count int64   `json:"count"`
	P50   float64 `json:"p50_ms"`
	P99   float64 `json:"p99_ms"`
	P999  float64 `json:"p999_ms"`
	Max   float64 `json:"max_ms"` // upper bound of the highest non-empty bin
}

// summarizeHist merges the histograms of buckets [from, to) and returns the percentiles.
func summarizeHist(hs []hist, from, to int) latencySummary {
	from = max(from, 0)
	to = min(to, len(hs))
	var merged [histBins]int64
	var n int64
	for b := from; b < to; b++ {
		for i := range hs[b] {
			v := hs[b][i].Load()
			merged[i] += v
			n += v
		}
	}
	s := latencySummary{Count: n}
	if n == 0 {
		return s
	}
	pct := func(q float64) float64 {
		rank := int64(q*float64(n-1)) + 1 // 1-based rank of the q-quantile
		var seen int64
		for i, v := range merged {
			seen += v
			if seen >= rank {
				return (histLow(i) + histHigh(i)) / 2 / 1000
			}
		}
		return histHigh(histBins-1) / 1000
	}
	s.P50, s.P99, s.P999 = pct(0.50), pct(0.99), pct(0.999)
	for i := histBins - 1; i >= 0; i-- {
		if merged[i] > 0 {
			s.Max = histHigh(i) / 1000
			break
		}
	}
	return s
}
