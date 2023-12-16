//go:build wasm
// +build wasm

package term

import (
	"github.com/refaktor/rye/env"
)

// DisplayBlock is dummy non-implementation for wasm for display builtin
func DisplayBlock(bloc env.Block, idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

// DisplayDict is dummy non-implementation for wasm for display builtin
func DisplayDict(bloc env.Dict, idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

// DisplaySpreadsheetRow is dummy non-implementation for wasm for display builtin
func DisplaySpreadsheetRow(bloc env.SpreadsheetRow, idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

// DisplayTable is dummy non-implementation for wasm for display builtin
func DisplayTable(bloc env.Spreadsheet, idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

// SaveCurPos is dummy non-implementation for wasm for display builtin
func SaveCurPos() {
	panic("unimplemented")
}
