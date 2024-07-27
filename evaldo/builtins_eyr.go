// builtins.go
package evaldo

import (
	"fmt"
	"os"

	"github.com/refaktor/rye/env"
)

// definiraj frame <builtin nargs arg0 arg1>
// definiraj stack []evalframe
// callbui kreira trenuten frame, nastavi bui nargs in vrne
// while loop pogleda naslednji arg, če je literal nastavi arg in poveča argc če je argc nargs potem pokliče frame in iz stacka potegne naslednjega, če ni potem zalopa
// 									če je builtin potem pusha trenuten frame na stack in kreira novega

const STACK_SIZE int = 1000

type EyrStack struct {
	D []env.Object
	I int
}

func NewEyrStack() *EyrStack {
	st := EyrStack{}
	st.D = make([]env.Object, STACK_SIZE)
	st.I = 0
	return &st
}

// IsEmpty checks if our stack is empty.
func (s *EyrStack) IsEmpty() bool {
	return s.I == 0
}

// Push adds a new number to the stack
func (s *EyrStack) Push(x env.Object) {
	//// *s = append(*s, x)
	if s.I+1 >= STACK_SIZE {
		fmt.Printf("stack overflow\n")
		os.Exit(0)
	}
	s.D[s.I] = x
	s.I++
	// appending takes a lot of time .. pushing values ...
}

// Pop removes and returns the top element of stack.
func (s *EyrStack) Pop() env.Object {
	if s.IsEmpty() {
		fmt.Printf("stack underflow\n")
		os.Exit(0)
	}
	s.I--
	x := s.D[s.I]
	return x
}

func Eyr_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, stack *EyrStack) *env.ProgramState {
	arg0 := bi.Cur0     //env.Object(bi.Cur0)
	var arg1 env.Object // := bi.Cur1
	var arg2 env.Object

	if bi.Argsn > 0 && bi.Cur0 == nil {
		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		arg0 = stack.Pop()
		if bi.Argsn == 1 {
			ps.Res = bi.Fn(ps, arg0, nil, nil, nil, nil)
			// stack.Push(ps.Res)
		}
	}
	if bi.Argsn > 1 && bi.Cur1 == nil {
		if checkFlagsBi(bi, ps, 1) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}

		arg1 = stack.Pop()
		if bi.Argsn == 2 {
			ps.Res = bi.Fn(ps, arg1, arg0, nil, nil, nil)
			// stack.Push(ps.Res)
		}
	}
	if bi.Argsn > 2 && bi.Cur2 == nil {
		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}

		arg2 = stack.Pop()
		if bi.Argsn == 3 {
			ps.Res = bi.Fn(ps, arg2, arg1, arg0, nil, nil)
			//stack.Push(ps.Res)
		}
	}
	return ps
}

func Eyr_EvalObject(es *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, session *env.RyeCtx, stack *EyrStack, bakein bool) *env.ProgramState {
	//fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.BuiltinType:
		bu := object.(env.Builtin)
		if bakein {
			es.Ser.Put(bu)
		} //es.Ser.SetPos(es.Ser.Pos() - 1)
		if checkFlagsBi(bu, es, 333) {
			return es
		}
		return Eyr_CallBuiltin(bu, es, leftVal, toLeft, stack)
	default:
		es.Res = object
		return es
	}
}

func Eyr_EvalWord(es *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, stack *EyrStack) *env.ProgramState {
	// LOCAL FIRST
	found, object, ctx := findWordValue(es, word)
	if found {
		es = Eyr_EvalObject(es, object, leftVal, toLeft, ctx, stack, true) //ww0128a *
		stack.Push(es.Res)
		return es
	} else {
		es.ErrorFlag = true
		if !es.FailureFlag {
			es.Res = *env.NewError2(5, "Word not found: "+word.Inspect(*es.Idx))
		}
		return es
	}
}

func Eyr_EvalLSetword(ps *env.ProgramState, word env.LSetword, leftVal env.Object, toLeft bool, stack *EyrStack) *env.ProgramState {
	idx := word.Index
	val := stack.Pop()
	ps.Ctx.Mod(idx, val)
	return ps
}

func Eyr_EvalExpression(es *env.ProgramState, stack *EyrStack) *env.ProgramState {
	object := es.Ser.Pop()
	trace2("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType:
			stack.Push(object)
		case env.DecimalType:
			stack.Push(object)
		case env.StringType:
			stack.Push(object)
		case env.BlockType:
			stack.Push(object)
		case env.WordType:
			rr := Eyr_EvalWord(es, object.(env.Word), nil, false, stack)
			return rr
		case env.OpwordType: // + and orther operators are basically opwords too
			rr := Eyr_EvalWord(es, object.(env.Opword), nil, false, stack)
			return rr
		case env.LSetwordType:
			rr := Eyr_EvalLSetword(es, object.(env.LSetword), nil, false, stack)
			return rr
		case env.BuiltinType:
			return Eyr_EvalObject(es, object, nil, false, nil, stack, false) //ww0128a *
		default:
			es.ErrorFlag = true
			es.Res = env.NewError("Not known type for Eyr")
		}
	} else {
		es.ErrorFlag = true
		es.Res = env.NewError("Not known type (nil)")
	}

	return es
}

func Eyr_EvalBlock(es *env.ProgramState, stack *EyrStack, full bool) *env.ProgramState {
	for es.Ser.Pos() < es.Ser.Len() {
		es = Eyr_EvalExpression(es, stack)
		if checkFlagsAfterBlock(es, 101) {
			return es
		}
		if es.ReturnFlag || es.ErrorFlag {
			return es
		}
	}
	if full {
		es.Res = *env.NewBlock(*env.NewTSeries(stack.D[0:stack.I]))
	} else {
		es.Res = stack.Pop()
	}
	return es
}

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

var Builtins_eyr = map[string]*env.Builtin{

	"eyr": {
		Argsn: 1,
		Doc:   "Evaluates Rye block as Eyr (postfix) stack based code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				stack := NewEyrStack()
				ser := ps.Ser
				ps.Ser = bloc.Series
				Eyr_EvalBlock(ps, stack, false)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "eyr")
			}
		},
	},

	"to-eyr": {
		Argsn: 1,
		Doc:   "Evaluates Rye block as Eyr (postfix) stack based code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				eBlock := env.NewBlock(*env.NewTSeries(make([]env.Object, 0)))
				CompileRyeToEyr(&bloc, ps, eBlock)
				return *eBlock
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "eyr")
			}
		},
	},

	"eyr\\full": {
		Argsn: 1,
		Doc:   "Evaluates Rye block as Eyr (postfix) stack based code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				stack := NewEyrStack()
				ser := ps.Ser
				ps.Ser = bloc.Series
				Eyr_EvalBlock(ps, stack, true)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "eyr")
			}
		},
	},

	"eyr\\loop": {
		Argsn: 2,
		Doc:   "Evaluates Rye block in loop as Eyr code (postfix stack based) N times.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					stack := NewEyrStack()
					for i := 0; int64(i) < cond.Value; i++ {
						ps = Eyr_EvalBlock(ps, stack, false)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "eyr-loop")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "eyr-loop")
			}
		},
	},
}
