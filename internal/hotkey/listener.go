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
	keycode  uint16
	modifier uint16
	events   chan Event
	done     chan struct{}
}

// ModifierCodes maps modifier names to gohook rawcode values.
var ModifierCodes = map[string]uint16{
	"alt":   56,  // KEY_LEFTALT
	"ctrl":  29,  // KEY_LEFTCTRL
	"shift": 42,  // KEY_LEFTSHIFT
	"super": 125, // KEY_LEFTMETA
}

// KeyCodes maps key names to gohook rawcode values.
var KeyCodes = map[string]uint16{
	"a": 30, "b": 48, "c": 46, "d": 32, "e": 18, "f": 33,
	"g": 34, "h": 35, "i": 23, "j": 36, "k": 37, "l": 38,
	"m": 50, "n": 49, "o": 24, "p": 25, "q": 16, "r": 19,
	"s": 31, "t": 20, "u": 22, "v": 47, "w": 17, "x": 45,
	"y": 21, "z": 44,
	"0": 11, "1": 2, "2": 3, "3": 4, "4": 5, "5": 6,
	"6": 7, "7": 8, "8": 9, "9": 10,
	"f1": 59, "f2": 60, "f3": 61, "f4": 62, "f5": 63, "f6": 64,
	"f7": 65, "f8": 66, "f9": 67, "f10": 68, "f11": 87, "f12": 88,
}

// NewListener creates a hotkey listener for the given modifier+key combination.
func NewListener(modifier, key string) *Listener {
	return &Listener{
		keycode:  KeyCodes[key],
		modifier: ModifierCodes[modifier],
		events:   make(chan Event, 16),
		done:     make(chan struct{}),
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

func (l *Listener) handleEvent(ev hook.Event, modDown *bool) {
	switch ev.Kind {
	case hook.KeyDown:
		if ev.Rawcode == l.modifier {
			*modDown = true
		}
		if ev.Rawcode == l.keycode && *modDown {
			select {
			case l.events <- Event{Type: Press}:
			default:
			}
		}
	case hook.KeyUp:
		if ev.Rawcode == l.modifier {
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
