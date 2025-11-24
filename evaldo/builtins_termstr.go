//go:build !b_no_term
// +build !b_no_term

package evaldo

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
)

var Builtins_termstr = map[string]*env.Builtin{

	//
	// ##### Terminal String ##### "Terminal ANSI code string functions"
	//

	// Tests:
	// equal { black |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for black text
	"black": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for black text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBlack())
		},
	},

	// Tests:
	// equal { red |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for red text
	"red": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for red text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorRed())
		},
	},

	// Tests:
	// equal { blue |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for blue text
	"blue": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for blue text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBlue())
		},
	},

	// Tests:
	// equal { green |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for green text
	"green": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for green text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorGreen())
		},
	},

	// Tests:
	// equal { yellow |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for yellow text
	"yellow": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for yellow text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorYellow())
		},
	},

	// Tests:
	// equal { magenta |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for magenta text
	"magenta": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for magenta text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorMagenta())
		},
	},

	// Tests:
	// equal { cyan |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for cyan text
	"cyan": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for cyan text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorCyan())
		},
	},

	// Tests:
	// equal { white |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for white text
	"white": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for white text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorWhite())
		},
	},

	// Tests:
	// equal { br-black |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright black text
	"br-black": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright black (gray) text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrBlack())
		},
	},

	// Tests:
	// equal { br-red |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright red text
	"br-red": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright red text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrRed())
		},
	},

	// Tests:
	// equal { br-blue |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright blue text
	"br-blue": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright blue text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrBlue())
		},
	},

	// Tests:
	// equal { br-green |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright green text
	"br-green": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright green text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrGreen())
		},
	},

	// Tests:
	// equal { br-yellow |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright yellow text
	"br-yellow": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright yellow text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrYellow())
		},
	},

	// Tests:
	// equal { br-magenta |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright magenta text
	"br-magenta": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright magenta text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrMagenta())
		},
	},

	// Tests:
	// equal { br-cyan |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright cyan text
	"br-cyan": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright cyan text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrCyan())
		},
	},

	// Tests:
	// equal { br-white |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright white text
	"br-white": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright white text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrWhite())
		},
	},

	// Tests:
	// equal { bg-black |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for black background
	"bg-black": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for black background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgBlack())
		},
	},

	// Tests:
	// equal { bg-red |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for red background
	"bg-red": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for red background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgRed())
		},
	},

	// Tests:
	// equal { bg-blue |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for blue background
	"bg-blue": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for blue background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgBlue())
		},
	},

	// Tests:
	// equal { bg-green |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for green background
	"bg-green": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for green background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgGreen())
		},
	},

	// Tests:
	// equal { bg-yellow |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for yellow background
	"bg-yellow": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for yellow background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgYellow())
		},
	},

	// Tests:
	// equal { bg-magenta |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for magenta background
	"bg-magenta": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for magenta background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgMagenta())
		},
	},

	// Tests:
	// equal { bg-cyan |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for cyan background
	"bg-cyan": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for cyan background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgCyan())
		},
	},

	// Tests:
	// equal { bg-white |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for white background
	"bg-white": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for white background color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBgWhite())
		},
	},

	// Tests:
	// equal { bold |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bold text
	"bold": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bold text style.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrBold())
		},
	},

	// Tests:
	// equal { underline |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for underlined text
	"underline": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for underlined text style.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrUnderline())
		},
	},

	// Tests:
	// equal { reset\bold |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code to reset bold text
	"reset\\bold": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string to reset bold text style.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrResetBold())
		},
	},

	// Tests:
	// equal { reset |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code to reset all styles
	"reset": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string to reset all terminal text styles and colors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrCloseProps())
		},
	},
}
