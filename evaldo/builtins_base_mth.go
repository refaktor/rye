package evaldo

// builtins_base_mth.go
//
// mth: fast single-pass infix math evaluator with proper operator precedence.
//
// Design: classic two-stack (values + operators) Shunting-yard algorithm that
// evaluates *inline* instead of producing an intermediate RPN block.
// This avoids the two-pass overhead of calc { } (shunting-yard → RPN block → Eyr eval).
//
// Supported: integer and decimal literals, getwords (?var) for variable access,
// blocks (  ) for parenthesised sub-expressions, and the operators:
//   + - * / // %    (arithmetic, precedence 1 - 2)
//   < > = <= >=     (comparison, precedence 0 - lowest)
//
// The result is an env.Integer, env.Decimal, or env.Boolean.

import (
	"github.com/refaktor/rye/env"
)

// mthApplyOp pops the top operator from ops and the top two values from vals,
// applies the operation, and pushes the result. Returns false on error.
func mthApplyOp(
	vals *[]env.Object,
	ops *[]int,
	idxAdd, idxSub, idxMul, idxDiv, idxIDiv, idxMod int,
	idxLt, idxGt, idxEq, idxLte, idxGte int,
) bool {
	if len(*vals) < 2 || len(*ops) < 1 {
		return false
	}
	right := (*vals)[len(*vals)-1]
	left := (*vals)[len(*vals)-2]
	*vals = (*vals)[:len(*vals)-1]
	op := (*ops)[len(*ops)-1]
	*ops = (*ops)[:len(*ops)-1]

	switch op {
	case idxAdd:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				(*vals)[len(*vals)-1] = *env.NewInteger(l.Value + r.Value)
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(float64(l.Value) + r.Value)
			default:
				return false
			}
		case env.Decimal:
			switch r := right.(type) {
			case env.Integer:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value + float64(r.Value))
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value + r.Value)
			default:
				return false
			}
		default:
			return false
		}

	case idxSub:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				(*vals)[len(*vals)-1] = *env.NewInteger(l.Value - r.Value)
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(float64(l.Value) - r.Value)
			default:
				return false
			}
		case env.Decimal:
			switch r := right.(type) {
			case env.Integer:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value - float64(r.Value))
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value - r.Value)
			default:
				return false
			}
		default:
			return false
		}

	case idxMul:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				(*vals)[len(*vals)-1] = *env.NewInteger(l.Value * r.Value)
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(float64(l.Value) * r.Value)
			default:
				return false
			}
		case env.Decimal:
			switch r := right.(type) {
			case env.Integer:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value * float64(r.Value))
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value * r.Value)
			default:
				return false
			}
		default:
			return false
		}

	case idxDiv:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				if r.Value == 0 {
					return false
				}
				(*vals)[len(*vals)-1] = *env.NewDecimal(float64(l.Value) / float64(r.Value))
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(float64(l.Value) / r.Value)
			default:
				return false
			}
		case env.Decimal:
			switch r := right.(type) {
			case env.Integer:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value / float64(r.Value))
			case env.Decimal:
				(*vals)[len(*vals)-1] = *env.NewDecimal(l.Value / r.Value)
			default:
				return false
			}
		default:
			return false
		}

	case idxIDiv:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				if r.Value == 0 {
					return false
				}
				(*vals)[len(*vals)-1] = *env.NewInteger(l.Value / r.Value)
			default:
				return false
			}
		default:
			return false
		}

	case idxMod:
		switch l := left.(type) {
		case env.Integer:
			switch r := right.(type) {
			case env.Integer:
				if r.Value == 0 {
					return false
				}
				(*vals)[len(*vals)-1] = *env.NewInteger(l.Value % r.Value)
			default:
				return false
			}
		default:
			return false
		}

	case idxLt:
		lf, rf, ok := mthToFloats(left, right)
		if !ok {
			return false
		}
		(*vals)[len(*vals)-1] = *env.NewBoolean(lf < rf)

	case idxGt:
		lf, rf, ok := mthToFloats(left, right)
		if !ok {
			return false
		}
		(*vals)[len(*vals)-1] = *env.NewBoolean(lf > rf)

	case idxEq:
		lf, rf, ok := mthToFloats(left, right)
		if !ok {
			// Fall back to generic equality
			(*vals)[len(*vals)-1] = *env.NewBoolean(left.Equal(right))
			return true
		}
		(*vals)[len(*vals)-1] = *env.NewBoolean(lf == rf)

	case idxLte:
		lf, rf, ok := mthToFloats(left, right)
		if !ok {
			return false
		}
		(*vals)[len(*vals)-1] = *env.NewBoolean(lf <= rf)

	case idxGte:
		lf, rf, ok := mthToFloats(left, right)
		if !ok {
			return false
		}
		(*vals)[len(*vals)-1] = *env.NewBoolean(lf >= rf)

	default:
		return false
	}
	return true
}

// mthToFloats coerces two objects to float64 for comparison operations.
func mthToFloats(a, b env.Object) (float64, float64, bool) {
	var fa, fb float64
	switch v := a.(type) {
	case env.Integer:
		fa = float64(v.Value)
	case env.Decimal:
		fa = v.Value
	default:
		return 0, 0, false
	}
	switch v := b.(type) {
	case env.Integer:
		fb = float64(v.Value)
	case env.Decimal:
		fb = v.Value
	default:
		return 0, 0, false
	}
	return fa, fb, true
}

