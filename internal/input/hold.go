package input

import (
	"github.com/willfindlay/edict/internal/hotkey"
)

// Hold implements hold-to-talk: press starts, release stops.
type Hold struct {
	recording bool
}

func NewHold() *Hold {
	return &Hold{}
}

func (h *Hold) HandleEvent(ev hotkey.Event) Action {
	switch ev.Type {
	case hotkey.Press:
		if !h.recording {
			h.recording = true
			return Start
		}
	case hotkey.Release:
		if h.recording {
			h.recording = false
			return Stop
		}
	}
	return None
}

func (h *Hold) HandleAudio(_ []int16) Action {
	return None
}

func (h *Hold) Reset() {
	h.recording = false
}
