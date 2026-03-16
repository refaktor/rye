package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// data that we want to optionally extract from builtins code and store as a general structure that
// * Rye runtime will be able to provide to the user
// * Test tool will run tests on
type builtinSection struct {
	name      string
	docstring string
	builtins  []builtinInfo
}

type builtinInfo struct {
	name      string         // from key value
	gentype   string         // optional from key value
	docstring string         // part of builtin definition
	doc       string         // free text at the top of the comment
	argsn     int            // Argsn field from builtin definition
	pure      bool           // Pure field from builtin definition
	args      []string       // extracted from comment or variable names
	returns   string
	argtypes  map[int][]string // extracted from MakeArgError calls: arg number → allowed types
	tests     []string         // extracted from comment
	examples  []string         // extracted from comment
	tags      []string         // extracted from comment
}

type counters struct {
	functions        int
	tested_functions int
	tests            int
	examples         int
}

// Helper function to get comments above the map key
func getCommentsAboveKey(fset *token.FileSet, comments []*ast.CommentGroup, keyPos token.Pos) string {
	for _, commentGroup := range comments {
		if fset.Position(commentGroup.End()).Line == fset.Position(keyPos).Line-1 {
			return commentGroup.Text()
		}
	}
	return ""
}

// Helper function to get comments above the map key
func parseCommentsAboveKey(input string, info *builtinInfo) (builtinInfo, *builtinSection) {

	const (
		inDoc = iota
		inTests
		inExamples
		inArgs
		inReturns
		x
	)
	position := inDoc

	// info := builtinInfo{}

	// Step 1: Split input into lines and trim whitespace
	lines := strings.Split(strings.TrimSpace(input), "\n")

	var section *builtinSection

	// Step 2: Separate header and tests
	var headerLines []string
	//	var testLines []string

	//	fmt.Println("!!!!!!!!!!!!!!!**************")

	re := regexp.MustCompile(`^##### ([A-Za-z0-9 ]+)#####\s+"([^"]*)"`)

	re_star := regexp.MustCompile("^\\* ")

	for _, line := range lines {
		line = strings.TrimSpace(line) // Remove leading and trailing whitespace
		// fmt.Println("LLLL:" + line)
		// fmt.Println(line)
		match := re.FindStringSubmatch(line)
		if match != nil {
			section = &builtinSection{match[1], match[2], make([]builtinInfo, 0)}
		}
		switch line {
		case "Tests:":
			position = inTests
			continue
		case "Example:":
			position = inExamples
			continue
		case "Args:":
			position = inArgs
			continue
		case "Returns:":
			position = inReturns
			continue
		}

		switch position {
		case inTests:
			info.tests = append(info.tests, line)
		case inDoc:
			headerLines = append(headerLines, line)
		case inExamples:
			info.examples = append(info.examples, line) // TODO --- examples can be multiline, there is a name also
		case inArgs:
			info.args = append(info.args, re_star.ReplaceAllString(line, ""))
		case inReturns:
			info.returns = info.returns + re_star.ReplaceAllString(line, "")
		}
	}
	// Step 3: Combine the header lines into a single string
	info.doc = strings.Join(headerLines, "\n")

	return *info, section
}

func outputInfo(sections *[]builtinSection) {
	for _, section := range *sections {
		fmt.Printf("section \"%s\" \"%s\" {\n", section.name, section.docstring) // name
		for _, info := range section.builtins {
			if len(info.tests) > 0 || len(info.args) > 0 || len(info.examples) > 0 || len(info.argtypes) > 0 {
				fmt.Printf("\tgroup \"%s\" \n", info.name) // name
				fmt.Printf("\t\"%s\"\n", info.docstring)   // docstring

				fmt.Print("\t{\n") // args block - contains metadata and arg descriptions

				// Output argsn and pure if available (inside args block)
				if info.argsn > 0 {
					fmt.Printf("\t\targsn %d\n", info.argsn)
				}
				if info.pure {
					fmt.Printf("\t\tpure\n")
				}

				// Output argument types extracted from MakeArgError (inside args block)
				if len(info.argtypes) > 0 {
					fmt.Print("\t\targtypes {\n")
					// Sort by arg number for consistent output
					keys := make([]int, 0, len(info.argtypes))
					for k := range info.argtypes {
						keys = append(keys, k)
					}
					sort.Ints(keys)
					for _, argNum := range keys {
						types := info.argtypes[argNum]
						fmt.Printf("\t\t\t%d [ %s ]\n", argNum, strings.Join(types, " "))
					}
					fmt.Print("\t\t}\n")
				}

				// Arg descriptions from comments
				for _, t := range info.args {
					fmt.Println("\t\targ `" + t + "`")
				}
				if len(info.returns) > 0 {
					fmt.Println("\t\treturns `" + info.returns + "`")
				}
				fmt.Println("\t}\n")

				// Tests block
				fmt.Print("\t{\n")
				for _, t := range info.tests {
					fmt.Println("\t\t" + t)
				}
				fmt.Println("\t}\n")

				// Examples block
				if len(info.examples) > 0 {
					fmt.Print("\t{\n`")
					for _, t := range info.examples {
						fmt.Println(t)
					}
					fmt.Println("`\t}\n")
				} else {
					fmt.Println("\t{\n\t}\n")
				}
			}
		}
		fmt.Println("}\n")
	}
}

