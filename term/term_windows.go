//go:build windows
// +build windows

package term

import (
	"syscall"
	"unsafe"
)

var (
	kernel32                          = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle                  = kernel32.NewProc("GetStdHandle")
	procReadConsoleInput              = kernel32.NewProc("ReadConsoleInputW")
	procGetNumberOfConsoleInputEvents = kernel32.NewProc("GetNumberOfConsoleInputEvents")
	procSetConsoleMode                = kernel32.NewProc("SetConsoleMode")
	procGetConsoleMode                = kernel32.NewProc("GetConsoleMode")
)

const (
	STD_INPUT_HANDLE = -10

	// Event types
	KEY_EVENT                = 0x0001
	MOUSE_EVENT              = 0x0002
	WINDOW_BUFFER_SIZE_EVENT = 0x0004
	MENU_EVENT               = 0x0008
	FOCUS_EVENT              = 0x0010

	// Control key states
	RIGHT_ALT_PRESSED  = 0x0001
	LEFT_ALT_PRESSED   = 0x0002
	RIGHT_CTRL_PRESSED = 0x0004
	LEFT_CTRL_PRESSED  = 0x0008
	SHIFT_PRESSED      = 0x0010

	// Console modes
	ENABLE_PROCESSED_INPUT = 0x0001
	ENABLE_LINE_INPUT      = 0x0002
	ENABLE_ECHO_INPUT      = 0x0004
	ENABLE_WINDOW_INPUT    = 0x0008
	ENABLE_MOUSE_INPUT     = 0x0010
	ENABLE_INSERT_MODE     = 0x0020
	ENABLE_QUICK_EDIT_MODE = 0x0040
	ENABLE_EXTENDED_FLAGS  = 0x0080
	ENABLE_AUTO_POSITION   = 0x0100
)

type InputRecord struct {
	EventType uint16
	_         uint16 // padding
	Event     [16]byte
}

type KeyEventRecord struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

// GetChar reads a character from the Windows console and returns ASCII code, key code, and error
func GetChar() (ascii int, keyCode int, err error) {
	// For Windows API constants, use int32 to preserve the negative value
	handle, _, err := procGetStdHandle.Call(uintptr(^uintptr(0) - 9))
	if handle == 0 {
		return 0, 0, err
	}

	// Save the current console mode
	var oldMode uint32
	ret, _, err := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&oldMode)))
	if ret == 0 {
		return 0, 0, err
	}

	// Set raw input mode (disable line input, echo, etc.)
	newMode := oldMode &^ (ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT | ENABLE_PROCESSED_INPUT)
	ret, _, err = procSetConsoleMode.Call(handle, uintptr(newMode))
	if ret == 0 {
		return 0, 0, err
	}

	// Restore console mode when we're done
	defer func() {
		procSetConsoleMode.Call(handle, uintptr(oldMode))
	}()

	// Wait for and read a single input event
	var numEvents uint32
	var record InputRecord
	for {
		ret, _, err = procGetNumberOfConsoleInputEvents.Call(handle, uintptr(unsafe.Pointer(&numEvents)))
		if ret == 0 {
			return 0, 0, err
		}

		if numEvents > 0 {
			var numRead uint32
			ret, _, err = procReadConsoleInput.Call(
				handle,
				uintptr(unsafe.Pointer(&record)),
				1,
				uintptr(unsafe.Pointer(&numRead)),
			)
			if ret == 0 {
				return 0, 0, err
			}

			// Process only key events where a key is pressed (not released)
			if record.EventType == KEY_EVENT {
				// Convert the event bytes to a KeyEventRecord
				keyEvent := (*KeyEventRecord)(unsafe.Pointer(&record.Event[0]))

				// Only process key down events
				if keyEvent.KeyDown != 0 {
					// Check for special keys
					switch keyEvent.VirtualKeyCode {
					case 0x26: // Up arrow
						return 0, 38, nil
					case 0x28: // Down arrow
						return 0, 40, nil
					case 0x25: // Left arrow
						return 0, 37, nil
					case 0x27: // Right arrow
						return 0, 39, nil
					case 0x1B: // Escape
						return 27, 0, nil
					case 0x0D: // Enter
						return 13, 0, nil
					case 0x08: // Backspace
						// This is the key part - ensure backspace is always ASCII 8
						// regardless of control key state
						return 8, 0, nil
					default:
						// For regular characters, return the Unicode character
						if keyEvent.UnicodeChar != 0 {
							// Check for Ctrl key combinations
							if (keyEvent.ControlKeyState & (LEFT_CTRL_PRESSED | RIGHT_CTRL_PRESSED)) != 0 {
								// For Ctrl+C
								if keyEvent.UnicodeChar == 3 {
									return 3, 0, nil
								}
								// For other Ctrl combinations, we could handle them here
								// But we'll let the UnicodeChar pass through
							}
							return int(keyEvent.UnicodeChar), 0, nil
						}
					}
				}
			}
		}
	}
}

// GetChar2 is similar to GetChar but returns the character as a string
func GetChar2() (letter string, ascii int, keyCode int, err error) {
	a, k, e := GetChar()
	if e != nil {
		return "", 0, 0, e
	}

	if a != 0 {
		letter = string(rune(a))
	}

	return letter, a, k, nil
}
