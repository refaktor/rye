package evaldo

import (
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/refaktor/rye/env"
)

type command struct {
	cmd *exec.Cmd
	// pipe holds any commands that pipe into cmd.
	pipe []*exec.Cmd
	// files holds files that must be closed after the command finishes.
	files []*os.File
}

func (c *command) StartPipe() error {
	var err error
	var stdout io.ReadCloser
	for i, cmd := range c.pipe {
		if i > 0 {
			cmd.Stdin = stdout
		}
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return err
		}
		err = cmd.Start()
		if err != nil {
			return err
		}
	}
	if stdout != nil {
		c.cmd.Stdin = stdout
	}
	return nil
}

func (c *command) Close() error {
	// TODO: should probably not return on first error
	for _, cmd := range c.pipe {
		err := cmd.Wait()
		if err != nil {
			return err
		}
	}
	for _, f := range c.files {
		err := f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func listToShArgs(ps *env.ProgramState, input ...any) ([]string, *env.Error) {
	var args []string
	for _, c := range input {
		switch it := c.(type) {
		case string:
			args = append(args, it)
		case env.String:
			args = append(args, it.Value)
		case int:
			args = append(args, strconv.Itoa(it))
		case env.Integer:
			args = append(args, strconv.Itoa(int(it.Value)))
		case env.List:
			newArgs, err := listToShArgs(ps, it.Data...)
			if err != nil {
				return nil, err
			}
			args = append(args, newArgs...)
		default:
			return nil, MakeBuiltinError(ps, "List data must be integer or string", "sh")
		}
	}
	return args, nil
}

var Builtins_sh = map[string]*env.Builtin{
	//
	// ##### Shell #####  "Calling other programs"
	//
	// Tests:
	// equal { sh { echo -n Hello World } |Output } "Hello World"
	// Args:
	// * command: block or list containing a command arguments
	// Returns:
	// * Command
	"sh": {
		Argsn: 1,
		Doc:   "Creates a shell command",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var args []string
			var pipe []*exec.Cmd
			switch input := arg0.(type) {
			case env.List:
				newArgs, err := listToShArgs(ps, input.Data...)
				if err != nil {
					return err
				}
				args = append(args, newArgs...)
			case env.Block:
				for _, c := range input.Series.S {
					switch it := c.(type) {
					case env.Word:
						args = append(args, it.Print(*ps.Idx))
					case env.Pipeword:
						word := it.ToWord().Print(*ps.Idx)
						if word[0] == '-' {
							// not sure why - makes a pipeword
							args = append(args, word)
						} else {
							pipecmd := exec.Command(args[0], args[1:]...)
							pipecmd.Stdin = os.Stdin
							pipecmd.Stderr = os.Stderr
							pipe = append(pipe, pipecmd)
							args = []string{word}
						}
					case env.Block:
						ser := ps.Ser
						ps.Ser = it.Series
						ps.BlockFile = it.FileName
						ps.BlockLine = it.Line
						EvalBlock(ps)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser = ser
						newArgs, err := listToShArgs(ps, ps.Res)
						if err != nil {
							// TODO: Better message
							return err
						}
						args = append(args, newArgs...)
					default:
						newArgs, err := listToShArgs(ps, it)
						if err != nil {
							// TODO: Better message
							return err
						}
						args = append(args, newArgs...)
					}
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sh")
			}

			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return *env.NewNative(ps.Idx, &command{cmd: cmd, pipe: pipe}, "command")
		},
	},
	"command//Dir!": {
		Argsn: 2,
		Doc:   "Change the working directory of a command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				c := r.Value.(*command)
				switch s := arg1.(type) {
				case env.Uri:
					c.cmd.Dir = s.Path
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "Dir!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Dir!")
			}
		},
	},
	// TODO: Stdin/Stderr
	"command//Stdout!": {
		Argsn: 2,
		Doc:   "Change the standard output of a command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				c := r.Value.(*command)
				switch s := arg1.(type) {
				case env.Uri:
					file, err := os.Create(s.Path)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "Stdout!")
					}
					c.cmd.Stdout = file
					c.files = append(c.files, file)
					return arg0
				case env.Native:
					if file, ok := s.Value.(*os.File); ok {
						c.cmd.Stdout = file
						c.files = append(c.files, file)
						return arg0
					}
					return MakeBuiltinError(ps, "arg1 must be of kind file", "Stdout!")
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "Stdout!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Stdout!")
			}
		},
	},
	"command//Run": {
		Argsn: 1,
		Doc:   "Start a command and wait for it to finish.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				c := r.Value.(*command)
				defer c.Close()
				err := c.StartPipe()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "Run")
				}
				err = c.cmd.Run()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "Run")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Run")
			}
			return nil
		},
	},
	"command//Output": {
		Argsn: 1,
		Doc:   "Start a command, wait for it to finish, and return its standard output",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				c := r.Value.(*command)
				defer c.Close()
				err := c.StartPipe()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "Output")
				}
				c.cmd.Stdout = nil
				out, err := c.cmd.Output()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "Output")
				}
				return *env.NewString(string(out))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Output")
			}
		},
	},
	"command//CombinedOutput": {
		Argsn: 1,
		Doc:   "Start a command, wait for it to finish, and return its combined standard output and error",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				c := r.Value.(*command)
				defer c.Close()
				err := c.StartPipe()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "CombinedOutput")
				}
				c.cmd.Stdout = nil
				c.cmd.Stderr = nil
				out, err := c.cmd.CombinedOutput()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "CombinedOutput")
				}
				return *env.NewString(string(out))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "CombinedOutput")
			}
		},
	},
}
