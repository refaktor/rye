package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

func makeBunchFromObject(ps *env.ProgramState, arg0 env.Object, name string) env.Object {
	switch value := arg0.(type) {
	case env.Block:
		bunch, err := env.NewBunchFromSeriesChecked(value.Series, ps.Idx)
		if err != nil {
			return MakeBuiltinError(ps, err.Error(), name)
		}
		return bunch
	case env.List:
		values := make([]env.Object, len(value.Data))
		for i, v := range value.Data {
			values[i] = env.ToRyeValue(v)
		}
		bunch, err := env.NewBunchFromValuesChecked(values, ps.Idx)
		if err != nil {
			return MakeBuiltinError(ps, err.Error(), name)
		}
		return bunch
	case env.Bunch:
		return value.DeepCopy()
	case *env.Bunch:
		return value.DeepCopy()
	default:
		return MakeArgError2(ps, 1, []env.Type{env.BlockType, env.ListType, env.BunchType}, name, arg0)
	}
}

func modifyBunchWord(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, add bool, name string) env.Object {
	key, ok := env.BunchKey(arg0, ps.Idx)
	if !ok {
		return MakeArgError2(ps, 1, []env.Type{env.StringType, env.WordType, env.TagwordType, env.IntegerType, env.DecimalType, env.BooleanType}, name, arg0)
	}

	switch target := arg1.(type) {
	case env.Bunch:
		if add {
			target.Add(key)
		} else {
			target.Remove(key)
		}
		return target
	case *env.Bunch:
		if add {
			target.Add(key)
		} else {
			target.Remove(key)
		}
		return target
	case env.Word, env.Tagword:
		var index int
		switch word := arg1.(type) {
		case env.Word:
			index = word.Index
		case env.Tagword:
			index = word.Index
		}
		val, found, ctx := ps.Ctx.Get2(index)
		if !found {
			return MakeBuiltinError(ps, "Word not found.", name)
		}
		if !ctx.IsVariable(index) {
			ps.FailureFlag = true
			return env.NewError("cannot modify constant '" + ps.Idx.GetWord(index) + "', use 'var' or mod-word for a variable'")
		}
		switch bunch := val.(type) {
		case env.Bunch:
			if add {
				bunch.Add(key)
			} else {
				bunch.Remove(key)
			}
			if !ctx.Mod(index, bunch) {
				return MakeBuiltinError(ps, "Unable to modify bunch variable.", name)
			}
			return bunch
		case *env.Bunch:
			if add {
				bunch.Add(key)
			} else {
				bunch.Remove(key)
			}
			if !ctx.Mod(index, bunch) {
				return MakeBuiltinError(ps, "Unable to modify bunch variable.", name)
			}
			return bunch
		default:
			return MakeBuiltinError(ps, fmt.Sprintf("Value of word is not a Bunch, got %s.", NameOfRyeType(val.Type())), name)
		}
	default:
		return MakeArgError2(ps, 2, []env.Type{env.BunchType, env.WordType, env.TagwordType}, name, arg1)
	}
}

// ##### Bunches ##### "Working with bunch values."

var builtins_bunch = map[string]*env.Builtin{
	// Tests:
	// equal { bunch { "a" "b" "a" } |length? } 2
	// equal { bunch { "a" "b" } |type? } 'bunch
	// equal { bunch { "a" "b" } |contains "a" } true
	// equal { bunch { "a" "b" } |contains "c" } false
	// equal { bunch { a b } |contains "a" } true
	// equal { bunch { 1 2 2 } |length? } 2
	// error { bunch { "a" { "x" } } } ; blocks are not valid bunch elements
	// Args:
	// * values: Block of scalar values, or an existing List/Bunch
	// Returns:
	// * a new Bunch containing the unique values
	"bunch": {
		Argsn: 1,
		Pure:  true,
		Doc:   "Constructs a Bunch from a Block, List, or another Bunch. Values are normalised to strings; blocks, lists and dicts are not valid elements.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return makeBunchFromObject(ps, arg0, "bunch")
		},
	},

	// Tests:
	// equal { to-bunch { "a" "b" "a" } |length? } 2
	// equal { to-bunch list { "a" "b" } |contains "b" } true
	// Args:
	// * values: Block, List, or Bunch to convert
	// Returns:
	// * a new Bunch containing the unique values
	"to-bunch": {
		Argsn: 1,
		Pure:  true,
		Doc:   "Converts a Block, List, or Bunch to a Bunch. Alias for bunch.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return makeBunchFromObject(ps, arg0, "to-bunch")
		},
	},

	// Tests:
	// equal { var 'seen bunch { "a" } add! "b" 'seen , seen |length? } 2
	// equal { var 'seen bunch { "a" } add! "a" 'seen , seen |length? } 1
	// equal { var 'seen bunch { "a" } add! "b" 'seen , seen |contains "b" } true
	// Args:
	// * value: Value to add
	// * bunch: Bunch value or Word/Tagword referring to a Bunch variable
	// Returns:
	// * the modified Bunch
	"add!": {
		Argsn: 2,
		Pure:  false,
		Doc:   "Adds a value to a Bunch in-place. The value is ignored if it is already present.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return modifyBunchWord(ps, arg0, arg1, true, "add!")
		},
	},

	// Tests:
	// equal { var 'seen bunch { "a" "b" } remove! "a" 'seen , seen |length? } 1
	// equal { var 'seen bunch { "a" "b" } remove! "c" 'seen , seen |length? } 2
	// equal { var 'seen bunch { "a" "b" } remove! "a" 'seen , seen |contains "a" } false
	// Args:
	// * value: Value to remove
	// * bunch: Bunch value or Word/Tagword referring to a Bunch variable
	// Returns:
	// * the modified Bunch
	"remove!": {
		Argsn: 2,
		Pure:  false,
		Doc:   "Removes a value from a Bunch in-place. Removing a missing value is a no-op.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return modifyBunchWord(ps, arg0, arg1, false, "remove!")
		},
	},
}
