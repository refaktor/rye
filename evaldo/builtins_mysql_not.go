//go:build no_mysql
// +build no_mysql

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_mysql = map[string]*env.Builtin{}
