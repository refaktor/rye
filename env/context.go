package env

import (
	"fmt"
	"sort"
	"strings"
)

// Context represents a unified interface for all context types (RyeCtx, PersistentCtx, etc.)
type Context interface {
	Object // Embed the existing Object interface

	// Core context operations
	Get(word int) (Object, bool)
	GetCurrent(word int) (Object, bool) // Gets only from current context, not parent chain
	Get2(word int) (Object, bool, Context)
	Set(word int, val Object) Object
	Mod(word int, val Object) bool
	Unset(word int, idxs *Idxs) Object
	SetNew(word int, val Object, idxs *Idxs) bool

	// Variable tracking
	MarkAsVariable(word int)
	IsVariable(word int) bool

	// Context hierarchy
	GetParent() Context
	SetParent(parent Context)

	// Context management
	Copy() Context
	Clear()
	GetState() map[int]Object

	// Utility methods
	GetWords(idxs Idxs) Block
	GetWordsAsStrings(idxs Idxs) Block
	Preview(idxs Idxs, filter string) string
	DumpBare(e Idxs) string

	// Context-specific fields (for compatibility)
	GetDoc() string
	SetDoc(doc string)
	GetKindWord() Word
	SetKindWord(kind Word)

	// Conversion methods for backward compatibility
	AsRyeCtx() *RyeCtx
}

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
	state     map[int]Object
	varFlags  map[int]bool    // Tracks which words are variables
	observers map[int][]Block // Observers for variable changes: word index -> observer blocks
	Parent    *RyeCtx
	Kind      Word
	Doc       string
	locked    bool
	IsClosure bool // Marks contexts captured by closures - should not be pooled
}

func NewEnv(par *RyeCtx) *RyeCtx {
	var e RyeCtx
	e.state = make(map[int]Object)
	e.varFlags = make(map[int]bool)
	e.observers = make(map[int][]Block)
	e.Parent = par
	return &e
}

func NewEnv2(par *RyeCtx, doc string) *RyeCtx {
	var e RyeCtx
	e.state = make(map[int]Object)
	e.varFlags = make(map[int]bool)
	e.observers = make(map[int][]Block)
	e.Parent = par
	e.Doc = doc
	return &e
}

func (e *RyeCtx) isContextOrParent(ctx *RyeCtx) bool {
	if ctx == nil {
		return false
	}
	// Check if ctx is the same as e or if ctx is a child of e
	// This means we check if e is in the parent chain of ctx
	current := ctx
	for current != nil {
		if current == e {
			return true
		}
		current = current.Parent
	}
	return false
}

func (e *RyeCtx) Copy() Context {
	nc := NewEnv(e.Parent)
	cp := make(map[int]Object)
	for k, v := range e.state {
		// move contexts of functions (closures) to this new context
		if fn, ok := v.(Function); ok {
			// fmt.Printf("**** Function found, fn.Ctx=%p, e=%p\n", fn.Ctx, e)
			// fmt.Printf("**** fn.Ctx.Parent=%p, e.Parent=%p\n", fn.Ctx.Parent, e.Parent)
			// Check if the function's context should be updated to the new context
			// This handles cases where the function was created with 'current' or similar
			// We should update the function's context if:
			// 1. It's exactly the same context (fn.Ctx == e)
			// 2. The function's context is the same as the context being copied (they have the same parent)
			// 3. The function's context is a child of the context being copied
			if fn.Ctx == e ||
				e.isContextOrParent(fn.Ctx) ||
				(fn.Ctx != nil &&
					fn.Ctx.Parent != nil &&
					e.Parent != nil &&
					fn.Ctx.Parent == e.Parent &&
					fn.Ctx != e.Parent) {
				//					fmt.Println("CTX IS THE SAME OR RELATED ... COPY")
				fn.Ctx = nc
				cp[k] = fn // store the modified function
			} else {
				// fmt.Println("CTX IS DIFFERENT - NOT COPYING")
				cp[k] = v // store the original function
			}
		} else {
			cp[k] = v
		}
	}
	cpVarFlags := make(map[int]bool)
	for k, v := range e.varFlags {
		cpVarFlags[k] = v
	}
	cpObservers := make(map[int][]Block)
	for k, v := range e.observers {
		// Make a copy of the observer slice
		observersCopy := make([]Block, len(v))
		copy(observersCopy, v)
		cpObservers[k] = observersCopy
	}
	nc.state = cp
	nc.varFlags = cpVarFlags
	nc.observers = cpObservers
	nc.Kind = e.Kind
	nc.locked = e.locked
	nc.IsClosure = e.IsClosure
	return nc
}