func outputMissing(sections *[]builtinSection) {
	fmt.Println("missing {") // name
	for _, section := range *sections {
		for _, info := range section.builtins {
			if len(info.tests) == 0 {
				fmt.Printf("\t %s %q\n", info.name, info.docstring) // docstring
			}
		}
	}
	fmt.Println("}\n")
}

func outputStats(cnt counters) {
	fmt.Println("stats {\n") // name
	fmt.Printf("\tfunctions       \t%d\n", cnt.functions)
	fmt.Printf("\ttested-functions\t%d\n", cnt.tested_functions)
	fmt.Printf("\ttests           \t%d\n", cnt.tests)
	fmt.Printf("\texamples        \t%d\n", cnt.examples)
	fmt.Printf("\n")
	fmt.Printf("\ttest-coverage   \t%.1f%%\n", 100*float64(cnt.tested_functions)/float64(cnt.functions))
	fmt.Printf("\ttests-per-func  \t%.1f\n", float64(cnt.tests)/float64(cnt.tested_functions))
	fmt.Println("}\n")
}

var (
	// fileName = flag.String("fiimle", "", "Path to the Rye file (default: none)")
	stats   = flag.Bool("stats", false, "Show stats about builtins file")
	ls      = flag.Bool("ls", false, "List builtins files")
	missing = flag.Bool("missing", false, "Lists functions missing the tests")
	help    = flag.Bool("help", false, "Displays this help message.")
)

func main() {

	flag.Usage = func() {
		fmt.Println("╭────────────────────────────────────────────────────────────────────────────────────────────---")
		fmt.Println("│ \033[1mrbit - rye builtin info tool - https://ryelang.org")
		fmt.Println("╰───────────────────────────────────────────────────────────────────────────────────────---")
		fmt.Println("\n Usage: \033[1mparse\033[0m [\033[1moptions\033[0m] [\033[1mfilename\033[0m]")
		flag.PrintDefaults()
		fmt.Println("\033[33m  rbit                                                       \033[36m# shows this help")
		fmt.Println("\033[33m  rbit ../../evaldo/builtins.go > ../../info/base.info.rye   \033[36m# generates the info file")
		fmt.Println("\033[33m  rbit -stats ../../evaldo/builtins.go                       \033[36m# gets builtins file stats")
		fmt.Println("\033[33m  rbit -ls ../../evaldo/                                     \033[36m# lists builtin files")
		fmt.Println("\033[33m  rbit -help                                                 \033[36m# shows this help")
		fmt.Println("\033[0m\n Thank you for trying out \033[1mRye\033[22m ...")
		fmt.Println("")
	}
	// Parse flags
	flag.Parse()
	args := flag.Args()

	if flag.NFlag() == 0 && flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	} else if *help {
		flag.Usage()
		os.Exit(0)
	} else if *ls {
		fmt.Println("TODO 1")
	} else {
		doParsing(args)
	}
	// asd
}

// extractArgTypesFromFn walks the Fn body and finds all MakeArgError calls
// to extract the allowed types for each argument
func extractArgTypesFromFn(fnBody *ast.BlockStmt) map[int][]string {
	argtypes := make(map[int][]string)

	ast.Inspect(fnBody, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if this is a MakeArgError call
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "MakeArgError" {
			return true
		}

		// Need at least 3 args: ps, argNum, types
		if len(call.Args) < 3 {
			return true
		}

		// Extract argument number (Args[1])
		argNumLit, ok := call.Args[1].(*ast.BasicLit)
		if !ok || argNumLit.Kind != token.INT {
			return true
		}
		argNum, _ := strconv.Atoi(argNumLit.Value)

		// Extract type slice (Args[2])
		typesLit, ok := call.Args[2].(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Extract each type from the slice
		var types []string
		for _, elt := range typesLit.Elts {
			if sel, ok := elt.(*ast.SelectorExpr); ok {
				// e.g., env.StringType → "String"
				typeName := strings.TrimSuffix(sel.Sel.Name, "Type")
				types = append(types, typeName)
			}
		}

		// Merge with existing types for this arg (some builtins have multiple error paths)
		existing := argtypes[argNum]
		for _, t := range types {
			if !containsString(existing, t) {
				existing = append(existing, t)
			}
		}
		argtypes[argNum] = existing

		return true
	})

	return argtypes
}

