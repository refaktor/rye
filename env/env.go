package env

import (
	"fmt"
	"strconv"
)

/* type Envi interface {
	Get(word int) (Object, bool)
	Set(word int, val Object) Object
} */

type Env struct {
	state  map[int]Object
	parent *Env
}

func NewEnv(par *Env) *Env {
	var e Env
	e.state = make(map[int]Object)
	e.parent = par
	return &e
}

func (e *Env) Probe(idxs Idxs) {
	fmt.Print("<ENV State: ")
	for k, v := range e.state {
		fmt.Print(strconv.FormatInt(int64(k), 10) + ": " + v.Inspect(idxs) + " ")
	}
	fmt.Println(">")
}

/*func (e *Env) Get(word int) (*Object, bool) {
	obj, exists := e.state[word]
	// recursively look at outer Environments ...
	// only specific functions should do this and ounly for function values ... but there is only global env maybe
	// this is simple environment setup, but we will for the sake of safety and speed change this probably
	// maybe some caching here ... or we could inject functions directly into locked series like some idea was to avoid variable lookup
	if !exists && e.parent != nil {
		par := *e.parent
		obj1, exists1 := par.Get(word)
		if exists1 {
			obj = *obj1
			exists = exists1
		}
	}
	return &obj, exists
}*/

func (e *Env) Get(word int) (Object, bool) {
	obj, exists := e.state[word]
	// recursively look at outer Environments ...
	// only specific functions should do this and ounly for function values ... but there is only global env maybe
	// this is simple environment setup, but we will for the sake of safety and speed change this probably
	// maybe some caching here ... or we could inject functions directly into locked series like some idea was to avoid variable lookup
	if !exists && e.parent != nil {
		par := *e.parent
		obj1, exists1 := par.Get(word)
		if exists1 {
			obj = obj1
			exists = exists1
		}
	}
	return obj, exists
}

func (e *Env) Set(word int, val Object) Object {
	e.state[word] = val
	return val
}

type ProgramState struct {
	Ser  TSeries // current block of code
	Res  Object  // result of expression
	Env  *Env    // Env object ()
	Idx  *Idxs   // Idx object (index of words)
	Args []int   // names of current arguments (indexes of names)
	Gen  *Gen    // map[int]map[int]Object  // list of Generic kinds / code
	Inj  Object  // Injected first value in a block evaluation
}

func NewProgramState(ser TSeries, idx *Idxs) *ProgramState {
	ps := ProgramState{
		ser,
		nil,
		NewEnv(nil),
		idx,
		make([]int, 6),
		NewGen(), //make(map[int]map[int]Object),
		nil,
	}
	return &ps
}

func AddToProgramState(ps *ProgramState, ser TSeries, idx *Idxs) *ProgramState {
	ps.Ser = ser
	ps.Res = nil
	ps.Idx = idx
	//ps.Env
	return ps
}

func SetValue(ps *ProgramState, word string, val Object) {
	idx, found := ps.Idx.GetIndex(word)
	if found {
		ps.Env.Set(idx, val)
	}
}
