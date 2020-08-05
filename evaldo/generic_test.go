package evaldo

import (
	"Ryelang/env"
	"Ryelang/loader"
	"fmt"

	//	"fmt"

	//"fmt"

	//	"fmt"

	"testing"
)

//
// Literal values
//

func _TestEvaldo_generic_register(t *testing.T) {
	input := "{ generic 'integer 'gadd fn { a b } { add a b } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	RegisterBuiltins(es)

	EvalBlock(es)

	es.Idx.Probe()

	es.Gen.Probe(*es.Idx)

	/*if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}*/

}

func TestEvaldo_generic_use(t *testing.T) {
	input := "{ generic 'integer 'Gadd fn { a } { add a 123 } Gadd 1000 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	RegisterBuiltins(es)

	fmt.Println("EVALUATING >>")

	EvalBlock(es)

	es.Idx.Probe()

	es.Gen.Probe(*es.Idx)

	fmt.Println("RES >>")

	es.Res.Trace("RES")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

}

func TestEvaldo_generic_use_2(t *testing.T) {
	input := "{ generic 'integer 'Add fn { a } { add a 123 } generic 'string 'Add fn { a } { join a \"--added--\" } Add 1000 Add \"woof\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	RegisterBuiltins(es)

	fmt.Println("EVALUATING >>")

	EvalBlock(es)

	es.Idx.Probe()

	es.Gen.Probe(*es.Idx)

	fmt.Println("RES >>")

	es.Res.Trace("RES")

	if es.Res.Type() != env.StringType {
		t.Error("Expected result type integer")
	}

}

func TestEvaldo_generic_function_2_args(t *testing.T) {
	input := "{ generic 'integer 'Add fn { a b } { add a b } Add 2 5 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	RegisterBuiltins(es)

	fmt.Println("EVALUATING >>")

	EvalBlock(es)

	es.Idx.Probe()

	es.Gen.Probe(*es.Idx)

	fmt.Println("RES >>")

	es.Res.Trace("RES")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

}

func TestEvaldo_generic_builtin_01(t *testing.T) {
	input := "{ generic 'integer 'add1 ?add add1 100 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	RegisterBuiltins(es)

	fmt.Println("EVALUATING >>")

	EvalBlock(es)

	es.Idx.Probe()

	es.Gen.Probe(*es.Idx)

	fmt.Println("RES >>")

	es.Res.Trace("RES")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
}

func TestEvaldo_generic_builtin(t *testing.T) {
	input := "{ generic 'integer 'add1 ?add add1 100 10 generic 'string 'add1 ?join add1 \"Wood\" \"Meow\" add1 100 11 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	RegisterBuiltins(es)

	fmt.Println("EVALUATING >>")

	EvalBlock(es)

	es.Idx.Probe()

	es.Gen.Probe(*es.Idx)

	fmt.Println("RES >>")

	es.Res.Trace("RES")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
}
