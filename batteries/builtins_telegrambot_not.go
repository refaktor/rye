//go:build no_telegram
// +build no_telegram

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_telegrambot = map[string]*env.Builtin{}
