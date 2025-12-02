// evaldo_rye0_no_opt.go - A simplified interpreter for the Rye language without word replacement optimization
package evaldo

/* // Rye0_findWordValue_NoOpt returns the value associated with a word in the current context.
// This version does NOT replace the word with the builtin in the series.
func Rye0_findWordValue_NoOpt(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	// Handle CPath type separately
	if cpath, ok := word1.(env.CPath); ok {
		return Rye0_findCPathValue(ps, cpath)
	}

	// Extract the word index from different word types
	var index int
	switch w := word1.(type) {
	case env.Word:
		index = w.Index
	case env.Opword:
		index = w.Index
	case env.Pipeword:
		index = w.Index
	default:
		return false, nil, nil
	}

	// First try to get the value from the current context
	object, found := ps.Ctx.Get(index)
	if found {
		return found, object, nil
	}

	// If not found in the current context and there's no parent, return not found
	if ps.Ctx.Parent == nil {
		return false, nil, nil
	}

	// Try to get the value directly from the parent context
	object, found = ps.Ctx.Parent.Get(index)
	if found {
		// No optimization here - we don't replace the word with the builtin
		return found, object, ps.Ctx.Parent
	}

	// If not found in the parent context, use the regular Get2 method to search up the context chain
	object, found, foundCtx := ps.Ctx.Get2(index)
	return found, object, foundCtx
}

// Rye0_EvalWord_NoOpt evaluates a word in the current context without the word replacement optimization.
func Rye0_EvalWord_NoOpt(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) *env.ProgramState {
	var firstVal env.Object
	found, object, session := Rye0_findWordValue_NoOpt(ps, word)
	pos := ps.Ser.GetPos()

	if !found {
		// Determine the kind for generic word lookup
		kind := 0
		if leftVal != nil {
			kind = leftVal.GetKind()
		}

		// Evaluate next expression if needed
		if (leftVal == nil && !pipeSecond) || pipeSecond {
			if !ps.Ser.AtLast() {
				Rye0_EvalExpression_DispatchType_NoOpt(ps)
				if ps.ReturnFlag {
					return ps
				}

				if pipeSecond {
					firstVal = ps.Res
					kind = firstVal.GetKind()
				} else {
					leftVal = ps.Res
					kind = leftVal.GetKind()
				}
			}
		}

		// Try to find a generic word
		if rword, ok := word.(env.Word); ok && leftVal != nil && ps.Ctx.Kind.Index != -1 {
			object, found = ps.Gen.Get(kind, rword.Index)
		}
	}

	if found {
		return Rye0_EvalObject(ps, object, leftVal, toLeft, session, pipeSecond, firstVal)
	} else {
		if !ps.FailureFlag {
			ps.Ser.SetPos(pos)
			setError(ps, "Word not found: "+word.Print(*ps.Idx))
		} else {
			ps.ErrorFlag = true
		}
		return ps
	}
}

// Rye0_EvalExpression_DispatchType_NoOpt evaluates a concrete expression without the word replacement optimization.
func Rye0_EvalExpression_DispatchType_NoOpt(ps *env.ProgramState) *env.ProgramState {
	object := ps.Ser.Pop()

	if object == nil {
		ps.ErrorFlag = true
		ps.Res = errMissingValue
		return ps
	}

	switch object.Type() {
	// Literal values evaluate to themselves
	case env.IntegerType, env.DecimalType, env.StringType, env.VoidType, env.UriType, env.EmailType:
		if !ps.SkipFlag {
			ps.Res = object
		}

	// Block handling
	case env.BlockType:
		if !ps.SkipFlag {
			return Rye0_EvaluateBlock_NoOpt(ps, object.(env.Block))
		}

	// Word types
	case env.TagwordType:
		// Create a Word directly without allocation
		ps.Res = env.Word{Index: object.(env.Tagword).Index}
		return ps
	case env.WordType:
		return Rye0_EvalWord_NoOpt(ps, object.(env.Word), nil, false, false)
	case env.CPathType:
		return Rye0_EvalWord_NoOpt(ps, object, nil, false, false)
	case env.BuiltinType:
		return Rye0_CallBuiltin(object.(env.Builtin), ps, nil, false, false, nil)
	case env.GenwordType:
		return Rye0_EvalGenword(ps, object.(env.Genword), nil, false)
	case env.SetwordType:
		return Rye0_EvalSetword(ps, object.(env.Setword))
	case env.ModwordType:
		return Rye0_EvalModword(ps, object.(env.Modword))
	case env.GetwordType:
		return Rye0_EvalGetword(ps, object.(env.Getword), nil, false)

	// Error handling
	case env.CommaType:
		setError(ps, "Expression guard inside expression")
	case env.ErrorType:
		setError(ps, "Error object encountered")

	// Unknown type
	default:
		setError(ps, "Unknown Rye value type: "+strconv.Itoa(int(object.Type())))
	}

	return ps
}

// Rye0_EvaluateBlock_NoOpt handles the evaluation of a block object without the word replacement optimization.
func Rye0_EvaluateBlock_NoOpt(ps *env.ProgramState, block env.Block) *env.ProgramState {
	// Save original series to restore later
	ser := ps.Ser

	switch block.Mode {
	case 1: // Eval blocks
		ps.Ser = block.Series
		// Pre-allocate the result slice with estimated capacity to avoid reallocations
		estimatedSize := ps.Ser.Len() - ps.Ser.Pos()
		res := make([]env.Object, 0, estimatedSize)

		for ps.Ser.Pos() < ps.Ser.Len() {
			Rye0_EvalExpression_CollectArg_NoOpt(ps, false)
			if Rye0_checkErrorReturnFlag(ps) {
				ps.Ser = ser // Restore original series
				return ps
			}
			res = append(res, ps.Res)
		}
		ps.Ser = ser // Restore original series

		// Create series and block in one step to reduce allocations
		series := env.NewTSeries(res)
		ps.Res = *env.NewBlock(*series)
	case 2:
		ps.Ser = block.Series
		EvalBlock_NoOpt(ps)
		ps.Ser = ser // Restore original series
	default:
		ps.Res = block
	}
	return ps
}

// Rye0_EvalExpression_CollectArg_NoOpt evaluates an expression with optional limitations without the word replacement optimization.
func Rye0_EvalExpression_CollectArg_NoOpt(ps *env.ProgramState, limited bool) *env.ProgramState {
	Rye0_EvalExpression_DispatchType_NoOpt(ps)
	return ps
}

// EvalBlock_NoOpt evaluates a block without the word replacement optimization.
func EvalBlock_NoOpt(ps *env.ProgramState) *env.ProgramState {
	for ps.Ser.Pos() < ps.Ser.Len() {
		Rye0_EvalExpression_CollectArg_NoOpt(ps, false)
		if Rye0_checkFlagsAfterBlock(ps, 101) || Rye0_checkErrorReturnFlag(ps) {
			return ps
		}
	}
	return ps
}
*/
