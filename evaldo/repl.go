//go:build !b_norepl && !wasm && !js

package evaldo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	// "github.com/eiannone/keyboard"
	"github.com/refaktor/keyboard"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
	"github.com/refaktor/rye/term"
)

var (
	history_fn = filepath.Join(os.TempDir(), ".rye_console_history")
	names      = []string{"add", "join", "return", "fn", "fail", "if"}
)

type ShellEd struct {
	CurrObj env.Function
	Pause   bool
	Askfor  []string
	Mode    string
	Return  env.Object
}

func genPrompt(shellEd *ShellEd, line string, multiline bool) (string, string) {
	if shellEd != nil && shellEd.Mode != "" {
		a := shellEd.Askfor
		if len(a) > 0 {
			x := a[0]
			a = a[1:]
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
			if multiline {
				return " > ", ""
			} else {
				return "×> ", ""
			}
		}
	}
}

func maybeDoShedCommands(line string, es *env.ProgramState, shellEd *ShellEd) {
	//fmt.Println(shellEd)
	line1 := strings.Split(line, " ")
	block := shellEd.CurrObj.Body
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
	//if block != nil {
	//block.Trace("TOP")
	//}
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

// terminal commands from buger/goterm

// Clear screen
func Clear() {
	fmt.Print("\033[2J")
}

// Move cursor to given position
func MoveCursor(x int, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

// Move cursor up relative the current position
func MoveCursorUp(bias int) {
	fmt.Printf("\033[%dA", bias)
}

// Move cursor down relative the current position
func MoveCursorDown(bias int) {
	fmt.Printf("\033[%dB", bias)
}

// Move cursor forward relative the current position
func MoveCursorForward(bias int) {
	fmt.Printf("\033[%dC", bias)
}

// Move cursor backward relative the current position
func MoveCursorBackward(bias int) {
	fmt.Printf("\033[%dD", bias)
}

type Repl struct {
	ps *env.ProgramState
	ml *term.MLState

	dialect     string
	showResults bool

	fullCode string

	stack         *env.EyrStack // part of PS no ... move there, remove here
	prevResult    env.Object
	captureStdout bool
}

func (r *Repl) recieveMessage(message string) {
	fmt.Print(message)
}

func (r *Repl) recieveLine(line string) string {
	log.Println("RECV LINE: " + line)
	res := r.evalLine(r.ps, line)
	if r.showResults && len(res) > 0 {
		fmt.Println(res)
	}
	return res
}

func (r *Repl) evalLine(es *env.ProgramState, code string) string {
	if es.LiveObj != nil {
		es.LiveObj.PsMutex.Lock()
		for _, update := range es.LiveObj.Updates {
			fmt.Println("\033[35m((Reloading " + update + "))\033[0m")
			block_, script_ := LoadScriptLocalFile(es, *env.NewUri1(es.Idx, "file://"+update))
			es.Res = EvaluateLoadedValue(es, block_, script_, true)
		}
		es.LiveObj.ClearUpdates()
		es.LiveObj.PsMutex.Unlock()
	}

	// More robust multiline input detection
	// Check for explicit multiline indicators or incomplete syntax
	multiline := false

	// 1. Check for explicit continuation character at the end (backslash)
	if len(code) > 0 && strings.HasSuffix(strings.TrimSpace(code), "\\") {
		multiline = true
		// Remove the continuation character for processing
		code = strings.TrimSuffix(strings.TrimSpace(code), "\\")
	}

	// 2. Check for unbalanced brackets/braces/parentheses
	if !multiline {
		openBraces := 0
		openBrackets := 0
		openParens := 0
		inString := false
		stringChar := ' '

		for _, char := range code {
			// Handle string literals to avoid counting brackets inside strings
			if (char == '"' || char == '`') && (stringChar == ' ' || stringChar == char) {
				if !inString {
					inString = true
					stringChar = char
				} else {
					inString = false
					stringChar = ' '
				}
				continue
			}

			if !inString {
				switch char {
				case '{':
					openBraces++
				case '}':
					openBraces--
				case '[':
					openBrackets++
				case ']':
					openBrackets--
				case '(':
					openParens++
				case ')':
					openParens--
				}
			}
		}

		// If any delimiters are unbalanced, consider it multiline
		multiline = openBraces > 0 || openBrackets > 0 || openParens > 0
	}

	// 3. Check for incomplete block definitions that end with a colon
	if !multiline && strings.HasSuffix(strings.TrimSpace(code), ":") {
		multiline = true
	}

	comment := regexp.MustCompile(`\s*;`)
	line := comment.Split(code, 2) //--- just very temporary solution for some comments in repl. Later should probably be part of loader ... maybe?
	lineReal := strings.Trim(line[0], "\t")

	output := ""
	if multiline {
		r.fullCode += lineReal + "\n"
	} else {
		r.fullCode += lineReal

		block, genv := loader.LoadString(r.fullCode, false)

		// Check if the result is an error
		if err, isError := block.(env.Error); isError {
			fmt.Println("\033[31mParsing error: " + err.Message + "\033[0m")
			r.fullCode = ""
			return ""
		}

		block1 := block.(env.Block)
		es = env.AddToProgramState(es, block1.Series, genv)

		// STDIO CAPTURE START

		// Define variables outside if statement
		var r1 *os.File
		var w *os.File
		var err error
		var oldStdout *os.File
		var stdoutCh chan string

		if r.captureStdout {
			// Create a pipe to capture stdout
			r1, w, err = os.Pipe()
			if err != nil {
				log.Printf("Failed to create pipe: %v", err)
				// resultCh <- "Error: Failed to capture output"
				// continue
			}

			// Save the original stdout
			oldStdout = os.Stdout
			// Replace stdout with our pipe writer
			os.Stdout = w

			// Create a channel for the captured output
			stdoutCh = make(chan string)

			// Start a goroutine to read from the pipe
			go func() {
				var buf bytes.Buffer
				_, err := io.Copy(&buf, r1)
				if err != nil {
					log.Printf("Error reading from pipe: %v", err)
				}
				stdoutCh <- buf.String()
			}()
		}

		// EVAL THE DO DIALECT
		if r.dialect == "rye" {
			EvalBlockInj(es, r.prevResult, true)
		} else if r.dialect == "eyr" {
			es.Dialect = env.EyrDialect
			Eyr_EvalBlock(es, true)
		} else if r.dialect == "math" {
			idxx, _ := es.Idx.GetIndex("math")
			s1, ok := es.Ctx.Get(idxx)
			if ok {
				switch ss := s1.(type) {
				case env.RyeCtx: /*  */
					es.Ctx = &ss
					// return s1
				}
			}
			res := DialectMath(es, block1)
			switch block := res.(type) {
			case env.Block:
				//stack := env.NewEyrStack()
				es.ResetStack()
				ser := es.Ser
				es.Ser = block.Series
				Eyr_EvalBlock(es, false)
				es.Ser = ser
			}
		} else {
			fmt.Println("Unknown dialect: " + r.dialect)
		}

		MaybeDisplayFailureOrError(es, genv, "repl / eval Line")

		if !es.ErrorFlag && es.Res != nil {
			r.prevResult = es.Res
			p := ""
			if env.IsPointer(es.Res) {
				p = "Ref"
			}
			resultStr := es.Res.Inspect(*genv)
			if r.dialect == "eyr" {
				resultStr = strings.Replace(resultStr, "Block:", "Stack:", 1) // TODO --- temp / hackish way ... make stack display itself or another approach
			}
			output = fmt.Sprint("\033[38;5;37m" + p + resultStr + "\x1b[0m")
		}

		if r.captureStdout {
			// STDOUT CAPTURE
			// Close the pipe writer to signal EOF to the reader
			w.Close()

			// Restore the original stdout
			os.Stdout = oldStdout

			// Get the captured stdout
			capturedOutput := <-stdoutCh

			log.Println("CAPTURED STDOUT")
			log.Println(capturedOutput)

			// Close the pipe reader
			r1.Close()
			// STDOUT CAPTURE END
			output = capturedOutput + output
		}
		es.ReturnFlag = false
		es.ErrorFlag = false
		es.FailureFlag = false

		r.fullCode = ""
		r.ml.AppendHistory(code)
		return output
	}

	r.ml.AppendHistory(code)
	return output
}

// constructKeyEvent maps a rune and keyboard.Key to a util.KeyEvent, which uses javascript key event codes
// only keys used in microliner are mapped
func constructKeyEvent(r rune, k keyboard.Key) term.KeyEvent {
	// fmt.Println(r)
	// fmt.Println(k)
	var ctrl bool
	alt := k == keyboard.KeyEsc
	var code int
	ch := string(r)

	// Check for Ctrl modifier with Backspace
	if k == keyboard.KeyBackspace || k == keyboard.KeyBackspace2 {
		code = 8
		// Ctrl+Backspace might send a different rune (e.g., \x17) or require modifier detection
		if r == '\x17' { // ETB character, common for Ctrl+Backspace in some terminals
			fmt.Println("*")
			ctrl = true
			ch = "backspace"
		}
	}

	switch k {
	case keyboard.KeyCtrlA:
		ch = "a"
		ctrl = true
	case keyboard.KeyCtrlS:
		ch = "s"
		ctrl = true
	case keyboard.KeyCtrlC:
		ch = "c"
		ctrl = true
	case keyboard.KeyCtrlB:
		ch = "b"
		ctrl = true
	case keyboard.KeyCtrlD:
		ch = "d"
		ctrl = true
	case keyboard.KeyCtrlE:
		ch = "e"
		ctrl = true
	case keyboard.KeyCtrlF:
		ch = "f"
		ctrl = true
	case keyboard.KeyCtrlK:
		ch = "k"
		ctrl = true
	case keyboard.KeyCtrlL:
		ch = "l"
		ctrl = true
	case keyboard.KeyCtrlN:
		ch = "n"
		ctrl = true
	case keyboard.KeyCtrlP:
		ch = "p"
		ctrl = true
	case keyboard.KeyCtrlU:
		ch = "u"
		ctrl = true
	case keyboard.KeyCtrlX:
		ch = "x"
		ctrl = true

	case keyboard.KeyEnter:
		code = 13
	case keyboard.KeyTab:
		code = 9
	case keyboard.KeyBackspace:
		ch = "backspace"
		ctrl = true
		code = 8 // Consistent with plain Backspace
	case keyboard.KeyBackspace2:
		code = 8
	// case keyboard.KeyAltBackspace:
	//	ch = "backspace"
	//	alt = true
	//	code = 8 // Consistent with plain Backspace
	case keyboard.KeyDelete:
		code = 46
	case keyboard.KeyArrowRight:
		code = 39
	case keyboard.KeyArrowLeft:
		code = 37
	case keyboard.KeyArrowUp:
		code = 38
	case keyboard.KeyArrowDown:
		code = 40
	case keyboard.KeyHome:
		code = 36
	case keyboard.KeyEnd:
		code = 35
	case keyboard.KeySpace:
		ch = " "
		code = 20
		//case keyboard.KeyCtrlBackspace:
		//	ch = "backspace"
		//	ctrl = true
		//	code = 8 // Consistent with plain Backspace
	}
	return term.NewKeyEvent(ch, code, ctrl, alt, false)
}

func isCursorAtBottom() bool { // TODO --- doesn't seem to work and probably don't need it ... test and remove if doesn't work
	// Implement a more robust check for the cursor's position if needed
	// For a simple approximation, you can check if the terminal height matches the current cursor position
	return true || os.Getenv("TERM_LINES") != "" && os.Getenv("TERM_LINES") == os.Getenv("TERM_ROW")
}

func DoRyeRepl(es *env.ProgramState, dialect string, showResults bool) { // here because of some odd options we were experimentally adding
	// Improved error handling for keyboard initialization
	err := keyboard.Open()
	if err != nil {
		log.Printf("Failed to initialize keyboard: %v", err)
		return
	}
	// Ensure keyboard is closed when function exits
	defer func() {
		fmt.Println("Closing keyboard...")
		if err := keyboard.Close(); err != nil {
			log.Printf("Error closing keyboard: %v", err)
		}
	}()

	c := make(chan term.KeyEvent)
	r := Repl{
		ps:            es,
		dialect:       dialect,
		showResults:   showResults,
		stack:         env.NewEyrStack(),
		captureStdout: false,
	}
	ml := term.NewMicroLiner(c, r.recieveMessage, r.recieveLine)
	r.ml = ml

	// Improved error handling for history file operations
	f, err := os.Open(history_fn)
	if err != nil {
		log.Printf("Could not open history file: %v", err)
	} else {
		defer f.Close() // Ensure file is closed even if ReadHistory panics

		if count, err := ml.ReadHistory(f); err != nil {
			log.Printf("Error reading history file: %v", err)
		} else {
			fmt.Printf("Read %d history entries\n", count)
		}
	}

	ml.SetCompleter(func(line string, mode int) (c []string) {
		// #IMPROV #IDEA words defined in current context should be bold
		// #IMPROV #Q how would we cycle just through words in current context?
		// #TODO don't display more than N words
		// #TODO make current word bold

		// # TRICK: we don't have the cursor position, but the caller code handles that already so we can suggest in the 	middle
		suggestions := make([]string, 0)
		var wordpart string
		spacePos := strings.LastIndex(line, " ")
		var prefix string
		if spacePos < 0 {
			// fmt.Println("*")
			wordpart = line
			prefix = ""
		} else {
			wordpart = strings.TrimSpace(line[spacePos:])
			fmt.Print("=(")
			fmt.Print(wordpart)
			fmt.Print(")=")
			prefix = line[0:spacePos] + " "
			if wordpart == "" { // we are probably 1 space after last word
				fmt.Println("+")
				return
			}
		}

		fmt.Print("=[")
		fmt.Print(wordpart)
		fmt.Print("]=")
		switch mode {
		case 0:
			for i := 0; i < es.Idx.GetWordCount(); i++ {
				// fmt.Print(es.Idx.GetWord(i))
				if strings.HasPrefix(es.Idx.GetWord(i), strings.ToLower(wordpart)) {
					c = append(c, prefix+es.Idx.GetWord(i))
					suggestions = append(suggestions, es.Idx.GetWord(i))
				} else if strings.HasPrefix("."+es.Idx.GetWord(i), strings.ToLower(wordpart)) {
					c = append(c, prefix+"."+es.Idx.GetWord(i))
					suggestions = append(suggestions, es.Idx.GetWord(i))
				} else if strings.HasPrefix("|"+es.Idx.GetWord(i), strings.ToLower(wordpart)) {
					c = append(c, prefix+"|"+es.Idx.GetWord(i))
					suggestions = append(suggestions, es.Idx.GetWord(i))
				}
			}
		case 1:
			for key := range es.Ctx.GetState() {
				// fmt.Print(es.Idx.GetWord(i))
				if strings.HasPrefix(es.Idx.GetWord(key), strings.ToLower(wordpart)) {
					c = append(c, prefix+es.Idx.GetWord(key))
					suggestions = append(suggestions, es.Idx.GetWord(key))
				} else if strings.HasPrefix("."+es.Idx.GetWord(key), strings.ToLower(wordpart)) {
					c = append(c, prefix+"."+es.Idx.GetWord(key))
					suggestions = append(suggestions, es.Idx.GetWord(key))
				} else if strings.HasPrefix("|"+es.Idx.GetWord(key), strings.ToLower(wordpart)) {
					c = append(c, prefix+"|"+es.Idx.GetWord(key))
					suggestions = append(suggestions, es.Idx.GetWord(key))
				}
			}
		}

		// TODO -- make this sremlines and use local term functions
		if isCursorAtBottom() {
			// If at the bottom, print a new line to create a space
			fmt.Println()
		}

		// Move the cursor one line down
		// term.CurDown(1) //"\033[B")

		// Delete the line
		term.ClearLine() //"\033[2K")

		// Print something
		term.ColorMagenta()
		fmt.Print(suggestions)
		term.CloseProps() //	fmt.Print("This is the new line.")

		// Move the cursor back to the previous line
		term.CurUp(1) //"\033[A")

		return
	})

	// Create context with timeout to prevent potential deadlocks
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	// Improved error handling for keyboard events
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Context done, keyboard handler exiting")
				return
			default:
				r, k, keyErr := keyboard.GetKey()
				if keyErr != nil {
					log.Printf("Keyboard error: %v", keyErr)
					// Don't break on transient errors, just continue
					continue
				}

				if k == keyboard.KeyCtrlC {
					// Handle Ctrl+C properly
					fmt.Println("\nKeyboard Ctrl+C in REPL detected. Use Ctrl-D to Exit.")
					// cancel()

					// Try to kill the process gracefully
					// err := util.KillProcess(os.Getpid())
					// if err != nil {
					//	log.Printf("Error killing process: %v", err)
					// }
					//return
				}

				// Send the key event to the channel
				select {
				case c <- constructKeyEvent(r, k):
					// Key event sent successfully
				case <-ctx.Done():
					// Context was cancelled while trying to send
					return
				}
			}
		}
	}(ctx)

	// Improved error handling for saving history
	defer func() {
		f, err := os.Create(history_fn)
		if err != nil {
			log.Printf("Error creating history file: %v", err)
			return
		}
		defer f.Close() // Ensure file is closed even if WriteHistory panics

		count, err := ml.WriteHistory(f)
		if err != nil {
			log.Printf("Error writing history file: %v", err)
		} else {
			fmt.Printf("Wrote %d history entries\n", count)
		}
	}()

	// Run the REPL with improved error handling
	_, err = ml.MicroPrompt("x> ", "", 0, ctx)
	if err != nil {
		log.Printf("MicroPrompt error: %v", err)
	}

	fmt.Println("End of Function in REPL...")
}

