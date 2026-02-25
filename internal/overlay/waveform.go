package overlay

import (
	"math"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	pointCount         = 20
	baseHeight         = 8.0
	maxAmplitudeHeight = 70.0
	waveCount          = 3
)

type waveSpec struct {
	glowColor      rl.Color // halo color
	spatialFreq    float64  // peaks across width (lower = broader)
	temporalFreq   float64  // breathing speed
	phaseOffset    float64  // time offset between waves
	harmonic2Freq  float64  // second harmonic spatial freq
	harmonic2Amp   float64  // second harmonic amplitude (0-1)
	harmonic2Speed float64  // second harmonic temporal speed
	centerYOffset  float64  // vertical offset from center (px)
	attackSpeed    float64  // amplitude rise speed (sound onset)
	decaySpeed     float64  // amplitude fall speed (sound offset)
	gradBaseSpeed  float64  // gradient scroll idle speed
	gradSpeechMul  float64  // gradient scroll speech multiplier
	gradFreq       float64  // gradient spatial frequency
	bodyThick      float32  // thickness for body spline
	coreThick      float32  // thickness for white core spline
	glowThick      float32  // thickness for glow source (pre-blur)
	bodyAlpha      float32  // alpha for body
	coreAlpha      float32  // alpha for white core
	glowAlpha      float32  // alpha for glow source drawn to RT
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
		attackSpeed:    8.0,
		decaySpeed:     2.0,
		gradBaseSpeed:  0.4,
		gradSpeechMul:  2.5,
		gradFreq:       0.8,
		bodyThick:      6,
		coreThick:      2,
		glowThick:      32,
		bodyAlpha:      0.50,
		coreAlpha:      0.85,
		glowAlpha:      0.70,
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
		attackSpeed:    14.0,
		decaySpeed:     3.5,
		gradBaseSpeed:  0.5,
		gradSpeechMul:  3.0,
		gradFreq:       0.9,
		bodyThick:      6,
		coreThick:      2,
		glowThick:      32,
		bodyAlpha:      0.50,
		coreAlpha:      0.85,
		glowAlpha:      0.70,
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
		attackSpeed:    20.0,
		decaySpeed:     5.0,
		gradBaseSpeed:  0.6,
		gradSpeechMul:  3.5,
		gradFreq:       1.0,
		bodyThick:      6,
		coreThick:      2,
		glowThick:      32,
		bodyAlpha:      0.50,
		coreAlpha:      0.85,
		glowAlpha:      0.70,
	},
}

// GLSL 330 single-direction Gaussian blur (9-tap, sigma ~8px).
// The direction uniform selects horizontal (1,0) or vertical (0,1).
const blurFragSrc = `#version 330
in vec2 fragTexCoord;
out vec4 finalColor;
uniform sampler2D texture0;
uniform vec2 direction;
uniform vec2 resolution;

float inBounds(vec2 uv) {
    return step(0.0, uv.y) * step(uv.y, 1.0);
}

void main() {
    vec2 texelSize = direction / resolution;
    float weights[9] = float[](
        0.10855, 0.10548, 0.09672, 0.08370, 0.06840,
        0.05276, 0.03840, 0.02637, 0.01710
    );
    vec4 color = texture(texture0, fragTexCoord) * weights[0];
    for (int i = 1; i < 9; i++) {
        vec2 offset = texelSize * float(i) * 2.0;
        vec2 uvPlus = fragTexCoord + offset;
        vec2 uvMinus = fragTexCoord - offset;
        color += texture(texture0, uvPlus) * weights[i] * inBounds(uvPlus);
        color += texture(texture0, uvMinus) * weights[i] * inBounds(uvMinus);
    }
    finalColor = color;
}
`

// Waveform renders audio amplitude as smooth glowing splines.
type Waveform struct {
	mu        sync.Mutex
	amplitude float64

	// Render-thread only (no mutex needed).
	waves  [waveCount]waveState
	time   float64
	width  int
	height int

	// GPU resources (render-thread only, created by InitGPU).
	glowRT     rl.RenderTexture2D
	blurRT     rl.RenderTexture2D
	blurShader rl.Shader
	dirLoc     int32
	sizeLoc    int32
}

// NewWaveform creates a waveform renderer for the given dimensions.
func NewWaveform(width, height int) *Waveform {
	return &Waveform{
		width:  width,
		height: height,
	}
}

