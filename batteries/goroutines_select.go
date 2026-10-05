//go:build !b_tiny

package batteries

import (
	"reflect"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// parseSelectCases evaluates the specification in the caller's context, but
// never leaves the caller pointing into the specification on failure or return.
// functions selects function handlers (select\fn) instead of block handlers (select).
// Cases and handlers have matching indexes; false means ps holds an error or return.
func parseSelectCases(ps *env.ProgramState, block env.Block, functions bool, name string) ([]reflect.SelectCase, []env.Object, bool) {
	ser := ps.Ser
	ps.Ser = block.Series
	ps.Ser.Reset()
	defer func() { ps.Ser = ser }()

	var cases []reflect.SelectCase
	var handlers []env.Object
	hasDefault := false
	// Record a syntax/type failure; evaluation errors keep their original result.
	fail := func(message string) ([]reflect.SelectCase, []env.Object, bool) {
		ps.FailureFlag = true
		ps.Res = evaldo.MakeBuiltinError(ps, message, name)
		return nil, nil, false
	}
	stopped := func() bool { return ps.ErrorFlag || ps.FailureFlag || ps.ReturnFlag }
	for ps.Ser.Pos() < ps.Ser.Len() {
		evaldo.EvalExpression_CollectArg(ps, false, false)
		if stopped() {
			return nil, nil, false
		}

		// A bare function is the default in select\fn; _ marks it in select.
		var selection reflect.SelectCase
		var handler env.Object
		switch selector := ps.Res.(type) {
		case env.Function:
			if !functions {
				return fail("expected a channel or _ before a block")
			}
			if selector.Argsn != 0 {
				return fail("default function must take 0 arguments")
			}
			selection.Dir = reflect.SelectDefault
			handler = selector
		case env.Void:
			if functions {
				return fail("expected a channel or a default function")
			}
			selection.Dir = reflect.SelectDefault
		case env.Native:
			ch, ok := selector.Value.(chan *env.Object)
			if !ok {
				return fail("expected a channel")
			}
			selection = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		default:
			return fail("expected a channel or default case")
		}

		if selection.Dir == reflect.SelectDefault {
			if hasDefault {
				return fail("select can only have one default case")
			}
			hasDefault = true
		}
		// Only a bare default function already includes its handler.
		if handler == nil {
			if ps.Ser.Pos() >= ps.Ser.Len() {
				return fail("missing handler after channel or default selector")
			}
			evaldo.EvalExpression_CollectArg(ps, false, false)
			if stopped() {
				return nil, nil, false
			}
			handler = ps.Res
		}
		if functions {
			fn, ok := handler.(env.Function)
			if !ok {
				return fail("case handler must be a function")
			}
			if fn.Argsn < 0 || fn.Argsn > 1 {
				return fail("channel handler must take 0 or 1 argument")
			}
		} else if _, ok := handler.(env.Block); !ok {
			return fail("case handler must be a block")
		}
		// reflect.Select panics above this limit.
		if len(cases) == 65536 {
			return fail("too many select cases")
		}
		cases = append(cases, selection)
		handlers = append(handlers, handler)
	}
	return cases, handlers, true
}

// runChannelSelect waits for a ready case (or falls back to default), then
// returns its handler's result. Errors and control-flow flags stay in ps.
func runChannelSelect(ps *env.ProgramState, specification env.Object, functions bool) env.Object {
	name := "select"
	if functions {
		name = "select\\fn"
	}
	block, ok := specification.(env.Block)
	if !ok {
		ps.FailureFlag = true
		return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType}, name)
	}
	cases, handlers, ok := parseSelectCases(ps, block, functions, name)
	if !ok {
		return ps.Res
	}

	// An explicitly empty specification intentionally blocks forever, like Go's
	// select {}. Invalid specifications have already returned an error above.
	chosen, received, open := reflect.Select(cases)
	var payload env.Object
	if cases[chosen].Dir == reflect.SelectRecv {
		// Buffered messages arrive normally; a drained closed channel fails.
		if !open {
			ps.FailureFlag = true
			return evaldo.MakeBuiltinError(ps, "channel closed", name)
		}
		message, valid := received.Interface().(*env.Object)
		if !valid || message == nil || *message == nil {
			ps.FailureFlag = true
			return evaldo.MakeBuiltinError(ps, "invalid channel message", name)
		}
		payload = *message
	}

	// Select is synchronous: copying ProgramState here loses control-flow flags
	// and unnecessarily copies shared runtime state.
	if functions {
		// Explicit arguments prevent handlers from consuming the caller's code.
		fn := handlers[chosen].(env.Function)
		if fn.Argsn == 0 {
			evaldo.CallFunctionWithArgs(fn, ps, nil)
		} else {
			evaldo.CallFunctionWithArgs(fn, ps, nil, payload)
		}
	} else {
		// Inject received values into blocks; defaults run without injection.
		ser := ps.Ser
		defer func() { ps.Ser = ser }()
		ps.Ser = handlers[chosen].(env.Block).Series
		ps.Ser.Reset()
		ps.Res = env.Void{}
		evaldo.EvalBlockInj(ps, payload, payload != nil)
	}
	return ps.Res
}
