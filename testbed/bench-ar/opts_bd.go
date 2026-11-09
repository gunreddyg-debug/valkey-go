//go:build bd

package main

import (
	"flag"

	"github.com/valkey-io/valkey-go"
)

// The bd builds (-tags bd): valkey-go with the prototype ClusterOption.LazyRefreshMaxDelay of the adaptive
// retry comparison (setup-src-ar.sh: pnbd, pnbdr).

var lazyMaxDelay = flag.Duration("lazy-max-delay", 0, "ClusterOption.LazyRefreshMaxDelay (bd builds only; 0: the library default, 1s to 30s by cluster size)")

func applyBDOptions(opt *valkey.ClusterOption, m map[string]string) {
	opt.LazyRefreshMaxDelay = *lazyMaxDelay
	m["LazyRefreshMaxDelay"] = opt.LazyRefreshMaxDelay.String()
}
