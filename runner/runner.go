//go:build !wasm

package runner

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/refaktor/rye/contrib"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
	"github.com/refaktor/rye/security"
	"github.com/refaktor/rye/util"
)

var (
	// fileName = flag.String("fiimle", "", "Path to the Rye file (default: none)")
	do     = flag.String("do", "", "Evaluates code after it loads a file or last save.")
	sdo    = flag.String("sdo", "", "Evaluates code after it loads a file or last save.")
	lang   = flag.String("lang", "rye", "Select a dialect / language (do, eyr, ...)")
	ctx    = flag.String("ctx", "", "Enter context or context chain")
	silent = flag.Bool("silent", false, "Console doesn't display return values")
	stin   = flag.String("stin", "no", "Inject first value from stdin")
	//	quit    = flag.Bool("quit", false, "Quits after executing.")
	console  = flag.Bool("console", false, "Enters console after a file is evaluated.")
	dual     = flag.Bool("dual", false, "Starts REPL in dual-mode with two parallel panels")
	template = flag.Bool("template", false, "Process file as a template, evaluating Rye code in {{ }} blocks")
	help     = flag.Bool("help", false, "Displays this help message.")

	// Seccomp options (Linux only) - using pure Go library
	SeccompProfile = flag.String("seccomp-profile", "", "Seccomp profile to use: strict, readonly")
	SeccompAction  = flag.String("seccomp-action", "errno", "Action on restricted syscalls: errno, kill, trap, log")

	// Landlock options (Linux only) - using landlock-go library
	LandlockEnabled = flag.Bool("landlock", false, "Enable landlock filesystem access control")
	LandlockProfile = flag.String("landlock-profile", "readonly", "Landlock profile: readonly, readexec, custom")
	LandlockPaths   = flag.String("landlock-paths", "", "Comma-separated list of paths to allow access to (for custom profile)")

	// Code signing options
	CodeSigEnforced = flag.Bool("codesig", false, "Enforce code signature verification")
)

// CurrentScriptDirectory stores the directory of the currently executing script
var CurrentScriptDirectory string

// GetScriptDirectory returns the directory of the currently executing script
func GetScriptDirectory() string {
	return CurrentScriptDirectory
}

// Error handling utilities
var (
	errorLogFile *os.File
	errorLogger  *log.Logger
	logErrors    bool = true
)

// initErrorLogging initializes the error logging system
func initErrorLogging() {
	// Try to open error log file
	var err error
	errorLogFile, err = os.OpenFile("rye_errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not open error log file: %v\n", err)
		errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// Create multi-writer to log to both stderr and file
		multiWriter := io.MultiWriter(os.Stderr, errorLogFile)
		errorLogger = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// logError logs an error with context if logging is enabled
func logError(err error, context string) {
	if logErrors && err != nil {
		errorLogger.Printf("%s: %v", context, err)
	}
}

// handleError handles an error with the specified context
// If fatal is true, the program will exit
func handleError(err error, context string, fatal bool) {
	if err != nil {
		logError(err, context)
		errMsg := fmt.Sprintf("Error in %s: %v", context, err)
		if fatal {
			fmt.Fprintln(os.Stderr, errMsg)
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, errMsg)
		}
	}
}

