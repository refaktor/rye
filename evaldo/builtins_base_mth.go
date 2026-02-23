package evaldo

// builtins_base_mth.go
//
// mth — fast single-pass infix math evaluator with proper operator precedence.
//
// Design: classic two-stack (values + operators) Shunting-yard algorithm that
// evaluates *inline* instead of producing an intermediate RPN block.
//
// Supported: integer and decimal literals, words (var) for variable access,
// blocks { } for parenthesised sub-expressions, and the operators:
//   + - * / // %    (arithmetic, precedence 2–3)
//   < > = <= >= !=  (comparison, precedence 1 — lowest)
//
// The result is an env.Integer, env.Decimal, or env.Boolean.
//
// Key design decision: operators are identified by their raw string ("+" etc.)
// via ps.Idx.GetWord — a simple slice index, O(1), no allocation.  This means
// no builtin dictionary lookup, no global state, no lazy init.
// Arithmetic is done directly in Go — no builtin function call overhead.

import (
	"github.com/refaktor/rye/env"
)

// Operator codes — small constants used on the operator stack.
const (
	mthOpNone = iota
	mthOpAdd  // +
	mthOpSub  // -
	mthOpMul  // *
	mthOpDiv  // /   → decimal
	mthOpIDiv // //  → integer
	mthOpMod  // %
	mthOpLt   // <
	mthOpGt   // >
	mthOpEq   // =
	mthOpLte  // <=
	mthOpGte  // >=
	mthOpNeq  // !=
)

// mthOpPrec returns the precedence of an operator code.
// 0 = unknown, 1 = comparison, 2 = add/sub, 3 = mul/div/mod.
func mthOpPrec(op int) int8 {
	switch op {
	case mthOpLt, mthOpGt, mthOpEq, mthOpLte, mthOpGte, mthOpNeq:
		return 1
	case mthOpAdd, mthOpSub:
		return 2
	case mthOpMul, mthOpDiv, mthOpIDiv, mthOpMod:
		return 3
	}
	return 0
}

// mthOpOf maps an opword string to an operator code.
// In Rye, arithmetic opwords are stored with a leading underscore (_+, _-, etc.).
// ps.Idx.GetWord is a plain slice index — O(1), no allocation.
func mthOpOf(name string) int {
	switch name {
	case "_+":
		return mthOpAdd
	case "_-":
		return mthOpSub
	case "_*":
		return mthOpMul
	case "_/":
		return mthOpDiv
	case "_//":
		return mthOpIDiv
	case "_%":
		return mthOpMod
	case "_<":
		return mthOpLt
	case "_>":
		return mthOpGt
	case "_=":
		return mthOpEq
	case "_<=":
		return mthOpLte
	case "_>=":
		return mthOpGte
	case "_!=":
		return mthOpNeq
	}
	return mthOpNone
}

// ---------------------------------------------------------------------------
// Stack-allocated state — zero heap for typical expressions (≤8 operands).
// ---------------------------------------------------------------------------

type mthState struct {
	vals [8]env.Object
	ops  [4]int8 // operator codes (fit in int8)
	vLen int
	oLen int
}

func (s *mthState) pushVal(v env.Object) {
	if s.vLen < len(s.vals) {
		s.vals[s.vLen] = v
		s.vLen++
	}
}

func (s *mthState) topVal() env.Object     { return s.vals[s.vLen-1] }
func (s *mthState) popVal() env.Object     { s.vLen--; return s.vals[s.vLen] }
func (s *mthState) setTopVal(v env.Object) { s.vals[s.vLen-1] = v }

func (s *mthState) pushOp(op int8) {
	if s.oLen < len(s.ops) {
		s.ops[s.oLen] = op
		s.oLen++
	}
}
func (s *mthState) topOp() int8 { return s.ops[s.oLen-1] }
func (s *mthState) popOp() int8 { s.oLen--; return s.ops[s.oLen] }

// ---------------------------------------------------------------------------
// Arithmetic — pure Go, no builtin lookup.
// ---------------------------------------------------------------------------

func mthToFloat(o env.Object) (float64, bool) {
	switch v := o.(type) {
	case env.Integer:
		return float64(v.Value), true
	case env.Decimal:
		return v.Value, true
	}
	return 0, false
}

