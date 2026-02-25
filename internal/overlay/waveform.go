package overlay

import (
	"math"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	barCount   = 40
	barPadding = 2
	minBarH    = 2
)

// Waveform renders audio amplitude as colored bars.
type Waveform struct {
	mu     sync.Mutex
	bars   [barCount]float64
	width  int
	height int
}

// NewWaveform creates a waveform renderer for the given dimensions.
func NewWaveform(width, height int) *Waveform {
	return &Waveform{
		width:  width,
		height: height,
	}
}

// AddSamples processes incoming audio samples and updates bar amplitudes.
func (w *Waveform) AddSamples(samples []int16) {
	if len(samples) == 0 {
		return
	}

	rms := normalizedRMS(samples)

	w.mu.Lock()
	defer w.mu.Unlock()

	// Shift bars left
	copy(w.bars[:], w.bars[1:])
	w.bars[barCount-1] = rms
}

// Draw renders the waveform bars. Must be called from the raylib render thread.
func (w *Waveform) Draw() {
	w.mu.Lock()
	bars := w.bars
	w.mu.Unlock()

	barW := float64(w.width)/float64(barCount) - barPadding
	if barW < 1 {
		barW = 1
	}

	for i, amp := range bars {
		barH := amp * float64(w.height)
		if barH < minBarH {
			barH = minBarH
		}
		if barH > float64(w.height) {
			barH = float64(w.height)
		}

		x := float64(i) * (barW + barPadding)
		y := float64(w.height) - barH

		color := barColor(amp)
		rl.DrawRectangleRounded(
			rl.Rectangle{
				X:      float32(x),
				Y:      float32(y),
				Width:  float32(barW),
				Height: float32(barH),
			},
			0.3, 4,
			color,
		)
	}
}

// Reset clears all bar amplitudes.
func (w *Waveform) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bars = [barCount]float64{}
}

// barColor returns a color based on amplitude: green -> cyan -> magenta.
func barColor(amp float64) rl.Color {
	// Clamp
	if amp < 0 {
		amp = 0
	}
	if amp > 1 {
		amp = 1
	}

	var r, g, b uint8
	if amp < 0.5 {
		// Green to cyan (0.0 -> 0.5)
		t := amp * 2
		r = 0
		g = 200
		b = uint8(200 * t)
	} else {
		// Cyan to magenta (0.5 -> 1.0)
		t := (amp - 0.5) * 2
		r = uint8(200 * t)
		g = uint8(200 * (1 - t))
		b = 200
	}

	return rl.Color{R: r, G: g, B: b, A: 220}
}

// normalizedRMS computes root-mean-square of int16 samples, normalized to [0, 1].
func normalizedRMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64
	for _, s := range samples {
		f := float64(s) / 32768.0
		sum += f * f
	}
	rms := math.Sqrt(sum / float64(len(samples)))

	// Apply gain to make visualization more responsive
	rms *= 3.0
	if rms > 1.0 {
		rms = 1.0
	}
	return rms
}
