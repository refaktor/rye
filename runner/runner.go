//go:build !wasm

package runner

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/term"

	"github.com/refaktor/rye/contrib"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
	"github.com/refaktor/rye/util"
)

var (
	// fileName = flag.String("fiimle", "", "Path to the Rye file (default: none)")
	do     = flag.String("do", "", "Evaluates code after it loads a file or last save.")
	lang   = flag.String("lang", "do", "Select a dialect / language (do, eyr, ...)")
	silent = flag.Bool("silent", false, "Console doesn't display return values")
	//	quit    = flag.Bool("quit", false, "Quits after executing.")
	console = flag.Bool("console", false, "Enters console after a file is evaluated.")
	help    = flag.Bool("help", false, "Displays this help message.")
)

func DoMain(regfn func(*env.ProgramState)) {
	flag.Usage = func() {
		fmt.Println("╭────────────────────────────────────────────────────────────────────────────────────────────---")
		fmt.Println("│ \033[1mRye\033[0m language. Visit \033[36mhttps://ryelang.org\033[0m to learn more.                            ")
		fmt.Println("╰───────────────────────────────────────────────────────────────────────────────────────---")
		fmt.Println("\n Usage: \033[1mrye\033[0m [\033[1moptions\033[0m] [\033[1mfilename\033[0m or \033[1mcommand\033[0m]")
		fmt.Println("\n To enter \033[1mRye console\033[0m provide no filename or command.")
		fmt.Println("\n \033[1mOptions\033[0m (optional)")
		flag.PrintDefaults()
		fmt.Println("\n \033[1mFilename:\033[0m (optional)")
		fmt.Println("  [filename]   \n       Executes a Rye file")
		fmt.Println("  .            \n       Executes a main.rye in current directory")
		fmt.Println("  [some/path]/.\n       Executes a main.rye on some path")
		fmt.Println("\n \033[1mCommands:\033[0m (optional)")
		fmt.Println("  cont\n     Continue console from the last save")
		fmt.Println("  here\n     Starts in Rye here mode (wip)")
		fmt.Println(" \033[1mExamples:\033[0m")
		fmt.Println("\033[33m  rye                                  \033[36m# enters console/REPL")
		fmt.Println("\033[33m  rye -do \"print 33 * 42\"              \033[36m# evaluates the do code")
		fmt.Println("\033[33m  rye -do 'name: \"Jim\"' console        \033[36m# evaluates the do code and enters console")
		fmt.Println("\033[33m  rye cont                             \033[36m# continues/loads last saved state and enters console")
		fmt.Println("\033[33m  rye -do 'print 10 + 10' cont         \033[36m# continues/loads last saved state, evaluates do code and enters console")
		fmt.Println("\033[33m  rye filename.rye                     \033[36m# evaluates filename.rye")
		fmt.Println("\033[33m  rye .                                \033[36m# evaluates main.rye in current directory")
		fmt.Println("\033[33m  rye some/path/.                      \033[36m# evaluates main.rye in some/path/")
		fmt.Println("\033[33m  rye -do 'print \"Hello\" path/.        \033[36m# evaluates main.rye in path/ and then do code")
		fmt.Println("\033[33m  rye -console file.rye                \033[36m# evaluates file.rye and enters console")
		fmt.Println("\033[33m  rye -do 'print 123' -console .       \033[36m# evaluates main.rye in current dir. evaluates do code and enters console")
		fmt.Println("\033[33m  rye -silent                          \033[36m# enters console in that doesn't show return values - silent mode")
		fmt.Println("\033[33m  rye -silent -console file.rye        \033[36m# evaluates file.re and enters console in silent mode")
		fmt.Println("\033[0m\n Thank you for trying out \033[1mRye\033[22m ...")
		fmt.Println("")
	}
	// Parse flags
	flag.Parse()

	evaldo.ShowResults = !*silent

	var code string
	if *do != "" {
		code = *do
	}

	// Check for --help flag
	if flag.NFlag() == 0 && flag.NArg() == 0 {
		if Option_Embed_Main {
			main_rye_file("buildtemp/main.rye", false, true, *console, code, regfn)
		} else if Option_Do_Main {
			ryeFile := dotsToMainRye(".")
			main_rye_file(ryeFile, false, true, *console, code, regfn)
		} else {
			main_rye_repl(os.Stdin, os.Stdout, true, false, *lang, regfn)
		}
	} else {
		// Check for --help flag
		if *help {
			flag.Usage()
			os.Exit(0)
		}

		args := flag.Args()
		// Check for subcommands (cont) and handle them
		if len(args) > 0 {
			if args[0] == "cont" {
				fmt.Println("[continuing...]")
				ryeFile := findLastConsoleSave()
				main_rye_file(ryeFile, false, true, true, code, regfn)
			} else if args[0] == "here" {
				main_rye_repl(os.Stdin, os.Stdout, true, true, *lang, regfn)
			} else {
				ryeFile := dotsToMainRye(args[0])
				main_rye_file(ryeFile, false, true, *console, code, regfn)
			}
		} else {
			if *do != "" {
				main_rye_file("", false, true, *console, code, regfn)
			} else {
				main_rye_repl(os.Stdin, os.Stdout, true, false, *lang, regfn)
			}
		}
	}
}

