package overlay

import (
	"log"

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

	// Preview text box settings.
	PreviewEnabled bool
	FontSize       int
	FontPath       string
	MaxLines       int
	Padding        int
}

const textboxGap = 8 // pixels between waveform and textbox

// Window manages the transparent always-on-top overlay for the waveform display.
type Window struct {
	cfg      WindowConfig
	waveform *Waveform
	textbox  *Textbox
	visible  bool
	commands chan Command

	// Render-thread state for dynamic resize.
	screenW int
	screenH int
	curH    int

	textboxLogged bool // diagnostic: log once when textbox first becomes visible
}

// NewWindow creates an overlay window. Must be called from the main goroutine.
func NewWindow(cfg WindowConfig) *Window {
	w := &Window{
		cfg:      cfg,
		waveform: NewWaveform(cfg.Width, cfg.Height),
		commands: make(chan Command, 16),
	}
	if cfg.PreviewEnabled {
		w.textbox = NewTextbox(cfg.Width, cfg.FontSize, cfg.MaxLines, cfg.Padding, cfg.FontPath)
	}
	return w
}

// Commands returns the channel for controlling overlay visibility.
func (w *Window) Commands() chan<- Command {
	return w.commands
}

// AddSamples feeds audio samples to the waveform visualization.
func (w *Window) AddSamples(samples []int16) {
	w.waveform.AddSamples(samples)
}

// SetPreviewText updates the live transcription preview. Thread-safe.
func (w *Window) SetPreviewText(text string) {
	if w.textbox != nil {
		w.textbox.SetText(text)
	}
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
			rl.FlagWindowMousePassthrough |
			rl.FlagMsaa4xHint,
	)

	rl.InitWindow(int32(w.cfg.Width), int32(w.cfg.Height), "edict")
	defer rl.CloseWindow()
	w.waveform.InitGPU()
	defer w.waveform.Close()
	if w.textbox != nil {
		w.textbox.InitGPU()
		defer w.textbox.Close()
	}
	hideFromTaskbar()

	// Must query after InitWindow (needs GLFW context)
	monitor := rl.GetCurrentMonitor()
	w.screenW = rl.GetMonitorWidth(monitor)
	w.screenH = rl.GetMonitorHeight(monitor)
	w.curH = w.cfg.Height
	posX := (w.screenW - w.cfg.Width) / 2
	posY := w.screenH - w.curH - 60
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
				if w.textbox != nil {
					w.textbox.SetText("")
				}
			case Hide:
				w.visible = false
				w.waveform.Reset()
				if w.textbox != nil {
					w.textbox.SetText("")
				}
				w.textboxLogged = false
				w.resizeWindow(w.cfg.Height)
			}
		default:
		}

		// Update textbox state (reflow, height) before resize check.
		textboxH := float32(0)
		if w.visible && w.textbox != nil {
			w.textbox.Update()
			textboxH = w.textbox.Height()
		}

		// Compute total height: textbox + gap + waveform.
		wantH := w.cfg.Height
		if textboxH > 0 {
			if !w.textboxLogged {
				log.Printf("textbox: h=%.0f lines=%d yOffset=0 waveH=%d",
					textboxH, len(w.textbox.lines), w.cfg.Height)
				w.textboxLogged = true
			}
			wantH = int(textboxH) + textboxGap + w.cfg.Height
		}
		if wantH != w.curH {
			w.resizeWindow(wantH)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})

		if w.visible {
			waveYOffset := float32(0)
			if w.textbox != nil && textboxH > 0 {
				w.textbox.Draw(0)
				waveYOffset = textboxH + textboxGap
			}
			w.waveform.Draw(waveYOffset)
		}

		rl.EndDrawing()
	}
}

// resizeWindow adjusts window height while keeping the bottom edge pinned.
func (w *Window) resizeWindow(h int) {
	w.curH = h
	rl.SetWindowSize(w.cfg.Width, h)
	posX := (w.screenW - w.cfg.Width) / 2
	posY := w.screenH - h - 60
	rl.SetWindowPosition(posX, posY)
}
