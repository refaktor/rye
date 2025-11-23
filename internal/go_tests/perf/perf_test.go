package perf

import (
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
)

/*
func TestEvaldo_perf_loop_1000(t *testing.T) {
	input := "{ aa: 1 print dotime { loop 10000000 { aa } } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)
}

func TestEvaldo_perf_loop_1000_word(t *testing.T) {
	input := "{ aa: 1 print dotime { loop 10000000 { aa aa } } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)
}

func TestEvaldo_perf_loop_1000_setword(t *testing.T) {
	input := "{ print dotime { loop 10000000 { aa: 1 aa: 2 } } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)
}
*/

func TestEvaldo_perf_loop_1000_func0(t *testing.T) {
	input := " loop 10000000 { oneone } }" //10000000
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)
}

func TestEvaldo_perf_loop_1000_func2(t *testing.T) {
	input := " loop 10000000 { add 1 2 } " // 10000000
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)
}

func TestEvaldo_perf_loop_1000_user_func2(t *testing.T) {
	input := " add1: fn { aa bb } { add aa bb } loop 10000000 { add1 1 2 } " // 10000000
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)
}

// after adding curry: 1.75s, 2.9s, 19s
