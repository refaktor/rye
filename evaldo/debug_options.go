package evaldo

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/refaktor/rye/env"
)

// OfferDebuggingOptions presents debugging options to the user when an error occurs
// Loops to offer options repeatedly until user chooses to continue or times out
func OfferDebuggingOptions(es *env.ProgramState, genv *env.Idxs, tag string) {
	// Cache file information outside the loop
	var showVSCodeOption bool
	var fileName string
	var lineNumber int

	if es.BlockFile != "" {
		fileName = es.BlockFile
		lineNumber = es.BlockLine
		showVSCodeOption = true
	}

	for {
		fmt.Print("\n\x1b[1;36mDebugging options:\x1b[0m\n")
		fmt.Println("  \x1b[1;32me\x1b[0m) Enter REPL console (\x1b[33menter-console\x1b[0m)")
		fmt.Println("  \x1b[1;32ml\x1b[0m) List current context (\x1b[33mlc\x1b[0m)")

		if showVSCodeOption {
			fmt.Printf("  \x1b[1;32mn\x1b[0m) Open in Neovim (\x1b[33mnvim +%d %s\x1b[0m)\n", lineNumber, fileName)
			fmt.Printf("  \x1b[1;32mc\x1b[0m) Open in VSCode (\x1b[33mcode -g %s:%d\x1b[0m)\n", fileName, lineNumber)
		}

		fmt.Println("  \x1b[1;32mx\x1b[0m) Exit program")
		fmt.Print("\n\x1b[1;36mChoose option [default x")
		if showVSCodeOption {
			fmt.Print("o")
		} else {
			fmt.Print("l")
		}
		fmt.Print("] (7 second timeout):\x1b[0m ")

		// Create channels for input and timeout
		inputCh := make(chan string, 1)
		timeoutCh := make(chan bool, 1)

		// Start goroutine to read input
		go func() {
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				inputCh <- ""
			} else {
				inputCh <- strings.TrimSpace(input)
			}
		}()

		// Start timeout timer
		go func() {
			time.Sleep(7 * time.Second)
			timeoutCh <- true
		}()

		// Wait for either input or timeout
		select {
		case input := <-inputCh:
			// User provided input within timeout
			switch input {
			case "e":
				HandleEnterConsoleOption(es, genv)
				// Continue the loop to offer options again
			case "l":
				HandleListContextOption(es, genv)
				// Continue the loop to offer options again
			case "c":
				if showVSCodeOption {
					HandleVSCodeOption(fileName, lineNumber)
					// Continue the loop to offer options again
				} else {
					fmt.Println("\x1b[31mInvalid option. VSCode option not available.\x1b[0m")
					// Continue the loop to offer options again
				}
			case "n":
				if showVSCodeOption {
					HandleNvimOption(fileName, lineNumber)
					// Continue the loop to offer options again
				} else {
					fmt.Println("\x1b[31mInvalid option. Neovim option not available.\x1b[0m")
					// Continue the loop to offer options again
				}
			case "x", "":
				fmt.Println("\x1b[36mExiting ...\x1b[0m")
				return // Exit the function and continue program execution
			default:
				fmt.Printf("\x1b[31mInvalid option '%s'. Please try again.\x1b[0m\n", input)
				// Continue the loop to offer options again
			}

		case <-timeoutCh:
			// Timeout occurred - exit program
			fmt.Print("\n\x1b[33mTimeout reached. Exiting ...\x1b[0m\n")
			os.Exit(1)
		}
	}
}

// HandleEnterConsoleOption enters the REPL console for debugging
func HandleEnterConsoleOption(es *env.ProgramState, genv *env.Idxs) {
	fmt.Println("\x1b[36mEntering REPL console for debugging...\x1b[0m")
	fmt.Println("\x1b[33mType 'exit' or Ctrl+D to return to execution.\x1b[0m")

	// Clear failure and error flags before entering console for clean debugging session
	es.FailureFlag = false
	es.ErrorFlag = false

	// Use the existing REPL system
	DoRyeRepl(es, "rye", true)
}

