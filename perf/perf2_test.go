package perf

import (
	"Ryelang/env"
	"Ryelang/evaldo"
	"Ryelang/loader"

	//	"fmt"
	"testing"
)

func TestEvaldo_function4_factorial_w_recur2(t *testing.T) {
	input := "{ factorial: fn { nn aa } { recur2if greater nn 0  subtract nn 1 multiply nn aa aa } loop 100000 { factorial 12 1 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)

	evaldo.EvalBlock(es)

	es.Res.Trace("returned")
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 479001600 {
		t.Error("Expected result value 479001600")
	}
}

func TestEvaldo_function4_factorial_w_recursive(t *testing.T) { //2.13s
	input := "{ factorial: fn { nn } { either greater nn 1 { multiply nn factorial subtract nn 1 } { 1 } } loop 100000 { factorial 12 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)

	evaldo.EvalBlock(es)

	es.Res.Trace("returned")
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 479001600 {
		t.Error("Expected result value 479001600")
	}
}

// fibonacci: fn { n } {  either lesser? n 3 { n } {  add fibonacci n - 2 fibonacci n - 1 } } x: miliseconds y: fibonacci 30 print miliseconds-from x y }'), 0, {}, 0)), 479001600)

func TestEvaldo_function4_fibonnaci_recursive(t *testing.T) { //17s
	input := "{ fibonacci: fn { nn } {  either lesser nn 2 { nn } {  add fibonacci subtract nn 2 fibonacci subtract nn 1 } } fibonacci 30 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)

	evaldo.EvalBlock(es)

	es.Res.Trace("returned")
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 832040 {
		t.Error("Expected result value 832040")
	}
}

func TestEvaldo_function4_fibonnaci_w_recur3(t *testing.T) { // 0s
	input := "{ fibonacci: fn { nn aa bb } { recur3if greater nn 1 subtract  nn 1  bb  add aa bb  either lesser nn 1 { aa } { bb } } fibonacci 30 0 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)

	evaldo.EvalBlock(es)

	es.Res.Trace("returned")
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 832040 {
		t.Error("Expected result value 832040")
	}
}
