package overlay

import (
	"math"
	"testing"
)

func TestNormalizedRMSEmpty(t *testing.T) {
	if rms := normalizedRMS(nil); rms != 0 {
		t.Errorf("expected 0, got %f", rms)
	}
}

func TestNormalizedRMSSilence(t *testing.T) {
	if rms := normalizedRMS([]int16{0, 0, 0, 0}); rms != 0 {
		t.Errorf("expected 0, got %f", rms)
	}
}

func TestNormalizedRMSCapped(t *testing.T) {
	// Full-scale signal with 3x gain should cap at 1.0
	samples := []int16{32767, 32767, 32767, 32767}
	rms := normalizedRMS(samples)
	if rms != 1.0 {
		t.Errorf("expected 1.0, got %f", rms)
	}
}

func TestNormalizedRMSMidRange(t *testing.T) {
	samples := []int16{5000, -5000, 5000, -5000}
	rms := normalizedRMS(samples)
	if rms <= 0 || rms >= 1.0 {
		t.Errorf("expected mid-range RMS, got %f", rms)
	}
}

func TestBarColorGreen(t *testing.T) {
	c := barColor(0.0)
	if c.R != 0 || c.G != 200 || c.B != 0 {
		t.Errorf("expected green at 0.0, got R=%d G=%d B=%d", c.R, c.G, c.B)
	}
}

func TestBarColorCyan(t *testing.T) {
	c := barColor(0.5)
	if c.R != 0 || c.G != 200 || c.B != 200 {
		t.Errorf("expected cyan at 0.5, got R=%d G=%d B=%d", c.R, c.G, c.B)
	}
}

func TestBarColorMagenta(t *testing.T) {
	c := barColor(1.0)
	if c.R != 200 || c.G != 0 || c.B != 200 {
		t.Errorf("expected magenta at 1.0, got R=%d G=%d B=%d", c.R, c.G, c.B)
	}
}

func TestBarColorClamp(t *testing.T) {
	// Values outside [0,1] should be clamped
	c := barColor(-0.5)
	cZero := barColor(0.0)
	if c != cZero {
		t.Errorf("negative value should clamp to 0: got %v, want %v", c, cZero)
	}

	c = barColor(1.5)
	cOne := barColor(1.0)
	if c != cOne {
		t.Errorf("value >1 should clamp to 1: got %v, want %v", c, cOne)
	}
}

func TestWaveformAddSamples(t *testing.T) {
	w := NewWaveform(400, 60)

	// Add samples multiple times
	for range 50 {
		w.AddSamples([]int16{10000, -10000, 5000, -5000})
	}

	// Verify bars are populated
	w.mu.Lock()
	defer w.mu.Unlock()

	nonZero := 0
	for _, b := range w.bars {
		if b > 0 {
			nonZero++
		}
	}
	if nonZero != barCount {
		t.Errorf("expected all %d bars to be non-zero after 50 additions, got %d", barCount, nonZero)
	}
}

func TestWaveformReset(t *testing.T) {
	w := NewWaveform(400, 60)
	w.AddSamples([]int16{10000, -10000})
	w.Reset()

	w.mu.Lock()
	defer w.mu.Unlock()

	for i, b := range w.bars {
		if b != 0 {
			t.Errorf("bar[%d] should be 0 after reset, got %f", i, b)
		}
	}
}

func TestWaveformAddEmptySamples(t *testing.T) {
	w := NewWaveform(400, 60)
	w.AddSamples(nil) // Should not panic

	w.mu.Lock()
	defer w.mu.Unlock()

	for i, b := range w.bars {
		if b != 0 {
			t.Errorf("bar[%d] should be 0 with no samples, got %f", i, b)
		}
	}
}

func TestBarColorGradientSmoothness(t *testing.T) {
	// Verify the gradient produces smooth transitions
	prev := barColor(0.0)
	for i := 1; i <= 100; i++ {
		amp := float64(i) / 100.0
		c := barColor(amp)
		// Each channel should change by at most ~4 per step (200/50 steps)
		dr := math.Abs(float64(c.R) - float64(prev.R))
		dg := math.Abs(float64(c.G) - float64(prev.G))
		db := math.Abs(float64(c.B) - float64(prev.B))
		if dr > 5 || dg > 5 || db > 5 {
			t.Errorf("gradient too jumpy at amp=%.2f: dR=%.0f dG=%.0f dB=%.0f", amp, dr, dg, db)
		}
		prev = c
	}
}
