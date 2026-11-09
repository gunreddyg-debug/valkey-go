//go:build pr1 && nsw

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/valkey-io/valkey-go"
)

// The pr1,nsw builds (-tags pr1,nsw): PR #209 after its second review, which keeps
// ClusterOption.MinRefreshInterval but no longer has MaxSlotWait (bin/bench-r2, bin/bench-r2big).

// buildName is the name of the valkey-go tree; build-ar.sh sets it with -ldflags -X.
var buildName = "r2"

var (
	minRefresh  = flag.Duration("min-refresh", 0, "ClusterOption.MinRefreshInterval (pr1 builds only)")
	maxSlotWait = flag.Duration("max-slot-wait", 0, "not available in nsw builds: must stay 0")
)

func applyPR1Options(opt *valkey.ClusterOption, m map[string]string) {
	if *maxSlotWait != 0 {
		fmt.Fprintln(os.Stderr, "bench: -max-slot-wait is not available in this build (no MaxSlotWait)")
		os.Exit(2)
	}
	opt.MinRefreshInterval = *minRefresh
	m["MinRefreshInterval"] = opt.MinRefreshInterval.String()
}
