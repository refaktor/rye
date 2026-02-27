//go:build !no_cli
// +build !no_cli

package evaldo

import (
	"fmt"
	"strings"

	"github.com/refaktor/rye/env"
)

// ArgSpec holds the specification for a single argument
type ArgSpec struct {
	Name         string     // Name for the result (from flagword long name or setword)
	ShortFlag    string     // Short flag name (e.g., "v")
	LongFlag     string     // Long flag name (e.g., "verbose")
	IsFlag       bool       // True if it's a boolean flag (no value)
	IsPositional bool       // True if it's a positional argument
	IsRequired   bool       // True if required
	IsList       bool       // True if can be repeated (collects into list)
	IsMany       bool       // True if positional accepts many values
	ValueType    string     // "string", "integer", "decimal", "boolean", "file", "any"
	Default      env.Object // Default value
	CheckBlock   *env.Block // Optional validation block
	Doc          string     // Documentation string
}

// SubcommandSpec holds the specification for a subcommand
type SubcommandSpec struct {
	Name string
	Args []ArgSpec
}

// CLISpec holds the complete CLI specification
type CLISpec struct {
	GlobalArgs  []ArgSpec
	Subcommands map[string]*SubcommandSpec
}

// ParsedArgs holds the result of parsing
type ParsedArgs struct {
	Values     map[string]env.Object
	Command    string // Name of subcommand if any
	Positional []env.Object
}

// CLI_ParseSpec parses the specification block into a CLISpec
func CLI_ParseSpec(es *env.ProgramState, specBlock env.Block) (*CLISpec, error) {
	spec := &CLISpec{
		GlobalArgs:  make([]ArgSpec, 0),
		Subcommands: make(map[string]*SubcommandSpec),
	}

	ser := specBlock.Series
	ser.Reset()

	for ser.Pos() < ser.Len() {
		obj := ser.Pop()

		switch item := obj.(type) {
		case env.Flagword:
			// Parse flag specification
			argSpec, err := parseArgSpec(es, &ser, item)
			if err != nil {
				return nil, err
			}
			spec.GlobalArgs = append(spec.GlobalArgs, *argSpec)

		case env.Word:
			wordName := es.Idx.GetWord(item.Index)
			if wordName == "subcommand" {
				// Parse subcommand block
				if ser.Pos() >= ser.Len() {
					return nil, fmt.Errorf("expected block after 'subcommand'")
				}
				subBlock, ok := ser.Pop().(env.Block)
				if !ok {
					return nil, fmt.Errorf("expected block after 'subcommand'")
				}
				subcommands, err := parseSubcommands(es, subBlock)
				if err != nil {
					return nil, err
				}
				spec.Subcommands = subcommands
			} else if wordName == "_" || strings.HasPrefix(wordName, "_") {
				// Positional argument
				argSpec, err := parsePositionalSpec(es, &ser, wordName)
				if err != nil {
					return nil, err
				}
				spec.GlobalArgs = append(spec.GlobalArgs, *argSpec)
			}

		case env.Setword:
			// Named argument using setword (alternative syntax)
			wordName := es.Idx.GetWord(item.Index)
			if strings.HasPrefix(wordName, "_") || wordName == "_" {
				argSpec, err := parsePositionalSpec(es, &ser, wordName)
				if err != nil {
					return nil, err
				}
				spec.GlobalArgs = append(spec.GlobalArgs, *argSpec)
			}
		}
	}

	return spec, nil
}

