//go:build !b_ollama
// +build !b_ollama

package ollama

import (
	"github.com/refaktor/rye/env"
)

var Builtins_ollama = map[string]*env.Builtin{}
