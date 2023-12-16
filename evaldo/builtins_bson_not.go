//go:build !b_bson
// +build !b_bson

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_bson = map[string]*env.Builtin{}
