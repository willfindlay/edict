package overlay

import (
	"math"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	pointCount         = 20
	baseHeight         = 8.0
	maxAmplitudeHeight = 45.0
	waveCount          = 3
	layerCount         = 5
)

type layerStyle struct {
	thick float32
	alpha float32
}

type waveSpec struct {
	glowColor      rl.Color               // halo color (outer glow layers)
	spatialFreq    float64                // peaks across width (lower = broader)
	temporalFreq   float64                // breathing speed
	phaseOffset    float64                // time offset between waves
	harmonic2Freq  float64                // second harmonic spatial freq
	harmonic2Amp   float64                // second harmonic amplitude (0-1)
	harmonic2Speed float64                // second harmonic temporal speed
	centerYOffset  float64                // vertical offset from center (px)
	lerpSpeed      float64                // amplitude smoothing speed
	gradBaseSpeed  float64                // gradient scroll idle speed
	gradSpeechMul  float64                // gradient scroll speech multiplier
	gradFreq       float64                // gradient spatial frequency
	layers         [layerCount]layerStyle // bloom, ultra-glow, glow, body, core
}

type waveState struct {
	smoothAmp     float64
	gradientPhase float64
}

var waves = [waveCount]waveSpec{
	// Ocean: few broad peaks, slow breathing, cyan glow.
	{
		glowColor:      rl.Color{R: 0, G: 200, B: 220, A: 255},
		spatialFreq:    0.35,
		temporalFreq:   0.8,
		phaseOffset:    0,
		harmonic2Freq:  0.55,
		harmonic2Amp:   0.3,
		harmonic2Speed: 1.1,
		centerYOffset:  -8,
		lerpSpeed:      4.0,
		gradBaseSpeed:  0.4,
		gradSpeechMul:  2.5,
		gradFreq:       0.8,
		layers: [layerCount]layerStyle{
			{thick: 50, alpha: 0.04}, // bloom (additive)
			{thick: 30, alpha: 0.08}, // ultra-glow (additive)
			{thick: 16, alpha: 0.20}, // glow (alpha)
			{thick: 6, alpha: 0.50},  // body (alpha)
			{thick: 2, alpha: 0.85},  // core (alpha, white)
		},
	},
	// Nebula: medium peaks, medium breathing, violet glow.
	{
		glowColor:      rl.Color{R: 140, G: 100, B: 255, A: 255},
		spatialFreq:    0.6,
		temporalFreq:   1.2,
		phaseOffset:    math.Pi / 3,
		harmonic2Freq:  0.85,
		harmonic2Amp:   0.25,
		harmonic2Speed: 1.5,
		centerYOffset:  0,
		lerpSpeed:      7.0,
		gradBaseSpeed:  0.5,
		gradSpeechMul:  3.0,
		gradFreq:       0.9,
		layers: [layerCount]layerStyle{
			{thick: 50, alpha: 0.04},
			{thick: 30, alpha: 0.08},
			{thick: 16, alpha: 0.20},
			{thick: 6, alpha: 0.50},
			{thick: 2, alpha: 0.85},
		},
	},
	// Aurora: many narrow peaks, fast breathing, hot pink glow.
	{
		glowColor:      rl.Color{R: 255, G: 100, B: 180, A: 255},
		spatialFreq:    0.95,
		temporalFreq:   1.8,
		phaseOffset:    2 * math.Pi / 3,
		harmonic2Freq:  1.3,
		harmonic2Amp:   0.2,
		harmonic2Speed: 2.2,
		centerYOffset:  8,
		lerpSpeed:      10.0,
		gradBaseSpeed:  0.6,
		gradSpeechMul:  3.5,
		gradFreq:       1.0,
		layers: [layerCount]layerStyle{
			{thick: 50, alpha: 0.04},
			{thick: 30, alpha: 0.08},
			{thick: 16, alpha: 0.20},
			{thick: 6, alpha: 0.50},
			{thick: 2, alpha: 0.85},
		},
	},
}

