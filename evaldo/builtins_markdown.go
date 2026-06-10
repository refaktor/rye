//go:build !no_markdown
// +build !no_markdown

package evaldo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
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

// Helper function to extract headings from markdown text
func extractHeadings(source string) []*env.String {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
	)

	doc := md.Parser().Parse(text.NewReader([]byte(source)))
	var headings []*env.String

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if heading, ok := n.(*ast.Heading); ok {
				headingText := string(heading.Text([]byte(source)))
				headings = append(headings, env.NewString(headingText))
			}
		}
		return ast.WalkContinue, nil
	})

	return headings
}

// Helper function to extract paragraphs from markdown text
func extractParagraphs(source string) []*env.String {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
	)

	doc := md.Parser().Parse(text.NewReader([]byte(source)))
	var paragraphs []*env.String

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if para, ok := n.(*ast.Paragraph); ok {
				var buf bytes.Buffer
				for c := para.FirstChild(); c != nil; c = c.NextSibling() {
					if textNode, ok := c.(*ast.Text); ok {
						buf.Write(textNode.Text([]byte(source)))
					} else {
						buf.WriteString(" ")
					}
				}
				if buf.Len() > 0 {
					paragraphs = append(paragraphs, env.NewString(strings.TrimSpace(buf.String())))
				}
			}
		}
		return ast.WalkContinue, nil
	})

	return paragraphs
}

// Helper function to extract links from markdown text
func extractLinks(source string) []map[string]*env.String {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
	)

	doc := md.Parser().Parse(text.NewReader([]byte(source)))
	var links []map[string]*env.String

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if link, ok := n.(*ast.Link); ok {
				destination := string(link.Destination)
				var text string
				for c := link.FirstChild(); c != nil; c = c.NextSibling() {
					if textNode, ok := c.(*ast.Text); ok {
						text = string(textNode.Text([]byte(source)))
						break
					}
				}

				linkMap := make(map[string]*env.String)
				linkMap["text"] = env.NewString(text)
				linkMap["url"] = env.NewString(destination)
				links = append(links, linkMap)
			}
		}
		return ast.WalkContinue, nil
	})

	return links
}

// Helper function to extract code blocks from markdown text
func extractCodeBlocks(source string) []map[string]*env.String {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
	)

	doc := md.Parser().Parse(text.NewReader([]byte(source)))
	var codeBlocks []map[string]*env.String

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			var lang string
			var content string

			if fenced, ok := n.(*ast.FencedCodeBlock); ok {
				lang = string(fenced.Language([]byte(source)))
				var buf bytes.Buffer
				for i := 0; i < fenced.Lines().Len(); i++ {
					line := fenced.Lines().At(i)
					buf.Write(line.Value([]byte(source)))
				}
				content = buf.String()
			} else if codeBlock, ok := n.(*ast.CodeBlock); ok {
				var buf bytes.Buffer
				for i := 0; i < codeBlock.Lines().Len(); i++ {
					line := codeBlock.Lines().At(i)
					buf.Write(line.Value([]byte(source)))
				}
				content = buf.String()
			}

			if content != "" {
				codeMap := make(map[string]*env.String)
				codeMap["content"] = env.NewString(content)
				codeMap["language"] = env.NewString(lang)
				codeBlocks = append(codeBlocks, codeMap)
			}
		}
		return ast.WalkContinue, nil
	})

	return codeBlocks
}

// Helper function to extract images from markdown text
func extractImages(source string) []map[string]*env.String {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
	)

	doc := md.Parser().Parse(text.NewReader([]byte(source)))
	var images []map[string]*env.String

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if img, ok := n.(*ast.Image); ok {
				destination := string(img.Destination)
				var alt string
				for c := img.FirstChild(); c != nil; c = c.NextSibling() {
					if textNode, ok := c.(*ast.Text); ok {
						alt = string(textNode.Text([]byte(source)))
						break
					}
				}

				imgMap := make(map[string]*env.String)
				imgMap["alt"] = env.NewString(alt)
				imgMap["src"] = env.NewString(destination)
				images = append(images, imgMap)
			}
		}
		return ast.WalkContinue, nil
	})

	return images
}

func markdownNodeText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	extractText(n, source, &buf)
	return buf.String()
}