// parseArgSpec parses a single argument specification
func parseArgSpec(es *env.ProgramState, ser *env.TSeries, flagword env.Flagword) (*ArgSpec, error) {
	spec := &ArgSpec{
		ShortFlag: flagword.GetShortName(*es.Idx),
		LongFlag:  flagword.GetLongName(*es.Idx),
		IsFlag:    false,
		ValueType: "any",
		Default:   env.NewVoid(),
	}

	// Use long name as the result name, fallback to short
	if spec.LongFlag != "" {
		spec.Name = spec.LongFlag
	} else {
		spec.Name = spec.ShortFlag
	}

	// Parse keywords following the flagword
	for ser.Pos() < ser.Len() {
		nextObj := ser.Peek()

		// Stop if we hit another flagword, word starting with _, or subcommand
		switch next := nextObj.(type) {
		case env.Flagword:
			return spec, nil
		case env.Word:
			wordName := es.Idx.GetWord(next.Index)
			if strings.HasPrefix(wordName, "_") || wordName == "subcommand" {
				return spec, nil
			}
		case env.Setword:
			wordName := es.Idx.GetWord(next.Index)
			if strings.HasPrefix(wordName, "_") {
				return spec, nil
			}
		}

		obj := ser.Pop()

		switch item := obj.(type) {
		case env.Word:
			wordName := es.Idx.GetWord(item.Index)
			switch wordName {
			case "flag":
				spec.IsFlag = true
				spec.ValueType = "boolean"
				spec.Default = env.NewBoolean(false)
			case "string":
				spec.ValueType = "string"
				spec.Default = env.NewString("")
			case "integer":
				spec.ValueType = "integer"
				spec.Default = env.NewInteger(0)
			case "decimal":
				spec.ValueType = "decimal"
				spec.Default = env.NewDecimal(0.0)
			case "boolean":
				spec.ValueType = "boolean"
				spec.Default = env.NewBoolean(false)
			case "file":
				spec.ValueType = "file"
				spec.Default = env.NewString("")
			case "any":
				spec.ValueType = "any"
				spec.Default = env.NewVoid()
			case "required":
				spec.IsRequired = true
			case "optional":
				spec.IsRequired = false
				// Check if next item is a default value
				if ser.Pos() < ser.Len() {
					nextVal := ser.Peek()
					switch v := nextVal.(type) {
					case env.Integer, env.String, env.Decimal, env.Boolean:
						spec.Default = v
						ser.Pop()
					}
				}
			case "list":
				spec.IsList = true
			case "check":
				// Get the check block
				if ser.Pos() < ser.Len() {
					if block, ok := ser.Pop().(env.Block); ok {
						spec.CheckBlock = &block
					}
				}
			case "doc":
				// Get the doc string
				if ser.Pos() < ser.Len() {
					if str, ok := ser.Pop().(env.String); ok {
						spec.Doc = str.Value
					}
				}
			}
		case env.Integer, env.String, env.Decimal, env.Boolean:
			// Direct default value
			spec.Default = item
		}
	}

	return spec, nil
}

// parsePositionalSpec parses a positional argument specification
func parsePositionalSpec(es *env.ProgramState, ser *env.TSeries, name string) (*ArgSpec, error) {
	spec := &ArgSpec{
		Name:         name,
		IsPositional: true,
		IsRequired:   false,
		ValueType:    "any",
		Default:      env.NewVoid(),
	}

	// Parse keywords
	for ser.Pos() < ser.Len() {
		nextObj := ser.Peek()

		// Stop if we hit another flagword or positional
		switch next := nextObj.(type) {
		case env.Flagword:
			return spec, nil
		case env.Word:
			wordName := es.Idx.GetWord(next.Index)
			if strings.HasPrefix(wordName, "_") || wordName == "subcommand" {
				return spec, nil
			}
		case env.Setword:
			wordName := es.Idx.GetWord(next.Index)
			if strings.HasPrefix(wordName, "_") {
				return spec, nil
			}
		}

		obj := ser.Pop()

		switch item := obj.(type) {
		case env.Word:
			wordName := es.Idx.GetWord(item.Index)
			switch wordName {
			case "positional":
				// Already set
			case "string":
				spec.ValueType = "string"
			case "integer":
				spec.ValueType = "integer"
			case "decimal":
				spec.ValueType = "decimal"
			case "file":
				spec.ValueType = "file"
			case "any":
				spec.ValueType = "any"
			case "many":
				spec.IsMany = true
			case "one":
				spec.IsRequired = true
				spec.IsMany = false
			case "required":
				spec.IsRequired = true
			case "optional":
				spec.IsRequired = false
				// Check for default
				if ser.Pos() < ser.Len() {
					nextVal := ser.Peek()
					switch v := nextVal.(type) {
					case env.Integer, env.String, env.Decimal, env.Boolean:
						spec.Default = v
						ser.Pop()
					}
				}
			case "check":
				if ser.Pos() < ser.Len() {
					if block, ok := ser.Pop().(env.Block); ok {
						spec.CheckBlock = &block
					}
				}
			case "doc":
				if ser.Pos() < ser.Len() {
					if str, ok := ser.Pop().(env.String); ok {
						spec.Doc = str.Value
					}
				}
			}
		}
	}

	return spec, nil
}