// Mth_EvalDirect is the fast single-pass math evaluator.
// It reads tokens from ps.Ser using the two-stack Shunting-yard algorithm
// and computes the result inline without creating an intermediate RPN block.
func Mth_EvalDirect(ps *env.ProgramState) env.Object {
	// Look up operator word indices once per call.
	// These are simple map lookups in ps.Idx - very cheap.
	idxAdd, _ := ps.Idx.GetIndex("_+")
	idxSub, _ := ps.Idx.GetIndex("_-")
	idxMul, _ := ps.Idx.GetIndex("_*")
	idxDiv, _ := ps.Idx.GetIndex("_/")
	idxIDiv, _ := ps.Idx.GetIndex("_//")
	idxMod, _ := ps.Idx.GetIndex("_%")
	idxLt, _ := ps.Idx.GetIndex("_<")
	idxGt, _ := ps.Idx.GetIndex("_>")
	idxEq, _ := ps.Idx.GetIndex("_=")
	idxLte, _ := ps.Idx.GetIndex("_<=")
	idxGte, _ := ps.Idx.GetIndex("_>=")

	// Precedence table. Higher number = tighter binding.
	// Comparison operators bind the most loosely so that
	// "5 + 5 < 10" evaluates as "(5 + 5) < 10".
	prec := map[int]int{
		idxLt: 0, idxGt: 0, idxEq: 0, idxLte: 0, idxGte: 0,
		idxAdd: 1, idxSub: 1,
		idxMul: 2, idxDiv: 2, idxIDiv: 2, idxMod: 2,
	}

	// Small inline stacks - avoid interface{} allocation for the common case.
	// Pre-allocated with reasonable capacity to avoid re-growing in typical expressions.
	vals := make([]env.Object, 0, 8)
	ops := make([]int, 0, 4)

	applyTop := func() bool {
		return mthApplyOp(&vals, &ops, idxAdd, idxSub, idxMul, idxDiv, idxIDiv, idxMod, idxLt, idxGt, idxEq, idxLte, idxGte)
	}

	for ps.Ser.Pos() < ps.Ser.Len() {
		object := ps.Ser.Pop()
		switch obj := object.(type) {

		case env.Integer:
			vals = append(vals, obj)

		case env.Decimal:
			vals = append(vals, obj)

		case env.Getword:
			// Variable access: ?varname
			val, found := ps.Ctx.Get(obj.Index)
			if !found {
				ps.ErrorFlag = true
				ps.Res = env.NewError("mth: variable not found: " + ps.Idx.GetWord(obj.Index))
				return ps.Res
			}
			vals = append(vals, val)

		case env.Block:
			// Parenthesised sub-expression: ( ... )
			ser1 := ps.Ser
			ps.Ser = obj.Series
			sub := Mth_EvalDirect(ps)
			ps.Ser = ser1
			if ps.ErrorFlag {
				return sub
			}
			vals = append(vals, sub)

		case env.Opword:
			p, known := prec[obj.Index]
			if !known {
				ps.ErrorFlag = true
				ps.Res = env.NewError("mth: unsupported operator: " + ps.Idx.GetWord(obj.Index))
				return ps.Res
			}
			// While the top of the operator stack has equal or higher precedence,
			// apply it before pushing the current operator.
			for len(ops) > 0 {
				topPrec := prec[ops[len(ops)-1]]
				if topPrec >= p {
					if !applyTop() {
						ps.ErrorFlag = true
						ps.Res = env.NewError("mth: type error in arithmetic expression")
						return ps.Res
					}
				} else {
					break
				}
			}
			ops = append(ops, obj.Index)

		default:
			ps.ErrorFlag = true
			ps.Res = env.NewError("mth: unexpected token type in block")
			return ps.Res
		}
	}

	// Drain remaining operators from the operator stack.
	for len(ops) > 0 {
		if !applyTop() {
			ps.ErrorFlag = true
			ps.Res = env.NewError("mth: type error in arithmetic expression")
			return ps.Res
		}
	}

	if len(vals) == 0 {
		return *env.NewVoid()
	}
	return vals[0]
}

var builtins_mth = map[string]*env.Builtin{
	// Tests:
	// equal { mth { 1 + 2 } } 3
	// equal { mth { 2 + 3 * 4 } } 14
	// equal { mth { ( 2 + 3 ) * 4 } } 20
	// equal { mth { 10 - 5 - 2 } } 3
	// equal { mth { 10 - 5 - 2 } } 3
	// equal { mth { 8 // 3 } } 2
	// equal { mth { 8 % 3 } } 2
	// equal { mth { 5 + 5 < 10 } } false
	// equal { mth { 5 + 5 <= 10 } } true
	// equal { mth { 5 + 4 < 10 } } true
	// equal { mth { 1 + 2 * 3 } } 7
	// equal { mth { ( 1 + 2 ) * 3 } } 9
	// equal { a: 5 mth { ?a * 2 } } 10
	// equal { a: 3 b: 4 mth { ?a * ?a + ?b * ?b } } 25
	// Args:
	// * block: Block containing an infix math expression
	// Returns:
	// * result of evaluating the expression with proper operator precedence
	"mth": {
		Argsn: 1,
		Doc:   "Fast infix math evaluator with proper operator precedence (+,-,*,/,//,%,<,>,=,<=,>=)",
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
