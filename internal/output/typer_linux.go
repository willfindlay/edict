//go:build linux

package output

import (
	"fmt"
	"os/exec"
	"strconv"
)

// Backend identifies the typing simulation tool on Linux.
type Backend string

const (
	Ydotool Backend = "ydotool"
	Xdotool Backend = "xdotool"
)

// linuxTyper simulates keyboard input via ydotool or xdotool.
type linuxTyper struct {
	backend Backend
	delayUs int
	execCmd func(name string, args ...string) *exec.Cmd
}

// NewTyper creates a typer with the specified backend and keystroke delay.
func NewTyper(backend Backend, delayUs int) Typer {
	return &linuxTyper{
		backend: backend,
		delayUs: delayUs,
		execCmd: exec.Command,
	}
}

func (t *linuxTyper) CheckAvailable() error {
	_, err := exec.LookPath(string(t.backend))
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", t.backend, err)
	}
	return nil
}

func (t *linuxTyper) Type(text string) error {
	if text == "" {
		return nil
	}

	args := t.buildArgs(text)
	cmd := t.execCmd(string(t.backend), args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w: %s", t.backend, err, string(output))
	}
	return nil
}

func (t *linuxTyper) buildArgs(text string) []string {
	switch t.backend {
	case Ydotool:
		args := []string{"type"}
		if t.delayUs > 0 {
			args = append(args, "--key-delay", strconv.Itoa(t.delayUs))
		}
		args = append(args, "--", text)
		return args
	case Xdotool:
		args := []string{"type"}
		if t.delayUs > 0 {
			// xdotool uses milliseconds
			delayMs := t.delayUs / 1000
			if delayMs < 1 {
				delayMs = 1
			}
			args = append(args, "--delay", strconv.Itoa(delayMs))
		}
		args = append(args, "--", text)
		return args
	default:
		return []string{"type", "--", text}
	}
}
