//go:build windows

package config

var validBackends = map[string]bool{
	"sendinput": true,
}

const validBackendList = "sendinput"

// Default returns the default configuration for Windows.
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
			Backend: "sendinput",
		},
		Overlay: OverlayConfig{
			Enabled: true,
			Width:   400,
			Height:  60,
			FPS:     60,
		},
	}
}

// validatePlatform performs Windows-specific config validation.
func validatePlatform(c *Config) []string {
	var errs []string
	if c.Context.WSLDistro == "" {
		errs = append(errs, "context.wsl_distro is required on Windows")
	}
	return errs
}
