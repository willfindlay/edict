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
	samples := []int16{5000, -5000, 5000, -5000}
	rms := normalizedRMS(samples)
	if rms <= 0 || rms >= 1.0 {
		t.Errorf("expected mid-range RMS, got %f", rms)
	}
}

func TestWaveformAddSamples(t *testing.T) {
	w := NewWaveform(400, 60)

	// Add samples multiple times
	for range 50 {
		w.AddSamples([]int16{10000, -10000, 5000, -5000})
	}

	// Verify targets are populated
	w.mu.Lock()
	defer w.mu.Unlock()

	nonZero := 0
	for _, v := range w.targets {
		if v > 0 {
			nonZero++
		}
	}
	if nonZero != pointCount {
		t.Errorf("expected all %d targets to be non-zero after 50 additions, got %d", pointCount, nonZero)
	}
}

func TestWaveformReset(t *testing.T) {
	w := NewWaveform(400, 60)
	w.AddSamples([]int16{10000, -10000})
	w.Reset()

	w.mu.Lock()
	defer w.mu.Unlock()

	for i, v := range w.targets {
		if v != 0 {
			t.Errorf("targets[%d] should be 0 after reset, got %f", i, v)
		}
	}
	for i, v := range w.smoothed {
		if v != 0 {
			t.Errorf("smoothed[%d] should be 0 after reset, got %f", i, v)
		}
	}
}

func TestWaveformAddEmptySamples(t *testing.T) {
	w := NewWaveform(400, 60)
	w.AddSamples(nil) // Should not panic

	w.mu.Lock()
	defer w.mu.Unlock()

	for i, v := range w.targets {
		if v != 0 {
			t.Errorf("targets[%d] should be 0 with no samples, got %f", i, v)
		}
	}
}
