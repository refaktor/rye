// Package embed provides a lightweight, dependency-minimal API for embedding
// the Rye scripting language in Go applications.
//
// # Build tags
//
// This sub-module is always built with the two tags below, which exclude
// optional heavy batteries from the dependency graph:
//
//	no_baseio  – exclude OS/filesystem/terminal I/O builtins and their deps
//	             (fsnotify, keyboard, seccomp, landlock, yaml, …)
//	no_vector  – exclude govector / primes math extensions
//
// The only external dependencies added to a consuming project are:
//
//	golang.org/x/text    – unicode case-folding (builtins_base_strings)
//	golang.org/x/crypto  – PBKDF2 (util/securesave)
//
// # Quick start
//
//	engine := embed.New()
//
//	engine.RegisterBuiltin("double", 1, func(ps *env.ProgramState, a0, _, _, _, _ env.Object) env.Object {
//	    return *env.NewInteger(a0.(env.Integer).Value * 2)
//	})
//
//	result, err := engine.Eval(`double 21`)
//	fmt.Println(result) // 42
package embed

import (
	"errors"
	"fmt"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
)

// BuiltinFn is the Go function signature required by Rye builtins.
// ps is the running program state; a0–a4 are positional arguments.
// Unused argument slots receive env.Void{}.
type BuiltinFn = env.BuiltinFunction

// Engine is a single Rye interpreter instance with its own word index and
// execution context.  Create one with [New] or [NewBlank].
type Engine struct {
	ps *env.ProgramState
}

// New creates an Engine pre-loaded with all base Rye builtins
// (arithmetic, strings, collections, control flow, …) but WITHOUT OS-level
// I/O (no file access, no shell commands, no os.Exit, no os.Args).
// This keeps the embedded sandbox safe by default.
// If you need file/shell access in the embedded engine, call
// evaldo.RegisterBaseIOBuiltins(engine.ProgramState()) after New().
// Register additional Go functions with [Engine.RegisterBuiltin].
func New() *Engine {
	block, idxs := loader.LoadStringNoPEG("", false)
	ps := env.NewProgramStateOLD(block.(env.Block).Series, idxs)
	evaldo.RegisterBaseBuiltins(ps)
	return &Engine{ps: ps}
}

// NewBlank creates an Engine with NO pre-registered builtins.
// Useful when you want complete control over the available vocabulary.
func NewBlank() *Engine {
	block, idxs := loader.LoadStringNoPEG("", false)
	ps := env.NewProgramStateOLD(block.(env.Block).Series, idxs)
	return &Engine{ps: ps}
}

// ProgramState returns the underlying *env.ProgramState for advanced
// integration (e.g. reading back Rye values, inspecting the word index,
// registering context-aware builtins manually).
func (e *Engine) ProgramState() *env.ProgramState {
	return e.ps
}

// ---------------------------------------------------------------------------
// Builtin registration
// ---------------------------------------------------------------------------

// RegisterBuiltin registers a Go function as a Rye builtin word.
//
//   - name     – the Rye word (e.g. "greet", "db-query")
//   - argCount – number of arguments the function accepts (0–5)
//   - fn       – the Go implementation (signature: BuiltinFn)
func (e *Engine) RegisterBuiltin(name string, argCount int, fn BuiltinFn) {
	idx := e.ps.Idx.IndexWord(name)
	b := env.NewBuiltin(fn, argCount, false, false, "")
	e.ps.Ctx.Set(idx, *b)
}

// RegisterBuiltinDoc is like [Engine.RegisterBuiltin] but also attaches a
// documentation string visible via the Rye `doc` word.
func (e *Engine) RegisterBuiltinDoc(name string, argCount int, doc string, fn BuiltinFn) {
	idx := e.ps.Idx.IndexWord(name)
	b := env.NewBuiltin(fn, argCount, false, false, doc)
	e.ps.Ctx.Set(idx, *b)
}

// ---------------------------------------------------------------------------
// Setting and getting Rye words from Go
// ---------------------------------------------------------------------------

// SetWord injects a Rye word as a constant into the current context.
// The word cannot be modified by Rye scripts (it is not a variable).
// Use [Engine.SetVar] when you need the script to be able to modify the value.
//
// Example:
//
//	engine.SetWord("port", *env.NewInteger(8080))
//	engine.Eval(`print "listening on port" + string port`)
func (e *Engine) SetWord(name string, val env.Object) {
	idx := e.ps.Idx.IndexWord(name)
	e.ps.Ctx.Set(idx, val)
}

// SetVar injects a Rye word as a mutable variable into the current context.
// Unlike [Engine.SetWord], the Rye script can modify this value using the
// modword operator (word::).
//
// Example:
//
//	engine.SetVar("score", *env.NewInteger(0))
//	engine.Eval(`score:: score + 10`)   // score is now 10
//	val, _ := engine.GetWord("score")
func (e *Engine) SetVar(name string, val env.Object) {
	idx := e.ps.Idx.IndexWord(name)
	e.ps.Ctx.SetVar(idx, val)
}

