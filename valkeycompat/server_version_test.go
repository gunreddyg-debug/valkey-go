package valkeycompat

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The suite runs against Valkey and Redis servers of different versions: in
// CI the RESP3 adapter talks to Redis Stack 7.4 and the RESP2 adapter to
// Redis 5, while local runs often use one Valkey version everywhere. Specs for
// newer commands therefore check the server version instead of the resp3 flag.

// serverAtLeast reports whether the server behind adapter is Valkey valkeyMin
// or later, or Redis redisMin or later. An empty minimum means that engine
// does not support the feature at all.
func serverAtLeast(adapter Cmdable, valkeyMin, redisMin string) bool {
	GinkgoHelper()
	engine, version := serverVersion(adapter)
	min := redisMin
	if engine == "valkey" {
		min = valkeyMin
	}
	return min != "" && versionAtLeast(version, min)
}

// skipBeforeServerVersion skips the running spec unless serverAtLeast holds.
// what names the command or option in the skip message.
func skipBeforeServerVersion(adapter Cmdable, valkeyMin, redisMin, what string) {
	GinkgoHelper()
	if serverAtLeast(adapter, valkeyMin, redisMin) {
		return
	}
	engine, version := serverVersion(adapter)
	need := []string{}
	if valkeyMin != "" {
		need = append(need, "Valkey "+valkeyMin)
	}
	if redisMin != "" {
		need = append(need, "Redis "+redisMin)
	}
	Skip(fmt.Sprintf("%s needs %s; the server is %s %s", what, strings.Join(need, " or "), engine, version))
}

func serverVersion(adapter Cmdable) (engine, version string) {
	GinkgoHelper()
	info, err := adapter.Info(ctx, "server").Result()
	Expect(err).NotTo(HaveOccurred())
	return parseServerVersion(info)
}

// parseServerVersion reads the engine and version from INFO server. Valkey
// reports valkey_version next to a fixed redis_version kept for compatibility.
func parseServerVersion(info string) (engine, version string) {
	engine = "redis"
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "valkey_version:"); ok {
			return "valkey", v
		}
		if v, ok := strings.CutPrefix(line, "redis_version:"); ok {
			version = v
		}
		if line == "server_name:valkey" {
			engine = "valkey"
		}
	}
	return engine, version
}

// versionAtLeast compares dotted versions numerically, so 9.10 is newer than 9.9.
func versionAtLeast(version, min string) bool {
	v, m := strings.Split(version, "."), strings.Split(min, ".")
	for i := range m {
		var a int
		if i < len(v) {
			a, _ = strconv.Atoi(v[i])
		}
		b, _ := strconv.Atoi(m[i])
		if a != b {
			return a > b
		}
	}
	return true
}

func TestParseServerVersion(t *testing.T) {
	for _, c := range []struct {
		info, engine, version string
	}{
		{"# Server\r\nredis_version:7.2.4\r\nserver_name:valkey\r\nvalkey_version:9.1.3\r\n", "valkey", "9.1.3"},
		{"# Server\r\nredis_version:7.2.5\r\nserver_name:valkey\r\n", "valkey", "7.2.5"},
		{"# Server\r\nredis_version:7.4.0\r\nredis_git_sha1:00000000\r\n", "redis", "7.4.0"},
	} {
		engine, version := parseServerVersion(c.info)
		if engine != c.engine || version != c.version {
			t.Errorf("parseServerVersion(%q) = %s %s, want %s %s", c.info, engine, version, c.engine, c.version)
		}
	}
}

func TestVersionAtLeast(t *testing.T) {
	for _, c := range []struct {
		version, min string
		want         bool
	}{
		{"9.1.3", "9.1", true},
		{"9.1.0", "9.1", true},
		{"8.0.12", "9.1", false},
		{"9.10.0", "9.9", true},
		{"7.4.0", "8.0", false},
		{"5.0.14", "6.2", false},
		{"255.255.255", "9.1", true},
	} {
		if got := versionAtLeast(c.version, c.min); got != c.want {
			t.Errorf("versionAtLeast(%q, %q) = %v, want %v", c.version, c.min, got, c.want)
		}
	}
}
