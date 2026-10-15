//go:build add_ssh
// +build add_ssh

package batteries

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/jinzhu/copier"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/util"
	gossh "golang.org/x/crypto/ssh"
)

// ryeSSHClientSession wraps an x/crypto/ssh client session together with the
// stdin pipe and captured stdout/stderr buffers, so Rye can start a command,
// stream data into its stdin, close stdin to signal EOF, and read results.
type ryeSSHClientSession struct {
	sess   *gossh.Session
	stdin  io.WriteCloser
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

// sshAddr ensures the host string contains a port, defaulting to 22.
func sshAddr(host string) string {
	if _, _, err := net.SplitHostPort(host); err == nil {
		return host
	}
	return net.JoinHostPort(host, "22")
}

// parseSSHUserinfo splits the SSH URI path (user[:pass]@host[:port]) into
// user, password and host[:port] parts.
func parseSSHUserinfo(path string) (user, pass, hostport string) {
	hostport = path
	if at := strings.LastIndex(path, "@"); at >= 0 {
		userinfo := path[:at]
		hostport = path[at+1:]
		if i := strings.Index(userinfo, ":"); i >= 0 {
			user = userinfo[:i]
			pass = userinfo[i+1:]
		} else {
			user = userinfo
		}
	}
	return
}

// dialSSHClient dials an SSH server with the given user, auth methods and
// host[:port] address.
func dialSSHClient(user string, auth []gossh.AuthMethod, hostport string) (*gossh.Client, error) {
	config := &gossh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}
	return gossh.Dial("tcp", sshAddr(hostport), config)
}

// runSSHCommand runs a command on an existing SSH client, capturing stdout and
// stderr separately and returning the exit code even on non-zero exit status.
func runSSHCommand(client *gossh.Client, cmd string) (stdout, stderr string, exitCode int, err error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", "", 0, err
	}
	defer sess.Close()

	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf

	err = sess.Run(cmd)
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*gossh.ExitError); ok {
			exitCode = ee.ExitStatus()
		} else {
			return "", "", 0, err
		}
	}
	return outBuf.String(), errBuf.String(), exitCode, nil
}

