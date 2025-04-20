//go:build !no_markdown
// +build !no_markdown

package evaldo

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Use the EmptyRM function from builtins_structures.go

// Parses the markdown handler block and creates a Dict with handlers for different markdown elements
func load_markdown_Dict(ps *env.ProgramState, block env.Block) (env.Dict, *env.Error) {
	data := make(map[string]any)
	rmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Word:
			key := ps.Idx.GetWord(obj1.Index)
			block.Series.Next()
			if nextObj, ok := block.Series.Peek().(env.Block); ok {
				block.Series.Next()
				rmap.Data[key] = nextObj
			} else {
				return EmptyRM(), MakeBuiltinError(ps, "Expected block after markdown section specifier.", "reader//do-markdown")
			}
		default:
			return EmptyRM(), MakeBuiltinError(ps, "Expected word specifying markdown section.", "reader//do-markdown")
		}
	}
	return rmap, nil
}

// Main function to process markdown content using goldmark
func do_markdown(ps *env.ProgramState, reader env.Object, rmap env.Dict) env.Object {
	file, ok := reader.(env.Native)
	if !ok {
		return MakeBuiltinError(ps, "Reader must be a file object.", "reader//do-markdown")
	}

	fileObj, ok := file.Value.(io.Reader)
	if !ok {
		return MakeBuiltinError(ps, "Reader must be a reader object.", "reader//do-markdown")
	}

	// Read all content from the reader
	content, err := ioutil.ReadAll(fileObj)
	if err != nil {
		return MakeBuiltinError(ps, err.Error(), "reader//do-markdown")
	}

	// Create a markdown parser with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,           // GitHub Flavored Markdown
			extension.Table,         // Tables
			extension.Strikethrough, // Strikethrough
			extension.TaskList,      // Task lists
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	// Parse the markdown content
	doc := md.Parser().Parse(text.NewReader(content))

	ser := ps.Ser
	// Process the AST
	processNode(ps, doc, content, rmap)

	ps.Ser = ser
	return ps.Res
}

