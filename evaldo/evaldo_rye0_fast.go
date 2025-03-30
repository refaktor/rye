package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

// Instruction is a function that executes a single instruction in the VM
type Instruction func(vm *Rye0VM) int

// Program represents a compiled Rye0 program
type Program struct {
	code []Instruction
}

// Rye0VM represents a virtual machine for executing compiled Rye0 code
type Rye0VM struct {
	ctx   *env.RyeCtx
	pctx  *env.RyeCtx
	gen   *env.Gen
	idx   *env.Idxs
	stack []env.Object
	Sp    int // Exported for use in benchmarks
	err   error
}

// NewRye0VM creates a new virtual machine for executing compiled Rye0 code
func NewRye0VM(ctx *env.RyeCtx, pctx *env.RyeCtx, gen *env.Gen, idx *env.Idxs) *Rye0VM {
	return &Rye0VM{
		ctx:   ctx,
		pctx:  pctx,
		gen:   gen,
		idx:   idx,
		stack: make([]env.Object, 1024), // Initial stack size
		Sp:    0,
		err:   nil,
	}
}

// Push pushes a value onto the stack
func (vm *Rye0VM) Push(val env.Object) {
	if vm.Sp >= len(vm.stack) {
		// Grow stack if needed
		newStack := make([]env.Object, len(vm.stack)*2)
		copy(newStack, vm.stack)
		vm.stack = newStack
	}
	vm.stack[vm.Sp] = val
	vm.Sp++
}

// Pop pops a value from the stack
func (vm *Rye0VM) Pop() env.Object {
	if vm.Sp <= 0 {
		vm.err = fmt.Errorf("stack underflow")
		return nil
	}
	vm.Sp--
	return vm.stack[vm.Sp]
}

// Execute executes a compiled program
func (vm *Rye0VM) Execute(p *Program) (env.Object, error) {
	code := p.code
	ip := 0

	// Reset VM state
	vm.Sp = 0
	vm.err = nil

	for ip < len(code) {
		ip += code[ip](vm)
		if vm.err != nil {
			return nil, vm.err
		}
	}

	if vm.Sp == 0 {
		return nil, nil
	}

	return vm.stack[vm.Sp-1], nil
}
