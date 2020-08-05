package failures

import (
	"../env"
	"../evaldo"
	"../loader"
	"testing"
)

func TestFailures_no_error1(t *testing.T) {
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
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}
