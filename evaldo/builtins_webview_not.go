// +build !b_webview

package evaldo

import "C"

import (
	"rye/env"
)

var Builtins_webview = map[string]*env.Builtin{}
