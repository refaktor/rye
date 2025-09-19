package evaldo

import (
	"regexp"

	"github.com/refaktor/rye/env"
)

var Builtins_regexp = map[string]*env.Builtin{

	//
	// ##### Regexp #####  "Go like Regular expressions"
	//
	// Tests:
	//  equal { regexp "[0-9]" |type? } 'native
	// Args:
	// * pattern: String containing a regular expression pattern
	// Returns:
	// * native regexp object or error if pattern is invalid
	"regexp": {
		Argsn: 1,
		Doc:   "Creates a compiled regular expression object from a pattern string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				val, err := regexp.Compile(s.Value)
				if err != nil {
					return MakeError(ps, err.Error())
				}
				return *env.NewNative(ps.Idx, val, "regexp")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "regexp")
			}
		},
	},

	// Tests:
	//  equal { regexp "[0-9]" |Is-match "5" } 1
	//  equal { regexp "[0-9]" |Is-match "a" } 0
	// Args:
	// * regexp: Native regexp object
	// * input: String to test against the pattern
	// Returns:
	// * integer 1 if the string matches the pattern, 0 otherwise
	"regexp//Is-match": {
		Argsn: 2,
		Doc:   "Tests if a string matches the regular expression pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).MatchString(val.Value)
					if res {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//Is-match")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//Is-match")
			}
		},
	},

	// Tests:
	//  equal { regexp "x([0-9]+)y" |Submatch? "x123y" } "123"
	// Args:
	// * regexp: Regular expression with capturing groups
	// * input: String to search in
	// Returns:
	// * string containing the first captured group or error if no submatch found
	"regexp//Submatch?": {
		Argsn: 2,
		Doc:   "Extracts the first captured group from a string using the regular expression.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindStringSubmatch(val.Value)
					if len(res) > 1 {
						return *env.NewString(res[1])
					} else {
						return MakeBuiltinError(ps, "No submatch.", "regexp//Submatch?")
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//Submatch?")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//Submatch?")
			}
		},
	},

	// Tests:
	//  equal { regexp "x([0-9]+)y" |Submatches? "x123y x234y" } { "123" }
	// Args:
	// * regexp: Regular expression with capturing groups
	// * input: String to search in
	// Returns:
	// * block containing all captured groups from the first match or error if no match found
	"regexp//Submatches?": {
		Argsn: 2,
		Doc:   "Extracts all captured groups from the first match as a block of strings.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindStringSubmatch(val.Value)
					if len(res) > 0 {
						col1 := make([]env.Object, len(res)-1)
						for i, row := range res {
							if i > 0 {
								col1[i-1] = *env.NewString(row)
							}
						}
						return *env.NewBlock(*env.NewTSeries(col1))
					}
					return MakeBuiltinError(env1, "No results", "Submatches?")
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	// Tests:
	//  equal { regexp "x([0-9]+)(y+)?" |Submatches\all? "x11yy x22" } { { "11" "yy" } { "22" "" } }
	// Args:
	// * regexp: Regular expression with capturing groups
	// * input: String to search in
	// Returns:
	// * block of blocks, each inner block containing the captured groups from one match
	"regexp//Submatches\\all?": {
		Argsn: 2,
		Doc:   "Extracts all captured groups from all matches as a nested block structure.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindAllStringSubmatch(val.Value, -1)
					if len(res) > 0 {
						blks := make([]env.Object, len(res))
						for i, mtch := range res {
							strs := make([]env.Object, len(mtch)-1)
							for j, row := range mtch {
								if j > 0 {
									strs[j-1] = *env.NewString(row)
								}
							}
							blks[i] = *env.NewBlock(*env.NewTSeries(strs))
						}
						return *env.NewBlock(*env.NewTSeries(blks))
					}
					return MakeBuiltinError(env1, "No results", "Submatches?")
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	// Tests:
	//  equal { regexp "[0-9]+" |Find-all "x123y x234y" } { "123" "234" }
	// Args:
	// * regexp: Regular expression pattern
	// * input: String to search in
	// Returns:
	// * block containing all matching substrings or error if no matches found
	"regexp//Find-all": {
		Argsn: 2,
		Doc:   "Finds all substrings matching the regular expression and returns them as a block.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindAllString(val.Value, -1)
					if len(res) > 0 {
						col1 := make([]env.Object, len(res))
						for i, row := range res {
							col1[i] = *env.NewString(row)
						}
						return *env.NewBlock(*env.NewTSeries(col1))
					}
					return MakeBuiltinError(env1, "No results", "Find-all")
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	// Tests:
	//	equal { regexp "[0-9]+c+" |Match? "aa33bb55cc" } "55cc"
	// Args:
	// * regexp: Regular expression pattern
	// * input: String to search in
	// Returns:
	// * string containing the first match or empty string if no match found
	"regexp//Match?": {
		Argsn: 2,
		Doc:   "Finds the first substring matching the regular expression.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindString(val.Value)
					if len(res) > 0 {
						return *env.NewString(res)
					}
					return MakeBuiltinError(ps, "No result", "Match?")
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//Match?")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//Match?")
			}
		},
	},

	// Tests:
	//  equal { regexp "[0-9]+" |Replace-all "x123y x234y" "XXX" } "xXXXy xXXXy"
	// Args:
	// * regexp: Regular expression pattern
	// * input: String to modify
	// * replacement: String to replace matches with
	// Returns:
	// * string with all matches replaced by the replacement string
	"regexp//Replace-all": {
		Argsn: 3,
		Doc:   "Replaces all occurrences of the regular expression pattern with the specified replacement string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch re := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					switch replac := arg2.(type) {
					case env.String:
						res := re.Value.(*regexp.Regexp).ReplaceAllString(val.Value, replac.Value)
						return *env.NewString(res)
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "regexp//Replace-all")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//Replace-all")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//Replace-all")
			}
		},
	},
}
