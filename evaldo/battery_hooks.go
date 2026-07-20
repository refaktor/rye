package evaldo

// battery_hooks.go – thin hooks that base evaldo code can call into battery
// functionality without creating an import cycle.
//
// Each hook is a package-level function variable, defaulting to a no-op or
// nil implementation.  The batteries package populates these variables during
// batteries.RegisterBatteries().

import "github.com/refaktor/rye/env"

// BatteryValidateHook is called by base builtins (e.g. "_<<", "assure-kind")
// that need to validate a Dict/RyeCtx against a spec block.
// Batteries set this to batteries.BuiValidate.
var BatteryValidateHook func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object = func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
	return MakeBuiltinError(ps, "Validation requires the batteries package (call batteries.RegisterBatteries first).", "validate")
}

// BatteryConvertHook is called by base builtins that perform Kind-based conversion.
// Batteries set this to batteries.BuiConvert.
var BatteryConvertHook func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object = func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
	return MakeBuiltinError(ps, "Conversion requires the batteries package (call batteries.RegisterBatteries first).", "convert")
}

// BatteryMarkdownDisplayHook is called by the base printing builtins to render
// markdown blocks in the REPL.  Batteries set this to their markdown renderer.
// The returned []interface{} should be the converted display items.
var BatteryMarkdownDisplayHook func(source string) []interface{} = func(source string) []interface{} {
	return nil // no-op without batteries
}

// PersistentCtxInterface is the interface evaldo's base builtins use to
// interact with a PersistentCtx without depending on its concrete type.
// The batteries package's PersistentCtx struct satisfies this interface.
type PersistentCtxInterface interface {
	env.Context
}

// BatteryEvalInPersistentCtxHook is called by "do\inside" when the context is
// a PersistentCtx.  Batteries set this to their EvalBlockInPersistentCtx.
var BatteryEvalInPersistentCtxHook func(ps *env.ProgramState, ctx PersistentCtxInterface) = func(ps *env.ProgramState, ctx PersistentCtxInterface) {
	// no-op without batteries
}

// BatteryEyrEvalBlockInsideHook is called by EvalBlockInj when the dialect is
// EyrDialect.  Batteries set this to Eyr_EvalBlockInside.
var BatteryEyrEvalBlockInsideHook func(ps *env.ProgramState, inj env.Object, injnow bool) = func(ps *env.ProgramState, inj env.Object, injnow bool) {
	// no-op without batteries
}

// BatteryEyrEvalBlockHook is called by the REPL when the dialect is "eyr".
// Batteries set this to Eyr_EvalBlock.
var BatteryEyrEvalBlockHook func(ps *env.ProgramState, full bool) = func(ps *env.ProgramState, full bool) {
	// no-op without batteries
}

// BatteryDialectMathHook is called by the REPL when the dialect is "math".
// Batteries set this to DialectMath.
var BatteryDialectMathHook func(ps *env.ProgramState, arg0 env.Object) env.Object = func(ps *env.ProgramState, arg0 env.Object) env.Object {
	return MakeBuiltinError(ps, "Math dialect requires the batteries package.", "math")
}
