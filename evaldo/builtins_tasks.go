package evaldo

import (
	"fmt"
	"sync"

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
		// position of function argument in builtin depends if it's withInput flag
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
	// Fork programstate specifically for another goroutine
	worker := ps.ForkForGoroutine()
	task := &ryeTask{done: make(chan struct{})}
	// Construct the handle before publication (symbol interning belongs here).
	handle := *env.NewNative(ps.Idx, task, "Rye-task")
	var complete func(env.Object, bool)
	if ps.TaskScope != nil {
		// complete is a function that is called on completion of the task
		// it also decrements the waitgroup that registerTask increments
		complete = ps.TaskScope.RegisterTask()
	}
	// goroutine is started
	go func() {
		defer func() {
			// closes the done channel
			close(task.done)
			// if complete is there call it with result ans failed status
			if complete != nil {
				complete(task.result, task.failed)
			}
		}()
		defer func() {
			if recovered := recover(); recovered != nil {
				task.failed = true
				task.result = env.NewError(fmt.Sprintf("task panic: %v", recovered))
			}
		}()
		// here we call the given function with 1 arg or no args
		// TODO -- we could provide a block of args als in future
		if withInput {
			// worker is ps (program state) for this gorotutine
			CallFunctionWithArgs(fn, worker, nil, input)
		} else {
			CallFunctionWithArgs(fn, worker, nil)
		}
		// We set result and failed status, before defer will be called
		task.result = worker.Res
		if task.result == nil {
			task.result = env.Void{}
		}
		task.failed = worker.FailureFlag || worker.ErrorFlag
	}()
	return handle
}

// taskScope permits registration by descendants while their ancestors remain
// counted. All registrations happen before launch, so Wait cannot finish early.
type taskScope struct {
	wg       sync.WaitGroup
	mu       sync.Mutex
	failures []env.Object
}

func (s *taskScope) RegisterTask() func(env.Object, bool) {
	// increment wait group
	s.wg.Add(1)
	s.mu.Lock()
	index := len(s.failures)
	s.failures = append(s.failures, nil)
	s.mu.Unlock()
	return func(result env.Object, failed bool) {
		if failed {
			s.mu.Lock()
			s.failures[index] = result
			s.mu.Unlock()
		}
		s.wg.Done()
	}
}

func withTasks(ps *env.ProgramState, value env.Object) env.Object {
	block, ok := value.(env.Block)
	if !ok {
		return MakeArgError(ps, 1, []env.Type{env.BlockType}, "with-tasks")
	}
	scope := &taskScope{}
	// we set Previous TaskScope and Series (curr block of code) to temp vars
	previous, series := ps.TaskScope, ps.Ser
	ps.TaskScope = scope
	// On defer we will set those back from temp vars
	defer func() { ps.TaskScope, ps.Ser = previous, series }()
	// Join even on a Go panic, but do not swallow that panic.
	defer scope.wg.Wait()
	// set current code to given block of code
	ps.Ser = block.Series
	ps.Ser.Reset()
	ps.Res = env.Void{}
	EvalBlockInj(ps, nil, false)
	scope.wg.Wait()
	// A body failure wins. Otherwise the first failed registration wins,
	// including over a pending return. Awaiting a failure does not remove it.
	if !ps.FailureFlag && !ps.ErrorFlag {
		for _, failure := range scope.failures {
			if failure != nil {
				// we set current failure to first failure in TaskScope
				ps.Res = failure
				ps.FailureFlag = true
				ps.ReturnFlag = false
				break
			}
		}
	}
	return ps.Res
}

// awaitAll validates before blocking, then joins every listed task even when
// an earlier task failed. Only listed handles are joined, not their descendants.
func awaitAll(ps *env.ProgramState, value env.Object) env.Object {
	block, ok := value.(env.Block)
	if !ok {
		return MakeArgError(ps, 1, []env.Type{env.BlockType}, "await-all")
	}
	// collect all tasks from the block, check for type
	tasks := make([]*ryeTask, len(block.Series.S))
	for i, item := range block.Series.S {
		native, ok := item.(env.Native)
		if ok {
			tasks[i], _ = native.Value.(*ryeTask)
		}
		if tasks[i] == nil || tasks[i].done == nil {
			return MakeBuiltinError(ps, fmt.Sprintf("expected a task at position %d", i+1), "await-all")
		}
	}
	// prepare block for results
	results := make([]env.Object, len(tasks))
	var failure env.Object
	// for each task wait for done signal, pickup the task result, set failure flags if neeedd
	for i, task := range tasks {
		<-task.done
		results[i] = task.result
		if task.failed && failure == nil {
			failure = task.result
		}
	}
	if failure != nil {
		ps.FailureFlag = true
		return failure
	}
	return *env.NewBlock(*env.NewTSeries(results))
}

var builtins_tasks = map[string]*env.Builtin{
	"await-all": {
		Argsn: 1, Effect: env.EffIO,
		Doc: "Joins a block of tasks and returns results in input order. Joins all before raising the first failure in input order. Validates all handles before waiting.",
		Fn: func(ps *env.ProgramState, a, b, c, d, e env.Object) env.Object {
			return awaitAll(ps, a)
		},
	},
	"with-tasks": {
		Argsn: 1, Effect: env.EffIO,
		Doc: "Evaluates a block, joins its tasks and descendants, and propagates failure. Does not cancel blocked tasks.",
		Fn:  func(ps *env.ProgramState, a, b, c, d, e env.Object) env.Object { return withTasks(ps, a) },
	},
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
