package evaldo

import (
	"fmt"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"

	. "github.com/refaktor/go-peg"
)

// Helper function to convert a Rye block to a PEG grammar string
func blockToPegGrammar(block env.Block, ps *env.ProgramState) (string, error) {
	var grammar strings.Builder

	for i := 0; i < block.Series.Len(); i += 2 {
		if i+1 >= block.Series.Len() {
			return "", fmt.Errorf("malformed PEG grammar block: expected rule name and definition pairs")
		}

		// Get rule name
		ruleName, ok := block.Series.Get(i).(env.Word)
		if !ok {
			return "", fmt.Errorf("rule name must be a word")
		}

		// Get rule definition
		ruleDefObj := block.Series.Get(i + 1)
		var ruleDef string

		switch def := ruleDefObj.(type) {
		case env.String:
			ruleDef = def.Value
		default:
			return "", fmt.Errorf("rule definition must be a string")
		}

		// Append to grammar
		grammar.WriteString(ps.Idx.GetWord(ruleName.Index))
		grammar.WriteString(" <- ")
		grammar.WriteString(ruleDef)
		grammar.WriteString("\n")
	}

	return grammar.String(), nil
}

// Helper function to create a parser action function that returns Rye values
func createParserAction(ps *env.ProgramState, actionBlock env.Block) func(v *Values, d Any) (Any, error) {
	return func(v *Values, d Any) (Any, error) {
		// Create a context with the parsed values
		ctx := env.NewEnv(ps.Ctx)

		// Add the Values object to the context
		valuesNative := env.NewNative(ps.Idx, nil, "PEG-Values")
		valuesNative.Value = v
		ctx.Set(ps.Idx.IndexWord("values"), *valuesNative)

		// Add token to the context
		ctx.Set(ps.Idx.IndexWord("token"), *env.NewString(v.Token()))

		// Add position to the context
		ctx.Set(ps.Idx.IndexWord("position"), *env.NewInteger(int64(v.Pos)))

		// Add choice to the context if it exists
		if v.Choice >= 0 {
			ctx.Set(ps.Idx.IndexWord("choice"), *env.NewInteger(int64(v.Choice)))
		}

		// Execute the action block in the new context
		oldCtx := ps.Ctx
		oldSer := ps.Ser
		ps.Ctx = ctx
		ps.Ser = actionBlock.Series
		EvalBlock(ps)
		if ps.ErrorFlag {
			ps.Ctx = oldCtx
			ps.Ser = oldSer
			return nil, fmt.Errorf("error in action block")
		}
		result := ps.Res
		ps.Ctx = oldCtx
		ps.Ser = oldSer

		return result, nil
	}
}

// Helper function to register actions for rules
func registerRuleActions(parser *Parser, actions env.Dict, ps *env.ProgramState) error {
	for key, value := range actions.Data {
		ruleName := key

		// Get the action block
		actionBlock, ok := value.(env.Block)
		if !ok {
			return fmt.Errorf("action for rule '%s' must be a block", ruleName)
		}

		// Get the rule from the parser
		rule, ok := parser.Grammar[ruleName]
		if !ok {
			return fmt.Errorf("rule '%s' not found in grammar", ruleName)
		}

		// Set the action function
		rule.Action = createParserAction(ps, actionBlock)
	}

	return nil
}

// Helper function to parse input using a PEG parser
func parseWithPeg(parser *Parser, input string, ps *env.ProgramState) (env.Object, error) {
	// Parse the input
	val, err := parser.ParseAndGetValue(input, nil)
	if err != nil {
		return *env.NewError(err.Error()), err
	}

	// Convert the result to a Rye value
	if val == nil {
		return env.Void{}, nil
	}

	switch result := val.(type) {
	case env.Object:
		return result, nil
	default:
		// Try to convert to a Rye value
		return env.ToRyeValue(result), nil
	}
}

