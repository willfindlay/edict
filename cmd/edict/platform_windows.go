//go:build windows

package main

import (
	"log"

	"github.com/willfindlay/edict/internal/config"
	edictctx "github.com/willfindlay/edict/internal/context"
)

func resolveHomeDir(cfg config.Config) string {
	unc, err := edictctx.WSLHomeUNC(cfg.Context.WSLDistro, cfg.Context.WSLHome)
	if err != nil {
		log.Printf("WSL home resolution failed: %v", err)
		return ""
	}
	return unc
}

func scanProcesses(cfg config.Config) []edictctx.ClaudeProcess {
	return edictctx.ScanClaudeProcessesWSL(cfg.Context.WSLDistro)
}