// mthApply pops the top operator and two values, computes the result in Go,
// and writes it back as the new stack top.  Returns false on type error.
func mthApply(s *mthState) bool {
	if s.vLen < 2 || s.oLen < 1 {
		return false
	}
	right := s.popVal()
	op := s.popOp()
	left := s.topVal() // result written here in-place

	switch op {
	case mthOpAdd:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				s.setTopVal(*env.NewInteger(l.Value + r.Value))
			case env.Decimal:
				s.setTopVal(*env.NewDecimal(float64(l.Value) + r.Value))
			default:
				return false
			}
		case env.Decimal:
			switch r := right.(type) {
			case env.Integer:
				s.setTopVal(*env.NewDecimal(l.Value + float64(r.Value)))
			case env.Decimal:
				s.setTopVal(*env.NewDecimal(l.Value + r.Value))
			default:
				return false
			}
		default:
			return false
		}

	case mthOpSub:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				s.setTopVal(*env.NewInteger(l.Value - r.Value))
			case env.Decimal:
				s.setTopVal(*env.NewDecimal(float64(l.Value) - r.Value))
			default:
				return false
			}
		case env.Decimal:
			switch r := right.(type) {
			case env.Integer:
				s.setTopVal(*env.NewDecimal(l.Value - float64(r.Value)))
			case env.Decimal:
				s.setTopVal(*env.NewDecimal(l.Value - r.Value))
			default:
				return false
			}
		default:
			return false
		}

	case mthOpMul:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				s.setTopVal(*env.NewInteger(l.Value * r.Value))
			case env.Decimal:
				s.setTopVal(*env.NewDecimal(float64(l.Value) * r.Value))
			default:
				return false
			}
		case env.Decimal:
			switch r := right.(type) {
			case env.Integer:
				s.setTopVal(*env.NewDecimal(l.Value * float64(r.Value)))
			case env.Decimal:
				s.setTopVal(*env.NewDecimal(l.Value * r.Value))
			default:
				return false
			}
		default:
			return false
		}

	case mthOpDiv:
		lf, ok1 := mthToFloat(left)
		rf, ok2 := mthToFloat(right)
		if !ok1 || !ok2 {
			return false
		}
		s.setTopVal(*env.NewDecimal(lf / rf))

	case mthOpIDiv:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				if r.Value == 0 {
					return false
				}
				s.setTopVal(*env.NewInteger(l.Value / r.Value))
			default:
				return false
			}
		default:
			return false
		}

	case mthOpMod:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				if r.Value == 0 {
					return false
				}
				s.setTopVal(*env.NewInteger(l.Value % r.Value))
			default:
				return false
			}
		default:
			return false
		}

	case mthOpLt:
		lf, ok1 := mthToFloat(left)
		rf, ok2 := mthToFloat(right)
		if !ok1 || !ok2 {
			return false
		}
		s.setTopVal(*env.NewBoolean(lf < rf))

	case mthOpGt:
		lf, ok1 := mthToFloat(left)
		rf, ok2 := mthToFloat(right)
		if !ok1 || !ok2 {
			return false
		}
		s.setTopVal(*env.NewBoolean(lf > rf))

	case mthOpEq:
		lf, ok1 := mthToFloat(left)
		rf, ok2 := mthToFloat(right)
		if ok1 && ok2 {
			s.setTopVal(*env.NewBoolean(lf == rf))
		} else {
			s.setTopVal(*env.NewBoolean(left.Equal(right)))
		}

	case mthOpNeq:
		lf, ok1 := mthToFloat(left)
		rf, ok2 := mthToFloat(right)
		if ok1 && ok2 {
			s.setTopVal(*env.NewBoolean(lf != rf))
		} else {
			s.setTopVal(*env.NewBoolean(!left.Equal(right)))
		}

	case mthOpLte:
		lf, ok1 := mthToFloat(left)
		rf, ok2 := mthToFloat(right)
		if !ok1 || !ok2 {
			return false
		}
		s.setTopVal(*env.NewBoolean(lf <= rf))

	case mthOpGte:
		lf, ok1 := mthToFloat(left)
		rf, ok2 := mthToFloat(right)
		if !ok1 || !ok2 {
			return false
		}
		s.setTopVal(*env.NewBoolean(lf >= rf))

	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Core evaluator
