//go:build !b_bson
// +build !b_bson

package evaldo

import (
	"rye/env"
)

var Builtins_bsona = map[string]*env.Builtin{}
