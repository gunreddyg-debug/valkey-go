package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// Minimal RESP2 client used by the topology watcher and the start-up checks. It talks to the nodes
// directly, never through the client under test or the discovery endpoint, so it does not affect
// what is measured.

type respError string

func (e respError) Error() string { return string(e) }

var shardsCmd = []byte("*2\r\n$7\r\nCLUSTER\r\n$6\r\nSHARDS\r\n")

// getSlotMigrationsCmd is CLUSTER GETSLOTMIGRATIONS, polled by the slot migration monitor of PR #209.
var getSlotMigrationsCmd = []byte("*2\r\n$7\r\nCLUSTER\r\n$17\r\nGETSLOTMIGRATIONS\r\n")

func encodeCommand(args ...string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	return []byte(b.String())
}

// readRESP reads one RESP2 value: string (simple or bulk), int64, []any, nil, or respError.
func readRESP(r *bufio.Reader) (any, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 3 || line[len(line)-2] != '\r' {
		return nil, fmt.Errorf("malformed RESP line %q", line)
	}
	body := line[1 : len(line)-2]
	switch line[0] {
	case '+':
		return body, nil
	case '-':
		return respError(body), nil
	case ':':
		return strconv.ParseInt(body, 10, 64)
	case '$':
		n, err := strconv.Atoi(body)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, nil
		}
		buf := make([]byte, n+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return string(buf[:n]), nil
	case '*':
		n, err := strconv.Atoi(body)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, nil
		}
		arr := make([]any, n)
		for i := range arr {
			if arr[i], err = readRESP(r); err != nil {
				return nil, err
			}
		}
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected RESP type %q", line[0])
}

// roundTrip sends one command on a fresh connection and returns the reply.
func roundTrip(addr string, timeout time.Duration, cmd []byte) (any, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write(cmd); err != nil {
		return nil, err
	}
	v, err := readRESP(bufio.NewReader(conn))
	if err != nil {
		return nil, err
	}
	if e, ok := v.(respError); ok {
		return nil, e
	}
	return v, nil
}

// flatMap turns a RESP2 flat key/value array into a map.
func flatMap(v any) map[string]any {
	arr, _ := v.([]any)
	m := make(map[string]any, len(arr)/2)
	for i := 0; i+1 < len(arr); i += 2 {
		if k, ok := arr[i].(string); ok {
			m[k] = arr[i+1]
		}
	}
	return m
}

type shardNode struct {
	ID     string
	Addr   string
	Role   string
	Health string
}

type shardInfo struct {
	Ranges [][2]int64
	Nodes  []shardNode
}

func (s shardInfo) owns(slot int64) bool {
	for _, r := range s.Ranges {
		if slot >= r[0] && slot <= r[1] {
			return true
		}
	}
	return false
}

// primaryNode returns the node of the shard with role master and health online.
func (s shardInfo) primaryNode() (shardNode, bool) {
	for _, n := range s.Nodes {
		if n.Role == "master" && n.Health == "online" {
			return n, true
		}
	}
	return shardNode{}, false
}

// onlinePrimary returns the address of a node of the shard with role master and health online.
func (s shardInfo) onlinePrimary() (string, bool) {
	for _, n := range s.Nodes {
		if n.Role == "master" && n.Health == "online" {
			return n.Addr, true
		}
	}
	return "", false
}

func parseShards(v any) ([]shardInfo, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, errors.New("CLUSTER SHARDS reply is not an array")
	}
	shards := make([]shardInfo, 0, len(arr))
	for _, sv := range arr {
		m := flatMap(sv)
		var s shardInfo
		slots, _ := m["slots"].([]any)
		for i := 0; i+1 < len(slots); i += 2 {
			from, _ := slots[i].(int64)
			to, _ := slots[i+1].(int64)
			s.Ranges = append(s.Ranges, [2]int64{from, to})
		}
		nodes, _ := m["nodes"].([]any)
		for _, nv := range nodes {
			nm := flatMap(nv)
			host, _ := nm["endpoint"].(string)
			if host == "" || host == "?" {
				host, _ = nm["ip"].(string)
			}
			port, _ := nm["port"].(int64)
			role, _ := nm["role"].(string)
			health, _ := nm["health"].(string)
			id, _ := nm["id"].(string)
			s.Nodes = append(s.Nodes, shardNode{ID: id, Addr: net.JoinHostPort(host, strconv.FormatInt(port, 10)), Role: role, Health: health})
		}
		shards = append(shards, s)
	}
	return shards, nil
}

// shardOf returns the shard that owns slot.
func shardOf(shards []shardInfo, slot int64) (shardInfo, bool) {
	for _, s := range shards {
		if s.owns(slot) {
			return s, true
		}
	}
	return shardInfo{}, false
}

func fetchShards(addr string, timeout time.Duration) ([]shardInfo, error) {
	v, err := roundTrip(addr, timeout, shardsCmd)
	if err != nil {
		return nil, err
	}
	return parseShards(v)
}
