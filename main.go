//+build linux darwin windows

package main

import (
	"os/user"
	"path/filepath"

	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	// "strconv"
	"strings"

	"net/http"
	"rye/env"
	"rye/evaldo"
	"rye/loader"
	"rye/util"

	"net/http/cgi"
	"rye/ryeco"
	// enable when using profiler
	// "github.com/pkg/profile"
)

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

func main() {
	if len(os.Args) == 1 {
		main_rye_repl(os.Stdin, os.Stdout)
	} else if len(os.Args) == 2 {
		if os.Args[1] == "shell" {
			main_rysh()
		} else if os.Args[1] == "web" {
			// main_httpd()
		} else if os.Args[1] == "ryeco" {
			main_ryeco()
		} else {
			main_rye_file(os.Args[1], false)
		}
	} else if len(os.Args) == 3 {
		if os.Args[1] == "ryk" {
			main_ryk(os.Args[2])
		} else if os.Args[1] == "cgi" {
			main_cgi_file(os.Args[2], false)
		} else if os.Args[1] == "sig" {
			main_rye_file(os.Args[2], true)
		}
	}
}

//
// main for awk like functionality with rye language
//

func main_ryk(code string) {

	block, genv := loader.LoadString(code, false)
	// make code composable, updatable ... so you can load by appending to existing program/state or initial block?
	// basically we need to have multiple toplevel blocks that can be evaluated by the same state
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		idx0 := es.Idx.IndexWord("_0") // turn to _0, _1 or something like it via separator later ..
		idx1 := es.Idx.IndexWord("_1") // turn to _0, _1 or something like it via separator later ..
		idx2 := es.Idx.IndexWord("_2") // turn to _0, _1 or something like it via separator later ..
		//printidx, _ := es.Idx.GetIndex("print")
		// val0, er := strconv.ParseInt(scanner.Text(), 10, 64)
		val0 := util.StringToFieldsWithQuoted(scanner.Text(), " ", "\"")
		// if er == nil {
		es.Ctx.Set(idx0, val0.Series.Get(0))
		es.Ctx.Set(idx1, val0.Series.Get(1))
		es.Ctx.Set(idx2, val0.Series.Get(2))
		evaldo.EvalBlockInj(es, val0, true)
		es.Ser.Reset()
		//} else {
		//	fmt.Println("error processing line: " + scanner.Text())
		// }
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func main_ryeco() {

	// this is experiment to create a golang equivalent of rye code
	// with same datatypes and using the same builtin code
	// so it gets compiled, so we can see what speeds do we get that way
	// defer profile.Start().Stop()
	//input := "{ loop 10000000 { add 1 2 } }"

	// so we need a golang loop and add rye function versions

	// ryeco_do(func() env.Object { return ryeco_loop(1000, func() env.Object { return ryeco_add(1, 2) }) })

	ryeco.Loop(env.Integer{10000000}, func() env.Object { return ryeco.Inc(env.Integer{1}) })

}

func main_rye_file(file string, sig bool) {

	//util.PrintHeader()
	//defer profile.Start(profile.CPUProfile).Stop()

	bcontent, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	content := string(bcontent)

	block, genv := loader.LoadString(content, sig)
	switch val := block.(type) {
	case env.Block:
		es := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(es)
		evaldo.EvalBlock(es)
		evaldo.MaybeDisplayFailureOrError(es, genv)
	case env.Error:
		fmt.Println(val.Message)
	}

}

func main_cgi_file(file string, sig bool) {

	if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//util.PrintHeader()
		//defer profile.Start(profile.CPUProfile).Stop()

		input := " whoami: \"Rye cgi 0.001 alpha\" ctx: 0 result: \"\" session: 0 w: 0 r: 0"
		block, genv := loader.LoadString(input, false)
		es := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(es)
		evaldo.EvalBlock(es)
		env.SetValue(es, "w", *env.NewNative(es.Idx, w, "Go-server-response-writer"))
		env.SetValue(es, "r", *env.NewNative(es.Idx, r, "Go-server-request"))

		bcontent, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		content := string(bcontent)

		block, genv = loader.LoadString(content, sig)
		switch val := block.(type) {
		case env.Block:
			es := env.AddToProgramState(es, block.(env.Block).Series, genv)
			evaldo.RegisterBuiltins(es)
			evaldo.EvalBlock(es)
			evaldo.MaybeDisplayFailureOrError(es, genv)
		case env.Error:
			fmt.Println(val.Message)
		}

	})); err != nil {
		fmt.Println(err)
	}

}

func main_rye_repl(in io.Reader, out io.Writer) {

	input := "name: \"Rye\" version: \"0.002 alpha\""
	user, _ := user.Current()
	profile_path := filepath.Join(user.HomeDir, ".rye-profile")

	if _, err := os.Stat(profile_path); err == nil {
		content, err := ioutil.ReadFile(profile_path)
		if err != nil {
			log.Fatal(err)
		}
		input = string(content)
	} else {
		fmt.Print("no profile")

	}

	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.EvalBlock(es)

	evaldo.DoRyeRepl(es)

}

func main_rysh() {
	reader := bufio.NewReader(os.Stdin)
	status := 1

	for status != 0 {
		// C.enableRawMode()
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
		// C.disableRawMode()
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