// InitGPU loads GPU resources for the blur bloom pipeline.
// Must be called from the render thread after rl.InitWindow.
func (w *Waveform) InitGPU() {
	w.glowRT = rl.LoadRenderTexture(int32(w.width), int32(w.height))
	w.blurRT = rl.LoadRenderTexture(int32(w.width), int32(w.height))
	w.blurShader = rl.LoadShaderFromMemory("", blurFragSrc)
	w.dirLoc = rl.GetShaderLocation(w.blurShader, "direction")
	w.sizeLoc = rl.GetShaderLocation(w.blurShader, "resolution")
	rl.SetShaderValue(w.blurShader, w.sizeLoc,
		[]float32{float32(w.width), float32(w.height)},
		rl.ShaderUniformVec2)
}

// Close frees GPU resources. Must be called from the render thread
// before rl.CloseWindow.
func (w *Waveform) Close() {
	rl.UnloadRenderTexture(w.glowRT)
	rl.UnloadRenderTexture(w.blurRT)
	rl.UnloadShader(w.blurShader)
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

	// Pre-compute control points for all waves.
	type wavePoints struct {
		points [pointCount + 2]rl.Vector2
	}
	var allPoints [waveCount]wavePoints

	for wi := 0; wi < waveCount; wi++ {
		spec := &waves[wi]
		ws := &w.waves[wi]

		// Asymmetric lerp: fast attack, slow decay.
		speed := spec.decaySpeed
		if amp > ws.smoothAmp {
			speed = spec.attackSpeed
		}
		alpha := 1.0 - math.Exp(-speed*dt)
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

	// Helper to compute per-segment gradient color.
	segColor := func(spec *waveSpec, ws *waveState, i int) rl.Color {
		t := 0.5 + 0.5*math.Sin(ws.gradientPhase+float64(i)*spec.gradFreq)
		bright := rl.Color{
			R: uint8(min(int(spec.glowColor.R)+60, 255)),
			G: uint8(min(int(spec.glowColor.G)+60, 255)),
			B: uint8(min(int(spec.glowColor.B)+60, 255)),
			A: 255,
		}
		return lerpColor(spec.glowColor, bright, t)
	}

	// Step 1: Draw body + core to screen (alpha blend).
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
			c := segColor(spec, ws, i)
			rl.DrawSplineSegmentCatmullRom(seg[0], seg[1], seg[2], seg[3], spec.bodyThick, rl.Fade(c, spec.bodyAlpha))
			rl.DrawSplineSegmentCatmullRom(seg[0], seg[1], seg[2], seg[3], spec.coreThick, rl.Fade(white, spec.coreAlpha))
		}
	}
	rl.EndBlendMode()

	// Step 2: Draw glow source to glowRT.
	rl.BeginTextureMode(w.glowRT)
	rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
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
			c := segColor(spec, ws, i)
			rl.DrawSplineSegmentCatmullRom(seg[0], seg[1], seg[2], seg[3], spec.glowThick, rl.Fade(c, spec.glowAlpha))
		}
	}
	rl.EndBlendMode()
	rl.EndTextureMode()

	// Render texture source rect (Y-flipped for OpenGL).
	flippedRect := rl.Rectangle{
		X: 0, Y: 0,
		Width:  float32(w.width),
		Height: -float32(w.height),
	}

	// Single-pass Gaussian blur (H→V) for a tight halo near the splines.
	rl.BeginTextureMode(w.blurRT)
	rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
	rl.BeginShaderMode(w.blurShader)
	rl.SetShaderValue(w.blurShader, w.dirLoc, []float32{1, 0}, rl.ShaderUniformVec2)
	rl.DrawTextureRec(w.glowRT.Texture, flippedRect, rl.Vector2{X: 0, Y: 0}, white)
	rl.EndShaderMode()
	rl.EndTextureMode()

	rl.BeginTextureMode(w.glowRT)
	rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
	rl.BeginShaderMode(w.blurShader)
	rl.SetShaderValue(w.blurShader, w.dirLoc, []float32{0, 1}, rl.ShaderUniformVec2)
	rl.DrawTextureRec(w.blurRT.Texture, flippedRect, rl.Vector2{X: 0, Y: 0}, white)
	rl.EndShaderMode()
	rl.EndTextureMode()

	// Composite blurred glow to screen (additive, drawn 8x for intensity).
	rl.BeginBlendMode(rl.BlendAdditive)
	for range 8 {
		rl.DrawTextureRec(w.glowRT.Texture, flippedRect, rl.Vector2{X: 0, Y: 0}, white)
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
	rms *= 8.0
	if rms > 1.0 {
		rms = 1.0
	}
	return rms
}
