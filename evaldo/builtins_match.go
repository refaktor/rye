package evaldo

import (
	"github.com/refaktor/rye/env"
)

type matchMode struct {
	tailMode int
}

// matchValue is a helper function that matches a value against a pattern
// It returns an integer 1 if the match is successful, or an error if the match fails
// Both value and pattern can be either single values or blocks
func matchValue(ps *env.ProgramState, value env.Object, pattern env.Object, mmode matchMode) (env.Object, matchMode) {
	// Check if both value and pattern are blocks
	valueBlock, valueIsBlock := value.(env.Block)
	patternBlock, patternIsBlock := pattern.(env.Block)

	// If pattern is a block, value must also be a block
	if patternIsBlock {
		if !valueIsBlock {
			ps.FailureFlag = true
			return MakeBuiltinError(ps, "Pattern is a block but value is not", "match"), mmode
		}

		// Both are blocks - check if they have the same length
		// if valueBlock.Series.Len() != patternBlock.Series.Len() {
		//	ps.FailureFlag = true
		//	return MakeBuiltinError(ps, "Block and pattern have different lengths", "match"), mmode
		// }

		j := 0
		// Iterate through the pattern and bind values
		for i := 0; i < patternBlock.Series.Len(); i++ {
			patternItem := patternBlock.Series.Get(i)
			var blockValue env.Object
			if mmode.tailMode == 0 {
				blockValue = valueBlock.Series.Get(j)
			} else {
				blockValue = *env.NewBlock(*env.NewTSeries(valueBlock.Series.S[j:]))
			}

			var result env.Object
			// Match each item in the block recursively
			result, mmode = matchValue(ps, blockValue, patternItem, mmode)
			if ps.FailureFlag {
				return result, mmode
			}
			if mmode.tailMode == 0 {
				j += 1
			}
		}

		if j == valueBlock.Series.Len() || mmode.tailMode == 1 {
			return *env.NewInteger(1), mmode
		} else {
			return makeError(ps, "Pattern too short for data"), mmode
		}
	}

	// Pattern is not a block - it's a single pattern element
	// Handle different pattern types
	switch word := pattern.(type) {
	case env.Word:
		ps.Ctx.Mod(word.Index, value)
	case env.Tagword:
		ps.Ctx.Mod(word.Index, value)
	case env.Getword:
		// Get the value from the context and check if it equals the value
		contextValue, found := ps.Ctx.Get(word.Index)
		if !found {
			ps.FailureFlag = true
			return MakeBuiltinError(ps, "Word not found in context: "+ps.Idx.GetWord(word.Index), "match"), mmode
		}
		if !contextValue.Equal(value) {
			ps.FailureFlag = true
			return MakeBuiltinError(ps, "Value doesn't match context value", "match"), mmode
		}
	case env.LModword:
		// Get the value from the context and check if it equals the value
		lmod, found := ps.Idx.GetIndex("")
		if found && word.Index == lmod {
			mmode.tailMode = 1
		}
	case env.Xword:
		// Check if the value matches the expected type
		if int(value.Type()) != word.Index {
			ps.FailureFlag = true
			return MakeBuiltinError(ps, "Expected type "+ps.Idx.GetWord(word.Index), "match"), mmode
		}
	case env.Block:
		// If the pattern is a block with mode 1 (square brackets), evaluate it with the value injected
		if word.Mode == 1 {
			// Save current series
			ser := ps.Ser
			// Set series to the block
			ps.Ser = word.Series
			// Evaluate the block with the value injected
			EvalBlockInjMultiDialect(ps, value, true)
			if ps.ErrorFlag {
				ps.Ser = ser
				return ps.Res, mmode
			}
			// Restore series
			ps.Ser = ser
		} else {
			// Pattern is a normal block, but we're matching against a single value
			// This should fail
			ps.FailureFlag = true
			return MakeBuiltinError(ps, "Pattern is a block but value is not", "match"), mmode
		}
	default:
		// If the pattern is not a word or block, check if it equals the value
		if !pattern.Equal(value) {
			ps.FailureFlag = true
			return MakeBuiltinError(ps, "Pattern doesn't match value", "match"), mmode
		}
	}

	// Return success
	return *env.NewInteger(1), mmode
}

// matchBlock is a helper function that matches a block against a pattern
// It returns an integer 1 if the match is successful, or an error if the match fails
// This is kept for backwards compatibility with match-block builtin
func matchBlock(ps *env.ProgramState, block env.Block, pattern env.Block) env.Object {
	v, _ := matchValue(ps, block, pattern, matchMode{0})
	return v
}

