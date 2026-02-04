//go:build no_docker || wasm || js
// +build no_docker wasm js

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_docker = map[string]*env.Builtin{}
