//go:build !wasm && !no_baseio
// +build !wasm,!no_baseio

package util

import (
	"fmt"
	"github.com/refaktor/keyboard"
)

func BeforeExit() {
	fmt.Println("Closing keyboard in Exit")
	keyboard.Close()
}
