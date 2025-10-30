package evaldo

import (
	"github.com/refaktor/rye/env"
)

// matchBlock is a helper function that matches a block against a pattern
// It returns an integer 1 if the match is successful, or an error if the match fails
func matchBlock(ps *env.ProgramState, block env.Block, pattern env.Block) env.Object {
	// Check if the block and pattern have the same length
	if block.Series.Len() != pattern.Series.Len() {
		ps.FailureFlag = true
		return MakeBuiltinError(ps, "Block and pattern have different lengths", "match-block")
	}

	// Iterate through the pattern and bind values
	for i := 0; i < pattern.Series.Len(); i++ {
		patternItem := pattern.Series.Get(i)
		blockValue := block.Series.Get(i)

		// If the pattern item is a word, bind the block value to it
		// If it's an xword, check the type of the block value
		// If it's a block with mode 1 (square brackets), evaluate it with the block value injected
		// Otherwise, check if the pattern item equals the block value
		switch word := patternItem.(type) {
		case env.Word:
			ps.Ctx.Set(word.Index, blockValue)
		case env.Tagword:
			ps.Ctx.Set(word.Index, blockValue)
		case env.Getword:
			// Get the value from the context and check if it equals the block value
			contextValue, found := ps.Ctx.Get(word.Index)
			if !found {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Word not found in context: "+ps.Idx.GetWord(word.Index), "match-block")
			}
			if !contextValue.Equal(blockValue) {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Value doesn't match context value", "match-block")
			}
		case env.Xword:
			// Check if the block value matches the expected type
			// Compare the type of the block value with the type specified by the xword
			if int(blockValue.Type()) != word.Index {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Expected type "+ps.Idx.GetWord(word.Index), "match-block")
			}
		case env.Block:
			// If the pattern item is a block with mode 1 (square brackets), evaluate it with the block value injected
			if word.Mode == 1 {
				// Save current series
				ser := ps.Ser
				// Set series to the block
				ps.Ser = word.Series
				// Evaluate the block with the block value injected
				EvalBlockInjMultiDialect(ps, blockValue, true)
				if ps.ErrorFlag {
					ps.Ser = ser
					return ps.Res
				}
				// Restore series
				ps.Ser = ser
				// Continue with the next pattern item
				continue
			} else {
				// If the pattern item is a block, recursively match it with the block value
				switch blockValueBlock := blockValue.(type) {
				case env.Block:
					// Recursively match the nested block
					result := matchBlock(ps, blockValueBlock, word)
					if ps.FailureFlag {
						return result
					}
				default:
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected a block for nested pattern matching", "match-block")
				}
			}
		default:
			// If the pattern item is not a word or block, check if it equals the block value
			if !patternItem.Equal(blockValue) {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Pattern item doesn't match block value", "match-block")
			}
		}
	}

	// Return success
	return *env.NewInteger(1)
}