// parseSubcommands parses the subcommand block
func parseSubcommands(es *env.ProgramState, block env.Block) (map[string]*SubcommandSpec, error) {
	subcommands := make(map[string]*SubcommandSpec)

	ser := block.Series
	ser.Reset()

	for ser.Pos() < ser.Len() {
		// Expect word (subcommand name) followed by block
		nameObj := ser.Pop()
		nameWord, ok := nameObj.(env.Word)
		if !ok {
			return nil, fmt.Errorf("expected subcommand name (word), got %T", nameObj)
		}
		cmdName := es.Idx.GetWord(nameWord.Index)

		if ser.Pos() >= ser.Len() {
			return nil, fmt.Errorf("expected block after subcommand name '%s'", cmdName)
		}

		cmdBlock, ok := ser.Pop().(env.Block)
		if !ok {
			return nil, fmt.Errorf("expected block after subcommand name '%s'", cmdName)
		}

		// Parse subcommand's arguments
		subSpec := &SubcommandSpec{
			Name: cmdName,
			Args: make([]ArgSpec, 0),
		}

		subSer := cmdBlock.Series
		subSer.Reset()

		for subSer.Pos() < subSer.Len() {
			obj := subSer.Pop()

			switch item := obj.(type) {
			case env.Flagword:
				argSpec, err := parseArgSpec(es, &subSer, item)
				if err != nil {
					return nil, err
				}
				subSpec.Args = append(subSpec.Args, *argSpec)

			case env.Word:
				wordName := es.Idx.GetWord(item.Index)
				if strings.HasPrefix(wordName, "_") || wordName == "_" {
					argSpec, err := parsePositionalSpec(es, &subSer, wordName)
					if err != nil {
						return nil, err
					}
					subSpec.Args = append(subSpec.Args, *argSpec)
				}
			}
		}

		subcommands[cmdName] = subSpec
	}

	return subcommands, nil
}

// flagwordMatches checks if a flagword from args matches a spec
func flagwordMatches(fw env.Flagword, spec *ArgSpec, idx *env.Idxs) bool {
	// Check if short flag matches
	if fw.HasShort() && spec.ShortFlag != "" {
		if fw.GetShortName(*idx) == spec.ShortFlag {
			return true
		}
	}
	// Check if long flag matches
	if fw.HasLong() && spec.LongFlag != "" {
		if fw.GetLongName(*idx) == spec.LongFlag {
			return true
		}
	}
	return false
}

// findFlagSpecForFlagword finds the spec that matches a given flagword
func findFlagSpecForFlagword(specs []ArgSpec, fw env.Flagword, idx *env.Idxs) *ArgSpec {
	for i := range specs {
		if flagwordMatches(fw, &specs[i], idx) {
			return &specs[i]
		}
	}
	return nil
}

// coerceToType coerces a Rye value to the expected type
func coerceToType(value env.Object, valueType string, es *env.ProgramState) (env.Object, error) {
	switch valueType {
	case "any":
		return value, nil
	case "integer":
		switch v := value.(type) {
		case env.Integer:
			return v, nil
		case env.Decimal:
			return *env.NewInteger(int64(v.Value)), nil
		case env.String:
			// Try to parse string as integer
			result, verr := evalInteger(v)
			if verr != nil {
				return nil, fmt.Errorf("expected integer, got string: %s", v.Value)
			}
			if resultObj, ok := result.(env.Object); ok {
				return resultObj, nil
			}
			return env.ToRyeValue(result), nil
		default:
			return nil, fmt.Errorf("expected integer, got %T", value)
		}
	case "decimal":
		switch v := value.(type) {
		case env.Decimal:
			return v, nil
		case env.Integer:
			return *env.NewDecimal(float64(v.Value)), nil
		case env.String:
			result, verr := evalDecimal(v)
			if verr != nil {
				return nil, fmt.Errorf("expected decimal, got string: %s", v.Value)
			}
			if resultObj, ok := result.(env.Object); ok {
				return resultObj, nil
			}
			return env.ToRyeValue(result), nil
		default:
			return nil, fmt.Errorf("expected decimal, got %T", value)
		}
	case "string", "file":
		switch v := value.(type) {
		case env.String:
			return v, nil
		case env.Uri:
			return *env.NewString(v.GetPath()), nil
		case env.Word:
			return *env.NewString(es.Idx.GetWord(v.Index)), nil
		case env.Integer:
			return *env.NewString(v.Print(*es.Idx)), nil
		default:
			return nil, fmt.Errorf("expected string, got %T", value)
		}
	case "boolean":
		switch v := value.(type) {
		case env.Boolean:
			return v, nil
		case env.Integer:
			return *env.NewBoolean(v.Value != 0), nil
		case env.String:
			result, verr := evalBoolean(v)
			if verr != nil {
				return nil, fmt.Errorf("expected boolean, got %T", value)
			}
			if resultObj, ok := result.(env.Object); ok {
				return resultObj, nil
			}
			return env.ToRyeValue(result), nil
		default:
			return nil, fmt.Errorf("expected boolean, got %T", value)
		}
	default:
		return value, nil
	}
}

