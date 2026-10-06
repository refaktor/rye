//go:build no_markdown
// +build no_markdown

package evaldo

import "github.com/refaktor/rye/env"

var Builtins_markdown = map[string]*env.Builtin{}

// MarkdownDisplayItems is unavailable when Markdown support is disabled.
func MarkdownDisplayItems(source string) []interface{} { return nil }
