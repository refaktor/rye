//go:build !b_sqlite
// +build !b_sqlite

package evaldo

import (
	"rye/env"
)

var Builtins_sqlite = map[string]*env.Builtin{}
