// builtins_test.go
package evaldo

import (
	"../env"
	"../loader"
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

func TestRawMap1(t *testing.T) {
	input := "{ raw-map { \"age\" 123 \"name\" \"Jimbo\" } 123 }" // POP doesn't make sense right now as builtins can't change objects now. If is this good / safety or bad /too restricting
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 123) {
		t.Error("Expected result value 123")
	}
}

func TestValidateRawMap(t *testing.T) {
	//input := "{ rm: raw-map { \"age1\" 1234 \"name\" \"Jimbo\" } vm: validate rm { age: optional \"123\" integer } get vm \"age\" }"
	input := "{ { \"age1\" 1234 \"name\" \"Jimbo\" } |raw-map |validate { age: optional \"123\" integer } |get \"age\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}
	if !(es.Res.(env.Integer).Value == 123) {
		t.Error("Expected result value 123")
	}
}

func TestValidateRawMap2(t *testing.T) { // passing thru
	input := "{ { \"age1\" 1234 \"name\" \"Jimbo\" } |raw-map |validate { name: optional \"JoeDoe\" string } |get \"name\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "Jimbo") {
		t.Error("Expected result value Jimbo")
	}
}

func TestValidateRawMap3(t *testing.T) { // using default
	input := "{ { \"age1\" 1234 \"name1\" \"Jimbo\" } |raw-map |validate { name: optional \"JoeDoe\" string } |get \"name\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "JoeDoe") {
		t.Error("Expected result value JoeDoe")
	}
}

func TestValidateRawMap4(t *testing.T) { // two keys, int as string
	input := "{ { \"age1\" 1234 \"name1\" \"Jimbo\" } |raw-map |validate { age: optional 111 string  name: optional \"JoeDoe\" string } |get \"age\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "111") {
		t.Error("Expected result value 111")
	}
}

func TestValidateRawMap5(t *testing.T) { // two keys, int as string, pass true
	input := "{ { \"age\" 1234 \"name1\" \"Jimbo\" } |raw-map |validate { age: required string } |get \"age\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "1234") {
		t.Error("Expected result value 1234")
	}
}

func TestValidateRawMapEmail1(t *testing.T) { // two keys, int as string, pass true
	input := "{ { \"em\" \"toto.bam@gmail.com\" } |raw-map |validate { em: required email } |get \"em\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.StringType {
		t.Error("Expected result type String")
	}
	if !(es.Res.(env.String).Value == "toto.bam@gmail.com") {
		t.Error("Expected result value toto.bam@gmail.com")
	}
}

func TestValidateRawMapDate1(t *testing.T) { // two keys, int as string, pass true
	input := "{ { \"da\" \"02.01.1020\" } |raw-map |validate { da: required date } |get \"da\" |print }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.DateType {
		t.Error("Expected result type Date")
	}
}

func TestValidateRawMapCheck1(t *testing.T) { // two keys, int as string, pass true
	input := "{ { \"num\" 100 } |raw-map |validate { num: optional 0 check \"too-low\" { .greater 50 } } |get \"num\" |print }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}
	if !(es.Res.(env.Integer).Value == 100) {
		t.Error("Expected result value 100")
	}

}

func TestValidateRawMapCalc1(t *testing.T) { // two keys, int as string, pass true
	input := "{ { \"numX\" 100 } |raw-map |validate { num: optional 1 calc { .add 10000 } } |get \"num\" |print }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	es.Res.Trace("dasasd")

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Integer).Value == 10001) {
		t.Error("Expected result value 123")
	}

}

func TestReturn(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 return 2 print 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Integer).Value == 2) {
		t.Error("Expected result value 2")
	}

}

func TestReturnDo(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 do { print 2 return 3 print 4 } print 5 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Integer).Value == 3) {
		t.Error("Expected result value 3")
	}
}

func TestReturnDoIf(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 do { print 2 if 1 { print 3 return 4 print 5 } print 6 } print 7 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Integer).Value == 4) {
		t.Error("Expected result value 4")
	}
}

func TestReturnInsideFn0(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 a: fn { } { print 11 return 22 print 33 } print 2 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Integer).Value == 2) {
		t.Error("Expected result value 2")
	}
}

