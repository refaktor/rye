//go:build no_markdown
// +build no_markdown

package batteries

import (
	"github.com/refaktor/rye/env"
)

type MarkdownDisplayItem struct {
	Type         string
	Content      string
	DisplayLines []string
	Level        int
	Language     string
}

func convertMarkdownDisplayItems(items []MarkdownDisplayItem) []interface{} {
	return []interface{}{}
}

func markdownDisplayItems(source string) []MarkdownDisplayItem {
	return []MarkdownDisplayItem{}
}

var Builtins_markdown = map[string]*env.Builtin{}
