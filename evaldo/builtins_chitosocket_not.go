//go:build !linux || no_chitosocket
// +build !linux no_chitosocket

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_chitosocket = map[string]*env.Builtin{}