// Waveform renders audio amplitude as smooth glowing splines.
type Waveform struct {
	mu        sync.Mutex
	amplitude float64

	// Render-thread only (no mutex needed).
	waves  [waveCount]waveState
	time   float64
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

// AddSamples processes incoming audio samples and updates the amplitude.
func (w *Waveform) AddSamples(samples []int16) {
	if len(samples) == 0 {
		return
	}

	rms := normalizedRMS(samples)

	w.mu.Lock()
	defer w.mu.Unlock()
	w.amplitude = rms
}

// Draw renders the waveform splines. Must be called from the raylib render thread.
func (w *Waveform) Draw() {
	w.mu.Lock()
	amp := w.amplitude
	w.mu.Unlock()

	dt := float64(rl.GetFrameTime())
	w.time += dt

	centerY := float64(w.height) / 2.0
	spacing := float64(w.width) / float64(pointCount-1)

	white := rl.Color{R: 255, G: 255, B: 255, A: 255}

	// Pre-compute control points for all waves to avoid recomputing per layer pass.
	type wavePoints struct {
		points [pointCount + 2]rl.Vector2
	}
	var allPoints [waveCount]wavePoints

	for wi := 0; wi < waveCount; wi++ {
		spec := &waves[wi]
		ws := &w.waves[wi]

		// Lerp smoothAmp toward current amplitude at this wave's speed.
		alpha := 1.0 - math.Exp(-spec.lerpSpeed*dt)
		ws.smoothAmp += (amp - ws.smoothAmp) * alpha

		// Advance gradient phase: faster when speaking.
		ws.gradientPhase += (spec.gradBaseSpeed + ws.smoothAmp*spec.gradSpeechMul) * dt

		envelope := baseHeight + ws.smoothAmp*maxAmplitudeHeight

		// Build control points with standing wave + Gaussian envelope.
		for i := 0; i < pointCount; i++ {
			x := float64(i) * spacing

			// Gaussian center envelope: peaks at center, tapers at edges.
			normX := 2.0*float64(i)/float64(pointCount-1) - 1.0
			env := math.Exp(-0.8 * normX * normX)

			// Standing wave: spatial * temporal separation.
			y1 := math.Sin(float64(i)*spec.spatialFreq) *
				math.Sin(w.time*spec.temporalFreq+spec.phaseOffset)
			// Second harmonic for organic variation.
			y2 := math.Sin(float64(i)*spec.harmonic2Freq) *
				math.Sin(w.time*spec.harmonic2Speed+spec.phaseOffset*1.3)
			dy := y1 + spec.harmonic2Amp*y2

			y := centerY + spec.centerYOffset + envelope*env*dy
			allPoints[wi].points[i+1] = rl.Vector2{X: float32(x), Y: float32(y)}
		}

		// Phantom edges for Catmull-Rom tangents.
		allPoints[wi].points[0] = rl.Vector2{
			X: allPoints[wi].points[1].X - float32(spacing),
			Y: allPoints[wi].points[1].Y,
		}
		allPoints[wi].points[pointCount+1] = rl.Vector2{
			X: allPoints[wi].points[pointCount].X + float32(spacing),
			Y: allPoints[wi].points[pointCount].Y,
		}
	}

	// Pass 1: Additive blend for bloom + ultra-glow (layers 0-1).
	// Overlapping glow from different waves brightens additively.
	rl.BeginBlendMode(rl.BlendAdditive)
	for wi := 0; wi < waveCount; wi++ {
		spec := &waves[wi]
		ws := &w.waves[wi]
		pts := &allPoints[wi]

		for i := 0; i < pointCount-1; i++ {
			seg := [4]rl.Vector2{
				pts.points[i],
				pts.points[i+1],
				pts.points[i+2],
				pts.points[i+3],
			}

			// Per-segment gradient color from glow color.
			t := 0.5 + 0.5*math.Sin(ws.gradientPhase+float64(i)*spec.gradFreq)
			// Vary brightness by lerping glow color with a brighter version.
			bright := rl.Color{
				R: uint8(min(int(spec.glowColor.R)+60, 255)),
				G: uint8(min(int(spec.glowColor.G)+60, 255)),
				B: uint8(min(int(spec.glowColor.B)+60, 255)),
				A: 255,
			}
			c := lerpColor(spec.glowColor, bright, t)

			for li := 0; li < 2; li++ {
				layer := &spec.layers[li]
				rl.DrawSplineSegmentCatmullRom(seg[0], seg[1], seg[2], seg[3], layer.thick, rl.Fade(c, layer.alpha))
			}
		}
	}
	rl.EndBlendMode()

	// Pass 2: Alpha blend for glow, body, core (layers 2-4).
	rl.BeginBlendMode(rl.BlendAlpha)
	for wi := 0; wi < waveCount; wi++ {
		spec := &waves[wi]
		ws := &w.waves[wi]
		pts := &allPoints[wi]

		for i := 0; i < pointCount-1; i++ {
			seg := [4]rl.Vector2{
				pts.points[i],
				pts.points[i+1],
				pts.points[i+2],
				pts.points[i+3],
			}

			t := 0.5 + 0.5*math.Sin(ws.gradientPhase+float64(i)*spec.gradFreq)
			bright := rl.Color{
				R: uint8(min(int(spec.glowColor.R)+60, 255)),
				G: uint8(min(int(spec.glowColor.G)+60, 255)),
				B: uint8(min(int(spec.glowColor.B)+60, 255)),
				A: 255,
			}
			c := lerpColor(spec.glowColor, bright, t)

			// Layers 2-3: glow and body use wave color.
			for li := 2; li < 4; li++ {
				layer := &spec.layers[li]
				rl.DrawSplineSegmentCatmullRom(seg[0], seg[1], seg[2], seg[3], layer.thick, rl.Fade(c, layer.alpha))
			}
			// Layer 4: core uses white.
			coreLayer := &spec.layers[4]
			rl.DrawSplineSegmentCatmullRom(seg[0], seg[1], seg[2], seg[3], coreLayer.thick, rl.Fade(white, coreLayer.alpha))
		}
	}
	rl.EndBlendMode()
}

// Reset clears all amplitudes.
func (w *Waveform) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.amplitude = 0
	w.waves = [waveCount]waveState{}
	w.time = 0
}

// lerpColor linearly interpolates between two colors.
func lerpColor(a, b rl.Color, t float64) rl.Color {
	ft := float32(t)
	return rl.Color{
		R: uint8(float32(a.R) + ft*(float32(b.R)-float32(a.R))),
		G: uint8(float32(a.G) + ft*(float32(b.G)-float32(a.G))),
		B: uint8(float32(a.B) + ft*(float32(b.B)-float32(a.B))),
		A: 255,
	}
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
	rms *= 5.0
	if rms > 1.0 {
		rms = 1.0
	}
	return rms
}
