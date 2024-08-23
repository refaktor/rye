//go:build no_psql
// +build no_psql

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_psql = map[string]*env.Builtin{}
