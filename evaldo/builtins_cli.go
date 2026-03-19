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
	CheckError   string     // Error message if check fails
	Doc          string     // Documentation string
}

// SubcommandSpec holds the specification for a subcommand
type SubcommandSpec struct {
	Name        string
	Args        []ArgSpec
	Subcommands map[string]*SubcommandSpec // Nested subcommands
	Doc         string                     // Documentation for subcommand
}

// CLISpec holds the complete CLI specification
type CLISpec struct {
	GlobalArgs  []ArgSpec
	Subcommands map[string]*SubcommandSpec
	ProgramName string // Optional program name for help
	ProgramDoc  string // Optional program description for help
}

// ParsedArgs holds the result of parsing
type ParsedArgs struct {
	Values      map[string]env.Object
	Command     string   // Full command path: "remote add"
	CommandPath []string // ["remote", "add"]
	Positional  []env.Object
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
			} else if wordName == "program" {
				// Program name for help generation
				if ser.Pos() < ser.Len() {
					if str, ok := ser.Pop().(env.String); ok {
						spec.ProgramName = str.Value
					}
				}
			} else if wordName == "description" {
				// Program description for help generation
				if ser.Pos() < ser.Len() {
					if str, ok := ser.Pop().(env.String); ok {
						spec.ProgramDoc = str.Value
					}
				}
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
				spec.Default = *env.NewBoolean(false)
			case "string":
				spec.ValueType = "string"
				spec.Default = *env.NewString("")
			case "integer":
				spec.ValueType = "integer"
				spec.Default = env.NewInteger(0)
			case "decimal":
				spec.ValueType = "decimal"
				spec.Default = env.NewDecimal(0.0)
			case "boolean":
				spec.ValueType = "boolean"
				spec.Default = *env.NewBoolean(false)
			case "file":
				spec.ValueType = "file"
				spec.Default = *env.NewString("")
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
				// Get the check block and optional error message
				// Syntax: check { block } "error message"
				if ser.Pos() < ser.Len() {
					if block, ok := ser.Pop().(env.Block); ok {
						spec.CheckBlock = &block
						// Check for error message after block
						if ser.Pos() < ser.Len() {
							if str, ok := ser.Peek().(env.String); ok {
								spec.CheckError = str.Value
								ser.Pop()
							}
						}
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
				// Get the check block and optional error message
				if ser.Pos() < ser.Len() {
					if block, ok := ser.Pop().(env.Block); ok {
						spec.CheckBlock = &block
						// Check for error message after block
						if ser.Pos() < ser.Len() {
							if str, ok := ser.Peek().(env.String); ok {
								spec.CheckError = str.Value
								ser.Pop()
							}
						}
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

// parseSubcommands parses the subcommand block (supports nesting)
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
			Name:        cmdName,
			Args:        make([]ArgSpec, 0),
			Subcommands: make(map[string]*SubcommandSpec),
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
				if wordName == "subcommand" {
					// Nested subcommand - recursive parsing
					if subSer.Pos() >= subSer.Len() {
						return nil, fmt.Errorf("expected block after nested 'subcommand' in '%s'", cmdName)
					}
					nestedBlock, ok := subSer.Pop().(env.Block)
					if !ok {
						return nil, fmt.Errorf("expected block after nested 'subcommand' in '%s'", cmdName)
					}
					nestedSubcommands, err := parseSubcommands(es, nestedBlock)
					if err != nil {
						return nil, err
					}
					subSpec.Subcommands = nestedSubcommands
				} else if wordName == "doc" {
					// Subcommand documentation
					if subSer.Pos() < subSer.Len() {
						if str, ok := subSer.Pop().(env.String); ok {
							subSpec.Doc = str.Value
						}
					}
				} else if strings.HasPrefix(wordName, "_") || wordName == "_" {
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

// stringFlagMatches checks if a string flag (e.g., "--verbose", "-v") matches a spec
func stringFlagMatches(flagStr string, spec *ArgSpec) bool {
	if strings.HasPrefix(flagStr, "--") {
		// Long flag
		longName := strings.TrimPrefix(flagStr, "--")
		return spec.LongFlag == longName
	} else if strings.HasPrefix(flagStr, "-") && len(flagStr) > 1 {
		// Short flag
		shortName := strings.TrimPrefix(flagStr, "-")
		return spec.ShortFlag == shortName
	}
	return false
}

// findFlagSpec finds the spec that matches a given flag (flagword or string)
func findFlagSpec(specs []ArgSpec, obj env.Object, idx *env.Idxs) *ArgSpec {
	switch item := obj.(type) {
	case env.Flagword:
		for i := range specs {
			if flagwordMatches(item, &specs[i], idx) {
				return &specs[i]
			}
		}
	case env.String:
		for i := range specs {
			if stringFlagMatches(item.Value, &specs[i]) {
				return &specs[i]
			}
		}
	}
	return nil
}

// isFlag checks if an object looks like a flag
func isFlag(obj env.Object) bool {
	switch item := obj.(type) {
	case env.Flagword:
		return true
	case env.String:
		return strings.HasPrefix(item.Value, "-")
	}
	return false
}

// getFlagName returns the flag name for error reporting
func getFlagName(obj env.Object, idx *env.Idxs) string {
	switch item := obj.(type) {
	case env.Flagword:
		return item.Print(*idx)
	case env.String:
		return item.Value
	}
	return "unknown"
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
			return env.Boolean{Value: v.Value != 0}, nil
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

// validateWithCheck runs the check block against a value
func validateWithCheck(es *env.ProgramState, value env.Object, spec *ArgSpec) error {
	if spec.CheckBlock == nil {
		return nil
	}

	// Save current series position
	ser := es.Ser

	// Create a copy of the block series for evaluation
	blockSer := spec.CheckBlock.Series
	blockSer.Reset()
	es.Ser = blockSer

	// Evaluate block with value injected
	EvalBlockInj(es, value, true)

	// Restore series
	es.Ser = ser

	// Check for error during evaluation
	if es.ErrorFlag {
		es.ErrorFlag = false
		errMsg := spec.CheckError
		if errMsg == "" {
			errMsg = "validation failed"
		}
		return fmt.Errorf("%s", errMsg)
	}

	// Check result - integer > 0 or truthy boolean means success
	switch result := es.Res.(type) {
	case env.Integer:
		if result.Value > 0 {
			return nil
		}
	case env.Boolean:
		if result.Value {
			return nil
		}
	}

	errMsg := spec.CheckError
	if errMsg == "" {
		errMsg = "validation failed"
	}
	return fmt.Errorf("%s", errMsg)
}

// isSubcommandWord checks if a word matches a subcommand name
func isSubcommandWord(obj env.Object, subcommands map[string]*SubcommandSpec, es *env.ProgramState) (string, bool) {
	switch item := obj.(type) {
	case env.Word:
		cmdName := es.Idx.GetWord(item.Index)
		if _, exists := subcommands[cmdName]; exists {
			return cmdName, true
		}
	case env.String:
		// Also support string subcommand names (from command line)
		if _, exists := subcommands[item.Value]; exists {
			return item.Value, true
		}
	}
	return "", false
}

// CLI_ParseArgs parses command line arguments (as Rye values) according to the spec
func CLI_ParseArgs(es *env.ProgramState, args env.Block, spec *CLISpec) (*ParsedArgs, map[string]env.Object) {
	result := &ParsedArgs{
		Values:      make(map[string]env.Object),
		CommandPath: make([]string, 0),
		Positional:  make([]env.Object, 0),
	}
	errors := make(map[string]env.Object)

	// Initialize defaults for global args
	for _, argSpec := range spec.GlobalArgs {
		if !argSpec.IsPositional {
			if argSpec.IsList {
				result.Values[argSpec.Name] = *env.NewList(make([]any, 0))
			} else {
				result.Values[argSpec.Name] = argSpec.Default
			}
		}
	}

	// Convert args to a slice for easier manipulation
	argsSer := args.Series
	argsSer.Reset()
	argsList := make([]env.Object, 0)
	for argsSer.Pos() < argsSer.Len() {
		argsList = append(argsList, argsSer.Pop())
	}

	// Track active specs and current subcommands
	activeSpecs := spec.GlobalArgs
	currentSubcommands := spec.Subcommands

	// Process args, looking for subcommands at each level
	i := 0
	for i < len(argsList) && len(currentSubcommands) > 0 {
		obj := argsList[i]

		// Skip flags and their values
		if isFlag(obj) {
			flagSpec := findFlagSpec(activeSpecs, obj, es.Idx)
			if flagSpec != nil && !flagSpec.IsFlag {
				i += 2 // Skip flag and value
			} else {
				i++ // Skip just the flag
			}
			continue
		}

		// Check if this is a subcommand
		if cmdName, ok := isSubcommandWord(obj, currentSubcommands, es); ok {
			result.CommandPath = append(result.CommandPath, cmdName)
			subCmd := currentSubcommands[cmdName]

			// Merge subcommand args with active specs
			activeSpecs = append(activeSpecs, subCmd.Args...)

			// Initialize subcommand defaults
			for _, argSpec := range subCmd.Args {
				if !argSpec.IsPositional {
					if argSpec.IsList {
						result.Values[argSpec.Name] = *env.NewList(make([]any, 0))
					} else {
						result.Values[argSpec.Name] = argSpec.Default
					}
				}
			}

			// Move to nested subcommands
			currentSubcommands = subCmd.Subcommands

			// Remove this subcommand from args
			argsList = append(argsList[:i], argsList[i+1:]...)
			// Don't increment i - we removed an element
		} else {
			i++
		}
	}

	// Set the full command path
	if len(result.CommandPath) > 0 {
		result.Command = strings.Join(result.CommandPath, " ")
	}

	// Second pass: parse remaining arguments
	positionalIndex := 0
	for i := 0; i < len(argsList); i++ {
		obj := argsList[i]

		if isFlag(obj) {
			// Find matching spec
			flagSpec := findFlagSpec(activeSpecs, obj, es.Idx)
			if flagSpec == nil {
				flagName := getFlagName(obj, es.Idx)
				errors[flagName] = *env.NewString("unknown flag")
				continue
			}

			if flagSpec.IsFlag {
				// Boolean flag - just set to true
				result.Values[flagSpec.Name] = *env.NewBoolean(true)
			} else {
				// Option that requires a value
				if i+1 >= len(argsList) {
					errors[flagSpec.Name] = *env.NewString("requires a value")
					continue
				}
				i++
				value := argsList[i]

				// Coerce to expected type
				coerced, err := coerceToType(value, flagSpec.ValueType, es)
				if err != nil {
					errors[flagSpec.Name] = *env.NewString(err.Error())
					continue
				}

				// Run check block if present
				if err := validateWithCheck(es, coerced, flagSpec); err != nil {
					errors[flagSpec.Name] = *env.NewString(err.Error())
					continue
				}

				if flagSpec.IsList {
					appendToListCli(result.Values, flagSpec.Name, coerced)
				} else {
					result.Values[flagSpec.Name] = coerced
				}
			}
		} else {
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
					coerced, err := coerceToType(obj, posSpec.ValueType, es)
					if err != nil {
						errors[posSpec.Name] = *env.NewString(err.Error())
					} else {
						// Run check block if present
						if err := validateWithCheck(es, coerced, posSpec); err != nil {
							errors[posSpec.Name] = *env.NewString(err.Error())
						} else {
							result.Positional = append(result.Positional, coerced)
						}
					}

					if !posSpec.IsMany {
						positionalIndex++
					}
				}
			} else {
				// Extra positional argument
				result.Positional = append(result.Positional, obj)
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
			result.Values[posSpec.Name] = *env.NewList(listData)

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
		values[name] = *env.NewList([]any{value})
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

// generateHelpText generates help text from a CLI specification
func generateHelpText(es *env.ProgramState, spec *CLISpec, cmdPath []string) string {
	var sb strings.Builder

	// Program name and description
	programName := spec.ProgramName
	if programName == "" {
		programName = "program"
	}

	// If we have a command path, we're generating help for a subcommand
	if len(cmdPath) > 0 {
		programName = programName + " " + strings.Join(cmdPath, " ")
	}

	sb.WriteString(fmt.Sprintf("Usage: %s [options]", programName))

	// Get the relevant subcommand spec if we have a path
	var activeSubcommands map[string]*SubcommandSpec
	var activeArgs []ArgSpec

	if len(cmdPath) > 0 {
		// Navigate to the subcommand
		currentSubs := spec.Subcommands
		var currentSpec *SubcommandSpec
		for _, cmd := range cmdPath {
			if sub, exists := currentSubs[cmd]; exists {
				currentSpec = sub
				currentSubs = sub.Subcommands
			}
		}
		if currentSpec != nil {
			activeSubcommands = currentSpec.Subcommands
			activeArgs = append(spec.GlobalArgs, currentSpec.Args...)
		}
	} else {
		activeSubcommands = spec.Subcommands
		activeArgs = spec.GlobalArgs
	}

	// Add subcommand indicator if there are subcommands
	if len(activeSubcommands) > 0 {
		sb.WriteString(" [command]")
	}

	// Add positional args indicator
	posSpecs := findAllPositionalSpecsCli(activeArgs)
	for _, pos := range posSpecs {
		name := strings.TrimPrefix(pos.Name, "_")
		if name == "" {
			name = "arg"
		}
		if pos.IsMany {
			sb.WriteString(fmt.Sprintf(" [%s...]", name))
		} else if pos.IsRequired {
			sb.WriteString(fmt.Sprintf(" <%s>", name))
		} else {
			sb.WriteString(fmt.Sprintf(" [%s]", name))
		}
	}

	sb.WriteString("\n")

	// Program description
	if spec.ProgramDoc != "" && len(cmdPath) == 0 {
		sb.WriteString(fmt.Sprintf("\n%s\n", spec.ProgramDoc))
	}

	// Options section
	flagSpecs := make([]ArgSpec, 0)
	for _, arg := range activeArgs {
		if !arg.IsPositional {
			flagSpecs = append(flagSpecs, arg)
		}
	}

	if len(flagSpecs) > 0 {
		sb.WriteString("\nOptions:\n")
		for _, arg := range flagSpecs {
			var flagStr string
			if arg.ShortFlag != "" && arg.LongFlag != "" {
				flagStr = fmt.Sprintf("  -%s, --%s", arg.ShortFlag, arg.LongFlag)
			} else if arg.LongFlag != "" {
				flagStr = fmt.Sprintf("      --%s", arg.LongFlag)
			} else {
				flagStr = fmt.Sprintf("  -%s", arg.ShortFlag)
			}

			// Add value placeholder for non-flags
			if !arg.IsFlag {
				typeName := strings.ToUpper(arg.ValueType)
				if typeName == "ANY" {
					typeName = "VALUE"
				}
				flagStr = fmt.Sprintf("%s %s", flagStr, typeName)
			}

			// Pad to align descriptions
			for len(flagStr) < 28 {
				flagStr += " "
			}

			// Add doc and required/default info
			doc := arg.Doc
			if arg.IsRequired {
				if doc != "" {
					doc += " (required)"
				} else {
					doc = "(required)"
				}
			} else if !arg.IsFlag {
				switch def := arg.Default.(type) {
				case env.String:
					if def.Value != "" {
						if doc != "" {
							doc += fmt.Sprintf(" (default: %q)", def.Value)
						} else {
							doc = fmt.Sprintf("(default: %q)", def.Value)
						}
					}
				case *env.String:
					if def.Value != "" {
						if doc != "" {
							doc += fmt.Sprintf(" (default: %q)", def.Value)
						} else {
							doc = fmt.Sprintf("(default: %q)", def.Value)
						}
					}
				case env.Integer:
					if doc != "" {
						doc += fmt.Sprintf(" (default: %d)", def.Value)
					} else {
						doc = fmt.Sprintf("(default: %d)", def.Value)
					}
				case *env.Integer:
					if doc != "" {
						doc += fmt.Sprintf(" (default: %d)", def.Value)
					} else {
						doc = fmt.Sprintf("(default: %d)", def.Value)
					}
				}
			}

			sb.WriteString(fmt.Sprintf("%s%s\n", flagStr, doc))
		}
	}

	// Subcommands section
	if len(activeSubcommands) > 0 {
		sb.WriteString("\nCommands:\n")
		for name, sub := range activeSubcommands {
			cmdStr := fmt.Sprintf("  %s", name)
			for len(cmdStr) < 20 {
				cmdStr += " "
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", cmdStr, sub.Doc))
		}
		sb.WriteString(fmt.Sprintf("\nRun '%s <command> --help' for more information on a command.\n", programName))
	}

	return sb.String()
}

// formatParseErrors formats error dict into user-friendly messages
func formatParseErrors(errors map[string]env.Object) string {
	var sb strings.Builder

	for name, errObj := range errors {
		var errStr string
		if s, ok := errObj.(env.String); ok {
			errStr = s.Value
		} else if s, ok := errObj.(*env.String); ok {
			errStr = s.Value
		} else {
			errStr = errObj.Print(env.Idxs{})
		}

		// Format the flag name nicely
		flagName := name
		if !strings.HasPrefix(name, "-") && !strings.HasPrefix(name, "_") {
			flagName = "--" + name
		}

		sb.WriteString(fmt.Sprintf("Error: %s %s\n", flagName, errStr))
	}

	return strings.TrimSuffix(sb.String(), "\n")
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

	// Add command-path as a list
	if len(result.CommandPath) > 0 {
		pathList := make([]any, len(result.CommandPath))
		for i, cmd := range result.CommandPath {
			pathList[i] = *env.NewString(cmd)
		}
		resultData["command-path"] = *env.NewList(pathList)
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
	// Args can be:
	// - Flagwords: -v, --verbose, -v|verbose (from Rye code)
	// - Strings: "--verbose", "-o" (from command line via rye .Args?)
	// - Integers: 123
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
	// String flag tests (for command line args):
	// equal { parse-args { "--verbose" } { -v|verbose flag } -> "verbose" } true
	// equal { parse-args { "-v" } { -v|verbose flag } -> "verbose" } true
	// equal { parse-args { "-o" "file.txt" } { -o|output string required } -> "output" } "file.txt"
	//
	// Subcommand tests:
	// equal { parse-args { init --force } { subcommand { init { -f|force flag } } } -> "command" } "init"
	// equal { parse-args { init --force } { subcommand { init { -f|force flag } } } -> "force" } true
	//
	// Nested subcommand tests:
	// equal { parse-args { remote add } { subcommand { remote { subcommand { add { } } } } } -> "command" } "remote add"
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

	// Tests:
	// ; Basic help generation
	// equal { generate-help { -v|verbose flag doc "Enable verbose output" } |type? } 'string
	//
	// Args:
	// * spec: Block containing argument specifications
	// Returns:
	// * String containing formatted help text
	"generate-help": {
		Argsn: 1,
		Doc:   "Generates help text from a CLI specification block.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			spec, ok := arg0.(env.Block)
			if !ok {
				es.FailureFlag = true
				return MakeArgError(es, 1, []env.Type{env.BlockType}, "generate-help")
			}

			cliSpec, err := CLI_ParseSpec(es, spec)
			if err != nil {
				es.FailureFlag = true
				return env.NewError2(400, "spec error: "+err.Error())
			}

			helpText := generateHelpText(es, cliSpec, nil)
			return *env.NewString(helpText)
		},
	},

	// Tests:
	// ; Help for specific subcommand
	// equal { generate-help\command { subcommand { init { -f|force flag } } } "init" |type? } 'string
	//
	// Args:
	// * spec: Block containing argument specifications
	// * command: String with command path (e.g., "remote add")
	// Returns:
	// * String containing formatted help text for the subcommand
	"generate-help\\command": {
		Argsn: 2,
		Doc:   "Generates help text for a specific subcommand.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			spec, ok := arg0.(env.Block)
			if !ok {
				es.FailureFlag = true
				return MakeArgError(es, 1, []env.Type{env.BlockType}, "generate-help\\command")
			}

			cmdPath := make([]string, 0)
			switch cmd := arg1.(type) {
			case env.String:
				if cmd.Value != "" {
					cmdPath = strings.Split(cmd.Value, " ")
				}
			case env.Block:
				ser := cmd.Series
				ser.Reset()
				for ser.Pos() < ser.Len() {
					obj := ser.Pop()
					if s, ok := obj.(env.String); ok {
						cmdPath = append(cmdPath, s.Value)
					} else if w, ok := obj.(env.Word); ok {
						cmdPath = append(cmdPath, es.Idx.GetWord(w.Index))
					}
				}
			default:
				es.FailureFlag = true
				return MakeArgError(es, 2, []env.Type{env.StringType, env.BlockType}, "generate-help\\command")
			}

			cliSpec, err := CLI_ParseSpec(es, spec)
			if err != nil {
				es.FailureFlag = true
				return env.NewError2(400, "spec error: "+err.Error())
			}

			helpText := generateHelpText(es, cliSpec, cmdPath)
			return *env.NewString(helpText)
		},
	},

	// Tests:
	// equal { format-parse-errors dict { "output" "required" } |type? } 'string
	//
	// Args:
	// * errors: Dict with error information (from failed parse-args)
	// Returns:
	// * String with formatted error messages
	"format-parse-errors": {
		Argsn: 1,
		Doc:   "Formats parse errors into human-readable messages.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch errObj := arg0.(type) {
			case env.Dict:
				errors := make(map[string]env.Object)
				for k, v := range errObj.Data {
					if obj, ok := v.(env.Object); ok {
						errors[k] = obj
					}
				}
				return *env.NewString(formatParseErrors(errors))
			case *env.Error:
				// Extract details from error if it has them
				if errObj.Values != nil {
					return *env.NewString(formatParseErrors(errObj.Values))
				}
				return *env.NewString(errObj.Message)
			default:
				es.FailureFlag = true
				return MakeArgError(es, 1, []env.Type{env.DictType}, "format-parse-errors")
			}
		},
	},
}
