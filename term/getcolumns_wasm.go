//go:build wasm
// +build wasm

package term

func GetTerminalColumns() int {
	return 50
}
