//go:build windows

package output

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

const (
	inputKeyboard    = 1
	keyeventfUnicode = 0x0004
	keyeventfKeyup   = 0x0002
)

// kbdInput matches the KEYBDINPUT struct.
type kbdInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

// inputRecord matches the INPUT struct with type INPUT_KEYBOARD.
type inputRecord struct {
	inputType uint32
	ki        kbdInput
	_padding  [8]byte // union padding to match C struct size
}

// Backend identifies the typing simulation tool on Windows.
type Backend string

const (
	SendInputBackend Backend = "sendinput"
)

// windowsTyper simulates keyboard input via Windows SendInput API.
type windowsTyper struct {
	delayUs int
}

// NewTyper creates a typer using the Windows SendInput API.
// The backend parameter is accepted for API compatibility but only "sendinput" is supported.
func NewTyper(_ Backend, delayUs int) Typer {
	return &windowsTyper{delayUs: delayUs}
}

func (t *windowsTyper) CheckAvailable() error {
	if err := procSendInput.Find(); err != nil {
		return fmt.Errorf("SendInput not available: %w", err)
	}
	return nil
}

func (t *windowsTyper) Type(text string) error {
	if text == "" {
		return nil
	}

	delay := time.Duration(t.delayUs) * time.Microsecond

	for _, r := range text {
		if err := t.sendChar(r); err != nil {
			return fmt.Errorf("SendInput failed for %q: %w", string(r), err)
		}
		if delay > 0 {
			time.Sleep(delay)
		}
	}
	return nil
}

func (t *windowsTyper) sendChar(r rune) error {
	// KEYEVENTF_UNICODE sends characters by Unicode codepoint directly,
	// avoiding VK code translation. Handles all characters correctly.
	down := inputRecord{
		inputType: inputKeyboard,
		ki: kbdInput{
			wScan:   uint16(r),
			dwFlags: keyeventfUnicode,
		},
	}
	up := inputRecord{
		inputType: inputKeyboard,
		ki: kbdInput{
			wScan:   uint16(r),
			dwFlags: keyeventfUnicode | keyeventfKeyup,
		},
	}

	inputs := [2]inputRecord{down, up}
	ret, _, err := procSendInput.Call(
		2,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
	if ret == 0 {
		return fmt.Errorf("SendInput returned 0: %v", err)
	}
	return nil
}
