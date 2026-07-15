//go:build no_crypto
// +build no_crypto

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_crypto = map[string]*env.Builtin{}
