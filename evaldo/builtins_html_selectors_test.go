package evaldo

import (
	"fmt"
	"rye/env"
	"rye/evaldo"
	"rye/loader"
	"testing"
)

func TestHtmlDialect_selectors_1(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><b class='my'>b2</b></div><b class='my'>b3</b>\" |string-reader |do-html { <div> { <b> .my [ a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1 {
		t.Error("wrong result")
	}
}

func TestHtmlDialect_selectors_2(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div><b class='my'>b2</b></div><b class='my'>b3</b>\" |string-reader |do-html { <b> .my [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 2 {
		t.Error("wrong result")
	}
}

func TestHtmlDialect_selectors_2b(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div id='mydiv1'><b class='my1'>b2</b></div><div class='mycls'><b class='my2'>b3</b></div>\" |string-reader |do-html { <div> { <b> .my1 [ a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1 {
		t.Error("wrong result")
	}
}

func TestHtmlDialect_selectors_2c(t *testing.T) {
	input := "{ a: 0 \"<b>b1</b><div class='mydiv1'><b>b2</b></div><div class='mydiv2'><b>b3</b></div>\" |string-reader |do-html { <div> .mydiv1 { <b> [ a: a + 1 ] } } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1 {
		t.Error("wrong result")
	}
}

func TestHtmlDialect_selectors_3(t *testing.T) {
	input := "{ a: 0 \"<b id='myid' class='my'>b1</b><div><b class='my'>b2</b></div><b class='my'>b3</b>\" |string-reader |do-html { <b> :myid [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1 {
		t.Error("wrong result")
	}
}

func TestHtmlDialect_selectors_4(t *testing.T) {
	input := "{ a: 0 \"<b id='myid' class='my'>b1</b><div><b class='my'>b2</b></div><b class='my'>b3</b>\" |string-reader |do-html { <b> .my :myid [ a: a + 1 ] } a }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)

	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)
	if es.Res.(env.Integer).Value != 1 {
		t.Error("wrong result")
	}
}
