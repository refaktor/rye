// main_test.go
package evaldo

import (
	"fmt"
	"rye/env"
	"rye/loader"

	"testing"
)

//
// Literal values
//

func TestEvaldo_load_integer2(t *testing.T) {
	input := "  123  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)

	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.(env.Block).Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}

	if es.Res != nil {
		t.Error("Expected result nil")
	}

	EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

}

func TestEvaldo_load_integer3(t *testing.T) {
	input := "  123  234  435  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)

	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 1 items")
	}
	if block.(env.Block).Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}

	EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

	if es.Res.(env.Integer).Value != 435 {
		t.Error("Expected result value 435")
	}

}

//
// words
//

func TestEvaldo_load_word1(t *testing.T) {
	input := "  someval  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	idx, found := es.Idx.GetIndex("someval")
	if found {
		es.Ctx.Set(idx, env.Integer{10101})

		if block.(env.Block).Series.Len() != 1 {
			t.Error("Expected 1 items")
		}
		if block.(env.Block).Series.Get(0).(env.Object).Type() != env.WordType {
			t.Error("Expected type word")
		}

		EvalBlock(es)

		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}

		if es.Res.(env.Integer).Value != 10101 {
			t.Error("Expected result value 10101")
		}
	} else {
		t.Error("Word not found in state")
	}
}

//
// setwords
//

func TestEvaldo_load_setword1_pass_val(t *testing.T) {
	input := "  otherval: 20202  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	idx, found := es.Idx.GetIndex("someval")
	if found {
		es.Ctx.Set(idx, env.Integer{10101})

		if block.(env.Block).Series.Len() != 2 {
			t.Error("Expected 2 items")
		}
		if block.(env.Block).Series.Get(0).(env.Object).Type() != env.SetwordType {
			t.Error("Expected type setword")
		}

		EvalBlock(es)

		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}

		if es.Res.(env.Integer).Value != 20202 {
			t.Error("Expected result value 20202")
		}
	} else {
		t.Error("Word not found in state")
	}
}

func TestEvaldo_load_setword1_stores_val(t *testing.T) {
	input := "  otherval: 40404 30303 someval otherval  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	idx, found := es.Idx.GetIndex("someval")
	if found {
		es.Ctx.Set(idx, env.Integer{10101})

		EvalBlock(es)

		fmt.Println(es.Idx.GetIndex("otherval"))

		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}

		if es.Res.(env.Integer).Value != 40404 {
			t.Error("Expected result value 40404")
		}
	} else {
		t.Error("Word not found in state")
	}
}

//
// builtins
//

func DISABLED_TestEvaldo_load_builtin(t *testing.T) {
	input := "  oneone  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)

	RegisterBuiltins(es)

	// TODO -- think how builtin will be available .. as a builtin object registered to word or as something in some other array??
	// TODO -- make it work ... at least in some simple way
	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

	if es.Res.(env.Integer).Value != 11 {
		t.Error("Expected result value 11")
	}
}

func DISABLED_TestEvaldo_load_builtin2(t *testing.T) {
	input := "  num: oneone 33  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 33 {
		t.Error("Expected result value 33")
	}
}

func DISABLED_TestEvaldo_load_builtin3(t *testing.T) {
	input := "  num: oneone 33 num  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 11 {
		t.Error("Expected result value 11")
	}
}

func TestEvaldo_load_builtin_1_arg(t *testing.T) {
	input := "  inc 1000  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1001 {
		t.Error("Expected result value 1001")
	}
}

func TestEvaldo_load_builtin_2_arg(t *testing.T) {
	input := "  1000 + 777  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1777 {
		t.Error("Expected result value 1777")
	}
}

func TestEvaldo_load_builtin_1_2_arg(t *testing.T) {
	input := "  1000 + inc inc 777  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1779 {
		t.Error("Expected result value 1779")
	}
}

func TestEvaldo_load_builtin_1_2_arg_setwords(t *testing.T) {
	input := "  sum: inc 1000 + inc inc 777 word: sum  sum + 10000  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 11780 {
		t.Error("Expected result value 11780")
	}
}

func TestEvaldo_load_builtin_loop(t *testing.T) {
	input := "  sum: 1 loop 1000 { sum: sum + 1 } sum  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1001 {
		t.Error("Expected result value 1001")
	}
}

func DISABLE_TestEvaldo_curry_1(t *testing.T) {
	input := "  add100: add 100 _ , add100 11  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if es.Res.(env.Integer).Value != 111 {
		t.Error("Expected result value 111")
	}
}

func TestEvaldo_curry_2(t *testing.T) {
	input := "  second: nth _ 2 , { 11 22 33 } .second  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if es.Res.(env.Integer).Value != 22 {
		t.Error("Expected result value 22")
	}
}

func TestEvaldo_load_lsetword1(t *testing.T) {
	input := "  30303 :lword 10101 lword  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	idx, found := es.Idx.GetIndex("lword")
	if found {
		es.Ctx.Set(idx, env.Integer{0})

		if block.(env.Block).Series.Len() != 4 {
			t.Error("Expected 4 items")
		}
		if block.(env.Block).Series.Get(1).(env.Object).Type() != env.LSetwordType {
			t.Error("Expected type setword")
		}

		EvalBlock(es)

		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}

		if es.Res.(env.Integer).Value != 30303 {
			t.Error("Expected result value 30303")
		}
	} else {
		t.Error("Word not found in state")
	}
}

func TestEvaldo_load_lsetword2(t *testing.T) {
	input := "  10 + 20 :lword 123456 lword  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

	if es.Res.(env.Integer).Value != 30 {
		t.Error("Expected result value 30")
	}
}

func TestEvaldo_load_lsetword3(t *testing.T) {
	input := "  10 + 20 + 20 |+ 50 :lword 123456 lword  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

	if es.Res.(env.Integer).Value != 100 {
		t.Error("Expected result value 100")
	}
}

func TestEvaldo_load_lsetword4(t *testing.T) {
	input := "  10 + 20 :sum1 + 20 |+ sum1 :lword 123456 lword  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}

	if es.Res.(env.Integer).Value != 80 {
		t.Error("Expected result value 80")
	}
}
