package main

import (
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

	// info := builtinInfo{}

	// Step 1: Split input into lines and trim whitespace
	lines := strings.Split(strings.TrimSpace(input), "\n")

	// Step 2: Separate header and tests
	var headerLines []string
	//	var testLines []string
	isTestSection := false

	fmt.Println("!!!!!!!!!!!!!!!**************")

	for _, line := range lines {
		line = strings.TrimSpace(line) // Remove leading and trailing whitespace
		fmt.Println("LLLL:" + line)
		fmt.Println(line)
		if line == "Tests:" {
			fmt.Println("***** TEST ****")
			isTestSection = true
			continue
		}

		if isTestSection {
			fmt.Println(line)
			info.tests = append(info.tests, line)
		} else {
			headerLines = append(headerLines, line)
		}
	}
	// Step 3: Combine the header lines into a single string
	info.doc = strings.Join(headerLines, "\n")
	return *info
}

func main() {

	infoList := make([]builtinInfo, 0)

	// Check if a filename is provided as an argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		return
	}

	// Get the filename from the first argument
	filename := os.Args[1]

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

							fmt.Printf("Key: %s\n", key.Value)
							// TODO NEXT - parse key into two values
							info.name = key.Value
							// Get comments above the key
							comment := getCommentsAboveKey(fset, node.Comments, key.Pos())
							if comment != "" {
								fmt.Printf("Comment above key: %s\n", strings.TrimSpace(comment))
								info = parseCommentsAboveKey(comment, &info)
								c.functions = c.functions + 1
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

	fmt.Println(infoList)

	fmt.Println("********")

	fmt.Println(c)
}
