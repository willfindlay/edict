//go:build windows

package overlay

import (
	"syscall"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	user32Win         = syscall.NewLazyDLL("user32.dll")
	procGetWindowLong = user32Win.NewProc("GetWindowLongW")
	procSetWindowLong = user32Win.NewProc("SetWindowLongW")
	procShowWindow    = user32Win.NewProc("ShowWindow")
)

const (
	gwlExStyle     = ^uintptr(20 - 1) // -20 as uintptr
	wsExToolWindow = 0x00000080
	wsExAppWindow  = 0x00040000
)

// hideFromTaskbar applies WS_EX_TOOLWINDOW to prevent the overlay from
// appearing in the Windows taskbar and alt-tab list.
func hideFromTaskbar() {
	hwnd := rl.GetWindowHandle()
	style, _, _ := procGetWindowLong.Call(uintptr(hwnd), gwlExStyle)
	style = (style | wsExToolWindow) &^ wsExAppWindow
	procSetWindowLong.Call(uintptr(hwnd), gwlExStyle, style)

	// Force Windows to re-read the style by toggling visibility.
	procShowWindow.Call(uintptr(hwnd), 0) // SW_HIDE
	procShowWindow.Call(uintptr(hwnd), 5) // SW_SHOW
}