// Process a markdown node and its children
func processNode(ps *env.ProgramState, n ast.Node, source []byte, rmap env.Dict) {
	// Process the current node
	switch node := n.(type) {
	case *ast.Document:
		// Document is the root node, just process its children
		// If there's a document handler, call it at the start
		if block, ok := rmap.Data["document"].(env.Block); ok {
			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString("start"), true)
		}

	case *ast.Heading:
		// Process headings (H1, H2, H3, etc.)
		level := node.Level
		headingKey := "h" + string(rune('0'+level))

		if block, ok := rmap.Data[headingKey].(env.Block); ok {
			text := string(node.Text(source))
			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(text), true)
		}

	case *ast.Paragraph:
		// Process paragraphs
		if block, ok := rmap.Data["paragraph"].(env.Block); ok {
			var buf bytes.Buffer
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if text, ok := c.(*ast.Text); ok {
					buf.Write(text.Text(source))
				} else {
					// Handle other types of nodes in paragraphs
					// Note: Hardbreak is not directly accessible, we'll handle line breaks differently
					buf.WriteString(" ")
				}
			}
			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(buf.String()), true)
		}

	case *ast.List:
		// Process lists
		listType := "unordered"
		if node.IsOrdered() {
			listType = "ordered"
		}

		if block, ok := rmap.Data["list"].(env.Block); ok {
			var items []string
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if item, ok := c.(*ast.ListItem); ok {
					var itemText bytes.Buffer
					for ic := item.FirstChild(); ic != nil; ic = ic.NextSibling() {
						if para, ok := ic.(*ast.Paragraph); ok {
							for pc := para.FirstChild(); pc != nil; pc = pc.NextSibling() {
								if text, ok := pc.(*ast.Text); ok {
									itemText.Write(text.Text(source))
								}
							}
						}
					}
					items = append(items, itemText.String())
				}
			}

			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(strings.Join(items, "\n")), true)
			EvalBlockInj(ps, *env.NewString(listType), true)
		}

	case *ast.ListItem:
		// Process individual list items if there's a specific handler
		if block, ok := rmap.Data["line-item"].(env.Block); ok {
			var itemText bytes.Buffer
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if para, ok := c.(*ast.Paragraph); ok {
					for pc := para.FirstChild(); pc != nil; pc = pc.NextSibling() {
						if text, ok := pc.(*ast.Text); ok {
							itemText.Write(text.Text(source))
						}
					}
				}
			}

			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(itemText.String()), true)
		}

	case *ast.CodeBlock, *ast.FencedCodeBlock:
		// Process code blocks
		var lang string
		var content string

		if fenced, ok := node.(*ast.FencedCodeBlock); ok {
			lang = string(fenced.Language(source))
			var buf bytes.Buffer
			for i := 0; i < fenced.Lines().Len(); i++ {
				line := fenced.Lines().At(i)
				buf.Write(line.Value(source))
			}
			content = buf.String()
		} else if codeBlock, ok := node.(*ast.CodeBlock); ok {
			var buf bytes.Buffer
			for i := 0; i < codeBlock.Lines().Len(); i++ {
				line := codeBlock.Lines().At(i)
				buf.Write(line.Value(source))
			}
			content = buf.String()
		}

		if block, ok := rmap.Data["code"].(env.Block); ok {
			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(content), true)
			if lang != "" {
				EvalBlockInj(ps, *env.NewString(lang), true)
			}
		}

	case *ast.Blockquote:
		// Process blockquotes
		if block, ok := rmap.Data["blockquote"].(env.Block); ok {
			var buf bytes.Buffer
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if para, ok := c.(*ast.Paragraph); ok {
					for pc := para.FirstChild(); pc != nil; pc = pc.NextSibling() {
						if text, ok := pc.(*ast.Text); ok {
							buf.Write(text.Text(source))
						}
					}
					buf.WriteString("\n")
				}
			}

			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(strings.TrimSpace(buf.String())), true)
		}

	case *ast.Link:
		// Process links
		if block, ok := rmap.Data["link"].(env.Block); ok {
			destination := string(node.Destination)
			var text string
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if textNode, ok := c.(*ast.Text); ok {
					text = string(textNode.Text(source))
					break
				}
			}

			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(text), true)
			EvalBlockInj(ps, *env.NewString(destination), true)
		}

	case *ast.Image:
		// Process images
		if block, ok := rmap.Data["image"].(env.Block); ok {
			destination := string(node.Destination)
			var alt string
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if textNode, ok := c.(*ast.Text); ok {
					alt = string(textNode.Text(source))
					break
				}
			}

			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(alt), true)
			EvalBlockInj(ps, *env.NewString(destination), true)
		}

	// Tables are handled by the extension package, not directly in ast
	// We'll check for table nodes differently
	default:
		// Check if it's a table node from the extension
		if block, ok := rmap.Data["table"].(env.Block); ok {
			// For tables, we'll extract text content in a simpler way
			var tableText bytes.Buffer
			extractText(n, source, &tableText)

			if tableText.Len() > 0 {
				ps.Ser = block.Series
				EvalBlockInj(ps, *env.NewString(tableText.String()), true)
				// Since we can't get the row count directly, we'll estimate based on newlines
				rowCount := strings.Count(tableText.String(), "\n") + 1
				EvalBlockInj(ps, *env.NewInteger(int64(rowCount)), true)
			}
		}

	case *ast.ThematicBreak:
		// Process horizontal rules
		if block, ok := rmap.Data["hr"].(env.Block); ok {
			ps.Ser = block.Series
			EvalBlockInj(ps, env.Void{}, true)
		}

	case *ast.Text:
		// Process text nodes (usually handled by parent nodes)
		// This is mainly for inline text that's not part of a larger structure
		if block, ok := rmap.Data["text"].(env.Block); ok {
			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(string(node.Text(source))), true)
		}

	case *ast.Emphasis:
		// Process emphasis (italic)
		if block, ok := rmap.Data["italic"].(env.Block); ok {
			var text string
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if textNode, ok := c.(*ast.Text); ok {
					text = string(textNode.Text(source))
					break
				}
			}

			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(text), true)
		}

		// Handle emphasis nodes (bold, italic, etc.)
		// Instead of checking for specific node types that might not exist,
		// we'll check for the Kind property which is more reliable
	}

	// Check for emphasis nodes (bold, italic, etc.) based on node type
	// Instead of using Kind constants which might not be exported,
	// we'll check the node type directly
	if _, ok := n.(*ast.Emphasis); ok {
		if block, ok := rmap.Data["italic"].(env.Block); ok {
			var textBuf bytes.Buffer
			extractText(n, source, &textBuf)

			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString(textBuf.String()), true)
		}
	}

	// Check for task list items
	// Since we can't directly check for TaskCheckBox type, we'll look for list items with specific properties
	if listItem, ok := n.(*ast.ListItem); ok {
		if block, ok := rmap.Data["task"].(env.Block); ok {
			// For task list items, we'll check if it has any children that might indicate it's a task
			// This is a simplified approach since we can't directly access task attributes
			var isTask bool
			var isChecked bool

			// Extract the text to see if it starts with [ ] or [x]
			var itemText bytes.Buffer
			extractText(listItem, source, &itemText)
			text := itemText.String()

			if strings.HasPrefix(text, "[ ]") {
				isTask = true
				isChecked = false
			} else if strings.HasPrefix(text, "[x]") || strings.HasPrefix(text, "[X]") {
				isTask = true
				isChecked = true
			}

			if isTask {
				ps.Ser = block.Series
				EvalBlockInj(ps, *env.NewBoolean(isChecked), true)
			}
		}
	}

	// Process children recursively
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		processNode(ps, c, source, rmap)
	}

	// If there's a document handler, call it at the end for the document node
	if _, ok := n.(*ast.Document); ok {
		if block, ok := rmap.Data["document"].(env.Block); ok {
			ps.Ser = block.Series
			EvalBlockInj(ps, *env.NewString("end"), true)
		}
	}
}

// Helper function to extract text from a node and its children
func extractText(n ast.Node, source []byte, buf *bytes.Buffer) {
	if text, ok := n.(*ast.Text); ok {
		buf.Write(text.Text(source))
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		extractText(c, source, buf)
	}
}

// Add a function to parse markdown to HTML
func markdown_to_html(ps *env.ProgramState, source string) (string, error) {
	// Create a markdown parser with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,           // GitHub Flavored Markdown
			extension.Table,         // Tables
			extension.Strikethrough, // Strikethrough
			extension.TaskList,      // Task lists
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

var Builtins_markdown = map[string]*env.Builtin{
	// Main markdown processing function
	"reader//do-markdown": {
		Argsn: 2,
		Doc:   "Processes Markdown using a streaming approach with section handlers.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_markdown_Dict(ps, arg1.(env.Block))
			if err != nil {
				ps.FailureFlag = true
				return err
			}
			return do_markdown(ps, arg0, rm)
		},
	},

	// Convert markdown to HTML
	"markdown->html": {
		Argsn: 1,
		Doc:   "Converts Markdown text to HTML.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "markdown->html")
			}

			html, err := markdown_to_html(ps, text.Value)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "markdown->html")
			}

			return *env.NewString(html)
		},
	},
}