func markdownParagraphText(node *ast.Paragraph, source []byte) string {
	var buf bytes.Buffer
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		switch child := c.(type) {
		case *ast.Text:
			text := child.Text(source)
			buf.Write(text)
			raw := string(child.Segment.Value(source))
			if child.HardLineBreak() || strings.HasSuffix(raw, "\n") {
				buf.WriteByte('\n')
			}
		case *ast.Link:
			linkText := markdownNodeText(child, source)
			if linkText != "" {
				buf.WriteString(linkText)
			}
		case *ast.Emphasis:
			emphText := markdownNodeText(child, source)
			if emphText != "" {
				buf.WriteString(emphText)
			}
		case *ast.CodeSpan:
			codeText := string(child.Text(source))
			if codeText != "" {
				buf.WriteString(codeText)
			}
		}
	}

	return buf.String()
}

// wrapTextToWidth wraps text to fit within the specified width, properly handling line breaks
func wrapTextToWidth(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	lines := strings.Split(text, "\n")
	wrappedLines := make([]string, 0)

	for _, line := range lines {
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Wrap long lines
		for len(line) > width {
			// Find the best break point (space or punctuation)
			breakPoint := width
			for i := width - 1; i >= width/2; i-- {
				if line[i] == ' ' || line[i] == '\t' || line[i] == '-' {
					breakPoint = i + 1
					break
				}
			}

			wrappedLines = append(wrappedLines, strings.TrimRight(line[:breakPoint], " \t"))
			line = strings.TrimLeft(line[breakPoint:], " \t")
		}

		if len(line) > 0 {
			wrappedLines = append(wrappedLines, line)
		}
	}

	return wrappedLines
}

// MarkdownDisplayItem represents a selectable markdown element with its type and content
type MarkdownDisplayItem struct {
	Type         string   // "heading", "paragraph", "code", "list", "quote", "hr", "link"
	Content      string   // Raw content
	DisplayLines []string // Formatted lines for display
	Level        int      // For headings (1-6), or 0 for others
	Language     string   // For code blocks
}

// convertMarkdownDisplayItems converts MarkdownDisplayItem slice to interface{} slice for DisplayMarkdownItems
func convertMarkdownDisplayItems(items []MarkdownDisplayItem) []interface{} {
	converted := make([]interface{}, len(items))
	for i, item := range items {
		itemMap := map[string]interface{}{
			"Type":         item.Type,
			"Content":      item.Content,
			"DisplayLines": item.DisplayLines,
			"Level":        item.Level,
			"Language":     item.Language,
		}
		converted[i] = itemMap
	}
	return converted
}

