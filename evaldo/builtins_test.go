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

func __TestEvaldo_load_builtin_either_2_in_func(t *testing.T) {
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

func TestEvaldo_strings(t *testing.T) {
	input := "{ a: left \"1234567\" 3 |right 2 , \"ABCDEF\" |middle 1 3 |join a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type string")
	}
	if !(es.Res.(env.String).Value == "BC23") {
		t.Error("Expected result value BC23")
	}
}

func TestEvaldo_strings_builtin_userfn_1(t *testing.T) {
	input := "{ a: join \"Woof\" \"Meow\" aa: fn { a } { join a a } aa a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type string")
	}
	if !(es.Res.(env.String).Value == "WoofMeowWoofMeow") {
		t.Error("Expected result value WoofMeowWoofMeow")
	}
}

func TestEvaldo_strings_pipe_op(t *testing.T) {
	input := "{ a: \"123ščž\" |join \"4567\" .right 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type string")
	}
	if !(es.Res.(env.String).Value == "123ščž567") {
		t.Error("Expected result value 123ščž567")
	}
}

func TestEvaldo_series_nth(t *testing.T) {
	input := "{ { 101 102 103 } .nth 2 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 102) {
		t.Error("Expected result value 102")
	}
}

func TestEvaldo_series_length(t *testing.T) {
	input := "{ { 101 102 103 } .length }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 3) {
		t.Error("Expected result value 3")
	}
}

func TestEvaldo_series_next_pop_peek(t *testing.T) {
	input := "{ a: { 101 102 103 } b: a .next .peek .add 10000  }" // POP doesn't make sense right now as builtins can't change objects now. If is this good / safety or bad /too restricting
	// we will still have to figure out. So internal position can't be changed. builtin can just return series with new pos
	// like next
	// but not return value and change the object in it's place
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 10102) {
		t.Error("Expected result value 10102")
	}
}

// { a: { 101 102 103 } b: a .nth 1 |add 100 }" // POP doesn't make sense right now as builtins can't change objects now. If is this good / safety or bad /too restricting
