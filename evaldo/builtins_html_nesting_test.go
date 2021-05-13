package evaldo

import (
	"fmt"
	"rye/env"
	"rye/evaldo"
	"rye/loader"
	"testing"
)

//
// Literal values
//

func TestHtmlDialect_nesting_2(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><b>b2</b></div><b>b3</b>\" |string-reader |do-html { <div> { <b> [ print \"%%%%\" a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1 {
		t.Error("Expected value 1")
	}
}

func TestHtmlDialect_nesting_3(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><b>b2</b></div><div><b>b3</b></div><b>b4</b4>\" |string-reader |do-html { <div> { <b> [ print \"%%%%\" a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 2 {
		t.Error("Expected value 2")
	}
}

func TestHtmlDialect_nesting_4(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><p><b>b2</b></p></div><div><b>b3</b></div><b>b4</b4>\" |string-reader |do-html { <div> { <p> { <b> [ print \"%%%%\" a: a + 1 ] } } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1 {
		t.Error("Expected value 1")
	}
}

func TestHtmlDialect_nesting_5(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><p><b>b2</b><i>not this</i><b>this too</b></p></div><div><b>b3</b></div><b>b4</b4>\" |string-reader |do-html { <div> { <p> { <b> [ print \"%%%%\" a: a + 1 ] } } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 2 {
		t.Error("Not correct")
	}
}

func TestHtmlDialect_nesting_start_1(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><b>b2</b></div><b>b3</b>\" |string-reader |do-html { [ a: a + 100 ] <div> { <b> [ a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 101 {
		t.Error("Not correct")
	}
}

func TestHtmlDialect_nesting_start_2(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><b>b2</b></div><b>b3</b>\" |string-reader |do-html { [ a: a + 1000 ] <div> { [ a: a + 100 ] <b> [ print \"%%%%\" a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1101 {
		t.Error("Not correct")
	}
}

func TestHtmlDialect_nesting_multipar(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><p><b>b2</b><i>not this</i><b>this too</b></p></div><div><b>b3</b></div><b>b4</b4>\" |string-reader |do-html { <div> { <p> { <b> , <i> [ print \"%%%%\" a: a + 1 ] } } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 3 {
		t.Error("Not correct")
	}
}

func TestHtmlDialect_nesting_multipar_2(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><i>and this</i></div><div><p><b>b2</b><i>not this</i><b>this too</b></p></div><div><b>b3</b></div><b>b4</b4>\" |string-reader |do-html { <div> { <p> { <b> , <i> [ print \"%%%%\" a: a + 1 ] } } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 3 {
		t.Error("Not correct")
	}
}

func TestHtmlDialect_nesting_multipar_3(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><p><i>and this</i></p></div><div><p><b>b2</b><i>not this</i><b>this too</b></p></div><div><b>b3</b></div><b>b4</b4>\" |string-reader |do-html { <div> { <p> { <b> , <i> [ print \"%%%%\" a: a + 1 ] } } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 4 {
		t.Error("Not correct")
	}
}
