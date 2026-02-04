//go:build no_docker
// +build no_docker

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_docker = map[string]*env.Builtin{}
