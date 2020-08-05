package contexts

import (
	"Ryelang/env"
	"Ryelang/evaldo"
	"Ryelang/loader"

	//	"fmt"
	"testing"
)

func TestContexts_1(t *testing.T) {
	input := "{ c: context { a: 1 b: \"in context\" } get c 'b }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "in context") {
		t.Error("Expected result value 'in context'")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestContexts_2(t *testing.T) {
	input := "{ c: context { a: 1 d: fn { } { add 1 2 } } ff: get c 'd }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Error")
	}
	if !(es.Res.(env.Error).Status == 5) {
		t.Error("Expected status 5")
	}
	if !es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestContexts_in_1(t *testing.T) {
	input := "{ c: context-in { d: fn { } { add 1 2 } } ff: get c 'd ff }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 3) {
		t.Error("Expected status 3")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func _____TestContexts_in_1(t *testing.T) {
	// doesn't work since function that is got from context and then ran runs in current context and doesn't have access to variable a
	// to run it in it's own context we must use the cpath (contextpath) that gives function also a context to run in
	input := "{ c: context-in { a: 1 b: \"in context\" d: fn { a } { add a 2 } } ff: get c 'd ff 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 12) {
		t.Error("Expected result value 12")
	}
	if !es.ErrorFlag {
		t.Error("Expected Error flag")
	}
}

func TestContexts_in_2(t *testing.T) {
	input := "{ c: context-in { a: 1 b: \"in context\" d: fn { a } { add a 2 } } c/d 998 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 1000) {
		t.Error("Expected result value 1000")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestContexts_in_3(t *testing.T) {
	input := "{ c: context-in { a: 1 d: fn { } { add a 1 } } c/d }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 2) {
		t.Error("Expected result value 2")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestContexts_in_4(t *testing.T) {
	input := "{ math3: context-in { a: 999 add3: fn { a b c } { add add a b c } my-add: fn { b } { add3 a 2 b } } math3/my-add 2 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 1003) {
		t.Error("Expected result value 1003")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestContexts_in_context(t *testing.T) {
	input := "{ users: context-in { admins: context-in { check: fn { id } { either id { \"OK\" } { \"WRONG\" } } } } adm: users/admins adm/check 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "OK") {
		t.Error("Expected result value OK")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestContexts_in_context_cpath3(t *testing.T) {
	input := "{ users: context-in { admins: context-in { check: fn { id } { either id { \"OK\" } { \"WRONG\" } } } } adm: users/admins/check 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "OK") {
		t.Error("Expected result value OK")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}
