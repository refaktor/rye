//go:build !b_no_term
// +build !b_no_term

package evaldo

import (
	"os"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
	goterm "golang.org/x/term"
)

var Builtins_term = map[string]*env.Builtin{

	"width?": {
		Argsn: 0,
		Doc:   "Get the terminal width",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fd := int(os.Stdout.Fd())
			width, _, err := goterm.GetSize(fd)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "width?")
			}
			return env.NewInteger(int64(width))
		},
	},
	"red": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorRed()
			return env.NewInteger(1)
		},
	},
	"blue": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBlue()
			return env.NewInteger(1)
		},
	},
	"green": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorGreen()
			return env.NewInteger(1)
		},
	},
	"orange": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorOrange()
			return env.NewInteger(1)
		},
	},
	"magenta": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorMagenta()
			return env.NewInteger(1)
		},
	},
	"cyan": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBlue()
			return env.NewInteger(1)
		},
	},
	"bold": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBold()
			return env.NewInteger(1)
		},
	},
	"reset\\bold": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ResetBold()
			return env.NewInteger(1)

		},
	},
	"reset\\all": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.CloseProps()
			return env.NewInteger(1)
		},
	},
}
