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

type EyrStack struct {
	D []env.Object
	I int
}

func NewEyrStack() *EyrStack {
	st := EyrStack{}
	st.D = make([]env.Object, 100)
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
	s.D[s.I] = x
	s.I++
	// appending takes a lot of time .. pushing values ...
	// try creating stack in advance and then just setting values
	// and see the difference TODO NEXT
}

// Pop removes and returns the top element of stack.
func (s *EyrStack) Pop() env.Object {
	if s.IsEmpty() {
		fmt.Printf("stack underflow\n")
		os.Exit(1)
	}

	/// i := len(*s) - 1
	/// x := (*s)[i]
	/// *s = (*s)[:i]
	s.I--
	x := s.D[s.I]
	return x
}

func Eyr_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, stack *EyrStack) *env.ProgramState {
	arg0 := bi.Cur0 //env.Object(bi.Cur0)
	arg1 := bi.Cur1

	if bi.Argsn > 0 && bi.Cur0 == nil {
		//fmt.Println(" ARG 1 ")
		//fmt.Println(ps.Ser.GetPos())
		// evalExprFn(ps, true)
		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		arg0 = stack.Pop()
	}
	if bi.Argsn > 1 && bi.Cur1 == nil {
		//evalExprFn(ps, true) // <---- THESE DETERMINE IF IT CONSUMES WHOLE EXPRESSION OR NOT IN CASE OF PIPEWORDS .. HM*... MAYBE WOULD COULD HAVE A WORD MODIFIER?? a: 2 |add 5 a:: 2 |add 5 print* --TODO
		if checkFlagsBi(bi, ps, 1) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		//fmt.Println(ps.Res)

		arg1 = stack.Pop()
	}
	ps.Res = bi.Fn(ps, arg1, arg0, nil, nil, nil)

	stack.Push(ps.Res)
	return ps
}

func Eyr_EvalObject(es *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, session *env.RyeCtx, stack *EyrStack, bakein bool) *env.ProgramState {
	//fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.BuiltinType:
		//fmt.Println(" BUIL**")
		//fmt.Println(es.Ser.GetPos())
		//fmt.Println(" BUILTIN**")
		bu := object.(env.Builtin)
		if bakein {
			es.Ser.Put(bu)
		} //es.Ser.SetPos(es.Ser.Pos() - 1)
		//es.Ser.SetPos(es.Ser.Pos() + 1)
		// OBJECT INJECTION EXPERIMENT
		// es.Ser.Put(bu)

		if checkFlagsBi(bu, es, 333) {
			return es
		}
		return Eyr_CallBuiltin(bu, es, leftVal, toLeft, stack)

		//es.Res.Trace("After builtin call")
		//return es
	default:
		//d object.Trace("DEFAULT**")
		es.Res = object
		//es.Res.Trace("After object returned")
		return es
	}
}

func Eyr_EvalWord(es *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, stack *EyrStack) *env.ProgramState {
	// LOCAL FIRST
	found, object, ctx := findWordValue(es, word)
	if found {
		trace("****33")
		return Eyr_EvalObject(es, object, leftVal, toLeft, ctx, stack, true) //ww0128a *
		//es.Res.Trace("After eval Object")
		//return es
	} else {
		trace("****34")
		es.ErrorFlag = true
		if !es.FailureFlag {
			es.Res = *env.NewError2(5, "Word not found: "+word.Inspect(*es.Idx))
		}
		return es
	}
}

func Eyr_EvalExpression(es *env.ProgramState, stack *EyrStack) *env.ProgramState {
	object := es.Ser.Pop()
	//es.Idx.Probe()
	trace2("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType:
			stack.Push(object)
		case env.StringType:
			stack.Push(object)
		case env.BlockType:
			stack.Push(object)
		case env.WordType:
			rr := Eyr_EvalWord(es, object.(env.Word), nil, false, stack)
			return rr
		case env.OpwordType:
			rr := Eyr_EvalWord(es, object.(env.Opword), nil, false, stack)
			return rr
		case env.BuiltinType:
			//fmt.Println("yoyo")
			return Eyr_EvalObject(es, object, nil, false, nil, stack, false) //ww0128a *
			//rr := Eyr_EvalWord(es, object.(env.Word), nil, false, stack)
			//return rr
		default:
			es.ErrorFlag = true
			es.Res = env.NewError("Not known type")
		}
	} else {
		es.ErrorFlag = true
		es.Res = env.NewError("Not known type")
	}

	return es
}

func Eyr_EvalBlock(es *env.ProgramState, stack *EyrStack) *env.ProgramState {
	for es.Ser.Pos() < es.Ser.Len() {
		es = Eyr_EvalExpression(es, stack)
		if checkFlagsAfterBlock(es, 101) {
			return es
		}
		if es.ReturnFlag || es.ErrorFlag {
			return es
		}
	}
	return es
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
				Eyr_EvalBlock(ps, stack)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "eyr")
			}
		},
	},

	"eyr-loop": {
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
						ps = Eyr_EvalBlock(ps, stack)
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
