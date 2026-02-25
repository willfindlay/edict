package input

import (
	"github.com/willfindlay/edict/internal/hotkey"
)

// Action represents what the pipeline should do in response to an input event.
type Action int

const (
	None  Action = iota
	Start        // Begin recording
	Stop         // Stop recording and transcribe
)

// Mode processes hotkey events and audio samples to determine recording actions.
type Mode interface {
	// HandleEvent processes a hotkey press/release and returns the resulting action.
	HandleEvent(ev hotkey.Event) Action

	// HandleAudio processes audio samples (used by VAD mode). Returns Stop when
	// silence is detected. Other modes return None.
	HandleAudio(samples []int16) Action

	// Reset returns the mode to its initial state.
	Reset()
}
