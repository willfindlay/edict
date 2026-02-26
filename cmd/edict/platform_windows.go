//go:build windows

package main

import (
	"os"
	"os/signal"
	"path/filepath"
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
