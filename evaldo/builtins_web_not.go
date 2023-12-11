//go:build !b_echo
// +build !b_echo

package evaldo

// import "C"

import (
	"rye/env"
)

var OutBuffer = "" // how does this work with multiple threads / ... in server use ... probably we would need some per environment variable, not global / global?

func PopOutBuffer() string {
	r := OutBuffer
	OutBuffer = ""
	return r
}

var Builtins_web = map[string]*env.Builtin{}
