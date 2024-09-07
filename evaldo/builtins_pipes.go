//go:build !no_pipes
// +build !no_pipes

package evaldo

import (
	"regexp"

	"github.com/refaktor/rye/env"

	"github.com/bitfield/script"
)

var Builtins_pipes = map[string]*env.Builtin{

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

	"wcl": {
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

	// GOPSUTIL

}
