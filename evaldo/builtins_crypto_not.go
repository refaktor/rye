//go:build no_crypto
// +build no_crypto

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_crypto = map[string]*env.Builtin{}
