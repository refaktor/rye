// +build b_tiny

package evaldo

import "C"

import (
	"rye/env"
)

var Builtins_webview = map[string]*env.Builtin{}
