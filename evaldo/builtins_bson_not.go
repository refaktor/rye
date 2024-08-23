//go:build no_bson
// +build no_bson

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_bson = map[string]*env.Builtin{}
