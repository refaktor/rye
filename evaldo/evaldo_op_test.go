// main_test.go
package evaldo

import (
	"Ryelang/env"
	"Ryelang/loader"

	"fmt"

	"testing"
)

func TestEvaldo_expression_maybeproctest_comma(t *testing.T) {
	input := "{ add add 1 2 3 , add 2 3 , add 2 4 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 6 {
		t.Error("Expected result value 6")
	}
}

func TestEvaldo_OPWORD_1(t *testing.T) {
	input := "{ 2 .add 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 5 {
		t.Error("Expected result value 5")
	}
}

func TestEvaldo_PIPEWORD_1(t *testing.T) {
	input := "{ 2 |add 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 5 {
		t.Error("Expected result value 5")
	}
}

func TestEvaldo_OPWORD_2(t *testing.T) {
	input := "{ 10 |subtract 3 .add 6 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1 {
		t.Error("Expected result value 1")
	}
}

func TestEvaldo_PIPEWORD_2(t *testing.T) {
	input := "{ 10 |subtract 3 |add 5 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 12 {
		t.Error("Expected result value 12")
	}
}

func TestEvaldo_OPWORD_3(t *testing.T) {
	//input := "{ r: 10 |subtract 3 .add 6 |add 10 .subtract 2 |subtract 3 , 2 .add 2 .add rr }"
	input := "{ inspect r: 10 |subtract 3 .add 6 |add 10 , 2 .add r }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 13 {
		t.Error("Expected result value 13")
	}
}

func _TestEvaldo_PIPEWORD_3(t *testing.T) {
	input := "{ 10 |subtract 3 |add 5 |subtract 10 |add 8 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 10 {
		t.Error("Expected result value 10")
	}
}

/*

func TestEvaldo_expression_first_opword_2(t *testing.T) {
	input := "{ 2 |add 3 |add 5 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 10 {
		t.Error("Expected result value 10")
	}
}

func TestEvaldo_expression_first_opword_mix(t *testing.T) {
	input := "{ add add 2 |add 3 |add 5 10 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 21 {
		t.Error("Expected result value 21")
	}
}

func TestEvaldo_expression_first_opword_4(t *testing.T) {
	input := "{ 100 |add 50 |subtract 33 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 117 {
		t.Error("Expected result value 117")
	}
}

func TestEvaldo_expression_first_opword_5(t *testing.T) {
	input := "{ 100 |add 50 |subtract 33 add 2 5 |subtract 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 4 {
		t.Error("Expected result value 4")
	}
}

func TestEvaldo_OPWORDS_USER_FUNC_1(t *testing.T) {
	input := "{ myadd: fn { a b } { add a b } 100 |myadd 50 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 150 {
		t.Error("Expected result value 150")
	}
}

func TestEvaldo_OPWORDS_USER_FUNC_2(t *testing.T) {
	input := "{ myadd: fn { a b } { add a b } 100 |myadd 50 |subtract 66 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 84 {
		t.Error("Expected result value 84")
	}
}

func TestEvaldo_OPWORDS_USER_FUNC_3(t *testing.T) {
	input := "{ myadd: fn { a b } { add a b } 100 |add 50 |subtract 20 |add 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 140 {
		t.Error("Expected result value 140")
	}
}

//
// IDEA .. WE COULD HAVE TWO OPWORD TYPES ONE WHICH EVALUATER STRICTLY L -> R AND OTHER THAT
// EVALS TO THE RIGHT EXPRESSION THIS WAY WE WOULD LOOSE THE NEED FOR PARENTHESIS IN MANY CASES
// W PARENS WE QUICKLY LOOSE THE FLOW ELEGANCE ... MAYBE THIS WOULD BE GOOD SOLUTION. LIKE |word and .word
// 100 |add 100 |subtract 100 .add 50
// vs.
// 100 |add 100 |subtract 100 .add 50
//
// | LR .L
//
// Hm .. could we also then make priority of other operators similar to math?
// 100 - 10 + 20 = 110 (- LR + LR)
// 100 - 10 * 2  = 80  (- LR * L)
// 10 * 4 + 30 * 2 - 10 / 5 = reb26 | py98 (* L + LR * L - LR / L) -- it seems it would work
//

// aleready implemented pipewords and opwords first eval LP secont ... HM .. but doesn't work as it should. one op/pipe word before now affects the L/LR .. which
// probably doesn't make sense ... I will continue on this later .. now just leave it as it is for now ... we will see how we can make it as we imagined first
// if it doesn't screw up our initial code that much and
// test show us that .subtract then consumes all opwords and pipewords on right. Which doesn't behave as I expected and can't be desirable.

func TestEvaldo_OPWORDS_USER_FUNC_toleft_1(t *testing.T) {
	input := "{ 150 |subtract 20 |add 20 |add 10 |add 10 |add 20 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 120 {
		t.Error("Expected result value 120")
	}
}

func __TestEvaldo_OPWORDS_USER_FUNC_toleft_2(t *testing.T) {
	input := "{ 150 |subtract 20 |add 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 140 {
		t.Error("Expected result value 140")
	}
}
*/
