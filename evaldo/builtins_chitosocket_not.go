//go:build no_chitosocket
// +build no_chitosocket

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_chitosocket = map[string]*env.Builtin{}