var Builtins_peg = map[string]*env.Builtin{
	// peg-grammar creates a new PEG grammar from a block of rule definitions
	// Args:
	// * grammar: Block containing alternating rule names and rule definitions
	// Returns:
	// * A native object containing the compiled PEG grammar
	"peg-grammar": {
		Argsn: 1,
		Doc:   "Creates a new PEG grammar from a block of rule definitions",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch grammarBlock := arg0.(type) {
			case env.Block:
				// Convert the block to a PEG grammar string
				grammarStr, err := blockToPegGrammar(grammarBlock, ps)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}

				// Create a new parser
				parser, err := NewParser(grammarStr)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(fmt.Sprintf("Error creating parser: %s", err.Error()))
				}

				// Create a native object to hold the parser
				parserNative := env.NewNative(ps.Idx, nil, "PEG-Grammar")
				parserNative.Value = parser

				return *parserNative
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "peg-grammar")
			}
		},
	},

	// peg-actions adds action functions to a PEG grammar
	// Args:
	// * grammar: Native object containing a PEG grammar
	// * actions: Dict mapping rule names to action blocks
	// Returns:
	// * The grammar object with actions added
	"peg-actions": {
		Argsn: 2,
		Doc:   "Adds action functions to a PEG grammar",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch grammarObj := arg0.(type) {
			case env.Native:
				kindWord := ps.Idx.GetWord(grammarObj.Kind.Index)
				if kindWord != "PEG-Grammar" {
					ps.FailureFlag = true
					return *env.NewError("First argument must be a PEG grammar")
				}

				parser, ok := grammarObj.Value.(*Parser)
				if !ok {
					ps.FailureFlag = true
					return *env.NewError("Invalid PEG grammar object")
				}

				switch actionsObj := arg1.(type) {
				case env.Dict:
					err := registerRuleActions(parser, actionsObj, ps)
					if err != nil {
						ps.FailureFlag = true
						return *env.NewError(err.Error())
					}

					return grammarObj
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType}, "peg-actions")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "peg-actions")
			}
		},
	},

	// peg-parse parses input using a PEG grammar
	// Args:
	// * grammar: Native object containing a PEG grammar
	// * input: String to parse
	// Returns:
	// * The parsed result as a Rye value
	"peg-parse": {
		Argsn: 2,
		Doc:   "Parses input using a PEG grammar",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch grammarObj := arg0.(type) {
			case env.Native:
				kindWord := ps.Idx.GetWord(grammarObj.Kind.Index)
				if kindWord != "PEG-Grammar" {
					ps.FailureFlag = true
					return *env.NewError("First argument must be a PEG grammar")
				}

				parser, ok := grammarObj.Value.(*Parser)
				if !ok {
					ps.FailureFlag = true
					return *env.NewError("Invalid PEG grammar object")
				}

				switch inputObj := arg1.(type) {
				case env.String:
					result, err := parseWithPeg(parser, inputObj.Value, ps)
					if err != nil {
						ps.FailureFlag = true
						return *env.NewError(fmt.Sprintf("Parse error: %s", err.Error()))
					}

					return result
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "peg-parse")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "peg-parse")
			}
		},
	},

	// peg-rule-names gets the names of all rules in a grammar
	// Args:
	// * grammar: Native object containing a PEG grammar
	// Returns:
	// * A block containing the names of all rules in the grammar
	"peg-rule-names": {
		Argsn: 1,
		Doc:   "Gets the names of all rules in a grammar",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch grammarObj := arg0.(type) {
			case env.Native:
				kindWord := ps.Idx.GetWord(grammarObj.Kind.Index)
				if kindWord != "PEG-Grammar" {
					ps.FailureFlag = true
					return *env.NewError("Argument must be a PEG grammar")
				}

				parser, ok := grammarObj.Value.(*Parser)
				if !ok {
					ps.FailureFlag = true
					return *env.NewError("Invalid PEG grammar object")
				}

				// Get all rule names
				ruleNames := make([]env.Object, 0, len(parser.Grammar))
				for name := range parser.Grammar {
					ruleNames = append(ruleNames, *env.NewString(name))
				}

				return *env.NewBlock(*env.NewTSeries(ruleNames))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "peg-rule-names")
			}
		},
	},

	// peg-values-get gets a value from a PEG Values object
	// Args:
	// * values: Native object containing a PEG Values object
	// * index: Integer index of the value to get
	// Returns:
	// * The value at the specified index
	"peg-values-get": {
		Argsn: 2,
		Doc:   "Gets a value from a PEG Values object",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch valuesObj := arg0.(type) {
			case env.Native:
				kindWord := ps.Idx.GetWord(valuesObj.Kind.Index)
				if kindWord != "PEG-Values" {
					ps.FailureFlag = true
					return *env.NewError("First argument must be a PEG Values object")
				}

				values, ok := valuesObj.Value.(*Values)
				if !ok {
					ps.FailureFlag = true
					return *env.NewError("Invalid PEG Values object")
				}

				switch indexObj := arg1.(type) {
				case env.Integer:
					index := int(indexObj.Value)
					if index < 0 || index >= len(values.Vs) {
						ps.FailureFlag = true
						return *env.NewError(fmt.Sprintf("Index %d out of range (0-%d)", index, len(values.Vs)-1))
					}

					// Convert the value to a Rye value
					if values.Vs[index] == nil {
						return env.Void{}
					}

					switch val := values.Vs[index].(type) {
					case env.Object:
						return val
					default:
						return env.ToRyeValue(val)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "peg-values-get")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "peg-values-get")
			}
		},
	},

	// peg-values-len gets the length of a PEG Values object
	// Args:
	// * values: Native object containing a PEG Values object
	// Returns:
	// * The number of values in the Values object
	"peg-values-len": {
		Argsn: 1,
		Doc:   "Gets the length of a PEG Values object",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch valuesObj := arg0.(type) {
			case env.Native:
				kindWord := ps.Idx.GetWord(valuesObj.Kind.Index)
				if kindWord != "PEG-Values" {
					ps.FailureFlag = true
					return *env.NewError("Argument must be a PEG Values object")
				}

				values, ok := valuesObj.Value.(*Values)
				if !ok {
					ps.FailureFlag = true
					return *env.NewError("Invalid PEG Values object")
				}

				return *env.NewInteger(int64(len(values.Vs)))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "peg-values-len")
			}
		},
	},

	// peg-load-string loads a string into Rye values using the standard Rye parser
	// Args:
	// * input: String to parse
	// Returns:
	// * The parsed Rye block
	"peg-load-string": {
		Argsn: 1,
		Doc:   "Loads a string into Rye values using the standard Rye parser",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inputObj := arg0.(type) {
			case env.String:
				block := loader.LoadStringNEW(inputObj.Value, false, ps)
				return block
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "peg-load-string")
			}
		},
	},
}
