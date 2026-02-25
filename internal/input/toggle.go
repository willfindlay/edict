package input

import (
	"github.com/willfindlay/edict/internal/hotkey"
)

// Toggle implements toggle mode: first press starts, second press stops.
type Toggle struct {
	recording bool
}

func NewToggle() *Toggle {
	return &Toggle{}
}

func (t *Toggle) HandleEvent(ev hotkey.Event) Action {
	if ev.Type != hotkey.Press {
		return None
	}

	if t.recording {
		t.recording = false
		return Stop
	}
	t.recording = true
	return Start
}

func (t *Toggle) HandleAudio(_ []int16) Action {
	return None
}

func (t *Toggle) Reset() {
	t.recording = false
}
