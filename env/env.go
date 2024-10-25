package env

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

/* type Envi interface {
	Get(word int) (Object, bool)
	Set(word int, val Object) Object
} */

// This is experimental env without map for Functions with up to two variables

type EnvR2 struct {
	Var1   Object
	Var2   Object
	parent *RyeCtx
	kind   Word
}

type RyeCtx struct {
	state  map[int]Object
	Parent *RyeCtx
	Kind   Word
	Doc    string
	locked bool
}

func NewEnv(par *RyeCtx) *RyeCtx {
	var e RyeCtx
	e.state = make(map[int]Object)
	e.Parent = par
	return &e
}

func NewEnv2(par *RyeCtx, doc string) *RyeCtx {
	var e RyeCtx
	e.state = make(map[int]Object)
	e.Parent = par
	e.Doc = doc
	return &e
}

func (e RyeCtx) Copy() *RyeCtx {
	nc := NewEnv(e.Parent)
	cp := make(map[int]Object)
	for k, v := range e.state {
		cp[k] = v
	}
	nc.state = cp
	nc.Kind = e.Kind
	nc.locked = e.locked
	return nc
}

func (e RyeCtx) GetState() map[int]Object {
	return e.state
}

func (e RyeCtx) Print(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("[Context (" + e.Kind.Print(idxs) + "): ")
	for k, v := range e.state {
		bu.WriteString(idxs.GetWord(k) + ": " + v.Inspect(idxs) + " ")
	}
	bu.WriteString("]")
	return bu.String()
}

const reset = "\x1b[0m"
const reset2 = "\033[39;49m"

const color_word = "\x1b[38;5;45m"
const color_word2 = "\033[38;5;214m"
const color_num2 = "\033[38;5;202m"
const color_string2 = "\033[38;5;148m"
const color_comment = "\033[38;5;247m"

func (e RyeCtx) Preview(idxs Idxs, filter string) string {
	var bu strings.Builder
	var ks string
	if e.GetKind() > 0 {
		ks = " (" + e.Kind.Print(idxs) + ") "
	}
	bu.WriteString("Context" + ks + ":")
	if e.Doc > "" {
		bu.WriteString("\n\r\"" + e.Doc + "\"")
	}
	arr := make([]string, 0)
	i := 0
	for k, v := range e.state {
		str1 := idxs.GetWord(k)
		if strings.Contains(str1, filter) {
			var color string
			switch idxs.GetWord(v.GetKind()) {
			case "builtin":
				color = color_word2
			case "context":
				color = color_num2
			case "function":
				color = color_word
			default:
				color = color_string2
			}
			arr = append(arr, str1+": "+reset+color_comment+v.Inspect(idxs)+reset+"|||"+color) // idxs.GetWord(v.GetKind()
			// bu.WriteString(" " + idxs.GetWord(k) + ": " + v.Inspect(idxs) + "\n")
			i += 1
		}
	}
	sort.Strings(arr)
	for aa := range arr {
		line := arr[aa]
		pars := strings.Split(line, "|||")
		bu.WriteString("\n\r " + pars[1] + pars[0])
	}
	return bu.String()
}

// Type returns the type of the Integer.
func (i RyeCtx) Type() Type {
	return CtxType
}

// Inspect returns a string representation of the Integer.
func (i RyeCtx) Inspect(e Idxs) string {
	return i.Print(e)
}

func (i RyeCtx) Trace(msg string) {
	fmt.Print(msg + "(env): ")
	//fmt.Println(i.Value)
}

func (i RyeCtx) GetKind() int {
	return i.Kind.Index
}

func (e RyeCtx) GetWords(idxs Idxs) Block {
	objs := make([]Object, len(e.state))
	idx := 0
	for k := range e.state {
		objs[idx] = String{idxs.GetWord(k)}
		idx += 1
	}
	return *NewBlock(*NewTSeries(objs))
}

func (i RyeCtx) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oCtx := o.(RyeCtx)
	if len(i.state) != len(oCtx.state) {
		return false
	}
	for k, v := range i.state {
		if !v.Equal(oCtx.state[k]) {
			return false
		}
	}
	if i.Parent != oCtx.Parent {
		return false
	}
	if i.Kind != oCtx.Kind {
		return false
	}
	if i.locked != oCtx.locked {
		return false
	}
	return true
}