// containsString checks if a string slice contains a specific string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func doParsing(args []string) {
	/// ###

	stemp := make([]builtinSection, 0)
	sectionList := &stemp
	// infoList := make([]builtinInfo, 0)

	if len(args) < 1 {
		fmt.Println("File argument missing")
		return
	}
	// Get the filename from the first argument
	filename := args[0]

	// Create a new token file set
	fset := token.NewFileSet()

	// Parse the Go source code into an AST
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing Go code:", err)
		return
	}

	c := counters{0, 0, 0, 0}

	var section *builtinSection

	section = &builtinSection{"Default", "", make([]builtinInfo, 0)}

	// Traverse the AST and find map literals
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			// Check if the literal is a map type
			if mapType, ok := x.Type.(*ast.MapType); ok {
				// Check if it's map[string]*env.Builtin or named Builtins_*
				isBuiltinMap := false

				// Check the map type: map[string]*env.Builtin
				if keyIdent, ok := mapType.Key.(*ast.Ident); ok {
					if keyIdent.Name == "string" {
						// Check if value is *env.Builtin
						if starExpr, ok := mapType.Value.(*ast.StarExpr); ok {
							if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
								if xIdent, ok := selExpr.X.(*ast.Ident); ok {
									if xIdent.Name == "env" && selExpr.Sel.Name == "Builtin" {
										isBuiltinMap = true
									}
								}
							}
						}
					}
				}

				// If not a builtin map by type, skip processing
				if !isBuiltinMap {
					return true
				}

				// Process each key-value pair in the map
				for _, elt := range x.Elts {
					info := builtinInfo{}
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						if key, ok := kv.Key.(*ast.BasicLit); ok {
							// Extract the key

							c.functions = c.functions + 1
							/// fmt.Printf("Key: %s\n", key.Value)
							// TODO NEXT - parse key into two values
							info.name = key.Value[1 : len(key.Value)-1]
							// Get comments above the key
							comment := getCommentsAboveKey(fset, node.Comments, key.Pos())
							if comment != "" {
								/// fmt.Printf("Comment above key: %s\n", strings.TrimSpace(comment))
								info, tempSection := parseCommentsAboveKey(comment, &info)
								if tempSection != nil {
									// fmt.Println(tempSection)
									if len(section.builtins) > 0 {
										*sectionList = append(*sectionList, *section)
									}
									section = tempSection
								}
								if len(info.tests) > 0 {
									c.tested_functions = c.tested_functions + 1
									c.tests = c.tests + len(info.tests)
								}
							}

							// Extract fields from the Builtin struct
							if compLit, ok := kv.Value.(*ast.CompositeLit); ok {
								for _, elt := range compLit.Elts {
									if kvField, ok := elt.(*ast.KeyValueExpr); ok {
										if keyField, ok := kvField.Key.(*ast.Ident); ok {
											switch keyField.Name {
											case "Doc":
												if docValue, ok := kvField.Value.(*ast.BasicLit); ok {
													// Extract the Doc value (removing quotes)
													docString := docValue.Value[1 : len(docValue.Value)-1]
													info.docstring = docString
												}
											case "Argsn":
												if argsValue, ok := kvField.Value.(*ast.BasicLit); ok {
													if argsValue.Kind == token.INT {
														info.argsn, _ = strconv.Atoi(argsValue.Value)
													}
												}
											case "Pure":
												if pureIdent, ok := kvField.Value.(*ast.Ident); ok {
													info.pure = pureIdent.Name == "true"
												}
											case "Fn":
												// Extract argument types from MakeArgError calls in the function body
												if fnLit, ok := kvField.Value.(*ast.FuncLit); ok {
													info.argtypes = extractArgTypesFromFn(fnLit.Body)
												}
											}
										}
									}
								}
							}
						}
					}
					section.builtins = append(section.builtins, info)
				}
			}
		}
		return true
	})

	// fmt.Println(section)
	*sectionList = append(*sectionList, *section)

	//	fmt.Println(infoList)

	//	fmt.Println("===================================================")

	if *stats {
		outputStats(c)
	} else if *missing {
		outputMissing(sectionList)
	} else {
		outputInfo(sectionList)
	}

	// 	fmt.Println("===================================================")

	// fmt.Println(c)
}
