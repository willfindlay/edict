//go:build windows

package context

import (
	"log"
	"os/exec"
	"strings"
)

// ScanClaudeProcesses finds running Claude Code processes by shelling into WSL.
// It scans /proc inside the given WSL distro for processes with CLAUDECODE=1,
// then translates the returned Linux CWD paths to UNC paths for file access.
func ScanClaudeProcessesWSL(distro string) []ClaudeProcess {
	if distro == "" {
		return nil
	}

	// Shell into WSL to find Claude Code processes and their CWDs.
	// For each /proc/<pid>/environ containing CLAUDECODE=1, output the CWD.
	script := `for p in /proc/[0-9]*/environ; do
  grep -qz CLAUDECODE=1 "$p" 2>/dev/null &&
  readlink -f "$(dirname "$p")/cwd" 2>/dev/null
done`

	cmd := exec.Command("wsl.exe", "-d", distro, "--", "sh", "-c", script)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("WSL scan failed: %v", err)
		return nil
	}

	var procs []ClaudeProcess
	seen := make(map[string]bool)

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		linuxCWD := strings.TrimSpace(line)
		if linuxCWD == "" || seen[linuxCWD] {
			continue
		}
		seen[linuxCWD] = true

		procs = append(procs, ClaudeProcess{
			PID:          0, // PID not meaningful across WSL boundary
			CWD:          WSLToUNC(distro, linuxCWD),
			CanonicalCWD: linuxCWD,
		})
	}

	return procs
}

// ScanClaudeProcesses is the default scanner entry point on Windows.
// It requires a WSL distro name; callers should use ScanClaudeProcessesWSL directly.
func ScanClaudeProcesses() []ClaudeProcess {
	// Without a distro name, we can't scan. Callers on Windows should use
	// ScanClaudeProcessesWSL with the configured distro.
	return nil
}
