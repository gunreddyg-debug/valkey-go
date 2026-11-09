//go:build !vd

package main

import "github.com/valkey-io/valkey-go"

// The builds without the vd tag: no -max-refresh-delay flag.

func applyVDOptions(*valkey.ClusterOption, map[string]string) {}
