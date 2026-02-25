//go:build windows

package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/willfindlay/edict/internal/config"
	edictctx "github.com/willfindlay/edict/internal/context"
)

func signalNotify(ch chan<- os.Signal) {
	signal.Notify(ch, os.Interrupt)
}

func configCandidates() []string {
	candidates := []string{"config.toml"}
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		candidates = append(candidates, filepath.Join(appdata, "edict", "config.toml"))
	}
	return candidates
}

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
