//go:build linux

package config

var validBackends = map[string]bool{
	"ydotool": true, "xdotool": true,
}

const validBackendList = "ydotool, xdotool"

// Default returns the default configuration for Linux.
func Default() Config {
	return Config{
		Whisper: WhisperConfig{
			ServerPath: "whisper-server",
			Host:       "127.0.0.1",
			Port:       9988,
			GPULayers:  -1,
			Threads:    4,
		},
		Hotkey: HotkeyConfig{
			Modifier: "alt",
			Key:      "m",
		},
		Input: InputConfig{
			Mode:         "hold",
			VADThreshold: 0.02,
			VADSilenceMs: 800,
		},
		Audio: AudioConfig{
			SampleRate: 16000,
			Channels:   1,
		},
		Output: OutputConfig{
			Backend: "ydotool",
		},
		Overlay: OverlayConfig{
			Enabled: true,
			Width:   400,
			Height:  60,
			FPS:     60,
		},
	}
}

// validatePlatform performs Linux-specific config validation.
func validatePlatform(_ *Config) []string {
	return nil
}