// CLI_ParseArgs parses command line arguments (as Rye values) according to the spec
func CLI_ParseArgs(es *env.ProgramState, args env.Block, spec *CLISpec) (*ParsedArgs, map[string]env.Object) {
	result := &ParsedArgs{
		Values:     make(map[string]env.Object),
		Positional: make([]env.Object, 0),
	}
	errors := make(map[string]env.Object)

	// Initialize defaults
	for _, argSpec := range spec.GlobalArgs {
		if !argSpec.IsPositional {
			if argSpec.IsList {
				result.Values[argSpec.Name] = env.NewList(make([]any, 0))
			} else {
				result.Values[argSpec.Name] = argSpec.Default
			}
		}
	}

	// Determine which specs to use (global or subcommand)
	activeSpecs := spec.GlobalArgs
	argsSer := args.Series
	argsSer.Reset()

	// First pass: check for subcommand (a plain word that matches a subcommand name)
	if len(spec.Subcommands) > 0 {
		tempSer := args.Series
		tempSer.Reset()
		foundSubcmdIdx := -1

		for idx := 0; tempSer.Pos() < tempSer.Len(); idx++ {
			obj := tempSer.Pop()

			// Skip flagwords and their values
			if fw, ok := obj.(env.Flagword); ok {
				flagSpec := findFlagSpecForFlagword(spec.GlobalArgs, fw, es.Idx)
				if flagSpec != nil && !flagSpec.IsFlag {
					// Skip the value
					if tempSer.Pos() < tempSer.Len() {
						tempSer.Pop()
						idx++
					}
				}
				continue
			}

			// Check if this word is a subcommand
			if word, ok := obj.(env.Word); ok {
				cmdName := es.Idx.GetWord(word.Index)
				if subCmd, exists := spec.Subcommands[cmdName]; exists {
					result.Command = cmdName
					foundSubcmdIdx = idx
					// Merge subcommand args with global args
					activeSpecs = append(spec.GlobalArgs, subCmd.Args...)

					// Initialize subcommand defaults
					for _, argSpec := range subCmd.Args {
						if !argSpec.IsPositional {
							if argSpec.IsList {
								result.Values[argSpec.Name] = env.NewList(make([]any, 0))
							} else {
								result.Values[argSpec.Name] = argSpec.Default
							}
						}
					}
					break
				}
			}
		}

		// Remove subcommand from processing if found
		if foundSubcmdIdx >= 0 {
			newData := make([]env.Object, 0)
			argsSer.Reset()
			for idx := 0; argsSer.Pos() < argsSer.Len(); idx++ {
				obj := argsSer.Pop()
				if idx != foundSubcmdIdx {
					newData = append(newData, obj)
				}
			}
			args = *env.NewBlock(*env.NewTSeries(newData))
		}
	}

	// Second pass: parse arguments
	argsSer = args.Series
	argsSer.Reset()
	positionalIndex := 0

	for argsSer.Pos() < argsSer.Len() {
		obj := argsSer.Pop()

		switch item := obj.(type) {
		case env.Flagword:
			// Find matching spec
			flagSpec := findFlagSpecForFlagword(activeSpecs, item, es.Idx)
			if flagSpec == nil {
				flagName := item.Print(*es.Idx)
				errors[flagName] = *env.NewString("unknown flag")
				continue
			}

			if flagSpec.IsFlag {
				// Boolean flag - just set to true
				result.Values[flagSpec.Name] = env.NewBoolean(true)
			} else {
				// Option that requires a value
				if argsSer.Pos() >= argsSer.Len() {
					errors[flagSpec.Name] = *env.NewString("requires a value")
					continue
				}
				value := argsSer.Pop()

				// Coerce to expected type
				coerced, err := coerceToType(value, flagSpec.ValueType, es)
				if err != nil {
					errors[flagSpec.Name] = *env.NewString(err.Error())
					continue
				}

				if flagSpec.IsList {
					appendToListCli(result.Values, flagSpec.Name, coerced)
				} else {
					result.Values[flagSpec.Name] = coerced
				}
			}

		default:
			// Positional argument
			posSpecs := findAllPositionalSpecsCli(activeSpecs)
			if positionalIndex < len(posSpecs) || (len(posSpecs) > 0 && posSpecs[len(posSpecs)-1].IsMany) {
				var posSpec *ArgSpec
				if positionalIndex < len(posSpecs) {
					posSpec = posSpecs[positionalIndex]
				} else if len(posSpecs) > 0 && posSpecs[len(posSpecs)-1].IsMany {
					posSpec = posSpecs[len(posSpecs)-1]
				}

				if posSpec != nil {
					coerced, err := coerceToType(item, posSpec.ValueType, es)
					if err != nil {
						errors[posSpec.Name] = *env.NewString(err.Error())
					} else {
						result.Positional = append(result.Positional, coerced)
					}

					if !posSpec.IsMany {
						positionalIndex++
					}
				}
			} else {
				// Extra positional argument
				result.Positional = append(result.Positional, item)
			}
		}
	}

	// Validate required arguments
	for _, argSpec := range activeSpecs {
		if argSpec.IsRequired && !argSpec.IsPositional {
			val, exists := result.Values[argSpec.Name]
			if !exists || isDefaultOrEmpty(val, argSpec.Default) {
				errors[argSpec.Name] = *env.NewString("required")
			}
		}
	}

	// Store positional args in result
	posSpecs := findAllPositionalSpecsCli(activeSpecs)
	for idx, posSpec := range posSpecs {
		if posSpec.IsMany {
			// Collect all remaining positional args from index onwards
			listData := make([]any, 0)
			for j := idx; j < len(result.Positional); j++ {
				listData = append(listData, result.Positional[j])
			}
			result.Values[posSpec.Name] = env.NewList(listData)

			// Check required
			if posSpec.IsRequired && len(listData) == 0 {
				errors[posSpec.Name] = *env.NewString("required")
			}
		} else {
			if idx < len(result.Positional) {
				result.Values[posSpec.Name] = result.Positional[idx]
			} else if posSpec.IsRequired {
				errors[posSpec.Name] = *env.NewString("required")
			} else {
				result.Values[posSpec.Name] = posSpec.Default
			}
		}
	}

	if len(errors) > 0 {
		return result, errors
	}

	return result, nil
}

