//go:build linux

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/willfindlay/edict/internal/config"
	edictctx "github.com/willfindlay/edict/internal/context"
)

func signalNotify(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

func configCandidates() []string {
	return []string{
		"config.toml",
		os.ExpandEnv("$HOME/.config/edict/config.toml"),
	}
}

func resolveHomeDir(_ config.Config) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func scanProcesses(_ config.Config) []edictctx.ClaudeProcess {
	return edictctx.ScanClaudeProcesses()
}
