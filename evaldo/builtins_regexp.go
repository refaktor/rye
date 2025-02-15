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
	// * pattern: regular expression
	"regexp": {
		Argsn: 1,
		Doc:   "Creates a Regular Expression native value.",
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
	//  equal { regexp "[0-9]" |is-match "5" } 1
	//  equal { regexp "[0-9]" |is-match "a" } 0
	// Args:
	// * regexp - native regexp value
	// * input - value to test for matching
	"regexp//is-match": {
		Argsn: 2,
		Doc:   "Check if string matches the given regular epression.",
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
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//is-match")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//is-match")
			}
		},
	},

	// Tests:
	//  equal { regexp "x([0-9]+)y" |submatch? "x123y" } "123"
	"regexp//submatch?": {
		Argsn: 2,
		Doc:   "Get the first submatch from string given the regular exprepesion.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindStringSubmatch(val.Value)
					if len(res) > 1 {
						return *env.NewString(res[1])
					} else {
						return MakeBuiltinError(ps, "No submatch.", "regexp//submatch?")
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//submatch?")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//submatch?")
			}
		},
	},

	// Tests:
	//  equal { regexp "x([0-9]+)y" |submatches? "x123y x234y" } { "123" }
	"regexp//submatches?": {
		Argsn: 2,
		Doc:   "Get all regexp submatches in a Block.",
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
					return MakeBuiltinError(env1, "No results", "submatches?")
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	// Tests:
	//  equal { regexp "x([0-9]+)(y+)?" |submatches\all? "x11yy x22" } { { "11" "yy" } { "22" "" } }
	"regexp//submatches\\all?": {
		Argsn: 2,
		Doc:   "Get all regexp submatches in a Block.",
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
					return MakeBuiltinError(env1, "No results", "submatches?")
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	// Tests:
	//  equal { regexp "[0-9]+" |find-all "x123y x234y" } { "123" "234" }
	"regexp//find-all": {
		Argsn: 2,
		Doc:   "Find all matches and return them in a Block",
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
					return MakeBuiltinError(env1, "No results", "submatches?")
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	// Tests:
	//	equal { regexp "[0-9]+c+" |match? "aa33bb55cc" } "55cc"
	// Args:
	// * regexp value
	// * input
	"regexp//match?": {
		Argsn: 2,
		Doc:   "Get the regexp match.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindString(val.Value)
					return *env.NewString(res)
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//match?")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//match?")
			}
		},
	},

	// Tests:
	//  equal { regexp "[0-9]+" |replace-all "x123y x234y" "XXX" } "xXXXy xXXXy"
	"regexp//replace-all": {
		Argsn: 3,
		Doc:   "Replace all mathes in a string given the regexp with another string.",
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
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "regexp//replace-all")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//replace-all")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//replace-all")
			}
		},
	},
}
