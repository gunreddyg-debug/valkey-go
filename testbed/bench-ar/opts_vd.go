//go:build vd

package main

import (
	"flag"

	"github.com/valkey-io/valkey-go"
)

// The vd builds (-tags vd): valkey-go with Vardhaman's draft ClusterOption.MaxRefreshDelay (his fork PR #9,
// commit 4c18d37), the maintainer-suggested configurable delay (setup: src/unvd).

var maxRefreshDelay = flag.Duration("max-refresh-delay", 0, "ClusterOption.MaxRefreshDelay (vd builds only; 0: the library default, 1s to 30s by cluster size; negative: no delay)")

func applyVDOptions(opt *valkey.ClusterOption, m map[string]string) {
	opt.MaxRefreshDelay = *maxRefreshDelay
	m["MaxRefreshDelay"] = opt.MaxRefreshDelay.String()
}
