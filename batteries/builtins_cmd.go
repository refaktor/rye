package batteries

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

type command struct {
	cmd *exec.Cmd
	// pipe holds any commands that pipe into cmd.
	pipe []*exec.Cmd
	// files holds files that must be closed after the command finishes.
	files   []*os.File
	started bool
	closed  bool
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
	var errs []error
	for _, cmd := range c.pipe {
		err := cmd.Wait()
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, f := range c.files {
		err := f.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func listToCmdArgs(ps *env.ProgramState, input ...any) ([]string, *env.Error) {
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
			newArgs, err := listToCmdArgs(ps, it.Data...)
			if err != nil {
				return nil, err
			}
			args = append(args, newArgs...)
		case env.Flagword:
			if it.HasLong() && it.HasShort() {
				return nil, evaldo.MakeBuiltinError(ps, fmt.Sprintf("ambiguous flagword %s", it.Print(*ps.Idx)), "cmd")
			}
			args = append(args, it.Print(*ps.Idx))
		case env.Uri:
			if it.GetProtocol().Print(*ps.Idx) == "file" {
				args = append(args, it.GetPath())
			} else {
				args = append(args, it.GetFullUri(*ps.Idx))
			}
		default:
			return nil, evaldo.MakeBuiltinError(ps, fmt.Sprintf("Command argument must be string, integer, flagword, URI, or list, got %T", c), "cmd")
		}
	}
	return args, nil
}

// pipeCommands creates a new command c1 | c2.
func pipeCommands(c1, c2 *command) *command {
	c1.cmd.Stdout = nil
	p := command{cmd: c2.cmd}
	p.pipe = append(p.pipe, c1.pipe...)
	p.pipe = append(p.pipe, c1.cmd)
	p.pipe = append(p.pipe, c2.pipe...)
	p.files = append(p.files, c1.files...)
	p.files = append(p.files, c2.files...)
	return &p
}

// commandOutputFd resolves a Rye argument to a command output file for stdout or stderr.
func commandOutputFd(ps *env.ProgramState, c *command, arg1 env.Object) (io.Writer, env.Object) {
	switch s := arg1.(type) {
	case env.Uri:
		file, err := os.Create(s.Path)
		if err != nil {
			return nil, evaldo.MakeBuiltinError(ps, err.Error(), "Stdout!")
		}
		c.files = append(c.files, file)
		return file, nil
	case env.Native:
		switch it := s.Value.(type) {
		case *os.File:
			c.files = append(c.files, it)
			return it, nil
		case io.Writer:
			return it, nil
		}
		return nil, evaldo.MakeBuiltinError(ps, "arg1 must be of kind file or writer", "Stdout!")
	default:
		return nil, evaldo.MakeArgError(ps, 2, []env.Type{env.UriType, env.NativeType}, "Stdout!")
	}
}

// commandInputFd resolves a Rye argument to a command input file for stdin.
func commandInputFd(ps *env.ProgramState, c *command, arg1 env.Object) (io.Reader, env.Object) {
	switch s := arg1.(type) {
	case env.Uri:
		file, err := os.Open(s.Path)
		if err != nil {
			return nil, evaldo.MakeBuiltinError(ps, err.Error(), "Stdin!")
		}
		c.files = append(c.files, file)
		return file, nil
	case env.Native:
		switch it := s.Value.(type) {
		case *os.File:
			c.files = append(c.files, it)
			return it, nil
		case io.Reader:
			return it, nil
		}
		return nil, evaldo.MakeBuiltinError(ps, "arg1 must be of kind file or reader", "Stdin!")
	default:
		return nil, evaldo.MakeArgError(ps, 2, []env.Type{env.UriType, env.NativeType}, "Stdin!")
	}
}

func commandFn(name string, fn func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object) func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	return func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
		switch r := arg0.(type) {
		case env.Native:
			c := r.Value.(*command)
			return fn(ps, c, arg0, arg1, arg2, arg3, arg4)
		default:
			return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, name)
		}
	}
}

