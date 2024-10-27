//go:build wasm
// +build wasm

package term

var ColumnNum int

func GetTerminalColumns() int {
	if ColumnNum == 0 {
		return 60
	} else {
		return ColumnNum
	}
}

func SetTerminalColumns(c int) {
	ColumnNum = c
}
