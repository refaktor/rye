//go:build add_psutil
// +build add_psutil

package evaldo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bitfield/script"
	"github.com/refaktor/rye/env"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// In request we return a raw-map, because it's very inside loop call, this is sparse call, and we get tons of fields, so it would be best
// to turn them to normal Rye map (which is now Env / later Context or something like it), and they query it from Rye.

func FileExists(filePath string) int {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0 // fmt.Println("File does not exist")
		} else {
			return -1 // fmt.Println("Error checking file:", err)
		}
	} else {
		return 1
	}
}

var Builtins_devops = map[string]*env.Builtin{

	"cd": {
		Argsn: 1,
		Doc:   "Changes current directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				new := filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
				res := FileExists(new)
				if res == 1 {
					ps.WorkingPath = filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
					return arg0
				} else if res == 0 {
					return MakeBuiltinError(ps, "Path doesn't exist", "cd")
				} else {
					return MakeBuiltinError(ps, "Error determining if path exists", "cd")
				}
				// TODO -- check if it exists
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "cd")
			}
		},
	},

	"mkdir": {
		Argsn: 1,
		Doc:   "Creates a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				newDir := filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
				err := os.Mkdir(newDir, 0755) // Create directory with permissions 0755
				if err != nil {
					return MakeBuiltinError(ps, "Error creating directory: "+err.Error(), "mkdir")
				} else {
					return arg0
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mkdir")
			}
		},
	},

	"mv": {
		Argsn: 2,
		Doc:   "Creates a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				switch path2 := arg1.(type) {
				case env.Uri:
					old := filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
					new := filepath.Join(filepath.Dir(ps.WorkingPath), path2.GetPath())
					err := os.Rename(old, new)
					if err != nil {
						fmt.Println("Error renaming file:", err)
						return MakeBuiltinError(ps, "Error renaming file: "+err.Error(), "mv")
					} else {
						return arg1
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.UriType}, "mv")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mv")
			}
		},
	},

	"cwd": {
		Argsn: 0,
		Doc:   "Returns current working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewUri1(ps.Idx, "file://"+ps.WorkingPath)
		},
	},

	"lsd": {
		Argsn: 0,
		Doc:   "Returns current working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			files, err := ioutil.ReadDir(ps.WorkingPath + "/")
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory:"+err.Error(), "ls")
			}

			items := make([]env.Object, len(files))

			for i, file := range files {
				// fmt.Println(file.Name()) // Print only file/directory names

				items[i] = *env.NewUri1(ps.Idx, "file://"+file.Name())
			}
			return *env.NewBlock(*env.NewTSeries(items))

		},
	},

	// SCRIPT PIPES

	"p-new-file": {
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

	"p-new-find-files": {
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

	"p-new-list-files": {
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

	"p-new-block": {
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

	"p-new-echo": {
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

	"p-new-if-exists": {
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

	"p-new-exec": {
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

	"p-exec": {
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

	"p-exec-for-each": {
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

	"p-string": {
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

	"p-write-file": {
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
						return *env.NewInteger(int64(i))
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

	"p-stdout": {
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

	"p-first-n": {
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

	"p-last-n": {
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

	"p-dirname": {
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

	"p-basename": {
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

	"p-count-lines": {
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

	"p-freq": {
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

	"p-column": {
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

	"p-jq": {
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

	"p-match": {
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

	"p-match-regexp": {
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

	"p-not-match": {
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

	"p-not-match-regexp": {
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

	"p-replace": {
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

	"p-replace-regexp": {
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

	"p-block": {
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

	"p-error": {
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

	// GOPSUTIL

	"host-info?": {
		Argsn: 0,
		Doc:   "Get information about the host system.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			v, err := host.Info()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "host-info?")
			}
			r := env.NewDict(make(map[string]any, 10))
			r.Data["hostname"] = *env.NewString(v.Hostname)
			r.Data["uptime"] = *env.NewInteger(int64(v.Uptime))
			r.Data["boot-time"] = *env.NewInteger(int64(v.BootTime))
			r.Data["procs"] = *env.NewInteger(int64(v.Procs))
			r.Data["os"] = *env.NewString(v.OS)
			r.Data["platform"] = *env.NewString(v.Platform)
			r.Data["platform-family"] = *env.NewString(v.PlatformFamily)
			r.Data["platform-version"] = *env.NewString(v.PlatformVersion)
			r.Data["kernel-version"] = *env.NewString(v.KernelVersion)
			r.Data["virtualization-system"] = *env.NewString(v.VirtualizationSystem)
			return *r
		},
	},
	"users?": {
		Argsn: 0,
		Doc:   "Get information about users as a spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			users, err := host.Users()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "users?")
			}
			fmt.Println(users)
			s := env.NewSpreadsheet([]string{"User", "Terminal", "Host", "Started"})
			for _, user := range users {
				vals := []any{
					*env.NewString(user.User),
					*env.NewString(user.Terminal),
					*env.NewString(user.Host),
					*env.NewInteger(int64(user.Started)),
				}
				s.AddRow(*env.NewSpreadsheetRow(vals, s))
			}
			return *s
		},
	},
	"load-avg?": {
		Argsn: 0,
		Doc:   "Get the load average as a dict representing load average over the last 1, 5, and 15 minutes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			v, err := load.Avg()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "load-avg?")
			}
			r := env.NewDict(make(map[string]any, 3))
			r.Data["1"] = *env.NewDecimal(v.Load1)
			r.Data["5"] = *env.NewDecimal(v.Load5)
			r.Data["15"] = *env.NewDecimal(v.Load15)
			return *r
		},
	},
	"virtual-memory?": {
		Argsn: 0,
		Doc:   "Get information about virtual memory usage.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			v, err := mem.VirtualMemory()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "virtual-memory?")
			}
			r := env.NewDict(make(map[string]any, 3))
			r.Data["total"] = *env.NewInteger(int64(v.Total))
			r.Data["free"] = *env.NewInteger(int64(v.Free))
			r.Data["used-percent"] = *env.NewDecimal(v.UsedPercent)
			return *r
		},
	},
	"disk-usage?": {
		Argsn: 0,
		Doc:   "Get disk usage information as a spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			partitions, err := disk.Partitions(true)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "disk-usage?")
			}
			s := env.NewSpreadsheet([]string{"Filesystem", "Size", "Used", "Available", "Capacity", "iused", "ifree", "%iused", "Mounted on"})
			for _, partition := range partitions {
				usage, err := disk.Usage(partition.Mountpoint)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "disk-usage?")
				}
				vals := []any{
					*env.NewString(partition.Device),
					*env.NewInteger(int64(usage.Total)),
					*env.NewInteger(int64(usage.Used)),
					*env.NewInteger(int64(usage.Free)),
					*env.NewDecimal(float64(usage.UsedPercent)),
					*env.NewInteger(int64(usage.InodesUsed)),
					*env.NewInteger(int64(usage.InodesFree)),
					*env.NewInteger(int64(usage.InodesUsedPercent)),
					*env.NewString(usage.Path),
				}
				s.AddRow(*env.NewSpreadsheetRow(vals, s))
			}
			return *s
		},
	},
	"pids?": {
		Argsn: 0,
		Doc:   "Get process pids as a block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pids, err := process.Pids()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "pids?")
			}

			pids2 := make([]env.Object, len(pids))
			for i, p := range pids {
				pids2[i] = env.NewInteger(int64(p))
			}
			return *env.NewBlock(*env.NewTSeries(pids2))
		},
	},
	"processes?": {
		Argsn: 0,
		Doc:   "Get information about all processes as a spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			processes, err := process.Processes()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "processes?")
			}
			s := proccesSpreadsheetBase()
			for _, process := range processes {
				processSpreadsheetAdd(s, process)
			}
			return *s
		},
	},
	"process": {
		Argsn: 1,
		Doc:   "Get information about process with a given PID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pid := arg0.(type) {
			case env.Integer:
				process, err := process.NewProcess(int32(pid.Value))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "process")
				}
				s := proccesSpreadsheetBase()
				processSpreadsheetAdd(s, process)
				return s.Rows[0].ToDict()
			default:
				return *MakeArgError(ps, 1, []env.Type{env.IntegerType}, "process")
			}
		},
	},
}

func proccesSpreadsheetBase() *env.Spreadsheet {
	return env.NewSpreadsheet([]string{
		"User",
		"PID",
		"Status",
		"%CPU",
		"%MEM",
		"VZS",
		"RSS",
		"Num Threads",
		"Num FDs",
		"Num Open Files",
		"Num Connections",
		"Started at",
		"CPU Time",
		"Command",
	})
}

func processSpreadsheetAdd(s *env.Spreadsheet, process *process.Process) {
	var status env.String
	stat, err := process.Status()
	if err == nil {
		status = *env.NewString(strings.Join(stat, " "))
	} else {
		status = *env.NewString("???")
	}

	var vzs env.Object
	var rss env.Object
	memInfo, err := process.MemoryInfo()
	if err == nil {
		vzs = *env.NewInteger(int64(memInfo.VMS))
		rss = *env.NewInteger(int64(memInfo.RSS))
	} else {
		vzs = *env.NewString("???")
		rss = *env.NewString("???")
	}

	var numOpenFiles env.Object
	openFiles, err := process.OpenFiles()
	if err == nil {
		numOpenFiles = *env.NewInteger(int64(len(openFiles)))
	} else {
		numOpenFiles = *env.NewString("???")
	}

	var numConnections env.Object
	connections, err := process.Connections()
	if err == nil {
		numConnections = *env.NewInteger(int64(len(connections)))
	} else {
		numConnections = *env.NewString("???")
	}

	var startedAt env.Object
	createTime, err := process.CreateTime()
	if err == nil {
		startedAt = *env.NewDate(time.UnixMilli(createTime))
	} else {
		startedAt = *env.NewString("???")
	}

	var cpuTime env.Object
	times, err := process.Times()
	if err == nil {
		dur := time.Duration(times.User+times.System) * time.Second
		cpuTime = *env.NewString(fmt.Sprintf("%02d:%02d.%02d", int(dur.Minutes()), int(dur.Seconds())%60, int(dur.Milliseconds())%1000))
	} else {
		cpuTime = *env.NewString("???")
	}

	vals := []any{
		maybeString(process.Username),
		process.Pid,
		status,
		maybeFloat64(process.CPUPercent),
		maybeFloat32(process.MemoryPercent),
		vzs,
		rss,
		maybeInt32(process.NumThreads),
		maybeInt32(process.NumFDs),
		numOpenFiles,
		numConnections,
		startedAt,
		cpuTime,
		maybeString(process.Cmdline),
	}
	s.AddRow(*env.NewSpreadsheetRow(vals, s))
}

func maybeString(f func() (string, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewString(s)
}

func maybeFloat64(f func() (float64, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewDecimal(s)
}

func maybeFloat32(f func() (float32, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewDecimal(float64(s))
}

func maybeInt32(f func() (int32, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewInteger(int64(s))
}
