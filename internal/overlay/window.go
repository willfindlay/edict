package overlay

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Command controls overlay visibility.
type Command int

const (
	Show Command = iota
	Hide
)

// WindowConfig holds overlay window settings.
type WindowConfig struct {
	Width   int
	Height  int
	FPS     int
	Enabled bool
}

// Window manages the transparent always-on-top overlay for the waveform display.
type Window struct {
	cfg      WindowConfig
	waveform *Waveform
	visible  bool
	commands chan Command
}

// NewWindow creates an overlay window. Must be called from the main goroutine.
func NewWindow(cfg WindowConfig) *Window {
	return &Window{
		cfg:      cfg,
		waveform: NewWaveform(cfg.Width, cfg.Height),
		commands: make(chan Command, 16),
	}
}

// Commands returns the channel for controlling overlay visibility.
func (w *Window) Commands() chan<- Command {
	return w.commands
}

// AddSamples feeds audio samples to the waveform visualization.
func (w *Window) AddSamples(samples []int16) {
	w.waveform.AddSamples(samples)
}

// Run starts the raylib event loop. Must be called from the main goroutine (OpenGL).
// Blocks until the window is closed.
func (w *Window) Run(done <-chan struct{}) {
	if !w.cfg.Enabled {
		// If overlay disabled, just wait for done signal
		<-done
		return
	}

	rl.SetConfigFlags(
		rl.FlagWindowTransparent |
			rl.FlagWindowUndecorated |
			rl.FlagWindowTopmost |
			rl.FlagWindowMousePassthrough,
	)

	rl.InitWindow(int32(w.cfg.Width), int32(w.cfg.Height), "edict")
	defer rl.CloseWindow()
	hideFromTaskbar()

	// Must query after InitWindow (needs GLFW context)
	monitor := rl.GetCurrentMonitor()
	screenW := rl.GetMonitorWidth(monitor)
	screenH := rl.GetMonitorHeight(monitor)
	posX := (screenW - int(w.cfg.Width)) / 2
	posY := screenH - int(w.cfg.Height) - 40
	rl.SetWindowPosition(posX, posY)
	rl.SetTargetFPS(int32(w.cfg.FPS))

	for !rl.WindowShouldClose() {
		select {
		case <-done:
			return
		case cmd := <-w.commands:
			switch cmd {
			case Show:
				w.visible = true
			case Hide:
				w.visible = false
				w.waveform.Reset()
			}
		default:
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})

		if w.visible {
			w.waveform.Draw()
		}

		rl.EndDrawing()
	}
}
