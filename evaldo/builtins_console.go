//go:build !b_norepl && !wasm && !js

package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

var Builtins_console = map[string]*env.Builtin{
	"enter-console": {
		Argsn: 1,
		Doc:   "Stops execution and gives you a Rye console, to test the code inside environment.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch name := arg0.(type) {
			case env.String:
				ser := ps.Ser
				/* ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser */
				//reader := bufio.NewReader(os.Stdin)

				fmt.Println("Welcome to console: \033[1m" + name.Value + "\033[0m")
				fmt.Println("* use \033[1mls\033[0m to list current context")
				fmt.Println("-------------------------------------------------------------")
				/*
					for {
						fmt.Print("{ rye dropin }")
						text, _ := reader.ReadString('\n')
						//fmt.Println(1111)
						// convert CRLF to LF
						text = strings.Replace(text, "\n", "", -1)
						//fmt.Println(1111)
						if strings.Compare("(lc)", text) == 0 {
							fmt.Println(ps.Ctx.Print(*ps.Idx))
						} else if strings.Compare("(r)", text) == 0 {
							ps.Ser = ser
							return ps.Res
						} else {
							// fmt.Println(1111)
							block, genv := loader.LoadString("{ " + text + " }")
							ps := env.AddToProgramState(ps, block.Series, genv)
							EvalBlock(ps)
							fmt.Println(ps.Res.Inspect(*ps.Idx))
						}
					}*/

				DoRyeRepl(ps, "do", ShowResults)
				fmt.Println("-------------------------------------------------------------")
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "enter-console")
			}
		},
	},
}
