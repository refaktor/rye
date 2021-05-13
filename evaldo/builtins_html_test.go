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

func TestHtmlDialect_basic_1(t *testing.T) {
	input := "{ a: 0 \"<html><body><p>p1</p><b>b1</b></body></html>\" |string-reader |do-html { <b> [ a: 100 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.(env.Integer).Value != 100 {
		t.Error("Expected value 100")
	}
}
func TestHtmlDialect_basic_2(t *testing.T) {
	input := "{ a: 0 \"<html><body><p>p1</p><b>b1</b></body></html>\" |string-reader |do-html { <i> [ a: 100 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.(env.Integer).Value != 0 {
		t.Error("Expected value 0")
	}
}
func TestHtmlDialect_nesting_1(t *testing.T) {
	input := "{ a: 0 \"<html><body><b>b1</b><div><b>b2</b></div><b>b3</b></body></html>\" |string-reader |do-html { <b> [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.(env.Integer).Value != 3 {
		t.Error("Expected value 3")
	}
}

func TestHtmlDialect_nesting_2(t *testing.T) {
	input := "{ a: 0 \"<html><body><b>b1</b><div><b>b2</b></div><b>b3</b></body></html>\" |string-reader |do-html { <div> { <b> [ a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.(env.Integer).Value != 1 {
		t.Error("Expected value 1")
	}
}

func TestHtmlDialect_selectors_1(t *testing.T) {
	input := "{ a: 0 \"<html><body><b>b1</b><b id='id1' class='class1'>b2</b><b>b3</b></body></html>\" |string-reader |do-html { <b> .class1 [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	if es.Res.(env.Integer).Value != 1 {
		t.Error("Expected value 1")
	}
}

func TestHtmlDialect_selectors_2(t *testing.T) {
	input := "{ a: 0 \"<html><body><b>b1</b><b id='id1' class='class1'>b2</b><b>b3</b></body></html>\" |string-reader |do-html { <b> :id1 [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	if es.Res.(env.Integer).Value != 1 {
		t.Error("Expected value 1")
	}
}

func TestHtmlDialect_selectors_3(t *testing.T) {
	input := "{ a: 0 \"<html><body><b>b1</b><b id='id1' class='class1'>b2</b><b>b3</b></body></html>\" |string-reader |do-html { <b> :id1 .class1 [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	if es.Res.(env.Integer).Value != 1 {
		t.Error("Expected value 1")
	}
}

func TestHtmlDialect_selectors_4(t *testing.T) {
	input := "{ a: 0 \"<html><body><b>b1</b><b id='id1' class='class1'>b2</b><b>b3</b></body></html>\" |string-reader |do-html { <b> :id1 .class2 [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	if es.Res.(env.Integer).Value != 0 {
		t.Error("Expected value 0")
	}
}

func TestHtmlDialect_selectors_5(t *testing.T) {
	input := "{ a: 0 \"<html><body><b>b1</b><b id='id1' class='class1'>b2</b><b>b3</b></body></html>\" |string-reader |do-html { <b> :id2 .class1 [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	if es.Res.(env.Integer).Value != 0 {
		t.Error("Expected value 0")
	}
}
