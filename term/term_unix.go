//go:build !windows && !wasm
// +build !windows,!wasm

package term

import (
	"github.com/pkg/term"
)

// GetChar reads a character from the Unix terminal and returns ASCII code, key code, and error
func GetChar() (ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")
	if err = term.RawMode(t); err != nil {
		return
	}
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	}
	// else {
	// Two characters read??
	// }
	if err = t.Restore(); err != nil {
		return
	}
	t.Close()
	return
}

// GetChar2 is similar to GetChar but returns the character as a string
func GetChar2() (letter string, ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")
	if err = term.RawMode(t); err != nil {
		return
	}
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 2 && bytes[0] == 27 && bytes[1] == 127 {
		// ESC followed by DEL/Backspace is Alt+Backspace
		// Return DEL (127) to signal Alt+Backspace for word deletion
		ascii = 127
		letter = string(rune(127))
	} else if numRead == 1 {
		ascii = int(bytes[0])
		letter = string(bytes[0])
	} else if numRead == 2 {
		letter = string(bytes[0:2])
	} else if numRead == 3 {
		letter = string(bytes)
	}
	if err = t.Restore(); err != nil {
		return
	}
	t.Close()
	return
}
