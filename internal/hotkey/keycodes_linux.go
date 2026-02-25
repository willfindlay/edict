//go:build linux

package hotkey

// ModifierCodes maps modifier names to Linux input event scancodes.
var ModifierCodes = map[string]uint16{
	"alt":   56,  // KEY_LEFTALT
	"ctrl":  29,  // KEY_LEFTCTRL
	"shift": 42,  // KEY_LEFTSHIFT
	"super": 125, // KEY_LEFTMETA
}

// KeyCodes maps key names to Linux input event scancodes.
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
