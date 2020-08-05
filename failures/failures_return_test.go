package failures

import (
	"rye/env"
	"rye/evaldo"
	"rye/loader"
	"fmt"

	//	"fmt"
	"testing"
)

func TestFailures_return1(t *testing.T) {
	input := "{ 1 return 2 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 2) {
		t.Error("Expected result value 2")
	}
}

func TestFailures_return2(t *testing.T) {
	input := "{ return 22 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 22) {
		t.Error("Expected result value 22")
	}
}
func TestFailures_return_opword(t *testing.T) {
	input := "{ 11 22 .return 33 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 22) {
		t.Error("Expected result value 22")
	}
}

func TestFailures_return_pipeword(t *testing.T) {
	input := "{ 11 22 |return 33 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 22) {
		t.Error("Expected result value 22")
	}
}

func TestFailures_return_opword2(t *testing.T) {
	input := "{ a: add 11 22 .return 33 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 22) {
		t.Error("Expected result value 22")
	}
}
func TestFailures_return_pipeword2(t *testing.T) {
	input := "{ a: add 11 22 |return 44 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 33) {
		t.Error("Expected result value 33")
	}
}
func TestFailures_return_infn(t *testing.T) {
	input := "{ f1: fn { } { a: add 11 22 |return 44 } 11 f1 55 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 55) {
		t.Error("Expected result value 55")
	}
}
func TestFailures_return_infn2(t *testing.T) {
	input := "{ f1: fn { } { a: add 11 22 |return 44 } 11 f1 |add 55 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 88) {
		t.Error("Expected result value 88")
	}
}
func TEST_FAILS_FIX_LATER_TestFailures_return_infn3(t *testing.T) {
	// FAILURE: pipeword return here acts like a opword
	//	input := "{ f1: fn { } { a: add 11 22 |return 44 } f2: fn { a } { a .return 1000 } 11 f1 |add f2 55 |return 999 }"
	input := "{ f1: fn { } { a: add 11 22 |return 44 } f2: fn { a } { a .return 1000 } 11 f1 |add f2 3 |return }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 88) {
		t.Error("Expected result value 88")
	}
}

// is this specific to return or any pipeword? make a adhoc test for it
func TestFailures_return_infn3_other_builtin(t *testing.T) {
	// This above is specific to return ... add works
	//input := "{ f1: fn { } { a: add 11 22 |return 44 } f2: fn { a } { a .return 1000 } 11 aa: f1 |add f2 3 return aa }" // works
	//input := "{ f1: fn { } { a: add 11 22 |return 44 } f2: fn { a } { a .return 1000 } 11 return f1 |add f2 3 }" // doesn't work
	//input := "{ f1: fn { } { a: add 11 22 |return 44 } f2: fn { a } { a .return 1000 } 11 f1 |add f2 3 |return }" // doesn't work
	// WE SHOULD FIND MINIMAL EXAMPLE ... does it work with builtin
	input := "{ return 10 |add 4 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 10) {
		t.Error("Expected result value 10")
	}
}

// after adding curry: 1.75s, 2.9s, 19s
