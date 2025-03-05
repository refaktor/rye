//go:build !wasm
// +build !wasm

package util

import (
	"fmt"
	"github.com/refaktor/keyboard"

)


func BeforeExit() {
	fmt.Println("Closing keyboard in Exit")
	keyboard.Close()
}
