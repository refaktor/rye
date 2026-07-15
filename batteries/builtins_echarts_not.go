//go:build no_echarts
// +build no_echarts

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_echarts = map[string]*env.Builtin{}
