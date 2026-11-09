//go:build pr1 && !nsw

package main

import (
	"flag"

	"github.com/valkey-io/valkey-go"
)

// The pr1 builds (-tags pr1): valkey-go with the core change of PR #20, which adds
// ClusterOption.MinRefreshInterval and ClusterOption.MaxSlotWait. bin/bench-pr1 (go.pr1.mod) and
// bin/bench-p, -p208, -p209, -p208209 (build-208-209.sh) are such builds.

// buildName is the name of the valkey-go tree; build-208-209.sh sets it with -ldflags -X.
var buildName = "pr1"

var (
	minRefresh  = flag.Duration("min-refresh", 0, "ClusterOption.MinRefreshInterval (pr1 builds only)")
	maxSlotWait = flag.Duration("max-slot-wait", 0, "ClusterOption.MaxSlotWait (pr1 builds only)")
)

func applyPR1Options(opt *valkey.ClusterOption, m map[string]string) {
	opt.MinRefreshInterval = *minRefresh
	opt.MaxSlotWait = *maxSlotWait
	m["MinRefreshInterval"] = opt.MinRefreshInterval.String()
	m["MaxSlotWait"] = opt.MaxSlotWait.String()
}
