package contexts

import (
	"Ryelang/env"
	"Ryelang/evaldo"
	"Ryelang/loader"
	"fmt"

	//	"fmt"
	"testing"
)

// _helpers: closure
// method: validates input, if validation fails returns validation result, else does the block of code and returns the sheet
// sql-method: same, but instead of doing the code it does the block as sql dialect
// later we build context in such way that these methods are only avaliable inside resource

/*
	sql-method: fn { validation block } {
		closure {
			validate input
			|^check { .to-sheet }
			|sql block
		}
	}
	method: fn { rules code } {
		closure { input } {
			validate input rules
			|^check { .to-sheet }
			|do-in code
		}
	}
*/

var Basis = `
	closure: fnc _ ?current-context _
	
	api-base: context {

		onetwo: fn {  } { 
			inc oneone
		}
	
		method-1: fn { rules code } { 
			closure { input } { 
				12345
			}
		}
		
		raw-method: fn { rules code } { 
			closure { input } { 
				validate input rules 
				|do-with code
			}
		}

		method: fn { rules code } { 
			closure { input } { 
				validate>ctx input rules 
				|do-in code
			}
		}

		onethree-1: method-1 { id: optional 0 } {
			10
		}

		onethree-2: method { id: optional 0 } {
			add 1000 id
		}
	}
`

// add onetwo get data 'id

func TestMinimal(t *testing.T) {
	input := "{" + Basis + `
	 
	word-resource: context {
		
	}
	
	api-base/onetwo
	
	` + "}"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}

	if !(es.Res.(env.Integer).Value == 12) {
		t.Error("Expected result value 12")
	}
}

func TestMinimal1(t *testing.T) {
	input := "{" + Basis + `
	 
	word-resource: context {
		
	}
	
	api-base/onethree-1 raw-map { id: 2 }
	
	` + "}"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}

	if !(es.Res.(env.Integer).Value == 12345) {
		t.Error("Expected result value 12345")
	}
}

func TestMinimal2(t *testing.T) {
	input := "{" + Basis + `
	 
	word-resource: context {
		
	}
	
	api-base/onethree-2 raw-map { "id" 33 }
	
	` + "}"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}

	if !(es.Res.(env.Integer).Value == 1033) {
		t.Error("Expected result value 1033")
	}
}

func TestExtend(t *testing.T) {
	input := "{" + Basis + `
	 
	word-resource: extend! api-base {

		addnums: method {
			num1: required integer
			num2: optional 0 integer
		} {
			add num1 num2
		}

	}
		
	word-resource/addnums raw-map { "num1" 10 "num2" 100 }
		
	` + "}"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println(es.Res)

	es.Res.Trace("ERR111")

	if es.Res.Type() != env.IntegerType {
		t.Error("Expected result type Integer")
	}

	if !(es.Res.(env.Integer).Value == 1033) {
		t.Error("Expected result value 1033")
	}
}
