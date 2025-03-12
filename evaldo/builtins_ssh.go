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

	//
	// ##### SSH ##### "SSH server functions"
	//
	// Tests:
	// equal { ssh-server "localhost:2222" |type? } 'native
	// Args:
	// * address: String containing host:port address to listen on
	// Returns:
	// * native SSH server object
	"ssh-server": {
		Argsn: 1,
		Doc:   "Creates a new SSH server that listens on the specified address.",
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

	// Tests:
	// equal { server: ssh-server "localhost:2222" , server |ssh-server//handle :{ |session| session |ssh-session//write "Hello" } |type? } 'native
	// Args:
	// * server: SSH server object
	// * handler: Function that receives an SSH session object
	// Returns:
	// * the SSH server object
	"ssh-server//handle": {
		Argsn: 2,
		Doc:   "Sets a handler function for SSH sessions on the server.",
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

	// Tests:
	// equal { server: ssh-server "localhost:2222" , server |ssh-server//password-auth :{ |pass| pass = "secret" } |type? } 'native
	// Args:
	// * server: SSH server object
	// * handler: Function that receives a password string and returns true/false
	// Returns:
	// * the SSH server object
	"ssh-server//password-auth": {
		Argsn: 2,
		Doc:   "Sets a password authentication handler for the SSH server.",
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

	// Tests:
	// equal { server: ssh-server "localhost:2222" , server |ssh-server//serve |type? } 'native
	// Args:
	// * server: SSH server object
	// Returns:
	// * the SSH server object
	"ssh-server//serve": {
		Argsn: 1,
		Doc:   "Starts the SSH server, listening for connections.",
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

	// Tests:
	// equal { session: ssh-session-mock , session |ssh-session//write "Hello" |type? } 'native
	// Args:
	// * session: SSH session object
	// * text: String to write to the session
	// Returns:
	// * the SSH session object
	"ssh-session//write": {
		Argsn: 2,
		Doc:   "Writes a string to an SSH session.",
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
