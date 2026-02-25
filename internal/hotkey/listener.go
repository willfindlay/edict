package hotkey

import (
	hook "github.com/robotn/gohook"
)

// EventType represents a hotkey press or release.
type EventType int

const (
	Press EventType = iota
	Release
)

// Event is emitted when the configured hotkey is pressed or released.
type Event struct {
	Type EventType
}

// Listener watches for global keyboard events using gohook.
type Listener struct {
	keycode   uint16
	modifiers [2]uint16
	events    chan Event
	done      chan struct{}
}

// NewListener creates a hotkey listener for the given modifier+key combination.
func NewListener(modifier, key string) *Listener {
	return &Listener{
		keycode:   KeyCodes[key],
		modifiers: ModifierCodes[modifier],
		events:    make(chan Event, 16),
		done:      make(chan struct{}),
	}
}

// Events returns the channel of hotkey events.
func (l *Listener) Events() <-chan Event {
	return l.events
}

// Start begins listening for keyboard events in a blocking manner.
// Call this from a dedicated goroutine.
func (l *Listener) Start() {
	evChan := hook.Start()
	defer hook.End()

	modDown := false

	for {
		select {
		case <-l.done:
			return
		case ev, ok := <-evChan:
			if !ok {
				return
			}
			l.handleEvent(ev, &modDown)
		}
	}
}

// Stop terminates the listener.
func (l *Listener) Stop() {
	select {
	case <-l.done:
	default:
		close(l.done)
	}
}

func (l *Listener) isModifier(rawcode uint16) bool {
	return rawcode == l.modifiers[0] || rawcode == l.modifiers[1]
}

func (l *Listener) handleEvent(ev hook.Event, modDown *bool) {
	switch ev.Kind {
	case hook.KeyDown:
		if l.isModifier(ev.Rawcode) {
			*modDown = true
		}
		if ev.Rawcode == l.keycode && *modDown {
			select {
			case l.events <- Event{Type: Press}:
			default:
			}
		}
	case hook.KeyUp:
		if l.isModifier(ev.Rawcode) {
			*modDown = false
		}
		if ev.Rawcode == l.keycode {
			select {
			case l.events <- Event{Type: Release}:
			default:
			}
		}
	}
}
