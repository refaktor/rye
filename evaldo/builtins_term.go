//go:build !b_no_term
// +build !b_no_term

package evaldo

import (
	"os"

	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
	goterm "golang.org/x/term"
)

var Builtins_term = map[string]*env.Builtin{

	"wrap": {
		Argsn: 2,
		Doc:   "Wraps string to certain width",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "wrap")
			}
			wdt, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "wrap")
			}
			mkd := wrap.String(text.Value, wdth.Value)
			return *env.NewString(mkd)
		},
	},

	"wrap\\words": {
		Argsn: 2,
		Doc:   "Wraps string to certain width",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "wrap")
			}
			wdt, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "wrap")
			}
			mkd := wordwrap.String(text.Value, wdth.Value)
			return *env.NewString(mkd)
		},
	},

	"indent": {
		Argsn: 2,
		Doc:   "Wraps string to certain width",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "wrap")
			}
			wdt, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "wrap")
			}
			mkd := indent.String(text.Value, wdth.Value)
			return *env.NewString(mkd)
		},
	},

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

	// font colors
	"black": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBlack()
			return env.NewInteger(1)
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
	"yellow": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorYellow()
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
	"white": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorWhite()
			return env.NewInteger(1)
		},
	},

	// font colors
	"str\\black": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBlack())
		},
	},
	"str\\red": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorRed())
		},
	},
	"str\\blue": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBlue())
		},
	},
	"str\\green": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorGreen())
		},
	},
	"str\\yellow": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorYellow())
		},
	},
	"str\\magenta": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorMagenta())
		},
	},
	"str\\cyan": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorCyan())
		},
	},
	"str\\white": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorWhite())
		},
	},

	// bright font colors
	"br-black": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrBlack()
			return env.NewInteger(1)
		},
	},
	"str\\br-black": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrBlack())
			return env.NewInteger(1)
		},
	},
	"br-red": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrRed()
			return env.NewInteger(1)
		},
	},
	"br-blue": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrBlue()
			return env.NewInteger(1)
		},
	},
	"br-green": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrGreen()
			return env.NewInteger(1)
		},
	},
	"br-yellow": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrYellow()
			return env.NewInteger(1)
		},
	},
	"br-magenta": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrMagenta()
			return env.NewInteger(1)
		},
	},
	"br-cyan": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrBlue()
			return env.NewInteger(1)
		},
	},
	"br-white": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrWhite()
			return env.NewInteger(1)
		},
	},

	// background colors
	"bg-black": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgBlack()
			return env.NewInteger(1)
		},
	},
	"bg-red": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgRed()
			return env.NewInteger(1)
		},
	},
	"bg-blue": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgBlue()
			return env.NewInteger(1)
		},
	},
	"bg-green": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgGreen()
			return env.NewInteger(1)
		},
	},
	"bg-yellow": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgYellow()
			return env.NewInteger(1)
		},
	},
	"bg-magenta": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgMagenta()
			return env.NewInteger(1)
		},
	},
	"bg-cyan": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgBlue()
			return env.NewInteger(1)
		},
	},
	"bg-white": {
		Argsn: 0,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgWhite()
			return env.NewInteger(1)
		},
	},

	// font styles
	"bold": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.Bold()
			return env.NewInteger(1)
		},
	},
	"underline": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.Underline()
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
	"reset\\all": { // TODO -- remove
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.CloseProps()
			return env.NewInteger(1)
		},
	},
	"reset": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.CloseProps()
			return env.NewInteger(1)
		},
	},
	"str\\reset": {
		Argsn: 0,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrCloseProps())
		},
	},
}
