package evaldo

import (
	"rye/env"
	//"rye/loader"
)

func registerGeneric(ps *env.ProgramState, kind int, word int, object env.Object) {
	// indexWord
	//idxs := loader.GetIdxs()
	//idx := idxs.IndexWord(word)
	// set global word with builtin
	ps.Gen.Set(kind, word, object)
}
