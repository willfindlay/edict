package overlay

import (
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
	samples := []int16{2000, -2000, 2000, -2000}
	rms := normalizedRMS(samples)
	if rms <= 0 || rms >= 1.0 {
		t.Errorf("expected mid-range RMS, got %f", rms)
	}
}

func TestWaveformAddSamples(t *testing.T) {
	w := NewWaveform(400, 80)
	w.AddSamples([]int16{10000, -10000, 5000, -5000})

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.amplitude == 0 {
		t.Error("expected non-zero amplitude after adding samples")
	}
}

func TestWaveformReset(t *testing.T) {
	w := NewWaveform(400, 80)
	w.AddSamples([]int16{10000, -10000})
	w.Reset()

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.amplitude != 0 {
		t.Errorf("amplitude should be 0 after reset, got %f", w.amplitude)
	}
	for i := 0; i < waveCount; i++ {
		if w.waves[i].smoothAmp != 0 {
			t.Errorf("waves[%d].smoothAmp should be 0 after reset, got %f", i, w.waves[i].smoothAmp)
		}
		if w.waves[i].gradientPhase != 0 {
			t.Errorf("waves[%d].gradientPhase should be 0 after reset, got %f", i, w.waves[i].gradientPhase)
		}
	}
}

func TestWaveformAddEmptySamples(t *testing.T) {
	w := NewWaveform(400, 80)
	w.AddSamples(nil) // Should not panic

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.amplitude != 0 {
		t.Errorf("amplitude should be 0 with no samples, got %f", w.amplitude)
	}
}
