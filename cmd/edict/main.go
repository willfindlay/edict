package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/willfindlay/edict/internal/config"
)

func main() {
	configPath := flag.String("config", "", "path to config.toml")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	_ = cfg
	fmt.Println("edict: ready")
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		// Try default locations
		candidates := []string{
			"config.toml",
			os.ExpandEnv("$HOME/.config/edict/config.toml"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return config.Load(c)
			}
		}
		// No config file found, use defaults
		return config.Default(), nil
	}
	return config.Load(path)
}
