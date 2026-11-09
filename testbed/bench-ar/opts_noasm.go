//go:build !asm

package main

import "github.com/valkey-io/valkey-go"

// The builds without the asm tag: valkey-go without PR #209, so no -slot-migration-awareness flag.

func applyASMOptions(*valkey.ClusterOption, map[string]string) {}