func DoMain(regfn func(*env.ProgramState) error) {
	// Initialize error logging
	initErrorLogging()
	defer func() {
		if errorLogFile != nil {
			errorLogFile.Close()
		}
	}()

	// Add panic recovery
	// defer func() {
	//	if r := recover(); r != nil {
	//		fmt.Fprintf(os.Stderr, "Recovered from panic in DoMain: %v\n", r)
	//		debug.PrintStack()
	//	}
	// }()

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
		fmt.Println("  cont[inue]\n     Continue console from the last save")
		fmt.Println("  here\n     Starts in Rye here mode (wip)")
		fmt.Println(" \033[1mExamples:\033[0m")
		fmt.Println("\033[33m  rye                                  \033[36m# enters console/REPL")
		fmt.Println("\033[33m  rye -do \"print 33 * 42\"              \033[36m# evaluates the do code")
		fmt.Println("\033[33m  rye -do 'name: \"Jim\"' console        \033[36m# evaluates the do code and enters console")
		fmt.Println("\033[33m  rye continue                             \033[36m# continues/loads last saved state and enters console")
		fmt.Println("\033[33m  rye -do 'print 10 + 10' cont         \033[36m# continues/loads last saved state, evaluates do code and enters console")
		fmt.Println("\033[33m  rye filename.rye                     \033[36m# evaluates filename.rye")
		fmt.Println("\033[33m  rye .                                \033[36m# evaluates main.rye in current directory")
		fmt.Println("\033[33m  rye some/path/.                      \033[36m# evaluates main.rye in some/path/")
		fmt.Println("\033[33m  rye -do 'print \"Hello\" path/.        \033[36m# evaluates main.rye in path/ and then do code")
		fmt.Println("\033[33m  rye -console file.rye                \033[36m# evaluates file.rye and enters console")
		fmt.Println("\033[33m  rye -do 'print 123' -console .       \033[36m# evaluates main.rye in current dir. evaluates do code and enters console")
		fmt.Println("\033[33m  rye -silent                          \033[36m# enters console in that doesn't show return values - silent mode")
		fmt.Println("\033[33m  rye -silent -console file.rye        \033[36m# evaluates file.re and enters console in silent mode")
		fmt.Println("\033[33m  rye -lang eyr                        \033[36m# enter console of stack based Eyr language")
		fmt.Println("\033[33m  rye -lang math                       \033[36m# enter console of math dialect")
		fmt.Println("\033[33m  rye -ctx os                          \033[36m# enter console and enter os context")
		fmt.Println("\033[33m  rye -ctx 'os pipes'                  \033[36m# enter console and enter os and then pipes context")
		fmt.Println("\033[33m  rye -template template.txt           \033[36m# processes template.txt, evaluating Rye code in {{ }} blocks")
		fmt.Println("\033[33m  rye                                  \033[36m# seccomp and landlock are disabled by default")
		fmt.Println("\033[33m  rye -seccomp-profile=strict          \033[36m# enable seccomp with the strict profile")
		fmt.Println("\033[33m  rye -seccomp-profile=readonly        \033[36m# enable seccomp with the readonly profile (blocks write operations)")
		fmt.Println("\033[33m  rye -seccomp-action=kill             \033[36m# terminate process on restricted syscalls")
		fmt.Println("\033[33m  rye -landlock                        \033[36m# enable landlock filesystem access control")
		fmt.Println("\033[33m  rye -landlock-profile=readonly       \033[36m# use the readonly profile (default)")
		fmt.Println("\033[33m  rye -landlock-profile=readexec       \033[36m# use the readexec profile (allows execution)")
		fmt.Println("\033[33m  rye -landlock-paths=/path1,/path2    \033[36m# specify paths to allow access to")
		fmt.Println("\033[0m\n Thank you for trying out \033[1mRye\033[22m ...")
		fmt.Println("")
	}
	// Parse flags
	flag.Parse()

	evaldo.ShowResults = !*silent

	doCode := ""

	if *do != "" {
		doCode = *do
	}

	if *sdo != "" {
		doCode = *sdo
		evaldo.ShowResults = false
	}

	ctxCode := ""
	if *ctx != "" {
		ctxs_ := strings.Split(*ctx, " ")
		for _, v := range ctxs_ {
			ctxCode = ctxCode + " cc " + v + " "
		}
	}

	code := ctxCode + " " + doCode

	if Option_Embed_Main {
		main_rye_file("buildtemp/main.rye", false, true, false, *console, code, *lang, regfn, *stin)
	} else {
		// Check for --help flag
		if flag.NFlag() == 0 && flag.NArg() == 0 {
			if Option_Embed_Main {
				fmt.Println("CASE OPT EMBED MAIN 2")
				main_rye_file("buildtemp/main.rye", false, true, false, *console, code, *lang, regfn, *stin)
			} else if Option_Do_Main {
				ryeFile := dotsToMainRye(".")
				main_rye_file(ryeFile, false, true, false, *console, code, *lang, regfn, *stin)
			} else {
				main_rye_repl(os.Stdin, os.Stdout, true, false, *lang, code, regfn)
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
				if args[0] == "cont" || args[0] == "continue" {
					fmt.Println("[continuing...]")
					ryeFile := findLastConsoleSave()
					main_rye_file(ryeFile, false, true, false, true, code, *lang, regfn, *stin)
				} else if args[0] == "shell" {
					main_rysh()
				} else if args[0] == "rwk" {
					main_ryk()
				} else if args[0] == "here" {
					if *do != "" {
						main_rye_file("", false, true, true, *console, code, *lang, regfn, *stin)
					} else {
						main_rye_repl(os.Stdin, os.Stdout, true, true, *lang, code, regfn)
					}
				} else if *template {
					processTemplate(args[0], regfn)
				} else {
					ryeFile := dotsToMainRye(args[0])
					main_rye_file(ryeFile, false, true, false, *console, code, *lang, regfn, *stin)
				}
			} else {
				if *do != "" || *sdo != "" {
					main_rye_file("", false, true, false, *console, code, *lang, regfn, *stin)
				} else {
					main_rye_repl(os.Stdin, os.Stdout, true, false, *lang, code, regfn)
				}
			}
		}
	}
}

