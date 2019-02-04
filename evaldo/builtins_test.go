// builtins_test.go
package evaldo

import (
	"Rejy_go_v1/env"
	"Rejy_go_v1/loader"
	"fmt"

	//"fmt"
	"testing"
)

//
// invoke builtins manually
//

func TestBuiltin_oneone(t *testing.T) {

	genv := loader.GetIdxs()
	builtin := builtins["oneone"]
	es := env.NewProgramState(env.TSeries{}, genv)
	obj := builtin.Fn(es, nil, nil, nil, nil, nil)
	fmt.Println(obj.(env.Object).Inspect(*genv))
	if obj.(env.Integer).Value != 11 {
		t.Error("Not 11 returned")
	}
}

func TestBuiltin_inc(t *testing.T) {
	genv := loader.GetIdxs()
	builtin := builtins["inc"]
	es := env.NewProgramState(env.TSeries{}, genv)
	obj := builtin.Fn(es, env.Integer{100}, nil, nil, nil, nil)
	fmt.Println(obj.(env.Object).Inspect(*genv))
	if obj.(env.Integer).Value != 101 {
		t.Error("Not 101 returned")
	}
}

func TestBuiltin_print(t *testing.T) {
	genv := loader.GetIdxs()
	builtin := builtins["print"]
	es := env.NewProgramState(env.TSeries{}, genv)
	obj := builtin.Fn(es, env.Integer{1010101010101}, nil, nil, nil, nil)
	fmt.Println(obj.(env.Object).Inspect(*genv))
	if obj.(env.Integer).Value != 1010101010101 {
		t.Error("Not 1010101010101 returned")
	}
}

//
// invoke functions via code
//

func TestEvaldo_load_builtin_if_1(t *testing.T) {
	input := "{ if 1 { 123 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 123 {
		t.Error("Expected result value 123")
	}
}

func TestEvaldo_load_builtin_do_1(t *testing.T) {
	input := "{ 123 do { add 2 3 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 5 {
		t.Error("Expected result value 123")
	}
}

func TestEvaldo_load_builtin_either_1(t *testing.T) {
	input := "{ either 1 { 1234 } { 9876 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	//fmt.Print(es.Res.Inspect(es.Idx))
	fmt.Print(es.Res)
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1234 {
		t.Error("Expected result value 1234")
	}
}

func TestEvaldo_load_builtin_either_2(t *testing.T) {
	input := "{ either 0 { 1234 } { 9876 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 9876 {
		t.Error("Expected result value 9876")
	}
}

func TestEvaldo_load_builtin_either_1_in_func(t *testing.T) {
	input := "{ func1: fn { aa } { either aa { 1234 } { 9876 } } loop 2 { func1 1 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1234 {
		t.Error("Expected result value 1234")
	}
}

func TestEvaldo_load_builtin_either_2_in_func(t *testing.T) {
	input := "{ func1: fn { aa } { either aa { 1234 } { 9876 } } loop 2 { func1 0 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 9876 {
		t.Error("Expected result value 9876")
	}
}

func TestEvaldo_load_builtin_either_3_in_func(t *testing.T) {
	input := "{ func1: fn { aa bb } { either multiply aa bb { add 100 bb } { add func1 1 bb 22 } } loop 2 { func1 0 100 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 222 {
		t.Error("Expected result value 222")
	}
}
