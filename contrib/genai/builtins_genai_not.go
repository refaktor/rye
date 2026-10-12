//go:build !b_genai
// +build !b_genai

package genai

import (
	"github.com/refaktor/rye/env"
)

var Builtins_genai = map[string]*env.Builtin{}