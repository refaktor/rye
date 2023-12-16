//go:build !b_sqlite
// +build !b_sqlite

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_sqlite = map[string]*env.Builtin{}
