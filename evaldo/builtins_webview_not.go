//go:build !b_webview
// +build !b_webview

package evaldo

import (
	"rye/env"
)

var Builtins_webview = map[string]*env.Builtin{}
