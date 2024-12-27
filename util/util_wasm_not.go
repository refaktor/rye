//go:build !wasm
// +build !wasm

package util

import (
	"fmt"
	"github.com/cszczepaniak/keyboard"

)


func BeforeExit() {
	fmt.Println("Closing keyboard in Exit")
	keyboard.Close()
}