func (i RyeCtx) Dump(e Idxs) string {
	var bu strings.Builder
	bu.WriteString("context {\n")
	//bu.WriteString(fmt.Sprintf("doc \"%s\"\n", i.Doc))
	bu.WriteString(i.DumpBare(e))
	bu.WriteString("}")
	return bu.String()
}

// DumpBare returns the string representation of the context without wraping it in context { ... }
func (i RyeCtx) DumpBare(e Idxs) string {
	var bu strings.Builder
	for j := 0; j < e.GetWordCount(); j++ {
		if val, ok := i.state[j]; ok {
			if val.Type() != BuiltinType {
				bu.WriteString(fmt.Sprintf("%s: %s\n", e.GetWord(j), val.Dump(e)))
			}
		}
	}
	return bu.String()
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

func (e *RyeCtx) Get(word int) (Object, bool) {
	obj, exists := e.state[word]
	// recursively look at outer Environments ...
	// only specific functions should do this and ounly for function values ... but there is only global env maybe
	// this is simple environment setup, but we will for the sake of safety and speed change this probably
	// maybe some caching here ... or we could inject functions directly into locked series like some idea was to avoid variable lookup
	if !exists && e.Parent != nil {
		par := *e.Parent
		obj1, exists1 := par.Get(word)
		if exists1 {
			obj = obj1
			exists = exists1
		}
	}
	return obj, exists
}

func (e *RyeCtx) Get2(word int) (Object, bool, *RyeCtx) {
	obj, exists := e.state[word]
	// recursively look at outer Environments ...
	// only specific functions should do this and ounly for function values ... but there is only global env maybe
	// this is simple environment setup, but we will for the sake of safety and speed change this probably
	// maybe some caching here ... or we could inject functions directly into locked series like some idea was to avoid variable lookup
	if !exists && e.Parent != nil {
		par := *e.Parent
		obj1, exists1, ctx := par.Get2(word)
		if exists1 {
			obj = obj1
			exists = exists1
			e = ctx
		}
	}
	return obj, exists, e
}

func (e *RyeCtx) Set(word int, val Object) Object {
	if _, exists := e.state[word]; exists {
		return NewError("Can't set already set word, try using modword! FIXME !")
	} else {
		e.state[word] = val
		return val
	}
}

func (e *RyeCtx) Unset(word int, idxs *Idxs) Object {
	if _, exists := e.state[word]; !exists {
		return *NewError("Can't unset non-existing word " + idxs.GetWord(word) + " in this context")
	} else {
		delete(e.state, word)
		return NewInteger(1)
	}
}

func (e *RyeCtx) SetNew(word int, val Object, idxs *Idxs) bool {
	if _, exists := e.state[word]; exists {
		return false
	} else {
		e.state[word] = val
		return true
	}
}

func (e *RyeCtx) Mod(word int, val Object) bool {
	// if _, exists := e.state[word]; exists {
	e.state[word] = val
	return true
	// } else {
	//	return NewError("Can't mod an unset word " + idxs.GetWord(word) + ", try using setword")
	// }
}

type ProgramState struct {
	Ser          TSeries // current block of code
	Res          Object  // result of expression
	Ctx          *RyeCtx // Env object ()
	PCtx         *RyeCtx // Env object () -- pure countext
	Idx          *Idxs   // Idx object (index of words)
	Args         []int   // names of current arguments (indexes of names)
	Gen          *Gen    // map[int]map[int]Object  // list of Generic kinds / code
	Inj          Object  // Injected first value in a block evaluation
	Injnow       bool
	ReturnFlag   bool
	ErrorFlag    bool
	FailureFlag  bool
	ForcedResult Object
	SkipFlag     bool
	InErrHandler bool
	ScriptPath   string // holds the path to the script that is being imported (doed) currently
	WorkingPath  string // holds the path to CWD (can be changed in program with specific functions)
	AllowMod     bool
	LiveObj      *LiveEnv
	Dialect      DoDialect
	Stack        *EyrStack
	Embedded     bool
}

type DoDialect int

const (
	Rye2Dialect DoDialect = 1
	EyrDialect  DoDialect = 2
)

func NewProgramState(ser TSeries, idx *Idxs) *ProgramState {
	ps := ProgramState{
		ser,
		nil,
		NewEnv(nil),
		NewEnv(nil),
		idx,
		make([]int, 6),
		NewGen(), //make(map[int]map[int]Object),
		nil,
		false,
		false,
		false,
		false,
		nil,
		false,
		false,
		"",
		"",
		false,
		nil, // NewLiveEnv(),
		Rye2Dialect,
		NewEyrStack(),
		false,
	}
	return &ps
}

func NewProgramStateNEW() *ProgramState {
	ps := ProgramState{
		*NewTSeries(make([]Object, 0)),
		nil,
		NewEnv(nil),
		NewEnv(nil),
		NewIdxs(),
		make([]int, 6),
		NewGen(), //make(map[int]map[int]Object),
		nil,
		false,
		false,
		false,
		false,
		nil,
		false,
		false,
		"",
		"",
		false,
		NewLiveEnv(),
		Rye2Dialect,
		NewEyrStack(),
		false,
	}
	return &ps
}

func (ps *ProgramState) Dump() string {
	return ps.Ctx.DumpBare(*ps.Idx)
}

func (ps *ProgramState) ResetStack() {
	ps.Stack = NewEyrStack()
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
		ps.Ctx.SetNew(idx, val, ps.Idx)
		switch valf := val.(type) {
		case Function:
			if valf.Pure {
				ps.PCtx.SetNew(idx, val, ps.Idx)
			}
		}
	}
}

