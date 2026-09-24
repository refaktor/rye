package baseio

// builtins_baseio.go - OS / file / shell / stdin / args builtins.
//
// These are kept separate from the pure-computation base builtins so that
// embedding use-cases (embed.New / RegisterBaseBuiltins) can opt-out of OS
// access entirely.  The full runner registers these via baseio.Register.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"github.com/landlock-lsm/go-landlock/landlock"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
	"github.com/refaktor/rye/security"
	"github.com/refaktor/rye/util"

	"golang.org/x/sync/errgroup"
	goterm "golang.org/x/term"
)

// builtins_baseio groups all builtins that touch the operating-system
// boundary: file I/O, shell commands, stdin, os.Exit, os.Args, and
// capture-stdout.  They are registered by baseio.Register.
var builtins_baseio = map[string]*env.Builtin{

	// -------------------------------------------------------------------------
	// Save / persist state
	// -------------------------------------------------------------------------

	"save\\current": {
		Argsn: 0,
		Doc:   "Saves current state of the program to a file.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			s := ps.Dump()
			fileName := fmt.Sprintf("console_%s.rye", time.Now().Format("060102_150405"))

			err := os.WriteFile(fileName, []byte(s), 0600)
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("error writing state: %s", err.Error()), "save\\state")
			}
			fmt.Println("State current context to \033[1m" + fileName + "\033[0m.")
			return *env.NewInteger(1)
		},
	},

	"save\\current\\secure": {
		Argsn: 0,
		Doc:   "Saves current state of the program to a file with password protection.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			s := ps.Dump()
			fileName := fmt.Sprintf("console_%s.rye.enc", time.Now().Format("060102_150405"))

			fmt.Print("Enter Password: ")
			bytePassword, err := goterm.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				panic(err)
			}
			password := string(bytePassword)

			util.SaveSecure(s, fileName, password)
			fmt.Println("State current context to \033[1m" + fileName + "\033[0m.")
			return *env.NewInteger(1)
		},
	},

	// -------------------------------------------------------------------------
	// File import / load (URI-based)
	// -------------------------------------------------------------------------

	"file-uri//Import": { // **
		Argsn: 1,
		Doc:   "Imports a file, loads and does it from script local path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				block_, script_ := evaldo.LoadScriptLocalFile(ps, s1)
				ps.Res = evaldo.EvaluateLoadedValue(ps, block_, script_, false)
				ps.ScriptPath = script_
				return ps.Res
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "import")
			}
		},
	},

	"file-uri//Import\\live": { // **
		Argsn: 1,
		Doc:   "Imports a file, loads and does it from script local path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				block_, script_ := evaldo.LoadScriptLocalFile(ps, s1)
				ps.Res = evaldo.EvaluateLoadedValue(ps, block_, script_, false)
				ps.LiveObj.Add(s1.GetPath()) // add to watcher
				ps.ScriptPath = script_
				return ps.Res
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "import\\live")
			}
		},
	},

	"file-uri//Load": { // **
		Argsn: 1,
		Doc:   "Loads a file URI into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				if s1.Scheme.Index == fileIdx {
					b, err := os.ReadFile(s1.GetPath())
					if err != nil {
						ps.FailureFlag = true
						return env.NewError(err.Error())
					}
					str = string(b)
				}
				scrip := ps.ScriptPath
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadString(str, false, ps)
				ps.ScriptPath = scrip
				return block
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Load")
			}
		},
	},

	"load\\mod\\file": { // **
		Argsn: 1,
		Doc:   "Loads a file URI into Rye values, allowing modification of words during load.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				if s1.Scheme.Index == fileIdx {
					b, err := os.ReadFile(s1.GetPath())
					if err != nil {
						ps.FailureFlag = true
						return env.NewError(err.Error())
					}
					str = string(b)
				}
				scrip := ps.ScriptPath
				ps.AllowMod = true
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadString(str, false, ps)
				ps.AllowMod = false
				ps.ScriptPath = scrip
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be a file URI")
			}
		},
	},

	"load\\live": { // **
		Argsn: 1,
		Doc:   "Loads a file URI into Rye values, allows modification of words, and watches for changes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				if s1.Scheme.Index == fileIdx {
					b, err := os.ReadFile(s1.GetPath())
					ps.LiveObj.Add(s1.GetPath()) // add to watcher
					if err != nil {
						ps.FailureFlag = true
						return env.NewError(err.Error())
					}
					str = string(b)
				}
				scrip := ps.ScriptPath
				ps.AllowMod = true
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadString(str, false, ps)
				ps.AllowMod = false
				ps.ScriptPath = scrip
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be a file URI")
			}
		},
	},

	// -------------------------------------------------------------------------
	// Process / shell
	// -------------------------------------------------------------------------

	// Tests:
	// equal { scmd "exit 0" } 0
	// equal { scmd "exit 3" } 3
	// error { scmd 123 }
	// Args:
	// * cmd: Shell command string to execute
	// Returns:
	// * integer exit status code
	"scmd": {
		Argsn: 1,
		Doc:   "Execute a shell command and return its exit status code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:
				r := exec.Command("sh", "-c", s0.Value) //nolint: gosec
				r.Stdout = os.Stdout
				r.Stderr = os.Stderr

				err := r.Run()
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						return *env.NewInteger(int64(exitError.ExitCode()))
					}
					fmt.Println(err)
					return *env.NewInteger(-1)
				}
				return *env.NewInteger(0)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "scmd")
			}
		},
	},

	// Tests:
	// equal { scmd\capture "echo hi" } "hi\n"
	// error { scmd\capture 123 }
	// Args:
	// * cmd: Shell command string to execute and capture output
	// Returns:
	// * string containing captured stdout (stderr on error)
	"scmd\\capture": {
		Argsn: 1,
		Doc:   "Execute a shell command and capture the output, return it as string",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:
				r := exec.Command("sh", "-c", s0.Value) //nolint: gosec
				var outb, errb bytes.Buffer
				r.Stdout = &outb
				r.Stderr = &errb

				err := r.Run()
				if err != nil {
					if errb.Len() > 0 {
						ps.FailureFlag = true
						return env.NewError("Command failed: " + errb.String())
					}
					if _, ok := err.(*exec.ExitError); ok && outb.Len() > 0 {
						return *env.NewString(outb.String())
					}
					ps.FailureFlag = true
					return env.NewError("Command failed: " + err.Error())
				}
				return *env.NewString(outb.String())
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "scmd\\capture")
			}
		},
	},

	// Example:
	// ; exit 0
	"exit": { // **
		Argsn: 1,
		Doc:   "Exits the process with the given integer status code (or 0 for any non-integer).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			util.BeforeExit()
			switch code := arg0.(type) {
			case env.Integer:
				os.Exit(int(code.Value))
				return nil
			default:
				fmt.Println(code.Inspect(*ps.Idx))
				os.Exit(0)
				return nil
			}
		},
	},

	// -------------------------------------------------------------------------
	// Rye-itself - args / history (requires os.Args / process context)
	// -------------------------------------------------------------------------

	// Rye-itself//Args?
	// Summary: Returns command line arguments as a block of parsed values.
	// Args: none
	// Returns: block of values (each arg parsed as integer, float, or string)
	// Example:
	//   Rye-itself//Args?
	"Rye-itself//Args?": {
		Argsn: 0,
		Doc:   "Returns command line arguments as a block of parsed values. Each argument is converted to appropriate type (integer, float, or string).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ryeItselfArgsParsed(ps)
		},
	},

	// Rye-itself//Args\\raw?
	// Summary: Returns raw command line arguments joined as a single string.
	// Args: separator (ignored currently; kept for future parity) – pass _
	// Returns: string
	// Example:
	//   Rye-itself//Args\\raw? _
	"Rye-itself//Args\\raw?": {
		Argsn: 1,
		Doc:   "Returns raw command line arguments joined as a single string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			firstArg := ryeFirstArgIdx(ps)
			if len(os.Args) > firstArg {
				return *env.NewString(strings.Join(os.Args[firstArg:], " "))
			}
			return *env.NewString("")
		},
	},

	// Rye-itself//History?
	// Summary: Returns last N lines from REPL history.
	// Args: rye-itself native (dot-dispatch), n: integer number of lines
	// Returns: block of strings
	// Notes: Requires REPL; otherwise returns an error
	// Example:
	//   Rye-itself//History? _ 10
	"Rye-itself//History?": {
		Argsn: 2,
		Doc:   "Returns a block of the last N lines from REPL history.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch n := arg1.(type) {
			case env.Integer:
				if ps.GetHistoryLast == nil {
					return evaldo.MakeBuiltinError(ps, "History not available (not running in REPL)", "Rye-itself//history")
				}
				lines := ps.GetHistoryLast(int(n.Value))
				objs := make([]env.Object, len(lines))
				for i, line := range lines {
					objs[i] = *env.NewString(line)
				}
				return *env.NewBlock(*env.NewTSeries(objs))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "Rye-itself//history")
			}
		},
	},

	// Rye-itself//Landlock
	// Summary: Create a Landlock builder object (Linux only) to configure and then enforce restrictions.
	// Args: none
	// Returns: native 'landlock' builder
	// Example:
	//   ll: Rye-itself//Landlock
	//   ll | landlock//Limit-to-cwd | landlock//Enforce
	"Rye-itself//Landlock": {
		Argsn: 0,
		Doc:   "Create a landlock builder object that can be configured and then enforced.",
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			return *env.NewNative(ps.Idx, security.NewLandlockBuilder(), "landlock")
		},
	},
	// landlock//Limit-to-cwd
	// Summary: Configure the builder to limit filesystem to current working directory.
	// Args: landlock native
	// Returns: same native (builder)
	// Notes: Configure only; use landlock//Enforce to apply.
	"landlock//Limit-to-cwd": {
		Argsn: 1,
		Doc:   "Limit landlock to current working directory (configure; does not enforce yet).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, _, _, _, _ env.Object) env.Object {
			if n, ok := arg0.(env.Native); ok {
				if b, ok2 := n.Value.(*security.LandlockBuilder); ok2 {
					b.LimitToCwd()
					return arg0
				}
			}
			return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Limit-to-cwd")
		},
	},
	// landlock//Allow
	// Summary: Allow access to a specific path (file or directory).
	// Args: landlock native, path (uri|string)
	// Returns: same native (builder)
	"landlock//Allow": {
		Argsn: 2,
		Doc:   "Allow access to a specific path (file or directory).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, _, _, _ env.Object) env.Object {
			var p string
			switch v := arg1.(type) {
			case env.Uri:
				p = v.Path
			case env.String:
				p = v.Value
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.UriType, env.StringType}, "Allow")
			}
			if n, ok := arg0.(env.Native); ok {
				if b, ok2 := n.Value.(*security.LandlockBuilder); ok2 {
					b.AllowPath(p)
					return arg0
				}
			}
			return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Allow")
		},
	},
	// landlock//Allow-exec
	// Summary: Allow execute on a file, or RX under a directory.
	// Args: landlock native, path (uri|string)
	// Returns: same native (builder)
	"landlock//Allow-exec": {
		Argsn: 2,
		Doc:   "Allow execute access to a specific file (or RX for all files under a directory).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, _, _, _ env.Object) env.Object {
			var p string
			switch v := arg1.(type) {
			case env.Uri:
				p = v.Path
			case env.String:
				p = v.Value
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.UriType, env.StringType}, "Allow-exec")
			}
			if n, ok := arg0.(env.Native); ok {
				if b, ok2 := n.Value.(*security.LandlockBuilder); ok2 {
					b.AllowExec(p)
					return arg0
				}
			}
			return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Allow-exec")
		},
	},
	// landlock//Enforce
	// Summary: Enforce the configured landlock ruleset (one-shot).
	// Args: landlock native
	// Returns: same native (builder)
	"landlock//Enforce": {
		Argsn: 1,
		Doc:   "Enforce the configured landlock ruleset exactly once.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, _, _, _, _ env.Object) env.Object {
			if n, ok := arg0.(env.Native); ok {
				if b, ok2 := n.Value.(*security.LandlockBuilder); ok2 {
					if err := b.Enforce(); err != nil {
						ps.FailureFlag = true
						return env.NewError("failed to enforce landlock: " + err.Error())
					}
					return arg0
				}
			}
			return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Enforce")
		},
	},

	// Rye-itself//Landlock-to-cwd
	// Summary: Immediately restrict filesystem access to current working directory using Landlock (Linux only).
	// Args: none
	// Returns: 'ok tagword on success; error on failure
	// Notes: Call early in the script; applies process-wide and is irreversible for current process.
	"Rye-itself//Landlock-to-cwd": {
		Argsn: 0,
		Doc:   "Restrict filesystem access to the current working directory and its subdirectories using Landlock (Linux only). Call early in the script.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			// Determine base dir: prefer ProgramState.WorkingPath
			base := ps.WorkingPath
			if base == "" {
				wd, err := os.Getwd()
				if err != nil {
					return *env.NewError(fmt.Sprintf("failed to get working directory: %v", err))
				}
				base = wd
			}
			abs, err := filepath.Abs(base)
			if err == nil {
				base = abs
			}

			// Build rules: read/write everything under base directory (dirs+files)
			rules := []landlock.Rule{
				landlock.RWDirs(base),
				landlock.RWFiles(base),
			}

			if err := landlock.V1.BestEffort().RestrictPaths(rules...); err != nil {
				return *env.NewError(fmt.Sprintf("failed to apply landlock: %v", err))
			}

			// Expose state for inspection
			os.Setenv("RYE_LANDLOCK_PROFILE", "cwd-rw")
			return env.Tagword{Index: ps.Idx.IndexWord("ok")}
		},
	},

	// Rye-itself//Landlock-only
	// Summary: Immediately restrict filesystem access to exactly the provided paths and modes.
	// Args: spec (string or block of strings). Spec format examples:
	//   "/proj:r,/proj/bin:rx,/tmp:rw" or block [ "/proj:r" "/proj/bin:rx" "/tmp:rw" ]
	// Modes:
	//   r  => read-only (files + dirs)
	//   rw => read-write (files + dirs)
	//   rx => read + execute (dir => RX for files under it; file => exec)
	// Returns: 'ok tagword on success; error on failure
	// Notes: Call early in the script; applies process-wide and is irreversible.
	"Rye-itself//Landlock-only": {
		Argsn: 1,
		Doc:   "Immediately restrict filesystem access to the provided list of paths with modes r, rw, or rx. Example: Rye-itself//Landlock-only \"/proj:r,/tmp:rw,/proj/bin:rx\"",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			// Parse spec(s)
			items := []string{}
			switch v := arg0.(type) {
			case env.String:
				for _, part := range strings.Split(v.Value, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						items = append(items, part)
					}
				}
			case env.Block:
				for _, o := range v.Series.S {
					if s, ok := o.(env.String); ok {
						items = append(items, strings.TrimSpace(s.Value))
					} else {
						return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType}, "Landlock-only")
					}
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType}, "Landlock-only")
			}

			if len(items) == 0 {
				return *env.NewError("Landlock-only: empty spec")
			}

			var rules []landlock.Rule
			for _, it := range items {
				// each item: path[:mode]
				path := it
				mode := "r"
				if i := strings.LastIndex(it, ":"); i >= 0 {
					path = strings.TrimSpace(it[:i])
					mode = strings.TrimSpace(it[i+1:])
				}
				if path == "" {
					return *env.NewError("Landlock-only: empty path in spec")
				}
				abs, err := filepath.Abs(path)
				if err == nil {
					path = abs
				}
				// Assemble rules based on mode
				switch mode {
				case "r":
					rules = append(rules, landlock.RODirs(path), landlock.ROFiles(path))
				case "rw":
					rules = append(rules, landlock.RWDirs(path), landlock.RWFiles(path))
				case "rx":
					// Read-only access; execution cannot be explicitly granted per-file via landlock-go API.
					// Best-effort: allow RO on the path; for executables, scripts should add the parent dir with rx in separate entries if needed.
					rules = append(rules, landlock.RODirs(path), landlock.ROFiles(path))
				default:
					return *env.NewError("Landlock-only: invalid mode (use r, rw, or rx)")
				}
			}

			if err := landlock.V1.BestEffort().RestrictPaths(rules...); err != nil {
				return *env.NewError(fmt.Sprintf("failed to apply landlock: %v", err))
			}
			os.Setenv("RYE_LANDLOCK_PROFILE", "custom")
			return env.Tagword{Index: ps.Idx.IndexWord("ok")}
		},
	},

	// Rye-itself//Is-unshare
	// Summary: Returns true if running inside Rye --unshare sandbox (Linux only).
	// Args: none
	// Returns: boolean
	// Notes: Checks env var RYE_UNSHARE_CHILD set by the runner in the unshare child.
	// Rye-itself//Is-unshare-fs
	// Summary: Returns true if filesystem namespace isolation is enabled under --unshare.
	// Args: none
	// Returns: boolean
	// Notes: Checks env var RYE_UNSHARE_FS set by the runner.
	"Rye-itself//Is-unshare-fs": {
		Argsn: 0,
		Doc:   "Returns true if filesystem namespace isolation is enabled under --unshare.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			return *env.NewBoolean(os.Getenv("RYE_UNSHARE_FS") == "1")
		},
	},

	// Rye-itself//Is-unshare-net
	// Summary: Returns true if network namespace isolation is enabled under --unshare.
	// Args: none
	// Returns: boolean
	// Notes: Checks env var RYE_UNSHARE_NET set by the runner.
	"Rye-itself//Is-unshare-net": {
		Argsn: 0,
		Doc:   "Returns true if network namespace isolation is enabled under --unshare.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			return *env.NewBoolean(os.Getenv("RYE_UNSHARE_NET") == "1")
		},
	},

	// Rye-itself//Is-unshare-pid
	// Summary: Returns true if PID namespace isolation is enabled under --unshare.
	// Args: none
	// Returns: boolean
	// Notes: Checks env var RYE_UNSHARE_PID set by the runner.
	"Rye-itself//Is-unshare-pid": {
		Argsn: 0,
		Doc:   "Returns true if PID namespace isolation is enabled under --unshare.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			return *env.NewBoolean(os.Getenv("RYE_UNSHARE_PID") == "1")
		},
	},

	// Rye-itself//Is-unshare-uts
	// Summary: Returns true if UTS namespace isolation is enabled under --unshare.
	// Args: none
	// Returns: boolean
	// Notes: Checks env var RYE_UNSHARE_UTS set by the runner.
	"Rye-itself//Is-unshare-uts": {
		Argsn: 0,
		Doc:   "Returns true if UTS namespace isolation is enabled under --unshare.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			return *env.NewBoolean(os.Getenv("RYE_UNSHARE_UTS") == "1")
		},
	},

	// Rye-itself//Unshare-config?
	// Summary: Returns a dictionary describing the active unshare isolation config.
	// Args: none
	// Returns: dict with keys: active, fs, net, pid, uts (booleans)
	// Notes: Reads environment variables set by the runner: RYE_UNSHARE_CHILD, RYE_UNSHARE_FS, RYE_UNSHARE_NET, RYE_UNSHARE_PID, RYE_UNSHARE_UTS.
	// Example:
	//   cc Rye-itself Unshare-config? |print
	"Rye-itself//Unshare-config?": {
		Argsn: 0,
		Doc:   "Returns a dictionary with unshare isolation flags: active, fs, net, pid, uts (booleans).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			d := make(map[string]any, 5)
			d["active"] = *env.NewBoolean(os.Getenv("RYE_UNSHARE_CHILD") == "1")
			d["fs"] = *env.NewBoolean(os.Getenv("RYE_UNSHARE_FS") == "1")
			d["net"] = *env.NewBoolean(os.Getenv("RYE_UNSHARE_NET") == "1")
			d["pid"] = *env.NewBoolean(os.Getenv("RYE_UNSHARE_PID") == "1")
			d["uts"] = *env.NewBoolean(os.Getenv("RYE_UNSHARE_UTS") == "1")
			return *env.NewDict(d)
		},
	},

	"Rye-itself//Is-unshare": {
		Argsn: 0,
		Doc:   "Returns true if current process runs inside Rye --unshare sandbox (Linux only).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			return *env.NewBoolean(os.Getenv("RYE_UNSHARE_CHILD") == "1")
		},
	},

	// Rye-itself//Is-dry-run
	// Summary: Returns true if Rye is in dry-run/scenario mode.
	// Args: none
	// Returns: boolean
	// Notes: True when --dry-run is used or scenario is active.
	// Rye-itself//Require-unshared
	// Summary: Fails and exits the program if unshare is not active.
	// Args: none
	// Returns: never on failure (process exits); 'ok tagword when already unshared
	// Notes: Intended for CI or scripts that must run sandboxed. Linux-only behavior; on non-Linux it will not detect unshare and will fail unless adapted.
	"Rye-itself//Require-unshared": {
		Argsn: 0,
		Doc:   "Fails and exits if unshare is not active (use in CI/safety-critical runs).",
		Pure:  false,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			if os.Getenv("RYE_UNSHARE_CHILD") == "1" {
				return env.Tagword{Index: ps.Idx.IndexWord("ok")}
			}
			fmt.Fprintln(os.Stderr, "Error: Rye requires --unshare for this script. Re-run with: rye --unshare ...")
			util.BeforeExit()
			os.Exit(1)
			return nil
		},
	},

	"Rye-itself//Is-dry-run": {
		Argsn: 0,
		Doc:   "Returns true if Rye is running in dry-run/scenario mode (activated via --dry-run or scenario context)",
		Pure:  true,
		Fn: func(ps *env.ProgramState, _ env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			if os.Getenv("RYE_DRY_RUN") == "1" {
				return *env.NewBoolean(true)
			}
			// Fallback to evaldo's scenario detector
			if evaldo_BatteryIsScenario(ps) {
				return *env.NewBoolean(true)
			}
			return *env.NewBoolean(false)
		},
	},

	// -------------------------------------------------------------------------
	// stdout capture
	// -------------------------------------------------------------------------

	// Tests:
	// equal { capture-stdout { print 1 } } "1\n"
	// error { capture-stdout 1 }
	"capture-stdout": { // **
		Argsn: 1,
		Doc:   "Executes a block of code while capturing all output to stdout, returning the captured output as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				old := os.Stdout // keep backup of the real stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				outC := make(chan string, 1000)
				g := errgroup.Group{}
				g.Go(func() error {
					var buf bytes.Buffer
					_, err := io.Copy(&buf, r)
					if err != nil {
						w.Close()
						os.Stdout = old
						fmt.Println(err.Error())
						return err
					}
					outC <- buf.String()
					return nil
				})

				ser := ps.Ser
				ps.Ser = bloc.Series
				ps.BlockFile = bloc.FileName
				ps.BlockLine = bloc.Line
				evaldo.Eval(ps)
				ps.Ser = ser

				w.Close()
				os.Stdout = old

				if err := g.Wait(); err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Error reading stdout: %v", err), "capture-stdout")
				}
				out := <-outC

				ps.SkipFlag = false
				evaldo.MaybeDisplayFailureOrError(ps, ps.Idx, "capture-stdout")

				if ps.ErrorFlag {
					return ps.Res
				}
				return *env.NewString(out)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType}, "capture-stdout")
			}
		},
	},
}

