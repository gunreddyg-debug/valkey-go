//go:build !bd

package main

import "github.com/valkey-io/valkey-go"

// The builds without the bd tag: no -lazy-max-delay flag.

func applyBDOptions(*valkey.ClusterOption, map[string]string) {}
