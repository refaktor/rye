package evaldo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/refaktor/rye/env"
)

// Declared effects: explicit input and output points in a script.
//
// input-effect evaluates a provider block unless an embedded scenario hook
// supplies a value. The resulting value is always passed through the validation
// dialect. output-effect executes a side-effect block, unless dry-run or
// confirm mode skips it.

var effectsStdin *bufio.Reader

func effectStdin() *bufio.Reader {
	if effectsStdin == nil {
		effectsStdin = bufio.NewReader(os.Stdin)
	}
	return effectsStdin
}

func effectLogWriter(ps *env.ProgramState) io.Writer {
	if ps.EffectMode.Log != nil {
		return ps.EffectMode.Log
	}
	return os.Stderr
}

func effectLogf(ps *env.ProgramState, format string, args ...any) {
	fmt.Fprintf(effectLogWriter(ps), format, args...)
}

func readEffectLine(ps *env.ProgramState) (string, bool) {
	line, err := effectStdin().ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if err != nil && line == "" {
		return "", false
	}
	return line, true
}

func effectName(ps *env.ProgramState, obj env.Object, argn int, builtinName string) (string, int, env.Object) {
	switch n := obj.(type) {
	case env.Tagword:
		return ps.Idx.GetWord(n.Index), n.Index, nil
	case env.Word:
		return ps.Idx.GetWord(n.Index), n.Index, nil
	case env.String:
		return n.Value, ps.Idx.IndexWord(n.Value), nil
	default:
		ps.FailureFlag = true
		return "", 0, MakeArgError(ps, argn, []env.Type{env.TagwordType, env.WordType, env.StringType}, builtinName)
	}
}

// validateEffectValue runs the validation dialect, but does not display
// validation errors: callers decide whether to re-prompt (ask mode) or return
// the failure to Rye (normal mode).
func validateEffectValue(ps *env.ProgramState, value env.Object, spec env.Block) env.Object {
	return BatteryValidateHook(ps, value, spec)
}

func askEffectValue(ps *env.ProgramState, name string, value env.Object, spec env.Block) env.Object {
	for {
		effectLogf(ps, "%s [%s]: ", name, value.Print(*ps.Idx))
		line, ok := readEffectLine(ps)
		if !ok {
			return validateEffectValue(ps, value, spec)
		}
		candidate := value
		if line != "" {
			candidate = *env.NewString(line)
		}

		savedFailure := ps.FailureFlag
		savedError := ps.ErrorFlag
		ps.FailureFlag = false
		ps.ErrorFlag = false

		validated := validateEffectValue(ps, candidate, spec)
		if ps.FailureFlag || ps.ErrorFlag {
			effectLogf(ps, "invalid %s: %s\n", name, validated.Print(*ps.Idx))
			ps.FailureFlag = false
			ps.ErrorFlag = false
			ps.FailureFlag = savedFailure
			ps.ErrorFlag = savedError
			continue
		}

		ps.FailureFlag = savedFailure
		ps.ErrorFlag = savedError
		return validated
	}
}

func runOutputEffect(ps *env.ProgramState, name string, payload env.Object, effect env.Block, returnPayload bool) env.Object {
	if isScenarioMode(ps) {
		effectLogf(ps, "would %s: %s\n", name, payload.Print(*ps.Idx))
		if BatteryScenarioCaptureOutputHook != nil {
			_ = BatteryScenarioCaptureOutputHook(ps, payload)
		}
		return payload
	}

	if ps.EffectMode.Confirm {
		effectLogf(ps, "run %s: %s? [y/N] ", name, payload.Print(*ps.Idx))
		line, ok := readEffectLine(ps)
		if !ok || !isYes(line) {
			return payload
		}
	}

	ser := ps.Ser
	ps.Ser = effect.Series
	EvalBlockInj(ps, nil, false)
	MaybeDisplayFailureOrError(ps, ps.Idx, "output-effect")
	ps.Ser = ser
	if ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag {
		return ps.Res
	}
	if returnPayload {
		return payload
	}
	return ps.Res
}

func isYes(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "y" || s == "yes"
}

var inputEffectFn func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object