// Helper functions

func findAllPositionalSpecsCli(specs []ArgSpec) []*ArgSpec {
	result := make([]*ArgSpec, 0)
	for i := range specs {
		if specs[i].IsPositional {
			result = append(result, &specs[i])
		}
	}
	return result
}

func appendToListCli(values map[string]env.Object, name string, value env.Object) {
	if existing, ok := values[name]; ok {
		if list, ok := existing.(*env.List); ok {
			list.Data = append(list.Data, value)
		} else if list, ok := existing.(env.List); ok {
			list.Data = append(list.Data, value)
			values[name] = list
		}
	} else {
		values[name] = env.NewList([]any{value})
	}
}

func isDefaultOrEmpty(val env.Object, defaultVal env.Object) bool {
	switch v := val.(type) {
	case env.String:
		if def, ok := defaultVal.(env.String); ok {
			return v.Value == def.Value
		}
		if def, ok := defaultVal.(*env.String); ok {
			return v.Value == def.Value
		}
		return v.Value == ""
	case *env.String:
		if def, ok := defaultVal.(env.String); ok {
			return v.Value == def.Value
		}
		if def, ok := defaultVal.(*env.String); ok {
			return v.Value == def.Value
		}
		return v.Value == ""
	case env.Void, *env.Void:
		return true
	}
	return false
}