// ---------------------------------------------------------------------------

// Mth_EvalDirect is the fast single-pass math evaluator.
// It reads tokens from ps.Ser using the Shunting-yard algorithm and computes
// the result inline.  No builtin lookup, no global state, no init.
func Mth_EvalDirect(ps *env.ProgramState) env.Object {
	var s mthState // stack-allocated, zero heap for typical expressions

	for ps.Ser.Pos() < ps.Ser.Len() {
		object := ps.Ser.Pop()
		switch obj := object.(type) {

		case env.Integer:
			s.pushVal(obj)

		case env.Decimal:
			s.pushVal(obj)

		case env.Word:
			// Variable access: varname
			val, found := ps.Ctx.Get(obj.Index)
			if !found {
				ps.ErrorFlag = true
				ps.Res = env.NewError("mth: variable not found: " + ps.Idx.GetWord(obj.Index))
				return ps.Res
			}
			s.pushVal(val)

		case env.Block:
			// Parenthesised sub-expression: { ... }
			ser1 := ps.Ser
			ps.Ser = obj.Series
			sub := Mth_EvalDirect(ps)
			ps.Ser = ser1
			if ps.ErrorFlag {
				return sub
			}
			s.pushVal(sub)

		case env.Opword:
			// Identify operator from its string — ps.Idx.GetWord is a slice index, O(1).
			op := mthOpOf(ps.Idx.GetWord(obj.Index))
			if op == mthOpNone {
				ps.ErrorFlag = true
				ps.Res = env.NewError("mth: unsupported operator: " + ps.Idx.GetWord(obj.Index))
				return ps.Res
			}
			p := mthOpPrec(op)
			// Shunting-yard: pop operators with equal or higher precedence first.
			for s.oLen > 0 && mthOpPrec(int(s.topOp())) >= p {
				if !mthApply(&s) {
					ps.ErrorFlag = true
					ps.Res = env.NewError("mth: type error in arithmetic expression")
					return ps.Res
				}
			}
			s.pushOp(int8(op))

		default:
			ps.ErrorFlag = true
			ps.Res = env.NewError("mth: unexpected token type in block")
			return ps.Res
		}
	}

	// Drain remaining operators.
	for s.oLen > 0 {
		if !mthApply(&s) {
			ps.ErrorFlag = true
			ps.Res = env.NewError("mth: type error in arithmetic expression")
			return ps.Res
		}
	}

	if s.vLen == 0 {
		return *env.NewVoid()
	}
	return s.vals[0]
}

// ---------------------------------------------------------------------------
// Builtin registration
// ---------------------------------------------------------------------------

var builtins_mth = map[string]*env.Builtin{
	// Tests:
	// equal { mth { 1 + 2 } } 3
	// equal { mth { 2 + 3 * 4 } } 14
	// equal { mth { 10 - 5 - 2 } } 3
	// equal { mth { 8 // 3 } } 2
	// equal { mth { 8 % 3 } } 2
	// equal { mth { 5 + 5 < 10 } } false
	// equal { mth { 5 + 5 <= 10 } } true
	// equal { mth { 5 + 4 < 10 } } true
	// equal { mth { 1 + 2 * 3 } } 7
	// equal { a: 5 mth { a * 2 } } 10
	// equal { a: 3 b: 4 mth { a * a + b * b } } 25
	// Args:
	// * block: Block containing an infix math expression
	// Returns:
	// * result of evaluating the expression with proper operator precedence
	"mth": {
		Argsn: 1,
		Doc:   "Fast infix math evaluator with proper operator precedence (+,-,*,/,//,%,<,>,=,<=,>=,!=). Single-pass Shunting-yard, direct Go arithmetic, no builtin lookup.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch blk := arg0.(type) {
			case env.Block:
				ser1 := ps.Ser
				ps.Ser = blk.Series
				result := Mth_EvalDirect(ps)
				ps.Ser = ser1
				return result
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "mth")
			}
		},
	},
}
