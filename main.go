// Rejy_go_v1 project main.go
package main

/*
extern void disableRawMode();
extern void enableRawMode();
*/
import "C"

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	//"regexp"
	"strconv"
	"strings"

	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"Rejy_go_v1/env"
	"Rejy_go_v1/evaldo"
	"Rejy_go_v1/loader"
	"Rejy_go_v1/util"

	//"Rejy_go_v1/util"
	//"fmt"
	//"strconv"
	"github.com/pkg/profile"
	//"github.com/pkg/term"
)

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
)

// REJY0 in GoLang

// contrary to JS rejy version, parser here already indexes word names into global word index, not evaluator.
// This means one intermediate step less (one data format less, and one conversion less)

// parser produces a tree of values words and blocks in an array.
// primitive values are stored unboxed, we can do this with series   []interface{}
// complex values are stored as struct { type, index } (words, setwords)
// functions are stored similarly. Probably argument count should be in struct too.

type TagType int
type RjType int
type Series []interface{}

type anyword struct {
	kind RjType
	idx  int
}

type node struct {
	kind  RjType
	value interface{}
}

var CODE []interface{}

//
// main function. Dispatches to appropriate mode function
//
// rye_repl - repl of rye language (like rebol repl)
// rye_file - loads a script file and runs it
// rysh - starts a rysh experimental shell
// ryk  - starts as an awk like tool with rye programming language

func main() {
	if len(os.Args) == 1 {
		main_rye_repl(os.Stdin, os.Stdout)
	} else if len(os.Args) == 2 {
		if os.Args[1] == "shell" {
			main_rysh()
		} else if os.Args[1] == "web" {
			main_httpd()
		} else {
			main_rye_file(os.Args[1])
		}
	} else if len(os.Args) == 3 && os.Args[1] == "ryk" {
		main_ryk(os.Args[2])
	}
}

//
// main for awk like functionality with rye language
//

