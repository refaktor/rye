package evaldo

import (
	"regexp"

	"github.com/refaktor/rye/env"
)

var Builtins_regexp = map[string]*env.Builtin{

	"regexp": {
		Argsn: 1,
		Doc:   "TODODOC",
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

	"regexp//matches": {
		Argsn: 2,
		Doc:   "TODODOC",
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
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "regexp//matches")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "regexp//matches")
			}
		},
	},

	"regexp//submatch?": {
		Argsn: 2,
		Doc:   "TODODOC",
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

	"regexp//submatches?": {
		Argsn: 2,
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

	"regexp//find-all": {
		Argsn: 2,
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

	"regexp//match?": {
		Argsn: 2,
		Doc:   "TODODOC",
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

	"regexp//replace-all": {
		Argsn: 3,
		Doc:   "TODODOC",
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