// BuiParseArgs is the main builtin function for parse-args
func BuiParseArgs(es *env.ProgramState, args env.Object, specBlock env.Object) env.Object {
	argsBlock, ok := args.(env.Block)
	if !ok {
		es.FailureFlag = true
		return MakeArgError(es, 1, []env.Type{env.BlockType}, "parse-args")
	}

	spec, ok := specBlock.(env.Block)
	if !ok {
		es.FailureFlag = true
		return MakeArgError(es, 2, []env.Type{env.BlockType}, "parse-args")
	}

	// Parse the specification
	cliSpec, err := CLI_ParseSpec(es, spec)
	if err != nil {
		es.FailureFlag = true
		return env.NewError2(400, "spec error: "+err.Error())
	}

	// Parse the arguments
	result, parseErrs := CLI_ParseArgs(es, argsBlock, cliSpec)
	if parseErrs != nil {
		es.FailureFlag = true
		return env.NewError4(400, "argument parsing error", nil, parseErrs)
	}

	// Convert result to Dict
	resultData := make(map[string]any)
	for k, v := range result.Values {
		resultData[k] = v
	}

	// Add command if subcommand was used
	if result.Command != "" {
		resultData["command"] = *env.NewString(result.Command)
	}

	return *env.NewDict(resultData)
}

var Builtins_cli = map[string]*env.Builtin{

	//
	// ##### CLI Argument Parsing dialect ##### "CLI argument parsing dialect for Rye"
	//
	// The parse-args function takes a block of Rye values (from parsed command line)
	// and a specification block, returning a dictionary with parsed values.
	//
	// Args are expected to be already parsed by the Rye loader, so they come as:
	// - Flagwords: -v, --verbose, -v|verbose
	// - Integers: 123
	// - Strings: "quoted string"
	// - Words: unquoted-word
	// - etc.
	//
	// Tests:
	// equal { parse-args { --verbose } { -v|verbose flag } -> "verbose" } true
	// equal { parse-args { -v } { -v|verbose flag } -> "verbose" } true
	// equal { parse-args { --output "file.txt" } { -o|output string required } -> "output" } "file.txt"
	// equal { parse-args { -o "file.txt" } { -o|output string required } -> "output" } "file.txt"
	// equal { parse-args { --count 5 } { -n|count integer optional 1 } -> "count" } 5
	// equal { parse-args { } { -n|count integer optional 3 } -> "count" } 3
	// equal { parse-args { "file1.txt" "file2.txt" } { _ positional string many } -> "_" |length? } 2
	// error { parse-args { } { -o|output string required } }
	//
	// Subcommand tests:
	// equal { parse-args { init --force } { subcommand { init { -f|force flag } } } -> "command" } "init"
	// equal { parse-args { init --force } { subcommand { init { -f|force flag } } } -> "force" } true
	//
	// Args:
	// * args: Block of Rye values representing command line arguments
	// * spec: Block containing argument specifications
	// Returns:
	// * Dict with parsed argument values or error if parsing fails
	"parse-args": {
		Argsn: 2,
		Doc:   "Parses command line arguments according to a specification block, returning a dictionary with the parsed values.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return BuiParseArgs(es, arg0, arg1)
		},
	},

	// Tests:
	// equal { parse-args\ctx { --verbose } { -v|verbose flag } -> 'verbose } true
	// equal { parse-args\ctx { -o "out.txt" } { -o|output string required } -> 'output } "out.txt"
	//
	// Args:
	// * args: Block of Rye values representing command line arguments
	// * spec: Block containing argument specifications
	// Returns:
	// * Context with parsed argument values or error if parsing fails
	"parse-args\\ctx": {
		Argsn: 2,
		Doc:   "Parses command line arguments according to a specification block, returning a context for easy field access.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result := BuiParseArgs(es, arg0, arg1)
			if es.FailureFlag {
				return result
			}
			switch dict := result.(type) {
			case env.Dict:
				// Convert Dict to Context
				ctx := env.NewEnv(nil)
				for k, v := range dict.Data {
					idx := es.Idx.IndexWord(k)
					if obj, ok := v.(env.Object); ok {
						ctx.Set(idx, obj)
					} else {
						ctx.Set(idx, env.ToRyeValue(v))
					}
				}
				return ctx
			default:
				return result
			}
		},
	},
}
