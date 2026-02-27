//go:build !no_io
// +build !no_io

package evaldo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/refaktor/rye/env"

	"net/http"
	//	"net/http/cgi"

	"github.com/jlaffaye/ftp"
)

var Builtins_io = map[string]*env.Builtin{

	//
	// ##### Console IO ##### "Console input and output functions"
	//
	// Example:
	//  ; File operations
	//  file: Create %output.txt
	//  file .Write "Hello, World!\n"
	//  file .Close
	//  Read %output.txt |print
	//
	//  ; Reader/Writer operations
	//  reader: Reader %data.txt
	//  content: reader .Read\string
	//  reader .Close
	//  print content
	//
	//  ; HTTP requests
	//  Get https://api.example.com/data |print
	//  Post https://api.example.com/data "{\"name\":\"test\"}" 'json |print
	//
	// Args:
	// * prompt: string to display as a prompt
	// Returns:
	// * string containing the user's input
	"input": {
		Argsn: 1,
		Doc:   "Prompts for and reads user input from the console.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				fmt.Print("" + str.Value)
				var input string
				fmt.Scanln(&input)
				// fmt.Print(input)
				/* reader := bufio.NewReader(os.Stdin)
				fmt.Print(str)
				inp, _ := reader.ReadString('\n')
				fmt.Println(inp) */
				return *env.NewString(input)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "input")
			}
		},
	},

	//
	// ##### File Operations ##### "File system operations and file manipulation"
	//
	// Tests:
	// equal { Open %data/file.txt |type? } 'native
	// equal { Open %data/file.txt |kind? } 'file
	// Args:
	// * path: uri representing the file to open
	// Returns:
	// * native file object
	"file-uri//Open": {
		Argsn: 1,
		Doc:   "Opens a file for reading.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.Open(s.Path)
				if err != nil {
					return makeError(ps, err.Error())
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Open")
			}
		},
	},

	// Tests:
	// equal { Open\append %data/file.txt |type? } 'native
	// equal { Open\append %data/file.txt |kind? } 'file
	// Args:
	// * path: uri representing the file to open for appending
	// Returns:
	// * native writer object
	"file-uri//Open\\append": {
		Argsn: 1,
		Doc:   "Opens a file for appending.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "file append operation blocked by readonly seccomp profile", "file-uri//Open\\append")
			}

			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.OpenFile(s.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "__openFile")
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "__openFile")
			}
		},
	},

	// Tests:
	// equal { Create %data/created.txt |type? } 'native
	// equal { Create %data/created.txt |kind? } 'file
	// Args:
	// * path: uri representing the file to create
	// Returns:
	// * native file object
	"file-uri//Create": {
		Argsn: 1,
		Doc:   "Creates a new file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "file creation blocked by readonly seccomp profile", "file-uri//Create")
			}

			switch s := arg0.(type) {
			case env.Uri:
				// path := strings.Split(s.Path, "://")
				file, err := os.Create(s.Path)
				if err != nil {
					ps.ReturnFlag = true
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "__create")
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				ps.ReturnFlag = true
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "__create")
			}
		},
	},

	// Tests:
	// equal { File-ext? %data/file.txt } ".txt"
	// equal { File-ext? %data/file.temp.png } ".png"
	// equal { File-ext? to-file "data/file.temp.png" } ".png"
	// Args:
	// * path: uri or string representing a file path
	// Returns:
	// * string containing the file extension (including the dot)
	"file-uri//File-ext?": {
		Argsn: 1,
		Doc:   "Gets the extension of a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				ext := filepath.Ext(s.Path)
				return *env.NewString(ext)
			/* case env.String:
			ext := filepath.Ext(s.Value)
			return *env.NewString(ext) */
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "file-ext?")
			}
		},
	},

	// Args:
	// * path: uri representing a file path
	// Returns:
	// * string containing the filename with extension
	"file-uri//Filename?": {
		Argsn: 1,
		Doc:   "Gets the filename with extension from a file path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				filename := filepath.Base(s.Path)
				return *env.NewString(filename)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Filename?")
			}
		},
	},

	// Args:
	// * path: uri representing a file path
	// Returns:
	// * string containing the filename without extension
	"file-uri//Stem?": {
		Argsn: 1,
		Doc:   "Gets the filename without extension from a file path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				base := filepath.Base(s.Path)
				stem := strings.TrimSuffix(base, filepath.Ext(base))
				return *env.NewString(stem)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Stem?")
			}
		},
	},

	// Args:
	// * path: uri representing a file path
	// Returns:
	// * string containing the directory path
	"file-uri//Dir?": {
		Argsn: 1,
		Doc:   "Gets the directory path from a file path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				dir := filepath.Dir(s.Path)
				return *env.NewString(dir)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Dir?")
			}
		},
	},

	// Args:
	// * path: uri representing a file path
	// Returns:
	// * block of strings containing path components
	"file-uri//Split": {
		Argsn: 1,
		Doc:   "Splits a file path into its components.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				pathStr := s.Path
				var components []env.Object

				// Split path and filter out empty components
				parts := strings.Split(pathStr, string(filepath.Separator))
				for _, part := range parts {
					if part != "" {
						components = append(components, *env.NewString(part))
					}
				}

				// Handle absolute paths - add root separator as first component
				if filepath.IsAbs(pathStr) && len(components) > 0 {
					components = append([]env.Object{*env.NewString(string(filepath.Separator))}, components...)
				}

				return *env.NewBlock(*env.NewTSeries(components))
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Split")
			}
		},
	},

	// Args:
	// * path: uri representing a file path
	// Returns:
	// * boolean indicating whether the path is absolute
	"file-uri//Is-absolute": {
		Argsn: 1,
		Doc:   "Checks if a file path is absolute.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				isAbs := filepath.IsAbs(s.Path)
				return *env.NewBoolean(isAbs)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Is-absolute")
			}
		},
	},

	// Tests:
	// equal { Does-exist %data/file.txt } true
	// equal { Does-exist %data/nonexistent.txt } false
	// Args:
	// * path: uri representing a file path
	// Returns:
	// * boolean indicating whether the file exists
	"file-uri//Does-exist": {
		Argsn: 1,
		Doc:   "Checks if a file exists.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				_, err := os.Stat(s.Path)
				if err == nil {
					return *env.NewBoolean(true)
				}
				if os.IsNotExist(err) {
					return *env.NewBoolean(false)
				}
				// For other errors (e.g., permission denied), return false
				return *env.NewBoolean(false)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Does-exist")
			}
		},
	},

	// Args:
	// * path: uri representing a file path to delete
	// Returns:
	// * true if the file was successfully deleted
	"file-uri//Delete": {
		Argsn: 1,
		Doc:   "Deletes a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "delete operation blocked by readonly seccomp profile", "file-uri//Delete")
			}

			switch s := arg0.(type) {
			case env.Uri:
				err := os.Remove(s.Path)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "file-uri//Delete")
				}
				return *env.NewBoolean(true)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Delete")
			}
		},
	},
	// Tests:
	// equal { Reader Open %data/file.txt |kind? } 'reader
	// Args:
	// * source: file object to read from
	// Returns:
	// * native reader object
	"file//Reader": {
		Argsn: 1,
		Doc:   "Creates a new reader from file object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				file, ok := s.Value.(*os.File)
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error opening file.", "file//Reader")
				}
				return *env.NewNative(ps.Idx, bufio.NewReader(file), "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Reader")
			}
		},
	},
	// Tests:
	// equal { Writer Open %data/file.txt |kind? } 'writer
	// Args:
	// * source: file object to read from
	// Returns:
	// * native reader object
	"file//Writer": {
		Argsn: 1,
		Doc:   "Creates a new reader from file object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				file, ok := s.Value.(*os.File)
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error opening file.", "file//Reader")
				}
				return *env.NewNative(ps.Idx, bufio.NewWriter(file), "writer")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Reader")
			}
		},
	},
	// Tests:
	// equal { Reader %data/file.txt |kind? } 'reader
	// Args:
	// * source: file uri to read from
	// Returns:
	// * native reader object
	"file-uri//Reader": {
		Argsn: 1,
		Doc:   "Creates a new reader from a file uri/path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.Open(s.Path)
				//trace3(path)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error opening file.", "file-uri//Reader")
				}
				return *env.NewNative(ps.Idx, bufio.NewReader(file), "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Read")
			}
		},
	},
	// Tests:
	// equal { reader "some string" |kind? } 'reader
	// Args:
	// * source: string to read from
	// Returns:
	// * native reader object
	"reader": {
		Argsn: 1,
		Doc:   "Creates a new reader from a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				return *env.NewNative(ps.Idx, bufio.NewReader(strings.NewReader(s.Value)), "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "reader")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * native reader object connected to standard input
	"stdin": {
		Argsn: 0,
		Doc:   "Gets a reader for standard input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, os.Stdin, "reader")
		},
	},

	// Args:
	// * none
	// Returns:
	// * native writer object connected to standard output
	"stdout": {
		Argsn: 0,
		Doc:   "Gets a writer for standard output.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(env1.Idx, os.Stdout, "writer")
		},
	},
	//
	// Args:
	// * none
	// Returns:
	// * native writer object connected to standard error
	"stderr": {
		Argsn: 0,
		Doc:   "Gets a writer for standard error.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(env1.Idx, os.Stderr, "writer")
		},
	},

	// TODO: add scanner ScanString method ... look at: https://stackoverflow.com/questions/47479564/go-bufio-readstring-in-loop-is-infinite

	// Tests:
	// equal { reader "some string" |Read\string } "some string"
	// Args:
	// * reader: native reader object
	// Returns:
	// * string containing all content from the reader
	"reader//Read\\string": {
		Argsn: 1,
		Doc:   "Reads all content from a reader as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				reader, ok := r.Value.(io.Reader)
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Not Reader", "__read\\string")
				}
				buf := new(strings.Builder)
				_, err := io.Copy(buf, reader)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "__read\\string")
				}
				return *env.NewString(buf.String())
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__read\\string")
			}

		},
	},

	// Args:
	// * reader: native reader object
	// Returns:
	// * empty string if successful
	"reader//Close": {
		Argsn: 1,
		Doc:   "Closes a reader if it implements io.Closer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				// Check if the reader implements io.Closer
				if closer, ok := r.Value.(io.Closer); ok {
					err := closer.Close()
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "reader//Close")
					}
					return *env.NewString("")
				}
				// If the reader doesn't implement io.Closer, just return success
				// (e.g., readers from strings don't need closing)
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "reader//Close")
			}
		},
	},

	// Args:
	// * reader: native reader object
	// * writer: native writer object
	// Returns:
	// * the reader object if successful
	"reader//Copy": {
		Argsn: 2,
		Doc:   "Copies all content from a reader to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				switch w := arg1.(type) {
				case env.Native:
					// Writer , Reader
					_, err := io.Copy(w.Value.(io.Writer), r.Value.(io.Reader))
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__copy")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "__copy")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__copy")
			}

		},
	},

	// We have duplication reader file TODO think about this ... is it worth
	// changing how kinds work, making them more complex? not sure yet
	// Args:
	// * file: native file object
	// * writer: native writer object
	// Returns:
	// * the file object if successful
	"file//Copy": {
		Argsn: 2,
		Doc:   "Copies content from a file to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				switch w := arg1.(type) {
				case env.Native:
					// Writer , Reader
					_, err := io.Copy(w.Value.(io.Writer), r.Value.(io.Reader))
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__copy")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "__copy")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__copy")
			}

		},
	},

	// Tests:
	// equal { Stat Open %data/file.txt |kind? } 'file-info
	// Args:
	// * file: native file object
	// Returns:
	// * native file-info object
	"file//Stat": {
		Argsn: 1,
		Doc:   "Gets file information (stat) for a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				info, err := r.Value.(*os.File).Stat()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "file//Stat")
				}
				return *env.NewNative(ps.Idx, info, "file-info")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Stat")
			}
		},
	},

	// Tests:
	// equal { Size? Stat Open %data/file.txt } 16
	// Args:
	// * file-info: native file-info object
	// Returns:
	// * integer representing the file size in bytes
	"file-info//Size?": {
		Argsn: 1,
		Doc:   "Gets the size of a file in bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				size := s.Value.(os.FileInfo).Size()
				return *env.NewInteger(size)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file-info//Size?")
			}
		},
	},

	// Tests:
	// equal { Read Open %data/file.txt } "hello text file\n"
	// Args:
	// * file: native file object
	// Returns:
	// * string containing the entire file content
	"file//Read": {
		Argsn: 1,
		Doc:   "Reads the entire content of a file as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				data, err := io.ReadAll(s.Value.(io.Reader))
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error reading file.", "Read")
				}
				return *env.NewString(string(data))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Read")
			}
		},
	},

	// Args:
	// * file: native file object
	// Returns:
	// * the same file object with position set to end of file
	"file//Seek\\end": {
		Argsn: 1,
		Doc:   "Seeks to the end of a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				reader, ok := s.Value.(*os.File)
				if !ok {
					return MakeBuiltinError(ps, "Native not io.Reader", "file//Seek\\end")
				}
				reader.Seek(0, os.SEEK_END)
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Seek\\end")
			}
		},
	},

	// Args:
	// * file: native file object
	// * content: string to write to the file
	// Returns:
	// * the file object if successful (allows chaining)
	"file//Write": {
		Argsn: 2,
		Doc:   "Writes a string directly to a file object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "write operation blocked by readonly seccomp profile", "file//Write")
			}

			switch s := arg0.(type) {
			case env.Native:
				file, ok := s.Value.(*os.File)
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Native not os.File", "file//Write")
				}
				switch content := arg1.(type) {
				case env.String:
					_, err := file.WriteString(content.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "file//Write")
					}
					return arg0 // Return the file object for chaining
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "file//Write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Write")
			}
		},
	},

	// Tests:
	// equal { Close Open %data/file.txt } ""
	// Args:
	// * file: native file object
	// Returns:
	// * empty string if successful
	"file//Close": {
		Argsn: 1,
		Doc:   "Closes an open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				err := s.Value.(*os.File).Close()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "__close")
				}
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__close")
			}

		},
	},

	// Tests:
	// equal { Read %data/file.txt } "hello text file\n"
	// Args:
	// * path: uri representing the file to read
	// Returns:
	// * string containing the entire file content
	"file-uri//Read": {
		Argsn: 1,
		Doc:   "Reads the entire content of a file as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				data, err := os.ReadFile(f.GetPath())
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "file-uri//Read")
				}
				return *env.NewString(string(data))
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Read")
			}
		},
	},

	// Tests:
	// equal { Read %data/file.txt } "hello text file\n"
	// Args:
	// * path: uri representing the file to read
	// Returns:
	// * native bytes object containing the file content
	"file-uri//Read\\bytes": {
		Argsn: 1,
		Doc:   "Reads the entire content of a file as bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				data, err := os.ReadFile(f.GetPath())
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "__fs_read_bytes")
				}
				return *env.NewNative(ps.Idx, data, "bytes")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "__fs_read_bytes")
			} // return __fs_read_bytes(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Tests:
	// equal { Read %data/file.txt } "hello text file\n"
	// Args:
	// * path: uri representing the file to read
	// Returns:
	// * block of strings, each representing a line from the file
	"file-uri//Read\\lines": {
		Argsn: 1,
		Doc:   "Reads a file and returns its content as a block of lines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				file, err := os.OpenFile(f.GetPath(), os.O_RDONLY, os.ModePerm)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "file-uri//Read\\lines")
				}
				defer file.Close()

				lines := make([]env.Object, 0)
				sc := bufio.NewScanner(file)
				for sc.Scan() {
					lines = append(lines, *env.NewString(sc.Text()))
				}
				if err := sc.Err(); err != nil {
					return MakeBuiltinError(ps, err.Error(), "file-uri//Read\\lines")
				}
				return *env.NewBlock(*env.NewTSeries(lines))
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Read\\lines")
			}
		},
	},

	// Tests:
	// equal { Write %data/write.txt "written\n" } "written\n"
	// Args:
	// * path: uri representing the file to write to
	// * content: string or bytes to write to the file
	// Returns:
	// * the content that was written
	"file-uri//Write": {
		Argsn: 2,
		Doc:   "Writes content to a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "write operation blocked by readonly seccomp profile", "file-uri//Write")
			}

			// If not in readonly mode, proceed with the original function
			switch f := arg0.(type) {
			case env.Uri:
				switch s := arg1.(type) {
				case env.String:
					err := os.WriteFile(f.GetPath(), []byte(s.Value), 0600)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__fs_write")
					}
					return arg1
				case env.Native:
					err := os.WriteFile(f.GetPath(), s.Value.([]byte), 0600)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__fs_write")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.NativeType}, "__fs_write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "__fs_write")
			}
		},
	},

	// TODO: make it generic of file schema
	// Args:
	// * bytes: Go-bytes native value to write
	// * path: string path to the file to write
	// Returns:
	// * integer 1 if successful
	"write\\bytes": {
		Argsn: 2,
		Doc:   "Writes bytes to a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bytesObj := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(bytesObj.GetKind()) != "Go-bytes" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "write-file")
				}
				switch path := arg1.(type) {
				case env.String:
					err := os.WriteFile(path.Value, bytesObj.Value.([]byte), 0644)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to write file: %v", err), "write-file")
					}
					return env.Integer{1} // Success indicator
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "write-file")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "write-file")
			}
		},
	},

	// Args:
	// * bytes1: first Go-bytes native value
	// * bytes2: second Go-bytes native value
	// Returns:
	// * combined bytes as a native bytes object
	"append\\bytes": {
		Argsn: 2,
		Doc:   "Appends two byte arrays into one.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bytes1 := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(bytes1.GetKind()) != "bytes" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "append-bytes")
				}
				switch bytes2 := arg1.(type) {
				case env.Native:
					if ps.Idx.GetWord(bytes2.GetKind()) != "bytes" {
						ps.FailureFlag = true
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "append-bytes")
					}
					combined := append(bytes1.Value.([]byte), bytes2.Value.([]byte)...)
					return *env.NewNative(ps.Idx, combined, "bytes")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "append-bytes")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "append-bytes")
			}
		},
	},

	// Args:
	// * writer: native writer object
	// * content: string to write
	// Returns:
	// * the writer object if successful
	"writer//Write": {
		Argsn: 2,
		Doc:   "Writes a string to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				// Allow writing to stdout/stderr but block writing to files
				switch ww := arg0.(type) {
				case env.Native:
					_, ok := ww.Value.(io.Writer)
					if !ok {
						return MakeBuiltinError(ps, "Native not io.Writer", "writer//Write")
					}
				}
			}

			switch s := arg1.(type) {
			case env.String:
				switch ww := arg0.(type) {
				case env.Native:
					// Try bufio.Writer first for WriteString support
					if bw, ok := ww.Value.(*bufio.Writer); ok {
						_, err := bw.WriteString(s.Value)
						if err != nil {
							return MakeBuiltinError(ps, "Error at write: "+err.Error(), "writer//Write")
						}
						return arg0
					}
					// Fall back to io.Writer interface (handles *os.File, etc.)
					if w, ok := ww.Value.(io.Writer); ok {
						_, err := io.WriteString(w, s.Value)
						if err != nil {
							return MakeBuiltinError(ps, "Error at write: "+err.Error(), "writer//Write")
						}
						return arg0
					}
					return MakeBuiltinError(ps, "Native not io.Writer", "writer//Write")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "writer//Write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "writer//Write")
			}
		},
	},

	/*
		"file-uri//Open": {
			Argsn: 1,
			Doc:   "Open a file, get a reader",
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch f := arg0.(type) {
				case env.Uri:
					file, err := os.Open(s.Path)
					//trace3(path)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Error opening file.", "file-uri//open")
					}
					return *env.NewNative(ps.Idx, bufio.NewReader(file), "file-uri//open")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file-uri//open")
				}
			},
		}, */

	// ---

	//
	// ##### HTTPs Operations ##### "Web requests and HTTP protocol functions"
	//
	// Args:
	// * url: uri representing the HTTPS URL to request
	// Returns:
	// * native reader object for the response body
	"https-uri//Open": {
		Argsn: 1,
		Doc:   "Opens a HTTPS GET request and returns a reader for the response body.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
				// defer cancel()
				proto := ps.Idx.GetWord(f.GetProtocol().Index)
				// req, err := http.NewRequestWithContext(ctx, http.MethodGet, proto+"://"+f.GetPath(), nil)
				req, err := http.NewRequest(http.MethodGet, proto+"://"+f.GetPath(), nil)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				// Print the HTTP Status Code and Status Name
				//mt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				// defer resp.Body.Close()
				// body, _ := io.ReadAll(resp.Body)

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewNative(ps.Idx, resp.Body, "https-uri://open")
				} else {
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v", resp.StatusCode)
					return MakeBuiltinError(ps, errMsg, "https-uri://Open")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-uri://Open")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTPS URL to request
	// Returns:
	// * string containing the response body
	"https-uri//Get": {
		Argsn: 1,
		Doc:   "Makes a HTTPS GET request and returns the response body as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				proto := ps.Idx.GetWord(f.GetProtocol().Index)
				req, err := http.NewRequest(http.MethodGet, proto+"://"+f.GetPath(), nil)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewString(string(body))
				} else {
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v", resp.StatusCode)
					return MakeBuiltinError(ps, errMsg, "https-uri//Get")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Get")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTPS URL to request
	// * data: string containing the request body
	// * content-type: word specifying the content type (e.g., 'json', 'text')
	// Returns:
	// * string containing the response body
	"https-uri//Post": {
		Argsn: 3,
		Doc:   "Makes a HTTPS POST request and returns the response body as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				switch d := arg1.(type) {
				case env.String:
					var tt string

					// Handle third argument - content-type
					switch t := arg2.(type) {
					case env.Word:
						// Existing behavior: map word to content-type
						tidx, terr := ps.Idx.GetIndex("json")
						tidx2, terr2 := ps.Idx.GetIndex("text")
						tidx3, terr3 := ps.Idx.GetIndex("urlencoded")
						tidx4, terr4 := ps.Idx.GetIndex("multipart")
						if terr && t.Index == tidx {
							tt = "application/json"
						} else if terr2 && t.Index == tidx2 {
							tt = "text/plain"
						} else if terr3 && t.Index == tidx3 {
							tt = "application/x-www-form-urlencoded"
						} else if terr4 && t.Index == tidx4 {
							tt = "multipart/form-data"
						} else {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "Wrong content type.", "https-uri//Post")
						}
					case env.String:
						// New behavior: use string directly as content-type
						tt = t.Value
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.WordType, env.StringType}, "https-uri//Post")
					}

					proto := ps.Idx.GetWord(f.GetProtocol().Index)
					req, err := http.NewRequest(http.MethodPost, proto+"://"+f.GetPath(), strings.NewReader(d.Value))
					if err != nil {
						ps.FailureFlag = true
						return *env.NewError(err.Error())
					}
					req.Header.Set("Content-Type", tt)
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						ps.FailureFlag = true
						return *env.NewError(err.Error())
					}
					defer resp.Body.Close()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "https-uri//Post")
					}

					if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
						return *env.NewString(string(body))
					} else {
						ps.FailureFlag = true
						return env.NewError2(resp.StatusCode, string(body))
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "https-uri//Post")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Post")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTP URL to request
	// Returns:
	// * string containing the response body
	"http-uri//Get": {
		Argsn: 1,
		Doc:   "Makes a HTTP GET request and returns the response body as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				proto := ps.Idx.GetWord(f.GetProtocol().Index)
				req, err := http.NewRequest(http.MethodGet, proto+"://"+f.GetPath(), nil)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewString(string(body))
				} else {
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v", resp.StatusCode)
					return MakeBuiltinError(ps, errMsg, "http-uri//Get")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Get")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTP URL to request
	// * data: string containing the request body
	// * content-type: word specifying the content type (e.g., 'json', 'text')
	// Returns:
	// * string containing the response body
	"http-uri//Post": {
		Argsn: 3,
		Doc:   "Makes a HTTP POST request and returns the response body as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				switch t := arg2.(type) {
				case env.Word:
					switch d := arg1.(type) {
					case env.String:
						var tt string
						tidx, terr := ps.Idx.GetIndex("json")
						tidx2, terr2 := ps.Idx.GetIndex("text")
						if terr && t.Index == tidx {
							tt = "application/json"
						} else if terr2 && t.Index == tidx2 {
							tt = "text/plain"
						} else {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "Wrong content type.", "http-uri//Post")
						}

						proto := ps.Idx.GetWord(f.GetProtocol().Index)
						req, err := http.NewRequest(http.MethodPost, proto+"://"+f.GetPath(), strings.NewReader(d.Value))
						if err != nil {
							ps.FailureFlag = true
							return *env.NewError(err.Error())
						}
						req.Header.Set("Content-Type", tt)
						resp, err := http.DefaultClient.Do(req)
						if err != nil {
							ps.FailureFlag = true
							return *env.NewError(err.Error())
						}
						defer resp.Body.Close()
						body, err := io.ReadAll(resp.Body)
						if err != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, err.Error(), "http-uri//Post")
						}

						if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
							return *env.NewString(string(body))
						} else {
							ps.FailureFlag = true
							return env.NewError2(resp.StatusCode, string(body))
						}
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "http-uri//Post")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 3, []env.Type{env.WordType}, "http-uri//Post")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Post")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTPS URL to request
	// * method: word specifying the HTTP method (e.g., 'GET', 'POST', 'PUT', 'DELETE')
	// * data: string containing the request body
	// Returns:
	// * native https-request object
	"https-uri//Request": {
		Argsn: 3,
		Doc:   "Creates a new HTTPS request object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				switch method := arg1.(type) {
				case env.Word:
					method1 := ps.Idx.GetWord(method.Index)
					if !(method1 == "GET" || method1 == "POST" || method1 == "PUT" || method1 == "DELETE") {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Wrong method.", "https-uri//Request")
					}
					switch data := arg2.(type) {
					case env.String:
						data1 := strings.NewReader(data.Value)
						proto := ps.Idx.GetWord(uri.GetProtocol().Index)
						req, err := http.NewRequest(method1, proto+"://"+uri.GetPath(), data1)
						if err != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, err.Error(), "https-uri//Request")
						}
						return *env.NewNative(ps.Idx, req, "https-request")
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "https-uri//Request")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "https-uri//Request")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Request")
			}
		},
	},

	// Args:
	// * request: native https-request object
	// * name: word representing the header name
	// * value: string containing the header value
	// Returns:
	// * the request object if successful
	"https-request//Header!": {
		Argsn: 3,
		Doc:   "Sets a header on a HTTPS request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch method := arg1.(type) {
				case env.Word:
					name := ps.Idx.GetWord(method.Index)
					switch data := arg2.(type) {
					case env.String:
						req.Value.(*http.Request).Header.Set(name, data.Value)
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "https-request//Header!")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "https-request//Header!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-request//Header!")
			}
		},
	},

	// Args:
	// * request: native https-request object
	// * username: string containing the username
	// * password: string containing the password
	// Returns:
	// * the request object if successful
	"https-request//Basic-auth!": {
		Argsn: 3,
		Doc:   "Sets Basic Authentication on a HTTPS request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch username := arg1.(type) {
				case env.String:
					switch password := arg2.(type) {
					case env.String:
						req.Value.(*http.Request).SetBasicAuth(username.Value, password.Value)
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "https-request//Basic-auth!")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "https-request//Basic-auth!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-request//Basic-auth!")
			}
		},
	},

	// Args:
	// * request: native https-request object
	// Returns:
	// * native https-response object (always returns response regardless of status code)
	"https-request//Call": {
		Argsn: 1,
		Doc:   "Executes a HTTPS request and returns the response object. Always returns the response regardless of status code (200, 404, 500, etc.) - use Status? to check the code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				client := &http.Client{}
				resp, err := client.Do(req.Value.(*http.Request))
				// defer resp.Body.Close() // TODO -- comment this and figure out goling bodyclose
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "https-request//Call")
				}
				// Always return the response object regardless of status code
				// Users can check the status code using https-response//Status?
				return *env.NewNative(ps.Idx, resp, "https-response")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-request//Call")
			}
		},
	},

	// Args:
	// * response: native https-response object
	// Returns:
	// * native reader object for the response body
	"https-response//Reader": {
		Argsn: 1,
		Doc:   "Gets a reader for the HTTPS response body that can be used with io.Copy.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				// Return the Body field, which implements io.Reader, not the entire Response
				return *env.NewNative(ps.Idx, resp.Value.(*http.Response).Body, "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-response//Reader")
			}
		},
	},

	// Args:
	// * response: native https-response object
	// Returns:
	// * string containing the response body
	"https-response//Read-body": {
		Argsn: 1,
		Doc:   "Reads the body of a HTTPS response as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				data, err := io.ReadAll(resp.Value.(*http.Response).Body)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "https-response//Read-body")
				}
				return *env.NewString(string(data))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-response//Read-body")
			}
		},
	},

	//
	// ##### Email Operations ##### "Email sending and SMTP communication"
	//
	// Args:
	// * to: email address to send to
	// * message: string containing the email message
	// Returns:
	// * integer 1 if successful
	"email//Send": {
		Argsn: 2,
		Doc:   "Sends an email to the specified address.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch to_ := arg0.(type) {
			case env.Email:
				switch msg := arg1.(type) {
				case env.String:
					idx, _ := ps.Idx.GetIndex("user-profile")
					uctx_, _ := ps.Ctx.Get(idx)
					uctx := uctx_.(*env.RyeCtx)
					fmt.Println(to_)
					fmt.Println(msg)
					fmt.Println(uctx)
					// TODO continue: uncomment and make it work
					/*
						from, _ := uctx.Get(ps.Idx.GetIndex("smtp-from"))
						password, _ := uctx.Get(ps.Idx.GetIndex("smtp-password"))
						server, _ := uctx.Get(ps.Idx.GetIndex("smtp-server"))
						port, _ := uctx.Get(ps.Idx.GetIndex("smtp-port"))
						// Receiver email address.
						// to := []string{
						//	to_.Value,
						//}
						// Message.
						// message := []byte(msg.Value)
						m := gomail.NewMessage()

						// Set E-Mail sender
						m.SetHeader("From", from)

						// Set E-Mail receivers
						m.SetHeader("To", to_.Address)

						// Set E-Mail subject
						m.SetHeader("Subject", msg.Value)

						// Set E-Mail body. You can set plain text or html with text/html
						m.SetBody("text/plain", msg.Value)

						// Settings for SMTP server
						d := gomail.NewDialer(server, port, from, password)

						// This is only needed when SSL/TLS certificate is not valid on server.
						// In production this should be set to false.
						//			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

						// Now send E-Mail
						if err := d.DialAndSend(m); err != nil {
							ps.FailureFlag = true
							return env.NewError(err.Error())
						}
					*/
					return *env.NewInteger(1)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "email//Send")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.EmailType}, "email//Send")
			}
		},
	},

	//
	// ##### FTP Operations ##### "File Transfer Protocol operations and connections"
	//
	// Args:
	// * server: uri representing the FTP server to connect to
	// Returns:
	// * native ftp-connection object
	"ftp-uri//Open": {
		Argsn: 1,
		Doc:   "Opens a connection to an FTP server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Uri:
				conn, err := ftp.Dial(s.Path)
				if err != nil {
					fmt.Println("Error connecting to FTP server:", err)
					return MakeBuiltinError(ps, "Error connecting to FTP server: "+err.Error(), "ftp-uri//Open")
				}
				//trace3(path)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error opening file.", "ftp-uri//Open")
				}
				return *env.NewNative(ps.Idx, conn, "ftp-connection")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "ftp-uri//Open")
			}
		},
	},

	// Args:
	// * connection: native ftp-connection object
	// * username: string containing the username
	// * password: string containing the password
	// Returns:
	// * the connection object if successful
	"ftp-connection//Login": {
		Argsn: 3,
		Doc:   "Logs in to an FTP server connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				username, ok := arg1.(env.String)
				if !ok {
					// TODO ARG ERROR
					return nil
				}
				pwd, ok := arg2.(env.String)
				if !ok {
					// TODO ARG ERROR
					return nil
				}
				err := s.Value.(*ftp.ServerConn).Login(username.Value, pwd.Value)
				if err != nil {
					// TODO
					fmt.Println("Error logging in:", err)
					return nil
				}
				return s
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "ftp-connection//Login")
			}
		},
	},

	// Args:
	// * connection: native ftp-connection object
	// * path: string containing the path of the file to retrieve
	// Returns:
	// * native reader object for the retrieved file
	"ftp-connection//Retrieve": {
		Argsn: 2,
		Doc:   "Retrieves a file from an FTP server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				path, ok := arg1.(env.String)
				if !ok {
					// TODO ARG ERROR
				}
				resp, err := s.Value.(*ftp.ServerConn).Retr(path.Value)
				if err != nil {
					fmt.Println("Error retrieving:", err)
					return nil
				}
				return *env.NewNative(ps.Idx, resp, "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "ftp-connection//Retrieve")
			}
		},
	},

	//
	// ##### File Monitoring ##### "File watching and tailing operations"
	//
	// Args:
	// * path: uri or string representing the file to tail
	// * follow: boolean indicating whether to follow the file for new content
	// * reopen: boolean indicating whether to reopen the file if it's rotated
	// Returns:
	// * native tail-file object that can be used to read lines as they are added
	"tail-file": {
		Argsn: 3,
		Doc:   "Tails a file, following it for new content. Used for monitoring log files.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var filePath string

			// Get the file path from either a Uri or String
			switch path := arg0.(type) {
			case env.Uri:
				filePath = path.GetPath()
			case env.String:
				filePath = path.Value
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "tail-file")
			}

			// Get follow option
			follow := true // Default to true
			if arg1 != nil {
				switch f := arg1.(type) {
				case env.Boolean:
					follow = f.Value
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BooleanType}, "tail-file")
				}
			}

			// Get reopen option
			reopen := true // Default to true
			if arg2 != nil {
				switch r := arg2.(type) {
				case env.Boolean:
					reopen = r.Value
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 3, []env.Type{env.BooleanType}, "tail-file")
				}
			}

			// Create tail configuration
			config := tail.Config{
				Follow: follow,
				ReOpen: reopen,
			}

			// Tail the file
			t, err := tail.TailFile(filePath, config)
			if err != nil {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, err.Error(), "tail-file")
			}

			return *env.NewNative(ps.Idx, t, "tail-file")
		},
	},

	// Args:
	// * tail: native tail-file object
	// Returns:
	// * string containing the next line from the file, or nil if no more lines
	"tail-file//Read-line": {
		Argsn: 1,
		Doc:   "Reads the next line from a tailed file. Blocks until a line is available.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch t := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(t.GetKind()) != "tail-file" {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected tail-file object", "tail-file//Read-line")
				}

				// Get the next line from the tail
				line, ok := <-t.Value.(*tail.Tail).Lines
				if !ok {
					// Channel is closed, no more lines
					return *env.NewVoid()
				}

				return *env.NewString(line.Text)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tail-file//Read-line")
			}
		},
	},

	// Args:
	// * tail: native tail-file object
	// Returns:
	// * empty string if successful
	"tail-file//Close": {
		Argsn: 1,
		Doc:   "Closes a tailed file, stopping the monitoring.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch t := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(t.GetKind()) != "tail-file" {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected tail-file object", "tail-file//Close")
				}

				err := t.Value.(*tail.Tail).Stop()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "tail-file//Close")
				}

				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tail-file//Close")
			}
		},
	},
}