// DeepCopy creates a deep copy of the RyeCtx using deep copying for all objects
func (e *RyeCtx) DeepCopy() Context {
	nc := NewEnv(e.Parent)
	cp := make(map[int]Object)
	for k, v := range e.state {
		// Use deep copying for all objects
		cp[k] = DeepCopyObject(v)
	}
	cpVarFlags := make(map[int]bool)
	for k, v := range e.varFlags {
		cpVarFlags[k] = v
	}
	cpObservers := make(map[int][]Block)
	for k, v := range e.observers {
		// Make a copy of the observer slice with deep copied blocks
		observersCopy := make([]Block, len(v))
		for i, block := range v {
			observersCopy[i] = DeepCopyObject(block).(Block)
		}
		cpObservers[k] = observersCopy
	}
	nc.state = cp
	nc.varFlags = cpVarFlags
	nc.observers = cpObservers
	nc.Kind = e.Kind
	nc.locked = e.locked
	nc.IsClosure = e.IsClosure
	return nc
}

func (e *RyeCtx) Clear() {
	clear(e.state)
}

func (e RyeCtx) GetState() map[int]Object {
	return e.state
}

func (e RyeCtx) Print(idxs Idxs) string {
	var bu strings.Builder
	totalWords := len(e.state)
	bu.WriteString(fmt.Sprintf("[Context (%s) \"%s\": %d words - ", e.Kind.Print(idxs), e.Doc, totalWords))

	// Collect keys and sort them for consistent output
	keys := make([]int, 0, len(e.state))
	for k := range e.state {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// Print only first 8 words
	count := 0
	maxWords := 8
	for _, k := range keys {
		if count >= maxWords {
			break
		}
		// v := e.state[k]
		bu.WriteString(idxs.GetWord(k) + ", ")
		//ctx, ok := v.(RyeCtx)
		// if ok || &ctx == &e {
		//	bu.WriteString(" [self reference] ")
		// } else {
		//	bu.WriteString(v.Inspect(idxs) + " ")
		// }
		count++
	}

	if totalWords > maxWords {
		bu.WriteString("...")
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
			switch idxs.GetWord(int(v.Type())) {
			case "builtin":
				color = color_word2
			case "context":
				color = color_num2
			case "function":
				color = color_word
			default:
				color = color_string2
			}
			var strVal string
			ctx, ok := v.(RyeCtx)
			if ok && &ctx == &e {
				strVal = " [self reference]"
			} else {
				strVal = v.Inspect(idxs)
			}
			arr = append(arr, str1+" "+reset+color_comment+strVal+reset+"|||"+color) // idxs.GetWord(v.GetKind()
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

// TODO -- unify these previews
func (e RyeCtx) PreviewByType(idxs Idxs, typeFilter string) string {
	var bu strings.Builder
	var ks string
	if e.GetKind() > 0 {
		ks = " (" + e.Kind.Print(idxs) + ") "
	}
	bu.WriteString("Context" + ks + " (filtered by type: " + typeFilter + "):")
	if e.Doc > "" {
		bu.WriteString("\n\r\"" + e.Doc + "\"")
	}
	arr := make([]string, 0)
	i := 0
	for k, v := range e.state {
		objectType := idxs.GetWord(int(v.Type()))
		if objectType == typeFilter {
			str1 := idxs.GetWord(k)
			var color string
			switch objectType {
			case "builtin":
				color = color_word2
			case "context":
				color = color_num2
			case "function":
				color = color_word
			default:
				color = color_string2
			}
			var strVal string
			ctx, ok := v.(RyeCtx)
			if ok && &ctx == &e {
				strVal = " [self reference]"
			} else {
				strVal = v.Inspect(idxs)
			}
			arr = append(arr, str1+" "+reset+color_comment+strVal+reset+"|||"+color)
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

func (e RyeCtx) PreviewByRegex(idxs Idxs, regexFilter interface{ MatchString(string) bool }) string {
	var bu strings.Builder
	var ks string
	if e.GetKind() > 0 {
		ks = " (" + e.Kind.Print(idxs) + ") "
	}
	bu.WriteString("Context" + ks + " (filtered by regex):")
	if e.Doc > "" {
		bu.WriteString("\n\r\"" + e.Doc + "\"")
	}
	arr := make([]string, 0)
	i := 0
	for k, v := range e.state {
		str1 := idxs.GetWord(k)
		if regexFilter.MatchString(str1) {
			var color string
			switch idxs.GetWord(int(v.Type())) {
			case "builtin":
				color = color_word2
			case "context":
				color = color_num2
			case "function":
				color = color_word
			default:
				color = color_string2
			}
			var strVal string
			ctx, ok := v.(RyeCtx)
			if ok && &ctx == &e {
				strVal = " [self reference]"
			} else {
				strVal = v.Inspect(idxs)
			}
			arr = append(arr, str1+" "+reset+color_comment+strVal+reset+"|||"+color)
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
	return ContextType
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

func (e RyeCtx) GetWordsAsStrings(idxs Idxs) Block {
	objs := make([]Object, len(e.state))
	idx := 0
	for k := range e.state {
		objs[idx] = *NewString(idxs.GetWord(k))
		idx += 1
	}
	return *NewBlock(*NewTSeries(objs))
}

func (e RyeCtx) GetWords(idxs Idxs) Block {
	objs := make([]Object, len(e.state))
	idx := 0
	for k := range e.state {
		objs[idx] = *NewWord(k)
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

func (e *RyeCtx) GetCurrent(word int) (Object, bool) {
	obj, exists := e.state[word]
	return obj, exists
}

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

func (e *RyeCtx) Get2(word int) (Object, bool, Context) {
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
			return obj, exists, ctx
		}
	}
	return obj, exists, e
}

func (e *RyeCtx) Set(word int, val Object) Object {
	if _, exists := e.state[word]; exists {
		return *NewError("Can't set already set word, try using modword! FIXME !")
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

// Mark a word as a variable
func (e *RyeCtx) MarkAsVariable(word int) {
	e.varFlags[word] = true
}

// Check if a word is a variable
func (e *RyeCtx) IsVariable(word int) bool {
	isVar, exists := e.varFlags[word]
	if exists && isVar {
		return true
	}
	// Not a variable in this context
	return false
}

// ModResult represents the result of a Mod operation
type ModResult int

const (
	ModOK              ModResult = 0 // Modification succeeded
	ModErrConstant     ModResult = 1 // Word is a constant, cannot be modified
	ModErrTypeMismatch ModResult = 2 // Type mismatch between existing and new value
)

func (e *RyeCtx) Mod(word int, val Object) bool {
	result, _ := e.ModWithInfo(word, val)
	return result == ModOK
}

// ModWithInfo modifies a word and returns detailed information about the result
// Returns (ModResult, existingType) where existingType is the type of the existing value if there's a mismatch
func (e *RyeCtx) ModWithInfo(word int, val Object) (ModResult, Type) {
	if existingVal, exists := e.state[word]; exists {
		// Word exists, check if it's a variable
		if !e.IsVariable(word) {
			// Cannot modify constants
			return ModErrConstant, 0
		}
		// Type check - compare Object.Type()
		if existingVal.Type() != val.Type() {
			return ModErrTypeMismatch, existingVal.Type()
		}
	} else {
		// Word doesn't exist, create it as a variable
		e.MarkAsVariable(word)
	}
	e.state[word] = val
	return ModOK, 0
}

// GetParent returns the parent context
func (e *RyeCtx) GetParent() Context {
	if e.Parent == nil {
		return nil
	}
	return e.Parent
}

// SetParent sets the parent context
func (e *RyeCtx) SetParent(parent Context) {
	if parent == nil {
		e.Parent = nil
	} else {
		// Type assert to *RyeCtx since that's what we're currently using
		if ryeCtx, ok := parent.(*RyeCtx); ok {
			e.Parent = ryeCtx
		}
	}
}

// GetDoc returns the documentation string
func (e *RyeCtx) GetDoc() string {
	return e.Doc
}

// SetDoc sets the documentation string
func (e *RyeCtx) SetDoc(doc string) {
	e.Doc = doc
}

// GetKindWord returns the Kind as a Word
func (e *RyeCtx) GetKindWord() Word {
	return e.Kind
}

// SetKindWord sets the Kind
func (e *RyeCtx) SetKindWord(kind Word) {
	e.Kind = kind
}

// AsRyeCtx returns the context as a *RyeCtx for backward compatibility
func (e *RyeCtx) AsRyeCtx() *RyeCtx {
	return e
}

// Observer management methods for context-level observers

// AddObserver registers an observer block for a specific word in this context
func (e *RyeCtx) AddObserver(wordIndex int, observerBlock Block) {
	e.observers[wordIndex] = append(e.observers[wordIndex], observerBlock)
}

// HasObservers checks if there are any observers for a word in this context
func (e *RyeCtx) HasObservers(wordIndex int) bool {
	observers, exists := e.observers[wordIndex]
	return exists && len(observers) > 0
}

// GetObservers returns a copy of observers for a word in this context
func (e *RyeCtx) GetObservers(wordIndex int) []Block {
	observers, exists := e.observers[wordIndex]
	if !exists || len(observers) == 0 {
		return nil
	}
	// Return a copy to avoid race conditions
	result := make([]Block, len(observers))
	copy(result, observers)
	return result
}

// RemoveObserver removes a specific observer from this context (for potential future use)
func (e *RyeCtx) RemoveObserver(wordIndex int, observerBlock Block) {
	observers, exists := e.observers[wordIndex]
	if !exists {
		return
	}

	// Remove the observer from the list
	for i, obs := range observers {
		if obs.Equal(observerBlock) {
			// Remove by swapping with last element and truncating
			observers[i] = observers[len(observers)-1]
			e.observers[wordIndex] = observers[:len(observers)-1]
			break
		}
	}

	// Clean up empty slices
	if len(e.observers[wordIndex]) == 0 {
		delete(e.observers, wordIndex)
	}
}

// LocationNode represents a source location marker in the code
// These nodes are inserted during parsing and ignored during evaluation
// They're only used for error reporting to find the nearest source location
type LocationNode struct {
	Filename   string
	Line       int
	Column     int
	SourceLine string // The actual line of source code for better error display
}

// Implement Object interface for LocationNode
func (ln LocationNode) Type() Type { return LocationNodeType }
func (ln LocationNode) Inspect(idxs Idxs) string {
	return fmt.Sprintf("LocationNode{%s:%d:%d}", ln.Filename, ln.Line, ln.Column)
}
func (ln LocationNode) Print(idxs Idxs) string {
	return ln.Inspect(idxs)
}
func (ln LocationNode) Trace(msg string) {
	fmt.Printf("%s LocationNode: %s:%d:%d\n", msg, ln.Filename, ln.Line, ln.Column)
}
func (ln LocationNode) GetKind() int { return 0 }
func (ln LocationNode) Equal(o Object) bool {
	if other, ok := o.(LocationNode); ok {
		return ln.Filename == other.Filename && ln.Line == other.Line && ln.Column == other.Column
	}
	return false
}
func (ln LocationNode) Dump(idxs Idxs) string {
	return ln.Inspect(idxs)
}

// String returns a formatted representation of the location
func (ln LocationNode) String() string {
	if ln.Filename == "" {
		return fmt.Sprintf("line %d, column %d", ln.Line, ln.Column)
	}
	return fmt.Sprintf("%s:%d:%d", ln.Filename, ln.Line, ln.Column)
}

// NewLocationNode creates a new LocationNode
func NewLocationNode(filename string, line, column int, sourceLine string) *LocationNode {
	return &LocationNode{
		Filename:   filename,
		Line:       line,
		Column:     column,
		SourceLine: sourceLine,
	}
}

type ProgramState struct {
	Ser           TSeries // current block of code
	Res           Object  // result of expression
	Ctx           *RyeCtx // Env object ()
	PCtx          *RyeCtx // Env object () -- pure countext
	Idx           *Idxs   // Idx object (index of words)
	Args          []int   // names of current arguments (indexes of names)
	Gen           *Gen    // map[int]map[int]Object  // list of Generic kinds / code
	Inj           Object  // Injected first value in a block evaluation
	Injnow        bool
	ReturnFlag    bool
	ErrorFlag     bool
	FailureFlag   bool
	InterruptFlag bool // Flag set when signal interruption is requested (Ctrl+C, Ctrl+Z)
	ForcedResult  Object
	SkipFlag      bool
	InErrHandler  bool
	ScriptPath    string // holds the path to the script that is being imported (doed) currently
	WorkingPath   string // holds the path to CWD (can be changed in program with specific functions)
	AllowMod      bool
	LiveObj       *LiveEnv
	Dialect       DoDialect
	Stack         *EyrStack
	Embedded      bool
	DeferBlocks   []Block   // blocks to be executed when function exits or program terminates
	ContextStack  []*RyeCtx // stack of previous contexts for ccb navigation
	// LastFailedCPathInfo map[string]interface{} // stores information about the last failed context path
	BlockFile string
	BlockLine int
}

type DoDialect int

const (
	Rye2Dialect  DoDialect = 1
	EyrDialect   DoDialect = 2
	Rye0Dialect  DoDialect = 3
	Rye00Dialect DoDialect = 4 // Simplified dialect for builtins and integers
)

func NewProgramState(ser TSeries, idx *Idxs) *ProgramState {
	ps := ProgramState{
		Ser:           ser,
		Res:           nil,
		Ctx:           NewEnv(nil),
		PCtx:          NewEnv(nil),
		Idx:           idx,
		Args:          make([]int, 6),
		Gen:           NewGen(),
		Inj:           nil,
		Injnow:        false,
		ReturnFlag:    false,
		ErrorFlag:     false,
		FailureFlag:   false,
		InterruptFlag: false,
		ForcedResult:  nil,
		SkipFlag:      false,
		InErrHandler:  false,
		ScriptPath:    "",
		WorkingPath:   "",
		AllowMod:      false,
		LiveObj:       nil,
		Dialect:       Rye2Dialect,
		Stack:         NewEyrStack(),
		Embedded:      false,
		DeferBlocks:   make([]Block, 0),
		ContextStack:  make([]*RyeCtx, 0),
		BlockFile:     "",
		BlockLine:     -1,
	}
	return &ps
}

func NewProgramStateNEW() *ProgramState {
	ps := ProgramState{
		Ser:           *NewTSeries(make([]Object, 0)),
		Res:           nil,
		Ctx:           NewEnv(nil),
		PCtx:          NewEnv(nil),
		Idx:           NewIdxs(),
		Args:          make([]int, 6),
		Gen:           NewGen(),
		Inj:           nil,
		Injnow:        false,
		ReturnFlag:    false,
		ErrorFlag:     false,
		FailureFlag:   false,
		InterruptFlag: false,
		ForcedResult:  nil,
		SkipFlag:      false,
		InErrHandler:  false,
		ScriptPath:    "",
		WorkingPath:   "",
		AllowMod:      false,
		LiveObj:       NewLiveEnv(),
		Dialect:       Rye2Dialect,
		Stack:         NewEyrStack(),
		Embedded:      false,
		DeferBlocks:   make([]Block, 0),
		ContextStack:  make([]*RyeCtx, 0),
		BlockFile:     "",
		BlockLine:     -1,
	}
	return &ps
}

func (ps *ProgramState) Dump() string {
	return ps.Ctx.DumpBare(*ps.Idx)
}

func (ps *ProgramState) ResetStack() {
	ps.Stack = NewEyrStack()
}

// PushContext adds current context to the context stack
func (ps *ProgramState) PushContext(ctx *RyeCtx) {
	ps.ContextStack = append(ps.ContextStack, ctx)
}

// PopContext removes and returns the most recent context from the stack
func (ps *ProgramState) PopContext() (*RyeCtx, bool) {
	if len(ps.ContextStack) == 0 {
		return nil, false
	}
	// Pop from end (LIFO - Last In, First Out)
	idx := len(ps.ContextStack) - 1
	ctx := ps.ContextStack[idx]
	ps.ContextStack = ps.ContextStack[:idx]
	return ctx, true
}

// ContextStackSize returns the number of contexts in the stack
func (ps *ProgramState) ContextStackSize() int {
	return len(ps.ContextStack)
}

func AddToProgramState(ps *ProgramState, ser TSeries, idx *Idxs) *ProgramState {
	ps.Ser = ser
	// ps.Res = nil
	ps.Idx = idx
	//ps.Env
	return ps
}

func AddToProgramStateNEWWithLocation(ps *ProgramState, block Block, idx *Idxs) *ProgramState {
	ps.Ser = block.Series
	// ps.Res = nil
	ps.Idx = idx
	ps.BlockFile = block.FileName
	ps.BlockLine = block.Line
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
	x := s.D[s.I-1-offset]
	return x
}
