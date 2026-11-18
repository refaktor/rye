package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

// ryeTask publishes its immutable completion record by closing done chan. Multiple
// awaiters can observe it. Mutable result values themselves are NOT copied.
// Tasks currently share closures and input values just like go/go-with. They
// provide execution-state isolation and completion, not heap isolation

type ryeTask struct {
	done   chan struct{}
	result env.Object
	failed bool
}

func spawnTask(ps *env.ProgramState, value env.Object, input env.Object, withInput bool) env.Object {
	name, arity := "spawn", 0
	if withInput {
		name, arity = "spawn-with", 1
	}
	fn, ok := value.(env.Function)
	if !ok {
		position := 1
		if withInput {
			position = 2
		}
		return MakeArgError(ps, position, []env.Type{env.FunctionType}, name)
	}
	if fn.Argsn != arity {
		return MakeBuiltinError(ps, fmt.Sprintf("function with %d arguments required", arity), name)
	}
	if !validateFunctionSpec(fn, ps) {
		return ps.Res
	}
	worker := ps.ForkForGoroutine()
	task := &ryeTask{done: make(chan struct{})}
	// Construct the handle before publication (symbol interning belongs here).
	handle := *env.NewNative(ps.Idx, task, "Rye-task")
	go func() {
		defer close(task.done)
		defer func() {
			if recovered := recover(); recovered != nil {
				task.failed = true
				task.result = env.NewError(fmt.Sprintf("task panic: %v", recovered))
			}
		}()
		if withInput {
			CallFunctionWithArgs(fn, worker, nil, input)
		} else {
			CallFunctionWithArgs(fn, worker, nil)
		}
		task.result = worker.Res
		if task.result == nil {
			task.result = env.Void{}
		}
		task.failed = worker.FailureFlag || worker.ErrorFlag
	}()
	return handle
}

var builtins_tasks = map[string]*env.Builtin{
	"spawn": {
		Argsn: 1, Effect: env.EffIO,
		Doc: "Starts a zero-argument function and returns a task. Contexts remain shared; await observes completion.",
		Fn: func(ps *env.ProgramState, a, b, c, d, e env.Object) env.Object {
			return spawnTask(ps, a, nil, false)
		},
	},
	"spawn-with": {
		Argsn: 2, Effect: env.EffIO,
		Doc: "Starts a one-argument function with a value and returns a task. Inputs and captured contexts are not copied.",
		Fn: func(ps *env.ProgramState, a, b, c, d, e env.Object) env.Object {
			return spawnTask(ps, b, a, true)
		},
	},
	"await": {
		Argsn: 1, Effect: env.EffIO,
		Doc: "Waits for a task and returns its result, or raises its failure. Repeated awaits are supported; results are not copied.",
		Fn: func(ps *env.ProgramState, a, b, c, d, e env.Object) env.Object {
			native, ok := a.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "await")
			}
			task, ok := native.Value.(*ryeTask)
			if !ok || task == nil {
				return MakeBuiltinError(ps, "expected a task", "await")
			}
			<-task.done
			if task.failed {
				ps.FailureFlag = true
			}
			return task.result
		},
	},
}
