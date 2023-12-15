//go:build b_psutil
// +build b_psutil

package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"

	"github.com/shirou/gopsutil/process"

	"github.com/shirou/gopsutil/mem"
)

func something1() {

	fmt.Print("1")
}

// In request we return a raw-map, because it's very inside loop call, this is sparse call, and we get tons of fields, so it would be best
// to turn them to normal Rye map (which is now Env / later Context or something like it), and they query it from Rye.

var Builtins_ps = map[string]*env.Builtin{

	"ps/virtual-memory": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			v, _ := mem.VirtualMemory()
			r := env.NewDict(make(map[string]any, 3))
			r.Data["total"] = v.Total
			r.Data["free"] = v.Free
			r.Data["used-percent"] = v.UsedPercent
			return *r
		},
	},
	"ps/pids": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pids, _ := process.Pids()
			pids2 := make([]env.Object, len(pids))
			for i, p := range pids {
				pids2[i] = env.Integer{int64(p)}
			}
			return env.NewBlock(*env.NewTSeries(pids2))
		},
	},
	"ps/process": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pids, _ := process.Pids()
			pids2 := make([]env.Object, len(pids))
			for i, p := range pids {
				pids2[i] = env.Integer{int64(p)}
			}
			return env.NewBlock(*env.NewTSeries(pids2))
		},
	},
}
