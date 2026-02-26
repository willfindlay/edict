package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Whisper.Host != "localhost" {
		t.Errorf("expected whisper host localhost, got %s", cfg.Whisper.Host)
	}
	if cfg.Whisper.Port != 9988 {
		t.Errorf("expected whisper port 9988, got %d", cfg.Whisper.Port)
	}
	if cfg.Hotkey.Modifier != "ralt" {
		t.Errorf("expected hotkey modifier ralt, got %s", cfg.Hotkey.Modifier)
	}
	if cfg.Hotkey.Key != "" {
		t.Errorf("expected empty hotkey key, got %s", cfg.Hotkey.Key)
	}
	if cfg.Input.Mode != "toggle" {
		t.Errorf("expected input mode toggle, got %s", cfg.Input.Mode)
	}
	if cfg.Audio.SampleRate != 16000 {
		t.Errorf("expected sample rate 16000, got %d", cfg.Audio.SampleRate)
	}
	if cfg.Output.Backend != "sendinput" {
		t.Errorf("expected backend sendinput, got %s", cfg.Output.Backend)
	}
	if !cfg.Overlay.Enabled {
		t.Error("expected overlay enabled by default")
	}
}

func TestDefaultValidates(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Errorf("default config should validate: %v", err)
	}
}

func TestLoadTOML(t *testing.T) {
	content := `
[whisper]
host = "192.168.1.100"
port = 8080
threads = 8

[hotkey]
modifier = "ctrl"
key = "r"

[input]
mode = "toggle"

[output]
backend = "sendinput"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cfg.Whisper.Host != "192.168.1.100" {
		t.Errorf("expected host 192.168.1.100, got %s", cfg.Whisper.Host)
	}
	if cfg.Whisper.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Whisper.Port)
	}
	if cfg.Whisper.Threads != 8 {
		t.Errorf("expected threads 8, got %d", cfg.Whisper.Threads)
	}
	if cfg.Hotkey.Modifier != "ctrl" || cfg.Hotkey.Key != "r" {
		t.Errorf("expected hotkey ctrl+r, got %s+%s", cfg.Hotkey.Modifier, cfg.Hotkey.Key)
	}
	if cfg.Input.Mode != "toggle" {
		t.Errorf("expected mode toggle, got %s", cfg.Input.Mode)
	}
	if cfg.Output.Backend != "sendinput" {
		t.Errorf("expected backend sendinput, got %s", cfg.Output.Backend)
	}
	// Defaults should be preserved for unset fields
	if cfg.Audio.SampleRate != 16000 {
		t.Errorf("expected default sample rate 16000, got %d", cfg.Audio.SampleRate)
	}
}

func TestValidateErrors(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Config)
		want   string
	}{
		{
			name:   "invalid port zero",
			modify: func(c *Config) { c.Whisper.Port = 0 },
			want:   "whisper.port",
		},
		{
			name:   "invalid port high",
			modify: func(c *Config) { c.Whisper.Port = 70000 },
			want:   "whisper.port",
		},
		{
			name:   "invalid modifier",
			modify: func(c *Config) { c.Hotkey.Modifier = "meta" },
			want:   "hotkey.modifier",
		},
		{
			name:   "invalid mode",
			modify: func(c *Config) { c.Input.Mode = "push" },
			want:   "input.mode",
		},
		{
			name:   "negative vad threshold",
			modify: func(c *Config) { c.Input.VADThreshold = -0.1 },
			want:   "input.vad_threshold",
		},
		{
			name:   "invalid backend",
			modify: func(c *Config) { c.Output.Backend = "wtype" },
			want:   "output.backend",
		},
		{
			name:   "zero overlay width",
			modify: func(c *Config) { c.Overlay.Width = 0 },
			want:   "overlay.width",
		},
		{
			name:   "fps too high",
			modify: func(c *Config) { c.Overlay.FPS = 500 },
			want:   "overlay.fps",
		},
		{
			name:   "zero threads",
			modify: func(c *Config) { c.Whisper.Threads = 0 },
			want:   "whisper.threads",
		},
		{
			name:   "invalid audio backend",
			modify: func(c *Config) { c.Audio.Backend = "coreaudio" },
			want:   "audio.backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.modify(&cfg)
			err := cfg.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("expected error containing %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("not valid [[[ toml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}
