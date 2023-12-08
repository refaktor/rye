//go:build !b_nats
// +build !b_nats

package evaldo

import (
	"fmt"
	"rye/env"
)

func strimp() { fmt.Println("") }

var Builtins_nats = map[string]*env.Builtin{}