var Builtins_match = map[string]*env.Builtin{
	// Tests:
	// Basic matching with words:
	// equal { match-block { 1 2 3 } { a b c } |a } 1
	// equal { match-block { 1 2 3 } { a b c } |b } 2
	// equal { match-block { 1 2 3 } { a b c } |c } 3
	// equal { try { match-block { 1 2 } { a b c } } |failed? } 1
	// equal { try { match-block { 1 2 3 4 } { a b c } } |failed? } 1
	//
	// Matching with literal values:
	// equal { match-block { 1 "test" 3 } { a "test" b } |a } 1
	// equal { match-block { 1 "test" 3 } { a "test" b } |b } 3
	// equal { try { match-block { 1 "hi" 3 } { a "bye" b } } |failed? } 1
	//
	// Matching with nested blocks:
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } |a } 123
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } |b } "hi"
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } |c } 44
	// equal { match-block { 123 { "hi" 44 } 789 } { a { b c } d } |d } 789
	// equal { match-block { 1 { 2 { 3 4 } 5 } 6 } { x { y { z w } v } u } |z } 3
	// equal { match-block { 1 { 2 { 3 4 } 5 } 6 } { x { y { z w } v } u } |w } 4
	// equal { try { match-block { 1 { 2 3 } 4 } { a { b "wrong" } c } } |failed? } 1
	//
	// Type checking with xwords:
	// equal { match-block { 123 } { <integer> } } 1
	// equal { match-block { "hello" } { <string> } } 1
	// equal { match-block { { 1 2 3 } } { <block> } } 1
	// equal { match-block { 123.45 } { <decimal> } } 1
	// equal { try { match-block { "not an integer" } { <integer> } } |failed? } 1
	// equal { try { match-block { 123 } { <string> } } |failed? } 1
	//
	// Matching with get-words (match against values in the current context):
	// equal { x: 234 match-block { 123 234 } { 123 ?x } } 1
	// equal { x: 234 try { match-block { 123 456 } { 123 ?x } } |failed? } 1
	//
	// Evaluating blocks with the value injected:
	// x: 0 , match-block { 1000 } { [ x:: . ] } , equal x 1000
	// a: 0 , b: 0 , match-block { 100 200 300 } { [ a:: . ] [ b:: . ] c } , equal a 100 , equal b 200 , equal c 300
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
		Doc:   "Tries to match and deconstruct a block into multiple values, with support for literal values, type checking, get-words, code evaluation, and nested blocks. If it doesn't match, returns failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				switch pattern := arg1.(type) {
				case env.Block:
					// Check if the block and pattern have the same length
					if block.Series.Len() != pattern.Series.Len() {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Block and pattern have different lengths", "match-block")
					}

					// Iterate through the pattern and bind values
					for i := 0; i < pattern.Series.Len(); i++ {
						patternItem := pattern.Series.Get(i)
						blockValue := block.Series.Get(i)

						// If the pattern item is a word, bind the block value to it
						// Otherwise, check if the pattern item equals the block value
						switch word := patternItem.(type) {
						case env.Word:
							ps.Ctx.Set(word.Index, blockValue)
						case env.Tagword:
							ps.Ctx.Set(word.Index, blockValue)
						case env.Getword:
							// Get the value from the context and check if it equals the block value
							contextValue, found := ps.Ctx.Get(word.Index)
							if !found {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, "Word not found in context: "+ps.Idx.GetWord(word.Index), "match-block")
							}
							if !contextValue.Equal(blockValue) {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, "Value doesn't match context value", "match-block")
							}
						case env.Xword:
							// Check if the block value matches the expected type
							// Compare the type of the block value with the type specified by the xword
							if int(blockValue.Type()) != word.Index {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, "Expected type "+ps.Idx.GetWord(word.Index), "match-block")
							}
						case env.Block:
							// If the pattern item is a block with mode 1 (square brackets), evaluate it with the block value injected
							if word.Mode == 1 {
								// Save current series
								ser := ps.Ser
								// Set series to the block
								ps.Ser = word.Series
								// Evaluate the block with the block value injected
								EvalBlockInjMultiDialect(ps, blockValue, true)
								if ps.ErrorFlag {
									ps.Ser = ser
									return ps.Res
								}
								// Restore series
								ps.Ser = ser
								// Continue with the next pattern item
								continue
							} else {
								// If the pattern item is a block, recursively match it with the block value
								switch blockValueBlock := blockValue.(type) {
								case env.Block:
									// Recursively match the nested block
									result := matchBlock(ps, blockValueBlock, word)
									if ps.FailureFlag {
										return result
									}
								default:
									ps.FailureFlag = true
									return MakeBuiltinError(ps, "Expected a block for nested pattern matching", "match-block")
								}
							}
						default:
							// If the pattern item is not a word or block, check if it equals the block value
							if !patternItem.Equal(blockValue) {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, "Pattern item doesn't match block value", "match-block")
							}
						}
					}

					// Return success
					return *env.NewInteger(1)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "match-block")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "match-block")
			}
		},
	},
}
