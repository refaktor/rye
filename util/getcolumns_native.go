//go:build !wasm
// +build !wasm

package util

import (
	tsize "github.com/kopoli/go-terminal-size"
)

func GetTerminalColumns() int {
	var s tsize.Size

	s, err := tsize.GetSize()
	if err == nil {
		return s.Width
	}
	return 50
}
