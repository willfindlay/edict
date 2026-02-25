package overlay

import (
	"os"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// premultiplyFragSrc is a minimal shader that premultiplies RGB by alpha
// for correct compositing on DWM transparent windows.
const premultiplyFragSrc = `#version 330
in vec2 fragTexCoord;
out vec4 finalColor;
uniform sampler2D texture0;

void main() {
    vec4 color = texture(texture0, fragTexCoord);
    color.rgb *= color.a;
    finalColor = color;
}
`

// Textbox renders word-wrapped text in a semi-transparent box.
// Thread safety: SetText is safe to call from any goroutine.
// All other methods are render-thread only.
type Textbox struct {
	mu    sync.Mutex
	text  string
	dirty bool

	// Config (immutable after construction).
	width    int
	fontSize int
	maxLines int
	padding  int
	fontPath string

	// Render-thread state.
	lines  []string
	font   rl.Font
	rt     rl.RenderTexture2D
	shader rl.Shader
	boxH   float32
}

// NewTextbox creates a textbox renderer. GPU resources are not allocated
// until InitGPU is called from the render thread.
func NewTextbox(width, fontSize, maxLines, padding int, fontPath string) *Textbox {
	return &Textbox{
		width:    width,
		fontSize: fontSize,
		maxLines: maxLines,
		padding:  padding,
		fontPath: fontPath,
	}
}

// InitGPU loads the font, render texture, and premultiply shader.
// Must be called from the render thread after rl.InitWindow.
func (tb *Textbox) InitGPU() {
	tb.font = loadFont(tb.fontPath, tb.fontSize)

	// Allocate an RT tall enough for max lines + padding.
	// lineSpacing = fontSize * 1.2, plus top/bottom padding.
	lineH := float32(tb.fontSize) * 1.2
	maxH := int32(lineH*float32(tb.maxLines) + float32(tb.padding*2))
	tb.rt = rl.LoadRenderTexture(int32(tb.width), maxH)
	rl.SetTextureFilter(tb.rt.Texture, rl.FilterBilinear)

	tb.shader = rl.LoadShaderFromMemory("", premultiplyFragSrc)
}

// Close frees GPU resources. Must be called from the render thread.
func (tb *Textbox) Close() {
	rl.UnloadRenderTexture(tb.rt)
	rl.UnloadShader(tb.shader)
	rl.UnloadFont(tb.font)
}

// SetText updates the displayed text. Thread-safe.
func (tb *Textbox) SetText(text string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.text != text {
		tb.text = text
		tb.dirty = true
	}
}

// Height returns the current box height in pixels (0 when empty).
// Render-thread only.
func (tb *Textbox) Height() float32 {
	return tb.boxH
}

// Update processes pending text changes and recomputes layout.
// Must be called before Height() each frame. Render-thread only.
func (tb *Textbox) Update() {
	tb.mu.Lock()
	text := tb.text
	dirty := tb.dirty
	tb.dirty = false
	tb.mu.Unlock()

	if dirty {
		tb.reflow(text)
	}

	if len(tb.lines) == 0 {
		tb.boxH = 0
	} else {
		lineH := float32(tb.fontSize) * 1.2
		tb.boxH = lineH*float32(len(tb.lines)) + float32(tb.padding*2)
	}
}

// Draw renders the text box at the given y offset. Render-thread only.
// Call Update() first to ensure layout is current.
//
// Diagnostic: bypasses RT/shader compositing, draws directly to the default
// framebuffer. If the textbox is visible with this path, the RT pipeline is
// the culprit. If still invisible, the issue is coordinates/viewport/DWM.
func (tb *Textbox) Draw(yOffset float32) {
	if len(tb.lines) == 0 {
		return
	}

	lineH := float32(tb.fontSize) * 1.2

	// Background rectangle (directly to screen).
	rl.DrawRectangleRounded(
		rl.Rectangle{X: 0, Y: yOffset, Width: float32(tb.width), Height: tb.boxH},
		0.03, 8,
		rl.Color{R: 0, G: 0, B: 0, A: 180},
	)

	// Draw each line of text (directly to screen).
	for i, line := range tb.lines {
		y := yOffset + float32(tb.padding) + lineH*float32(i)
		rl.DrawTextEx(tb.font, line,
			rl.Vector2{X: float32(tb.padding), Y: y},
			float32(tb.fontSize), 0,
			rl.Color{R: 255, G: 255, B: 255, A: 230},
		)
	}
}

// reflow performs word wrapping. Render-thread only.
func (tb *Textbox) reflow(text string) {
	if text == "" {
		tb.lines = nil
		return
	}

	maxW := float32(tb.width - tb.padding*2)
	words := splitWords(text)
	var lines []string
	var current string

	for _, word := range words {
		candidate := current
		if candidate != "" {
			candidate += " "
		}
		candidate += word

		w := rl.MeasureTextEx(tb.font, candidate, float32(tb.fontSize), 0).X
		if w > maxW && current != "" {
			lines = append(lines, current)
			current = word
		} else {
			current = candidate
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	// Cap at maxLines, dropping oldest.
	if len(lines) > tb.maxLines {
		lines = lines[len(lines)-tb.maxLines:]
	}

	tb.lines = lines
}

// splitWords splits text on spaces, keeping words intact.
func splitWords(text string) []string {
	var words []string
	start := -1
	for i, r := range text {
		if r == ' ' || r == '\t' || r == '\n' {
			if start >= 0 {
				words = append(words, text[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		words = append(words, text[start:])
	}
	return words
}

// loadFont attempts to load the specified font, falling back to raylib default.
func loadFont(path string, size int) rl.Font {
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			f := rl.LoadFontEx(path, int32(size), nil)
			if f.BaseSize > 0 {
				rl.SetTextureFilter(f.Texture, rl.FilterBilinear)
				return f
			}
		}
	}
	return rl.GetFontDefault()
}