func findLastConsoleSave() string {
	// Read directory entries
	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err) // TODO --- report better
		return ""
	}

	files := make([]string, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}
		if strings.HasPrefix(entry.Name(), "console_") {
			files = append(files, entry.Name())
		}
	}

	if len(files) == 0 {
		return ""
	}

	sort.Strings(files)

	return files[len(files)-1]
}

func dotsToMainRye(ryeFile string) string {
	re := regexp.MustCompile(`^\.$|/\.$`)
	if re.MatchString(ryeFile) {
		main_path := ryeFile[:len(ryeFile)-1] + "main.rye"
		if _, err := os.Stat(main_path); err == nil || Option_Embed_Main {
			_, err := os.ReadFile(main_path)
			if err != nil {
				log.Fatal(err)
			}
			return main_path
		} else {
			fmt.Println("There was no main.rye")
		}
	}
	return ryeFile
}

//
// main for awk like functionality with rye language
//

func main_ryk() {
	argIdx := 2
	ignore := 0
	separator := " "
	input := " 1 "

	// 	fmt.Print("preload")

	profile_path := ".ryk-preload"

	if _, err := os.Stat(profile_path); err == nil {
		content, err := os.ReadFile(profile_path)
		if err != nil {
			log.Fatal(err)
		}
		input = string(content)
	}

	block, genv := loader.LoadString(input, false)
	//block, genv := loader.LoadString("{ }", false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)
	evaldo.EvalBlock(es)

	if len(os.Args) >= 4 {
		if os.Args[argIdx] == "--skip" {
			ignore = 1
			argIdx++
		}
		if os.Args[argIdx] == "--csv" {
			separator = ","
			argIdx++
		}
		if os.Args[argIdx] == "--tsv" {
			separator = "\t"
			argIdx++
		}
	}

	var filter *regexp.Regexp
	var filterBlock *env.Object

	if len(os.Args) >= 5 {
		if os.Args[argIdx] == "--begin" {
			block, genv := loader.LoadString(os.Args[argIdx+1], false)
			es = env.AddToProgramState(es, block.(env.Block).Series, genv)
			evaldo.EvalBlockInj(es, es.ForcedResult, true)
			es.Ser.Reset()
			argIdx += 2
		}
		if os.Args[argIdx] == "--filter" {
			code := os.Args[argIdx+1]
			if code[0] == '/' {
				filter = regexp.MustCompilePOSIX(code[1 : len(code)-1])
			} else {
				filterBlock1, genv1 := loader.LoadString(code, false)
				es = env.AddToProgramState(es, filterBlock1.(env.Block).Series, genv1)
				filterBlock = &filterBlock1
			}
			argIdx += 2
		}
	}

	code := os.Args[argIdx]

	block1, genv1 := loader.LoadString(code, false)
	es = env.AddToProgramState(es, block1.(env.Block).Series, genv1)
	// make code composable, updatable ... so you can load by appending to existing program/state or initial block?
	// basically we need to have multiple toplevel blocks that can be evaluated by the same state

	scanner := bufio.NewScanner(os.Stdin)
	nn := 1
	for scanner.Scan() {
		doLine := true
		if filter != nil {
			doLine = filter.MatchString(scanner.Text())
		}
		if ignore > 0 {
			ignore--
		} else {
			if doLine {
				//fmt.Println(scanner.Text())
				N := es.Idx.IndexWord("n") // turn to _0, _1 or something like it via separator later ..
				L := es.Idx.IndexWord("l") // turn to _0, _1 or something like it via separator later ..
				//idx1 := es.Idx.IndexWord("f1") // turn to _0, _1 or something like it via separator later ..
				//idx2 := es.Idx.IndexWord("f2") // turn to _0, _1 or something like it via separator later ..
				//printidx, _ := es.Idx.GetIndex("print")
				// val0, er := strconv.ParseInt(scanner.Text(), 10, 64)
				val0 := util.StringToFieldsWithQuoted(scanner.Text(), separator, "\"")
				// if er == nil {
				es.Ctx.Set(N, *env.NewInteger(int64(nn)))
				es.Ctx.Set(L, *env.NewInteger(int64(val0.Series.Len())))
				if filterBlock != nil {
					blk := *filterBlock
					es.Ser = blk.(env.Block).Series
					evaldo.EvalBlockInj(es, val0, true)
					es.Ser.Reset()
					doLine = util.IsTruthy(es.Res)
				}
				if doLine {
					es.Ser = block1.(env.Block).Series
					evaldo.EvalBlockInj(es, val0, true)
					es.Ser.Reset()
				}
				//} else {
				//	fmt.Println("error processing line: " + scanner.Text())
				// }
			}
			nn++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	argIdx += 1

	if len(os.Args) >= argIdx+2 {
		if os.Args[argIdx] == "--end" {
			block, genv := loader.LoadString(os.Args[argIdx+1], false)
			es = env.AddToProgramState(es, block.(env.Block).Series, genv)
			evaldo.EvalBlockInj(es, es.ForcedResult, true)
			es.Ser.Reset()
		}
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

	// ryeco.Loop(env.Integer{10000000}, func() env.Object { return ryeco.Inc(env.Integer{1}) })

}

func main_rye_file(file string, sig bool, subc bool, interactive bool, code string, regfn func(*env.ProgramState)) {
	info := true

	//defer profile.Start(profile.CPUProfile).Stop()

	var content string

	if len(file) > 4 && file[len(file)-4:] == ".enc" {
		fmt.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		password := string(bytePassword)

		content = util.ReadSecure(file, password)
	} else if file != "" {
		var bcontent []byte
		var err error
		if Option_Embed_Main {
			bcontent, err = Rye_files.ReadFile(file)
		} else {
			bcontent, err = os.ReadFile(file)
		}
		if err != nil {
			log.Fatal(err)
		}
		content = string(bcontent)
	} else {
		content = ""
	}

	if info {
		pattern := regexp.MustCompile(`^; (#[^\n]*)`)

		lines := pattern.FindAllStringSubmatch(content, -1)

		for _, line := range lines {
			if line[1] != "" {
				fmt.Println(line[1])
			}
		}
	}

	ps := env.NewProgramStateNEW()
	ps.ScriptPath = file
	ps.WorkingPath, _ = os.Getwd() // TODO -- WHAT SHOULD WE DO IF GETWD FAILS?
	evaldo.RegisterBuiltins(ps)
	contrib.RegisterBuiltins(ps, &evaldo.BuiltinNames)
	regfn(ps)
	// current.RegisterBuiltins(ps)
	// ctx := ps.Ctx
	// ps.Ctx = env.NewEnv(ctx)
	//ES = ps
	// evaldo.ShowResults = false

	block := loader.LoadStringNEW(" "+content+"\n"+code, sig, ps)
	switch val := block.(type) {
	case env.Block:

		//	block, genv := loader.LoadString(content+"\n"+code, sig)
		//	switch val := block.(type) {
		//	case env.Block:
		//es := env.NewProgramState(block.(env.Block).Series, genv)
		//evaldo.RegisterBuiltins(es)
		// contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

		ps = env.AddToProgramState(ps, val.Series, ps.Idx)

		if subc {
			ctx := ps.Ctx
			ps.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlock(ps)
		evaldo.MaybeDisplayFailureOrError(ps, ps.Idx)

		if interactive {
			evaldo.DoRyeRepl(ps, "do", evaldo.ShowResults)
		}

	case env.Error:
		fmt.Println(util.TermError(val.Message))
	}
}

func main_rye_file_OLD(file string, sig bool, subc bool, interactive bool, code string) {
	info := true
	//util.PrintHeader()
	//defer profile.Start(profile.CPUProfile).Stop()

	var content string

	if file[len(file)-4:] == ".enc" {
		fmt.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		password := string(bytePassword)

		content = util.ReadSecure(file, password)
	} else {
		bcontent, err := os.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		content = string(bcontent)
	}

	if info {
		pattern := regexp.MustCompile(`^; (#[^\n]*)`)

		lines := pattern.FindAllStringSubmatch(content, -1)

		for _, line := range lines {
			if line[1] != "" {
				fmt.Println(line[1])
			}
		}
	}

	block, genv := loader.LoadString(content+"\n"+code, sig)
	switch val := block.(type) {
	case env.Block:
		es := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(es)
		contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

		if subc {
			ctx := es.Ctx
			es.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlock(es)
		evaldo.MaybeDisplayFailureOrError(es, genv)

		if interactive {
			evaldo.DoRyeRepl(es, "do", evaldo.ShowResults)
		}

	case env.Error:
		fmt.Println(val.Message)
	}
}

func main_cgi_file(file string, sig bool) {
	if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//util.PrintHeader()
		//defer profile.Start(profile.CPUProfile).Stop()

		input := " 123 " //" whoami: \"Rye cgi 0.001 alpha\" ctx: 0 result: \"\" session: 0 w: 0 r: 0"
		block, genv := loader.LoadString(input, false)
		es := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(es)
		contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

		evaldo.EvalBlock(es)
		env.SetValue(es, "w", *env.NewNative(es.Idx, w, "Go-server-response-writer"))
		env.SetValue(es, "r", *env.NewNative(es.Idx, r, "Go-server-request"))

		bcontent, err := os.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		content := string(bcontent)

		block, genv = loader.LoadString(content, sig)
		switch val := block.(type) {
		case env.Block:
			es = env.AddToProgramState(es, block.(env.Block).Series, genv)
			evaldo.RegisterBuiltins(es)
			contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

			evaldo.EvalBlock(es)
			evaldo.MaybeDisplayFailureOrError(es, genv)
		case env.Error:
			fmt.Println(val.Message)
		}
	})); err != nil {
		fmt.Println(err)
	}
}

func main_rye_repl(_ io.Reader, _ io.Writer, subc bool, here bool, lang string, regfn func(*env.ProgramState)) {
	input := " " // "name: \"Rye\" version: \"0.011 alpha\""
	// userHomeDir, _ := os.UserHomeDir()
	// profile_path := filepath.Join(userHomeDir, ".rye-profile")

	fmt.Println("Welcome to Rye console. Use ls to list current or lsp and lsp\\ \"prin\" to list parent contexts.")

	//if _, err := os.Stat(profile_path); err == nil {
	//content, err := os.ReadFile(profile_path)
	//if err != nil {
	//	log.Fatal(err)
	//}
	// input = string(content)
	//} else {
	//		fmt.Println("There was no profile.")
	//}

	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)
	regfn(es)

	evaldo.EvalBlock(es)

	if subc {
		ctx := es.Ctx
		es.Ctx = env.NewEnv(ctx) // make new context with no parent
	}

	if here {
		if _, err := os.Stat(".rye-here"); err == nil {
			content, err := os.ReadFile(".rye-here")
			if err != nil {
				log.Fatal(err)
			}
			inputH := string(content)
			block, genv := loader.LoadString(inputH, false)
			block1 := block.(env.Block)
			es = env.AddToProgramState(es, block1.Series, genv)
			evaldo.EvalBlock(es)
		} else {
			fmt.Println("There was no `here` file.")
		}
	}

	evaldo.DoRyeRepl(es, lang, evaldo.ShowResults)
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
				// cursorPos = len(line)
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
		// Read the keyboard input.
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
			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			return os.Chdir(userHomeDir)
		}
		// Change the directory and return the error.
		return os.Chdir(args[1])
	case "exit":
		os.Exit(0)
	}

	// Prepare the command to execute.
	// REMOVED 20231205
	// Subprocess launched with a potential tainted input or cmd arguments (gosec)
	// cmd := exec.Command(args[0], args[1:]...)

	//
	// look at this page on how to capture the output and pass it through:
	// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
	// we tested it and it worked, but changed ls to multiline output. This could be also used to do some output formatting like
	// adding colors. Maybe other shells use similar approaches and so this should work
	// With this, we could "move cursor in the background string and get the word under the string like we wanted to
	// but maybe we would screw up some commands ... test it like ls | wc -l could function a little differently.
	// or htop (which exited in previous case also) and less which worked with default stream handling
	// maybe we could detect these special commands and let them through as they are and maybe when we are piping
	// we wouldn't buffer output since it won't be displayed anyway
	//
	// OK. so we solved the ls mystery. It displays / outputs uniformly. So this could be done it seems.
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
	// REMOVED 20231205 -- as above
	//cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	// Execute the command and return the error.
	// return cmd.Run()
	return nil
}
