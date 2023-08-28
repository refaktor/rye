//go:build !b_crypto
// +build !b_crypto

package evaldo

import (
	"rye/env"
)

var Builtins_crypto = map[string]*env.Builtin{}