func markdownDisplayItems(source string) []MarkdownDisplayItem {
	sourceBytes := []byte(source)
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
	)

	// Get terminal width for proper text wrapping
	size, err := term.GetTerminalSize()
	width := size.Width
	if err != nil || width <= 0 {
		width = 80 // Fallback default
	}
	// Reserve some space for display margins and cursor
	displayWidth := width - 4

	doc := md.Parser().Parse(text.NewReader(sourceBytes))
	items := make([]MarkdownDisplayItem, 0)

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			headingText := strings.TrimSpace(string(node.Text(sourceBytes)))
			if headingText != "" {
				fullHeading := "H" + strconv.Itoa(node.Level) + ": " + headingText
				// Wrap heading if it's too long
				wrappedLines := wrapTextToWidth(fullHeading, displayWidth)
				displayLines := make([]string, len(wrappedLines))
				for i, line := range wrappedLines {
					displayLines[i] = term.StrBold() + line + term.StrCloseProps()
				}
				
				item := MarkdownDisplayItem{
					Type:         "heading",
					Content:      headingText,
					DisplayLines: displayLines,
					Level:        node.Level,
				}
				items = append(items, item)
			}
		case *ast.Paragraph:
			child := node.FirstChild()
			if child != nil && child.NextSibling() == nil {
				if link, ok := child.(*ast.Link); ok {
					linkText := markdownNodeText(link, sourceBytes)
					destination := string(link.Destination)
					if linkText == "" {
						linkText = destination
					}
					linkLine := fmt.Sprintf("Link: %s -> %s", linkText, destination)
					// Wrap link if it's too long
					wrappedLines := wrapTextToWidth(linkLine, displayWidth)
					
					item := MarkdownDisplayItem{
						Type:         "link",
						Content:      destination,
						DisplayLines: wrappedLines,
					}
					items = append(items, item)
					return ast.WalkSkipChildren, nil
				}
			}

			paragraph := markdownParagraphText(node, sourceBytes)
			if paragraph == "" {
				item := MarkdownDisplayItem{
					Type:         "paragraph",
					Content:      "",
					DisplayLines: []string{""},
				}
				items = append(items, item)
				return ast.WalkContinue, nil
			}
			
			// Remove trailing newline for wrapping
			paragraph = strings.TrimRight(paragraph, "\n")
			
			// Wrap paragraph text to fit terminal width
			wrappedLines := wrapTextToWidth(paragraph, displayWidth)
			
			item := MarkdownDisplayItem{
				Type:         "paragraph",
				Content:      paragraph,
				DisplayLines: wrappedLines,
			}
			items = append(items, item)
		case *ast.CodeBlock, *ast.FencedCodeBlock:
			// Handle code blocks
			var content string
			var language string
			if fenced, ok := node.(*ast.FencedCodeBlock); ok {
				language = string(fenced.Language(sourceBytes))
				var buf bytes.Buffer
				for i := 0; i < fenced.Lines().Len(); i++ {
					line := fenced.Lines().At(i)
					buf.Write(line.Value(sourceBytes))
				}
				content = buf.String()
			} else if codeBlock, ok := node.(*ast.CodeBlock); ok {
				var buf bytes.Buffer
				for i := 0; i < codeBlock.Lines().Len(); i++ {
					line := codeBlock.Lines().At(i)
					buf.Write(line.Value(sourceBytes))
				}
				content = buf.String()
			}
			
			if content != "" {
				codeLines := strings.Split(strings.TrimRight(content, "\n"), "\n")
				displayLines := make([]string, len(codeLines))
				for i, line := range codeLines {
					// Don't wrap code lines, just truncate if too long
					if len(line) > displayWidth {
						line = line[:displayWidth-3] + "..."
					}
					displayLines[i] = "  " + line // Indent code
				}
				
				item := MarkdownDisplayItem{
					Type:         "code",
					Content:      content,
					DisplayLines: displayLines,
					Language:     language,
				}
				items = append(items, item)
			}
		case *ast.List:
			// Handle lists as complete blocks - collect all list items
			var listItems []string
			var listContent strings.Builder
			
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if listItem, ok := c.(*ast.ListItem); ok {
					var itemText bytes.Buffer
					for ic := listItem.FirstChild(); ic != nil; ic = ic.NextSibling() {
						if para, ok := ic.(*ast.Paragraph); ok {
							for pc := para.FirstChild(); pc != nil; pc = pc.NextSibling() {
								if text, ok := pc.(*ast.Text); ok {
									itemText.Write(text.Text(sourceBytes))
								}
							}
						}
					}
					
					if itemText.Len() > 0 {
						itemContent := itemText.String()
						listItems = append(listItems, itemContent)
						listContent.WriteString("• " + itemContent + "\n")
					}
				}
			}
			
			if len(listItems) > 0 {
				displayLines := make([]string, 0)
				for _, itemContent := range listItems {
					prefix := "• "
					itemWidth := displayWidth - len(prefix)
					wrappedLines := wrapTextToWidth(itemContent, itemWidth)
					
					for i, line := range wrappedLines {
						if i == 0 {
							displayLines = append(displayLines, prefix+line)
						} else {
							displayLines = append(displayLines, "  "+line)
						}
					}
				}
				
				item := MarkdownDisplayItem{
					Type:         "list",
					Content:      strings.TrimRight(listContent.String(), "\n"),
					DisplayLines: displayLines,
				}
				items = append(items, item)
			}
			
			return ast.WalkSkipChildren, nil
		case *ast.Blockquote:
			// Handle blockquotes
			var buf bytes.Buffer
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if para, ok := c.(*ast.Paragraph); ok {
					for pc := para.FirstChild(); pc != nil; pc = pc.NextSibling() {
						if text, ok := pc.(*ast.Text); ok {
							buf.Write(text.Text(sourceBytes))
						}
					}
				}
			}
			
			if buf.Len() > 0 {
				quoteText := strings.TrimSpace(buf.String())
				prefix := "> "
				quoteWidth := displayWidth - len(prefix)
				wrappedLines := wrapTextToWidth(quoteText, quoteWidth)
				
				displayLines := make([]string, len(wrappedLines))
				for i, line := range wrappedLines {
					displayLines[i] = prefix + line
				}
				
				item := MarkdownDisplayItem{
					Type:         "quote",
					Content:      quoteText,
					DisplayLines: displayLines,
				}
				items = append(items, item)
			}
		case *ast.ThematicBreak:
			// Handle horizontal rules
			ruler := strings.Repeat("─", min(displayWidth, 40))
			item := MarkdownDisplayItem{
				Type:         "hr",
				Content:      "---",
				DisplayLines: []string{ruler},
			}
			items = append(items, item)
		}

		return ast.WalkContinue, nil
	})

	return items
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var Builtins_markdown = map[string]*env.Builtin{

	//
	// ##### Markdown ##### "Functions for processing and converting Markdown documents"
	//
	// Tests:
	// ; Tested manually since it requires file input
	// Args:
	// * reader: File reader or similar object to read markdown from
	// * handlers: Block containing section handlers (h1, h2, paragraph, code, etc.)
	// Returns:
	// * the result of processing the markdown document
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

	// Create a new Markdown value from string
	"markdown": {
		Argsn: 1,
		Doc:   "Creates a Markdown value from a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch text := arg0.(type) {
			case env.String:
				return *env.NewMarkdown(text.Value)
			case env.Markdown:
				return text
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.MarkdownType}, "markdown")
			}
		},
	},

	// Get raw markdown text from Markdown value
	"markdown//text": {
		Argsn: 1,
		Doc:   "Gets the raw markdown text from a Markdown value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			md, ok := arg0.(env.Markdown)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.MarkdownType}, "markdown//text")
			}
			return *env.NewString(md.Value)
		},
	},

	// Get length of markdown text
	"markdown//length": {
		Argsn: 1,
		Doc:   "Gets the length of the markdown text in characters.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			md, ok := arg0.(env.Markdown)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.MarkdownType}, "markdown//length")
			}
			return *env.NewInteger(int64(len(md.Value)))
		},
	},

	// Convert Markdown value to HTML
	"markdown//to-html": {
		Argsn: 1,
		Doc:   "Converts a Markdown value to HTML.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			md, ok := arg0.(env.Markdown)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.MarkdownType}, "markdown//to-html")
			}

			html, err := markdown_to_html(ps, md.Value)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "markdown//to-html")
			}

			return *env.NewString(html)
		},
	},

	// Extract headings from markdown text
	"markdown//headings": {
		Argsn: 1,
		Doc:   "Extracts all headings from markdown text as a list of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			md, ok := arg0.(env.Markdown)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.MarkdownType}, "markdown//headings")
			}

			headings := extractHeadings(md.Value)
			result := make([]any, len(headings))
			for i, heading := range headings {
				result[i] = heading
			}
			return *env.NewList(result)
		},
	},

	// Extract paragraphs from markdown text
	"markdown//paragraphs": {
		Argsn: 1,
		Doc:   "Extracts all paragraphs from markdown text as a list of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			md, ok := arg0.(env.Markdown)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.MarkdownType}, "markdown//paragraphs")
			}

			paragraphs := extractParagraphs(md.Value)
			result := make([]any, len(paragraphs))
			for i, paragraph := range paragraphs {
				result[i] = paragraph
			}
			return *env.NewList(result)
		},
	},

	// Extract links from markdown text
	"markdown//links": {
		Argsn: 1,
		Doc:   "Extracts all links from markdown text as a list of dictionaries with 'text' and 'url' keys.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			md, ok := arg0.(env.Markdown)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.MarkdownType}, "markdown//links")
			}

			links := extractLinks(md.Value)
			result := make([]any, len(links))
			for i, link := range links {
				linkDict := make(map[string]any)
				linkDict["text"] = link["text"]
				linkDict["url"] = link["url"]
				result[i] = linkDict
			}
			return *env.NewList(result)
		},
	},
}