func findLastConsoleSave() string {
	// Read directory entries
	entries, err := os.ReadDir(".")
	if err != nil {
		handleError(err, "reading directory for console saves", false)
		return ""
	}

	files := make([]string, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}
		if strings.HasPrefix(entry.Name(), "console_") {
			// Verify file is readable
			if _, err := os.Stat(entry.Name()); err != nil {
				handleError(err, fmt.Sprintf("checking console save file %s", entry.Name()), false)
				continue // Skip this file but continue processing others
			}
			files = append(files, entry.Name())
		}
	}

	if len(files) == 0 {
		fmt.Println("No console save files found")
		return ""
	}

	sort.Strings(files)
	latestFile := files[len(files)-1]

	// Final verification
	if _, err := os.ReadFile(latestFile); err != nil {
		handleError(err, fmt.Sprintf("reading latest console save file %s", latestFile), false)
		// Try the next most recent file if available
		if len(files) > 1 {
			fmt.Printf("Trying previous save file: %s\n", files[len(files)-2])
			return files[len(files)-2]
		}
		return ""
	}

	return latestFile
}

func dotsToMainRye(ryeFile string) string {
	re, err := regexp.Compile(`^\.$|/\.$`)
	if err != nil {
		handleError(err, "compiling regex in dotsToMainRye", false)
		return ryeFile
	}

	if re.MatchString(ryeFile) {
		main_path := ryeFile[:len(ryeFile)-1] + "main.rye"
		if _, err := os.Stat(main_path); err == nil || Option_Embed_Main {
			_, err := os.ReadFile(main_path)
			if err != nil {
				handleError(err, fmt.Sprintf("reading main.rye at %s", main_path), false)
				return ryeFile // Return original file on error
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
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered from panic in main_ryk: %v\n", r)
			debug.PrintStack()
		}
	}()

	argIdx := 2
	ignore := 0
	separator := ""
	input := " 1 "

	// 	fmt.Print("preload")

	profile_path := ".ryk-preload"

	if _, err := os.Stat(profile_path); err == nil {
		content, err := os.ReadFile(profile_path)
		if err != nil {
			handleError(err, fmt.Sprintf("reading profile file %s", profile_path), false)
			// Continue with default input instead of fatal error
		} else {
			input = string(content)
		}
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
		if os.Args[argIdx] == "--ssp" {
			separator = " "
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
			evaldo.MaybeDisplayFailureOrError(es, es.Idx, "rwk begin")
			es.Ser.Reset()
			argIdx += 2
		}
		if os.Args[argIdx] == "--filter" {
			code := os.Args[argIdx+1]
			if code[0] == '/' {
				// MustCompilePOSIX panics on error, so no error handling needed
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
				var val0 env.Object
				if separator == "" {
					val := *env.NewString(scanner.Text())
					es.Ctx.Set(L, *env.NewInteger(int64(len(val.Value))))
					val0 = val
				} else {
					val := util.StringToFieldsWithQuoted(scanner.Text(), separator, "\"")
					es.Ctx.Set(L, *env.NewInteger(int64(val.Series.Len())))
					val0 = val
				}
				// if er == nil {
				es.Ctx.Set(N, *env.NewInteger(int64(nn)))
				if filterBlock != nil {
					blk := *filterBlock
					es.Ser = blk.(env.Block).Series
					evaldo.EvalBlockInj(es, val0, true)
					evaldo.MaybeDisplayFailureOrError(es, es.Idx, "rwk main filter")
					es.Ser.Reset()
					doLine = util.IsTruthy(es.Res)
				}
				if doLine {
					es.Ser = block1.(env.Block).Series
					evaldo.EvalBlockInj(es, val0, true)
					evaldo.MaybeDisplayFailureOrError(es, es.Idx, "rwk main doline")
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
		handleError(err, "scanning input in main_ryk", false)
		// Consider additional recovery actions if needed
	}

	argIdx += 1

	if len(os.Args) >= argIdx+2 {
		if os.Args[argIdx] == "--end" {
			block, genv := loader.LoadString(os.Args[argIdx+1], false)
			es = env.AddToProgramState(es, block.(env.Block).Series, genv)
			evaldo.EvalBlockInj(es, es.ForcedResult, true)
			evaldo.MaybeDisplayFailureOrError(es, es.Idx, "rwk end")
			es.Ser.Reset()
		}
	}
}

func main_ryeco() {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered from panic in main_ryeco: %v\n", r)
			debug.PrintStack()
		}
	}()

	// this is experiment to create a golang equivalent of rye code
	// with same datatypes and using the same builtin code
	// so it gets compiled, so we can see what speeds do we get that way
	// defer profile.Start().Stop()
	//input := "{ loop 10000000 { add 1 2 } }"

	// so we need a golang loop and add rye function versions

	// ryeco_do(func() env.Object { return ryeco_loop(1000, func() env.Object { return ryeco_add(1, 2) }) })

	// ryeco.Loop(env.Integer{10000000}, func() env.Object { return ryeco.Inc(env.Integer{1}) })

}

func main_rye_file(file string, sig bool, subc bool, here bool, interactive bool, code string, lang string, regfn func(*env.ProgramState) error, stin string) {
	// Add defer to recover from panics
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered from panic in main_rye_file: %v\n", r)
			// Log stack trace
			debug.PrintStack()
		}
	}()

	logFile, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Override sig parameter with CodeSigEn flag if it's set
	if *CodeSigEnforced {
		sig = true
	}

	// Store the script directory for code signing auto-enforcement
	if file != "" {
		scriptDir := filepath.Dir(file)
		if scriptDir == "." {
			// Get absolute path if it's a relative path
			if absPath, err := filepath.Abs(scriptDir); err == nil {
				scriptDir = absPath
			}
		}
		CurrentScriptDirectory = scriptDir
	}

	// fmt.Println("RYE FILE")
	info := true

	//defer profile.Start(profile.CPUProfile).Stop()

	var content string

	if len(file) > 4 && file[len(file)-4:] == ".enc" {
		fmt.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			handleError(err, "reading password", true)
			return
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
			handleError(err, fmt.Sprintf("reading file %s", file), true)
			return
		}
		content = string(bcontent)
	} else {
		content = ""
	}

	if info {
		pattern, err := regexp.Compile(`^; (#[^\n]*)`)
		if err != nil {
			handleError(err, "compiling info pattern regex", false)
		} else {
			lines := pattern.FindAllStringSubmatch(content, -1)
			for _, line := range lines {
				if line[1] != "" {
					fmt.Println(line[1])
				}
			}
		}
	}

	// READ STDIN IF

	var stValue env.Object
	stValue = *env.NewString("")

	if stin == "all" || stin == "a" { // TODO add modes like lines, maybe load / lines, do / lines)
		var stInput string
		stReader := bufio.NewReader(os.Stdin)
		for {
			stLine, err := stReader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					handleError(err, "reading from stdin", false)
				}
				break
			}
			stInput += stLine
		}
		stValue = *env.NewString(stInput)
	}

	ps := env.NewProgramStateNEW()
	ps.Embedded = Option_Embed_Main
	ps.ScriptPath = file

	workingPath, err := os.Getwd()
	if err != nil {
		handleError(err, "getting working directory", false)
		workingPath = "." // Use current directory as fallback
	}
	ps.WorkingPath = workingPath

	evaldo.RegisterBuiltins(ps)
	evaldo.RegisterVarBuiltins(ps)
	contrib.RegisterBuiltins(ps, &evaldo.BuiltinNames)
	if err := regfn(ps); err != nil {
		fmt.Println(err.Error())
		return
	}

	if here {
		if _, err := os.Stat(".rye-here"); err == nil {
			content, err := os.ReadFile(".rye-here")
			if err != nil {
				handleError(err, "reading .rye-here file", false)
				fmt.Println("Could not read .rye-here file")
			} else {
				inputH := string(content)
				block := loader.LoadStringNEW(inputH, security.CurrentCodeSigEnabled, ps)
				block1 := block.(env.Block)
				ps = env.AddToProgramState(ps, block1.Series, ps.Idx)
				evaldo.EvalBlockInjMultiDialect(ps, nil, false)
				evaldo.MaybeDisplayFailureOrError(ps, ps.Idx, "main rye file")
			}
		} else {
			fmt.Println("There was no `here` file.")
		}
	}

	// current.RegisterBuiltins(ps)
	// ctx := ps.Ctx
	// ps.Ctx = env.NewEnv(ctx)
	//ES = ps
	// evaldo.ShowResults = false

	block := loader.LoadStringNEW(" "+content+"\n"+code, security.CurrentCodeSigEnabled, ps)
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

		if lang == "eyr" {
			// fmt.Println("****")
			ps.Dialect = env.EyrDialect
		}

		evaldo.EvalBlockInjMultiDialect(ps, stValue, true)
		evaldo.MaybeDisplayFailureOrError(ps, ps.Idx, "main rye file #2")

		if interactive {
			evaldo.DoRyeRepl(ps, "rye", evaldo.ShowResults)
		} else {
			if file == "" && evaldo.ShowResults { // TODO -- move this to some instance ... to ProgramState? or is more of a ReplState?
				fmt.Println(ps.Res.Print(*ps.Idx))
			}
		}

	case env.Error:
		fmt.Println(util.TermError(val.Message))
	}
}

