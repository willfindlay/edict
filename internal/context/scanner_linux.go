//go:build linux

package context

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ScanClaudeProcesses finds running Claude Code processes by checking /proc.
// Primary detection: CLAUDECODE=1 in environ. Fallback: "claude" in comm.
func ScanClaudeProcesses() []ClaudeProcess {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var procs []ClaudeProcess
	seen := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		if !isClaudeProcess(pid) {
			continue
		}

		cwd := readCWD(pid)
		if cwd == "" || seen[cwd] {
			continue
		}
		seen[cwd] = true

		procs = append(procs, ClaudeProcess{
			PID:          pid,
			CWD:          cwd,
			CanonicalCWD: cwd,
		})
	}

	return procs
}

func isClaudeProcess(pid int) bool {
	environPath := filepath.Join("/proc", strconv.Itoa(pid), "environ")
	data, err := os.ReadFile(environPath)
	if err == nil {
		for _, env := range bytes.Split(data, []byte{0}) {
			if string(env) == "CLAUDECODE=1" {
				return true
			}
		}
	}

	commPath := filepath.Join("/proc", strconv.Itoa(pid), "comm")
	comm, err := os.ReadFile(commPath)
	if err == nil {
		name := strings.TrimSpace(string(comm))
		if strings.Contains(name, "claude") {
			return true
		}
	}

	return false
}

func readCWD(pid int) string {
	cwdPath := filepath.Join("/proc", strconv.Itoa(pid), "cwd")
	target, err := os.Readlink(cwdPath)
	if err != nil {
		return ""
	}
	return target
}
