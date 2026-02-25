//go:build linux

package main

import (
	"os"

	"github.com/willfindlay/edict/internal/config"
	edictctx "github.com/willfindlay/edict/internal/context"
)

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