/*  THIS WAS DISABLED TEMP FOR WASM MODE .. 20250116 func DoGeneralInput(es *env.ProgramState, prompt string) {
	line := liner.NewLiner()
	defer line.Close()
	if code, err := line.SimplePrompt(prompt); err == nil {
		es.Res = *env.NewString(code)
	} else {
		log.Print("Error reading line: ", err)
	}
}

func DoGeneralInputField(es *env.ProgramState, prompt string) {
	line := liner.NewLiner()
	defer line.Close()
	if code, err := line.SimpleTextField(prompt, 5); err == nil {
		es.Res = *env.NewString(code)
	} else {
		log.Print("Error reading line: ", err)
	}
}
*/

/* func DoRyeRepl_OLD(es *env.ProgramState, showResults bool) { // here because of some odd options we were experimentally adding
	codestr := "a: 100\nb: \"jim\"\nprint 10 + 20 + b"
	codelines := strings.Split(codestr, ",\n")

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
		if _, err := line.ReadHistory(f); err != nil {
			log.Print("Error reading history file: ", err)
		}
		f.Close()
	}
	//const PROMPT = "\x1b[6;30;42m Rye \033[m "

	shellEd := ShellEd{env.Function{}, false, make([]string, 0), "", nil}

	// nek variable bo z listo wordow bo ki jih želi setirat v tem okolju in dokler ne pride čez bo repl spraševal za njih
	// name funkcije pa bo prikazal v promptu dokler smo noter , spet en state var
	// isti potek bi lahko uporabili za kreirat live validation dialekte npr daš primer podatka Dict npr za input in potem pišeš dialekt
	// in preverjaš rezultat ... tako z hitrim reset in ponovi workflowon in prikazom rezultata
	// to s funkcijo se bo dalo čist dobro naredit ... potem pa tudi s kontekstom ne vidim kaj bi bil problem

	line2 := ""

	var prevResult env.Object

	multiline := false

	for {
		prompt, arg := genPrompt(&shellEd, line2, multiline)

		if code, err := line.Prompt(prompt); err == nil {
			// strip comment

			es.LiveObj.PsMutex.Lock()
			for _, update := range es.LiveObj.Updates {
				fmt.Println("\033[35m((Reloading " + update + "))\033[0m")
				block_, script_ := LoadScriptLocalFile(es, *env.NewUri1(es.Idx, "file://"+update))
				es.Res = EvaluateLoadedValue(es, block_, script_, true)
			}
			es.LiveObj.ClearUpdates()
			es.LiveObj.PsMutex.Unlock()

			multiline = len(code) > 1 && code[len(code)-1:] == " "

			comment := regexp.MustCompile(`\s*;`)
			line1 := comment.Split(code, 2) //--- just very temporary solution for some comments in repl. Later should probably be part of loader ... maybe?
			//fmt.Println(line1)
			lineReal := strings.Trim(line1[0], "\t")

			// fmt.Println("*" + lineReal + "*")

			// JM20201008
			if lineReal == "111" {
				for _, c := range codelines {
					fmt.Println(c)
				}
			}

			// check for #shed commands
			maybeDoShedCommands(lineReal, es, &shellEd)

			///fmt.Println(lineReal[len(lineReal)-3 : len(lineReal)])

			if multiline {
				line2 += lineReal + "\n"
			} else {
				line2 += lineReal

				if strings.Trim(line2, " \t\n\r") == "" {
					// ignore
				} else if strings.Compare("((show-results))", line2) == 0 {
					showResults = true
				} else if strings.Compare("((hide-results))", line2) == 0 {
					showResults = false
				} else if strings.Compare("((return))", line2) == 0 {
					// es.Ser = ser
					// fmt.Println("")
					return
				} else {
					//fmt.Println(lineReal)
					block, genv := loader.LoadString(line2, false)
					block1 := block.(env.Block)
					es = env.AddToProgramState(es, block1.Series, genv)

					// EVAL THE DO DIALECT
					EvalBlockInj(es, prevResult, true)

					if arg != "" {
						if arg == "<-return->" {
							shellEd.Return = es.Res
						} else {
							es.Ctx.Set(es.Idx.IndexWord(arg), es.Res)
						}
					} else {
						if shellEd.Mode != "" {
							if !shellEd.Pause {
								shellEd.CurrObj.Body.Series.AppendMul(block1.Series.GetAll())
							}
						}

						MaybeDisplayFailureOrError(es, genv)

						if !es.ErrorFlag && es.Res != nil {
							prevResult = es.Res
							// TEMP - make conditional
							// print the result
							if showResults {
								fmt.Println("\033[38;5;37m" + es.Res.Inspect(*genv) + "\x1b[0m..")
							}
							if es.Res != nil && shellEd.Mode != "" && !shellEd.Pause && es.Res == shellEd.Return {
								fmt.Println(" <- the correct value was returned")
							}
						}

						es.ReturnFlag = false
						es.ErrorFlag = false
						es.FailureFlag = false
					}
				}

				line2 = ""
			}

			line.AppendHistory(code)
		} else if err == liner.ErrPromptAborted {
			// log.Print("Aborted")
			break
			//		} else if err == liner.ErrJMCodeUp {
			/* } else if err == liner.ErrCodeUp { ... REMOVED 04.01.2022 for cleaning , figure out why I added it and if it still makes sense
			fmt.Println("")
			for _, c := range codelines {
				fmt.Println(c)
			}
			MoveCursorUp(len(codelines)) * /
		} else {
			log.Print("Error reading line: ", err)
			break
		}
	}

 	if f, err := os.Create(history_fn); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		if _, err := line.WriteHistory(f); err != nil {
			log.Print("Error writing history file: ", err)
		}
		f.Close()
	}
}
*/
