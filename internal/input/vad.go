package input

import (
	"math"
	"time"

	"github.com/willfindlay/edict/internal/hotkey"
)

// VAD implements voice-activity-detection mode: press starts recording,
// recording stops automatically when silence is detected for the configured
// duration.
type VAD struct {
	threshold  float64
	silenceMs  int
	recording  bool
	silentFrom time.Time
	inSilence  bool
}

func NewVAD(threshold float64, silenceMs int) *VAD {
	return &VAD{
		threshold: threshold,
		silenceMs: silenceMs,
	}
}

func (v *VAD) HandleEvent(ev hotkey.Event) Action {
	if ev.Type != hotkey.Press {
		return None
	}

	if v.recording {
		// Manual stop override
		v.recording = false
		v.inSilence = false
		return Stop
	}
	v.recording = true
	v.inSilence = false
	return Start
}

func (v *VAD) HandleAudio(samples []int16) Action {
	if !v.recording {
		return None
	}

	rms := RMS(samples)

	if rms < v.threshold {
		if !v.inSilence {
			v.inSilence = true
			v.silentFrom = time.Now()
		} else if time.Since(v.silentFrom).Milliseconds() >= int64(v.silenceMs) {
			v.recording = false
			v.inSilence = false
			return Stop
		}
	} else {
		v.inSilence = false
	}

	return None
}

func (v *VAD) Reset() {
	v.recording = false
	v.inSilence = false
}

// RMS computes the root-mean-square of PCM int16 samples, normalized to [0, 1].
func RMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64
	for _, s := range samples {
		f := float64(s) / 32768.0
		sum += f * f
	}
	return math.Sqrt(sum / float64(len(samples)))
}
