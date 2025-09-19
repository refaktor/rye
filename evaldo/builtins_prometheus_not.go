//go:build no_prometheus
// +build no_prometheus

package evaldo

import (
	"github.com/refaktor/rye/env"
)

// Builtins_prometheus is a placeholder map when Prometheus support is not enabled
var Builtins_prometheus = map[string]*env.Builtin{
	"prometheus-not-enabled": {
		Argsn: 0,
		Doc:   "Placeholder function when Prometheus support is not enabled",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return MakeBuiltinError(ps, "Prometheus support is not enabled. Rebuild Rye with -tags prometheus", "prometheus-not-enabled")
		},
	},
}