var Builtins_ssh = map[string]*env.Builtin{

	//
	// ##### SSH ##### "SSH server functions"
	//
	// Example:
	//  server: ssh-server "localhost:2222"
	//  server |Password-auth { pass = "secret123" }
	//  server |Handle fn { session } {
	//    session .Write "Welcome to Rye SSH server!\n"
	//    session .Write "Type 'exit' to quit.\n"
	//  }
	//  server |Serve
	//
	// Tests:
	// ; equal { ssh-server "localhost:2222" |type? } 'native
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
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "ssh-server")
			}

		},
	},

	// Tests:
	// ; equal { server: ssh-server "localhost:2222" , server |Handle { session |Write "Hello" } |type? } 'native
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
						evaldo.CallFunctionWithArgs(handler, ps, nil, *env.NewNative(ps.Idx, s, "ssh-session"))
						// Check for errors after calling handler and print to server console
						if ps.FailureFlag || ps.ErrorFlag {
							println("Error in SSH handler: " + ps.Res.Inspect(*ps.Idx))
						}
					})
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.FunctionType}, "ssh-server//Handle")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Handle")
			}
		},
	},

	// Tests:
	// ; equal { server: ssh-server "localhost:2222" , server |Password-auth { pass = "secret" } |type? } 'native
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
						evaldo.CallFunctionWithArgs(handler, ps, nil, *env.NewString(pass))
						return util.IsTruthy(ps.Res)
					})
					server.Value.(*ssh.Server).SetOption(pwda)
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.FunctionType}, "ssh-server//Password-auth")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Password-auth")
			}
		},
	},

	// Args:
	// * server: SSH server object
	// * handler: Function that receives a public key string and returns true/false
	// Returns:
	// * the SSH server object
	"ssh-server//Public-key-auth": {
		Argsn: 2,
		Doc:   "Sets a public key authentication handler for the SSH server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.Function:
					pka := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						copier.Copy(&psTemp, &ps)
						keyStr := strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key)))
						evaldo.CallFunctionWithArgs(handler, ps, nil, *env.NewString(keyStr))
						return util.IsTruthy(ps.Res)
					})
					server.Value.(*ssh.Server).SetOption(pka)
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.FunctionType}, "ssh-server//Public-key-auth")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Public-key-auth")
			}
		},
	},

	// Args:
	// * server: SSH server object
	// * path: String path to a PEM private key file used as host key
	// Returns:
	// * the SSH server object
	"ssh-server//Host-key-file": {
		Argsn: 2,
		Doc:   "Loads a private key file and uses it as the SSH server host key.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch path := arg1.(type) {
				case env.String:
					if err := server.Value.(*ssh.Server).SetOption(ssh.HostKeyFile(path.Value)); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-server//Host-key-file")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-server//Host-key-file")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Host-key-file")
			}
		},
	},

	// Args:
	// * server: SSH server object
	// * banner: String banner shown to clients before authentication
	// Returns:
	// * the SSH server object
	"ssh-server//Banner": {
		Argsn: 2,
		Doc:   "Sets the banner string sent to clients before authentication.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch banner := arg1.(type) {
				case env.String:
					server.Value.(*ssh.Server).Banner = banner.Value
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-server//Banner")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Banner")
			}
		},
	},

	// Tests:
	// ; equal { server: ssh-server "localhost:2222" , server |Serve |type? } 'native
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
					return evaldo.MakeError(ps, err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Serve")
			}
		},
	},

	// Args:
	// * server: SSH server object
	// Returns:
	// * the SSH server object
	"ssh-server//Close": {
		Argsn: 1,
		Doc:   "Closes all active listeners and connections on the SSH server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				if err := server.Value.(*ssh.Server).Close(); err != nil {
					return evaldo.MakeError(ps, err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Close")
			}
		},
	},

	// Args:
	// * server: SSH server object
	// Returns:
	// * the SSH server object
	"ssh-server//Shutdown": {
		Argsn: 1,
		Doc:   "Gracefully shuts down the SSH server, closing listeners without interrupting active connections.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				if err := server.Value.(*ssh.Server).Shutdown(context.Background()); err != nil {
					return evaldo.MakeError(ps, err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-server//Shutdown")
			}
		},
	},

	// Tests:
	// ; equal { session: ssh-session-mock , session |Write "Hello" |type? } 'native
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
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-session//Write")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Write")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// * length: maximum number of bytes to read
	// Returns:
	// * String read from the session, or error on failure
	"ssh-session//Read": {
		Argsn: 2,
		Doc:   "Reads up to a given number of bytes from an SSH session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				switch length := arg1.(type) {
				case env.Integer:
					buf := make([]byte, length.Value)
					n, err := session.Value.(ssh.Session).Read(buf)
					if err != nil && err != io.EOF {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-session//Read")
					}
					return *env.NewString(string(buf[:n]))
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ssh-session//Read")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Read")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// * code: integer exit status
	// Returns:
	// * the SSH session object
	"ssh-session//Exit": {
		Argsn: 2,
		Doc:   "Sends an exit status and closes the SSH session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					if err := session.Value.(ssh.Session).Exit(int(code.Value)); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-session//Exit")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ssh-session//Exit")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Exit")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * String username used for the connection
	"ssh-session//User": {
		Argsn: 1,
		Doc:   "Returns the username used when establishing the SSH connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				return *env.NewString(session.Value.(ssh.Session).User())
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//User")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * String remote address of the client
	"ssh-session//Remote-addr": {
		Argsn: 1,
		Doc:   "Returns the remote address of the SSH connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				return *env.NewString(session.Value.(ssh.Session).RemoteAddr().String())
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Remote-addr")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * String local address of the server
	"ssh-session//Local-addr": {
		Argsn: 1,
		Doc:   "Returns the local address of the SSH connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				return *env.NewString(session.Value.(ssh.Session).LocalAddr().String())
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Local-addr")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * Block of shell-parsed command arguments
	"ssh-session//Command": {
		Argsn: 1,
		Doc:   "Returns the shell-parsed command arguments provided by the user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				cmd := session.Value.(ssh.Session).Command()
				objs := make([]env.Object, len(cmd))
				for i, c := range cmd {
					objs[i] = *env.NewString(c)
				}
				return *env.NewBlock(*env.NewTSeries(objs))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Command")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * String raw command provided by the user
	"ssh-session//Raw-command": {
		Argsn: 1,
		Doc:   "Returns the exact raw command string provided by the user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				return *env.NewString(session.Value.(ssh.Session).RawCommand())
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Raw-command")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * Block of "key=value" environment strings
	"ssh-session//Environ": {
		Argsn: 1,
		Doc:   "Returns the environment variables set for the session as a block of key=value strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				envs := session.Value.(ssh.Session).Environ()
				objs := make([]env.Object, len(envs))
				for i, e := range envs {
					objs[i] = *env.NewString(e)
				}
				return *env.NewBlock(*env.NewTSeries(objs))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Environ")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * String subsystem requested by the user
	"ssh-session//Subsystem": {
		Argsn: 1,
		Doc:   "Returns the subsystem requested by the user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				return *env.NewString(session.Value.(ssh.Session).Subsystem())
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Subsystem")
			}
		},
	},

	// Args:
	// * session: SSH session object
	// Returns:
	// * String authorized-keys formatted public key, or error if none used
	"ssh-session//Public-key": {
		Argsn: 1,
		Doc:   "Returns the public key used to authenticate the session, in authorized_keys format.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				pk := session.Value.(ssh.Session).PublicKey()
				if pk == nil {
					return evaldo.MakeBuiltinError(ps, "No public key used for authentication", "ssh-session//Public-key")
				}
				return *env.NewString(strings.TrimSpace(string(gossh.MarshalAuthorizedKey(pk))))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-session//Public-key")
			}
		},
	},

	//
	// ##### SSH client #####
	//
	// Example:
	//  cli: Open ssh://me:secret@example.com:22
	//  run cli "uname -a"
	// Args:
	// * uri: SSH URI in the form ssh://[user[:pass]@]host[:port]
	// Returns:
	// * native SSH client object
	"ssh-uri//Open": {
		Argsn: 1,
		Doc:   "Opens an SSH client connection. URI format: ssh://user[:pass]@host[:port].",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				user, pass, hostport := parseSSHUserinfo(uri.GetPath())
				if user == "" {
					return evaldo.MakeBuiltinError(ps, "SSH URI must include a user: ssh://user[:pass]@host[:port]", "ssh-uri//Open")
				}

				var auth []gossh.AuthMethod
				if pass != "" {
					auth = append(auth, gossh.Password(pass))
				}

				client, err := dialSSHClient(user, auth, hostport)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "Failed to connect: "+err.Error(), "ssh-uri//Open")
				}
				return *env.NewNative(ps.Idx, client, "ssh-client")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "ssh-uri//Open")
			}
		},
	},

	// Example:
	//  cli: Open\pwd ssh://me@example.com:22 "secret"
	//  run cli "uname -a"
	// Args:
	// * uri: SSH URI in the form ssh://user@host[:port]
	// * password: String password for authentication
	// Returns:
	// * native SSH client object
	"ssh-uri//Open\\pwd": {
		Argsn: 2,
		Doc:   "Opens an SSH client connection with a separate password parameter. URI format: ssh://user@host[:port].",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				user, _, hostport := parseSSHUserinfo(uri.GetPath())
				if user == "" {
					return evaldo.MakeBuiltinError(ps, "SSH URI must include a user: ssh://user@host[:port]", "ssh-uri//Open\\pwd")
				}

				switch pwd := arg1.(type) {
				case env.String:
					client, err := dialSSHClient(user, []gossh.AuthMethod{gossh.Password(pwd.Value)}, hostport)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, "Failed to connect: "+err.Error(), "ssh-uri//Open\\pwd")
					}
					return *env.NewNative(ps.Idx, client, "ssh-client")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-uri//Open\\pwd")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "ssh-uri//Open\\pwd")
			}
		},
	},

	// Example:
	//  cli: Open\key ssh://me@example.com:22 %~/.ssh/id_rsa
	//  run cli "uname -a"
	// Args:
	// * uri: SSH URI in the form ssh://user@host[:port]
	// * key: String path to a private key file
	// Returns:
	// * native SSH client object
	"ssh-uri//Open\\key": {
		Argsn: 2,
		Doc:   "Opens an SSH client connection using a private key file. URI format: ssh://user@host[:port].",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				user, _, hostport := parseSSHUserinfo(uri.GetPath())
				if user == "" {
					return evaldo.MakeBuiltinError(ps, "SSH URI must include a user: ssh://user@host[:port]", "ssh-uri//Open\\key")
				}

				var keyPath string
				switch keyArg := arg1.(type) {
				case env.String:
					keyPath = keyArg.Value
				case env.Uri:
					keyPath = keyArg.GetPath()
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType, env.UriType}, "ssh-uri//Open\\key")
				}

				keyBytes, err := os.ReadFile(keyPath)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "Unable to read private key file: "+err.Error(), "ssh-uri//Open\\key")
				}
				signer, err := gossh.ParsePrivateKey(keyBytes)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "Unable to parse private key: "+err.Error(), "ssh-uri//Open\\key")
				}

				client, err := dialSSHClient(user, []gossh.AuthMethod{gossh.PublicKeys(signer)}, hostport)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "Failed to connect: "+err.Error(), "ssh-uri//Open\\key")
				}
				return *env.NewNative(ps.Idx, client, "ssh-client")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "ssh-uri//Open\\key")
			}
		},
	},

	// Args:
	// * client: SSH client object
	// * command: String command to run remotely
	// Returns:
	// * Dict with stdout, stderr and exit keys
	"run": {
		Argsn: 2,
		Doc:   "Runs a command on a remote SSH client and returns a dict with 'stdout', 'stderr' and 'exit'.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				cli, ok := client.Value.(*gossh.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client", "run")
				}
				switch cmd := arg1.(type) {
				case env.String:
					stdout, stderr, exitCode, err := runSSHCommand(cli, cmd.Value)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "run")
					}
					return *env.NewDict(map[string]any{"stdout": stdout, "stderr": stderr, "exit": int64(exitCode)})
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "run")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "run")
			}
		},
	},

	// Args:
	// * client: SSH client object
	// Returns:
	// * the SSH client object
	"ssh-client//Close": {
		Argsn: 1,
		Doc:   "Closes the SSH client connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				cli, ok := client.Value.(*gossh.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client", "ssh-client//Close")
				}
				if err := cli.Close(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client//Close")
				}
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client//Close")
			}
		},
	},

	// Args:
	// * client: SSH client object
	// Returns:
	// * native SSH client session object
	"ssh-client//New-session": {
		Argsn: 1,
		Doc:   "Creates a new SSH session on the client, with stdin/stdout/stderr prepared for use.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				cli, ok := client.Value.(*gossh.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client", "ssh-client//New-session")
				}
				sess, err := cli.NewSession()
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client//New-session")
				}
				stdin, err := sess.StdinPipe()
				if err != nil {
					sess.Close()
					return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client//New-session")
				}
				rs := &ryeSSHClientSession{
					sess:   sess,
					stdin:  stdin,
					stdout: &bytes.Buffer{},
					stderr: &bytes.Buffer{},
				}
				sess.Stdout = rs.stdout
				sess.Stderr = rs.stderr
				return *env.NewNative(ps.Idx, rs, "ssh-client-session")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client//New-session")
			}
		},
	},

	// Args:
	// * client: SSH client object
	// * command: String command to run remotely
	// Returns:
	// * Dict with stdout, stderr and exit keys
	"ssh-client//Run": {
		Argsn: 2,
		Doc:   "Runs a command on a remote SSH client and returns a dict with 'stdout', 'stderr' and 'exit'.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				cli, ok := client.Value.(*gossh.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client", "ssh-client//Run")
				}
				switch cmd := arg1.(type) {
				case env.String:
					stdout, stderr, exitCode, err := runSSHCommand(cli, cmd.Value)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client//Run")
					}
					return *env.NewDict(map[string]any{"stdout": stdout, "stderr": stderr, "exit": int64(exitCode)})
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-client//Run")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client//Run")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// * command: String command to start remotely
	// Returns:
	// * the SSH client session object
	"ssh-client-session//Start": {
		Argsn: 2,
		Doc:   "Starts a command on the SSH session (non-blocking).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Start")
				}
				switch cmd := arg1.(type) {
				case env.String:
					if err := rs.sess.Start(cmd.Value); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Start")
					}
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-client-session//Start")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Start")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// * data: String or bytes to write to standard input
	// Returns:
	// * the SSH client session object
	"ssh-client-session//Stdin": {
		Argsn: 2,
		Doc:   "Writes a string or bytes to the session's standard input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Stdin")
				}
				switch data := arg1.(type) {
				case env.String:
					if _, err := io.WriteString(rs.stdin, data.Value); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Stdin")
					}
					return arg0
				case env.Native:
					if b, ok := data.Value.([]byte); ok {
						if _, err := rs.stdin.Write(b); err != nil {
							return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Stdin")
						}
						return arg0
					}
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-client-session//Stdin")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType, env.NativeType}, "ssh-client-session//Stdin")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Stdin")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// Returns:
	// * the SSH client session object
	"ssh-client-session//Close-stdin": {
		Argsn: 1,
		Doc:   "Closes the session's standard input, signaling end-of-file to the remote program.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Close-stdin")
				}
				if err := rs.stdin.Close(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Close-stdin")
				}
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Close-stdin")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// Returns:
	// * Integer exit code of the command
	"ssh-client-session//Wait": {
		Argsn: 1,
		Doc:   "Waits for the command to finish and returns its exit code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Wait")
				}
				err := rs.sess.Wait()
				exitCode := 0
				if err != nil {
					if ee, ok := err.(*gossh.ExitError); ok {
						exitCode = ee.ExitStatus()
					} else {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Wait")
					}
				}
				return *env.NewInteger(int64(exitCode))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Wait")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// * command: String command to run remotely
	// Returns:
	// * Dict with stdout, stderr and exit keys
	"ssh-client-session//Run": {
		Argsn: 2,
		Doc:   "Runs a command on the session and waits, returning a dict with 'stdout', 'stderr' and 'exit'.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Run")
				}
				switch cmd := arg1.(type) {
				case env.String:
					err := rs.sess.Run(cmd.Value)
					exitCode := 0
					if err != nil {
						if ee, ok := err.(*gossh.ExitError); ok {
							exitCode = ee.ExitStatus()
						} else {
							return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Run")
						}
					}
					return *env.NewDict(map[string]any{
						"stdout": rs.stdout.String(),
						"stderr": rs.stderr.String(),
						"exit":   int64(exitCode),
					})
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-client-session//Run")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Run")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// Returns:
	// * String captured standard output
	"ssh-client-session//Stdout": {
		Argsn: 1,
		Doc:   "Returns the captured standard output of the session as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Stdout")
				}
				return *env.NewString(rs.stdout.String())
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Stdout")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// Returns:
	// * String captured standard error
	"ssh-client-session//Stderr": {
		Argsn: 1,
		Doc:   "Returns the captured standard error of the session as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Stderr")
				}
				return *env.NewString(rs.stderr.String())
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Stderr")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// Returns:
	// * the SSH client session object
	"ssh-client-session//Shell": {
		Argsn: 1,
		Doc:   "Starts a login shell on the SSH session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Shell")
				}
				if err := rs.sess.Shell(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Shell")
				}
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Shell")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// * name: String environment variable name
	// * value: String environment variable value
	// Returns:
	// * the SSH client session object
	"ssh-client-session//Setenv": {
		Argsn: 3,
		Doc:   "Sets an environment variable on the SSH session before starting the command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Setenv")
				}
				switch name := arg1.(type) {
				case env.String:
					switch value := arg2.(type) {
					case env.String:
						if err := rs.sess.Setenv(name.Value, value.Value); err != nil {
							return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Setenv")
						}
						return arg0
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "ssh-client-session//Setenv")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ssh-client-session//Setenv")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Setenv")
			}
		},
	},

	// Args:
	// * session: SSH client session object
	// Returns:
	// * the SSH client session object
	"ssh-client-session//Close": {
		Argsn: 1,
		Doc:   "Closes the SSH client session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				rs, ok := session.Value.(*ryeSSHClientSession)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected SSH client session", "ssh-client-session//Close")
				}
				if err := rs.sess.Close(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "ssh-client-session//Close")
				}
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ssh-client-session//Close")
			}
		},
	},
}
