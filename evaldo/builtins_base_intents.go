package evaldo

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/refaktor/rye/env"
)

// input-intent builtin implements semantic input points with optional scenario override
// and validation via the validation dialect (through BatteryValidateHook).
//
// Usage:
//   age: input-intent 'user-age { integer } { read-file %user.json |-> "age" }
//
// Semantics:
//   1) Determine the intent name from the first argument (word, tagword, or string).
//   2) If a scenario override is available, use it; otherwise evaluate the provider block.
//      Scenario sources (in order):
//        - BatteryScenarioGetHook(ps, name)
//        - ps.Ctx/intents (Dict or Context)
//        - ps.Ctx/scenario/intents (Dict or Context)
//   3) Always validate the resulting value with the validation dialect block.
//      Requires batteries to be registered; otherwise BatteryValidateHook will return an error.
var builtins_intents = map[string]*env.Builtin{
	"input-intent": {
		Argsn: 3,
		Doc:   "Declares a semantic input point: input-intent name validator-block provider-block. In scenario mode, returns the mocked value for the given intent name; otherwise evaluates the provider block. The resulting value is always validated using the validation dialect block.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// 1) Extract intent name
			var intentName string
			var intentWordIdx int
			switch n := arg0.(type) {
			case env.Tagword:
				intentName = ps.Idx.GetWord(n.Index)
				intentWordIdx = n.Index
			case env.Word:
				intentName = ps.Idx.GetWord(n.Index)
				intentWordIdx = n.Index
			case env.String:
				intentName = n.Value
				intentWordIdx = ps.Idx.IndexWord(intentName)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.TagwordType, env.WordType, env.StringType}, "input-intent")
			}

			// 2) Determine value: scenario override or evaluate provider block
			var value env.Object

			// 2a) Try scenario via hook if batteries provide it
			if BatteryScenarioGetHook != nil {
				if ok, v := BatteryScenarioGetHook(ps, intentName); ok {
					value = v
				}
			}

			// 2b) If no hook override, look for intents in context(s)
			if value == nil {
				// Try top-level 'intents word first
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

			// 2c) Try nested scenario/intents structure: scenario { intents: ... }
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
							// look for 'intents inside this context
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

			// 2d) If still no value, evaluate the provider block
			if value == nil {
				switch provider := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = provider.Series
					EvalBlockInj(ps, nil, false)
					MaybeDisplayFailureOrError(ps, ps.Idx, "input-intent")
					ps.Ser = ser
					if ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag {
						return ps.Res
					}
					value = ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "input-intent")
				}
			}

			// 3) Validate the value using the validation dialect block
			switch spec := arg1.(type) {
			case env.Block:
				// Use BatteryValidateHook to avoid direct dependency on batteries
				return BatteryValidateHook(ps, value, spec)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "input-intent")
			}
		},
	},

	// output-intent builtin: logs the payload and conditionally executes the side-effect block.
	// Usage:
	//   output-intent { payload } { side-effect-block }
	// Behavior:
	//   - Always appends a log entry to output.log with timestamp and payload Inspect.
	//   - If scenario mode is active, it does NOT execute the side-effect block (simulation mode).
	//   - Otherwise, evaluates the side-effect block normally.
	//   - Returns the payload value (so it can be captured or piped if desired).
	"output-intent": {
		Argsn: 2,
		Doc:   "Logs an output payload to output.log and executes the side-effect block unless in scenario mode, where the side-effect is skipped. Returns the payload.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			payload := arg0
			// Log to output.log
			logLine := fmt.Sprintf("%s | %s\n", time.Now().Format(time.RFC3339), payload.Inspect(*ps.Idx))
			fn := "output.log"
			if ps.WorkingPath != "" {
				fn = filepath.Join(ps.WorkingPath, fn)
			}
			_ = appendToFile(fn, []byte(logLine))

			// If a batteries hook wants to capture outputs, let it (non-blocking decision for side-effects)
			if BatteryScenarioCaptureOutputHook != nil {
				_ = BatteryScenarioCaptureOutputHook(ps, payload)
			}

			// Scenario mode: skip executing the side-effect block
			if isScenarioMode(ps) {
				return payload
			}

			// Execute the side-effect block normally
			switch blk := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = blk.Series
				EvalBlockInj(ps, nil, false)
				MaybeDisplayFailureOrError(ps, ps.Idx, "output-intent")
				ps.Ser = ser
				if ps.ErrorFlag || ps.ReturnFlag || ps.FailureFlag {
					return ps.Res
				}
				return payload // Preserve pass-through feel; payload stays as returned value
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "output-intent")
			}
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

// Helper: detect scenario mode via batteries hook or presence of 'scenario' in context
func isScenarioMode(ps *env.ProgramState) bool {
	if BatteryIsScenarioHook != nil && BatteryIsScenarioHook(ps) {
		return true
	}
	if idx, found := ps.Idx.GetIndex("scenario"); found {
		if obj, ok := ps.Ctx.Get(idx); ok && obj != nil {
			return true
		}
	}
	return false
}
