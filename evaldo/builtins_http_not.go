//go:build !b_http
// +build !b_http

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_http = map[string]*env.Builtin{}
