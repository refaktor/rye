package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

func CompileWord(block *env.Block, ps *env.ProgramState, word env.Word, eyrBlock *env.Block) {
	// LOCAL FIRST
	found, object, _ := findWordValue(ps, word)
	pos := ps.Ser.GetPos()
	if found {
		switch obj := object.(type) {
		case env.Integer:
			eyrBlock.Series.Append(obj)
		case env.Builtin:
			for i := 0; i < obj.Argsn; i++ {
				// fmt.Println("**")
				block = CompileStepRyeToEyr(block, ps, eyrBlock)
			}
			eyrBlock.Series.Append(word)
		}
	} else {
		ps.ErrorFlag = true
		if !ps.FailureFlag {
			ps.Ser.SetPos(pos)
			ps.Res = env.NewError2(5, "word not found: "+word.Print(*ps.Idx))
		}
	}
}

func CompileRyeToEyr(block *env.Block, ps *env.ProgramState, eyrBlock *env.Block) *env.Block {
	for block.Series.Pos() < block.Series.Len() {
		block = CompileStepRyeToEyr(block, ps, eyrBlock)
	}
	return block
}

func CompileStepRyeToEyr(block *env.Block, ps *env.ProgramState, eyrBlock *env.Block) *env.Block {
	// for block.Series.Pos() < block.Series.Len() {
	switch xx := block.Series.Pop().(type) {
	case env.Word:
		// 	fmt.Println("W")
		CompileWord(block, ps, xx, eyrBlock)
		// get value of word
		// if function
		// get argnum
		// add argnum args to mstack (values, words or compiled expressions (recur))
		// add word to mstack
		// else add word to value list
	case env.Opword:
		fmt.Println("O")
	case env.Pipeword:
		fmt.Println("P")
	case env.Integer:
		// fmt.Println("I")
		eyrBlock.Series.Append(xx)
	}
	// }
	return block
}
