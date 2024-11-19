package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// data that we want to optionally extract from builtins code and store as a general structure that
// * Rye runtime will be able to provide to the user
// * Test tool will run tests on
type builtinInfo struct {
	name      string     // from key value
	gentype   string     // optional from key value
	docstring string     // part of builtin definition
	doc       string     // free text at the top of the comment
	nargs     int        // part of builtin definition
	args      []string   // extracted from comment or variable names
	argtypes  [][]string // extracted from switch statements or conversions
	tests     []string   // extracted from comment
	examples  []string   // extracted from comment
	tags      []string   // extracted from comment
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
func parseCommentsAboveKey(input string, info *builtinInfo) builtinInfo {

	const (
		inDoc = iota
		inTests
		inExamples
		inArgs
	)
	position := inDoc

	// info := builtinInfo{}

	// Step 1: Split input into lines and trim whitespace
	lines := strings.Split(strings.TrimSpace(input), "\n")

	// Step 2: Separate header and tests
	var headerLines []string
	//	var testLines []string

	//	fmt.Println("!!!!!!!!!!!!!!!**************")

	for _, line := range lines {
		line = strings.TrimSpace(line) // Remove leading and trailing whitespace
		// fmt.Println("LLLL:" + line)
		// fmt.Println(line)
		switch line {
		case "Tests:":
			position = inTests
			continue
		case "Examples:":
			position = inExamples
			continue
		case "Args:":
			position = inArgs
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
			info.args = append(info.args, line)
		}
	}
	// Step 3: Combine the header lines into a single string
	info.doc = strings.Join(headerLines, "\n")
	return *info
}

func outputInfo(infos []builtinInfo) {
	fmt.Println("section \"base\" \"base text\" {\n") // name
	for _, info := range infos {
		if len(info.tests) > 0 {
			fmt.Printf("\tgroup %s \n", info.name)   // name
			fmt.Printf("\t\"%s\"\n", info.docstring) // docstring

			fmt.Print("\t{\n") // args
			for _, t := range info.args {
				fmt.Println("\t\targ \"" + t + "\"")
			}
			fmt.Println("\t}\n")

			fmt.Print("\t{\n")
			for _, t := range info.tests {
				fmt.Println("\t\t" + t)
			}
			fmt.Println("\t}\n")
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
	stats = flag.Bool("stats", false, "Show stats about builtins file")
	ls    = flag.Bool("ls", false, "List builtins files")
	help  = flag.Bool("help", false, "Displays this help message.")
)

func main() {

	flag.Usage = func() {
		fmt.Println("╭────────────────────────────────────────────────────────────────────────────────────────────---")
		fmt.Println("│ \033[1mrbit - rye builtin info tool")
		fmt.Println("╰───────────────────────────────────────────────────────────────────────────────────────---")
		fmt.Println("\n Usage: \033[1mparse\033[0m [\033[1moptions\033[0m] [\033[1mfilename\033[0m or \033[1mcommand\033[0m]")
		flag.PrintDefaults()
		fmt.Println("\033[33m  rbit                                                         \033[36m# shows helo")
		fmt.Println("\033[33m  rbit ../../evaldo/builtins.go > ../../info/base.info.rye   \033[36m# generates the info file")
		fmt.Println("\033[33m  rbit -stats ../../evaldo/builtins.go                          \033[36m# gets bi coverage stats")
		fmt.Println("\033[33m  rbit -ls ../../evaldo/                                        \033[36m# lists bi files")
		fmt.Println("\033[0m\n Thank you for trying out \033[1mRye\033[22m ...")
		fmt.Println("")
	}
	// Parse flags
	flag.Parse()
	args := flag.Args()

	if flag.NFlag() == 0 && flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	} else if *ls {
		fmt.Println("TODO 1")
	} else {
		doParsing(args)
	}
	// asd
}

func doParsing(args []string) {
	/// ###

	infoList := make([]builtinInfo, 0)

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

	// Traverse the AST and find map literals
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			// Check if the literal is a map type
			if _, ok := x.Type.(*ast.MapType); ok {
				// Process each key-value pair in the map
				for _, elt := range x.Elts {
					info := builtinInfo{}
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						if key, ok := kv.Key.(*ast.BasicLit); ok {
							// Extract the key

							c.functions = c.functions + 1
							/// fmt.Printf("Key: %s\n", key.Value)
							// TODO NEXT - parse key into two values
							info.name = key.Value
							// Get comments above the key
							comment := getCommentsAboveKey(fset, node.Comments, key.Pos())
							if comment != "" {
								/// fmt.Printf("Comment above key: %s\n", strings.TrimSpace(comment))
								info = parseCommentsAboveKey(comment, &info)
								if len(info.tests) > 0 {
									c.tested_functions = c.tested_functions + 1
									c.tests = c.tests + len(info.tests)
								}
							}
						}
					}
					infoList = append(infoList, info)
				}
			}
		}
		return true
	})

	//	fmt.Println(infoList)

	//	fmt.Println("===================================================")

	if *stats {
		outputStats(c)
	} else {
		outputInfo(infoList)
	}

	// 	fmt.Println("===================================================")

	// fmt.Println(c)
}
