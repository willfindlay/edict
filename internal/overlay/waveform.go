package overlay

import (
	"math"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	pointCount         = 20
	lerpSpeed          = 10.0
	baseHeight         = 4.0
	maxAmplitudeHeight = 22.0

	glowThick = 14.0
	bodyThick = 8.0
	coreThick = 3.0

	glowAlpha = float32(0.15)
	bodyAlpha = float32(0.35)
	coreAlpha = float32(0.70)
)

var (
	tealColor = rl.Color{R: 0, G: 210, B: 210, A: 255}
	pinkColor = rl.Color{R: 255, G: 100, B: 180, A: 255}
)

// Waveform renders audio amplitude as smooth glowing splines.
type Waveform struct {
	mu       sync.Mutex
	targets  [pointCount]float64
	smoothed [pointCount]float64
	time     float64
	width    int
	height   int
}

// NewWaveform creates a waveform renderer for the given dimensions.
func NewWaveform(width, height int) *Waveform {
	return &Waveform{
		width:  width,
		height: height,
	}
}

// AddSamples processes incoming audio samples and updates point amplitudes.
func (w *Waveform) AddSamples(samples []int16) {
	if len(samples) == 0 {
		return
	}

	rms := normalizedRMS(samples)

	w.mu.Lock()
	defer w.mu.Unlock()

	copy(w.targets[:], w.targets[1:])
	w.targets[pointCount-1] = rms
}

// Draw renders the waveform splines. Must be called from the raylib render thread.
func (w *Waveform) Draw() {
	w.mu.Lock()
	targets := w.targets
	w.mu.Unlock()

	dt := float64(rl.GetFrameTime())
	w.time += dt

	// Lerp smoothed values toward targets.
	alpha := 1.0 - math.Exp(-lerpSpeed*dt)
	for i := range w.smoothed {
		w.smoothed[i] += (targets[i] - w.smoothed[i]) * alpha
	}

	centerY := float64(w.height) / 2.0

	type layer struct {
		color       rl.Color
		phaseOffset float64
	}
	layers := [2]layer{
		{tealColor, 0},
		{pinkColor, math.Pi / 3},
	}

	rl.BeginBlendMode(rl.BlendAdditive)
	for _, l := range layers {
		// Build control points: 20 visible + 2 phantom edges for Catmull-Rom.
		points := make([]rl.Vector2, pointCount+2)
		spacing := float64(w.width) / float64(pointCount-1)

		for i := 0; i < pointCount; i++ {
			x := float64(i) * spacing
			envelope := baseHeight + w.smoothed[i]*maxAmplitudeHeight
			y := centerY + envelope*math.Sin(w.time*1.5+float64(i)*0.8+l.phaseOffset)
			points[i+1] = rl.Vector2{X: float32(x), Y: float32(y)}
		}

		// Phantom edges: extend tangent from first/last visible points.
		points[0] = rl.Vector2{
			X: points[1].X - float32(spacing),
			Y: points[1].Y,
		}
		points[pointCount+1] = rl.Vector2{
			X: points[pointCount].X + float32(spacing),
			Y: points[pointCount].Y,
		}

		// Three passes: glow, body, core.
		rl.DrawSplineCatmullRom(points, glowThick, rl.Fade(l.color, glowAlpha))
		rl.DrawSplineCatmullRom(points, bodyThick, rl.Fade(l.color, bodyAlpha))
		rl.DrawSplineCatmullRom(points, coreThick, rl.Fade(l.color, coreAlpha))
	}
	rl.EndBlendMode()
}

// Reset clears all amplitudes.
func (w *Waveform) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.targets = [pointCount]float64{}
	w.smoothed = [pointCount]float64{}
	w.time = 0
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

	// Apply gain to make visualization more responsive.
	rms *= 3.0
	if rms > 1.0 {
		rms = 1.0
	}
	return rms
}
