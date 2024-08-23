//go:build no_http
// +build no_http

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_http = map[string]*env.Builtin{}
