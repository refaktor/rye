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
	// equal { server: ssh-server "localhost:2222" , server |ssh-server//Handle :{ |session| session |ssh-session//Write "Hello" } |type? } 'native
	// Args:
	// * server: SSH server object
	// * handler: Function that receives an SSH session object
	// Returns:
	// * the SSH server object
	"ssh-server//Handle": {
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
						CallFunctionWithArgs(handler, ps, nil, *env.NewNative(ps.Idx, s, "ssh-session"))
						// Check for errors after calling handler and print to server console
						if ps.FailureFlag || ps.ErrorFlag {
							println("Error in SSH handler: " + ps.Res.Inspect(*ps.Idx))
						}
					})
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "ssh-server//Handle")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Handle")
			}
		},
	},

	// Tests:
	// equal { server: ssh-server "localhost:2222" , server |ssh-server//Password-auth :{ |pass| pass = "secret" } |type? } 'native
	// Args:
	// * server: SSH server object
	// * handler: Function that receives a password string and returns true/false
	// Returns:
	// * the SSH server object
	"ssh-server//Password-auth": {
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
						CallFunctionWithArgs(handler, ps, nil, *env.NewString(pass))
						return util.IsTruthy(ps.Res)
					})
					server.Value.(*ssh.Server).SetOption(pwda)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "ssh-server//Password-auth")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Password-auth")
			}
		},
	},

	// Tests:
	// equal { server: ssh-server "localhost:2222" , server |ssh-server//Serve |type? } 'native
	// Args:
	// * server: SSH server object
	// Returns:
	// * the SSH server object, or error if unable to serve
	"ssh-server//Serve": {
		Argsn: 1,
		Doc:   "Starts the SSH server, listening for connections.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				err := server.Value.(*ssh.Server).ListenAndServe()
				if err != nil {
					return makeError(ps, err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Serve")
			}
		},
	},

	// Tests:
	// equal { session: ssh-session-mock , session |ssh-session//Write "Hello" |type? } 'native
	// Args:
	// * session: SSH session object
	// * text: String to write to the session
	// Returns:
	// * the SSH session object
	"ssh-session//Write": {
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
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-session//Write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Write")
			}
		},
	},
}
