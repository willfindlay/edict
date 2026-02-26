//go:build windows

package config

var validBackends = map[string]bool{
	"sendinput": true,
}

const validBackendList = "sendinput"

var validAudioBackends = map[string]bool{
	"wasapi": true, "dsound": true, "winmm": true,
}

const validAudioBackendList = "wasapi, dsound, winmm"

// Default returns the default configuration for Windows.
func Default() Config {
	return Config{
		Whisper: WhisperConfig{
			ServerPath: "whisper-server",
			Host:       "localhost",
			Port:       9988,
			GPULayers:  -1,
			Threads:    4,
		},
		Hotkey: HotkeyConfig{
			Modifier: "ralt",
		},
		Input: InputConfig{
			Mode:         "toggle",
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
			Width:   700,
			Height:  80,
			FPS:     60,
		},
		Preview: PreviewConfig{
			Enabled:    true,
			IntervalMs: 1500,
			FontSize:   20,
			FontPath:   `C:\Windows\Fonts\segoeui.ttf`,
			MaxLines:   6,
			Padding:    12,
		},
	}
}

// validatePlatform performs Windows-specific config validation.
func validatePlatform(c *Config) []string {
	return nil
}