// HandleListContextOption lists the current context
func HandleListContextOption(es *env.ProgramState, genv *env.Idxs) {
	fmt.Println("\x1b[36mCurrent context:\x1b[0m")

	// Display current context contents
	ctxMap := es.Ctx.GetState()
	if len(ctxMap) == 0 {
		fmt.Println("  \x1b[33m(empty context)\x1b[0m")
		return
	}

	fmt.Printf("\x1b[1;36m  Context contains %d items:\x1b[0m\n", len(ctxMap))

	// Sort keys for consistent display
	keys := make([]string, 0, len(ctxMap))
	for idx := range ctxMap {
		keys = append(keys, genv.GetWord(idx))
	}

	// Simple bubble sort for the keys
	for i := 0; i < len(keys)-1; i++ {
		for j := 0; j < len(keys)-i-1; j++ {
			if keys[j] > keys[j+1] {
				keys[i], keys[j+1] = keys[j+1], keys[j]
			}
		}
	}

	for _, key := range keys {
		idx := genv.IndexWord(key)
		if value, exists := es.Ctx.Get(idx); exists {
			valueStr := value.Inspect(*genv)
			if len(valueStr) > 60 {
				valueStr = valueStr[:57] + "..."
			}
			fmt.Printf("    \x1b[1;32m%s\x1b[0m: %s\n", key, valueStr)
		}
	}

	// Also show parent context if it exists
	if es.Ctx.Parent != nil {
		fmt.Println("\x1b[33m  (parent context available)\x1b[0m")
	}
}

// HandleVSCodeOption opens VSCode at the specified file and line
func HandleVSCodeOption(fileName string, lineNumber int) {
	target := fmt.Sprintf("%s:%d", fileName, lineNumber)
	fmt.Printf("\x1b[36mOpening VSCode at %s...\x1b[0m\n", target)

	cmd := exec.Command("code", "-g", target)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("\x1b[31mError opening VSCode: %v\x1b[0m\n", err)
		fmt.Println("\x1b[33mMake sure VSCode is installed and the 'code' command is in your PATH.\x1b[0m")
	} else {
		fmt.Println("\x1b[32mVSCode opened successfully.\x1b[0m")
	}
}

// HandleNvimOption opens Neovim at the specified file and line
func HandleNvimOption(fileName string, lineNumber int) {
	lineArg := fmt.Sprintf("+%d", lineNumber)
	fmt.Printf("\x1b[36mOpening Neovim at %s:%d...\x1b[0m\n", fileName, lineNumber)

	// Try nvim in a new terminal (most common approach)
	cmd := exec.Command("gnome-terminal", "--", "nvim", lineArg, fileName)
	err := cmd.Start()
	if err == nil {
		fmt.Println("\x1b[32mNeovim opened successfully in new terminal.\x1b[0m")
		return
	}

	// Try with xterm if gnome-terminal is not available
	cmd = exec.Command("xterm", "-e", "nvim", lineArg, fileName)
	err = cmd.Start()
	if err == nil {
		fmt.Println("\x1b[32mNeovim opened successfully in xterm.\x1b[0m")
		return
	}

	// Try with konsole (KDE terminal)
	cmd = exec.Command("konsole", "-e", "nvim", lineArg, fileName)
	err = cmd.Start()
	if err == nil {
		fmt.Println("\x1b[32mNeovim opened successfully in konsole.\x1b[0m")
		return
	}

	// If all terminal attempts failed, show manual instruction
	fmt.Printf("\x1b[31mCould not automatically open Neovim in a terminal.\x1b[0m\n")
	fmt.Printf("\x1b[33mTo open manually, run: \x1b[1mnvim %s %s\x1b[0m\n", lineArg, fileName)
	fmt.Println("\x1b[33mOr open a terminal and run the command above.\x1b[0m")
}