var builtins_effects = map[string]*env.Builtin{

	// ##### Declared effects ##### "Explicit input and output points with validation, dry-run, ask and confirm."

	// Args:
	// * name: Word, String, or Tagword identifying the effect.
	// * validator: Validation-dialect block applied to the resulting value.
	// * provider: Block that produces the real value.
	// Returns:
	// * The validated value.
	// Example:
	// input-effect 'username { string } { "real-user" }
	"input-effect": {
		Argsn: 3,
		Doc:   "Declares an input effect: input-effect name validator-block provider-block. Evaluates the provider (unless an embedded scenario hook supplies a value), then validates the value with the validation dialect. With -eff-ask, prompts after the provider and validates the typed answer.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			intentName, intentWordIdx, nameErr := effectName(ps, arg0, 1, "input-effect")
			if nameErr != nil {
				return nameErr
			}

			spec, ok := arg1.(env.Block)
			if !ok {
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "input-effect")
			}
			provider, ok := arg2.(env.Block)
			if !ok {
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.BlockType}, "input-effect")
			}

			var value env.Object

			// Embedded scenario hook (Go embedders).
			if BatteryScenarioGetHook != nil {
				if found, v := BatteryScenarioGetHook(ps, intentName); found {
					value = v
				}
			}

			// Backwards-compatible Rye-level mocks: a top-level `intents`
			// dict/context, or `scenario { intents: {...} }`.
			if value == nil {
				if idx, found := ps.Idx.GetIndex("intents"); found {
					if obj, ok := ps.Ctx.Get(idx); ok && obj != nil {
						switch o := obj.(type) {
						case env.Dict:
							if v, ok2 := o.Data[intentName]; ok2 {
								value = env.ToRyeValue(v)
							}
						case *env.RyeCtx:
							if v, ok2 := o.Get(intentWordIdx); ok2 {
								value = v
							}
						}
					}
				}
			}
			if value == nil {
				if sidx, found := ps.Idx.GetIndex("scenario"); found {
					if sobj, ok := ps.Ctx.Get(sidx); ok && sobj != nil {
						switch sc := sobj.(type) {
						case env.Dict:
							if raw, ok2 := sc.Data["intents"]; ok2 {
								switch intents := raw.(type) {
								case env.Dict:
									if v, ok3 := intents.Data[intentName]; ok3 {
										value = env.ToRyeValue(v)
									}
								case *env.RyeCtx:
									if v, ok3 := intents.Get(intentWordIdx); ok3 {
										value = v
									}
								}
							}
						case *env.RyeCtx:
							if idx2, okName := ps.Idx.GetIndex("intents"); okName {
								if it, okIt := sc.Get(idx2); okIt && it != nil {
									switch intents := it.(type) {
									case env.Dict:
										if v, ok3 := intents.Data[intentName]; ok3 {
											value = env.ToRyeValue(v)
										}
									case *env.RyeCtx:
										if v, ok3 := intents.Get(intentWordIdx); ok3 {
											value = v
										}
									}
								}
							}
						}
					}
				}
			}

			// No override: evaluate the provider block.
			if value == nil {
				ser := ps.Ser
				ps.Ser = provider.Series
				EvalBlockInj(ps, nil, false)
				ps.Ser = ser
				if ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag {
					return ps.Res
				}
				value = ps.Res
			}

			if ps.EffectMode.Ask {
				return askEffectValue(ps, intentName, value, spec)
			}
			return validateEffectValue(ps, value, spec)
		},
	},

	// Args:
	// * name: Word, String, or Tagword identifying the effect.
	// * payload: Value describing what is being done.
	// * effect: Block that performs the side effect.
	// Returns:
	// * The effect block's result when executed, or the payload when skipped.
	// Example:
	// output-effect 'user-login "alice" { print "logged in" }
	"output-effect": {
		Argsn: 3,
		Doc:   "Declares an output effect: output-effect name payload effect-block. Executes the effect block normally. With -eff-dryrun it logs `would name: payload` and skips the effect. With -eff-confirm it asks before executing.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			name, _, nameErr := effectName(ps, arg0, 1, "output-effect")
			if nameErr != nil {
				return nameErr
			}
			effect, ok := arg2.(env.Block)
			if !ok {
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.BlockType}, "output-effect")
			}
			return runOutputEffect(ps, name, arg1, effect, false)
		},
	},

	// Args:
	// * name: Word, String, or Tagword identifying the input.
	// * validator: Validation-dialect block applied to the resulting value.
	// * provider: Block that produces the real value.
	// Returns:
	// * The validated value.
	// Example:
	// input-intent 'username { string } { "real-user" }
	// Deprecated alias for input-effect. Kept so existing scripts continue to work.
	"input-intent": {
		Argsn: 3,
		Doc:   "Deprecated alias for input-effect. Declares an input with validation-dialect validation.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return inputEffectFn(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Args:
	// * payload: Name or value to log and return.
	// * fields: Block of additional values to include in the log entry.
	// * side-effect: Block to execute outside dry-run mode.
	// Returns:
	// * The payload.
	// Example:
	// output-intent "user-login" { "alice" } { print "logged in" }
	// Deprecated alias for output-effect with the old shape:
	// output-intent payload { fields } { side-effect }.
	"output-intent": {
		Argsn: 3,
		Doc:   "Deprecated alias for output-effect using the old shape: output-intent payload { fields } { side-effect }.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			payload := arg0

			var fieldsStr string
			fields, ok := arg1.(env.Block)
			if !ok {
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "output-intent")
			}
			effect, ok := arg2.(env.Block)
			if !ok {
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.BlockType}, "output-intent")
			}

			serSaved := ps.Ser
			ps.Ser = fields.Series
			EvalBlockInj(ps, nil, false)
			ps.Ser = serSaved
			if ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag {
				return ps.Res
			}
			if ps.Res != nil && ps.Res.Type() != env.VoidType {
				switch v := ps.Res.(type) {
				case env.Block:
					parts := make([]string, 0, v.Series.Len())
					for _, it := range v.Series.GetAll() {
						if it == nil {
							continue
						}
						if iv, ok := it.(env.String); ok {
							parts = append(parts, iv.Value)
						} else {
							parts = append(parts, it.Inspect(*ps.Idx))
						}
					}
					fieldsStr = strings.Join(parts, " ")
				case env.List:
					parts := make([]string, 0, len(v.Data))
					for _, raw := range v.Data {
						if raw == nil {
							continue
						}
						switch iv := raw.(type) {
						case env.String:
							parts = append(parts, iv.Value)
						case string:
							parts = append(parts, iv)
						default:
							if obj, ok := raw.(env.Object); ok {
								parts = append(parts, obj.Inspect(*ps.Idx))
							} else {
								parts = append(parts, fmt.Sprintf("%v", raw))
							}
						}
					}
					fieldsStr = strings.Join(parts, " ")
				case env.String:
					fieldsStr = v.Value
				default:
					fieldsStr = ps.Res.Inspect(*ps.Idx)
				}
			}

			name := payload.Print(*ps.Idx)
			inline := name
			if fieldsStr != "" && fieldsStr != "void" {
				inline = fmt.Sprintf("%s | %s", name, fieldsStr)
			}

			if isScenarioMode(ps) {
				effectLogf(ps, "would %s: %s\n", name, inline)
				logLine := fmt.Sprintf("%s | %s\n", time.Now().Format(time.RFC3339), inline)
				fn := "output.log"
				if ps.WorkingPath != "" {
					fn = filepath.Join(ps.WorkingPath, fn)
				}
				_ = appendToFile(fn, []byte(logLine))
				if BatteryScenarioCaptureOutputHook != nil {
					_ = BatteryScenarioCaptureOutputHook(ps, payload)
				}
				return payload
			}

			if ps.EffectMode.Confirm {
				effectLogf(ps, "run %s: %s? [y/N] ", name, inline)
				line, ok := readEffectLine(ps)
				if !ok || !isYes(line) {
					return payload
				}
			}

			if BatteryScenarioCaptureOutputHook != nil {
				_ = BatteryScenarioCaptureOutputHook(ps, payload)
			}

			ser := ps.Ser
			ps.Ser = effect.Series
			EvalBlockInj(ps, nil, false)
			MaybeDisplayFailureOrError(ps, ps.Idx, "output-intent side-effect")
			ps.Ser = ser
			if ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag {
				return ps.Res
			}
			return payload
		},
	},
}

func init() {
	inputEffectFn = builtins_effects["input-effect"].Fn
}

// Helper: append to file (best-effort; errors ignored for now as logging shouldn't break execution)
func appendToFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// isScenarioMode reports whether output effects should be skipped without
// running. It reads process effect mode, not user words: a script binding
// `scenario:` must not silently disable its side effects.
func isScenarioMode(ps *env.ProgramState) bool {
	if ps.EffectMode.DryRun {
		return true
	}
	if BatteryIsScenarioHook != nil && BatteryIsScenarioHook(ps) {
		return true
	}
	if os.Getenv("RYE_EFF_DRY_RUN") == "1" || os.Getenv("RYE_DRY_RUN") == "1" {
		return true
	}
	return false
}