var Builtins_match = map[string]*env.Builtin{

	//
	// ##### Match dialect ##### "pattern matching and deconstruction"
	//
	// Tests:
	// ; Basic matching with words:
	// equal { match-block { 1 2 3 } { a b c } a } 1
	// equal { match-block { 1 2 3 } { a b c } b } 2
	// equal { match-block { 1 2 3 } { a b c } c } 3
	// ; error { match-block { 1 2 } { a b c } }
	// ; error { match-block { 1 2 3 4 } { a b c } }
	//
	// ; Matching with literal values:
	// equal { match-block { 1 "test" 3 } { a "test" b } a } 1
	// equal { match-block { 1 "test" 3 } { a "test" b } b } 3
	// ; equal { try { match-block { 1 "hi" 3 } { a "bye" b } } |failed? } 1
	//
	// ; Matching with nested blocks:
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } a } 123
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } b } "hi"
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } c } 44
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } d } 789
	// equal { match-block { 1 { 2 { 3 4 } 5 } 6 } { x { y { z w } v } u } z } 3
	// equal { match-block { 1 { 2 { 3 4 } 5 } 6 } { x { y { z w } v } u } w } 4
	// ; equal { try { match-block { 1 { 2 3 } 4 } { a { b "wrong" } c } } |failed? } 1
	//
	// ; Type checking with xwords:
	// equal { match-block { 123 } { <integer> } } true
	// equal { match-block { "hello" } { <string> } } true
	// equal { match-block { { 1 2 3 } } { <block> } } true
	// equal { match-block { 123.45 } { <decimal> } } true
	// ; equal { try { match-block { "not an integer" } { <integer> } } |failed? } true
	// ; equal { try { match-block { 123 } { <string> } } |failed? } true
	//
	// ;Matching with get-words (match against values in the current context):
	// equal { x: 234 match-block { 123 234 } { 123 ?x } } true
	// ; equal { x: 234 try { match-block { 123 456 } { 123 ?x } } |failed? } true
	//
	// ; Evaluating blocks with the value injected:
	// ; x: 0 , match-block { 1000 } { [ x:: . ] } , equal x 1000
	// ; a: 0 , b: 0 , match-block { 100 200 300 } { [ a:: . ] [ b:: . ] c } , equal a 100 , equal b 200 , equal c 300
	// Args:
	// * block: Block to match and deconstruct
	// * pattern: Block containing:
	//   - Words to bind values to
	//   - Literal values to match
	//   - Xwords to check value types (e.g., <integer>, <string>)
	//   - Get-words to match against values in the current context (e.g., ?x)
	//   - Square brackets [ ] to evaluate code with the value injected
	//   - Nested blocks for recursive matching
	// Returns:
	// * Integer 1 if successful, or failure if the match fails
	"match-block": {
		Argsn: 2,
		Doc:   "Tries to match and deconstruct a value into multiple values, with support for literal values, type checking, get-words, code evaluation, and nested blocks. First argument can be any value. If it doesn't match, returns failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			value := arg0

			// Convert value to a block if it's not already
			var valueBlock env.Block
			switch v := value.(type) {
			case env.Block:
				valueBlock = v
			default:
				// Wrap single value in a block
				valueBlock = env.Block{
					Series: *env.NewTSeries([]env.Object{value}),
				}
			}

			// Pattern must be a block
			switch pattern := arg1.(type) {
			case env.Block:
				return matchBlock(ps, valueBlock, pattern)
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "match-block")
			}
		},
	},

	// match takes a value and a block of pattern/action pairs
	// Each pair consists of:
	//   1. A pattern (can be a single element or a block for matching multiple values)
	//   2. A code block to evaluate if the pattern matches
	// It tests each pair in sequence until one matches, then evaluates and returns the result
	// If no pattern matches, it returns failure
	//
	// Tests:
	// ; Basic type matching with single values:
	// equal { match 5 { <integer> { "got integer" } <string> { "got string" } } } "got integer"
	// equal { match "hello" { <integer> { "got integer" } <string> { "got string" } } } "got string"
	// equal { match 3.14 { <integer> { "int" } <decimal> { "decimal" } <string> { "string" } } } "decimal"
	//
	// ; Block deconstruction (simple):
	// equal { match { 1 2 } { { a b } { a + b } { x } { x * 2 } } } 3
	// equal { match { 101 202 } { { } { 0 } { a b } { a + b } } } 303
	// equal { match { 10 20 30 } { { x y z } { x + y + z } } } 60
	//
	// ; Literal value matching:
	// equal { match 100 { 50 { "fifty" } 100 { "hundred" } 200 { "two hundred" } } } "hundred"
	// equal { match { 100 } { { 50 } { "fifty" } { 100 } { "hundred" } { 200 } { "two hundred" } } } "hundred"
	//
	// ; Partial pattern matching:
	// equal { match { 10 20 } { { a 20 } { a * 2 } { 10 b } { b + 5 } } } 20
	// equal { match { 10 20 } { { 10 b } { b + 5 } { a 20 } { a * 2 } } } 25
	//
	// ; Rule blocks with arity and type patterns:
	// equal { match { } { { } { 'no-data } { a b } { 'has-data } } } 'no-data
	// equal { match { "Jim" "30" } { { <string> <string> } { 'bad-format } { <string> <integer> } { 'good-format } } } 'bad-format
	// equal { match { "Jim" 30 } { { <string> <string> } { 'bad-format } { <string> <integer> } { 'good-format } } } 'good-format
	//
	// ; Complex nested block patterns:
	// equal { match { 200 { "OK" 30 } } { { code { "OK" result } } { result } { code { "ERR" } } { 0 } } } 30
	// equal { match { 404 { "ERR" } } { { code { "OK" result } } { result } { code { "ERR" } } { code } } } 404
	// equal { match { "user" { "John" "Doe" 30 } } { { "user" { first last age } } { first } } } "John"
	//
	// ; First :: rest pattern (head and tail matching):
	// equal { match { 11 22 33 } { { a :: bb } { a } } } 11
	// equal { match { 11 22 33 } { { a :: bb } { bb } } } { 22 33 }
	// equal { match { 11 22 33 } { { a b :: cc } { a + b } } } 33
	// equal { match { 11 22 33 44 } { { a b :: cc } { cc } } } { 33 44 }
	// equal { match { { 12 23 } { 23 34 } { 34 45 } } { { { a b } :: more } { a + b } } } 35
	//
	// ; Mixed type patterns:
	// equal { match { 1 2 3 } { { <integer> <integer> <integer> } { 'three-ints } { <string> <string> } { 'two-strings } } } 'three-ints
	// equal { match { "hello" "world" } { { <integer> <integer> <integer> } { 'three-ints } { <string> <string> } { 'two-strings } } } 'two-strings
	//
	// ; Error handling:
	// ; equal { try { match 5.5 { <integer> { "got integer" } <string> { "got string" } } } |failed? } true
	// ; equal { try { match { 1 2 3 } { { a b } { a + b } } } |failed? } true
	//
	// Args:
	// * value: Any value to match against patterns (can be a block or any other value)
	// * patterns: Block containing pairs:
	//   - First element: pattern (can be a single element like <integer> or a block like { a b })
	//   - Second element: code block to evaluate if pattern matches
	// Returns:
	// * The result of evaluating the code block of the first matching pattern, or failure if no pattern matches
	"match": {
		Argsn: 2,
		Doc:   "Matches a value against a series of pattern/action pairs. Pattern can be a single element or block. Returns the result of the first matching action, or failure if none match.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			value := arg0

			switch patternsBlock := arg1.(type) {
			case env.Block:
				// Iterate through pairs in the patterns block
				if patternsBlock.Series.Len()%2 != 0 {
					return MakeBuiltinError(ps, "Patterns block must contain an even number of elements (pairs)", "match")
				}

				for i := 0; i < patternsBlock.Series.Len(); i += 2 {
					pattern := patternsBlock.Series.Get(i)
					action := patternsBlock.Series.Get(i + 1)

					// Action must be a block
					actionBlock, actionOk := action.(env.Block)
					if !actionOk {
						return MakeBuiltinError(ps, "Action must be a block", "match")
					}

					// Save the current failure flag state
					oldFailureFlag := ps.FailureFlag
					ps.FailureFlag = false

					// Try to match the value against the pattern directly
					// This allows matchValue to properly distinguish between single values and blocks
					matchResult, _ := matchValue(ps, value, pattern, matchMode{0})

					// If match succeeded (no failure flag), evaluate and return the action
					if !ps.FailureFlag {
						// Restore the old failure flag
						ps.FailureFlag = oldFailureFlag

						// Evaluate the action block
						ser := ps.Ser
						ps.Ser = actionBlock.Series
						EvalBlockInj(ps, value, true)
						ps.Ser = ser
						return ps.Res
					}

					// Match failed, restore the old failure flag and continue to next pattern
					ps.FailureFlag = oldFailureFlag
					_ = matchResult // Suppress unused variable warning
				}

				// No pattern matched, return failure
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "No pattern matched the value", "match")
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "match")
			}
		},
	},
}
