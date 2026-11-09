package main

import (
	"fmt"
	"strconv"
	"strings"
)

// crc16 is CRC-16/XMODEM, the hash of Valkey cluster key slots.
func crc16(b []byte) uint16 {
	var crc uint16
	for _, c := range b {
		crc ^= uint16(c) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// keySlot returns the cluster slot of key, honoring {hash tags}.
func keySlot(key string) int64 {
	if s := strings.IndexByte(key, '{'); s >= 0 {
		if e := strings.IndexByte(key[s+1:], '}'); e > 0 {
			key = key[s+1 : s+1+e]
		}
	}
	return int64(crc16([]byte(key)) % 16384)
}

// keysInShard returns n keys whose slots the shard owns.
func keysInShard(n int, s shardInfo) ([]string, error) {
	keys := make([]string, 0, n)
	for i := 0; len(keys) < n; i++ {
		if i > 100*n+100000 {
			return nil, fmt.Errorf("found only %d keys in %v", len(keys), s.Ranges)
		}
		k := "lf:" + strconv.Itoa(i)
		if s.owns(keySlot(k)) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// keysInSlots returns n keys whose slots are in [lo, hi].
func keysInSlots(n int, lo, hi int64) ([]string, error) {
	keys := make([]string, 0, n)
	limit := 100*n*int(16384/(hi-lo+1)+1) + 100000
	for i := 0; len(keys) < n; i++ {
		if i > limit {
			return nil, fmt.Errorf("found only %d keys in slots %d-%d", len(keys), lo, hi)
		}
		k := "lf:" + strconv.Itoa(i)
		if s := keySlot(k); s >= lo && s <= hi {
			keys = append(keys, k)
		}
	}
	return keys, nil
}
