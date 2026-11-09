//go:build asm

package main

import (
	"flag"
	"strconv"

	"github.com/valkey-io/valkey-go"
)

// The asm builds (-tags asm): valkey-go with PR #209, which adds ClusterOption.SlotMigrationAwareness
// and its polling intervals. bin/bench-u209, -u208209, -p209 and -p208209 are such builds.

var (
	slotMigrationAwareness  = flag.Bool("slot-migration-awareness", false, "ClusterOption.SlotMigrationAwareness: the slot migration monitor of PR #209 (asm builds only)")
	slotMigrationIdle       = flag.Duration("slot-migration-idle", 0, "ClusterOption.SlotMigrationIdleInterval (asm builds only; 0: the library default, 30s)")
	slotMigrationActive     = flag.Duration("slot-migration-active", 0, "ClusterOption.SlotMigrationActiveInterval (asm builds only; 0: the library default, 250ms)")
	slotMigrationFinalizing = flag.Duration("slot-migration-finalizing", 0, "ClusterOption.SlotMigrationFinalizingInterval (asm builds only; 0: the library default, 20ms)")
)

func applyASMOptions(opt *valkey.ClusterOption, m map[string]string) {
	opt.SlotMigrationAwareness = *slotMigrationAwareness
	opt.SlotMigrationIdleInterval = *slotMigrationIdle
	opt.SlotMigrationActiveInterval = *slotMigrationActive
	opt.SlotMigrationFinalizingInterval = *slotMigrationFinalizing
	m["SlotMigrationAwareness"] = strconv.FormatBool(opt.SlotMigrationAwareness)
	for k, v := range map[string]string{
		"SlotMigrationIdleInterval":       opt.SlotMigrationIdleInterval.String(),
		"SlotMigrationActiveInterval":     opt.SlotMigrationActiveInterval.String(),
		"SlotMigrationFinalizingInterval": opt.SlotMigrationFinalizingInterval.String(),
	} {
		if v != "0s" {
			m[k] = v
		}
	}
}
