package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// runFailScript executes a fail script (fail-primary.sh or planned-failover.sh) with args and extra
// environment, echoes its output with run timestamps, and records when it reports the failure
// (FAILED ...), the end of it (RESTORED ...), and the other machine-readable lines (FLAGGED_FAIL,
// PROMOTE, PROMOTED, FAILED_OVER).
func runFailScript(script string, args, env []string, t0 time.Time, st *stats, keep func(string)) error {
	cmd := exec.Command(script, args...)
	cmd.Env = append(os.Environ(), env...)
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd.Stdout, cmd.Stderr = pw, pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return err
	}
	pw.Close()
	sc := bufio.NewScanner(pr)
	for sc.Scan() {
		at := time.Since(t0)
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "FAILED "):
			st.failAt.CompareAndSwap(0, int64(at))
		case strings.HasPrefix(line, "RESTORED "):
			st.restoredAt.CompareAndSwap(0, int64(at))
		case strings.HasPrefix(line, "FLAGGED_FAIL "):
			st.flaggedAt.CompareAndSwap(0, int64(at))
		case strings.HasPrefix(line, "PROMOTE "):
			st.promoteAt.CompareAndSwap(0, int64(at))
		case strings.HasPrefix(line, "PROMOTED "):
			st.promotedAt.CompareAndSwap(0, int64(at))
		case strings.HasPrefix(line, "FAILED_OVER "):
			st.failedOverAt.CompareAndSwap(0, int64(at))
		}
		out := fmt.Sprintf("[t=%6.3fs] %s", at.Seconds(), line)
		fmt.Println(scriptName(script)+":", out)
		keep(out)
	}
	pr.Close()
	return cmd.Wait()
}

func scriptName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}
