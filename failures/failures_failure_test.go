package failures

import (
	"../env"
	"../evaldo"
	"../loader"

	//	"fmt"
	"testing"
)

func TestFailures_nofailure1(t *testing.T) {
	input := "{ print 111 return 123 print 333 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 123) {
		t.Error("Expected result value 123")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
	if es.FailureFlag {
		t.Error("Expected No Failure flag")
	}
}

func TestFailures_failure1(t *testing.T) {
	input := "{ print 111 failure 123 print 333 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 123) {
		t.Error("Expected result value 123")
	}
	if !(es.FailureFlag) {
		t.Error("Expected Fail flag")
	}
	if !(es.ErrorFlag) {
		t.Error("Expected Error flag")
	}
}

func TestFailures_fail1(t *testing.T) {
	input := "{ print 111 fail 123 print 333 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 123) {
		t.Error("Expected result value 123")
	}
	if !(es.FailureFlag) {
		t.Error("Expected Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestFailures_fail2(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } f 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 333) {
		t.Error("Expected result value 333")
	}
	if !(es.FailureFlag) {
		t.Error("Expected Fail flag")
	}
	if !(es.ErrorFlag) {
		t.Error("Expected Error flag")
	}
}

func TestFailures_fail3(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } 2 |f 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 333) {
		t.Error("Expected result value 333")
	}
	if !(es.FailureFlag) {
		t.Error("Expected Fail flag")
	}
	if !(es.ErrorFlag) {
		t.Error("Expected Error flag")
	}
}

func TestFailures_fail4(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } add 2 .f 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 333) {
		t.Error("Expected result value 333")
	}
	if !(es.FailureFlag) {
		t.Error("Expected Fail flag")
	}
	if !(es.ErrorFlag) {
		t.Error("Expected Error flag")
	}
}

func TestFailures_fail5(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } add 2 f 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Error")
	}
	if !(es.Res.(env.Error).Status == 333) {
		t.Error("Expected result value 333")
	}
	if !(es.FailureFlag) {
		t.Error("Expected Fail flag")
	}
	if !(es.ErrorFlag) {
		t.Error("Expected Error flag")
	}
}

// disarm

func TestFailures_disarmfailure1(t *testing.T) {
	input := "{ print 111 disarm failure 123 print 333 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 333) {
		t.Error("Expected result value 333")
	}
	if es.FailureFlag {
		t.Error("Expected No Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestFailures_disarmfailure2(t *testing.T) {
	input := "{ either 1 { 123 } { failure 999 } either 0 { 321 } { failure 888 } 444 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 888) {
		t.Error("Expected result value 888")
	}
	if !es.FailureFlag {
		t.Error("Expected Fail flag")
	}
	if !es.ErrorFlag {
		t.Error("Expected Error flag")
	}
} // QUESTION: should failure on the end of block turn not Error or not? Probably only the function return should reset it
// maybe don't turn to error at the end of the block
func TestFailures_disarmfailure3(t *testing.T) {
	input := "{ either 1 { 123 } { failure 999 } disarm either 0 { 321 } { failure 888 } 444 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 888) {
		t.Error("Expected result value 888")
	}
	if !es.FailureFlag {
		t.Error("Expected Fail flag")
	}
	if !es.ErrorFlag {
		t.Error("Expected Error flag")
	}
}

func TestFailures_disarmfailure4(t *testing.T) {
	input := "{ either 1 { 123 } { failure 999 } either 0 { 321 } { disarm failure 888 } 444 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 444) {
		t.Error("Expected result value 444")
	}
	if es.FailureFlag {
		t.Error("Expected No Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

// after adding curry: 1.75s, 2.9s, 19s

func TestFailures_disarmfail2(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } disarm f 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Error).Status == 333) {
		t.Error("Expected result value 333")
	}
	if es.FailureFlag {
		t.Error("Expected No Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestFailures_disarmfail3(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } 2 |f |disarm |status |add 1  }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 334) {
		t.Error("Expected result value 334")
	}
	if es.FailureFlag {
		t.Error("Expected Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected Error flag")
	}
}
func TestFailures_disarmfail4(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } add 2 .f .disarm .status 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 334) {
		t.Error("Expected result value 334")
	}
	if es.FailureFlag {
		t.Error("Expected Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected Error flag")
	}
}

func TestFailures_disarmfail5(t *testing.T) {
	input := "{ f: fn { a } { 123 fail 333 444 } add 2 status disarm f 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 335) {
		t.Error("Expected result value 335")
	}
	if es.FailureFlag {
		t.Error("Expected Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected Error flag")
	}
}

func TestFailures_assert1(t *testing.T) {
	input := "{ ^assert 0 \"some error\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Error")
	}
	if !(es.Res.(env.Error).Message == "some error") {
		t.Error("Expected result value some error")
	}
	if !es.FailureFlag {
		t.Error("Expected Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected Error flag")
	}
}
func TestFailures_assert2(t *testing.T) {
	input := "{ ^assert 1 \"some error\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.VoidType {
		t.Error("Expected result type Error")
	}
	if es.FailureFlag {
		t.Error("Expected No Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}
func TestFailures_assert2check(t *testing.T) {
	input := "{ f1: fn { a } { ^assert 0 \"some error\" } 123 f1 1 |^check \"some other error\"  }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Error")
	}
	if !es.FailureFlag {
		t.Error("Expected Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}

func TestFailures_assert2check2(t *testing.T) {
	input := "{ f1: fn { a } { ^assert 0 \"some error\" } f2: fn { } { f1 1 |^check \"some other error\" } print disarm f2 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Error")
	}
	if es.FailureFlag {
		t.Error("Expected No Fail flag")
	}
	if es.ErrorFlag {
		t.Error("Expected No Error flag")
	}
}
