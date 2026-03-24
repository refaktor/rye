//go:build !no_pipes
// +build !no_pipes

package evaldo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"

	"github.com/bitfield/script"
)

var Builtins_pipes = map[string]*env.Builtin{

	//
	// ##### Pipes ##### "Unix-like pipe operations for data processing"
	//

	// Example: cat %file.txt |into-string
	// Args:
	// * path: URI path to the file
	// Returns:
	// * script-pipe containing the file contents
	"cat": {
		Argsn: 1,
		Doc:   "Creates a new pipe object from a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				p := script.File(path.GetPath())
				return *env.NewNative(ps.Idx, p, "script-pipe")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "p-new-file")
			}
		},
	},

	// Example: find %/home/user |match ".go" |wcl
	// Args:
	// * path: URI path to the directory
	// Returns:
	// * script-pipe containing file paths, one per line
	"find": {
		Argsn: 1,
		Doc:   "Creates a pipe object listing all the files in the directory and its subdirectories recursively, one per line.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				p := script.FindFiles(path.GetPath())
				return *env.NewNative(ps.Idx, p, "script-pipe")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "p-new-find-files")
			}
		},
	},

	// Tests:
	// equal { list %./testdata/ |wc\l > 0 } true
	// Args:
	// * path: URI path to directory or glob pattern string
	// Returns:
	// * script-pipe containing file paths, one per line
	"list": {
		Argsn: 1,
		Doc:   "Creates a pipe object listing all the files in the directory, one per line. Accepts and URI or glob pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				p := script.ListFiles(path.GetPath())
				return *env.NewNative(ps.Idx, p, "script-pipe")
			case env.String:
				p := script.ListFiles(path.Value)
				return *env.NewNative(ps.Idx, p, "script-pipe")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "p-new-list-files")
			}
		},
	},

	// Tests:
	// equal { from-block { "apple" "banana" "cherry" } |into-block } { "apple" "banana" "cherry" }
	// Args:
	// * block: Block of strings to convert to pipe lines
	// Returns:
	// * script-pipe with each string as a line
	"from-block": {
		Argsn: 1,
		Doc:   "Creates a pipe object from a block of strings, one per line.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				strs := make([]string, len(block.Series.S))
				for i, item := range block.Series.S {
					switch s := item.(type) {
					case env.String:
						strs[i] = s.Value
					default:
						return MakeBuiltinError(ps, "Block must contain only strings", "p-new-block")
					}
				}
				p := script.Slice(strs)
				return *env.NewNative(ps.Idx, p, "script-pipe")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "p-new-block")
			}
		},
	},

	// Tests:
	// equal { echo "hello world" |into-string } "hello world\n"
	// equal { echo "line1\nline2" |into-block } { "line1" "line2" }
	// Args:
	// * str: String content for the pipe
	// Returns:
	// * script-pipe containing the string
	"echo": {
		Argsn: 1,
		Doc:   "Creates a pipe object from a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				p := script.Echo(s.Value)
				return *env.NewNative(ps.Idx, p, "script-pipe")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "p-new-echo")
			}
		},
	},

	// Tests:
	// equal { cat\opt %./nonexistent-file.txt |into-string } ""
	// Args:
	// * path: URI path to the file
	// Returns:
	// * script-pipe containing file contents, or empty pipe if file doesn't exist
	"cat\\opt": {
		Argsn: 1,
		Doc:   "Creates a pipe object from a file if it exists, otherwise returns an empty pipe object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				p := script.IfExists(path.GetPath())
				return *env.NewNative(ps.Idx, p, "script-pipe")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "p-new-if-exists")
			}
		},
	},

	// Tests:
	// equal { exec "echo hello" |into-string |trim } "hello"
	// Args:
	// * cmd: Shell command string to execute
	// Returns:
	// * script-pipe containing command output
	"exec": {
		Argsn: 1,
		Doc:   "Creates a pipe object from a command that is executed.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cmd := arg0.(type) {
			case env.String:
				p := script.Exec(cmd.Value)
				return *env.NewNative(ps.Idx, p, "script-pipe")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "p-new-exec")
			}
		},
	},

	// Tests:
	// equal { echo "hello" |exec\in "cat" |into-string |trim } "hello"
	// Args:
	// * pipe: script-pipe whose contents are sent as input to the command
	// * cmd: Shell command string to execute
	// Returns:
	// * script-pipe containing command output
	"exec\\in": {
		Argsn: 2,
		Doc:   "Executes a command by sending it the contents of the pipe as input and returns a pipe object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch cmd := arg1.(type) {
					case env.String:
						newPipe := pipe.Exec(cmd.Value)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-exec")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-exec")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-exec")
			}
		},
	},

	// Tests:
	// equal { echo "hello\nworld" |exec\each "echo {{.}}" |into-block } { "hello" "world" }
	// Args:
	// * pipe: script-pipe with lines to process
	// * cmd: Go template string for command; {{.}} is replaced with each line
	// Returns:
	// * script-pipe with combined output of all commands
	"exec\\each": {
		Argsn: 2,
		Doc:   "Executes a command from a Go template for each line in the pipe and returns a pipe object with the output of each command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch cmd := arg1.(type) {
					case env.String:
						newPipe := pipe.ExecForEach(cmd.Value)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-exec-for-each")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-exec-for-each")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-exec-for-each")
			}
		},
	},

	// Tests:
	// equal { echo "hello" |into-string } "hello\n"
	// equal { from-block { "a" "b" } |into-string } "a\nb\n"
	// Args:
	// * pipe: script-pipe to read from
	// Returns:
	// * string containing all pipe contents
	"into-string": {
		Argsn: 1,
		Doc:   "Returns pipe contents as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					str, err := pipe.String()
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "p-string")
					}
					return *env.NewString(str)
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-string")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-string")
			}
		},
	},

	// Tests:
	// ; into-file writes contents to a file (tested via side effect)
	// Args:
	// * pipe: script-pipe to read from
	// * path: URI path to the output file
	// Returns:
	// * integer number of bytes written
	"into-file": {
		Argsn: 2,
		Doc:   "Writes pipe contents to a file and returns the number of bytes written.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch path := arg1.(type) {
					case env.Uri:
						i, err := pipe.WriteFile(path.GetPath())
						if err != nil {
							return MakeBuiltinError(ps, err.Error(), "p-write-file")
						}
						return *env.NewInteger(i)
					default:
						return MakeArgError(ps, 2, []env.Type{env.UriType}, "p-write-file")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-write-file")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-write-file")
			}
		},
	},

	// Tests:
	// ; out prints to stdout (tested via side effect)
	// Args:
	// * pipe: script-pipe to read from
	// Returns:
	// * integer number of bytes written to stdout
	"out": {
		Argsn: 1,
		Doc:   "Prints pipe contents to stdout and returns the number of bytes written.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					i, err := pipe.Stdout()
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "p-stdout")
					}
					return *env.NewInteger(int64(i))
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-stdout")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-stdout")
			}
		},
	},

	// Tests:
	// equal { echo "a\nb\nc\nd\ne" |head 3 |into-block } { "a" "b" "c" }
	// equal { echo "a\nb" |head 5 |into-block } { "a" "b" }
	// Args:
	// * pipe: script-pipe to read from
	// * n: number of lines to keep from the beginning
	// Returns:
	// * script-pipe with only the first n lines
	"head": {
		Argsn: 2,
		Doc:   "Returns a pipe with the first n lines from the pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch n := arg1.(type) {
					case env.Integer:
						newPipe := pipe.First(int(n.Value))
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "p-first-n")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-first-n")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-first-n")
			}
		},
	},

	// Tests:
	// equal { echo "a\nb\nc\nd\ne" |tail 3 |into-block } { "c" "d" "e" }
	// equal { echo "a\nb" |tail 5 |into-block } { "a" "b" }
	// Args:
	// * pipe: script-pipe to read from
	// * n: number of lines to keep from the end
	// Returns:
	// * script-pipe with only the last n lines
	"tail": {
		Argsn: 2,
		Doc:   "Returns a pipe with the last n lines from the pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch n := arg1.(type) {
					case env.Integer:
						newPipe := pipe.Last(int(n.Value))
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "p-last-n")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-last-n")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-last-n")
			}
		},
	},

	// Tests:
	// equal { echo "/home/user/file.txt" |dirname |into-string |trim } "/home/user"
	// equal { echo "/a/b/c\n/d/e/f" |dirname |into-block } { "/a/b" "/d/e" }
	// Args:
	// * pipe: script-pipe containing file paths, one per line
	// Returns:
	// * script-pipe with directory component of each path
	"dirname": {
		Argsn: 1,
		Doc:   "Reads paths from the pipe, one per line, and returns the directory component of each.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Dirname()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-dirname")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-dirname")
			}
		},
	},

	// Tests:
	// equal { echo "/home/user/file.txt" |basename |into-string |trim } "file.txt"
	// equal { echo "/a/b/c\n/d/e/f" |basename |into-block } { "c" "f" }
	// Args:
	// * pipe: script-pipe containing file paths, one per line
	// Returns:
	// * script-pipe with only the filename component of each path
	"basename": {
		Argsn: 1,
		Doc:   "Reads paths from the pipe, one per line, and removes any leading directory components from each.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Basename()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-basename")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-basename")
			}
		},
	},

	// Tests:
	// equal { echo "a\nb\nc" |wc\l } 3
	// equal { echo "" |wc\l } 1
	// Args:
	// * pipe: script-pipe to count lines from
	// Returns:
	// * integer count of lines
	"wc\\l": {
		Argsn: 1,
		Doc:   "Returns the number of lines in a pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					n, err := pipe.CountLines()
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "p-count-lines")
					}
					return *env.NewInteger(int64(n))
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-count-lines")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-count-lines")
			}
		},
	},
	/*
		"wc": {
			Argsn: 1,
			Doc:   "Returns the number of lines in a pipe.",
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch p := arg0.(type) {
				case env.Native:
					switch pipe := p.Value.(type) {
					case *script.Pipe:
						n, err := pipe.CountCharacters()
						if err != nil {
							return MakeBuiltinError(ps, err.Error(), "p-count-lines")
						}
						return *env.NewInteger(int64(n))
					default:
						return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-count-lines")
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-count-lines")
				}
			},
		},
	*/

	// Tests:
	// equal { echo "a\nb\na\na\nb" |freq |head 1 |into-string |trim } "3 a"
	// Args:
	// * pipe: script-pipe to analyze
	// Returns:
	// * script-pipe with "count value" lines sorted by frequency descending
	"freq": {
		Argsn: 1,
		Doc:   "Returns a pipe object with the frequency of each line in a pipe in descending order.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Freq()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-freq")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-freq")
			}
		},
	},

	// Tests:
	// equal { echo "a b c\nd e f" |column 2 |into-block } { "b" "e" }
	// equal { echo "one two three" |column 1 |into-string |trim } "one"
	// Args:
	// * pipe: script-pipe with whitespace-delimited columns
	// * n: column number to extract (1-indexed)
	// Returns:
	// * script-pipe with the specified column from each line
	"column": {
		Argsn: 2,
		Doc:   "Returns a pipe object with the column of each line of input, where the first column is column 1, and columns are delimited by Unicode whitespace.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch n := arg1.(type) {
					case env.Integer:
						newPipe := pipe.Column(int(n.Value))
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "p-column")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-column")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-column")
			}
		},
	},

	// Tests:
	// ; equal { echo `{"name":"test"}` |jq ".name" |into-string |trim } `"test"`
	// Args:
	// * pipe: script-pipe containing JSON data
	// * query: jq query string
	// Returns:
	// * script-pipe with jq query results
	"jq": {
		Argsn: 2,
		Doc:   "Executes the jq command on the pipe whose contents are presumed to be JSON and returns a new pipe object with the output.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch jq := arg1.(type) {
					case env.String:
						newPipe := pipe.JQ(jq.Value)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-jq")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-jq")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-jq")
			}
		},
	},

	// Tests:
	// equal { echo "apple\nbanana\napricot" |match "ap" |into-block } { "apple" "apricot" }
	// equal { echo "hello\nworld" |match "x" |into-block } { }
	// Args:
	// * pipe: script-pipe to filter
	// * str: substring to search for in each line
	// Returns:
	// * script-pipe with only lines containing the substring
	"match": {
		Argsn: 2,
		Doc:   "Returns a pipe object with lines that match the string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch s := arg1.(type) {
					case env.String:
						newPipe := pipe.Match(s.Value)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-match")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-match")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-match")
			}
		},
	},

	// Tests:
	// equal { echo "apple\nbanana\napricot" |match\regexp regexp "^a" |into-block } { "apple" "apricot" }
	// Args:
	// * pipe: script-pipe to filter
	// * regexp: compiled regexp native to match against each line
	// Returns:
	// * script-pipe with only lines matching the regexp
	"match\\regexp": {
		Argsn: 2,
		Doc:   "Returns a pipe object with lines that match the regular expression.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch r := arg1.(type) {
					case env.Native:
						switch regxp := r.Value.(type) {
						case *regexp.Regexp:
							newPipe := pipe.MatchRegexp(regxp)
							return *env.NewNative(ps.Idx, newPipe, "script-pipe")
						default:
							return MakeNativeArgError(ps, 2, []string{"regexp"}, "p-match-regexp")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "p-match-regexp")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-match-regexp")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-match-regexp")
			}
		},
	},

	// Tests:
	// equal { echo "apple\nbanana\napricot" |not-match "ap" |into-block } { "banana" }
	// equal { echo "hello\nworld" |not-match "o" |into-block } { }
	// Args:
	// * pipe: script-pipe to filter
	// * str: substring to reject in each line
	// Returns:
	// * script-pipe with only lines NOT containing the substring
	"not-match": {
		Argsn: 2,
		Doc:   "Returns a pipe object with lines that do not match the string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch s := arg1.(type) {
					case env.String:
						newPipe := pipe.Reject(s.Value)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-not-match")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-not-match")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-not-match")
			}
		},
	},

	// Tests:
	// equal { echo "apple\nbanana\napricot" |not-match\regexp regexp "^a" |into-block } { "banana" }
	// Args:
	// * pipe: script-pipe to filter
	// * regexp: compiled regexp native to reject matches
	// Returns:
	// * script-pipe with only lines NOT matching the regexp
	"not-match\\regexp": {
		Argsn: 2,
		Doc:   "Returns a pipe object with lines that do not match the regular expression.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch r := arg1.(type) {
					case env.Native:
						switch regxp := r.Value.(type) {
						case *regexp.Regexp:
							newPipe := pipe.RejectRegexp(regxp)
							return *env.NewNative(ps.Idx, newPipe, "script-pipe")
						default:
							return MakeNativeArgError(ps, 2, []string{"regexp"}, "p-not-match-regexp")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "p-not-match-regexp")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-not-match-regexp")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-not-match-regexp")
			}
		},
	},

	// Tests:
	// equal { echo "hello world" |replace "world" "rye" |into-string |trim } "hello rye"
	// equal { echo "aaa" |replace "a" "b" |into-string |trim } "bbb"
	// Args:
	// * pipe: script-pipe to process
	// * search: string to search for
	// * replacement: string to replace with
	// Returns:
	// * script-pipe with all occurrences replaced
	"replace": {
		Argsn: 3,
		Doc:   "Replaces all occurrences of a string with another string in the pipe and returns a new pipe object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch search := arg1.(type) {
					case env.String:
						switch replace := arg2.(type) {
						case env.String:
							newPipe := pipe.Replace(search.Value, replace.Value)
							return *env.NewNative(ps.Idx, newPipe, "script-pipe")
						default:
							return MakeArgError(ps, 3, []env.Type{env.StringType}, "p-replace")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-replace")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-replace")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-replace")
			}
		},
	},

	// Tests:
	// equal { echo "hello123world" |replace\regexp regexp "[0-9]+" "-" |into-string |trim } "hello-world"
	// Args:
	// * pipe: script-pipe to process
	// * regexp: compiled regexp native pattern to search for
	// * replacement: string to replace matches with
	// Returns:
	// * script-pipe with all regexp matches replaced
	"replace\\regexp": {
		Argsn: 3,
		Doc:   "Replaces all occurrences of strings that match the regexp pattern with a string in the pipe and returns a new pipe object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch searchR := arg1.(type) {
					case env.Native:
						switch searchRegexp := searchR.Value.(type) {
						case *regexp.Regexp:
							switch replace := arg2.(type) {
							case env.String:
								newPipe := pipe.ReplaceRegexp(searchRegexp, replace.Value)
								return *env.NewNative(ps.Idx, newPipe, "script-pipe")
							default:
								return MakeArgError(ps, 3, []env.Type{env.StringType}, "p-replace-regexp")
							}
						default:
							return MakeNativeArgError(ps, 2, []string{"regexp"}, "p-replace-regexp")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "p-replace-regexp")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-replace-regexp")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-replace-regexp")
			}
		},
	},

	// Tests:
	// equal { echo "a\nb\nc" |into-block } { "a" "b" "c" }
	// equal { echo "single" |into-block } { "single" }
	// Args:
	// * pipe: script-pipe to convert
	// Returns:
	// * block of strings, one per line from the pipe
	"into-block": {
		Argsn: 1,
		Doc:   "Returns a block of strings with the contents of the pipe, one per line.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					lines, err := pipe.Slice()
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "p-block")
					}
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-block")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-block")
			}
		},
	},

	// Tests:
	// equal { echo "hello" |error\opt |type? } 'void
	// Args:
	// * pipe: script-pipe to check for errors
	// Returns:
	// * error if pipe has an error, void otherwise
	"error\\opt": {
		Argsn: 1,
		Doc:   "Returns the error from the pipe, if any.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					err := pipe.Error()
					if err != nil {
						return *env.NewError(err.Error())
					}
					return env.Void{}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-error")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-error")
			}
		},
	},

	// Tests:
	// equal { echo "hello" |sha256sum |length? } 64
	// Args:
	// * pipe: script-pipe to hash
	// Returns:
	// * string containing hex-encoded SHA-256 hash
	"sha256sum": {
		Argsn: 1,
		Doc:   "Returns the hex-encoded SHA-256 hash of the entire contents of the pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					sha256, err := pipe.SHA256Sum()
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "pipes/sha256sum")
					}
					return *env.NewString(sha256)
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "pipes/sha256sum")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pipes/sha256sum")
			}
		},
	},

	// Tests:
	// ; sha256sums computes hash of files listed in pipe
	// Args:
	// * pipe: script-pipe containing file paths, one per line
	// Returns:
	// * script-pipe with SHA-256 hash of each file, one per line
	"sha256sums": {
		Argsn: 1,
		Doc:   "Reads paths from the pipe, one per line, and produces the hex-encoded SHA-256 hash of each corresponding file, one per line.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.SHA256Sums()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "pipes/sha256sums")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pipes/sha256sums")
			}
		},
	},

	// Tests:
	// equal { echo "aGVsbG8=" |decodeBase64 |into-string |trim } "hello"
	// Args:
	// * pipe: script-pipe containing base64-encoded data
	// Returns:
	// * script-pipe with decoded data
	"decodeBase64": {
		Argsn: 1,
		Doc:   "decodeBase64 produces the string represented by the base64 encoded input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.DecodeBase64()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "decodeBase64")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "decodeBase64")
			}
		},
	},

	// Tests:
	// equal { echo "hello" |encodeBase64 |into-string |trim } "aGVsbG8K"
	// Args:
	// * pipe: script-pipe containing data to encode
	// Returns:
	// * script-pipe with base64-encoded data
	"encodeBase64": {
		Argsn: 1,
		Doc:   "encodeBase64 produces the base64 encoding of the input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.EncodeBase64()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "encodeBase64")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "encodeBase64")
			}
		},
	},

	// Tests:
	// equal { echo "a\nb\nc" |join |into-string |trim } "a b c"
	// Args:
	// * pipe: script-pipe with multiple lines
	// Returns:
	// * script-pipe with all lines joined by spaces into a single line
	"join": {
		Argsn: 1,
		Doc:   "joins all the lines in the pipe's contents into a single space-separated string, which will always end with a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Join()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-join")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-join")
			}
		},
	},

	// Tests:
	// equal { exec "exit 0" |exit-status } 0
	// Args:
	// * pipe: script-pipe from a command execution
	// Returns:
	// * integer exit status of the command (0 for success)
	"exit-status": {
		Argsn: 1,
		Doc:   "Returns the integer exit status of a previous command. This will be zero unless the pipe's error status is set and the error matches the pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					status := pipe.ExitStatus()
					return *env.NewInteger(int64(status))
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-exit-status")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-exit-status")
			}
		},
	},

	// Tests:
	// ; args returns command-line arguments (depends on how program is invoked)
	// Args:
	// * (none)
	// Returns:
	// * script-pipe with command-line arguments, one per line
	"args": {
		Argsn: 0,
		Doc:   "Creates a pipe containing the program's command-line arguments from os.Args, excluding the program name, one per line.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			newPipe := script.Args()
			return *env.NewNative(ps.Idx, newPipe, "script-pipe")
		},
	},

	// Tests:
	// ; concat reads file paths and outputs their combined contents
	// Args:
	// * pipe: script-pipe containing file paths, one per line
	// Returns:
	// * script-pipe with concatenated contents of all files
	"concat": {
		Argsn: 1,
		Doc:   "concat reads paths from the pipe, one per line, and produces the contents of all the corresponding files in sequence. If there are any errors (for example, non-existent files), these will be ignored, execution will continue, and the pipe's error status will not be set.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Concat()
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-concat")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-concat")
			}
		},
	},

	// Tests:
	// equal { echo "hello" |close } 0
	// Args:
	// * pipe: script-pipe to close
	// Returns:
	// * 0 on success, error otherwise
	"close": {
		Argsn: 1,
		Doc:   "Closes the pipe's associated reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					closeErr := pipe.Close()
					if closeErr != nil {
						return *env.NewError("Error closing pipe")
					}
					return *env.NewInteger(0)
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-close")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-close")
			}
		},
	},

	// Tests:
	// ; get makes HTTP GET request (requires network)
	// Args:
	// * pipe: script-pipe with request body content
	// * url: URL string to send GET request to
	// Returns:
	// * script-pipe with server response
	"get": {
		Argsn: 2,
		Doc:   "Get makes an HTTP GET request to url, sending the contents of the pipe as the request body, and produces the server's response.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch s := arg1.(type) {
					case env.String:
						newPipe := pipe.Get(s.Value)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-get")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-get")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-get")
			}
		},
	},

	// Tests:
	// ; post makes HTTP POST request (requires network)
	// Args:
	// * pipe: script-pipe with request body content
	// * url: URL string to send POST request to
	// Returns:
	// * script-pipe with server response
	"post": {
		Argsn: 2,
		Doc:   "Post makes an HTTP POST request to url, using the contents of the pipe as the request body, and produces the server's response.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch s := arg1.(type) {
					case env.String:
						newPipe := pipe.Post(s.Value)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-post")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-post")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-post")
			}
		},
	},

	// Tests:
	// equal { new |into-string } ""
	// Args:
	// * (none)
	// Returns:
	// * script-pipe with empty content
	"new": {
		Argsn: 0,
		Doc:   "new creates a new pipe with an empty reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			newPipe := script.NewPipe()
			return *env.NewNative(ps.Idx, newPipe, "script-pipe")
		},
	},

	// Tests:
	// equal { echo "hello" |wait } 0
	// Args:
	// * pipe: script-pipe to wait for completion
	// Returns:
	// * 0 on success, error if pipe had an error
	"wait": {
		Argsn: 1,
		Doc:   "Wait reads the pipe to completion and returns any error present on the pipe, or 0 otherwise..",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					waitErr := pipe.Wait()
					if waitErr != nil {
						return *env.NewError("Error in pipe during waiting")
					}
					return *env.NewInteger(0)
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-wait")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-wait")
			}
		},
	},

	// Tests:
	// ; append-to-file appends to file (tested via side effect)
	// Args:
	// * pipe: script-pipe to read from
	// * path: file path string or URI to append to
	// Returns:
	// * integer number of bytes written
	"append-to-file": {
		Argsn: 2,
		Doc:   "append-to-file appends the contents of the pipe to the file path, creating it if necessary, and returns the number of bytes successfully written, or an error.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch s := arg1.(type) {
					case env.String:
						writtenBytes, err := pipe.AppendFile(s.Value)
						if err != nil {
							return *env.NewError("Error while appending data to files.")
						}
						return *env.NewInteger(writtenBytes)
					case env.Uri:
						writtenBytes, err := pipe.AppendFile(s.Path)
						if err != nil {
							return *env.NewError("Error while appending data to rey-file.")
						}
						return *env.NewInteger(writtenBytes)
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType, env.UriType}, "p-append-to-file")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-append-to-file")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-append-to-file")
			}
		},
	},

	// Tests:
	// ; stdin reads from standard input (interactive)
	// Args:
	// * (none)
	// Returns:
	// * script-pipe reading from stdin
	"stdin": {
		Argsn: 0,
		Doc:   "Stdin creates a pipe that reads from os.Stdin.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			newPipe := script.Stdin()
			return *env.NewNative(ps.Idx, newPipe, "script-pipe")
		},
	},

	// Tests:
	// equal { echo "hello" |error } 0
	// Args:
	// * pipe: script-pipe to check for errors
	// Returns:
	// * error if pipe has an error, 0 otherwise
	"error": {
		Argsn: 1,
		Doc:   "error - returns any error present on the pipe, or 0 otherwise.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					waitErr := pipe.Error()
					if waitErr != nil {
						return *env.NewError("Error in pipe: " + waitErr.Error())
					}
					return *env.NewInteger(0)
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-error")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-error")
			}
		},
	},

	// Tests:
	// equal { echo "hello" |set-error "test error" |error |type? } 'error
	// Args:
	// * pipe: script-pipe to set error on
	// * err: error message string
	// Returns:
	// * script-pipe with error set
	"set-error": {
		Argsn: 2,
		Doc:   "set-error sets the error err on the pipe and return it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch errStr := arg1.(type) {
					case env.String:
						err := errors.New(errStr.Value)
						pipe.SetError(err)
						return *env.NewNative(ps.Idx, pipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "p-set-error")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "p-set-error")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "p-set-error")
			}
		},
	},

	// Example: cat %logfile.log |extract\regexp regexp `apitoken "([^"]+)` |into-block
	// Args:
	// * pipe: script-pipe to read from
	// * regexp: compiled regexp native; if the pattern has a capture group, group 1 is extracted, otherwise the whole match
	// Returns:
	// * script-pipe with one extracted match per output line (like grep -oP)
	"extract\\regexp": {
		Argsn: 2,
		Doc:   "For each line in the pipe, extracts all regexp matches (capture group 1 if present, otherwise the whole match) and outputs each match on a separate line. Equivalent to grep -oP with a capture group.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch r := arg1.(type) {
					case env.Native:
						switch rx := r.Value.(type) {
						case *regexp.Regexp:
							newPipe := pipe.Filter(func(r io.Reader, w io.Writer) error {
								scanner := bufio.NewScanner(r)
								for scanner.Scan() {
									line := scanner.Text()
									matches := rx.FindAllStringSubmatch(line, -1)
									for _, match := range matches {
										if len(match) > 1 {
											fmt.Fprintln(w, match[1])
										} else {
											fmt.Fprintln(w, match[0])
										}
									}
								}
								return scanner.Err()
							})
							return *env.NewNative(ps.Idx, newPipe, "script-pipe")
						default:
							return MakeNativeArgError(ps, 2, []string{"regexp"}, "extract\\regexp")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "extract\\regexp")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "extract\\regexp")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "extract\\regexp")
			}
		},
	},

	// Tests:
	// equal { echo "banana\napple\ncherry" |sort |into-block } { "apple" "banana" "cherry" }
	// Args:
	// * pipe: script-pipe whose lines should be sorted
	// Returns:
	// * script-pipe with lines sorted alphabetically (ascending)
	"sort": {
		Argsn: 1,
		Doc:   "Sorts the lines of the pipe alphabetically in ascending order.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Filter(func(r io.Reader, w io.Writer) error {
						scanner := bufio.NewScanner(r)
						var lines []string
						for scanner.Scan() {
							lines = append(lines, scanner.Text())
						}
						if err := scanner.Err(); err != nil {
							return err
						}
						sort.Strings(lines)
						for _, line := range lines {
							fmt.Fprintln(w, line)
						}
						return nil
					})
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "sort")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "sort")
			}
		},
	},

	// Tests:
	// equal { echo "banana\napple\ncherry" |sort\reverse |into-block } { "cherry" "banana" "apple" }
	// Args:
	// * pipe: script-pipe whose lines should be sorted
	// Returns:
	// * script-pipe with lines sorted alphabetically in reverse (descending) order
	"sort\\reverse": {
		Argsn: 1,
		Doc:   "Sorts the lines of the pipe alphabetically in descending (reverse) order.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Filter(func(r io.Reader, w io.Writer) error {
						scanner := bufio.NewScanner(r)
						var lines []string
						for scanner.Scan() {
							lines = append(lines, scanner.Text())
						}
						if err := scanner.Err(); err != nil {
							return err
						}
						sort.Sort(sort.Reverse(sort.StringSlice(lines)))
						for _, line := range lines {
							fmt.Fprintln(w, line)
						}
						return nil
					})
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "sort\\reverse")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "sort\\reverse")
			}
		},
	},

	// Tests:
	// equal { echo "apple\napple\nbanana\napple" |uniq |into-block } { "apple" "banana" "apple" }
	// Args:
	// * pipe: script-pipe whose adjacent duplicate lines should be removed
	// Returns:
	// * script-pipe with consecutive duplicate lines removed (like the uniq command)
	"uniq": {
		Argsn: 1,
		Doc:   "Removes consecutive duplicate lines from the pipe, like the Unix uniq command. Combine with sort to remove all duplicates.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Filter(func(r io.Reader, w io.Writer) error {
						scanner := bufio.NewScanner(r)
						last := ""
						first := true
						for scanner.Scan() {
							line := scanner.Text()
							if first || line != last {
								fmt.Fprintln(w, line)
								last = line
								first = false
							}
						}
						return scanner.Err()
					})
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "uniq")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "uniq")
			}
		},
	},

	// GOPSUTIL

	// ========================================
	// GROUP 1 — Trivial (pure Go, stateless per-line)
	// ========================================

	// Tests:
	// equal { echo "  hello  \n  world  " |trim-lines |into-block } { "hello" "world" }
	// Args:
	// * pipe: script-pipe to process
	// Returns:
	// * script-pipe with leading and trailing whitespace removed from each line
	"trim-lines": {
		Argsn: 1,
		Doc:   "Trims leading and trailing whitespace from each line in the pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.FilterLine(func(line string) string {
						return strings.TrimSpace(line)
					})
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "trim-lines")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "trim-lines")
			}
		},
	},

	// Tests:
	// equal { echo "line1\nline2" |prefix-lines ">> " |into-block } { ">> line1" ">> line2" }
	// Args:
	// * pipe: script-pipe to process
	// * prefix: string to prepend to each line
	// Returns:
	// * script-pipe with prefix added to the beginning of each line
	"prefix-lines": {
		Argsn: 2,
		Doc:   "Prepends a string to the beginning of each line in the pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch prefix := arg1.(type) {
					case env.String:
						newPipe := pipe.FilterLine(func(line string) string {
							return prefix.Value + line
						})
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "prefix-lines")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "prefix-lines")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prefix-lines")
			}
		},
	},

	// ========================================
	// GROUP 2 — Simple (pure Go, stateful, sequential)
	// ========================================

	// Tests:
	// equal { echo "a\nb\nc\nd\ne" |skip 2 |into-block } { "c" "d" "e" }
	// equal { echo "a\nb\nc" |skip 0 |into-block } { "a" "b" "c" }
	// equal { echo "a\nb" |skip 5 |into-block } { }
	// Args:
	// * pipe: script-pipe to process
	// * n: number of lines to skip from the beginning
	// Returns:
	// * script-pipe with the first n lines removed
	"skip": {
		Argsn: 2,
		Doc:   "Skips the first n lines from the pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch n := arg1.(type) {
					case env.Integer:
						skipCount := int(n.Value)
						newPipe := pipe.Filter(func(r io.Reader, w io.Writer) error {
							scanner := bufio.NewScanner(r)
							lineNum := 0
							for scanner.Scan() {
								lineNum++
								if lineNum > skipCount {
									fmt.Fprintln(w, scanner.Text())
								}
							}
							return scanner.Err()
						})
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "skip")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "skip")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "skip")
			}
		},
	},

	// Tests:
	// equal { echo "apple\nbanana\ncherry" |number-lines |into-block } { "1: apple" "2: banana" "3: cherry" }
	// Args:
	// * pipe: script-pipe to process
	// Returns:
	// * script-pipe with each line prefixed by its line number (1-indexed) and ": "
	"number-lines": {
		Argsn: 1,
		Doc:   "Prefixes each line with its line number (1-indexed).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					newPipe := pipe.Filter(func(r io.Reader, w io.Writer) error {
						scanner := bufio.NewScanner(r)
						lineNum := 0
						for scanner.Scan() {
							lineNum++
							fmt.Fprintf(w, "%d: %s\n", lineNum, scanner.Text())
						}
						return scanner.Err()
					})
					return *env.NewNative(ps.Idx, newPipe, "script-pipe")
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "number-lines")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "number-lines")
			}
		},
	},

	// Tests:
	// equal { echo "hello world\nhello there\ngoodbye" |count-matches "hello" } 2
	// equal { echo "one\ntwo\nthree" |count-matches "x" } 0
	// Args:
	// * pipe: script-pipe to process
	// * str: string to search for
	// Returns:
	// * integer count of lines containing the string
	"count-matches": {
		Argsn: 2,
		Doc:   "Counts the number of lines containing the given string. This is a sink operation that consumes the pipe.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch s := arg1.(type) {
					case env.String:
						lines, err := pipe.Slice()
						if err != nil {
							return MakeBuiltinError(ps, err.Error(), "count-matches")
						}
						count := 0
						for _, line := range lines {
							if strings.Contains(line, s.Value) {
								count++
							}
						}
						return *env.NewInteger(int64(count))
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "count-matches")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "count-matches")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "count-matches")
			}
		},
	},

	// Tests:
	// equal { echo "before\nSTART\nmiddle1\nmiddle2\nEND\nafter" |between "START" "END" |into-block } { "middle1" "middle2" }
	// equal { echo "a\nSTART\nb\nEND\nc\nSTART\nd\nEND\ne" |between "START" "END" |into-block } { "b" "d" }
	// Args:
	// * pipe: script-pipe to process
	// * start: string marking the start of the range (exclusive)
	// * end: string marking the end of the range (exclusive)
	// Returns:
	// * script-pipe with only the lines between start and end markers (markers themselves are excluded)
	"between": {
		Argsn: 3,
		Doc:   "Returns lines between start and end marker lines (exclusive). Can handle multiple ranges.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch startStr := arg1.(type) {
					case env.String:
						switch endStr := arg2.(type) {
						case env.String:
							newPipe := pipe.Filter(func(r io.Reader, w io.Writer) error {
								scanner := bufio.NewScanner(r)
								inside := false
								for scanner.Scan() {
									line := scanner.Text()
									if strings.Contains(line, endStr.Value) {
										inside = false
										continue
									}
									if inside {
										fmt.Fprintln(w, line)
									}
									if strings.Contains(line, startStr.Value) {
										inside = true
									}
								}
								return scanner.Err()
							})
							return *env.NewNative(ps.Idx, newPipe, "script-pipe")
						default:
							return MakeArgError(ps, 3, []env.Type{env.StringType}, "between")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "between")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "between")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "between")
			}
		},
	},

	// ========================================
	// GROUP 3 — Moderate (uses existing script library method)
	// ========================================

	// Tests:
	// ; tee\file writes to a file while passing through (tested via side effect)
	// Args:
	// * pipe: script-pipe to process
	// * path: URI path to the file to write to
	// Returns:
	// * script-pipe that passes through unchanged while also writing to the file
	"tee\\file": {
		Argsn: 2,
		Doc:   "Writes the pipe contents to a file while also passing them through. Like the Unix tee command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch path := arg1.(type) {
					case env.Uri:
						file, err := os.Create(path.GetPath())
						if err != nil {
							return MakeBuiltinError(ps, err.Error(), "tee\\file")
						}
						newPipe := pipe.Tee(file)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					case env.String:
						file, err := os.Create(path.Value)
						if err != nil {
							return MakeBuiltinError(ps, err.Error(), "tee\\file")
						}
						newPipe := pipe.Tee(file)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.UriType, env.StringType}, "tee\\file")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "tee\\file")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tee\\file")
			}
		},
	},

	// ========================================
	// GROUP 4 — Complex (calls back into the Rye evaluator)
	// ========================================

	// Tests:
	// equal { echo "hello\nworld" |map-each { .uppercase } |into-block } { "HELLO" "WORLD" }
	// equal { echo "1\n2\n3" |map-each { .to-integer + 10 |to-string } |into-block } { "11" "12" "13" }
	// Args:
	// * pipe: script-pipe to process
	// * block: Rye block to evaluate for each line; receives the line as injected value, should return a string
	// Returns:
	// * script-pipe with each line transformed by the block
	"map-each": {
		Argsn: 2,
		Doc:   "Transforms each line by evaluating a Rye block with the line injected. The block should return a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch block := arg1.(type) {
					case env.Block:
						// Eager evaluation: collect all lines
						lines, err := pipe.Slice()
						if err != nil {
							return MakeBuiltinError(ps, err.Error(), "map-each")
						}

						// Process each line through the Rye block
						result := make([]string, 0, len(lines))
						ser := ps.Ser
						for _, line := range lines {
							ps.Ser = block.Series
							ps.Ser.SetPos(0)
							EvalBlockInj(ps, *env.NewString(line), true)
							if ps.ErrorFlag {
								ps.Ser = ser
								return MakeBuiltinError(ps, "Error during evaluation of block", "map-each")
							}
							if ps.FailureFlag {
								ps.Ser = ser
								return ps.Res
							}
							// Convert result to string
							switch res := ps.Res.(type) {
							case env.String:
								result = append(result, res.Value)
							default:
								// Try to convert to string representation
								result = append(result, ps.Res.Print(*ps.Idx))
							}
						}
						ps.Ser = ser

						// Create new pipe from result
						newPipe := script.Slice(result)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "map-each")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "map-each")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "map-each")
			}
		},
	},

	// Tests:
	// equal { echo "apple\nbanana\napricot\ncherry" |filter-each { .has-prefix "a" } |into-block } { "apple" "apricot" }
	// equal { echo "1\n2\n3\n4\n5" |filter-each { .to-integer > 2 } |into-block } { "3" "4" "5" }
	// Args:
	// * pipe: script-pipe to process
	// * block: Rye block to evaluate for each line; receives the line as injected value, should return a truthy/falsy value
	// Returns:
	// * script-pipe with only lines for which the block returns a truthy value
	"filter-each": {
		Argsn: 2,
		Doc:   "Filters lines by evaluating a Rye block with each line injected. Only lines for which the block returns truthy are kept.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				switch pipe := p.Value.(type) {
				case *script.Pipe:
					switch block := arg1.(type) {
					case env.Block:
						// Eager evaluation: collect all lines
						lines, err := pipe.Slice()
						if err != nil {
							return MakeBuiltinError(ps, err.Error(), "filter-each")
						}

						// Filter each line through the Rye block
						result := make([]string, 0)
						ser := ps.Ser
						for _, line := range lines {
							ps.Ser = block.Series
							ps.Ser.SetPos(0)
							EvalBlockInj(ps, *env.NewString(line), true)
							if ps.ErrorFlag {
								ps.Ser = ser
								return MakeBuiltinError(ps, "Error during evaluation of block", "filter-each")
							}
							if ps.FailureFlag {
								ps.Ser = ser
								return ps.Res
							}
							// Check if result is truthy
							if util.IsTruthy(ps.Res) {
								result = append(result, line)
							}
						}
						ps.Ser = ser

						// Create new pipe from result
						newPipe := script.Slice(result)
						return *env.NewNative(ps.Idx, newPipe, "script-pipe")
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "filter-each")
					}
				default:
					return MakeNativeArgError(ps, 1, []string{"script-pipe"}, "filter-each")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "filter-each")
			}
		},
	},
}
