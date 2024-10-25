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

// Helper function to get comments above the map key
func getCommentsAboveKey(fset *token.FileSet, comments []*ast.CommentGroup, keyPos token.Pos) string {
	for _, commentGroup := range comments {
		if fset.Position(commentGroup.End()).Line == fset.Position(keyPos).Line-1 {
			return commentGroup.Text()
		}
	}
	return ""
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
							}
							// TODO NEXT - make a function that parses the comment extracting doc, args (opt), examples, tests and tags
							info.doc = strings.TrimSpace(comment)
						}
					}
					infoList = append(infoList, info)
				}
			}
		}
		return true
	})

	fmt.Println(infoList)
}