func main_cgi_file(file string, sig bool) {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered from panic in main_cgi_file: %v\n", r)
			debug.PrintStack()
		}
	}()

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
			handleError(err, fmt.Sprintf("reading CGI file %s", file), true)
			fmt.Fprintf(w, "Error reading CGI file: %v", err)
			return
		}

		content := string(bcontent)

		block, genv = loader.LoadString(content, sig)
		switch val := block.(type) {
		case env.Block:
			es = env.AddToProgramState(es, block.(env.Block).Series, genv)
			evaldo.RegisterBuiltins(es)
			contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

			evaldo.EvalBlock(es)
			evaldo.MaybeDisplayFailureOrError(es, genv, "main cgi file")
		case env.Error:
			fmt.Fprintf(w, "Error: %s", val.Message)
		}
	})); err != nil {
		handleError(err, "serving CGI", false)
	}
}

func main_rye_repl(_ io.Reader, _ io.Writer, subc bool, here bool, lang string, code string, regfn func(*env.ProgramState) error) {
	// Add panic recovery
	//defer func() {
	//	if r := recover(); r != nil {
	//		fmt.Fprintf(os.Stderr, "Recovered from panic in main_rye_repl: %v\n", r)
	//		debug.PrintStack()
	//	}
	//}()

	logFile, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// fmt.Println("RYE REPL")
	input := code // "name: \"Rye\" version: \"0.011 alpha\""
	// userHomeDir, _ := os.UserHomeDir()
	// profile_path := filepath.Join(userHomeDir, ".rye-profile")

	if !*dual {
		// fmt.Println("Welcome to Rye console. Use lc to list current or lcp and lcp\\ \"pri\" to list parent contexts.")
		fmt.Println("Welcome to Rye console. We're still W-I-P. Visit \033[38;5;14mryelang.org\033[0m for more info.")
		fmt.Println("- \033[38;5;246mtype in lcp (list context parent) too see functions, or lc to see your context\033[0m")
		//fmt.Println("--------------------------------------------------------------------------------")
	}

	// Uncomment and fix the profile loading code if needed
	//if _, err := os.Stat(profile_path); err == nil {
	//	content, err := os.ReadFile(profile_path)
	//	if err != nil {
	//		handleError(err, "reading profile file", false)
	//	} else {
	//		input = string(content)
	//	}
	//} else {
	//	fmt.Println("There was no profile.")
	//}

	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	evaldo.RegisterVarBuiltins(es)
	contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)
	if err := regfn(es); err != nil {
		fmt.Println(err.Error())
		return
	}

	if lang == "eyr" {
		es.Dialect = env.EyrDialect
	}
	evaldo.EvalBlockInjMultiDialect(es, nil, false)

	if subc {
		ctx := es.Ctx
		es.Ctx = env.NewEnv(ctx) // make new context with no parent
	}

	if here {
		if _, err := os.Stat(".rye-here"); err == nil {
			content, err := os.ReadFile(".rye-here")
			if err != nil {
				handleError(err, "reading .rye-here file", false)
				fmt.Println("Could not read .rye-here file")
			} else {
				inputH := string(content)
				block, genv := loader.LoadString(inputH, false)
				if blockErr, ok := block.(env.Error); ok {
					handleError(fmt.Errorf("%s", blockErr.Message), "parsing .rye-here file", false)
				} else {
					block1 := block.(env.Block)
					es = env.AddToProgramState(es, block1.Series, genv)
					evaldo.EvalBlock(es)
				}
			}
		} else {
			fmt.Println("There was no `here` file.")
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(c) // Ensure signal resources are released

	go func() {
		sig := <-c
		fmt.Println()
		fmt.Println("Captured signal:", sig)
		// Perform cleanup or other actions here
		// os.Exit(0)
	}()
	//fmt.Println("Waiting for signal")

	if *dual {
		// Create a second program state for the right panel
		rightEs := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(rightEs)
		contrib.RegisterBuiltins(rightEs, &evaldo.BuiltinNames)
		if err := regfn(rightEs); err != nil {
			fmt.Println(err.Error())
			return
		}

		if lang == "eyr" {
			rightEs.Dialect = env.EyrDialect
		}
		evaldo.EvalBlockInjMultiDialect(rightEs, nil, false)

		if subc {
			ctx := rightEs.Ctx
			rightEs.Ctx = env.NewEnv(ctx) // make new context with no parent
		}

		// Start dual REPL
		evaldo.DoRyeDualRepl(es, rightEs, lang, evaldo.ShowResults)
	} else {
		evaldo.DoRyeRepl(es, lang, evaldo.ShowResults)
	}
}

func main_rysh() {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered from panic in main_rysh: %v\n", r)
			debug.PrintStack()
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	status := 1

	for status != 0 {
		// C.enableRawMode()
		wd, err := os.Getwd()
		if err != nil {
			handleError(err, "getting working directory in shell", false)
			wd = "unknown_dir" // Fallback
		}
		fmt.Print("\033[36m" + wd + " -> " + "\033[m")

		line, cursorPos, shellEditor := "", 0, false

		for {
			c, err := reader.ReadByte()
			if err != nil {
				handleError(err, "reading input in shell", false)
				break
			}
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
				c1, err := reader.ReadByte()
				if err != nil {
					handleError(err, "reading escape sequence in shell", false)
					break
				}
				if c1 == '[' {
					c2, err := reader.ReadByte()
					if err != nil {
						handleError(err, "reading escape sequence in shell", false)
						break
					}
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

// processTemplate reads a template file and processes it by evaluating Rye code in {{ }} blocks
func processTemplate(file string, regfn func(*env.ProgramState) error) {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered from panic in processTemplate: %v\n", r)
			debug.PrintStack()
		}
	}()

	// Read the template file
	content, err := os.ReadFile(file)
	if err != nil {
		handleError(err, fmt.Sprintf("reading template file %s", file), true)
		return
	}

	// Create a program state for evaluating Rye code
	ps := env.NewProgramStateNEW()
	ps.ScriptPath = file

	workingPath, err := os.Getwd()
	if err != nil {
		handleError(err, "getting working directory", false)
		workingPath = "." // Use current directory as fallback
	}
	ps.WorkingPath = workingPath

	// Register builtins
	evaldo.RegisterBuiltins(ps)
	evaldo.RegisterVarBuiltins(ps)
	contrib.RegisterBuiltins(ps, &evaldo.BuiltinNames)
	if err := regfn(ps); err != nil {
		fmt.Println(err.Error())
		return
	}

	// Regular expression to find {{ ... }} blocks (with (?s) flag to match across multiple lines)
	re := regexp.MustCompile(`(?s)\{\{\s*(.*?)\s*\}\}`)

	// Process the template
	result := re.ReplaceAllStringFunc(string(content), func(match string) string {
		// Extract the Rye code from the match
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match // Return the original match if no submatch found
		}

		ryeCode := submatch[1]

		// Create a block to evaluate
		block := loader.LoadStringNEW(ryeCode, false, ps)

		// Check for errors in the code
		if blockErr, ok := block.(env.Error); ok {
			fmt.Fprintf(os.Stderr, "Error in template code %s: %s\n", ryeCode, blockErr.Message)
			return fmt.Sprintf("[ERROR: %s]", blockErr.Message)
		}

		// Set up for capturing stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Create a channel to receive the captured output
		outC := make(chan string)

		// Copy the output in a separate goroutine
		go func() {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error capturing output: %v\n", err)
			}
			outC <- buf.String()
		}()

		// Evaluate the block
		switch val := block.(type) {
		case env.Block:
			ser := ps.Ser
			ps.Ser = val.Series
			evaldo.EvalBlock(ps)
			ps.Ser = ser
		}

		// Restore stdout and get the captured output
		w.Close()
		os.Stdout = old
		out := <-outC

		// If there was an error during evaluation, return an error message
		if ps.ErrorFlag {
			return fmt.Sprintf("[ERROR: %s]", ps.Res.Print(*ps.Idx))
		}

		// Return the captured output (without trailing newline if present)
		return strings.TrimSuffix(out, "\n")
	})

	// Print the processed template
	fmt.Print(result)
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
				handleError(err, "getting user home directory", false)
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			if err := os.Chdir(userHomeDir); err != nil {
				handleError(err, fmt.Sprintf("changing directory to %s", userHomeDir), false)
				return fmt.Errorf("failed to change directory to %s: %w", userHomeDir, err)
			}
			return nil
		}
		// Change the directory and return the error.
		if err := os.Chdir(args[1]); err != nil {
			handleError(err, fmt.Sprintf("changing directory to %s", args[1]), false)
			return fmt.Errorf("failed to change directory to %s: %w", args[1], err)
		}
		return nil
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
