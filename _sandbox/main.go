// rye project main.go
package main

/*
extern void disableRawMode();
extern void enableRawMode();
*/
import "C"

import (
	"path/filepath"
	"regexp"

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

	"rye/env"
	"rye/evaldo"
	"rye/loader"
	"rye/ryeco"
	"rye/util"

	"github.com/peterh/liner"

	//"rye/util"
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
		} else if os.Args[1] == "ryeco" {
			main_ryeco()
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
			es.Ctx.Set(curIdx, env.Integer{val0})
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

func main_ryeco() {

	// this is experiment to create a golang equivalent of rye code
	// with same datatypes and using the same builtin code
	// so it gets compiled, so we can see what speeds do we get that way
	//defer profile.Start().Stop()
	//input := "{ loop 10000000 { add 1 2 } }"

	// so we need a golang loop and add rye function versions

	//	ryeco_do(func() env.Object { return ryeco_loop(1000, func() env.Object { return ryeco_add(1, 2) }) })

	ryeco.Loop(env.Integer{10000000}, func() env.Object { return ryeco.Add(env.Integer{1}, env.Integer{2}) })

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
	env.SetValue(es, "ctx", *env.NewNative(es.Idx, c, "Rye-echo-context"))
	env.SetValue(es, "session", *env.NewNative(es.Idx, sess, "Rye-echo-session"))
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
	defer profile.Start(profile.CPUProfile).Stop()

	input := "{ loop 50000000 { add 1 2 } }"
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

func main_rye_repl_OLD(in io.Reader, out io.Writer) {

	//util.PrintHeader()
	//defer profile.Start().Stop()

	input := "{ name: \"Rye\" version: \"0.002 alpha\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	fmt.Println("")
	fmt.Println(".:/|\\:. Rye shell v0.014 alpha .:/|\\:.")
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

		line1 := strings.Split(line, "//") //--- just very temporary solution for some comments in repl. Later should probably be part of loader ... maybe?
		//fmt.Println(line1[0])
		block, genv := loader.LoadString("{ " + line1[0] + " }")
		es := env.AddToProgramState(es, block.Series, genv)
		evaldo.EvalBlock(es)
		if es.FailureFlag {
			fmt.Println("\x1b[33m" + "Failure" + "\x1b[0m")
		}
		if es.ErrorFlag {
			fmt.Println("\x1b[31m" + "Critical error:\n" + es.Res + "\x1b[0m")
		}
		es.ReturnFlag = false
		es.ErrorFlag = false
		es.FailureFlag = false

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

var (
	history_fn = filepath.Join(os.TempDir(), ".rye_repl_history")
	names      = []string{"add", "join", "return", "fn", "fail", "if"}
)

type ShellEd struct {
	CurrObj env.Function
	Pause   bool
	Askfor  []string
	Mode    string
	Return  env.Object
}

func main_rye_repl(in io.Reader, out io.Writer) {

	input := "{ name: \"Rye\" version: \"0.002 alpha\" }"
	block, genv := loader.LoadString(input)
	es := env.NewProgramState(block.Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		for i := 0; i < es.Idx.GetWordCount(); i++ {
			if strings.HasPrefix(es.Idx.GetWord(i), strings.ToLower(line)) {
				c = append(c, es.Idx.GetWord(i))
			}
		}
		return
	})

	if f, err := os.Open(history_fn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
	//const PROMPT = "\x1b[6;30;42m Rye \033[m "

	shellEd := ShellEd{env.Function{}, false, make([]string, 0), "", nil}

	// nek variable bo z listo wordow bo ki jih želi setirat v tem okolju in dokler ne pride čez bo repl spraševal za njih
	// name funkcije pa bo prikazal v promptu dokler smo noter , spet en state var
	// isti potek bi lahko uporabili za kreirat live validation dialekte npr daš primer podatka rawmap npr za input in potem pišeš dialekt
	// in preverjaš rezultat ... tako z hitrim reset in ponovi workflowon in prikazom rezultata
	// to s funkcijo se bo dalo čist dobro naredit ... potem pa tudi s kontekstom ne vidim kaj bi bil problem

	line2 := ""

	for {

		prompt, arg := genPrompt(&shellEd, line2)

		if code, err := line.Prompt(prompt); err == nil {
			// strip comment

			multiline := len(code) > 1 && code[len(code)-1:len(code)] == " "

			comment := regexp.MustCompile(`\s//`)
			line1 := comment.Split(code, 2) //--- just very temporary solution for some comments in repl. Later should probably be part of loader ... maybe?
			//fmt.Println(line1)
			lineReal := strings.Trim(line1[0], "\t")

			// check for #shed commands
			maybeDoShedCommands(lineReal, es, &shellEd)

			///fmt.Println(lineReal[len(lineReal)-3 : len(lineReal)])

			if multiline {

				line2 += lineReal + "\n"

			} else {

				line2 += lineReal

				//fmt.Println(lineReal)
				block, genv := loader.LoadString("{ " + line2 + " }")
				es := env.AddToProgramState(es, block.Series, genv)
				evaldo.EvalBlock(es)

				if arg != "" {
					if arg == "<-return->" {
						shellEd.Return = es.Res
					} else {
						es.Ctx.Set(es.Idx.IndexWord(arg), es.Res)
					}
				} else {
					if shellEd.Mode != "" {
						if !shellEd.Pause {
							//fmt.Println(shellEd.CurrObj)
							//fmt.Println("SETIRAM ******* ")
							shellEd.CurrObj.Body.Series.AppendMul(block.Series.GetAll())
							//shellEd.CurrObj.(env.Function).Body.Series.AppendMul(block.Series.GetAll())
							//fmt.Println(shellEd.CurrObj)

						}
					}
					if es.FailureFlag {
						fmt.Println("\x1b[33m" + "Failure" + "\x1b[0m")
					}
					if es.ErrorFlag {
						fmt.Println("\x1b[31m" + "Critical error:\n" + es.Res.Inspect(*genv) + "\x1b[0m")
					}
					es.ReturnFlag = false
					es.ErrorFlag = false
					es.FailureFlag = false

					if es.Res != nil {
						fmt.Print("\033[38;5;37m" + es.Res.Inspect(*genv) + "\x1b[0m")
						if es.Res != nil && shellEd.Mode != "" && !shellEd.Pause && es.Res == shellEd.Return {
							fmt.Println(" <- the correct value was returned")
						} else {
							fmt.Println("")
						}
					}
				}

				line2 = ""

			}

			line.AppendHistory(code)
		} else if err == liner.ErrPromptAborted {
			log.Print("Aborted")
			break
		} else {
			log.Print("Error reading line: ", err)
			break
		}

	}

	if f, err := os.Create(history_fn); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}

func genPrompt(shellEd *ShellEd, line string) (string, string) {
	if shellEd.Mode != "" {
		a := shellEd.Askfor
		if len(a) > 0 {
			x, a := a[0], a[1:]
			shellEd.Askfor = a
			return "{ Rye - value of " + x + " } ", x
		} else if shellEd.Return == nil {
			return "{ Rye - expected return value } ", "<-return->"
		}
		return "{ Rye " + shellEd.Mode + "} ", ""
	} else {
		if len(line) > 0 {

			return "        ", ""
		} else {

			return "{ Rye } ", ""
		}
	}
}

func maybeDoShedCommands(line string, es *env.ProgramState, shellEd *ShellEd) {
	//fmt.Println(shellEd)
	line1 := strings.Split(line, " ")
	var block env.Block
	block = shellEd.CurrObj.Body
	switch line1[0] {
	case "#ra":
		//es.Res.Trace("ADD1")
		block.Series.Append(es.Res)
		//es.Res.Trace("ADD2")
	case "#in":
		//fmt.Println("in")
		//es.Res.Trace("*es.Idx")
		//b := es.Res.(env.Block)
		//fmt.Println(es.Res)
		shellEd.Mode = "fn"
		fn := es.Res.(env.Function)
		words := fn.Spec.Series.GetAll()
		for _, x := range words {
			shellEd.Askfor = append(shellEd.Askfor, es.Idx.GetWord(x.(env.Word).Index))
		}
		shellEd.CurrObj = es.Res.(env.Function)
		//fmt.Println(shellEd)
	case "#ls":
		fmt.Println(shellEd.CurrObj.Inspect(*es.Idx))
	case "#s":
		i := es.Idx.IndexWord(line1[1])
		es.Ctx.Set(i, shellEd.CurrObj)
	case "#.":
		shellEd.Pause = true
	case "#>":
		shellEd.Pause = false
	case "#out":
		shellEd.Mode = ""
	}
}

func maybeDoShedCommandsBlk(line string, es *env.ProgramState, block *env.Block, shed_pause *bool) {
	if block != nil {
		//block.Trace("TOP")
	}
	line1 := strings.Split(line, " ")
	switch line1[0] {
	case "#in":
		//fmt.Println("in")
		//es.Res.Trace("*es.Idx")
		//b := es.Res.(env.Block)
		*block = es.Res.(env.Block)
		//block.Trace("*es.Idx")
	case "#ls":
		fmt.Println(block.Inspect(*es.Idx))
	case "#ra":
		//es.Res.Trace("ADD1")
		block.Series.Append(es.Res)
		//es.Res.Trace("ADD2")
	case "#s":
		i := es.Idx.IndexWord(line1[1])
		es.Ctx.Set(i, *block)
	case "#,":
		*shed_pause = true
	case "#>":
		*shed_pause = false
	case "#out":
		block = nil
	}
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
