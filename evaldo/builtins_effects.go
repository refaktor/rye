package evaldo

import (
	"bufio"
	"fmt"
	"io"
	"os"
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

// effectAuditf appends a timestamped line to the optional -eff-log file. Errors
// are deliberately ignored: auditing must never break the script.
func effectAuditf(ps *env.ProgramState, format string, args ...any) {
	if ps.EffectMode.AuditFile == "" {
		return
	}
	line := time.Now().Format(time.RFC3339) + " " + fmt.Sprintf(format, args...) + "\n"
	_ = appendToFile(ps.EffectMode.AuditFile, []byte(line))
}

// evalEffectBlock runs a provider/effect block so that a failure at the end of
// the block is returned to the caller instead of being elevated to an error at
// top level. That is what makes `input-effect ... |fix {...}` work.
func evalEffectBlock(ps *env.ProgramState, block env.Block) {
	ser := ps.Ser
	callDepth := ps.CallDepth
	ps.Ser = block.Series
	ps.CallDepth++
	EvalBlockInj(ps, nil, false)
	ps.CallDepth = callDepth
	ps.Ser = ser
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
	return runOutputEffectText(ps, name, payload, payload.Print(*ps.Idx), effect, returnPayload)
}

func runOutputEffectText(ps *env.ProgramState, name string, payload env.Object, payloadText string, effect env.Block, returnPayload bool) env.Object {
	// Replay always skips real effects. The recorded stand-in (or payload) is
	// returned so downstream computation stays deterministic.
	if tape := ps.EffectMode.Replay; tape != nil {
		ps.EffectMode.OutputsSkipped++
		entry, found, matched := tape.NextOutput(name, payload, ps.Idx)
		if found && matched {
			effectAuditf(ps, "replay %s %s ok", name, payloadText)
		} else {
			effectAuditf(ps, "replay %s %s mismatch", name, payloadText)
		}
		if returnPayload {
			return payload
		}
		if entry.Result != nil {
			return entry.Result
		}
		return payload
	}

	if isScenarioMode(ps) {
		ps.EffectMode.OutputsSkipped++
		effectLogf(ps, "would %s: %s\n", name, payloadText)
		effectAuditf(ps, "would %s %s", name, payloadText)
		if BatteryScenarioCaptureOutputHook != nil {
			_ = BatteryScenarioCaptureOutputHook(ps, payload)
		}
		return payload
	}

	if ps.EffectMode.Confirm {
		effectLogf(ps, "run %s: %s? [y/N] ", name, payloadText)
		line, ok := readEffectLine(ps)
		if !ok || !isYes(line) {
			ps.EffectMode.OutputsSkipped++
			effectAuditf(ps, "skip %s %s", name, payloadText)
			return payload
		}
	}

	if BatteryScenarioCaptureOutputHook != nil {
		_ = BatteryScenarioCaptureOutputHook(ps, payload)
	}

	evalEffectBlock(ps, effect)
	ps.EffectMode.OutputsExecuted++
	failed := ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag
	if failed {
		ps.EffectMode.OutputsFailed++
		effectAuditf(ps, "exec %s %s fail", name, payloadText)
	} else {
		effectAuditf(ps, "exec %s %s ok", name, payloadText)
	}

	if ps.EffectMode.Record != nil {
		result := ps.Res
		if returnPayload {
			result = payload
		}
		ps.EffectMode.Record.Add(env.EffectEntry{Kind: "output", Name: name, Value: payload, Result: result})
	}

	if failed {
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
		Effect: env.EffInput,
		Argsn:  3,
		Doc:    "Declares an input effect: input-effect name validator-block provider-block. Evaluates the provider (unless an embedded scenario hook supplies a value), then validates the value with the validation dialect. With -eff-ask it prompts after the provider; with -eff-replay it reads the value from the tape; with -eff-record it records the validated value.",
		Pure:   false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			intentName, _, nameErr := effectName(ps, arg0, 1, "input-effect")
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
			fromTape := false

			// Replay: inputs come from the tape. A missing input is an error
			// unless -eff-replay-fallback is set, in which case the provider runs.
			if tape := ps.EffectMode.Replay; tape != nil {
				if entry, found := tape.NextInput(intentName); found {
					value = entry.Value
					fromTape = true
				} else if !ps.EffectMode.ReplayFallback {
					tape.Mismatches = append(tape.Mismatches, "no recorded value for input '"+intentName+"'")
					ps.ErrorFlag = true
					return env.NewError("no recorded value for input '" + intentName + "'")
				}
			}

			// Embedded scenario hook (Go embedders). User words named
			// `intents` or `scenario` intentionally have no effect.
			if value == nil && BatteryScenarioGetHook != nil {
				if found, v := BatteryScenarioGetHook(ps, intentName); found {
					value = v
				}
			}

			// No override: evaluate the provider block.
			if value == nil {
				evalEffectBlock(ps, provider)
				if ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag {
					return ps.Res
				}
				value = ps.Res
			}

			ps.EffectMode.Inputs++

			var validated env.Object
			if ps.EffectMode.Ask {
				validated = askEffectValue(ps, intentName, value, spec)
			} else {
				validated = validateEffectValue(ps, value, spec)
			}

			if !ps.ErrorFlag && !ps.FailureFlag && ps.EffectMode.Record != nil && !fromTape {
				ps.EffectMode.Record.Add(env.EffectEntry{Kind: "input", Name: intentName, Value: validated})
			}
			return validated
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
		Effect: env.EffOutput,
		Argsn:  3,
		Doc:    "Declares an output effect: output-effect name payload effect-block. Executes the effect block normally. With -eff-dry-run it logs `would name: payload` and skips the effect. With -eff-confirm it asks before executing. With -eff-replay it skips the effect and compares the payload with the tape.",
		Pure:   false,
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
	return false
}
