//go:build !no_io
// +build !no_io

package baseio

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"gopkg.in/gomail.v2"

	"net/http"
	//	"net/http/cgi"

	"github.com/jlaffaye/ftp"
)

// consoleInputReader is shared so repeated input calls do not lose data
// buffered from stdin (for example when input is piped).
var consoleInputReader *bufio.Reader

func readConsoleLine() string {
	if consoleInputReader == nil {
		consoleInputReader = bufio.NewReader(os.Stdin)
	}
	line, err := consoleInputReader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return ""
	}
	return strings.TrimRight(line, "\r\n")
}

// fileReader keeps a reference to the underlying file so that reader//Close
// actually releases the file descriptor. A bare *bufio.Reader does not
// implement io.Closer, which previously made reader//Close a silent no-op.
type fileReader struct {
	*bufio.Reader
	file *os.File
}

func newFileReader(file *os.File) *fileReader {
	return &fileReader{Reader: bufio.NewReader(file), file: file}
}

func (r *fileReader) Close() error {
	return r.file.Close()
}

// fileWriter keeps a reference to the underlying file so that writer//Close can
// flush buffered data and then close the file. Without this, data written via
// file//Writer could remain in the buffer and be lost on exit.
type fileWriter struct {
	*bufio.Writer
	file *os.File
}

func newFileWriter(file *os.File) *fileWriter {
	return &fileWriter{Writer: bufio.NewWriter(file), file: file}
}

// Close flushes any buffered data before closing the underlying file. Flush
// errors take precedence because they indicate that data was not persisted.
func (w *fileWriter) Close() error {
	flushErr := w.Flush()
	closeErr := w.file.Close()
	if flushErr != nil {
		return flushErr
	}
	return closeErr
}

// underlyingFile returns the *os.File behind a plain file native or one of the
// buffered reader/writer wrappers.
func underlyingFile(value any) (*os.File, bool) {
	switch f := value.(type) {
	case *os.File:
		return f, true
	case *fileReader:
		return f.file, true
	case *fileWriter:
		return f.file, true
	default:
		return nil, false
	}
}

// isStandardStream reports whether f is one of the process standard streams,
// which must not be closed by writer//Close/reader//Close.
func isStandardStream(f *os.File) bool {
	return f == os.Stdin || f == os.Stdout || f == os.Stderr
}

func asHTTPRequest(value any) (*http.Request, bool) {
	req, ok := value.(*http.Request)
	return req, ok
}

func asHTTPResponse(value any) (*http.Response, bool) {
	resp, ok := value.(*http.Response)
	return resp, ok
}

func asFTPConnection(value any) (*ftp.ServerConn, bool) {
	conn, ok := value.(*ftp.ServerConn)
	return conn, ok
}

