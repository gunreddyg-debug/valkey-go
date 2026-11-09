//go:build !pr1

package main

import "github.com/valkey-io/valkey-go"

// The builds without the pr1 tag: valkey-go without PR #20, which has no MinRefreshInterval or
// MaxSlotWait, so these builds have no -min-refresh or -max-slot-wait flags.

// buildName is the name of the valkey-go tree; build-208-209.sh sets it with -ldflags -X.
var buildName = "upstream"

func applyPR1Options(*valkey.ClusterOption, map[string]string) {}
