//go:build !no_os && !b_wasm

package evaldo

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/refaktor/go-find"
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

var Builtins_os = map[string]*env.Builtin{

	//
	// ##### OS ##### "OS related functions"
	//
	// Example:
	//  cc os
	//  print cwd?
	//  ls |for { .print }
	//  ls\ 'files |for { .print }
	//  host-info? -> "hostname" |print
	//  load-avg? -> "1" |print
	//  virtual-memory? -> "total" |print
	//  disk-usage? |print
	//  processes? |first |print
	//
	// Tests:
	// ; equal { cc os cd %/tmp cwd? cd %.. } %/tmp
	// Args:
	// * none
	// Returns:
	// * uri representing the current working directory
	"cwd?": {
		Argsn: 0,
		Doc:   "Gets the current working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			path, err := os.Getwd()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "cwd?")
			}
			return *env.NewUri1(ps.Idx, "file://"+path)
		},
	},

	// Tests:
	//  equal { cc os does-exist %main.rye } true
	// Args:
	// * path: uri representing the file or directory to check
	// Returns:
	// * boolean: true if exists, false if not
	"does-exist": {
		Argsn: 1,
		Doc:   "Checks if a file or directory exists.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := filepath.Join(ps.WorkingPath, path.GetPath())
				res := FileExists(filePath)
				if res == -1 {
					return MakeBuiltinError(ps, "Error checking if path exists", "does-exists")
				}
				return *env.NewBoolean(res == 1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "does-exist")
			}
		},
	},

	// Tests:
	// ; equal { cc os cd %/tmp cwd?  } %/tmp
	// Args:
	// * path: uri representing the directory to change to
	// Returns:
	// * the same uri if successful
	"cd": {
		Argsn: 1,
		Doc:   "Changes the current working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:

				err := os.Chdir(path.GetPath())
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "cd")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "cd")
			}
		},
	},

	// Args:
	// * variable_name: string containing the name of the environment variable
	// Returns:
	// * string containing the value of the environment variable
	"env?": {
		Argsn: 1,
		Doc:   "Gets the value of an environment variable.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch variable_name := arg0.(type) {
			case env.String:

				val, ok := os.LookupEnv(variable_name.Value)
				if !ok {
					return MakeBuiltinError(ps, "Variable couldn't be read", "env?")
				}
				return *env.NewString(val)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "env?")
			}
		},
	},

	// Tests:
	// ; equal { cc os mkdir %delme } %delme
	// Args:
	// * path: uri representing the directory to create
	// Returns:
	// * the same uri if successful
	"mkdir": {
		Argsn: 1,
		Doc:   "Creates a new directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				newDir := filepath.Join(ps.WorkingPath, path.GetPath())
				/*fmt.Println("-----------------------------")
				fmt.Println(filepath.Dir(ps.WorkingPath))
				fmt.Println(ps.WorkingPath)
				fmt.Println(path.GetPath())
				fmt.Println(newDir)*/
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

	// Args:
	// * none
	// Returns:
	// * uri representing the created temporary directory
	"mktmp": {
		Argsn: 0,
		Doc:   "Creates a new temporary directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir, err := os.MkdirTemp("", "rye-tmp-")
			if err != nil {
				return MakeBuiltinError(ps, "Error creating temporary directory: "+err.Error(), "mktmp")
			}
			return *env.NewUri1(ps.Idx, "file://"+dir)
		},
	},

	// Args:
	// * path: uri representing the file or empty directory to remove
	// Returns:
	// * the uri if successful
	// Tags: #file #delete
	"rm": {
		Argsn: 1,
		Doc:   "Removes a file or empty directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := filepath.Join(ps.WorkingPath, path.GetPath())
				err := os.Remove(filePath)
				if err != nil {
					return MakeBuiltinError(ps, "Error removing file: "+err.Error(), "rm")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "rm")
			}
		},
	},

	// Args:
	// * path: uri representing the directory to remove (including contents)
	// Returns:
	// * the uri if successful
	// Tags: #file #delete
	"rmdir": {
		Argsn: 1,
		Doc:   "Removes a directory and all its contents recursively.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				dirPath := filepath.Join(ps.WorkingPath, path.GetPath())
				err := os.RemoveAll(dirPath)
				if err != nil {
					return MakeBuiltinError(ps, "Error removing directory: "+err.Error(), "rmdir")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "rmdir")
			}
		},
	},

	// Args:
	// * source: uri representing the source file
	// * destination: uri representing the destination file
	// Returns:
	// * destination uri if successful
	// Tags: #file #copy
	"cp": {
		Argsn: 2,
		Doc:   "Copies a file to a new location.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch dst := arg1.(type) {
				case env.Uri:
					srcPath := filepath.Join(ps.WorkingPath, src.GetPath())
					dstPath := filepath.Join(ps.WorkingPath, dst.GetPath())

					// Read source file
					data, err := os.ReadFile(srcPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error reading source file: "+err.Error(), "cp")
					}

					// Get source file permissions
					srcInfo, err := os.Stat(srcPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error getting source file info: "+err.Error(), "cp")
					}

					// Write to destination with same permissions
					err = os.WriteFile(dstPath, data, srcInfo.Mode())
					if err != nil {
						return MakeBuiltinError(ps, "Error writing destination file: "+err.Error(), "cp")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "cp")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "cp")
			}
		},
	},

	// Args:
	// * path: uri representing the file or directory
	// Returns:
	// * dict with keys: name, size, mode, mod-time, is-dir
	// Tags: #file #info
	"file-info?": {
		Argsn: 1,
		Doc:   "Gets detailed information about a file or directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := filepath.Join(ps.WorkingPath, path.GetPath())
				info, err := os.Stat(filePath)
				if err != nil {
					return MakeBuiltinError(ps, "Error getting file info: "+err.Error(), "file-info?")
				}
				r := env.NewDict(make(map[string]any, 5))
				r.Data["name"] = *env.NewString(info.Name())
				r.Data["size"] = *env.NewInteger(info.Size())
				r.Data["mode"] = *env.NewString(info.Mode().String())
				r.Data["mod-time"] = *env.NewDate(info.ModTime())
				r.Data["is-dir"] = *env.NewBoolean(info.IsDir())
				return *r
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-info?")
			}
		},
	},

	// Args:
	// * path: uri representing the path to check
	// Returns:
	// * boolean: true if path is a directory, false otherwise
	// Tags: #file #check
	"is-dir?": {
		Argsn: 1,
		Doc:   "Checks if a path is a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := filepath.Join(ps.WorkingPath, path.GetPath())
				info, err := os.Stat(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						return *env.NewBoolean(false)
					}
					return MakeBuiltinError(ps, "Error checking path: "+err.Error(), "is-dir?")
				}
				return *env.NewBoolean(info.IsDir())
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "is-dir?")
			}
		},
	},

	// Args:
	// * path: uri representing the path to check
	// Returns:
	// * boolean: true if path is a regular file, false otherwise
	// Tags: #file #check
	"is-file?": {
		Argsn: 1,
		Doc:   "Checks if a path is a regular file (not a directory).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := filepath.Join(ps.WorkingPath, path.GetPath())
				info, err := os.Stat(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						return *env.NewBoolean(false)
					}
					return MakeBuiltinError(ps, "Error checking path: "+err.Error(), "is-file?")
				}
				return *env.NewBoolean(info.Mode().IsRegular())
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "is-file?")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * uri representing the user's home directory
	// Tags: #file #directory
	"home-dir?": {
		Argsn: 0,
		Doc:   "Gets the current user's home directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir, err := os.UserHomeDir()
			if err != nil {
				return MakeBuiltinError(ps, "Error getting home directory: "+err.Error(), "home-dir?")
			}
			return *env.NewUri1(ps.Idx, "file://"+dir)
		},
	},

	// Args:
	// * none
	// Returns:
	// * uri representing the system's temporary directory
	// Tags: #file #directory
	"tmp-dir?": {
		Argsn: 0,
		Doc:   "Gets the system's default temporary directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir := os.TempDir()
			return *env.NewUri1(ps.Idx, "file://"+dir)
		},
	},

	// Args:
	// * pattern: string containing a glob pattern (e.g., "*.txt", "data/*.csv")
	// Returns:
	// * block of uris matching the pattern
	// Tags: #file #search
	"glob": {
		Argsn: 1,
		Doc:   "Returns files matching a glob pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pattern := arg0.(type) {
			case env.String:
				patternPath := filepath.Join(ps.WorkingPath, pattern.Value)
				matches, err := filepath.Glob(patternPath)
				if err != nil {
					return MakeBuiltinError(ps, "Error in glob pattern: "+err.Error(), "glob")
				}
				items := make([]env.Object, len(matches))
				for i, match := range matches {
					items[i] = *env.NewUri1(ps.Idx, "file://"+match)
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "glob")
			}
		},
	},

	// Args:
	// * source: uri representing the source file or directory
	// * destination: uri representing the destination file or directory
	// Returns:
	// * destination uri if successful
	"mv": {
		Argsn: 2,
		Doc:   "Moves or renames a file or directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch dst := arg1.(type) {
				case env.Uri:
					oldPath := filepath.Join(ps.WorkingPath, src.GetPath())
					newPath := filepath.Join(ps.WorkingPath, dst.GetPath())
					err := os.Rename(oldPath, newPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error moving file: "+err.Error(), "mv")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "mv")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mv")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * block of uris representing files and directories in the current directory
	"ls": {
		Argsn: 0,
		Doc:   "Lists files and directories in the current directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			files, err := os.ReadDir(".")
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

	// Args:
	// * filter: file-uri to list directory, word 'dirs' or 'files' to filter by type, string for partial name matching, or regexp to match names
	// Returns:
	// * block of uris representing filtered files or directories in the specified directory or current directory
	"ls\\": {
		Argsn: 1,
		Doc:   "Lists files or directories with absolute paths when possible. If argument is a file-uri, lists that directory (or returns failure if not a directory). Otherwise filters current directory: 'dirs' for directories only, 'files' for files only, a string for partial name matching, or a regexp to match names.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var targetDir string = "."
			var doFiltering bool = true
			var filterArg env.Object = arg0

			// Check if argument is a file-uri
			if uri, ok := arg0.(env.Uri); ok {
				targetPath := resolvePath(ps.WorkingPath, uri.GetPath())
				
				// Check if path exists and is a directory
				info, err := os.Stat(targetPath)
				if err != nil {
					if os.IsNotExist(err) {
						return MakeBuiltinError(ps, fmt.Sprintf("path doesn't exist: %s", uri.GetPath()), "ls\\")
					}
					if os.IsPermission(err) {
						return MakeBuiltinError(ps, fmt.Sprintf("permission denied: %s", uri.GetPath()), "ls\\")
					}
					return MakeBuiltinError(ps, fmt.Sprintf("error accessing path %s: %s", uri.GetPath(), err.Error()), "ls\\")
				}
				
				if !info.IsDir() {
					return MakeBuiltinError(ps, fmt.Sprintf("not a directory (is file): %s", uri.GetPath()), "ls\\")
				}
				
				targetDir = targetPath
				doFiltering = false
			}

			files, err := os.ReadDir(targetDir)
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory:"+err.Error(), "ls\\")
			}

			var items []env.Object

			if !doFiltering {
				// If we're listing a specific directory, return all files with absolute paths
				for _, file := range files {
					absPath, err := filepath.Abs(filepath.Join(targetDir, file.Name()))
					if err != nil {
						// Fallback to relative path if absolute path fails
						items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
					} else {
						items = append(items, *env.NewUri1(ps.Idx, "file://"+absPath))
					}
				}
			} else {
				// Apply filtering based on the argument type
				switch filterArg := filterArg.(type) {
				case env.Word:
					filter := ps.Idx.GetWord(filterArg.Index)
					if filter != "dirs" && filter != "files" {
						return MakeBuiltinError(ps, "Word filter must be 'dirs' or 'files'", "ls\\")
					}

					for _, file := range files {
						include := false
						if filter == "dirs" {
							include = file.IsDir()
						} else if filter == "files" {
							include = !file.IsDir()
						}

						if include {
							absPath, err := filepath.Abs(filepath.Join(targetDir, file.Name()))
							if err != nil {
								// Fallback to relative path if absolute path fails
								items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
							} else {
								items = append(items, *env.NewUri1(ps.Idx, "file://"+absPath))
							}
						}
					}

				case env.String:
					// String does partial matching on file/directory names
					pattern := filterArg.Value
					for _, file := range files {
						if strings.Contains(file.Name(), pattern) {
							absPath, err := filepath.Abs(filepath.Join(targetDir, file.Name()))
							if err != nil {
								// Fallback to relative path if absolute path fails
								items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
							} else {
								items = append(items, *env.NewUri1(ps.Idx, "file://"+absPath))
							}
						}
					}

				case env.Native:
					// Check if it's a regexp
					if ps.Idx.GetWord(filterArg.Kind.Index) != "regexp" {
						return MakeBuiltinError(ps, "Native object must be a regexp", "ls\\")
					}

					regex := filterArg.Value.(*regexp.Regexp)
					for _, file := range files {
						if regex.MatchString(file.Name()) {
							absPath, err := filepath.Abs(filepath.Join(targetDir, file.Name()))
							if err != nil {
								// Fallback to relative path if absolute path fails
								items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
							} else {
								items = append(items, *env.NewUri1(ps.Idx, "file://"+absPath))
							}
						}
					}

				default:
					return MakeArgError(ps, 1, []env.Type{env.UriType, env.WordType, env.StringType, env.NativeType}, "ls\\")
				}
			}

			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	// Args:
	// * none
	// Returns:
	// * block of uris representing directories in the current directory
	"ls\\dirs": {
		Argsn: 0,
		Doc:   "Lists only directories in the current directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			files, err := os.ReadDir(".")
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory:"+err.Error(), "ls\\dirs")
			}

			var items []env.Object
			for _, file := range files {
				if file.IsDir() {
					items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
				}
			}
			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	// Args:
	// * none
	// Returns:
	// * block of uris representing files (non-directories) in the current directory
	"ls\\files": {
		Argsn: 0,
		Doc:   "Lists only files (non-directories) in the current directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			files, err := os.ReadDir(".")
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory:"+err.Error(), "ls\\files")
			}

			var items []env.Object
			for _, file := range files {
				if !file.IsDir() {
					items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
				}
			}
			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	// Args:
	// * filter: string for partial name matching, or regexp to match names
	// Returns:
	// * block of uris representing filtered directories in the current directory
	"ls\\dirs\\": {
		Argsn: 1,
		Doc:   "Lists directories in the current directory with filtering. Use a string for partial name matching, or a regexp to match names.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			files, err := os.ReadDir(".")
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory:"+err.Error(), "ls\\dirs\\")
			}

			var items []env.Object

			switch filterArg := arg0.(type) {
			case env.String:
				// String does partial matching on directory names
				pattern := filterArg.Value
				for _, file := range files {
					if file.IsDir() && strings.Contains(file.Name(), pattern) {
						items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
					}
				}

			case env.Native:
				// Check if it's a regexp
				if ps.Idx.GetWord(filterArg.Kind.Index) != "regexp" {
					return MakeBuiltinError(ps, "Native object must be a regexp", "ls\\dirs\\")
				}

				regex := filterArg.Value.(*regexp.Regexp)
				for _, file := range files {
					if file.IsDir() && regex.MatchString(file.Name()) {
						items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
					}
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.NativeType}, "ls\\dirs\\")
			}

			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	// Args:
	// * filter: string for partial name matching, or regexp to match names
	// Returns:
	// * block of uris representing filtered files (non-directories) in the current directory
	"ls\\files\\": {
		Argsn: 1,
		Doc:   "Lists files (non-directories) in the current directory with filtering. Use a string for partial name matching, or a regexp to match names.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			files, err := os.ReadDir(".")
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory:"+err.Error(), "ls\\files\\")
			}

			var items []env.Object

			switch filterArg := arg0.(type) {
			case env.String:
				// String does partial matching on file names
				pattern := filterArg.Value
				for _, file := range files {
					if !file.IsDir() && strings.Contains(file.Name(), pattern) {
						items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
					}
				}

			case env.Native:
				// Check if it's a regexp
				if ps.Idx.GetWord(filterArg.Kind.Index) != "regexp" {
					return MakeBuiltinError(ps, "Native object must be a regexp", "ls\\files\\")
				}

				regex := filterArg.Value.(*regexp.Regexp)
				for _, file := range files {
					if !file.IsDir() && regex.MatchString(file.Name()) {
						items = append(items, *env.NewUri1(ps.Idx, "file://"+file.Name()))
					}
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.NativeType}, "ls\\files\\")
			}

			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	// Args:
	// * path: uri representing source path
	// * link: uri representing link path to create
	// Returns:
	// * link uri if successful
	// Tags: #file #link
	"symlink": {
		Argsn: 2,
		Doc:   "Creates a symbolic link.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch link := arg1.(type) {
				case env.Uri:
					srcPath := resolvePath(ps.WorkingPath, src.GetPath())
					linkPath := resolvePath(ps.WorkingPath, link.GetPath())
					err := os.Symlink(srcPath, linkPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error creating symlink: "+err.Error(), "symlink")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "symlink")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "symlink")
			}
		},
	},

	// Args:
	// * path: uri representing a path to check
	// Returns:
	// * boolean: true if path is a symbolic link
	// Tags: #file #check
	"is-symlink?": {
		Argsn: 1,
		Doc:   "Checks if a path is a symbolic link.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := resolvePath(ps.WorkingPath, path.GetPath())
				info, err := os.Lstat(filePath) // Use Lstat to not follow symlinks
				if err != nil {
					if os.IsNotExist(err) {
						return *env.NewBoolean(false)
					}
					return MakeBuiltinError(ps, "Error checking path: "+err.Error(), "is-symlink?")
				}
				return *env.NewBoolean(info.Mode()&os.ModeSymlink != 0)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "is-symlink?")
			}
		},
	},

	// Args:
	// * path: uri representing a symbolic link
	// Returns:
	// * uri representing the target of the symlink
	// Tags: #file #link
	"readlink": {
		Argsn: 1,
		Doc:   "Reads the target of a symbolic link.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				linkPath := resolvePath(ps.WorkingPath, path.GetPath())
				target, err := os.Readlink(linkPath)
				if err != nil {
					return MakeBuiltinError(ps, "Error reading symlink: "+err.Error(), "readlink")
				}
				return *env.NewUri1(ps.Idx, "file://"+target)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "readlink")
			}
		},
	},

	// Args:
	// * path: uri representing file or directory path
	// * mode: integer representing file mode (e.g., 0755)
	// Returns:
	// * path uri if successful
	// Tags: #file #permissions
	"chmod": {
		Argsn: 2,
		Doc:   "Changes file or directory permissions.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				switch mode := arg1.(type) {
				case env.Integer:
					filePath := resolvePath(ps.WorkingPath, path.GetPath())
					err := os.Chmod(filePath, os.FileMode(mode.Value))
					if err != nil {
						return MakeBuiltinError(ps, "Error changing permissions: "+err.Error(), "chmod")
					}
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "chmod")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "chmod")
			}
		},
	},

	// Args:
	// * path: uri representing file or directory
	// Returns:
	// * string representing the absolute path
	// Tags: #file #path
	"abs-path": {
		Argsn: 1,
		Doc:   "Returns the absolute path for a given path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				absPath, err := filepath.Abs(resolvePath(ps.WorkingPath, path.GetPath()))
				if err != nil {
					return MakeBuiltinError(ps, "Error getting absolute path: "+err.Error(), "abs-path")
				}
				return *env.NewUri1(ps.Idx, "file://"+absPath)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "abs-path")
			}
		},
	},

	// Args:
	// * path: uri representing a path
	// Returns:
	// * uri representing the directory containing the path
	// Tags: #file #path
	"dirname": {
		Argsn: 1,
		Doc:   "Returns the directory part of a path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				dir := filepath.Dir(path.GetPath())
				return *env.NewUri1(ps.Idx, "file://"+dir)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "dirname")
			}
		},
	},

	// Args:
	// * path: uri representing a path
	// Returns:
	// * string representing the filename part of the path
	// Tags: #file #path
	"basename": {
		Argsn: 1,
		Doc:   "Returns the filename part of a path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				base := filepath.Base(path.GetPath())
				return *env.NewString(base)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "basename")
			}
		},
	},

	// Args:
	// * path: uri representing a file path
	// Returns:
	// * string representing the file extension (including the dot)
	// Tags: #file #path
	"file-ext": {
		Argsn: 1,
		Doc:   "Returns the file extension of a path (including the dot).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				ext := filepath.Ext(path.GetPath())
				return *env.NewString(ext)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-ext")
			}
		},
	},

	// Args:
	// * paths: uri or block of uris to join
	// Returns:
	// * uri representing the joined path
	// Tags: #file #path
	"join-path": {
		Argsn: 1,
		Doc:   "Joins path elements into a single path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch paths := arg0.(type) {
			case env.Block:
				pathStrs := make([]string, paths.Series.Len())
				for i := 0; i < paths.Series.Len(); i++ {
					item := paths.Series.Get(i)
					switch p := item.(type) {
					case env.Uri:
						pathStrs[i] = p.GetPath()
					case env.String:
						pathStrs[i] = p.Value
					default:
						return MakeBuiltinError(ps, "Block must contain only uris or strings", "join-path")
					}
				}
				joined := filepath.Join(pathStrs...)
				return *env.NewUri1(ps.Idx, "file://"+joined)
			case env.Uri:
				// If single uri passed, just return it
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.BlockType}, "join-path")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * string representing the current user's username
	// Tags: #system #user
	"whoami": {
		Argsn: 0,
		Doc:   "Returns the current user's username.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			username := os.Getenv("USER")
			if username == "" {
				username = os.Getenv("USERNAME") // Windows
			}
			if username == "" {
				return MakeBuiltinError(ps, "Unable to determine username", "whoami")
			}
			return *env.NewString(username)
		},
	},

	// Args:
	// * command: string representing the command to check
	// Returns:
	// * uri representing the path to the executable, or error if not found
	// Tags: #system #command
	"which": {
		Argsn: 1,
		Doc:   "Finds the path to an executable command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cmd := arg0.(type) {
			case env.String:
				path, err := findExecutable(cmd.Value)
				if err != nil {
					return MakeBuiltinError(ps, "Command not found: "+cmd.Value, "which")
				}
				return *env.NewUri1(ps.Idx, "file://"+path)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "which")
			}
		},
	},

	// Args:
	// * source: uri representing source directory
	// * destination: uri representing destination directory  
	// Returns:
	// * destination uri if successful
	// Tags: #file #copy #recursive
	"cp-r": {
		Argsn: 2,
		Doc:   "Recursively copies a directory and all its contents.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch dst := arg1.(type) {
				case env.Uri:
					srcPath := resolvePath(ps.WorkingPath, src.GetPath())
					dstPath := resolvePath(ps.WorkingPath, dst.GetPath())
					err := copyRecursive(srcPath, dstPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error copying recursively: "+err.Error(), "cp-r")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "cp-r")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "cp-r")
			}
		},
	},

	// Args:
	// * path: uri representing directory path to create
	// Returns:
	// * path uri if successful  
	// Tags: #file #directory
	"mkdir-p": {
		Argsn: 1,
		Doc:   "Creates a directory and any necessary parent directories.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				dirPath := resolvePath(ps.WorkingPath, path.GetPath())
				err := os.MkdirAll(dirPath, 0755)
				if err != nil {
					return MakeBuiltinError(ps, "Error creating directory: "+err.Error(), "mkdir-p")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mkdir-p")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * dictionary containing information about the host system (hostname, uptime, OS, etc.)
	"host-info?": {
		Argsn: 0,
		Doc:   "Gets information about the host system.",
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
	// Args:
	// * none
	// Returns:
	// * table containing information about users (user, terminal, host, started)
	"users?": {
		Argsn: 0,
		Doc:   "Gets information about users currently logged into the system.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			users, err := host.Users()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "users?")
			}
			s := env.NewTable([]string{"User", "Terminal", "Host", "Started"})
			for _, user := range users {
				vals := []any{
					*env.NewString(user.User),
					*env.NewString(user.Terminal),
					*env.NewString(user.Host),
					*env.NewInteger(int64(user.Started)),
				}
				s.AddRow(*env.NewTableRow(vals, s))
			}
			return *s
		},
	},
	// Args:
	// * none
	// Returns:
	// * dictionary with keys "1", "5", and "15" representing load averages
	"load-avg?": {
		Argsn: 0,
		Doc:   "Gets the system load average over the last 1, 5, and 15 minutes.",
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
	// Args:
	// * none
	// Returns:
	// * dictionary containing information about virtual memory (total, free, used-percent)
	"virtual-memory?": {
		Argsn: 0,
		Doc:   "Gets information about virtual memory usage.",
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
	// Args:
	// * none
	// Returns:
	// * table containing disk usage information for all partitions
	"disk-usage?": {
		Argsn: 0,
		Doc:   "Gets disk usage information for all partitions.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			partitions, err := disk.Partitions(true)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "disk-usage?")
			}
			s := env.NewTable([]string{"Filesystem", "Size", "Used", "Available", "Capacity", "iused", "ifree", "%iused", "Mounted on"})
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
					*env.NewDecimal(usage.UsedPercent),
					*env.NewInteger(int64(usage.InodesUsed)),
					*env.NewInteger(int64(usage.InodesFree)),
					*env.NewInteger(int64(usage.InodesUsedPercent)),
					*env.NewString(usage.Path),
				}
				s.AddRow(*env.NewTableRow(vals, s))
			}
			return *s
		},
	},
	// Args:
	// * none
	// Returns:
	// * block of integers representing process IDs
	"pids?": {
		Argsn: 0,
		Doc:   "Gets a list of all process IDs currently running.",
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
	// Args:
	// * none
	// Returns:
	// * table containing detailed information about all running processes
	"processes?": {
		Argsn: 0,
		Doc:   "Gets detailed information about all running processes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			processes, err := process.Processes()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "processes?")
			}
			s := proccesTableBase()
			for _, process := range processes {
				processTableAdd(s, process)
			}
			return *s
		},
	},
	// Args:
	// * pid: integer process ID
	// Returns:
	// * dictionary containing detailed information about the specified process
	"process": {
		Argsn: 1,
		Doc:   "Gets detailed information about a specific process by PID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pid := arg0.(type) {
			case env.Integer:
				process, err := process.NewProcess(int32(pid.Value))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "process")
				}
				s := proccesTableBase()
				processTableAdd(s, process)
				return s.Rows[0].ToDict()
			default:
				return *MakeArgError(ps, 1, []env.Type{env.IntegerType}, "process")
			}
		},
	},

	// Args:
	// * ip: string containing an IP address
	// Returns:
	// * block of strings containing hostnames associated with the IP
	"lookup-address": {
		Argsn: 1,
		Doc:   "Performs a reverse DNS lookup to get hostnames for an IP address.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ip := arg0.(type) {
			case env.String:
				names, err := net.LookupAddr(ip.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "lookup-address")
				}
				items := make([]env.Object, len(names))
				for i, name := range names {
					items[i] = *env.NewString(name)
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "lookup-address")
			}
		},
	},

	// Args:
	// * hostname: string containing a hostname
	// Returns:
	// * block of strings containing IP addresses associated with the hostname
	"lookup-ip": {
		Argsn: 1,
		Doc:   "Performs a DNS lookup to get IP addresses for a hostname.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch hostname := arg0.(type) {
			case env.String:
				ips, err := net.LookupIP(hostname.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "lookup-ip")
				}
				items := make([]env.Object, len(ips))
				for i, ip := range ips {
					items[i] = *env.NewString(ip.String())
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "lookup-ip")
			}
		},
	},

	// Args:
	// * value: string to write to the clipboard
	// Returns:
	// * the same string if successful
	"write\\clipboard": {
		Argsn: 1,
		Doc:   "Writes a string value to the system clipboard.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.String:
				err := clipboard.WriteAll(val.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "write\\clipboard")
				}
				return arg0
			default:
				return *MakeArgError(ps, 1, []env.Type{env.StringType}, "write\\clipboard")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * string containing the current contents of the clipboard
	"read\\clipboard": {
		Argsn: 0,
		Doc:   "Reads the current contents of the system clipboard.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			val, err := clipboard.ReadAll()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "read\\clipboard")
			}
			return *env.NewString(val)
		},
	},

	//
	// ##### Archive Functions ##### "Functions for creating and extracting archives"
	//

	// Creates a .tar.gz archive from a directory or file.
	// Args:
	// * source: uri representing the file or directory to archive
	// * destination: uri representing the output .tar.gz file
	// Returns:
	// * destination uri if successful
	// Tags: #archive #compress
	"tgz": {
		Argsn: 2,
		Doc:   "Creates a .tar.gz archive from a file or directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch dst := arg1.(type) {
				case env.Uri:
					srcPath := resolvePath(ps.WorkingPath, src.GetPath())
					dstPath := resolvePath(ps.WorkingPath, dst.GetPath())
					err := createTarGz(srcPath, dstPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error creating tar.gz: "+err.Error(), "tgz")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "tgz")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "tgz")
			}
		},
	},

	// Extracts a .tar.gz archive to a directory.
	// Args:
	// * source: uri representing the .tar.gz file
	// * destination: uri representing the output directory
	// Returns:
	// * destination uri if successful
	// Tags: #archive #extract
	"un-tgz": {
		Argsn: 2,
		Doc:   "Extracts a .tar.gz archive to a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch dst := arg1.(type) {
				case env.Uri:
					srcPath := resolvePath(ps.WorkingPath, src.GetPath())
					dstPath := resolvePath(ps.WorkingPath, dst.GetPath())
					err := extractTarGz(srcPath, dstPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error extracting tar.gz: "+err.Error(), "un-tgz")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "un-tgz")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "un-tgz")
			}
		},
	},

	// Creates a .zip archive from a directory or file.
	// Args:
	// * source: uri representing the file or directory to archive
	// * destination: uri representing the output .zip file
	// Returns:
	// * destination uri if successful
	// Tags: #archive #compress
	"zip": {
		Argsn: 2,
		Doc:   "Creates a .zip archive from a file or directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch dst := arg1.(type) {
				case env.Uri:
					srcPath := resolvePath(ps.WorkingPath, src.GetPath())
					dstPath := resolvePath(ps.WorkingPath, dst.GetPath())
					err := createZip(srcPath, dstPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error creating zip: "+err.Error(), "zip")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "zip")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "zip")
			}
		},
	},

	// Extracts a .zip archive to a directory.
	// Args:
	// * source: uri representing the .zip file
	// * destination: uri representing the output directory
	// Returns:
	// * destination uri if successful
	// Tags: #archive #extract
	"unzip": {
		Argsn: 2,
		Doc:   "Extracts a .zip archive to a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch dst := arg1.(type) {
				case env.Uri:
					srcPath := resolvePath(ps.WorkingPath, src.GetPath())
					dstPath := resolvePath(ps.WorkingPath, dst.GetPath())
					err := extractZip(srcPath, dstPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error extracting zip: "+err.Error(), "unzip")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "unzip")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "unzip")
			}
		},
	},

	//
	// ##### Find ##### "File finding functions using go-find library"
	//
	// Example:
	//  find %src |name "*.go" |type 'file |eval
	//  find %. |max-depth! 2 |name "*.txt" |eval
	//  find %/home |min-depth! 1 |max-depth! 3 |regex regexp "test.*\.go$" |eval
	//

	// Creates a new finder starting from the given path(s).
	// Args:
	// * path: uri or block of uris representing starting paths
	// Returns:
	// * native finder object
	// Tags: #find #files
	"find": {
		Argsn: 1,
		Doc:   "Creates a new file finder starting from the given path(s).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				finder := find.NewFind(path.GetPath())
				return *env.NewNative(ps.Idx, finder, "finder")
			case env.Block:
				paths := make([]string, 0, path.Series.Len())
				for i := 0; i < path.Series.Len(); i++ {
					item := path.Series.Get(i)
					switch p := item.(type) {
					case env.Uri:
						paths = append(paths, p.GetPath())
					case env.String:
						paths = append(paths, p.Value)
					default:
						return MakeBuiltinError(ps, "Block must contain only uris or strings", "find")
					}
				}
				finder := find.NewFind(paths...)
				return *env.NewNative(ps.Idx, finder, "finder")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.BlockType}, "find")
			}
		},
	},

	// Sets the minimum depth for the finder.
	// Args:
	// * finder: native finder object
	// * depth: integer minimum depth
	// Returns:
	// * native finder object (for chaining)
	// Tags: #find #filter
	"finder//min-depth!": {
		Argsn: 2,
		Doc:   "Sets the minimum depth for file traversal.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					switch depth := arg1.(type) {
					case env.Integer:
						finder.MinDepth(int(depth.Value))
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "finder//min-depth!")
					}
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//min-depth!")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//min-depth!")
			}
		},
	},

	// Sets the maximum depth for the finder.
	// Args:
	// * finder: native finder object
	// * depth: integer maximum depth
	// Returns:
	// * native finder object (for chaining)
	// Tags: #find #filter
	"finder//max-depth!": {
		Argsn: 2,
		Doc:   "Sets the maximum depth for file traversal.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					switch depth := arg1.(type) {
					case env.Integer:
						finder.MaxDepth(int(depth.Value))
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "finder//max-depth!")
					}
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//max-depth!")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//max-depth!")
			}
		},
	},

	// Filters by file type.
	// Args:
	// * finder: native finder object
	// * type: word 'file or 'dir (or string "f" or "d")
	// Returns:
	// * native finder object (for chaining)
	// Tags: #find #filter
	"finder//type": {
		Argsn: 2,
		Doc:   "Filters results by type: 'file (or \"f\") for files, 'dir (or \"d\") for directories.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					var typeStr string
					switch t := arg1.(type) {
					case env.Word:
						word := ps.Idx.GetWord(t.Index)
						switch word {
						case "file":
							typeStr = "f"
						case "dir":
							typeStr = "d"
						default:
							return MakeBuiltinError(ps, "Type must be 'file or 'dir", "finder//type")
						}
					case env.String:
						if t.Value != "f" && t.Value != "d" {
							return MakeBuiltinError(ps, "Type must be \"f\" or \"d\"", "finder//type")
						}
						typeStr = t.Value
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "finder//type")
					}
					finder.Type(typeStr)
					return arg0
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//type")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//type")
			}
		},
	},

	// Filters by file name using glob pattern.
	// Args:
	// * finder: native finder object
	// * pattern: string glob pattern (e.g., "*.go", "test_*")
	// Returns:
	// * native finder object (for chaining)
	// Tags: #find #filter
	"finder//name": {
		Argsn: 2,
		Doc:   "Filters results by file name using a glob pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					switch pattern := arg1.(type) {
					case env.String:
						finder.Name(pattern.Value)
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "finder//name")
					}
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//name")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//name")
			}
		},
	},

	// Filters by full path using glob pattern.
	// Args:
	// * finder: native finder object
	// * pattern: string glob pattern for the whole path
	// Returns:
	// * native finder object (for chaining)
	// Tags: #find #filter
	"finder//whole-name": {
		Argsn: 2,
		Doc:   "Filters results by full path using a glob pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					switch pattern := arg1.(type) {
					case env.String:
						finder.WholeName(pattern.Value)
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "finder//whole-name")
					}
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//whole-name")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//whole-name")
			}
		},
	},

	// Filters by regular expression on the full path.
	// Args:
	// * finder: native finder object
	// * regex: native regexp object
	// Returns:
	// * native finder object (for chaining)
	// Tags: #find #filter
	"finder//regex": {
		Argsn: 2,
		Doc:   "Filters results by regular expression on the full path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					switch r := arg1.(type) {
					case env.Native:
						if regex, ok := r.Value.(*regexp.Regexp); ok {
							finder.Regex(regex)
							return arg0
						}
						return MakeBuiltinError(ps, "Expected regexp object", "finder//regex")
					default:
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "finder//regex")
					}
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//regex")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//regex")
			}
		},
	},

	// Filters for empty files or directories.
	// Args:
	// * finder: native finder object
	// Returns:
	// * native finder object (for chaining)
	// Tags: #find #filter
	"finder//empty": {
		Argsn: 1,
		Doc:   "Filters for empty files or directories.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					finder.Empty()
					return arg0
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//empty")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//empty")
			}
		},
	},

	// Executes the find operation and returns results.
	// Args:
	// * finder: native finder object
	// Returns:
	// * block of uris representing found files/directories
	// Tags: #find #execute
	"finder//eval": {
		Argsn: 1,
		Doc:   "Executes the find operation and returns matching paths as uris.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Native:
				if finder, ok := f.Value.(*find.Find); ok {
					results, err := finder.Evaluate()
					if err != nil {
						return MakeBuiltinError(ps, "Error evaluating find: "+err.Error(), "finder//eval")
					}
					items := make([]env.Object, len(results))
					for i, path := range results {
						items[i] = *env.NewUri1(ps.Idx, "file://"+path)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				}
				return MakeBuiltinError(ps, "Expected finder object", "finder//eval")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "finder//eval")
			}
		},
	},

	//
	// ##### Additional OS utilities ##### "Common OS operations made simple"
	//

	// Gets the size of a file in bytes.
	// Args:
	// * path: uri representing the file
	// Returns:
	// * integer size in bytes
	// Tags: #file #size
	"file-size": {
		Argsn: 1,
		Doc:   "Gets the size of a file in bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := resolvePath(ps.WorkingPath, path.GetPath())
				info, err := os.Stat(filePath)
				if err != nil {
					return MakeBuiltinError(ps, "Error getting file size: "+err.Error(), "file-size")
				}
				return *env.NewInteger(info.Size())
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-size")
			}
		},
	},

	// Touches a file (creates if doesn't exist, updates access time if exists).
	// Args:
	// * path: uri representing the file to touch
	// Returns:
	// * path uri if successful
	// Tags: #file #create
	"touch": {
		Argsn: 1,
		Doc:   "Creates a file if it doesn't exist, or updates its access time if it does.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				filePath := resolvePath(ps.WorkingPath, path.GetPath())
				now := time.Now()
				err := os.Chtimes(filePath, now, now)
				if os.IsNotExist(err) {
					// File doesn't exist, create it
					file, err := os.Create(filePath)
					if err != nil {
						return MakeBuiltinError(ps, "Error creating file: "+err.Error(), "touch")
					}
					file.Close()
				} else if err != nil {
					return MakeBuiltinError(ps, "Error touching file: "+err.Error(), "touch")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "touch")
			}
		},
	},

	// Creates a hard link to a file.
	// Args:
	// * source: uri representing the source file
	// * link: uri representing the hard link to create
	// Returns:
	// * link uri if successful
	// Tags: #file #link
	"hardlink": {
		Argsn: 2,
		Doc:   "Creates a hard link to a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch src := arg0.(type) {
			case env.Uri:
				switch link := arg1.(type) {
				case env.Uri:
					srcPath := resolvePath(ps.WorkingPath, src.GetPath())
					linkPath := resolvePath(ps.WorkingPath, link.GetPath())
					err := os.Link(srcPath, linkPath)
					if err != nil {
						return MakeBuiltinError(ps, "Error creating hard link: "+err.Error(), "hardlink")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.UriType}, "hardlink")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "hardlink")
			}
		},
	},

	// Gets file or directory count in a directory.
	// Args:
	// * path: uri representing the directory (optional, defaults to current dir)
	// Returns:
	// * integer count of items
	// Tags: #file #count
	"count-dir": {
		Argsn: 1,
		Doc:   "Counts the number of files and directories in a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var dirPath string
			switch path := arg0.(type) {
			case env.Uri:
				dirPath = resolvePath(ps.WorkingPath, path.GetPath())
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "count-dir")
			}

			files, err := os.ReadDir(dirPath)
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory: "+err.Error(), "count-dir")
			}
			return *env.NewInteger(int64(len(files)))
		},
	},

	// Gets the current process ID.
	// Args:
	// * none
	// Returns:
	// * integer process ID
	// Tags: #system #process
	"pid": {
		Argsn: 0,
		Doc:   "Gets the current process ID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(int64(os.Getpid()))
		},
	},

	// Gets the parent process ID.
	// Args:
	// * none
	// Returns:
	// * integer parent process ID
	// Tags: #system #process
	"ppid": {
		Argsn: 0,
		Doc:   "Gets the parent process ID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(int64(os.Getppid()))
		},
	},

	// Sleep for a specified duration.
	// Args:
	// * duration: integer seconds to sleep
	// Returns:
	// * integer seconds slept
	// Tags: #system #time
	"sleep": {
		Argsn: 1,
		Doc:   "Sleeps for the specified number of seconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch seconds := arg0.(type) {
			case env.Integer:
				time.Sleep(time.Duration(seconds.Value) * time.Second)
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},

	//
	// ##### TMUX ##### "tmux terminal multiplexer functions using gotmux library"
	//

	// Args:
	// * none
	// Returns:
	// * boolean indicating if tmux is available
	// Tags: #tmux #check
	"tmux-available?": {
		Argsn: 0,
		Doc:   "Checks if tmux is available on the system.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			_, err := exec.LookPath("tmux")
			return *env.NewBoolean(err == nil)
		},
	},

	// Args:
	// * session-name: string name for the new session
	// Returns:
	// * native tmux session object
	// Tags: #tmux #session
	"tmux-new-session": {
		Argsn: 1,
		Doc:   "Creates a new tmux session with the given name using gotmux library.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			tmux, err := gotmux.DefaultTmux()
			if err != nil {
				return MakeBuiltinError(ps, "Failed to connect to tmux: "+err.Error(), "tmux-new-session")
			}

			switch name := arg0.(type) {
			case env.String:
				session, err := tmux.NewSession(&gotmux.SessionOptions{
					Name: name.Value,
				})
				if err != nil {
					return MakeBuiltinError(ps, "Failed to create session: "+err.Error(), "tmux-new-session")
				}
				return *env.NewNative(ps.Idx, session, "tmux-session")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "tmux-new-session")
			}
		},
	},

	// Args:
	// * none
	// Returns:
	// * block of native tmux session objects
	// Tags: #tmux #session #list
	"tmux-list-sessions": {
		Argsn: 0,
		Doc:   "Lists all tmux sessions using gotmux library.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			tmux, err := gotmux.DefaultTmux()
			if err != nil {
				return MakeBuiltinError(ps, "Failed to connect to tmux: "+err.Error(), "tmux-list-sessions")
			}

			sessions, err := tmux.ListSessions()
			if err != nil {
				return MakeBuiltinError(ps, "Failed to list sessions: "+err.Error(), "tmux-list-sessions")
			}

			items := make([]env.Object, len(sessions))
			for i, session := range sessions {
				items[i] = *env.NewNative(ps.Idx, session, "tmux-session")
			}
			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	// Args:
	// * session-name: string name of the session to get
	// Returns:
	// * native tmux session object
	// Tags: #tmux #session
	"tmux-get-session": {
		Argsn: 1,
		Doc:   "Gets a tmux session by name using gotmux library.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			tmux, err := gotmux.DefaultTmux()
			if err != nil {
				return MakeBuiltinError(ps, "Failed to connect to tmux: "+err.Error(), "tmux-get-session")
			}

			switch name := arg0.(type) {
			case env.String:
				session, err := tmux.GetSessionByName(name.Value)
				if err != nil {
					return MakeBuiltinError(ps, "Failed to get session: "+err.Error(), "tmux-get-session")
				}
				return *env.NewNative(ps.Idx, session, "tmux-session")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "tmux-get-session")
			}
		},
	},

	// Args:
	// * session: native tmux session object
	// Returns:
	// * native tmux window object
	// Tags: #tmux #window
	"tmux-new-window": {
		Argsn: 1,
		Doc:   "Creates a new window in the given tmux session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					window, err := session.New()
					if err != nil {
						return MakeBuiltinError(ps, "Failed to create window: "+err.Error(), "tmux-new-window")
					}
					return *env.NewNative(ps.Idx, window, "tmux-window")
				}
				return MakeBuiltinError(ps, "Expected tmux-session object", "tmux-new-window")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-new-window")
			}
		},
	},

	// Args:
	// * session: native tmux session object
	// * window-name: string name for the window
	// Returns:
	// * native tmux window object
	// Tags: #tmux #window
	"tmux-new-window-named": {
		Argsn: 2,
		Doc:   "Creates a new window in the given tmux session with a specific name.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					switch name := arg1.(type) {
					case env.String:
						window, err := session.NewWindow(&gotmux.NewWindowOptions{
							WindowName: name.Value,
						})
						if err != nil {
							return MakeBuiltinError(ps, "Failed to create named window: "+err.Error(), "tmux-new-window-named")
						}
						return *env.NewNative(ps.Idx, window, "tmux-window")
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "tmux-new-window-named")
					}
				}
				return MakeBuiltinError(ps, "Expected tmux-session object", "tmux-new-window-named")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-new-window-named")
			}
		},
	},

	// Args:
	// * session: native tmux session object
	// Returns:
	// * block of native tmux window objects
	// Tags: #tmux #window #list
	"tmux-list-windows": {
		Argsn: 1,
		Doc:   "Lists all windows in the given tmux session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					windows, err := session.ListWindows()
					if err != nil {
						return MakeBuiltinError(ps, "Failed to list windows: "+err.Error(), "tmux-list-windows")
					}
					
					items := make([]env.Object, len(windows))
					for i, window := range windows {
						items[i] = *env.NewNative(ps.Idx, window, "tmux-window")
					}
					return *env.NewBlock(*env.NewTSeries(items))
				}
				return MakeBuiltinError(ps, "Expected tmux-session object", "tmux-list-windows")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-list-windows")
			}
		},
	},

	// Args:
	// * window: native tmux window object
	// * index: integer pane index (0-based)
	// Returns:
	// * native tmux pane object
	// Tags: #tmux #pane
	"tmux-get-pane": {
		Argsn: 2,
		Doc:   "Gets a pane from a tmux window by index.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				if window, ok := win.Value.(*gotmux.Window); ok {
					switch idx := arg1.(type) {
					case env.Integer:
						pane, err := window.GetPaneByIndex(int(idx.Value))
						if err != nil {
							return MakeBuiltinError(ps, "Failed to get pane: "+err.Error(), "tmux-get-pane")
						}
						return *env.NewNative(ps.Idx, pane, "tmux-pane")
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "tmux-get-pane")
					}
				}
				return MakeBuiltinError(ps, "Expected tmux-window object", "tmux-get-pane")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-get-pane")
			}
		},
	},

	// Args:
	// * pane: native tmux pane object
	// Returns:
	// * native tmux pane object
	// Tags: #tmux #pane
	"tmux-split-pane": {
		Argsn: 1,
		Doc:   "Splits a tmux pane horizontally.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				if pane, ok := p.Value.(*gotmux.Pane); ok {
					err := pane.Split()
					if err != nil {
						return MakeBuiltinError(ps, "Failed to split pane: "+err.Error(), "tmux-split-pane")
					}
					return arg0
				}
				return MakeBuiltinError(ps, "Expected tmux-pane object", "tmux-split-pane")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-split-pane")
			}
		},
	},

	// Args:
	// * pane: native tmux pane object
	// * command: string command to send
	// Returns:
	// * native tmux pane object
	// Tags: #tmux #pane #command
	"tmux-send-keys": {
		Argsn: 2,
		Doc:   "Sends keys/command to a tmux pane.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				if pane, ok := p.Value.(*gotmux.Pane); ok {
					switch cmd := arg1.(type) {
					case env.String:
						err := pane.SendKeys(cmd.Value)
						if err != nil {
							return MakeBuiltinError(ps, "Failed to send keys: "+err.Error(), "tmux-send-keys")
						}
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "tmux-send-keys")
					}
				}
				return MakeBuiltinError(ps, "Expected tmux-pane object", "tmux-send-keys")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-send-keys")
			}
		},
	},

	// Args:
	// * session: native tmux session object
	// Returns:
	// * native tmux session object
	// Tags: #tmux #session
	"tmux-kill-session": {
		Argsn: 1,
		Doc:   "Kills/destroys a tmux session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					err := session.Kill()
					if err != nil {
						return MakeBuiltinError(ps, "Failed to kill session: "+err.Error(), "tmux-kill-session")
					}
					return arg0
				}
				return MakeBuiltinError(ps, "Expected tmux-session object", "tmux-kill-session")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-kill-session")
			}
		},
	},
}

