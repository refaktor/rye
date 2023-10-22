// main_test.go
package evaldo

import (
	"rye/env"
	"rye/loader"

	"fmt"

	"testing"
)

//
// Function
//

func TestEvaldo_function1_just_return_integer(t *testing.T) {
	input := "  fun1  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")
	if found {
		body := []env.Object{env.Integer{234}}
		spec := []env.Object{}
		es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

		EvalBlock(es)

		fmt.Print(es.Res.Inspect(*es.Idx))
		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}
		if es.Res.(env.Integer).Value != 234 {
			t.Error("Expected result value 1001")
		}
	} else {
		t.Error("Word not found in index")
	}
}

func TestEvaldo_function1_just_return_integer_in_loop(t *testing.T) {
	input := "  loop 10000 { fun1 }  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")
	if found {
		body := []env.Object{env.Integer{2345}}
		spec := []env.Object{}
		es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

		EvalBlock(es)

		fmt.Print(es.Res.Inspect(*es.Idx))
		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}
		if es.Res.(env.Integer).Value != 2345 {
			t.Error("Expected result value 1001")
		}
	} else {
		t.Error("Word not found in index")
	}
}

func TestEvaldo_function2_call_builtin_in_func(t *testing.T) {
	input := "  fun1  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")

	if found {
		incidx, found1 := es.Idx.GetIndex("inc")
		//printidx, _ := es.Idx.GetIndex("print")
		if found1 { // env.Word{printidx},
			body := []env.Object{env.Word{incidx}, env.Word{incidx}, env.Word{incidx}, env.Integer{330}}
			spec := []env.Object{}
			es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

			EvalBlock(es)

			fmt.Print(es.Res.Inspect(*es.Idx))
			if es.Res.Type() != env.IntegerType {
				t.Error("Expected result type integer")
			}
			if es.Res.(env.Integer).Value != 333 {
				t.Error("Expected result value 1001")
			}
		} else {
			t.Error("Builting inc Word not found in index")
		}
	} else {
		t.Error("Word not found in index")
	}
}

func TestEvaldo_function3_arg_unit_func(t *testing.T) {
	input := "  fun1 789  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")

	if found {
		aaaidx := es.Idx.IndexWord("aaa")
		//printidx, _ := es.Idx.GetIndex("print")
		body := []env.Object{env.Word{aaaidx}}
		spec := []env.Object{env.Word{aaaidx}}
		es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

		EvalBlock(es)

		fmt.Print(es.Res.Inspect(*es.Idx))
		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}
		if es.Res.(env.Integer).Value != 789 {
			t.Error("Expected result value 789")
		}
	} else {
		t.Error("Word not found in index")
	}
}

func TestEvaldo_function3_arg_inc_func(t *testing.T) {
	input := "  fun1 999  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")

	if found {
		incidx, found1 := es.Idx.GetIndex("inc")
		//printidx, _ := es.Idx.GetIndex("print")
		if found1 { // env.Word{printidx},

			aaaidx := es.Idx.IndexWord("aaa")
			//printidx, _ := es.Idx.GetIndex("print")
			spec := []env.Object{env.Word{aaaidx}}
			body := []env.Object{env.Word{incidx}, env.Word{aaaidx}}
			es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

			EvalBlock(es)

			fmt.Print(es.Res.Inspect(*es.Idx))
			if es.Res.Type() != env.IntegerType {
				t.Error("Expected result type integer")
			}
			if es.Res.(env.Integer).Value != 1000 {
				t.Error("Expected result value 1000")
			}
		} else {
			t.Error("Word not found in index")
		}
	} else {
		t.Error("Word not found in index")
	}
}

func TestEvaldo_function3_arg_unit_func_loop(t *testing.T) {
	input := "  loop 2 { fun1 999 }  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")

	if found {
		//incidx, found1 := es.Idx.GetIndex("inc")
		//printidx, _ := es.Idx.GetIndex("print")
		//if found1 { // env.Word{printidx},

		aaaidx := es.Idx.IndexWord("aaa")
		//printidx, _ := es.Idx.GetIndex("print")
		spec := []env.Object{env.Word{aaaidx}}
		body := []env.Object{env.Word{aaaidx}}
		es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

		EvalBlock(es)

		fmt.Print(es.Res.Inspect(*es.Idx))
		if es.Res.Type() != env.IntegerType {
			t.Error("Expected result type integer")
		}
		if es.Res.(env.Integer).Value != 999 {
			t.Error("Expected result value 999")
		}
		//} else {
		//	t.Error("Word not found in index")
		//}
	} else {
		t.Error("Word not found in index")
	}
}

func TestEvaldo_function3_arg_inc_func_loop(t *testing.T) {
	//input := "  fun1 999  "
	input := "  loop 1000 { fun1 999 }  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")

	if found {
		incidx, found1 := es.Idx.GetIndex("inc")
		//printidx, _ := es.Idx.GetIndex("print")
		if found1 { // env.Word{printidx},

			aaaidx := es.Idx.IndexWord("aaa")
			//printidx, _ := es.Idx.GetIndex("print")
			spec := []env.Object{env.Word{aaaidx}}
			body := []env.Object{env.Word{incidx}, env.Word{aaaidx}}
			es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

			EvalBlock(es)

			fmt.Print(es.Res.Inspect(*es.Idx))
			if es.Res.Type() != env.IntegerType {
				t.Error("Expected result type integer")
			}
			if es.Res.(env.Integer).Value != 1000 {
				t.Error("Expected result value 1000")
			}
		} else {
			t.Error("Word not found in index")
		}
	} else {
		t.Error("Word not found in index")
	}
}

func _TestEvaldo_function4_simple_recur(t *testing.T) {
	input := "  fun1 1  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	idx, found := es.Idx.GetIndex("fun1")

	if found {
		incidx, found1 := es.Idx.GetIndex("inc")
		recuridx, _ := es.Idx.GetIndex("recur1if")
		printidx, _ := es.Idx.GetIndex("print")
		greateridx, _ := es.Idx.GetIndex("greater")
		if found1 { // env.Word{printidx},

			aaaidx := es.Idx.IndexWord("aaa")
			//printidx, _ := es.Idx.GetIndex("print")
			spec := []env.Object{env.Word{aaaidx}}
			body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
			es.Ctx.Set(idx, *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), *env.NewBlock(*env.NewTSeries(body)), false))

			EvalBlock(es)

			fmt.Print(es.Res.Inspect(*es.Idx))
			if es.Res.Type() != env.IntegerType {
				t.Error("Expected result type integer")
			}
			if es.Res.(env.Integer).Value != 100 {
				t.Error("Expected result value 100")
			}
		} else {
			t.Error("Word not found in index")
		}
	} else {
		t.Error("Word not found in index")
	}
}

func TestEvaldo_function4_simple_fn_in_code(t *testing.T) {
	input := "  fun1: fn { aa } { aa } fun1 1234  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1234 {
		t.Error("Expected result value 1234")
	}
}

func TestEvaldo_function4_simple_fn_in_code2(t *testing.T) {
	input := "  fun1: fn { aa } { inc aa } fun1 1234  "
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type integer")
	}
	if es.Res.(env.Integer).Value != 1235 {
		t.Error("Expected result value 1235")
	}
}
