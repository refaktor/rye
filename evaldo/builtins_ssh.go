//go:build add_ssh
// +build add_ssh

package evaldo

import (
	"io"

	"github.com/gliderlabs/ssh"
	"github.com/jinzhu/copier"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

var Builtins_ssh = map[string]*env.Builtin{

	"ssh-server": {
		Argsn: 1,
		Doc:   "Create new ssh server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				return *env.NewNative(ps.Idx, &ssh.Server{Addr: addr.Value}, "ssh-server")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ssh-server")
			}

		},
	},
	"ssh-server//handle": {
		Argsn: 2,
		Doc:   "HTTP handle function for server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.Function:
					server.Value.(*ssh.Server).Handle(func(s ssh.Session) {
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						copier.Copy(&psTemp, &ps)
						CallFunction(handler, ps, *env.NewNative(ps.Idx, s, "ssh-session"), false, nil)
					})
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "ssh-server//handle")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//handle")
			}
		},
	},
	"ssh-server//password-auth": {
		Argsn: 2,
		Doc:   "HTTP handler for password authentication.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.Function:
					pwda := ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						copier.Copy(&psTemp, &ps)
						newPs := CallFunction(handler, ps, *env.NewString(pass), false, nil)
						return util.IsTruthy(newPs.Res)
					})
					server.Value.(*ssh.Server).SetOption(pwda)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "ssh-server//password-auth")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//password-auth")
			}
		},
	},
	"ssh-server//serve": {
		Argsn: 1,
		Doc:   "Listen and serve new server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				server.Value.(*ssh.Server).ListenAndServe()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//serve")
			}
		},
	},

	"ssh-session//write": {
		Argsn: 2,
		Doc:   "SSH session write function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					io.WriteString(session.Value.(ssh.Session), val.Value)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-session//write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//write")
			}
		},
	},
}
