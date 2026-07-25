//go:build no_mqtt
// +build no_mqtt

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_mqtt = map[string]*env.Builtin{}
