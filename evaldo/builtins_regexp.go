package evaldo

import (
	"regexp"
	"rye/env"
)

var Builtins_regexp = map[string]*env.Builtin{

	"regexp": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				val, err := regexp.Compile(s.Value)
				if err != nil {
					return MakeError(env1, err.Error())
				}
				return *env.NewNative(env1.Idx, val, "regexp")
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"regexp//matches": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"regexp//submatch?": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindStringSubmatch(val.Value)
					if len(res) > 1 {
						return *env.NewString(res[1])
					} else {
						return MakeError(env1, "No submatch")
					}
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
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				switch s := arg0.(type) {
				case env.Native:
					res := s.Value.(*regexp.Regexp).FindString(val.Value)
					return *env.NewString(res)
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"regexp//replace-all": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch re := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					switch replac := arg2.(type) {
					case env.String:
						res := re.Value.(*regexp.Regexp).ReplaceAllString(val.Value, replac.Value)
						return *env.NewString(res)
					default:
						return MakeError(env1, "Arg2 not Native")
					}
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},
}
