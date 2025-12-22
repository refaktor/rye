//go:build !b_no_term
// +build !b_no_term

package evaldo

import (
	"fmt"
	"os"
	"time"

	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
	goterm "golang.org/x/term"
)

var Builtins_term = map[string]*env.Builtin{

	//
	// ##### Terminal ##### "Terminal formatting and styling functions"
	//
	// Tests:
	// equal { "Hello World" |wrap 5 } "Hello\nWorld"
	// Args:
	// * text: String to wrap
	// * width: Integer width to wrap at
	// Returns:
	// * string wrapped at the specified width
	"wrap": {
		Argsn: 2,
		Doc:   "Wraps a string to a specified width by inserting newlines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "wrap")
			}
			wdth, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "wrap")
			}
			mkd := wrap.String(text.Value, int(wdth.Value))
			return *env.NewString(mkd)
		},
	},

	// Tests:
	// equal { "Hello World" |wrap\words 5 } "Hello\nWorld"
	// Args:
	// * text: String to wrap
	// * width: Integer width to wrap at
	// Returns:
	// * string wrapped at the specified width, preserving word boundaries
	"wrap\\words": {
		Argsn: 2,
		Doc:   "Wraps a string to a specified width, preserving word boundaries.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "wrap\\words")
			}
			wdth, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "wrap\\words")
			}
			mkd := wordwrap.String(text.Value, int(wdth.Value))
			return *env.NewString(mkd)
		},
	},

	// Tests:
	// equal { "Hello\nWorld" |indent 2 } "  Hello\n  World"
	// Args:
	// * text: String to indent
	// * spaces: Integer number of spaces to indent each line
	// Returns:
	// * string with each line indented by the specified number of spaces
	"indent": {
		Argsn: 2,
		Doc:   "Indents each line of a string by a specified number of spaces.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "indent")
			}
			wdth, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "indent")
			}
			mkd := indent.String(text.Value, uint(wdth.Value))
			return *env.NewString(mkd)
		},
	},

	// Tests:
	// equal { width? |type? } 'integer
	// Args:
	// * none
	// Returns:
	// * integer width of the terminal in characters
	"width?": {
		Argsn: 0,
		Doc:   "Returns the current width of the terminal in characters.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fd := int(os.Stdout.Fd())
			width, _, err := goterm.GetSize(fd)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "width?")
			}
			return env.NewInteger(int64(width))
		},
	},

	// Tests:
	// stdout { black print "Black text" reset } "Black text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"black": {
		Argsn: 0,
		Doc:   "Sets terminal text color to black.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBlack()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { red print "Red text" reset } "Red text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"red": {
		Argsn: 0,
		Doc:   "Sets terminal text color to red.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorRed()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { blue print "Blue text" reset } "Blue text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"blue": {
		Argsn: 0,
		Doc:   "Sets terminal text color to blue.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBlue()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { green print "Green text" reset } "Green text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"green": {
		Argsn: 0,
		Doc:   "Sets terminal text color to green.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorGreen()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { yellow print "Yellow text" reset } "Yellow text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"yellow": {
		Argsn: 0,
		Doc:   "Sets terminal text color to yellow.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorYellow()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { magenta print "Magenta text" reset } "Magenta text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"magenta": {
		Argsn: 0,
		Doc:   "Sets terminal text color to magenta.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorMagenta()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { cyan print "Cyan text" reset } "Cyan text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"cyan": {
		Argsn: 0,
		Doc:   "Sets terminal text color to cyan.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBlue()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { white print "White text" reset } "White text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"white": {
		Argsn: 0,
		Doc:   "Sets terminal text color to white.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorWhite()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// equal { str\black |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for black text
	"str\\black": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for black text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBlack())
		},
	},

	// Tests:
	// equal { str\red |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for red text
	"str\\red": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for red text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorRed())
		},
	},

	// Tests:
	// equal { str\blue |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for blue text
	"str\\blue": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for blue text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBlue())
		},
	},

	// Tests:
	// equal { str\green |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for green text
	"str\\green": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for green text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorGreen())
		},
	},

	// Tests:
	// equal { str\yellow |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for yellow text
	"str\\yellow": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for yellow text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorYellow())
		},
	},

	// Tests:
	// equal { str\magenta |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for magenta text
	"str\\magenta": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for magenta text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorMagenta())
		},
	},

	// Tests:
	// equal { str\cyan |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for cyan text
	"str\\cyan": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for cyan text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorCyan())
		},
	},

	// Tests:
	// equal { str\white |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for white text
	"str\\white": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for white text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorWhite())
		},
	},

	// Tests:
	// stdout { br-black print "Bright Black text" reset } "Bright Black text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-black": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright black (gray).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrBlack()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// equal { str\br-black |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code for bright black text
	"str\\br-black": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string for bright black (gray) text color.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrColorBrBlack())
		},
	},

	// Tests:
	// stdout { br-red print "Bright Red text" reset } "Bright Red text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-red": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright red.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrRed()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { br-blue print "Bright Blue text" reset } "Bright Blue text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-blue": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright blue.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrBlue()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { br-green print "Bright Green text" reset } "Bright Green text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-green": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright green.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrGreen()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { br-yellow print "Bright Yellow text" reset } "Bright Yellow text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-yellow": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright yellow.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrYellow()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { br-magenta print "Bright Magenta text" reset } "Bright Magenta text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-magenta": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright magenta.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrMagenta()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { br-cyan print "Bright Cyan text" reset } "Bright Cyan text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-cyan": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright cyan.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrBlue()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { br-white print "Bright White text" reset } "Bright White text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"br-white": {
		Argsn: 0,
		Doc:   "Sets terminal text color to bright white.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBrWhite()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-black print "Black background" reset } "Black background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-black": {
		Argsn: 0,
		Doc:   "Sets terminal background color to black.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgBlack()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-red print "Red background" reset } "Red background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-red": {
		Argsn: 0,
		Doc:   "Sets terminal background color to red.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgRed()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-blue print "Blue background" reset } "Blue background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-blue": {
		Argsn: 0,
		Doc:   "Sets terminal background color to blue.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgBlue()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-green print "Green background" reset } "Green background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-green": {
		Argsn: 0,
		Doc:   "Sets terminal background color to green.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgGreen()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-yellow print "Yellow background" reset } "Yellow background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-yellow": {
		Argsn: 0,
		Doc:   "Sets terminal background color to yellow.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgYellow()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-magenta print "Magenta background" reset } "Magenta background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-magenta": {
		Argsn: 0,
		Doc:   "Sets terminal background color to magenta.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgMagenta()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-cyan print "Cyan background" reset } "Cyan background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-cyan": {
		Argsn: 0,
		Doc:   "Sets terminal background color to cyan.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgBlue()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bg-white print "White background" reset } "White background"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bg-white": {
		Argsn: 0,
		Doc:   "Sets terminal background color to white.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ColorBgWhite()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bold print "Bold text" reset } "Bold text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"bold": {
		Argsn: 0,
		Doc:   "Sets terminal text style to bold.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.Bold()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { underline print "Underlined text" reset } "Underlined text"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"underline": {
		Argsn: 0,
		Doc:   "Sets terminal text style to underlined.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.Underline()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// stdout { bold print "Bold" reset\bold print " not bold" } "Bold not bold"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"reset\\bold": {
		Argsn: 0,
		Doc:   "Resets terminal text bold style.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.ResetBold()
			return env.NewInteger(1)

		},
	},

	// Tests:
	// stdout { red bold print "Red bold" reset print " normal" } "Red bold normal"
	// Args:
	// * none
	// Returns:
	// * integer 1 (success indicator)
	"reset": {
		Argsn: 0,
		Doc:   "Resets all terminal text styles and colors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.CloseProps()
			return env.NewInteger(1)
		},
	},

	// Tests:
	// equal { str\reset |type? } 'string
	// Args:
	// * none
	// Returns:
	// * string containing ANSI escape code to reset all styles
	"str\\reset": {
		Argsn: 0,
		Doc:   "Returns ANSI escape code string to reset all terminal text styles and colors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(term.StrCloseProps())
		},
	},

	// like time-it - maybe change name
	"spin-it": {
		Argsn: 2,
		Doc:   "Takes a block of code and a message string, shows a spinner with the message while evaluating",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				switch msg := arg0.(type) {
				case env.String:

					done := make(chan bool)
					spinnerDone := make(chan bool)

					go func() {
						ser := ps.Ser
						ps.Ser = bloc.Series
						ps.BlockFile = bloc.FileName
						ps.BlockLine = bloc.Line
						EvalBlock(ps)
						MaybeDisplayFailureOrError(ps, ps.Idx, "spin-it")
						ps.Ser = ser
						done <- true
					}()

					// Spinner characters
					spinner := []rune{'|', '/', '-', '\\'}
					i := 0

					// Show spinner until process completes
					go func() {
						for {
							select {
							case <-done:
								fmt.Printf("\r")
								spinnerDone <- true
								return
							default:
								fmt.Printf("\r%s %c", msg.Value, spinner[i%len(spinner)])
								i++
								time.Sleep(120 * time.Millisecond)
							}
						}
					}()

					// Wait for the spinner to finish
					<-spinnerDone
					return ps.Res
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType}, "spin-it")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "spin-it")
			}
		},
	},
}
