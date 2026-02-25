package input

import (
	"math"
	"testing"

	"github.com/willfindlay/edict/internal/hotkey"
)

// --- Hold mode tests ---

func TestHoldPressRelease(t *testing.T) {
	h := NewHold()

	if a := h.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Start {
		t.Errorf("press: expected Start, got %v", a)
	}
	if a := h.HandleEvent(hotkey.Event{Type: hotkey.Release}); a != Stop {
		t.Errorf("release: expected Stop, got %v", a)
	}
}

func TestHoldDoublePress(t *testing.T) {
	h := NewHold()

	h.HandleEvent(hotkey.Event{Type: hotkey.Press})
	if a := h.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != None {
		t.Errorf("double press: expected None, got %v", a)
	}
}

func TestHoldReleaseWithoutPress(t *testing.T) {
	h := NewHold()

	if a := h.HandleEvent(hotkey.Event{Type: hotkey.Release}); a != None {
		t.Errorf("release without press: expected None, got %v", a)
	}
}

func TestHoldReset(t *testing.T) {
	h := NewHold()
	h.HandleEvent(hotkey.Event{Type: hotkey.Press})
	h.Reset()

	if a := h.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Start {
		t.Errorf("after reset press: expected Start, got %v", a)
	}
}

func TestHoldIgnoresAudio(t *testing.T) {
	h := NewHold()
	if a := h.HandleAudio([]int16{1000, 2000}); a != None {
		t.Errorf("hold HandleAudio: expected None, got %v", a)
	}
}

// --- Toggle mode tests ---

func TestToggleOnOff(t *testing.T) {
	tg := NewToggle()

	if a := tg.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Start {
		t.Errorf("first press: expected Start, got %v", a)
	}
	if a := tg.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Stop {
		t.Errorf("second press: expected Stop, got %v", a)
	}
	if a := tg.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Start {
		t.Errorf("third press: expected Start, got %v", a)
	}
}

func TestToggleIgnoresRelease(t *testing.T) {
	tg := NewToggle()

	if a := tg.HandleEvent(hotkey.Event{Type: hotkey.Release}); a != None {
		t.Errorf("release: expected None, got %v", a)
	}
}

func TestToggleReset(t *testing.T) {
	tg := NewToggle()
	tg.HandleEvent(hotkey.Event{Type: hotkey.Press}) // Start
	tg.Reset()

	if a := tg.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Start {
		t.Errorf("after reset: expected Start, got %v", a)
	}
}

// --- VAD mode tests ---

func TestVADStartOnPress(t *testing.T) {
	v := NewVAD(0.02, 800)

	if a := v.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Start {
		t.Errorf("press: expected Start, got %v", a)
	}
}

func TestVADManualStop(t *testing.T) {
	v := NewVAD(0.02, 800)
	v.HandleEvent(hotkey.Event{Type: hotkey.Press})

	if a := v.HandleEvent(hotkey.Event{Type: hotkey.Press}); a != Stop {
		t.Errorf("second press: expected Stop, got %v", a)
	}
}

func TestVADIgnoresReleaseAndAudioWhenNotRecording(t *testing.T) {
	v := NewVAD(0.02, 800)

	if a := v.HandleEvent(hotkey.Event{Type: hotkey.Release}); a != None {
		t.Errorf("release: expected None, got %v", a)
	}
	if a := v.HandleAudio([]int16{0, 0, 0}); a != None {
		t.Errorf("audio when not recording: expected None, got %v", a)
	}
}

func TestVADSilenceDetection(t *testing.T) {
	// Use 0ms silence threshold so it triggers immediately
	v := NewVAD(0.5, 0)
	v.HandleEvent(hotkey.Event{Type: hotkey.Press})

	// First call establishes silence start
	silent := []int16{0, 0, 0, 0}
	v.HandleAudio(silent)

	// Second call checks elapsed time (>= 0ms)
	if a := v.HandleAudio(silent); a != Stop {
		t.Errorf("expected Stop on silence, got %v", a)
	}
}

func TestVADLoudAudioResetsTimer(t *testing.T) {
	v := NewVAD(0.01, 0)
	v.HandleEvent(hotkey.Event{Type: hotkey.Press})

	// Silent
	v.HandleAudio([]int16{0, 0, 0})

	// Loud (resets silence timer)
	loud := []int16{20000, -20000, 15000, -15000}
	if a := v.HandleAudio(loud); a != None {
		t.Errorf("loud audio: expected None, got %v", a)
	}
}

func TestVADReset(t *testing.T) {
	v := NewVAD(0.02, 800)
	v.HandleEvent(hotkey.Event{Type: hotkey.Press})
	v.Reset()

	if a := v.HandleAudio([]int16{0, 0}); a != None {
		t.Errorf("after reset: expected None (not recording), got %v", a)
	}
}

// --- RMS tests ---

func TestRMSEmpty(t *testing.T) {
	if rms := RMS(nil); rms != 0 {
		t.Errorf("expected 0, got %f", rms)
	}
}

func TestRMSSilence(t *testing.T) {
	if rms := RMS([]int16{0, 0, 0, 0}); rms != 0 {
		t.Errorf("expected 0, got %f", rms)
	}
}

func TestRMSFullScale(t *testing.T) {
	// All samples at max amplitude
	samples := []int16{32767, 32767, 32767, 32767}
	rms := RMS(samples)
	// Should be very close to 1.0
	if math.Abs(rms-1.0) > 0.001 {
		t.Errorf("expected ~1.0, got %f", rms)
	}
}

func TestRMSMixed(t *testing.T) {
	samples := []int16{1000, -1000, 1000, -1000}
	rms := RMS(samples)
	expected := 1000.0 / 32768.0
	if math.Abs(rms-expected) > 0.001 {
		t.Errorf("expected ~%f, got %f", expected, rms)
	}
}