// LiveEnv -- a experiment in live realoading

type LiveEnv struct {
	Active  bool
	Watcher *fsnotify.Watcher
	PsMutex sync.Mutex
	Updates []string
}

func NewLiveEnv() *LiveEnv {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err) // TODO -- if this fails show error in red, but make it so that rye runs anyway (check if null at repl for starters)
		return nil
	}

	// defer watcher.Close()

	// Watch current directory for changes in any Go source file (*.go)

	liveEnv := &LiveEnv{true, watcher, sync.Mutex{}, make([]string, 0)}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					// fmt.Println("LiveEnv file changed:", event.Name)
					liveEnv.PsMutex.Lock()
					liveEnv.Updates = append(liveEnv.Updates, event.Name)
					liveEnv.PsMutex.Unlock()
				}
			case err := <-watcher.Errors:
				fmt.Println("LiveEnv error watching files:", err)
			}
		}
	}()

	return liveEnv
}

func (le *LiveEnv) Add(file string) {
	err := le.Watcher.Add(".")
	if err != nil {
		fmt.Println("LiveEnv: Error adding directory to watch:", err)
	}
}

func (le *LiveEnv) ClearUpdates() {
	le.Updates = make([]string, 0)
}

const STACK_SIZE int = 1000

type EyrStack struct {
	D []Object
	I int
}

func NewEyrStack() *EyrStack {
	st := EyrStack{}
	st.D = make([]Object, STACK_SIZE)
	st.I = 0
	return &st
}

// IsEmpty checks if our stack is empty.
func (s *EyrStack) IsEmpty() bool {
	return s.I == 0
}

// Push adds a new number to the stack
func (s *EyrStack) Push(es *ProgramState, x Object) {
	//// *s = append(*s, x)
	if s.I+1 >= STACK_SIZE {
		es.ErrorFlag = true
		es.Res = NewError("stack overflow (maxed)")
		return
	}
	s.D[s.I] = x
	s.I++
	// appending takes a lot of time .. pushing values ...
}

// Pop removes and returns the top element of stack.
func (s *EyrStack) Pop(es *ProgramState) Object {
	if s.IsEmpty() {
		es.ErrorFlag = true
		es.Res = NewError("stack underflow (empty)")
		return es.Res
	}
	s.I--
	x := s.D[s.I]
	return x
}

// Pop removes and returns the top element of stack.
func (s *EyrStack) Peek(es *ProgramState, offset int) Object {
	if s.IsEmpty() {
		es.ErrorFlag = true
		es.Res = NewError("stack underflow (empty 2)")
		return es.Res
	}
	if s.I-offset < 0 {
		es.ErrorFlag = true
		es.Res = NewError("stack underflow (offset)")
		return es.Res
	}
	x := s.D[s.I-offset]
	return x
}