// ---------------------------------------------------------------------------
// Private helpers shared by the args builtins
// ---------------------------------------------------------------------------

func ryeFirstArgIdx(ps *env.ProgramState) int {
	if ps.Embedded {
		return 1
	}
	return 2
}

func ryeItselfArgsParsed(ps *env.ProgramState) env.Object {
	firstArg := ryeFirstArgIdx(ps)
	if firstArg >= len(os.Args) {
		return *env.NewBlock(*env.NewTSeries([]env.Object{}))
	}

	args := os.Args[firstArg:]
	lst := make([]env.Object, len(args))

	intRe := regexp.MustCompile("^[+-]?[0-9]+$")
	floatRe := regexp.MustCompile("^[+-]?[0-9]*\\.[0-9]+$")

	for i, arg := range args {
		if intRe.MatchString(arg) {
			if num, err := strconv.ParseInt(arg, 10, 64); err == nil {
				lst[i] = *env.NewInteger(num)
				continue
			}
		}
		if floatRe.MatchString(arg) {
			if num, err := strconv.ParseFloat(arg, 64); err == nil {
				lst[i] = *env.NewDecimal(num)
				continue
			}
		}
		lst[i] = *env.NewString(arg)
	}
	return *env.NewBlock(*env.NewTSeries(lst))
}

// Bridge to evaldo.isScenarioMode without export; we re-use its logic parts available:
func evaldo_BatteryIsScenario(ps *env.ProgramState) bool {
	// If batteries expose a hook, use it
	if evaldo.BatteryIsScenarioHook != nil && evaldo.BatteryIsScenarioHook(ps) {
		return true
	}
	// Check context sentinel 'scenario'
	if idx, found := ps.Idx.GetIndex("scenario"); found {
		if obj, ok := ps.Ctx.Get(idx); ok && obj != nil {
			return true
		}
	}
	return false
}
