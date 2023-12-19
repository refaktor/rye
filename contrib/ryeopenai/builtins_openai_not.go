//go:build !b_openai
// +build !b_openai

package ryeopenai

import (
	"github.com/refaktor/rye/env"
)

var Builtins_openai = map[string]*env.Builtin{}
