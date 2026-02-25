package output

import (
	"os/exec"
	"testing"
)

func TestBuildArgsYdotool(t *testing.T) {
	ty := NewTyper(Ydotool, 0)
	args := ty.buildArgs("hello world")

	want := []string{"type", "--", "hello world"}
	assertArgs(t, args, want)
}

func TestBuildArgsYdotoolWithDelay(t *testing.T) {
	ty := NewTyper(Ydotool, 500)
	args := ty.buildArgs("test")

	want := []string{"type", "--key-delay", "500", "--", "test"}
	assertArgs(t, args, want)
}

func TestBuildArgsXdotool(t *testing.T) {
	ty := NewTyper(Xdotool, 0)
	args := ty.buildArgs("hello")

	want := []string{"type", "--", "hello"}
	assertArgs(t, args, want)
}

func TestBuildArgsXdotoolWithDelay(t *testing.T) {
	ty := NewTyper(Xdotool, 5000) // 5000 us = 5 ms
	args := ty.buildArgs("test")

	want := []string{"type", "--delay", "5", "--", "test"}
	assertArgs(t, args, want)
}

func TestTypeEmpty(t *testing.T) {
	ty := NewTyper(Ydotool, 0)
	err := ty.Type("")
	if err != nil {
		t.Errorf("typing empty string should not error: %v", err)
	}
}

func TestTypeMockExec(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	ty := NewTyper(Ydotool, 0)
	ty.execCmd = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		// Return a command that succeeds
		return exec.Command("true")
	}

	err := ty.Type("hello world")
	if err != nil {
		t.Fatalf("type failed: %v", err)
	}

	if capturedName != "ydotool" {
		t.Errorf("expected command 'ydotool', got %q", capturedName)
	}

	wantArgs := []string{"type", "--", "hello world"}
	assertArgs(t, capturedArgs, wantArgs)
}

func assertArgs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args length mismatch: got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("arg[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}