func main_ryk(code string) {

	block, genv := loader.LoadString(code)
	// make code composable, updatable ... so you can load by appending to existing program/state or initial block?
	// basically we need to have multiple toplevel blocks that can be evaluated by the same state
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		curIdx := es.Idx.IndexWord("line") // turn to _0, _1 or something like it via separator later ..
		//printidx, _ := es.Idx.GetIndex("print")
		val0, er := strconv.ParseInt(scanner.Text(), 10, 64)
		if er == nil {
			es.Env.Set(curIdx, env.Integer{val0})
			evaldo.EvalBlock(es)
			es.Ser.Reset()
		} else {
			fmt.Println("error processing line: " + scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	//util.PrintHeader()
	///// defer profile.Start().Stop()

	//input := "{ loop 10000000 { add 1 2 } }"

	/*genv := loader.GetIdxs()
		ps := evaldo.ProgramState{}

		// Parse
		loader1 := loader.NewLoader()
		input := "{ 123 word 3 { setword: 23 } end 12 word }"
		val, _ := loader1.ParseAndGetValue(input, nil)
		loader.InspectNode(val)
		evaldo.EvalBlock(ps, val.(env.Object))
		fmt.Println(val)

		genv.Probe()

	a	fmt.Println(strconv.FormatInt(int64(genv.GetWordCount()), 10))*/

}

func main_httpd() {

	//util.PrintHeader()
	//	defer profile.Start().Stop()

	// Echo instance
	e := echo.New()

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret122849320"))))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", serve)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// Handler
func serve(c echo.Context) error {

	sess, _ := session.Get("session", c)
	/*sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}*/
	//sess.Values["foo"] = env.String{"barca"}

	input := "{ whoami: \"Rye webserver 0.001 alpha\" ctx: 0 result: \"\" session: 0 }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	//evaldo.RegisterBuiltins2(evaldo.Buil, ps *env.ProgramState) {
	evaldo.EvalBlock(es)

	env.SetValue(es, "ctx", env.Native{c})
	env.SetValue(es, "session", env.Native{sess})
	//env.SetValue(es, "ctx", env.String{"YOYOYO"})

	b1, err := ioutil.ReadFile("web/index.html")
	if err != nil {
		fmt.Print(err)
	}

	var bu strings.Builder

	b := string(b1)

	idx1 := 0

	for {
		ii := util.IndexOfAt(b, "<%", idx1)
		if ii > -1 {
			ii2 := util.IndexOfAt(b, "%>", ii)
			if ii2 > -1 {

				fmt.Println("OUTPUT")
				fmt.Println(idx1)
				fmt.Println(ii)
				fmt.Println(b[idx1:ii])
				bu.WriteString(b[idx1:ii])

				displayBlock := false
				if string(b[ii+2]) == "?" {
					ii++
					displayBlock = true
				}

				code := b[ii+2 : ii2]

				if displayBlock {
					bu.WriteString("<pre class='rye-code'>" + code + "</pre>")
				}
				fmt.Print("CODE: ")
				fmt.Println(code)
				block, genv := loader.LoadString("{ " + code + " }")
				es := env.AddToProgramState(es, block.Series, genv)
				evaldo.EvalBlock(es)
				if displayBlock {
					bu.WriteString("<div class='rye-result'>")
				}
				bu.WriteString(evaldo.PopOutBuffer())
				/*
					ACTIVATE LATER FOR <%= tags %>
						if es.Res != nil {
						//fmt.Println("\x1b[6;30;42m" + es.Res.Inspect(*genv) + "\x1b[0m")
						if es.Res.Type() == env.IntegerType {
							bu.WriteString(strconv.FormatInt(es.Res.(env.Integer).Value, 10))
						} else if es.Res.Type() == env.StringType {
							bu.WriteString(es.Res.(env.String).Value)
						}
					} */
				if displayBlock {
					bu.WriteString("</div>")
				}
				idx1 = ii2 + 2
			} else {
				fmt.Println("Err 214")
				break
			}
		} else {
			bu.WriteString(b[idx1:])
			break
		}
	}

	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return c.HTML(http.StatusOK, err.Error()) // tempsol: make this work better .. RETURN http error

	}

	return c.HTML(http.StatusOK, bu.String())
}

func main_rye_file(file string) {

	//util.PrintHeader()
	defer profile.Start().Stop()

	input := "{ loop 10000000 { add 1 2 } }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	/*genv := loader.GetIdxs()
		ps := evaldo.ProgramState{}

		// Parse
		loader1 := loader.NewLoader()
		input := "{ 123 word 3 { setword: 23 } end 12 word }"
		val, _ := loader1.ParseAndGetValue(input, nil)
		loader.InspectNode(val)
		evaldo.EvalBlock(ps, val.(env.Object))
		fmt.Println(val)

		genv.Probe()

	a	fmt.Println(strconv.FormatInt(int64(genv.GetWordCount()), 10))*/

}

func main_rye_repl(in io.Reader, out io.Writer) {

	//util.PrintHeader()
	//defer profile.Start().Stop()

	input := "{ name: \"Rye\" version: \"0.002 alpha\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println("")
	fmt.Println(".:/|\\:. Rye shell v0.006 alpha .:/|\\:.")
	fmt.Println("")

	const PROMPT = "\n\x1b[6;30;42m Rye \033[m "

	scanner := bufio.NewScanner(in)

	for {

		fmt.Print(PROMPT)

		scanned := scanner.Scan()

		if !scanned {
			return
		}

		line := scanner.Text()

		block, genv := loader.LoadString("{ " + line + " }")
		es := env.AddToProgramState(es, block.Series, genv)
		evaldo.EvalBlock(es)

		if es.Res != nil {
			//fmt.Println("\x1b[6;30;42m" + es.Res.Inspect(*genv) + "\x1b[0m")
			fmt.Println("\x1b[0;49;32m" + es.Res.Inspect(*genv) + "\x1b[0m")
		}
	}

	/*genv := loader.GetIdxs()
		ps := evaldo.ProgramState{}

		// Parse
		loader1 := loader.NewLoader()
		input := "{ 123 word 3 { setword: 23 } end 12 word }"
		val, _ := loader1.ParseAndGetValue(input, nil)
		loader.InspectNode(val)
		evaldo.EvalBlock(ps, val.(env.Object))
		fmt.Println(val)

		genv.Probe()

	a	fmt.Println(strconv.FormatInt(int64(genv.GetWordCount()), 10))*/

}

func main_rysh() {
	reader := bufio.NewReader(os.Stdin)
	status := 1

	for status != 0 {
		C.enableRawMode()
		wd, _ := os.Getwd()
		fmt.Print("\033[36m" + wd + " -> " + "\033[m")

		line, cursorPos, shellEditor := "", 0, false

		for {

			c, _ := reader.ReadByte()
			//fmt.Print(c)
			if c == 13 {
				//line = line[:len(line)-1]
				fmt.Println()
				shellEditor = false
				cursorPos = len(line)
				fmt.Print("YOLO")
				break
			}
			shellEditor = false

			if c == 27 {
				c1, _ := reader.ReadByte()
				if c1 == '[' {
					c2, _ := reader.ReadByte()
					switch c2 {
					/*case 'A':
						if len(HISTMEM) != 0 && histCounter < len(HISTMEM) {
							for cursorPos > 0 {
								fmt.Printf("\b\033[J")
								cursorPos--
							}
							line = strings.Split(HISTMEM[histCounter], "::")[2]
							fmt.Printf(line)
							cursorPos = len(line)
							histCounter++
						}
					case 'B':
						if len(HISTMEM) != 0 && histCounter > 0 {
							for cursorPos > 0 {
								fmt.Printf("\b\033[J")
								cursorPos--
							}
							histCounter--
							line = strings.Split(HISTMEM[histCounter], "::")[2]
							fmt.Printf(line)
							cursorPos = len(line)
						}*/
					case 'C':
						if cursorPos < len(line) {
							fmt.Printf("\033[C")
							cursorPos++
						}
					case 'D':
						if cursorPos > 0 {
							fmt.Printf("\033[D")
							cursorPos--
						}
					case 'A':
						fmt.Printf("\033[A")
					case 'B':
						fmt.Printf("\033[2K\r")
						fmt.Print("lovely")
					}
				}
				continue
			}
			// backspace was pressed
			if c == 127 {
				if cursorPos > 0 {
					if cursorPos != len(line) {
						temp, oldLength := line[cursorPos:], len(line)
						fmt.Printf("\b\033[K%s", temp)
						for oldLength != cursorPos {
							fmt.Printf("\033[D")
							oldLength--
						}
						line = line[:cursorPos-1] + temp
						cursorPos--
					} else {
						fmt.Print("\b\033[K")
						line = line[:len(line)-1]
						cursorPos--
					}
				}
				continue
			}
			// ctrl-c was pressed
			if c == 3 {
				fmt.Println("^C")
				/// discard = true
				break
			}
			// ctrl-d was pressed
			if c == 4 {
				os.Exit(0)
			}
			// the enter key was pressed
			if c == 13 {
				fmt.Println()
				break
			}

			if cursorPos == len(line) {
				fmt.Printf("%c", c)
				line += string(c)
				cursorPos = len(line)
			} else {
				temp, oldLength := line[cursorPos:], len(line)
				fmt.Printf("\033[K%c%s", c, temp)
				for oldLength != cursorPos {
					fmt.Printf("\033[D")
					oldLength--
				}
				line = line[:cursorPos] + string(c) + temp
				cursorPos++
			}
			if c == '\\' {
				fmt.Print("NEWLINE")
				shellEditor = true
				fmt.Print(shellEditor)
			}
		}
		// Read the keyboad input.
		//input, err := reader.ReadString('\n')
		//if err != nil {
		//	fmt.Fprintln(os.Stderr, err)
		//}
		fmt.Println("OUT OUT")
		C.disableRawMode()
		// Handle the execution of the input.
		if err := execInput(line); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")

func execInput(input string) error {
	// Remove the newline character.
	input = strings.TrimSuffix(input, "\n")

	// Split the input separate the command and the arguments.
	args := strings.Split(input, " ")

	// Check for built-in commands.
	switch args[0] {
	case "cd":
		// 'cd' to home with empty path not yet supported.
		if len(args) < 2 {
			return os.Chdir(os.Getenv("HOME"))
		}
		// Change the directory and return the error.
		return os.Chdir(args[1])
	case "exit":
		os.Exit(0)
	}

	// Prepare the command to execute.
	cmd := exec.Command(args[0], args[1:]...)

	//
	// look at this page on how to capture the output and pass it through:
	// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
	// we tested it and it worked, but changed ls to multileine outpu. This could be also used to do some ouput formatting like
	// adding colors. Maybe other shells use similar approaches and so this should work
	// With this, we could "move cursor in the background string and get the word under the string like we wanted to
	// but maybe we would screw up some commands ... test it like ls | wc -l could function a little differently.
	// or htop (which exited in previous case also) and less which worked with default stream handling
	// maybe we could detect these special commands and let them through as they are and maybe when we are piping
	// we wouldn't buffer output since it won't be displayed anyway
	//
	// OK. so we solved the ls mistery. It displays / outputs uniformly. So this could be done it seems.
	// https://unix.stackexchange.com/questions/10421/output-from-ls-has-newlines-but-displays-on-a-single-line-why
	//
	// looked at the --color option and saw:
	// Color codes are emitted only on standard output; not in pipes or redirection.
	// We could capture and print in same case. Unless pipes or redirection. (What exactly is redirection?)
	//
	// if we can do this ... copy current word succsesfully at least for last command then this would be at least one good reason
	// to do this shell outside rye stuff or json output stuff.
	// idea ... instead of going over past commands we could also go over past outputs and reuse them via cursors
	// idea ... alt would be like smart modified. alt up would go to previous command. alt left/right would move over words etc
	// alt-u w would use the word , alt-l the line ... there would be some dynamic behaviout that would enable you to type commands and reuse
	// words, lines, previous commands so you wouldn't have to keep moving up and down to combine typing and using / taking
	// like alt-b would take cursor back to where it last took ... taking would take you down to typing
	// maybe the command line below could show the currently selected word / selection and alt-enter would take it and execute it
	// alt-m would select more .. more words lines etc alt-l less. or alt-n next (right) and alt-p previous (left).
	//
	// idea ... ability to document your procedures, undo them, comment them, save them to recipe files
	//
	// idea task-contexts ... pack histories in specific task-contexts. So you can switch to some context and get
	// see recipes, get the history of commands from that task ... taks-context can be remembered by path you are in
	// so you can also list task context related to current path or list them via paths furtner in.
	//

	// Set the correct output device.
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error.
	return cmd.Run()
}
