//go:build no_telegram
// +build no_telegram

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_telegrambot = map[string]*env.Builtin{}