// GetWord retrieves the value of a Rye word from the current context.
// Returns (nil, false) if the word has not been set.
func (e *Engine) GetWord(name string) (env.Object, bool) {
	idx, found := e.ps.Idx.GetIndex(name)
	if !found {
		return nil, false
	}
	obj, exists := e.ps.Ctx.Get(idx)
	if !exists {
		return nil, false
	}
	return obj, true
}

// ---------------------------------------------------------------------------
// Evaluation
// ---------------------------------------------------------------------------

// Eval parses and evaluates Rye source code.
// Returns the result serialised as a human-readable string, or an error if
// parsing or evaluation fails.
func (e *Engine) Eval(code string) (string, error) {
	obj, err := e.EvalGetObject(code)
	if err != nil {
		return "", err
	}
	if obj == nil {
		return "", nil
	}
	return obj.Print(*e.ps.Idx), nil
}

// EvalGetObject parses and evaluates Rye source code and returns the raw
// env.Object result.  Callers can type-assert to env.Integer, env.String,
// env.Block, etc.
func (e *Engine) EvalGetObject(code string) (env.Object, error) {
	block := loader.LoadString(code, false, e.ps)
	if errObj, ok := block.(env.Error); ok {
		return nil, fmt.Errorf("parse error: %s", errObj.Message)
	}
	// Update the series on the existing program state, preserving ctx & idx.
	e.ps.SetBlock(block.(env.Block))
	e.ps.ErrorFlag = false
	e.ps.ReturnFlag = false
	e.ps.FailureFlag = false
	evaldo.Eval(e.ps)
	if e.ps.ErrorFlag {
		return nil, fmt.Errorf("eval error: %s", e.ps.Res.Print(*e.ps.Idx))
	}
	return e.ps.Res, nil
}

// EvalBytes is a convenience wrapper that evaluates a []byte source.
// Useful when the script is read from a file:
//
//	src, _ := os.ReadFile("config.rye")
//	result, err := engine.EvalBytes(src)
func (e *Engine) EvalBytes(src []byte) (string, error) {
	return e.Eval(string(src))
}

// ---------------------------------------------------------------------------
// Typed result helpers
// ---------------------------------------------------------------------------

// EvalString evaluates code and returns the result as a Go string.
// Returns an error if the result is not a Rye String.
func (e *Engine) EvalString(code string) (string, error) {
	obj, err := e.EvalGetObject(code)
	if err != nil {
		return "", err
	}
	s, ok := obj.(env.String)
	if !ok {
		return "", errors.New("result is not a string: " + obj.Print(*e.ps.Idx))
	}
	return s.Value, nil
}

// EvalInteger evaluates code and returns the result as a Go int64.
// Returns an error if the result is not a Rye Integer.
func (e *Engine) EvalInteger(code string) (int64, error) {
	obj, err := e.EvalGetObject(code)
	if err != nil {
		return 0, err
	}
	i, ok := obj.(env.Integer)
	if !ok {
		return 0, errors.New("result is not an integer: " + obj.Print(*e.ps.Idx))
	}
	return i.Value, nil
}

// EvalDecimal evaluates code and returns the result as a Go float64.
// Returns an error if the result is not a Rye Decimal.
func (e *Engine) EvalDecimal(code string) (float64, error) {
	obj, err := e.EvalGetObject(code)
	if err != nil {
		return 0, err
	}
	d, ok := obj.(env.Decimal)
	if !ok {
		return 0, errors.New("result is not a decimal: " + obj.Print(*e.ps.Idx))
	}
	return d.Value, nil
}

// EvalBool evaluates code and returns the result as a Go bool.
// Returns an error if the result is not a Rye Integer with value 0 or 1
// (Rye's boolean representation).
func (e *Engine) EvalBool(code string) (bool, error) {
	obj, err := e.EvalGetObject(code)
	if err != nil {
		return false, err
	}
	switch v := obj.(type) {
	case env.Integer:
		return v.Value != 0, nil
	default:
		return false, errors.New("result is not a boolean (integer): " + obj.Print(*e.ps.Idx))
	}
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

// Reset re-initialises the engine with a fresh context and re-registers all
// base Rye builtins (no OS I/O, matching the behaviour of [New]).
// Custom builtins registered with RegisterBuiltin are
// removed — call RegisterBuiltin again to re-add them, or use [New] to
// create a fresh engine.
func (e *Engine) Reset() {
	block, idxs := loader.LoadStringNoPEG("", false)
	e.ps = env.NewProgramStateOLD(block.(env.Block).Series, idxs)
	evaldo.RegisterBaseBuiltins(e.ps)
}