func TestReturnInsideFn1(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 a: fn { } { print 11 return 22 print 33 } print 2 print a print 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Integer).Value == 3) {
		t.Error("Expected result value 3")
	}
}

func TestReturnInsideFn2(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 a: fn { } { print 11 return 22 print 33 } b: fn { } { print 111 return a print 333 } print 2 print b print 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Integer).Value == 3) {
		t.Error("Expected result value 3")
	}
}

func TestCriticalFailureInsideFn2InBuiltinExpr(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 a: fn { } { print 11 error 22 print 33 } b: fn { } { print 111 return a print 333 } print 2 add b 2 print 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Date")
	}

	if !(es.Res.(env.Error).Status == 22) {
		t.Error("Expected result value 22")
	}
}

func TestCriticalFailureInsideFn2InFnExpr(t *testing.T) { // two keys, int as string, pass true
	input := "{ print 1 a: fn { } { print 11 error 22 print 33 } b: fn { a } { print 111 print a print 333 } print 2 b a print 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.ErrorType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Error).Status == 22) {
		t.Error("Expected result value 22")
	}
}

// FNC

func Test_Fnc_make_adder(t *testing.T) { // two keys, int as string, pass true
	input := "{ make-adder: fn { b } { fnc { a } context { b: b } { add a b } } add2: make-adder 2 add5: make-adder 5 add2 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 12) {
		t.Error("Expected result value 12")
	}
}

func Test_Fnc_make_adder_2(t *testing.T) { // two keys, int as string, pass true
	input := "{ make-adder: fn { b } { fnc { a } current-context { add a b } } add2: make-adder 2 add5: make-adder 5 add add2 10 add5 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 27) {
		t.Error("Expected result value 27")
	}
}

func Test_Fnc_make_adder_3_closure(t *testing.T) { // two keys, int as string, pass true
	input := "{ closure: fn { a b } { fnc a parent-context b } make-adder: fn { b } { closure { a } { add a b } } add3: make-adder 3 add5: make-adder 5 add add3 10 add5 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 28) {
		t.Error("Expected result value 28")
	}
}

func Test_Fnc_make_adder_4_in_context(t *testing.T) { // two keys, int as string, pass true
	input := "{ closure: fn { a b } { fnc a parent-context b } adder: fn { b } { closure { a } { add a b } } api: context { add9: adder 9 add5: adder 5 } add api/add9 10 api/add5 10 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 34) {
		t.Error("Expected result value 34")
	}
}

// { a: { 101 102 103 } b: a .nth 1 |add 100 }" // POP doesn't make sense right now as builtins can't change objects now. If is this good / safety or bad /too restricting

// MAP FILTER SEEK

func Test_hofs_map_1(t *testing.T) { // two keys, int as string, pass true
	input := "{ a: { 1 4 8 } map a { .inc } |nth 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 2) {
		t.Error("Expected result value 2")
	}
}

func Test_hofs_map_2(t *testing.T) { // two keys, int as string, pass true
	input := "{ a: { 1 4 8 } map a add 10 _ |nth 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 11) {
		t.Error("Expected result value 11")
	}
}

func Test_hofs_filter_1(t *testing.T) { // two keys, int as string, pass true
	input := "{ a: { 1 4 8 } filter a { .greater 5 } |nth 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 8) {
		t.Error("Expected result value 8")
	}
}

func Test_hofs_filter_2(t *testing.T) { // two keys, int as string, pass true
	input := "{ a: { 1 4 8 } filter a lesser _ 5 |nth 1 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 1) {
		t.Error("Expected result value 1")
	}
}

func Test_hofs_seek_1(t *testing.T) { // two keys, int as string, pass true
	input := "{ a: { 1 4 8 } seek a { .greater 5 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 8) {
		t.Error("Expected result value 8")
	}
}

func Test_hofs_seek_2(t *testing.T) { // two keys, int as string, pass true
	input := "{ a: { 1 4 8 } seek a greater _ 3 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	RegisterBuiltins(es)

	EvalBlock(es)

	fmt.Print(es.Res.Inspect(*es.Idx))
	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Error")
	}

	if !(es.Res.(env.Integer).Value == 4) {
		t.Error("Expected result value 4")
	}
}