// resolvePath resolves a path - if it's absolute, returns it as-is; if relative, joins with workingPath
func resolvePath(workingPath, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workingPath, path)
}

// createTarGz creates a .tar.gz archive from the source path
func createTarGz(srcPath, dstPath string) error {
	outFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	var baseDir string
	if srcInfo.IsDir() {
		baseDir = filepath.Base(srcPath)
	}

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		if baseDir != "" {
			relPath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return err
			}
			header.Name = filepath.Join(baseDir, relPath)
		} else {
			header.Name = filepath.Base(path)
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})
}

// extractTarGz extracts a .tar.gz archive to the destination path
func extractTarGz(srcPath, dstPath string) error {
	inFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	gzReader, err := gzip.NewReader(inFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dstPath, header.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(dstPath)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}
	return nil
}

// createZip creates a .zip archive from the source path
func createZip(srcPath, dstPath string) error {
	outFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	var baseDir string
	if srcInfo.IsDir() {
		baseDir = filepath.Base(srcPath)
	}

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			relPath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return err
			}
			header.Name = filepath.Join(baseDir, relPath)
		} else {
			header.Name = filepath.Base(path)
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// extractZip extracts a .zip archive to the destination path
func extractZip(srcPath, dstPath string) error {
	reader, err := zip.OpenReader(srcPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath := filepath.Join(dstPath, file.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(dstPath)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}

		if err := os.Chmod(targetPath, file.Mode()); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to find an executable in PATH
func findExecutable(cmd string) (string, error) {
	// Check if command is an absolute path
	if filepath.IsAbs(cmd) {
		if _, err := os.Stat(cmd); err == nil {
			return cmd, nil
		}
		return "", fmt.Errorf("command not found")
	}

	// Search in PATH
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return "", fmt.Errorf("PATH environment variable not set")
	}

	pathDirs := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range pathDirs {
		if dir == "" {
			continue
		}
		fullPath := filepath.Join(dir, cmd)
		
		// Try with common executable extensions on Windows
		extensions := []string{""}
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
			extensions = []string{"", ".exe", ".bat", ".cmd", ".com"}
		}
		
		for _, ext := range extensions {
			testPath := fullPath + ext
			if info, err := os.Stat(testPath); err == nil && !info.IsDir() {
				// Check if file is executable (basic check)
				if info.Mode().Perm()&0111 != 0 || ext != "" {
					return testPath, nil
				}
			}
		}
	}
	return "", fmt.Errorf("command not found in PATH")
}

// Helper function to recursively copy directories
func copyRecursive(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		// Create destination directory
		if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
			return err
		}

		// Read source directory
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}

		// Copy each entry recursively
		for _, entry := range entries {
			srcPath := filepath.Join(src, entry.Name())
			dstPath := filepath.Join(dst, entry.Name())
			
			if err := copyRecursive(srcPath, dstPath); err != nil {
				return err
			}
		}
	} else {
		// Copy file
		srcFile, err := os.Open(src)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Create destination directory if needed
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		// Copy permissions
		if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func proccesTableBase() *env.Table {
	return env.NewTable([]string{
		"User",
		"PID",
		"TTY",
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

func processTableAdd(s *env.Table, process *process.Process) {
	var tty env.Object
	terminal, err := process.Terminal()
	if err == nil && terminal != "" {
		tty = *env.NewString(terminal)
	} else {
		tty = *env.NewString("?")
	}

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
		tty,
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
	s.AddRow(*env.NewTableRow(vals, s))
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