var Builtins_cmd = map[string]*env.Builtin{
	//
	// ##### Command Operations #####  "Running other programs"
	//
	// Tests:
	// equal { cmd { echo -n Hello World } |Output } "Hello World"
	// equal { cmd { echo -n Hello World |tr A-Z a-z |sed "s/hello/goodbye/" } |Output } "goodbye world"
	// equal { cmd { echo -n "1 + 1 =" { 1 + 1 } } |Output } "1 + 1 = 2"
	// equal { args: list { "two" "arguments" } cmd { printf "'%s' " ?args } |Output } "'two' 'arguments' "
	// Args:
	// * command: block or list containing a command arguments
	// Returns:
	// * native command object
	"cmd": {
		Argsn: 1,
		Doc:   "Create a command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var args []string
			var pipe []*exec.Cmd
			switch input := arg0.(type) {
			case env.List:
				newArgs, err := listToCmdArgs(ps, input.Data...)
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
						if len(args) == 0 {
							return evaldo.MakeBuiltinError(ps, "missing command before pipe", "cmd")
						}
						pipecmd := exec.Command(args[0], args[1:]...)
						pipecmd.Stdin = os.Stdin
						pipecmd.Stderr = os.Stderr
						pipe = append(pipe, pipecmd)
						args = []string{word}
					case env.Getword:
						evaldo.EvalGetword(ps, it, nil, false)
						if ps.ErrorFlag {
							return ps.Res
						}
						newArgs, err := listToCmdArgs(ps, ps.Res)
						if err != nil {
							return err
						}
						args = append(args, newArgs...)
					case env.Block:
						ser := ps.Ser
						ps.Ser = it.Series
						ps.BlockFile = it.FileName
						ps.BlockLine = it.Line
						evaldo.Eval(ps)
						evaldo.MaybeDisplayFailureOrError(ps, ps.Idx, "cmd")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						ps.Ser = ser
						newArgs, err := listToCmdArgs(ps, ps.Res)
						if err != nil {
							// TODO: Better message
							return err
						}
						args = append(args, newArgs...)
					default:
						newArgs, err := listToCmdArgs(ps, it)
						if err != nil {
							// TODO: Better message
							return err
						}
						args = append(args, newArgs...)
					}
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "cmd")
			}

			if len(args) == 0 {
				return evaldo.MakeBuiltinError(ps, "missing command", "cmd")
			}
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return *env.NewNative(ps.Idx, &command{cmd: cmd, pipe: pipe}, "command")
		},
	},
	// Tests:
	// equal { cmd { pwd } |Dir! %/ |Output |trim } "/"
	// Args:
	// * command: native command object
	// * dir: path to the working directory
	// Returns:
	// * the original command object
	"command//Dir!": {
		Argsn: 2,
		Doc:   "Change the working directory of a command.",
		Fn: commandFn("Dir!", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg1.(type) {
			case env.Uri:
				c.cmd.Dir = s.Path
				return arg0
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.UriType}, "Dir!")
			}
		}),
	},
	// Args:
	// * command: native command object
	// * input: path to an input file, a native file object, or a reader
	// Returns:
	// * the original command object
	"command//Stdin!": {
		Argsn: 2,
		Doc:   "Change the standard input of a command.",
		Fn: commandFn("Stdin!", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			out, ret := commandInputFd(ps, c, arg1)
			if ret != nil {
				return ret
			}
			c.cmd.Stdin = out
			return arg0
		}),
	},
	// Args:
	// * command: native command object
	// * output: path to an output file, a native file object, or a writer
	// Returns:
	// * the original command object
	"command//Stdout!": {
		Argsn: 2,
		Doc:   "Change the standard output of a command.",
		Fn: commandFn("Stdout!", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			out, ret := commandOutputFd(ps, c, arg1)
			if ret != nil {
				return ret
			}
			c.cmd.Stdout = out
			return arg0
		}),
	},
	// Args:
	// * command: native command object
	// * output: path to an output file, a native file object, or a writer
	// Returns:
	// * the original command object
	"command//Stderr!": {
		Argsn: 2,
		Doc:   "Change the standard error of a command.",
		Fn: commandFn("Stderr!", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			out, ret := commandOutputFd(ps, c, arg1)
			if ret != nil {
				return ret
			}
			c.cmd.Stderr = out
			return arg0
		}),
	},
	// Tests:
	// equal { cmd { echo -n Hello World } |Pipe cmd { tr a-z A-Z } |Output } "HELLO WORLD"
	// Args:
	// * c1: first command writing to the pipe
	// * c2: second command reading from the pipe
	// Returns:
	// * new native command object
	"command//Pipe": {
		Argsn: 2,
		Doc:   "Pipe the output of the first command to the input of the second command.",
		Fn: commandFn("Pipe", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch it := arg1.(type) {
			case env.Native:
				if c2, ok := it.Value.(*command); ok {
					return *env.NewNative(ps.Idx, pipeCommands(c, c2), "command")
				} else {
					return evaldo.MakeBuiltinError(ps, "arg1 must be of kind command", "Pipe!")
				}
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "Pipe")
			}
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * boolean true if the command executed successfully
	// Tests:
	// error { cmd { false } |Run }
	"command//Run": {
		Argsn: 1,
		Doc:   "Start a command and wait for it to finish.",
		Fn: commandFn("Run", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			defer c.Close()
			err := c.StartPipe()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Run")
			}
			err = c.cmd.Run()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Run")
			}
			return *env.NewBoolean(true)
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * the same command after starting it (does not wait)
	"command//Start": {
		Argsn: 1,
		Doc:   "Start a command (and any piped commands) without waiting.",
		Fn: commandFn("Start", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if c.started {
				return arg0
			}
			if err := c.StartPipe(); err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Start")
			}
			if err := c.cmd.Start(); err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Start")
			}
			c.started = true
			return arg0
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * boolean true; kills the process (and any started piped processes)
	"command//Kill": {
		Argsn: 1,
		Doc:   "Kill a started command (and any started piped commands).",
		Fn: commandFn("Kill", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var firstErr error
			if c.cmd != nil && c.cmd.Process != nil {
				if err := c.cmd.Process.Kill(); err != nil && firstErr == nil {
					firstErr = err
				}
			}
			for _, pc := range c.pipe {
				if pc != nil && pc.Process != nil {
					if err := pc.Process.Kill(); err != nil && firstErr == nil {
						firstErr = err
					}
				}
			}
			_ = c.Close()
			c.closed = true
			if firstErr != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, firstErr.Error(), "Kill")
			}
			return *env.NewBoolean(true)
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * process id of the main command or -1 if not started
	"command//Pid": {
		Argsn: 1,
		Doc:   "Return the process id of the main command (or -1 if not started).",
		Fn: commandFn("Pid", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if c.cmd == nil || c.cmd.Process == nil {
				return *env.NewInteger(-1)
			}
			return *env.NewInteger(int64(c.cmd.Process.Pid))
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * exit status (or list for pipelines) after waiting
	"command//Wait": {
		Argsn: 1,
		Doc:   "Wait for a started command to finish and return its exit status.",
		Fn: commandFn("Wait", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if !c.started {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "command not started", "Wait")
			}
			err := c.cmd.Wait()
			_ = c.Close()
			c.closed = true
			status := c.cmd.ProcessState.ExitCode()
			if err != nil && status == -1 {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Wait")
			}
			if len(c.pipe) > 0 {
				var statusList []any
				for _, pc := range c.pipe {
					statusList = append(statusList, *env.NewInteger(int64(pc.ProcessState.ExitCode())))
				}
				statusList = append(statusList, *env.NewInteger(int64(status)))
				return *env.NewList(statusList)
			}
			return *env.NewInteger(int64(status))
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * for simple command: exit status integer
	// * for pipeline: list of exit status integers
	// Tests:
	// equal { cmd { true } |Status } 0
	// equal { cmd { false } |Status } 1
	// equal { cmd { false |true } |Status } list [ 1 0 ]
	"command//Status": {
		Argsn: 1,
		Doc:   "Start a command, wait for it to finish, and return its exit status.",
		Fn: commandFn("Status", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			err := c.StartPipe()
			if err != nil {
				c.Close()
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Status")
			}
			err = c.cmd.Run()
			c.Close()
			status := c.cmd.ProcessState.ExitCode()
			if err != nil && status == -1 {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Status")
			}
			if len(c.pipe) > 0 {
				var statusList []any
				for _, cmd := range c.pipe {
					statusList = append(statusList, *env.NewInteger(int64(cmd.ProcessState.ExitCode())))
				}
				statusList = append(statusList, *env.NewInteger(int64(status)))
				return *env.NewList(statusList)
			}
			return *env.NewInteger(int64(status))
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * string containing data written to the command's standard output
	"command//Output": {
		Argsn: 1,
		Doc:   "Start a command, wait for it to finish, and return its standard output",
		Fn: commandFn("Output", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			defer c.Close()
			err := c.StartPipe()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Output")
			}
			c.cmd.Stdout = nil
			out, err := c.cmd.Output()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Output")
			}
			return *env.NewString(string(out))
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * string containing data written to the command's standard output and error
	"command//CombinedOutput": {
		Argsn: 1,
		Doc:   "Start a command, wait for it to finish, and return its combined standard output and error",
		Fn: commandFn("CombinedOutput", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			defer c.Close()
			err := c.StartPipe()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "CombinedOutput")
			}
			c.cmd.Stdout = nil
			c.cmd.Stderr = nil
			out, err := c.cmd.CombinedOutput()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "CombinedOutput")
			}
			return *env.NewString(string(out))
		}),
	},
	// Args:
	// * command: native command object
	// Returns:
	// * the original command object after redirecting stdin/stdout/stderr to /dev/null
	"command//Detach!": {
		Argsn: 1,
		Doc:   "Redirect stdin, stdout, and stderr of a command to /dev/null.",
		Fn: commandFn("Detach!", func(ps *env.ProgramState, c *command, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Open /dev/null once and reuse for all three.
			f, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "Detach!")
			}
			c.files = append(c.files, f)
			c.cmd.Stdin = f
			c.cmd.Stdout = f
			c.cmd.Stderr = f
			return arg0
		}),
	},
}
