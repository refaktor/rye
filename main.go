// Rejy_go_v1 project main.go
package main

import (
	"Rejy_go_v1/env"
	"Rejy_go_v1/evaldo"
	"Rejy_go_v1/loader"

	//"Rejy_go_v1/util"
	//"fmt"
	//"strconv"
	"github.com/pkg/profile"
)

// REJY0 in GoLang

// contrary to JS rejy version, parser here already indexes word names into global word index, not evaluator.
// This means one intermediate step less (one data format less, and one conversion less)

// parser produces a tree of values words and blocks in an array.
// primitive values are stored unboxed, we can do this with series   []interface{}
// complex values are stored as struct { type, index } (words, setwords)
// functions are stored similarly. Probably argument count should be in struct too.

type TagType int
type RjType int
type Series []interface{}

type anyword struct {
	kind RjType
	idx  int
}

type node struct {
	kind  RjType
	value interface{}
}

var CODE []interface{}

func main() {

	//util.PrintHeader()
	defer profile.Start().Stop()

	input := "{ loop 10000000 { add 1 2 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	/*genv := loader.GetIdxs()
	ps := evaldo.ProgramState{}

	// Parse
	loader1 := loader.NewLoader()
	input := "{ 123 word 3 { setword: 23 } end 12 word }"
	val, _ := loader1.ParseAndGetValue(input, nil)
	loader.InspectNode(val)
	evaldo.EvalBlock(ps, val.(env.Object))
	fmt.Println(val)

	genv.Probe()

	fmt.Println(strconv.FormatInt(int64(genv.GetWordCount()), 10))*/

}
