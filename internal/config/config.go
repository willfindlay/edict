package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Whisper WhisperConfig `toml:"whisper"`
	Hotkey  HotkeyConfig  `toml:"hotkey"`
	Input   InputConfig   `toml:"input"`
	Audio   AudioConfig   `toml:"audio"`
	Output  OutputConfig  `toml:"output"`
	Overlay OverlayConfig `toml:"overlay"`
	Context ContextConfig `toml:"context"`
}

type WhisperConfig struct {
	ServerPath string `toml:"server_path"`
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	ModelPath  string `toml:"model_path"`
	GPULayers  int    `toml:"gpu_layers"`
	Threads    int    `toml:"threads"`
}

type HotkeyConfig struct {
	Modifier string `toml:"modifier"`
	Key      string `toml:"key"`
}

type InputConfig struct {
	Mode         string  `toml:"mode"`
	VADThreshold float64 `toml:"vad_threshold"`
	VADSilenceMs int     `toml:"vad_silence_ms"`
}

type AudioConfig struct {
	SampleRate int    `toml:"sample_rate"`
	Channels   int    `toml:"channels"`
	Backend    string `toml:"backend"`
}

type OutputConfig struct {
	Backend          string `toml:"backend"`
	KeystrokeDelayUs int    `toml:"keystroke_delay_us"`
}

type OverlayConfig struct {
	Enabled bool `toml:"enabled"`
	Width   int  `toml:"width"`
	Height  int  `toml:"height"`
	FPS     int  `toml:"fps"`
}

// ContextConfig controls Claude Code process detection and context file access.
// On Windows, edict shells into WSL to find Claude Code processes and reads
// context files via \\wsl.localhost\<distro>\... UNC paths.
type ContextConfig struct {
	WSLDistro string `toml:"wsl_distro"`
	WSLHome   string `toml:"wsl_home"`
}

func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

var validModifiers = map[string]bool{
	"alt": true, "lalt": true, "ralt": true,
	"ctrl": true, "lctrl": true, "rctrl": true,
	"shift": true, "lshift": true, "rshift": true,
	"super": true, "lsuper": true, "rsuper": true,
}

var validModes = map[string]bool{
	"hold": true, "toggle": true, "auto": true,
}

func (c *Config) Validate() error {
	var errs []string

	if c.Whisper.Port < 1 || c.Whisper.Port > 65535 {
		errs = append(errs, "whisper.port must be between 1 and 65535")
	}
	if c.Whisper.Threads < 1 {
		errs = append(errs, "whisper.threads must be at least 1")
	}

	if !validModifiers[c.Hotkey.Modifier] {
		errs = append(errs, fmt.Sprintf("hotkey.modifier must be one of: alt, lalt, ralt, ctrl, lctrl, rctrl, shift, lshift, rshift, super, lsuper, rsuper (got %q)", c.Hotkey.Modifier))
	}

	if !validModes[c.Input.Mode] {
		errs = append(errs, fmt.Sprintf("input.mode must be one of: hold, toggle, auto (got %q)", c.Input.Mode))
	}
	if c.Input.VADThreshold < 0 || c.Input.VADThreshold > 1 {
		errs = append(errs, "input.vad_threshold must be between 0.0 and 1.0")
	}
	if c.Input.VADSilenceMs < 0 {
		errs = append(errs, "input.vad_silence_ms must be non-negative")
	}

	if c.Audio.SampleRate < 8000 || c.Audio.SampleRate > 48000 {
		errs = append(errs, "audio.sample_rate must be between 8000 and 48000")
	}
	if c.Audio.Channels < 1 || c.Audio.Channels > 2 {
		errs = append(errs, "audio.channels must be 1 or 2")
	}
	if c.Audio.Backend != "" && !validAudioBackends[c.Audio.Backend] {
		errs = append(errs, fmt.Sprintf("audio.backend must be one of: %s (got %q)", validAudioBackendList, c.Audio.Backend))
	}

	if !validBackends[c.Output.Backend] {
		errs = append(errs, fmt.Sprintf("output.backend must be one of: %s (got %q)", validBackendList, c.Output.Backend))
	}

	if c.Overlay.Width < 1 {
		errs = append(errs, "overlay.width must be positive")
	}
	if c.Overlay.Height < 1 {
		errs = append(errs, "overlay.height must be positive")
	}
	if c.Overlay.FPS < 1 || c.Overlay.FPS > 240 {
		errs = append(errs, "overlay.fps must be between 1 and 240")
	}

	errs = append(errs, validatePlatform(c)...)

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
