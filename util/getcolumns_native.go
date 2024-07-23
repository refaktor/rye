//go:build !wasm
// +build !wasm

package util

import (
	"syscall"
	"unsafe"
)

type winSize struct {
	row, col       uint16
	xpixel, ypixel uint16
}

func GetTerminalColumns() int {
	var ws winSize
	// TODO -- MOVE THIS OUTSIDE ... in case of native use this, in case of browser check how we can get the numer of columns
	ok, _, _ := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout),
		syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	if int(ok) < 0 {
		return 50
	}
	columns := int(ws.col)
	return columns
}