func asNetConnection(value any) (net.Conn, bool) {
	conn, ok := value.(net.Conn)
	return conn, ok
}

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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Prompts for and reads user input from the console.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				fmt.Print(str.Value)
				return *env.NewString(readConsoleLine())
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "input")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Opens a file for reading.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.Open(s.Path)
				if err != nil {
					return evaldo.MakeError(ps, err.Error())
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Open")
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
		Effect: env.EffOutput,
		Argsn:  1,
		Doc:    "Opens a file for appending.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "file append operation blocked by readonly seccomp profile", "file-uri//Open\\append")
			}

			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.OpenFile(s.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "__openFile")
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "__openFile")
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
		Effect: env.EffOutput,
		Argsn:  1,
		Doc:    "Creates a new file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "file creation blocked by readonly seccomp profile", "file-uri//Create")
			}

			switch s := arg0.(type) {
			case env.Uri:
				// path := strings.Split(s.Path, "://")
				file, err := os.Create(s.Path)
				if err != nil {
					ps.ReturnFlag = true
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "__create")
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				ps.ReturnFlag = true
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "__create")
			}
		},
	},

	// Tests:
	// equal { Ext? %data/file.txt } ".txt"
	// equal { Ext? %data/file.temp.png } ".png"
	// equal { Ext? file "data/file.temp.png" } ".png"
	// Args:
	// * path: uri or string representing a file path
	// Returns:
	// * string containing the file extension (including the dot)
	"file-uri//Ext?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the extension of a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				ext := filepath.Ext(s.Path)
				return *env.NewString(ext)
			case env.String:
				ext := filepath.Ext(s.Value)
				return *env.NewString(ext)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "Ext?")
			}
		},
	},

	// Tests:
	// equal { Name? %data/file.txt } "file.txt"
	// Args:
	// * path: uri representing a file path
	// Returns:
	// * string containing the filename with extension
	"file-uri//Name?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the filename with extension from a file path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				filename := filepath.Base(s.Path)
				return *env.NewString(filename)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "Name?")
			}
		},
	},

	// Tests:
	// equal { Stem? %data/file.txt } "file"
	// Args:
	// * path: uri representing a file path
	// Returns:
	// * string containing the filename without extension
	"file-uri//Stem?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the filename without extension from a file path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				base := filepath.Base(s.Path)
				stem := strings.TrimSuffix(base, filepath.Ext(base))
				return *env.NewString(stem)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Stem?")
			}
		},
	},

	// Tests:
	// equal { Dir? %data/file.txt } "data"
	// Args:
	// * path: uri representing a file path
	// Returns:
	// * string containing the directory path
	"file-uri//Dir?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the directory path from a file path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				dir := filepath.Dir(s.Path)
				return *env.NewString(dir)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Dir?")
			}
		},
	},

	// Tests:
	// equal { Split %data/file.txt } { "data" "file.txt" }
	// Args:
	// * path: uri representing a file path
	// Returns:
	// * block of strings containing path components
	"file-uri//Split": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Splits a file path into its components.",
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
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Split")
			}
		},
	},

	// Tests:
	// equal { Is-absolute %/tmp } true
	// equal { Is-absolute %data/file.txt } false
	// Args:
	// * path: uri representing a file path
	// Returns:
	// * boolean indicating whether the path is absolute
	"file-uri//Is-absolute": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Checks if a file path is absolute.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				isAbs := filepath.IsAbs(s.Path)
				return *env.NewBoolean(isAbs)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Is-absolute")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Checks if a file exists.",
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
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Does-exist")
			}
		},
	},

	// Args:
	// * path: uri representing a file path to delete
	// Returns:
	// * true if the file was successfully deleted
	"file-uri//Delete": {
		Effect: env.EffOutput,
		Argsn:  1,
		Doc:    "Deletes a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "delete operation blocked by readonly seccomp profile", "file-uri//Delete")
			}

			switch s := arg0.(type) {
			case env.Uri:
				err := os.Remove(s.Path)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "file-uri//Delete")
				}
				return *env.NewBoolean(true)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Delete")
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
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Creates a new reader from file object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				file, ok := s.Value.(*os.File)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error opening file.", "file//Reader")
				}
				return *env.NewNative(ps.Idx, newFileReader(file), "reader")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Reader")
			}
		},
	},
	// Tests:
	// equal { Writer Open %data/file.txt |kind? } 'writer
	// Args:
	// * source: file object to write to
	// Returns:
	// * native writer object
	"file//Writer": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Creates a new writer from a file object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				file, ok := s.Value.(*os.File)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error opening file.", "file//Writer")
				}
				return *env.NewNative(ps.Idx, newFileWriter(file), "writer")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Writer")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Creates a new reader from a file uri/path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.Open(s.Path)
				//trace3(path)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error opening file.", "file-uri//Reader")
				}
				return *env.NewNative(ps.Idx, newFileReader(file), "reader")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Read")
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
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Creates a new reader from a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				return *env.NewNative(ps.Idx, bufio.NewReader(strings.NewReader(s.Value)), "reader")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "reader")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * native reader object connected to standard input
	"stdin": {
		Effect: env.EffNone,
		Argsn:  0,
		Doc:    "Gets a reader for standard input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, os.Stdin, "reader")
		},
	},

	// Args:
	// * none
	// Returns:
	// * native writer object connected to standard output
	"stdout": {
		Effect: env.EffNone,
		Argsn:  0,
		Doc:    "Gets a writer for standard output.",
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
		Effect: env.EffNone,
		Argsn:  0,
		Doc:    "Gets a writer for standard error.",
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads all content from a reader as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				reader, ok := r.Value.(io.Reader)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Not Reader", "__read\\string")
				}
				buf := new(strings.Builder)
				_, err := io.Copy(buf, reader)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "__read\\string")
				}
				return *env.NewString(buf.String())
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "__read\\string")
			}

		},
	},

	// Tests:
	// equal { Reader %data/file.txt |Close |type? } 'string
	// Args:
	// * reader: native reader object
	// Returns:
	// * empty string if successful
	"reader//Close": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Closes a reader if it implements io.Closer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				// Check if the reader implements io.Closer
				if closer, ok := r.Value.(io.Closer); ok {
					// Never close the process standard streams through a reader.
					if f, isFile := underlyingFile(r.Value); isFile && isStandardStream(f) {
						return *env.NewString("")
					}
					err := closer.Close()
					if err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "reader//Close")
					}
					return *env.NewString("")
				}
				// If the reader doesn't implement io.Closer, just return success
				// (e.g., readers from strings don't need closing)
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "reader//Close")
			}
		},
	},

	// Args:
	// * reader: native reader object
	// * writer: native writer object
	// Returns:
	// * the reader object if successful
	"reader//Copy": {
		Effect: env.EffIO,
		Argsn:  2,
		Doc:    "Copies all content from a reader to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				switch w := arg1.(type) {
				case env.Native:
					reader, rOK := r.Value.(io.Reader)
					writer, wOK := w.Value.(io.Writer)
					if !rOK || !wOK {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "reader//Copy expects a reader and a writer", "reader//Copy")
					}
					_, err := io.Copy(writer, reader)
					if err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "reader//Copy")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "reader//Copy")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "reader//Copy")
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
		Effect: env.EffIO,
		Argsn:  2,
		Doc:    "Copies content from a file to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				switch w := arg1.(type) {
				case env.Native:
					reader, rOK := r.Value.(io.Reader)
					writer, wOK := w.Value.(io.Writer)
					if !rOK || !wOK {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "file//Copy expects a reader and a writer", "file//Copy")
					}
					_, err := io.Copy(writer, reader)
					if err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "file//Copy")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "file//Copy")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Copy")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Gets file information (stat) for a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				file, ok := underlyingFile(r.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected a file object", "file//Stat")
				}
				info, err := file.Stat()
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "file//Stat")
				}
				return *env.NewNative(ps.Idx, info, "file-info")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Stat")
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
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the size of a file in bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				info, ok := s.Value.(os.FileInfo)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected a file-info object", "file-info//Size?")
				}
				size := info.Size()
				return *env.NewInteger(size)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file-info//Size?")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads the entire content of a file as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				reader, ok := s.Value.(io.Reader)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected a reader object", "file//Read")
				}
				data, err := io.ReadAll(reader)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error reading file.", "file//Read")
				}
				return *env.NewString(string(data))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Read")
			}
		},
	},

	// Args:
	// * file: native file object
	// Returns:
	// * the same file object with position set to end of file
	"file//Seek\\end": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Seeks to the end of a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				file, ok := underlyingFile(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected a file object", "file//Seek\\end")
				}
				if _, err := file.Seek(0, io.SeekEnd); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "file//Seek\\end")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Seek\\end")
			}
		},
	},

	// Args:
	// * file: native file object
	// * content: string to write to the file
	// Returns:
	// * the file object if successful (allows chaining)
	"file//Write": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Writes a string directly to a file object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "write operation blocked by readonly seccomp profile", "file//Write")
			}

			switch s := arg0.(type) {
			case env.Native:
				var writer io.Writer
				if fw, ok := s.Value.(*fileWriter); ok {
					// Keep using the buffered writer so direct writes and buffered
					// writes preserve their order.
					writer = fw
				} else {
					file, ok := underlyingFile(s.Value)
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Expected a file object", "file//Write")
					}
					writer = file
				}
				switch content := arg1.(type) {
				case env.String:
					_, err := io.WriteString(writer, content.Value)
					if err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "file//Write")
					}
					return arg0 // Return the file object for chaining
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "file//Write")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Write")
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
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Closes an open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				closer, ok := s.Value.(io.Closer)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected a closable file object", "file//Close")
				}
				if f, isFile := underlyingFile(s.Value); isFile && isStandardStream(f) {
					return *env.NewString("")
				}
				err := closer.Close()
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "file//Close")
				}
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//Close")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads the entire content of a file as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				data, err := os.ReadFile(f.GetPath())
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "file-uri//Read")
				}
				return *env.NewString(string(data))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Read")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads the entire content of a file as bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				data, err := os.ReadFile(f.GetPath())
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "__fs_read_bytes")
				}
				return *env.NewBytes(data)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "__fs_read_bytes")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads a file and returns its content as a block of lines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				file, err := os.OpenFile(f.GetPath(), os.O_RDONLY, os.ModePerm)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "file-uri//Read\\lines")
				}
				defer file.Close()

				lines := make([]env.Object, 0)
				sc := bufio.NewScanner(file)
				for sc.Scan() {
					lines = append(lines, *env.NewString(sc.Text()))
				}
				if err := sc.Err(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "file-uri//Read\\lines")
				}
				return *env.NewBlock(*env.NewTSeries(lines))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Read\\lines")
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
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Writes content to a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "write operation blocked by readonly seccomp profile", "file-uri//Write")
			}

			// If not in readonly mode, proceed with the original function
			switch f := arg0.(type) {
			case env.Uri:
				var data []byte
				switch s := arg1.(type) {
				case env.String:
					data = []byte(s.Value)
				case env.Bytes:
					data = s.Value
				case env.Native:
					if ps.Idx.GetWord(s.GetKind()) != "Go-bytes" {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Expected a Go-bytes object", "file-uri//Write")
					}
					bytesValue, ok := s.Value.([]byte)
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Go-bytes object does not contain a byte slice", "file-uri//Write")
					}
					data = bytesValue
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType, env.BytesType, env.NativeType}, "file-uri//Write")
				}
				if err := os.WriteFile(f.GetPath(), data, 0644); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "file-uri//Write")
				}
				return arg1
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Write")
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
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Writes bytes to a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bytesObj := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(bytesObj.GetKind()) != "Go-bytes" {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "write\\bytes")
				}
				data, ok := bytesObj.Value.([]byte)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Go-bytes object does not contain a byte slice", "write\\bytes")
				}
				switch path := arg1.(type) {
				case env.String:
					err := os.WriteFile(path.Value, data, 0644)
					if err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Failed to write file: %v", err), "write\\bytes")
					}
					return env.Integer{1} // Success indicator
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "write\\bytes")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "write\\bytes")
			}
		},
	},

	// Args:
	// * bytes1: first Go-bytes native value
	// * bytes2: second Go-bytes native value
	// Returns:
	// * combined bytes as a native bytes object
	"append\\bytes": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Appends two byte arrays into one.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bytes1 := arg0.(type) {
			case env.Bytes:
				switch bytes2 := arg1.(type) {
				case env.Bytes:
					combined := append(bytes1.Value, bytes2.Value...)
					return *env.NewBytes(combined)
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.BytesType}, "append-bytes")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BytesType}, "append-bytes")
			}
		},
	},

	// Args:
	// * writer: native writer object
	// * content: string to write
	// Returns:
	// * the writer object if successful
	"writer//Write": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Writes a string to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if we're in readonly mode
			profile, exists := os.LookupEnv("RYE_SECCOMP_PROFILE")
			if exists && profile == "readonly" {
				// Allow writing to stdout/stderr but block file-backed writers.
				if ww, ok := arg0.(env.Native); ok {
					if f, isFile := underlyingFile(ww.Value); isFile {
						if !isStandardStream(f) {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, "write operation blocked by readonly seccomp profile", "writer//Write")
						}
					} else if _, isBuffered := ww.Value.(*bufio.Writer); isBuffered {
						// The backing destination of an arbitrary buffered writer is unknown.
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "write operation blocked by readonly seccomp profile", "writer//Write")
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
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, "Error at write: "+err.Error(), "writer//Write")
						}
						return arg0
					}
					// Fall back to io.Writer interface (handles *os.File, *fileWriter, etc.)
					if w, ok := ww.Value.(io.Writer); ok {
						_, err := io.WriteString(w, s.Value)
						if err != nil {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, "Error at write: "+err.Error(), "writer//Write")
						}
						return arg0
					}
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not io.Writer", "writer//Write")
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "writer//Write")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "writer//Write")
			}
		},
	},

	// Args:
	// * writer: native writer object
	// Returns:
	// * empty string if successful
	"writer//Flush": {
		Effect: env.EffOutput,
		Argsn:  1,
		Doc:    "Flushes buffered data to the underlying writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch w := arg0.(type) {
			case env.Native:
				if flusher, ok := w.Value.(interface{ Flush() error }); ok {
					if err := flusher.Flush(); err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "writer//Flush")
					}
				}
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "writer//Flush")
			}
		},
	},

	// Args:
	// * writer: native writer object
	// Returns:
	// * empty string if successful
	"writer//Close": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Flushes and closes a writer when it implements io.Closer. Standard streams are flushed but never closed.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch w := arg0.(type) {
			case env.Native:
				if f, isFile := underlyingFile(w.Value); isFile && isStandardStream(f) {
					if flusher, ok := w.Value.(interface{ Flush() error }); ok {
						if err := flusher.Flush(); err != nil {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, err.Error(), "writer//Close")
						}
					}
					return *env.NewString("")
				}
				if closer, ok := w.Value.(io.Closer); ok {
					if err := closer.Close(); err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "writer//Close")
					}
					return *env.NewString("")
				}
				if flusher, ok := w.Value.(interface{ Flush() error }); ok {
					if err := flusher.Flush(); err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "writer//Close")
					}
					return *env.NewString("")
				}
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "Native not io.Closer or flusher", "writer//Close")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "writer//Close")
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
						return evaldo.MakeBuiltinError(ps, "Error opening file.", "file-uri//open")
					}
					return *env.NewNative(ps.Idx, bufio.NewReader(file), "file-uri//open")
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "file-uri//open")
				}
			},
		}, */

	// ---

	//
	// ##### HTTPs Operations ##### "Web requests and HTTP protocol functions"
	//

	// Tests:
	// equal { Ext? uri "https://example.com/data/file.txt" } ".txt"
	// Args:
	// * url: uri representing the HTTPS URL
	// Returns:
	// * string containing the extension of the last path component
	"https-uri//Ext?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the extension of the last path component of an HTTPS URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				ext := path.Ext(s.Path)
				return *env.NewString(ext)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Ext?")
			}
		},
	},

	// Tests:
	// equal { Name? uri "https://example.com/data/file.txt" } "file.txt"
	// Args:
	// * url: uri representing the HTTPS URL
	// Returns:
	// * string containing the last path component with extension
	"https-uri//Name?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the last path component (filename) from an HTTPS URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				name := path.Base(s.Path)
				return *env.NewString(name)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Name?")
			}
		},
	},

	// Tests:
	// equal { Stem? uri "https://example.com/data/file.txt" } "file"
	// Args:
	// * url: uri representing the HTTPS URL
	// Returns:
	// * string containing the last path component without extension
	"https-uri//Stem?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the last path component (filename) without its extension from an HTTPS URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				base := path.Base(s.Path)
				stem := strings.TrimSuffix(base, path.Ext(base))
				return *env.NewString(stem)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Stem?")
			}
		},
	},

	// Tests:
	// equal { Dir? uri "https://example.com/data/file.txt" } "example.com/data"
	// Args:
	// * url: uri representing the HTTPS URL
	// Returns:
	// * string containing the directory portion of the URL path (including host)
	"https-uri//Dir?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the directory portion of an HTTPS URL path (including the host).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				dir := path.Dir(s.Path)
				return *env.NewString(dir)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Dir?")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTPS URL to request
	// Returns:
	// * native reader object for the response body
	"https-uri//Open": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Opens a HTTPS GET request and returns a reader for the response body.",
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
					return *env.NewNative(ps.Idx, resp.Body, "reader")
				} else {
					resp.Body.Close()
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v", resp.StatusCode)
					return evaldo.MakeBuiltinError(ps, errMsg, "https-uri://Open")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri://Open")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTPS URL to request
	// Returns:
	// * string containing the response body
	"https-uri//Get": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Makes a HTTPS GET request and returns the response body as a string.",
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
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "https-uri//Get")
				}

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewString(string(body))
				} else {
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v", resp.StatusCode)
					return evaldo.MakeBuiltinError(ps, errMsg, "https-uri//Get")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Get")
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
		Effect: env.EffOutput,
		Argsn:  3,
		Doc:    "Makes a HTTPS POST request and returns the response body as a string.",
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
							return evaldo.MakeBuiltinError(ps, "Wrong content type.", "https-uri//Post")
						}
					case env.String:
						// New behavior: use string directly as content-type
						tt = t.Value
					default:
						ps.FailureFlag = true
						return evaldo.MakeArgError(ps, 3, []env.Type{env.WordType, env.StringType}, "https-uri//Post")
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
						return evaldo.MakeBuiltinError(ps, err.Error(), "https-uri//Post")
					}

					if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
						return *env.NewString(string(body))
					} else {
						ps.FailureFlag = true
						return env.NewError2(resp.StatusCode, string(body))
					}
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "https-uri//Post")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Post")
			}
		},
	},

	// Tests:
	// equal { Ext? uri "http://example.com/data/file.txt" } ".txt"
	// Args:
	// * url: uri representing the HTTP URL
	// Returns:
	// * string containing the extension of the last path component
	"http-uri//Ext?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the extension of the last path component of an HTTP URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				ext := path.Ext(s.Path)
				return *env.NewString(ext)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Ext?")
			}
		},
	},

	// Tests:
	// equal { Name? uri "http://example.com/data/file.txt" } "file.txt"
	// Args:
	// * url: uri representing the HTTP URL
	// Returns:
	// * string containing the last path component with extension
	"http-uri//Name?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the last path component (filename) from an HTTP URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				name := path.Base(s.Path)
				return *env.NewString(name)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Name?")
			}
		},
	},

	// Tests:
	// equal { Stem? uri "http://example.com/data/file.txt" } "file"
	// Args:
	// * url: uri representing the HTTP URL
	// Returns:
	// * string containing the last path component without extension
	"http-uri//Stem?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the last path component (filename) without its extension from an HTTP URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				base := path.Base(s.Path)
				stem := strings.TrimSuffix(base, path.Ext(base))
				return *env.NewString(stem)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Stem?")
			}
		},
	},

	// Tests:
	// equal { Dir? uri "http://example.com/data/file.txt" } "example.com/data"
	// Args:
	// * url: uri representing the HTTP URL
	// Returns:
	// * string containing the directory portion of the URL path (including host)
	"http-uri//Dir?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the directory portion of an HTTP URL path (including the host).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				dir := path.Dir(s.Path)
				return *env.NewString(dir)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Dir?")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTP URL to request
	// Returns:
	// * native reader object for the response body
	"http-uri//Open": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Opens a HTTP GET request and returns a reader for the response body.",
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

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewNative(ps.Idx, resp.Body, "reader")
				} else {
					resp.Body.Close()
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v", resp.StatusCode)
					return evaldo.MakeBuiltinError(ps, errMsg, "http-uri://Open")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri://Open")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTP URL to request
	// Returns:
	// * string containing the response body
	"http-uri//Get": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Makes a HTTP GET request and returns the response body as a string.",
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
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "http-uri//Get")
				}

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewString(string(body))
				} else {
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v", resp.StatusCode)
					return evaldo.MakeBuiltinError(ps, errMsg, "http-uri//Get")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Get")
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
		Effect: env.EffOutput,
		Argsn:  3,
		Doc:    "Makes a HTTP POST request and returns the response body as a string.",
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
							return evaldo.MakeBuiltinError(ps, "Wrong content type.", "http-uri//Post")
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
							return evaldo.MakeBuiltinError(ps, err.Error(), "http-uri//Post")
						}

						if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
							return *env.NewString(string(body))
						} else {
							ps.FailureFlag = true
							return env.NewError2(resp.StatusCode, string(body))
						}
					default:
						ps.FailureFlag = true
						return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "http-uri//Post")
					}
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 3, []env.Type{env.WordType}, "http-uri//Post")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Post")
			}
		},
	},

	// Args:
	// * url: uri representing the HTTP URL to request
	// * method: word specifying the HTTP method (e.g., 'GET', 'POST', 'PUT', 'DELETE')
	// * data: string containing the request body
	// Returns:
	// * native http-request object
	"http-uri//Request": {
		Effect: env.EffIO,
		Argsn:  3,
		Doc:    "Creates a new HTTP request object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				switch method := arg1.(type) {
				case env.Word:
					method1 := ps.Idx.GetWord(method.Index)
					if !(method1 == "GET" || method1 == "POST" || method1 == "PUT" || method1 == "DELETE") {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Wrong method.", "http-uri//Request")
					}
					switch data := arg2.(type) {
					case env.String:
						data1 := strings.NewReader(data.Value)
						proto := ps.Idx.GetWord(uri.GetProtocol().Index)
						req, err := http.NewRequest(method1, proto+"://"+uri.GetPath(), data1)
						if err != nil {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, err.Error(), "http-uri//Request")
						}
						return *env.NewNative(ps.Idx, req, "http-request")
					default:
						ps.FailureFlag = true
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "http-uri//Request")
					}
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "http-uri//Request")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "http-uri//Request")
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
		Effect: env.EffIO,
		Argsn:  3,
		Doc:    "Creates a new HTTPS request object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				switch method := arg1.(type) {
				case env.Word:
					method1 := ps.Idx.GetWord(method.Index)
					if !(method1 == "GET" || method1 == "POST" || method1 == "PUT" || method1 == "DELETE") {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Wrong method.", "https-uri//Request")
					}
					switch data := arg2.(type) {
					case env.String:
						data1 := strings.NewReader(data.Value)
						proto := ps.Idx.GetWord(uri.GetProtocol().Index)
						req, err := http.NewRequest(method1, proto+"://"+uri.GetPath(), data1)
						if err != nil {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, err.Error(), "https-uri//Request")
						}
						return *env.NewNative(ps.Idx, req, "https-request")
					default:
						ps.FailureFlag = true
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "https-uri//Request")
					}
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "https-uri//Request")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "https-uri//Request")
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
		Effect: env.EffNone,
		Argsn:  3,
		Doc:    "Sets a header on a HTTPS request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch method := arg1.(type) {
				case env.Word:
					name := ps.Idx.GetWord(method.Index)
					switch data := arg2.(type) {
					case env.String:
						request, ok := asHTTPRequest(req.Value)
						if !ok {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, "Expected an https-request object", "https-request//Header!")
						}
						request.Header.Set(name, data.Value)
						return arg0
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "https-request//Header!")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "https-request//Header!")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-request//Header!")
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
		Effect: env.EffNone,
		Argsn:  3,
		Doc:    "Sets Basic Authentication on a HTTPS request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch username := arg1.(type) {
				case env.String:
					switch password := arg2.(type) {
					case env.String:
						request, ok := asHTTPRequest(req.Value)
						if !ok {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, "Expected an https-request object", "https-request//Basic-auth!")
						}
						request.SetBasicAuth(username.Value, password.Value)
						return arg0
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "https-request//Basic-auth!")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "https-request//Basic-auth!")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-request//Basic-auth!")
			}
		},
	},

	// Args:
	// * request: native https-request object
	// Returns:
	// * native https-response object (always returns response regardless of status code)
	"https-request//Call": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Executes a HTTPS request and returns the response object. Always returns the response regardless of status code (200, 404, 500, etc.) - use Status? to check the code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				request, ok := asHTTPRequest(req.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an https-request object", "https-request//Call")
				}
				client := &http.Client{}
				resp, err := client.Do(request)
				// defer resp.Body.Close() // TODO -- comment this and figure out goling bodyclose
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "https-request//Call")
				}
				// Always return the response object regardless of status code
				// Users can check the status code using https-response//Status?
				return *env.NewNative(ps.Idx, resp, "https-response")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-request//Call")
			}
		},
	},

	// Args:
	// * request: native http-request object
	// * name: word representing the header name
	// * value: string containing the header value
	// Returns:
	// * the request object if successful
	"http-request//Header!": {
		Effect: env.EffNone,
		Argsn:  3,
		Doc:    "Sets a header on a HTTP request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch method := arg1.(type) {
				case env.Word:
					name := ps.Idx.GetWord(method.Index)
					switch data := arg2.(type) {
					case env.String:
						request, ok := asHTTPRequest(req.Value)
						if !ok {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, "Expected an http-request object", "http-request//Header!")
						}
						request.Header.Set(name, data.Value)
						return arg0
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "http-request//Header!")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "http-request//Header!")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "http-request//Header!")
			}
		},
	},

	// Args:
	// * request: native http-request object
	// * username: string containing the username
	// * password: string containing the password
	// Returns:
	// * the request object if successful
	"http-request//Basic-auth!": {
		Effect: env.EffNone,
		Argsn:  3,
		Doc:    "Sets Basic Authentication on a HTTP request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch username := arg1.(type) {
				case env.String:
					switch password := arg2.(type) {
					case env.String:
						request, ok := asHTTPRequest(req.Value)
						if !ok {
							ps.FailureFlag = true
							return evaldo.MakeBuiltinError(ps, "Expected an http-request object", "http-request//Basic-auth!")
						}
						request.SetBasicAuth(username.Value, password.Value)
						return arg0
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "http-request//Basic-auth!")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "http-request//Basic-auth!")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "http-request//Basic-auth!")
			}
		},
	},

	// Args:
	// * request: native http-request object
	// Returns:
	// * native http-response object (always returns response regardless of status code)
	"http-request//Call": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Executes a HTTP request and returns the response object. Always returns the response regardless of status code (200, 404, 500, etc.) - use Status? to check the code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				request, ok := asHTTPRequest(req.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an http-request object", "http-request//Call")
				}
				client := &http.Client{}
				resp, err := client.Do(request)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "http-request//Call")
				}
				// Always return the response object regardless of status code
				// Users can check the status code using http-response//Status?
				return *env.NewNative(ps.Idx, resp, "http-response")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "http-request//Call")
			}
		},
	},

	// Args:
	// * response: native https-response object
	// Returns:
	// * native reader object for the response body
	"https-response//Reader": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Gets a reader for the HTTPS response body that can be used with io.Copy.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				// Return the Body field, which implements io.Reader, not the entire Response
				response, ok := asHTTPResponse(resp.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an https-response object", "https-response//Reader")
				}
				return *env.NewNative(ps.Idx, response.Body, "reader")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-response//Reader")
			}
		},
	},

	// Args:
	// * response: native https-response object
	// Returns:
	// * string containing the response body
	"https-response//Read-body": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads the body of a HTTPS response as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				response, ok := asHTTPResponse(resp.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an https-response object", "https-response//Read-body")
				}
				data, err := io.ReadAll(response.Body)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "https-response//Read-body")
				}
				return *env.NewString(string(data))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-response//Read-body")
			}
		},
	},

	// Args:
	// * response: native http-response object
	// Returns:
	// * native reader object for the response body
	"http-response//Reader": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Gets a reader for the HTTP response body that can be used with io.Copy.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				// Return the Body field, which implements io.Reader, not the entire Response
				response, ok := asHTTPResponse(resp.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an http-response object", "http-response//Reader")
				}
				return *env.NewNative(ps.Idx, response.Body, "reader")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "http-response//Reader")
			}
		},
	},

	// Args:
	// * response: native http-response object
	// Returns:
	// * string containing the response body
	"http-response//Read-body": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads the body of a HTTP response as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				response, ok := asHTTPResponse(resp.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an http-response object", "http-response//Read-body")
				}
				data, err := io.ReadAll(response.Body)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "http-response//Read-body")
				}
				return *env.NewString(string(data))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "http-response//Read-body")
			}
		},
	},

	// Args:
	// * response: native https-response object
	// Returns:
	// * empty string if close succeeds
	"https-response//Close-body!": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Closes the HTTPS response body to free network resources.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				response, ok := asHTTPResponse(resp.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an https-response object", "https-response//Close-body!")
				}
				err := response.Body.Close()
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "https-response//Close-body!")
				}
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-response//Close-body!")
			}
		},
	},

	// Args:
	// * response: native http-response object
	// Returns:
	// * empty string if close succeeds
	"http-response//Close-body!": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Closes the HTTP response body to free network resources.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch resp := arg0.(type) {
			case env.Native:
				response, ok := asHTTPResponse(resp.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected an http-response object", "http-response//Close-body!")
				}
				if err := response.Body.Close(); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "http-response//Close-body!")
				}
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "http-response//Close-body!")
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
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Sends an email to the specified address.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch to_ := arg0.(type) {
			case env.Email:
				switch msg := arg1.(type) {
				case env.String:
					profileIdx, ok := ps.Idx.GetIndex("user-profile")
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Missing user-profile context", "email//Send")
					}
					profileObj, ok := ps.Ctx.Get(profileIdx)
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Missing user-profile context", "email//Send")
					}
					profile, ok := profileObj.(*env.RyeCtx)
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "user-profile is not a context", "email//Send")
					}

					getString := func(name string) (string, bool) {
						wordIdx, ok := ps.Idx.GetIndex(name)
						if !ok {
							return "", false
						}
						value, ok := profile.Get(wordIdx)
						if !ok {
							return "", false
						}
						switch v := value.(type) {
						case env.String:
							return v.Value, true
						case env.Uri:
							return v.GetPath(), true
						case env.Word:
							return ps.Idx.GetWord(v.Index), true
						case env.Email:
							return v.Address, true
						default:
							return "", false
						}
					}

					from, ok := getString("smtp-from")
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "user-profile is missing 'smtp-from'", "email//Send")
					}
					password, _ := getString("smtp-password")
					server, ok := getString("smtp-server")
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "user-profile is missing 'smtp-server'", "email//Send")
					}

					port := 587
					if portIdx, ok := ps.Idx.GetIndex("smtp-port"); ok {
						if portObj, ok := profile.Get(portIdx); ok {
							switch v := portObj.(type) {
							case env.Integer:
								port = int(v.Value)
							case env.String:
								if _, err := fmt.Sscanf(v.Value, "%d", &port); err != nil {
									ps.FailureFlag = true
									return evaldo.MakeBuiltinError(ps, "'smtp-port' is not an integer", "email//Send")
								}
							default:
								ps.FailureFlag = true
								return evaldo.MakeBuiltinError(ps, "'smtp-port' is not an integer", "email//Send")
							}
						}
					}

					m := gomail.NewMessage()
					m.SetHeader("From", from)
					m.SetHeader("To", to_.Address)
					m.SetHeader("Subject", msg.Value)
					m.SetBody("text/plain", msg.Value)
					d := gomail.NewDialer(server, port, from, password)
					if err := d.DialAndSend(m); err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, err.Error(), "email//Send")
					}
					return *env.NewInteger(1)
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "email//Send")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.EmailType}, "email//Send")
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
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Opens a connection to an FTP server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Uri:
				conn, err := ftp.Dial(s.Path)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error connecting to FTP server: "+err.Error(), "ftp-uri//Open")
				}
				return *env.NewNative(ps.Idx, conn, "ftp-connection")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "ftp-uri//Open")
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
		Effect: env.EffIO,
		Argsn:  3,
		Doc:    "Logs in to an FTP server connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(s.Kind.Index) != "ftp-connection" {
					return evaldo.MakeBuiltinError(ps, "Expected FTP connection.", "ftp-connection//Login")
				}
				username, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Login")
				}
				pwd, ok := arg2.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "ftp-connection//Login")
				}
				conn, ok := s.Value.(*ftp.ServerConn)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Login")
				}
				if err := conn.Login(username.Value, pwd.Value); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error logging in to FTP server: "+err.Error(), "ftp-connection//Login")
				}
				return s
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Login")
			}
		},
	},

	// Args:
	// * connection: native ftp-connection object
	// * path: string containing the path of the file to retrieve
	// Returns:
	// * native reader object for the retrieved file
	"ftp-connection//Retrieve": {
		Effect: env.EffInput,
		Argsn:  2,
		Doc:    "Retrieves a file from an FTP server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(s.Kind.Index) != "ftp-connection" {
					return evaldo.MakeBuiltinError(ps, "Expected FTP connection.", "ftp-connection//Retrieve")
				}
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Retrieve")
				}
				path, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Retrieve")
				}
				resp, err := conn.Retr(path.Value)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error retrieving file from FTP server: "+err.Error(), "ftp-connection//Retrieve")
				}
				return *env.NewNative(ps.Idx, resp, "reader")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Retrieve")
			}
		},
	},

	// Examples:
	// equal { ftp: Open ftp://localhost:21	 ftp .Close |type? } 'integer
	// Args:
	// * connection: native ftp-connection object
	// Returns:
	// * integer 1 for success
	// * error if close fails
	"ftp-connection//Close": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Closes an FTP server connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(s.Kind.Index) != "ftp-connection" {
					return evaldo.MakeBuiltinError(ps, "Expected FTP connection.", "ftp-connection//Close")
				}
				conn, ok := s.Value.(*ftp.ServerConn)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Close")
				}
				err := conn.Quit()
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Error closing FTP connection: %v", err.Error()), "ftp-connection//Close")
				}
				return *env.NewInteger(1)
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Close")
			}
		},
	},

	// Examples:
	// ; Requires ftp server on localhost:21 with anonymous or test creds
	// ; Example smoke tests (skip in CI if no server):
	// ; conn: ftp-uri: Open ftp://localhost:21
	// ; conn: conn .Login "anonymous" "anonymous" ; or real creds
	// ; conn .Cwd "/" |kind? |equal 'native
	// ; conn .Mkdir "rye-test" |kind? |equal 'native
	// ; conn .List "/" |type? |equal 'block
	// ; conn .Store "/rye-test/hello.txt" "hi" |kind? |equal 'native
	// ; r: conn .Retrieve "/rye-test/hello.txt" |kind? |equal 'reader
	// ; r .Close |drop
	// ; lst: conn .List "/rye-test"
	// ; lst |length |>= 1 |assert
	// ; conn .Rename "/rye-test/hello.txt" "/rye-test/hello2.txt" |kind? |equal 'native
	// ; conn .Delete "/rye-test/hello2.txt" |kind? |equal 'native
	// ; conn .Rmdir "/rye-test" |kind? |equal 'native
	// ; conn .Close |equal 1
	// Additional FTP operations
	// Args:
	// * connection: native FTP connection.
	// * path: String directory path to list.
	//
	// Returns:
	// * Block of dicts describing each directory entry.
	"ftp-connection//List": {
		Effect: env.EffInput,
		Argsn:  2,
		Doc:    "Lists directory entries. Returns a block of maps with full info.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(s.Kind.Index) != "ftp-connection" {
					return evaldo.MakeBuiltinError(ps, "Expected FTP connection.", "ftp-connection//List")
				}
				var path string
				if arg1 != nil {
					switch p := arg1.(type) {
					case env.String:
						path = p.Value
					default:
						ps.FailureFlag = true
						return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//List")
					}
				}
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//List")
				}
				entries, err := conn.List(path)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "ftp-connection//List")
				}
				// Build block of maps
				vals := make([]env.Object, 0, len(entries))
				for _, e := range entries {
					typeStr := "file"
					switch e.Type {
					case ftp.EntryTypeFolder:
						typeStr = "dir"
					case ftp.EntryTypeLink:
						typeStr = "link"
					}
					m := env.NewDict(map[string]any{
						"name":   *env.NewString(e.Name),
						"type":   *env.NewString(typeStr),
						"size":   *env.NewInteger(int64(e.Size)),
						"time":   *env.NewTime(e.Time),
						"target": *env.NewString(e.Target),
					})
					vals = append(vals, *m)
				}
				return *env.NewBlock(*env.NewTSeries(vals))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//List")
			}
		},
	},
	// Examples:
	// ; conn .Cwd "/" |kind? |equal 'native
	// Args:
	// * connection: native FTP connection.
	// * path: String directory to change to.
	//
	// Returns:
	// * the FTP connection.
	"ftp-connection//Cwd": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Changes working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Cwd")
				}
				p, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Cwd")
				}
				if err := conn.ChangeDir(p.Value); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "ftp-connection//Cwd")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Cwd")
			}
		},
	},
	// Examples:
	// ; conn .Mkdir "rye-test" |kind? |equal 'native
	// Args:
	// * connection: native FTP connection.
	// * path: String directory to create.
	//
	// Returns:
	// * the FTP connection.
	"ftp-connection//Mkdir": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Creates a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Mkdir")
				}
				p, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Mkdir")
				}
				if err := conn.MakeDir(p.Value); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "ftp-connection//Mkdir")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Mkdir")
			}
		},
	},
	// Examples:
	// ; conn .Rmdir "rye-test" |kind? |equal 'native
	// Args:
	// * connection: native FTP connection.
	// * path: String directory to remove.
	//
	// Returns:
	// * the FTP connection.
	"ftp-connection//Rmdir": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Removes a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Rmdir")
				}
				p, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Rmdir")
				}
				if err := conn.RemoveDir(p.Value); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "ftp-connection//Rmdir")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Rmdir")
			}
		},
	},
	// Examples:
	// ; conn .Delete "/rye-test/hello.txt" |kind? |equal 'native
	// Args:
	// * connection: native FTP connection.
	// * path: String file to delete.
	//
	// Returns:
	// * the FTP connection.
	"ftp-connection//Delete": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Deletes a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Delete")
				}
				p, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Delete")
				}
				if err := conn.Delete(p.Value); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "ftp-connection//Delete")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Delete")
			}
		},
	},
	// Examples:
	// ; conn .Rename "/rye-test/a.txt" "/rye-test/b.txt" |kind? |equal 'native
	// Args:
	// * connection: native FTP connection.
	// * from: String current path.
	// * to: String new path.
	//
	// Returns:
	// * the FTP connection.
	"ftp-connection//Rename": {
		Effect: env.EffOutput,
		Argsn:  3,
		Doc:    "Renames or moves a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Rename")
				}
				from, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Rename")
				}
				to, ok := arg2.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "ftp-connection//Rename")
				}
				if err := conn.Rename(from.Value, to.Value); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "ftp-connection//Rename")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Rename")
			}
		},
	},
	// Examples:
	// ; conn .Store "/rye-test/up.txt" "content" |kind? |equal 'native
	// Args:
	// * connection: native FTP connection.
	// * path: String remote destination path.
	// * reader: Reader supplying the file contents.
	//
	// Returns:
	// * the FTP connection.
	"ftp-connection//Store": {
		Effect: env.EffOutput,
		Argsn:  3,
		Doc:    "Uploads data to a remote path. Accepts string or reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				conn, ok := asFTPConnection(s.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid FTP connection.", "ftp-connection//Store")
				}
				path, ok := arg1.(env.String)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ftp-connection//Store")
				}
				var rdr io.Reader
				switch v := arg2.(type) {
				case env.String:
					rdr = strings.NewReader(v.Value)
				case env.Native:
					// expect a reader
					if ps.Idx.GetWord(v.Kind.Index) != "reader" {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Expected reader for data", "ftp-connection//Store")
					}
					if r, ok := v.Value.(io.Reader); ok {
						rdr = r
					} else {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Invalid reader", "ftp-connection//Store")
					}
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType, env.NativeType}, "ftp-connection//Store")
				}
				if err := conn.Stor(path.Value, rdr); err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "ftp-connection//Store")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ftp-connection//Store")
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
		Effect: env.EffInput,
		Argsn:  3,
		Doc:    "Tails a file, following it for new content. Used for monitoring log files.",
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
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "tail-file")
			}

			// Get follow option
			follow := true // Default to true
			if arg1 != nil {
				switch f := arg1.(type) {
				case env.Boolean:
					follow = f.Value
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.BooleanType}, "tail-file")
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
					return evaldo.MakeArgError(ps, 3, []env.Type{env.BooleanType}, "tail-file")
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
				return evaldo.MakeBuiltinError(ps, err.Error(), "tail-file")
			}

			return *env.NewNative(ps.Idx, t, "tail-file")
		},
	},

	// Args:
	// * tail: native tail-file object
	// Returns:
	// * string containing the next line from the file, or nil if no more lines
	"tail-file//Read-line": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads the next line from a tailed file. Blocks until a line is available.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch t := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(t.GetKind()) != "tail-file" {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected tail-file object", "tail-file//Read-line")
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
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tail-file//Read-line")
			}
		},
	},

	// Args:
	// * tail: native tail-file object
	// Returns:
	// * empty string if successful
	"tail-file//Close": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Closes a tailed file, stopping the monitoring.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch t := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(t.GetKind()) != "tail-file" {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected tail-file object", "tail-file//Close")
				}

				err := t.Value.(*tail.Tail).Stop()
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "tail-file//Close")
				}

				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tail-file//Close")
			}
		},
	},

	//
	// ##### Unix Domain Socket Operations ##### "Unix domain socket operations for IPC"
	//
	// Args:
	// * path: uri representing the Unix socket path to connect to
	// Returns:
	// * native unix-connection object
	"unix-uri//Open": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Opens a connection to a Unix domain socket.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				conn, err := net.Dial("unix", s.Path)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error connecting to Unix socket: "+err.Error(), "unix-uri//Open")
				}
				return *env.NewNative(ps.Idx, conn, "unix-connection")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "unix-uri//Open")
			}
		},
	},

	// Args:
	// * connection: native unix-connection object
	// * data: string to write to the socket
	// Returns:
	// * the connection object if successful (allows chaining)
	"unix-connection//Write": {
		Effect: env.EffOutput,
		Argsn:  2,
		Doc:    "Writes data to a Unix domain socket connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch conn := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(conn.GetKind()) != "unix-connection" {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected unix-connection object", "unix-connection//Write")
				}
				switch data := arg1.(type) {
				case env.String:
					connection, ok := asNetConnection(conn.Value)
					if !ok {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Invalid unix-connection object", "unix-connection//Write")
					}
					_, err := connection.Write([]byte(data.Value))
					if err != nil {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "Error writing to Unix socket: "+err.Error(), "unix-connection//Write")
					}
					return arg0 // Return the connection for chaining
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "unix-connection//Write")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "unix-connection//Write")
			}
		},
	},

	// Args:
	// * connection: native unix-connection object
	// Returns:
	// * string containing data read from the socket
	"unix-connection//Read": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Reads data from a Unix domain socket connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch conn := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(conn.GetKind()) != "unix-connection" {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected unix-connection object", "unix-connection//Read")
				}
				connection, ok := asNetConnection(conn.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid unix-connection object", "unix-connection//Read")
				}
				buf := make([]byte, 1024)
				n, err := connection.Read(buf)
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Error reading from Unix socket: "+err.Error(), "unix-connection//Read")
				}
				return *env.NewString(string(buf[:n]))
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "unix-connection//Read")
			}
		},
	},

	// Args:
	// * connection: native unix-connection object
	// Returns:
	// * empty string if successful
	"unix-connection//Close": {
		Effect: env.EffIO,
		Argsn:  1,
		Doc:    "Closes a Unix domain socket connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch conn := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(conn.GetKind()) != "unix-connection" {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Expected unix-connection object", "unix-connection//Close")
				}
				connection, ok := asNetConnection(conn.Value)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid unix-connection object", "unix-connection//Close")
				}
				err := connection.Close()
				if err != nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, err.Error(), "unix-connection//Close")
				}
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "unix-connection//Close")
			}
		},
	},
}
