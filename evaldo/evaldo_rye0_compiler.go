package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

// Rye0_CompileBlock compiles a block of code into a Program
func Rye0_CompileBlock(ps *env.ProgramState) *Program {
	// Create a new program
	program := &Program{
		code: make([]Instruction, 0),
	}

	// Save the original position
	origPos := ps.Ser.GetPos()

	// Compile each expression in the block
	for ps.Ser.Pos() < ps.Ser.Len() {
		instr := Rye0_CompileExpression(ps)
		program.code = append(program.code, instr)
	}

	// Reset the position
	ps.Ser.SetPos(origPos)

	return program
}

// Rye0_CompileExpression compiles a single expression into an Instruction
func Rye0_CompileExpression(ps *env.ProgramState) Instruction {
	object := ps.Ser.Pop()

	if object == nil {
		return func(vm *Rye0VM) int {
			vm.err = fmt.Errorf("expected Rye value but it's missing")
			return 1
		}
	}

	switch object.Type() {
	// Literal values evaluate to themselves
	case env.IntegerType:
		val := object.(env.Integer)
		return func(vm *Rye0VM) int {
			vm.Push(val)
			return 1
		}
	case env.DecimalType:
		val := object.(env.Decimal)
		return func(vm *Rye0VM) int {
			vm.Push(val)
			return 1
		}
	case env.StringType:
		val := object.(env.String)
		return func(vm *Rye0VM) int {
			vm.Push(val)
			return 1
		}
	case env.VoidType, env.UriType, env.EmailType:
		return func(vm *Rye0VM) int {
			vm.Push(object)
			return 1
		}

	// Block handling
	case env.BlockType:
		block := object.(env.Block)
		return func(vm *Rye0VM) int {
			vm.Push(block)
			return 1
		}

	// Word types
	case env.TagwordType:
		word := object.(env.Tagword)
		return func(vm *Rye0VM) int {
			vm.Push(*env.NewWord(word.Index))
			return 1
		}
	case env.WordType:
		word := object.(env.Word)
		return func(vm *Rye0VM) int {
			// Look up the word in the context
			val, found := vm.ctx.Get(word.Index)
			if !found && vm.ctx.Parent != nil {
				val, found = vm.ctx.Parent.Get(word.Index)
			}

			if !found {
				vm.err = fmt.Errorf("word not found: %s", vm.idx.GetWord(word.Index))
				return 1
			}

			// Handle different types of values
			switch val.Type() {
			case env.FunctionType:
				// Call the function
				// For simplicity, we're not handling function calls with arguments here
				// In a real implementation, you'd need to compile and execute the function body
				vm.err = fmt.Errorf("function calls not implemented in fast evaluator yet")
				return 1
			case env.BuiltinType:
				// Call the builtin
				builtin := val.(env.Builtin)

				// Check if we have enough arguments on the stack
				if builtin.Argsn > 0 && vm.Sp < builtin.Argsn {
					vm.err = fmt.Errorf("not enough arguments for builtin: %s", vm.idx.GetWord(word.Index))
					return 1
				}

				// Get arguments from the stack
				var args [5]env.Object
				for i := 0; i < builtin.Argsn; i++ {
					args[builtin.Argsn-i-1] = vm.Pop()
				}

				// Call the builtin function
				result := builtin.Fn(nil, args[0], args[1], args[2], args[3], args[4])
				vm.Push(result)
				return 1
			default:
				// Just push the value
				vm.Push(val)
				return 1
			}
		}
	case env.SetwordType:
		return func(vm *Rye0VM) int {
			// Assume the next instruction will push a value onto the stack
			// We'll set the word to that value
			return 1
		}
	case env.ModwordType:
		return func(vm *Rye0VM) int {
			// Assume the next instruction will push a value onto the stack
			// We'll modify the word with that value
			return 1
		}

	// Error handling
	case env.CommaType:
		return func(vm *Rye0VM) int {
			vm.err = fmt.Errorf("expression guard inside expression")
			return 1
		}
	case env.ErrorType:
		return func(vm *Rye0VM) int {
			vm.err = fmt.Errorf("error object encountered")
			return 1
		}

	// Operator word
	case env.OpwordType:
		opword := object.(env.Opword)
		return func(vm *Rye0VM) int {
			// Look up the operator in the context
			val, found := vm.ctx.Get(opword.Index)
			if !found && vm.ctx.Parent != nil {
				val, found = vm.ctx.Parent.Get(opword.Index)
			}

			if !found {
				vm.err = fmt.Errorf("operator not found: %s", vm.idx.GetWord(opword.Index))
				return 1
			}

			// Handle different types of values
			switch val.Type() {
			case env.BuiltinType:
				// Call the builtin
				builtin := val.(env.Builtin)
				// For simplicity, we're not handling builtin calls with arguments here
				result := builtin.Fn(nil, nil, nil, nil, nil, nil)
				vm.Push(result)
				return 1
			default:
				vm.Push(val)
				return 1
			}
		}

	// Unknown type
	default:
		return func(vm *Rye0VM) int {
			vm.err = fmt.Errorf("unknown Rye value type: %d", object.Type())
			return 1
		}
	}
}

// Rye0_FastEvalBlock evaluates a block of code using the fast evaluator
func Rye0_FastEvalBlock(ps *env.ProgramState) *env.ProgramState {
	// Compile the block
	program := Rye0_CompileBlock(ps)

	// Create a VM
	vm := NewRye0VM(ps.Ctx, ps.PCtx, ps.Gen, ps.Idx)

	// Execute the program
	result, err := vm.Execute(program)
	if err != nil {
		ps.ErrorFlag = true
		ps.Res = env.NewError(err.Error())
		return ps
	}

	// Set the result
	ps.Res = result
	return ps
}
