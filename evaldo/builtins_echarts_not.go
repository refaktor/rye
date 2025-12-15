//go:build no_echarts
// +build no_echarts

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_echarts = map[string]*env.Builtin{}
