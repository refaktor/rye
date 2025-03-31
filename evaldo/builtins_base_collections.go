package evaldo

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/refaktor/rye/env"

	// JM 20230825	"github.com/refaktor/rye/term"
	"strconv"

	"github.com/refaktor/rye/util"
)

var builtins_collection = map[string]*env.Builtin{

	//
	// ##### Collections ##### ""
	//
	// Tests:
	// equal { random { 1 2 3 } |type? } 'integer
	// equal { random { 1 2 3 } |contains* { 1 2 3 } } 1
	// Args:
	// * block: Block of values to select from
	// Returns:
	// * a randomly selected value from the block
	"random": {
		Argsn: 1,
		Doc:   "Selects a random value from a block of values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Block:
				val, err := rand.Int(rand.Reader, big.NewInt(int64(len(arg.Series.S))))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "random")
				}
				return arg.Series.S[int(val.Int64())]
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "random")
			}
		},
	},

	// Tests:
	// equal { unpack { { 123 } { 234 } } } { 123 234 }
	// equal { unpack { { { 123 } } { 234 } } } { { 123 } 234 }
	// ; equal { unpack list { list { 1 2 } list { 3 4 } } } list { 1 2 3 4 }
	// Args:
	// * collection: Block or list of blocks/lists to unpack
	// Returns:
	// * a flattened block or list with all inner blocks/lists unpacked
	"unpack": {
		Argsn: 1,
		Doc:   "Flattens a block of blocks or list of lists by one level, combining all inner collections into a single collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				res := make([]env.Object, 0)
				for _, val := range bloc.Series.S {
					switch val_ := val.(type) {
					case env.Block:
						res = append(res, val_.Series.S...)
						//for _, val2 := range val_.Series.S {
						//	res = append(res, val2)
						// }
					default:
						res = append(res, val)
					}
				}
				return *env.NewBlock(*env.NewTSeries(res))
			case env.List:
				res := make([]any, 0)
				for _, val := range bloc.Data {
					switch val_ := val.(type) {
					case env.List:
						res = append(res, val_.Data...)
						//for _, val2 := range val_.Series.S {
						//	res = append(res, val2)
						// }
					default:
						res = append(res, val)
					}
				}
				return *env.NewList(res)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "unpack")
			}
		},
	},

	// Tests:
	// equal { sample { 1 2 3 4 } 2 |length? } 2
	// equal { sample { 123 123 123 123 } 3 -> 0 } 123
	// ; equal { sample list { 1 2 3 4 5 } 3 |length? } 3
	// Args:
	// * collection: Block, list or table to sample from
	// * count: Number of random elements to select
	// Returns:
	// * a new collection with randomly selected elements
	"sample": {
		Argsn: 2,
		Doc:   "Randomly selects a specified number of elements from a collection without replacement.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			num, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "sample")
			}
			switch arg := arg0.(type) {
			case env.Block:
				blkLen := len(arg.Series.S)
				if blkLen < int(num.Value) {
					return MakeBuiltinError(ps, "size smaller than sample size", "sample")
				}
				indexes := util.GenSampleIndexes(len(arg.Series.S), int(num.Value))
				newl := make([]env.Object, len(indexes))
				for i := 0; i < len(indexes); i++ {
					newl[i] = arg.Series.S[indexes[i]]
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			case env.Table:
				blkLen := len(arg.Rows)
				if blkLen < int(num.Value) {
					return MakeBuiltinError(ps, "size smaller than sample size", "sample")
				}
				indexes := util.GenSampleIndexes(len(arg.Rows), int(num.Value))
				nspr := env.NewTable(arg.Cols)

				newl := make([]env.TableRow, len(indexes))
				for i := 0; i < len(indexes); i++ {
					newl[i] = arg.Rows[indexes[i]]
				}
				nspr.Rows = newl
				return *nspr
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sample")
			}
		},
	},

	// Tests:
	// equal { max { 8 2 10 6 } } 10
	// equal { max list { 8 2 10 6 } } 10
	// equal { try { max { } } |type? } 'error
	// equal { try { max list { } } |type? } 'error
	// Args:
	// * collection: Block or list of comparable values
	// Returns:
	// * the maximum value in the collection
	"max": { // **
		Argsn: 1,
		Doc:   "Finds the maximum value in a block or list of comparable values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch data := arg0.(type) {
			case env.Block:
				var max env.Object
				l := data.Series.Len()
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "max")
				}
				for i := 0; i < l; i++ {
					if max == nil || greaterThan(ps, data.Series.Get(i), max) {
						max = data.Series.Get(i)
					}
				}
				return max
			case env.List:
				max := math.SmallestNonzeroFloat64
				l := len(data.Data)
				if l == 0 {
					return MakeBuiltinError(ps, "List is empty.", "max")
				}
				var isMaxInt bool
				for i := 0; i < l; i++ {
					switch val1 := data.Data[i].(type) {
					case int64:
						if float64(val1) > max {
							max = float64(val1)
							isMaxInt = true
						}
					case float64:
						if val1 > max {
							max = val1
							isMaxInt = false
						}
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "max")
					}
				}
				if isMaxInt {
					return *env.NewInteger(int64(max))
				} else {
					return *env.NewDecimal(max)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "max")
			}
		},
	},
	// Tests:
	// equal { min { 8 2 10 6 } } 2
	// equal { min list { 8 2 10 6 } } 2
	// equal { try { min { } } |type? } 'error
	// equal { try { min list { } } |type? } 'error
	// Args:
	// * collection: Block or list of comparable values
	// Returns:
	// * the minimum value in the collection
	"min": { // **
		Argsn: 1,
		Doc:   "Finds the minimum value in a block or list of comparable values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch data := arg0.(type) {
			case env.Block:
				var min env.Object
				l := data.Series.Len()
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "min")
				}
				for i := 0; i < l; i++ {
					if min == nil || greaterThan(ps, min, data.Series.Get(i)) {
						min = data.Series.Get(i)
					}
				}
				return min
			case env.List:
				l := len(data.Data)
				if l == 0 {
					return MakeBuiltinError(ps, "List is empty.", "min")
				}
				var isMinInt bool
				min := math.MaxFloat64
				for i := 0; i < l; i++ {
					switch val1 := data.Data[i].(type) {
					case int64:
						if float64(val1) < min {
							min = float64(val1)
							isMinInt = true
						}
					case float64:
						if val1 < min {
							min = val1
							isMinInt = false
						}
					case env.Integer: // TODO -- think about what values really List should hold and when / how it should be used
						if float64(val1.Value) < min {
							min = float64(val1.Value)
							isMinInt = true
						}
					case *env.Integer: // TODO -- think about what values really List should hold and when / how it should be used
						if float64(val1.Value) < min {
							min = float64(val1.Value)
							isMinInt = true
						}
					default:
						fmt.Println(data.Data[i])
						fmt.Printf("t1: %T\n", data.Data[i])
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "min")
					}
				}
				if isMinInt {
					return *env.NewInteger(int64(min))
				} else {
					return *env.NewDecimal(min)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "min")
			}
		},
	},

	// Tests:
	// equal { avg { 8 2 10 6 } } 6.5
	// equal { avg list { 8 2 10 6 } } 6.5
	// equal { avg { 1 2 3 } } 2.0
	// equal { try { avg { } } |type? } 'error
	// equal { try { avg list { } } |type? } 'error
	// Args:
	// * collection: Block, list or vector of numeric values
	// Returns:
	// * the arithmetic mean (average) of the values as a decimal
	"avg": { // **
		Argsn: 1,
		Doc:   "Calculates the arithmetic mean (average) of numeric values in a collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "avg")
				}
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum += float64(val1.Value)
					case env.Decimal:
						sum += val1.Value
					default:
						return MakeBuiltinError(ps, "Block type should be Integer or Decimal.", "avg")
					}
				}
				return *env.NewDecimal(sum / float64(l))
			case env.List:
				l := len(block.Data)
				if l == 0 {
					return MakeBuiltinError(ps, "List is empty.", "avg")
				}
				for i := 0; i < l; i++ {
					obj := block.Data[i]
					switch val1 := obj.(type) {
					case int64:
						sum += float64(val1)
					case float64:
						sum += val1
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "avg")
					}
				}
				return *env.NewDecimal(sum / float64(l))
			case env.Vector:
				return *env.NewDecimal(block.Value.Mean())
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.VectorType}, "avg")
			}
		},
	},
	// Tests:
	// equal { sum { 8 2 10 6 } } 26
	// equal { sum { 8 2 10 6.5 } } 26.5
	// equal { sum { } } 0
	// equal { sum list { 8 2 10 6 } } 26
	// equal { sum list { 8 2 10 6.5 } } 26.5
	// equal { sum list { } } 0
	// Args:
	// * collection: Block, list or vector of numeric values
	// Returns:
	// * the sum of all values (integer if all values are integers, decimal otherwise)
	"sum": { // **
		Argsn: 1,
		Doc:   "Calculates the sum of all numeric values in a collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum += float64(val1.Value)
					case env.Decimal:
						sum += val1.Value
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "Block type should be Integer or Decimal.", "sum")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}
			case env.List:
				l := len(block.Data)
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Data[i]
					switch val1 := obj.(type) {
					case int64:
						sum += float64(val1)
					case float64:
						sum += val1
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "sum")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}

			case env.Vector:
				return *env.NewDecimal(block.Value.Sum())
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.VectorType}, "sum")
			}
		},
	},
	// Tests:
	// equal { mul { 1 2 3 4 5 } } 120
	// equal { mul { 1 2.0 3.3 4 5 } } 132.0
	// equal { mul { 2 3 4 } } 24
	// Args:
	// * collection: Block, list or vector of numeric values
	// Returns:
	// * the product of all values (integer if all values are integers, decimal otherwise)
	"mul": { // **
		Argsn: 1,
		Doc:   "Calculates the product of all numeric values in a collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64 = 1
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum *= float64(val1.Value)
					case env.Decimal:
						sum *= val1.Value
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "Block type should be Integer or Decimal.", "mul")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}
			case env.List:
				l := len(block.Data)
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Data[i]
					switch val1 := obj.(type) {
					case int64:
						sum *= float64(val1)
					case float64:
						sum *= val1
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "mul")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}

			case env.Vector:
				return *env.NewDecimal(block.Value.Sum())
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.VectorType}, "mul")
			}
		},
	},

	// BASIC SERIES FUNCTIONS

	// Tests:
	// equal { first { 1 2 3 4 } } 1
	// equal { first "abcde" } "a"
	// equal { first list { 1 2 3 } } 1
	// ; equal { first table { 'a 'b } { 1 2 } { 3 4 } } table-row { 'a 1 'b 2 }
	// Args:
	// * collection: Block, list, string or table to get the first item from
	// Returns:
	// * the first item in the collection
	"first": { // **
		Argsn: 1,
		Doc:   "Retrieves the first item from a collection (block, list, string, or table).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "first")
				}
				return s1.Series.Get(int(0))
			case env.List:
				if len(s1.Data) == 0 {
					return MakeBuiltinError(ps, "List is empty.", "first")
				}
				return env.ToRyeValue(s1.Data[int(0)])
			case env.String:
				str := []rune(s1.Value)
				if len(str) == 0 {
					return MakeBuiltinError(ps, "String is empty.", "first")
				}
				return *env.NewString(string(str[0]))
			case env.Table:
				if s1.NRows() == 0 {
					return MakeBuiltinError(ps, "Table is empty.", "first")
				}
				return s1.GetRow(ps, int(0))
			case *env.Table:
				if s1.NRows() == 0 {
					return MakeBuiltinError(ps, "Table is empty.", "first")
				}
				return s1.GetRow(ps, int(0))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType, env.BlockType, env.StringType, env.ListType}, "first")
			}
		},
	},

	// Tests:
	// equal { rest { 1 2 3 4 } } { 2 3 4 }
	// equal { rest "abcde" } "bcde"
	// equal { rest list { 1 2 3 } } list { 2 3 }
	// equal { rest { 1 } } { }
	// Args:
	// * collection: Block, list or string to get all but the first item from
	// Returns:
	// * a new collection containing all items except the first one
	"rest": { // **
		Argsn: 1,
		Doc:   "Creates a new collection with all items except the first one from the input collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "rest")
				}
				return *env.NewBlock(*env.NewTSeries(s1.Series.S[1:]))
			case env.List:
				if len(s1.Data) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "rest")
				}
				return env.NewList(s1.Data[int(1):])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 1 {
					return MakeBuiltinError(ps, "String has only one element.", "rest")
				}
				return *env.NewString(string(str[1:]))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.StringType, env.ListType}, "rest")
			}
		},
	},

	// Tests:
	// equal { rest\from { 1 2 3 4 5 6 } 3 } { 4 5 6 }
	// equal { rest\from "abcdefg" 1 } "bcdefg"
	// equal { rest\from list { 1 2 3 4 } 2 } list { 3 4 }
	// equal { rest\from { 1 2 3 } 0 } { 1 2 3 }
	// Args:
	// * collection: Block, list or string to get items from
	// * n: Integer position to start from (0-based)
	// Returns:
	// * a new collection containing all items starting from position n
	"rest\\from": { // **
		Argsn: 2,
		Doc:   "Creates a new collection with all items starting from the specified position in the input collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					if len(s1.Series.S) == 0 {
						return MakeBuiltinError(ps, "Block is empty.", "rest\\from")
					}
					if len(s1.Series.S) <= int(num.Value) {
						return MakeBuiltinError(ps, fmt.Sprintf("Block has less than %d elements.", num.Value+1), "rest\\from")
					}
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[int(num.Value):]))
				case env.List:
					if len(s1.Data) == 0 {
						return MakeBuiltinError(ps, "List is empty.", "rest\\from")
					}
					if len(s1.Data) <= int(num.Value) {
						return MakeBuiltinError(ps, fmt.Sprintf("List has less than %d elements.", num.Value+1), "rest\\from")
					}
					return env.NewList(s1.Data[int(num.Value):])
				case env.String:
					str := []rune(s1.Value)
					if len(str) == 0 {
						return MakeBuiltinError(ps, "String is empty.", "rest\\from")
					}
					if len(str) <= int(num.Value) {
						return MakeBuiltinError(ps, fmt.Sprintf("String has less than %d elements.", num.Value+1), "rest\\from")
					}
					return *env.NewString(string(str[int(num.Value):]))
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "rest\\from")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "rest\\from")
			}
		},
	},

	// Tests:
	// equal { tail { 1 2 3 4 5 6 7 } 3 } { 5 6 7 }
	// equal { tail "abcdefg" 4 } "defg"
	// equal { tail list { 1 2 3 4 } 1 } list { 4 }
	// equal { tail { 1 2 } 5 } { 1 2 }
	// Args:
	// * collection: Block, list, string or table to get the last items from
	// * n: Number of items to retrieve from the end
	// Returns:
	// * a new collection containing the last n items
	"tail": { // **
		Argsn: 2,
		Doc:   "Creates a new collection with the last n items from the input collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				numVal := int(num.Value)
				switch s1 := arg0.(type) {
				case env.Block:
					if len(s1.Series.S) == 0 {
						return *env.NewBlock(*env.NewTSeries([]env.Object{}))
					}
					if len(s1.Series.S) < numVal {
						numVal = len(s1.Series.S)
					}
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[len(s1.Series.S)-numVal:]))
				case env.List:
					if len(s1.Data) == 0 {
						return *env.NewList([]any{})
					}
					if len(s1.Data) < numVal {
						numVal = len(s1.Data)
					}
					return *env.NewList(s1.Data[len(s1.Data)-numVal:])
				case env.String:
					str := []rune(s1.Value)
					if len(str) == 0 {
						return *env.NewString("")
					}
					if len(str) < numVal {
						numVal = len(str)
					}
					return *env.NewString(string(str[len(str)-numVal:]))
				case env.Table:
					nspr := env.NewTable(s1.Cols)
					nspr.Rows = s1.Rows[len(s1.Rows)-numVal:]
					return *nspr
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "tail")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "tail")
			}
		},
	},

	// Tests:
	// equal { second { 123 234 345 } } 234
	// equal { second "abc" } "b"
	// equal { second list { 10 20 30 } } 20
	// Args:
	// * collection: Block, list or string to get the second item from
	// Returns:
	// * the second item in the collection
	"second": { // **
		Argsn: 1,
		Doc:   "Retrieves the second item from a collection (block, list, or string).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) < 2 {
					return MakeBuiltinError(ps, "Block has no second element.", "second")
				}
				return s1.Series.Get(1)
			case env.List:
				if len(s1.Data) < 2 {
					return MakeBuiltinError(ps, "List has no second element.", "second")
				}
				return env.ToRyeValue(s1.Data[1])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 2 {
					return MakeBuiltinError(ps, "String has no second element.", "second")
				}
				return *env.NewString(string(str[1]))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "second")
			}
		},
	},

	// Tests:
	// equal { third { 123 234 345 } } 345
	// equal { third "abcde" } "c"
	// equal { third list { 10 20 30 40 } } 30
	// Args:
	// * collection: Block, list or string to get the third item from
	// Returns:
	// * the third item in the collection
	"third": {
		Argsn: 1,
		Doc:   "Retrieves the third item from a collection (block, list, or string).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) < 3 {
					return MakeBuiltinError(ps, "Block has no third element.", "third")
				}
				return s1.Series.Get(int(2))
			case env.List:
				if len(s1.Data) < 3 {
					return MakeBuiltinError(ps, "List has no third element.", "third")
				}
				return env.ToRyeValue(s1.Data[2])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 3 {
					return MakeBuiltinError(ps, "String has no third element.", "third")
				}
				return *env.NewString(string(str[2]))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "third")
			}
		},
	},

	// Tests:
	// equal { last { 1 2 } } 2
	// equal { last "abcd" } "d"
	// equal { last list { 4 5 6 } } 6
	// equal { try { last { } } |type? } 'error
	// Args:
	// * collection: Block, list or string to get the last item from
	// Returns:
	// * the last item in the collection
	"last": { // **
		Argsn: 1,
		Doc:   "Retrieves the last item from a collection (block, list, or string).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "last")
				}
				return s1.Series.Get(s1.Series.Len() - 1)
			case env.List:
				if len(s1.Data) == 0 {
					return MakeBuiltinError(ps, "List is empty.", "last")
				}
				return env.ToRyeValue(s1.Data[len(s1.Data)-1])
			case env.String:
				if len(s1.Value) == 0 {
					return MakeBuiltinError(ps, "String is empty.", "last")
				}
				return *env.NewString(s1.Value[len(s1.Value)-1:])
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "last")
			}
		},
	},

	// Tests:
	// equal { head { 4 5 6 7 } 3 } { 4 5 6 }
	// equal { head "abcdefg" 2 } "ab"
	// equal { head "abcdefg" 4 } "abcd"
	// equal { head list { 10 20 30 40 } 2 } list { 10 20 }
	// equal { head { 4 5 6 7 } -2 } { 4 5 }
	// equal { head "abcdefg" -1 } "abcdef"
	// equal { head "abcdefg" -5 } "ab"
	// equal { head list { 10 20 30 40 } -1 } list { 10 20 30 }
	// Args:
	// * collection: Block, list, string or table to get the first items from
	// * n: Number of items to retrieve (if positive) or number to exclude from the end (if negative)
	// Returns:
	// * a new collection containing the first n items or all but the last |n| items
	"head": { // **
		Argsn: 2,
		Doc:   "Creates a new collection with the first n items from the input collection, or all but the last |n| items if n is negative.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				numVal := int(num.Value)
				switch s1 := arg0.(type) {
				case env.Block:
					if len(s1.Series.S) == 0 {
						return *env.NewBlock(*env.NewTSeries([]env.Object{}))
					}
					if len(s1.Series.S) < numVal {
						numVal = len(s1.Series.S)
					}
					if numVal < 0 {
						numVal = len(s1.Series.S) + numVal // warn: numVal is negative so we must add
					}
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[0:numVal]))
				case env.List:
					if len(s1.Data) == 0 {
						return *env.NewList([]any{})
					}
					if len(s1.Data) < numVal {
						numVal = len(s1.Data)
					}
					if numVal < 0 {
						numVal = len(s1.Data) + numVal // warn: numVal is negative so we must add
					}
					return *env.NewList(s1.Data[0:numVal])
				case env.String:
					str := []rune(s1.Value)
					if len(str) == 0 {
						return *env.NewString("")
					}
					if len(str) < numVal {
						numVal = len(str)
					}
					if numVal < 0 {
						numVal = len(str) + numVal // warn: numVal is negative so we must add
					}
					return *env.NewString(string(str[0:numVal]))
				case env.Table:
					nspr := env.NewTable(s1.Cols)
					nspr.Rows = s1.Rows[0:numVal]
					return *nspr
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType, env.TableType}, "head")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "head")
			}
		},
	},

	// Tests:
	// equal { nth { 1 2 3 4 5 } 4 } 4
	// equal { nth { "a" "b" "c" "d" "e" } 2 } "b"
	// equal { nth "abcde" 3 } "c"
	// equal { nth list { 10 20 30 40 } 2 } 20
	// ; equal { nth table { 'a 'b } { 1 2 } { 3 4 } 2 } table-row { 'a 3 'b 4 }
	// Args:
	// * collection: Block, list, string or table to get an item from
	// * n: Position of the item to retrieve (1-based)
	// Returns:
	// * the item at position n in the collection
	"nth": { // **
		Argsn: 2,
		Doc:   "Retrieves the item at the specified position (1-based indexing) from a collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					if num.Value > int64(s1.Series.Len()) {
						return MakeBuiltinError(ps, fmt.Sprintf("Block has less than %d elements.", num.Value), "nth")
					}
					return s1.Series.Get(int(num.Value - 1))
				case env.List:
					if num.Value > int64(len(s1.Data)) {
						return MakeBuiltinError(ps, fmt.Sprintf("List has less than %d elements.", num.Value), "nth")
					}
					return env.ToRyeValue(s1.Data[int(num.Value-1)])
				case env.String:
					str := []rune(s1.Value)
					if num.Value > int64(len(str)) {
						return MakeBuiltinError(ps, fmt.Sprintf("String has less than %d elements.", num.Value), "nth")
					}
					return *env.NewString(string(str[num.Value-1 : num.Value]))
				case env.Table:
					rows := s1.GetRows()
					if num.Value > int64(len(rows)) {
						return MakeBuiltinError(ps, fmt.Sprintf("Spreadhseet has less than %d rows.", num.Value), "nth")
					}
					return rows[num.Value-1]
				case *env.Table:
					rows := s1.GetRows()
					if num.Value > int64(len(rows)) {
						return MakeBuiltinError(ps, fmt.Sprintf("Spreadhseet has less than %d rows.", num.Value), "nth")
					}
					return rows[num.Value-1]
				default:
					return MakeArgError(ps, 1,
						[]env.Type{env.BlockType, env.ListType, env.StringType, env.TableType},
						"nth")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "nth")
			}
		},
	},

	// Tests:
	// equal { dict { "a" 1 "b" 2 "c" 3 } |values } list { 1 2 3 }
	// equal { dict { "x" 10 "y" 20 } |values |length? } 2
	// Args:
	// * dict: Dictionary object to extract values from
	// Returns:
	// * list containing all values from the dictionary
	"values": { // **
		Argsn: 1,
		Doc:   "Extracts all values from a dictionary and returns them as a list.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dict := arg0.(type) {
			case env.Dict:
				newl := make([]any, 0)
				for _, v := range dict.Data {
					newl = append(newl, v)
				}
				return *env.NewList(newl)
			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "values")
			}
		},
	},

	// Tests:
	// equal { sort { 6 12 1 } } { 1 6 12 }
	// equal { sort x: { 6 12 1 } x } { 6 12 1 }
	// equal { sort { "b" "c" "a" } } { "a" "b" "c" }
	// equal { sort list { 5 3 1 4 } } list { 1 3 4 5 }
	// equal { sort "cba" } "abc"
	// Args:
	// * collection: Block, list or string to sort
	// Returns:
	// * a new collection with items sorted in ascending order
	"sort": { // TODO -- make sort (not in place) and  decide if sort will only work on ref
		Argsn: 1,
		Doc:   "Creates a new collection with items sorted in ascending order.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				copied := make([]env.Object, len(block.Series.S))
				copy(copied, block.Series.S)
				sort.Sort(RyeBlockSort(copied))
				return *env.NewBlock(*env.NewTSeries(copied))
			case env.List:
				copied := make([]any, len(block.Data))
				copy(copied, block.Data)
				sort.Sort(RyeListSort(copied))
				return *env.NewList(copied)
			case env.String:
				copied := []rune(block.Value)
				// copy(copied, block.Data)
				sort.Sort(RyeStringSort(copied))
				return *env.NewString(string(copied))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "sort")
			}
		},
	},

	// Tests:
	// error { x: { 6 12 1 } , sort! x }
	// equal { x: ref { 6 12 1 } , sort! x , x } { 1 6 12 }
	// equal { x: ref list { 5 3 1 4 } , sort! x , x } list { 1 3 4 5 }
	// Args:
	// * collection: Reference to a block or list to sort in-place
	// Returns:
	// * the sorted collection (same reference, modified in-place)
	"sort!": { // TODO -- make sort (not in place) and  decide if sort will only work on ref
		Argsn: 1,
		Doc:   "Sorts a block or list in-place in ascending order and returns the modified collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case *env.Block:
				ss := block.Series.S
				sort.Sort(RyeBlockSort(ss))
				return *env.NewBlock(*env.NewTSeries(ss))
			case *env.List:
				ss := block.Data
				sort.Sort(RyeListSort(ss))
				return *env.NewList(ss)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sort!") // TODO make it report ref
			}
		},
	},

	// Tests:
	// equal { sort\by { 6 12 1 } fn { a b } { a < b } } { 1 6 12 }
	// equal { sort\by { 6 12 1 } fn { a b } { a > b } } { 12 6 1 }
	// equal { sort\by { { x 6 } { x 12 } { x 1 } } fn { a b } { second a |< second b } } { { x 1 } { x 6 } { x 12 } }
	// equal { sort\by list { 5 3 1 4 } fn { a b } { a > b } } list { 5 4 3 1 }
	// Args:
	// * collection: Block or list to sort
	// * comparator: Function that takes two arguments and returns a truthy value if they are in the correct order
	// Returns:
	// * a new collection with items sorted according to the comparator function
	"sort\\by": { // TODO -- make sort (not in place) and  decide if sort will only work on ref
		Argsn: 2,
		Doc:   "Creates a new collection with items sorted according to a custom comparison function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				switch fn := arg1.(type) {
				case env.Function:
					copied := make([]env.Object, len(block.Series.S))
					copy(copied, block.Series.S)
					sorter := RyeBlockCustomSort{copied, fn, ps}
					sort.Sort(sorter)
					return *env.NewBlock(*env.NewTSeries(copied))
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sort!")
				}
			case env.List:
				switch fn := arg1.(type) {
				case env.Function:
					copied := make([]any, len(block.Data))
					copy(copied, block.Data)
					sorter := RyeListCustomSort{copied, fn, ps}
					sort.Sort(sorter)
					return *env.NewList(copied)
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sort!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sort!")
			}
		},
	},

	// Tests:
	// equal { list { 3 2 3 5 3 2 } .unique |sort } list { 2 3 5 }
	// equal { unique list { 1 1 2 2 3 } |sort } list { 1 2 3 }
	// equal { unique list { 1 1 2 2 } |sort } list { 1 2 }
	// equal { unique { 1 1 2 2 3 } |sort } { 1 2 3 }
	// equal { unique { 1 1 2 2 } |sort } { 1 2 }
	// equal { unique "aabbc" |length? } 3
	// equal { unique "ab" |length? } 2
	// Args:
	// * collection: Block, list or string to remove duplicates from
	// Returns:
	// * a new collection with duplicate values removed
	"unique": { // **
		Argsn: 1,
		Doc:   "Creates a new collection with duplicate values removed, keeping only the first occurrence of each value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.List:
				ss := block.Data

				// Create a map to store the unique values.
				// uniqueValues := make(map[string]bool)
				uniqueValues := make(map[any]bool)

				// Iterate over the slice and add the elements to the map.
				for _, element := range ss {
					// uniqueValues[env.ToRyeValue(element).Print(*ps.Idx)] = true
					uniqueValues[element] = true
				}

				// Create a new slice to store the unique values.
				uniqueSlice := make([]any, 0, len(uniqueValues))

				// Iterate over the map and add the keys to the new slice.
				for key := range uniqueValues {
					uniqueSlice = append(uniqueSlice, key)
				}
				return *env.NewList(uniqueSlice)
			case env.Block:
				uniqueList := util.RemoveDuplicate(ps, block.Series.S)
				return *env.NewBlock(*env.NewTSeries(uniqueList))
			case env.String:
				strSlice := make([]env.Object, 0)
				// create string to object slice
				for _, value := range block.Value {
					// if want to block  space then we can add here condition
					strSlice = append(strSlice, env.ToRyeValue(value))
				}
				uniqueStringSlice := util.RemoveDuplicate(ps, strSlice)
				uniqueStr := ""
				// converting object to string and append final
				for _, value := range uniqueStringSlice {
					uniqueStr = uniqueStr + env.RyeToRaw(value, ps.Idx).(string)
				}
				return *env.NewString(uniqueStr)
			default:
				return MakeArgError(ps, 1, []env.Type{env.ListType, env.BlockType, env.StringType}, "unique")
			}
		},
	},

	// Tests:
	// equal { reverse { 3 1 2 3 } } { 3 2 1 3 }
	// equal { reverse "abcd" } "dcba"
	// equal { reverse list { 1 2 3 4 } } list { 4 3 2 1 }
	// equal { reverse { } } { }
	// Args:
	// * collection: Block, list or string to reverse
	// Returns:
	// * a new collection with items in reverse order
	"reverse": { // **
		Argsn: 1,
		Doc:   "Creates a new collection with items in reverse order from the input collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				a := make([]env.Object, len(block.Series.S))
				copy(a, block.Series.S)
				for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
					a[left], a[right] = a[right], a[left]
				}
				// sort.Sort(RyeBlockSort(ss))
				return *env.NewBlock(*env.NewTSeries(a))
			case env.String:
				s := block.Value
				reversed := ""
				for i := len(s) - 1; i >= 0; i-- {
					reversed += string(s[i])
				}
				return *env.NewString(reversed)
			case env.List: // TODO - check if this all is needed, decide in global if list is a temporary data format or keeper
				//                                                should list turn to block after being processed or remain lists?
				// Create slice of env.Object
				dataSlice := make([]env.Object, 0)
				for _, v := range block.Data {
					dataSlice = append(dataSlice, env.ToRyeValue(v))
				}
				// Reverse slice data
				for left, right := 0, len(dataSlice)-1; left < right; left, right = left+1, right-1 {
					dataSlice[left], dataSlice[right] = dataSlice[right], dataSlice[left]
				}
				// Create list frol slice data
				reverseList := make([]any, 0, len(dataSlice))
				for _, value := range dataSlice {
					reverseList = append(reverseList, value)
				}
				return *env.NewList(reverseList)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.StringType, env.ListType}, "reverse!")
			}
		},
	},

	// Tests:
	// error { reverse! { 3 1 2 3 } }
	// equal { reverse! ref { 3 1 2 3 } } { 3 2 1 3 }
	// equal { x: ref list { 1 2 3 } , reverse! x , x } list { 3 2 1 }
	// Args:
	// * collection: Reference to a block or list to reverse in-place
	// Returns:
	// * the reversed collection (same reference, modified in-place)
	"reverse!": { // **
		Argsn: 1,
		Doc:   "Reverses a block or list in-place and returns the modified collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case *env.Block:
				a := block.Series.S
				for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
					a[left], a[right] = a[right], a[left]
				}
				// sort.Sort(RyeBlockSort(ss))
				return *env.NewBlock(*env.NewTSeries(a))
			case *env.List:
				// Create slice of env.Object
				dataSlice := make([]env.Object, 0)
				for _, v := range block.Data {
					dataSlice = append(dataSlice, env.ToRyeValue(v))
				}
				// Reverse slice data
				for left, right := 0, len(dataSlice)-1; left < right; left, right = left+1, right-1 {
					dataSlice[left], dataSlice[right] = dataSlice[right], dataSlice[left]
				}
				// Create list frol slice data
				reverseList := make([]any, 0, len(dataSlice))
				for _, value := range dataSlice {
					reverseList = append(reverseList, value)
				}
				return *env.NewList(reverseList)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.StringType, env.ListType}, "reverse!")
			}
		},
	},

	// Tests:
	// equal { "abcd" .concat "cde" } "abcdcde"
	// equal { concat { 1 2 3 4 } { 2 4 5 } } { 1 2 3 4 2 4 5 }
	// equal { 123 .concat "abc" } "123abc"
	// Args:
	// * value1: First value (string, integer, block) to concatenate
	// * value2: Second value to concatenate with the first
	// Returns:
	// * result of concatenating the two values
	"concat": {
		Argsn: 2,
		Doc:   "Joins two series values together.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(strconv.Itoa(int(s1.Value)) + s2.Value)
				case env.Integer:
					return *env.NewString(strconv.Itoa(int(s1.Value)) + strconv.Itoa(int(s2.Value)))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType}, "concat")
				}
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(s1.Value + s2.Value)
				case env.Integer:
					return *env.NewString(s1.Value + strconv.Itoa(int(s2.Value)))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType}, "concat")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				case env.Object:
					s := &s1.Series
					s1.Series = *s.Append(b2)
					return s1
				default:
					return MakeBuiltinError(ps, "If Arg 1 is Block then Arg 2 should be Block or Object type.", "concat")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.StringType, env.BlockType}, "concat")
			}
		},
	},

	// Tests:
	// equal { "A" ++ "b" } "Ab"
	// equal { "A" ++ 1 } "A1"
	// equal { { 1 2 } ++ { 3 4 } } { 1 2 3 4 }
	// equal { dict { "a" 1 } |++ { "b" 2 } } dict { "a" 1 "b" 2 }
	// equal { dict { "a" 1 } |++ dict { "b" 2 } } dict { "a" 1 "b" 2 }
	// Args:
	// * value1: First value (string, block, dict, etc.)
	// * value2: Second value to join
	// Returns:
	// * result of joining the values, type depends on input types
	"_++": {
		Argsn: 2,
		Doc:   "Joins two values together, with behavior depending on types: concatenates strings, joins blocks, merges dictionaries, etc.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(s1.Value + s2.Value)
				case env.Integer:
					return *env.NewString(s1.Value + strconv.Itoa(int(s2.Value)))
				case env.Decimal:
					return *env.NewString(s1.Value + strconv.FormatFloat(s2.Value, 'f', -1, 64))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType, env.DecimalType}, "_++")
				}
			case env.Uri:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+s2.Value)
				case env.Integer:
					return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+strconv.Itoa(int(s2.Value)))
				case env.Block: // -- TODO turn tagwords and valvar sb strings.Builderues to uri encoded values , turn files into paths ... think more about it
					var str strings.Builder
					sepa := ""
					for i := 0; i < s2.Series.Len(); i++ {
						switch node := s2.Series.Get(i).(type) {
						case env.Word:
							_, err := str.WriteString(sepa + ps.Idx.GetWord(node.Index) + "=")
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for Word type.", "_++")
							}
							sepa = "&"
						case env.String:
							_, err := str.WriteString(node.Value)
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for String type.", "_++")
							}
						case env.Integer:
							_, err := str.WriteString(strconv.Itoa(int(node.Value)))
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for Integer type.", "_++")
							}
						case env.Uri:
							_, err := str.WriteString(node.GetPath())
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for Uri type.", "_++")
							}
						default:
							return MakeBuiltinError(ps, "Value in node is not word, string, int or uri type.", "_++")
						}
					}
					return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+str.String())
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType, env.BlockType}, "_++")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				default:
					return MakeBuiltinError(ps, "Value in Block is not block type.", "_++")
				}
			case env.Dict:
				switch b2 := arg1.(type) {
				case env.Dict:
					return env.MergeTwoDicts(s1, b2)
				case env.Block:
					return env.MergeDictAndBlock(s1, b2.Series, ps.Idx)
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType, env.BlockType}, "_++")
				}
			case env.TableRow:
				switch b2 := arg1.(type) {
				case env.Dict:
					return env.AddTableRowAndDict(s1, b2)
				case env.Block:
					return env.AddTableRowAndBlock(s1, b2.Series, ps.Idx)
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType, env.BlockType}, "_++")
				}
			case env.Time:
				switch b2 := arg1.(type) {
				case env.Integer:
					v := s1.Value.Add(time.Duration(b2.Value * 1000000))
					return *env.NewTime(v)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "_++")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.DictType, env.TableRowType, env.TimeType, env.UriType}, "_++")
			}
		},
	},

	/*
			case env.String:
			switch s2 := arg1.(type) {
			case env.String:
				return *env.NewString(s1.Value + s2.Value)
			case env.Integer:
				return *env.NewString(s1.Value + strconv.Itoa(int(s2.Value)))
			case env.Decimal:
				return *env.NewString(s1.Value + strconv.FormatFloat(s2.Value, 'f', -1, 64))
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType, env.DecimalType}, "_+")
			}
		case env.Uri:
			switch s2 := arg1.(type) {
			case env.String:
				return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+s2.Value)
			case env.Integer:
				return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+strconv.Itoa(int(s2.Value)))
			case env.Block: // -- TODO turn tagwords and valvar sb strings.Builderues to uri encoded values , turn files into paths ... think more about it
				var str strings.Builder
				sepa := ""
				for i := 0; i < s2.Series.Len(); i++ {
					switch node := s2.Series.Get(i).(type) {
					case env.Word:
						_, err := str.WriteString(sepa + ps.Idx.GetWord(node.Index) + "=")
						if err != nil {
							return MakeBuiltinError(ps, "WriteString failed for Word type.", "_+")
						}
						sepa = "&"
					case env.String:
						_, err := str.WriteString(node.Value)
						if err != nil {
							return MakeBuiltinError(ps, "WriteString failed for String type.", "_+")
						}
					case env.Integer:
						_, err := str.WriteString(strconv.Itoa(int(node.Value)))
						if err != nil {
							return MakeBuiltinError(ps, "WriteString failed for Integer type.", "_+")
						}
					case env.Uri:
						_, err := str.WriteString(node.GetPath())
						if err != nil {
							return MakeBuiltinError(ps, "WriteString failed for Uri type.", "_+")
						}
					default:
						return MakeBuiltinError(ps, "Value in node is not word, string, int or uri type.", "_+")
					}
				}
				return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+str.String())
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType, env.BlockType}, "_+")
			}
		case env.Block:
			switch b2 := arg1.(type) {
			case env.Block:
				s := &s1.Series
				s1.Series = *s.AppendMul(b2.Series.GetAll())
				return s1
			default:
				return MakeBuiltinError(ps, "Value in Block is not block type.", "_+")
			}
		case env.Dict:
			switch b2 := arg1.(type) {
			case env.Dict:
				return env.MergeTwoDicts(s1, b2)
			case env.Block:
				return env.MergeDictAndBlock(s1, b2.Series, ps.Idx)
			default:
				return MakeArgError(ps, 2, []env.Type{env.DictType, env.BlockType}, "_+")
			}
		case env.TableRow:
			switch b2 := arg1.(type) {
			case env.Dict:
				return env.AddTableRowAndDict(s1, b2)
			case env.Block:
				return env.AddTableRowAndBlock(s1, b2.Series, ps.Idx)
			default:
				return MakeArgError(ps, 2, []env.Type{env.DictType, env.BlockType}, "_+")
			}

	*/

	// Tests:
	// ; equal { "abcd" .union "cde" } "abcde"
	// equal { union { 1 2 3 4 } { 2 4 5 } |length? } 5 ; order is not certain
	// equal { union list { 1 2 3 4 } list { 2 4 5 } |length? } 5 ; order is not certain
	// equal { union { 8 2 } { 1 9 } |sort } { 1 2 8 9 }
	// equal { union { 1 2 } { } |sort } { 1 2 }
	// equal { union { } { 1 9 } |sort }  { 1 9 }
	// equal { union { } { } } { }
	// equal { union list { 1 2 } list { 1 2 3 4 } |sort } list { 1 2 3 4 }
	// equal { union list { 1 2 } list { 1 } |sort } list { 1 2 }
	// equal { union list { 1 2 } list { } |sort } list { 1 2 }
	// equal { union list { } list { 1 2 } |sort } list { 1 2 }
	// equal { union list { } list { } } list { }
	// Args:
	// * collection1: First block or list
	// * collection2: Second block or list
	// Returns:
	// * a new collection containing all unique values from both collections
	"union": {
		Argsn: 2,
		Doc:   "Combines two collections, removing any duplicate values to create a union of all unique values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					union := util.UnionOfBlocks(ps, s1, b2)
					return *env.NewBlock(*env.NewTSeries(union))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "union")
				}
			case env.List:
				switch l2 := arg1.(type) {
				case env.List:
					union := util.UnionOfLists(ps, s1, l2)
					return *env.NewList(union)
				default:
					return MakeArgError(ps, 2, []env.Type{env.ListType}, "union")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "union")
			}
		},
	},

	// Tests:
	// equal { range 1 5 } { 1 2 3 4 5 }
	// equal { range 5 10 } { 5 6 7 8 9 10 }
	// equal { range -2 2 } { -2 -1 0 1 2 }
	// Args:
	// * start: Integer starting value (inclusive)
	// * end: Integer ending value (inclusive)
	// Returns:
	// * a block containing all integers from start to end, inclusive
	"range": { // **
		Argsn: 2,
		Doc:   "Creates a block containing all integers from the start value to the end value, inclusive.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch i1 := arg0.(type) {
			case env.Integer:
				switch i2 := arg1.(type) {
				case env.Integer:
					objs := make([]env.Object, i2.Value-i1.Value+1)
					idx := 0
					for i := i1.Value; i <= i2.Value; i++ {
						objs[idx] = *env.NewInteger(i)
						idx += 1
					}
					return *env.NewBlock(*env.NewTSeries(objs))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "range")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "range")
			}
		},
	},
	// Tests:
	// equal { { } .is-empty } 1
	// equal { dict { } |is-empty } 1
	// equal { table { 'a 'b } { } |is-empty } 1
	// equal { "abc" .is-empty } 0
	// equal { { 1 2 3 } .is-empty } 0
	// Args:
	// * collection: String, block, dict, list, table, context or vector to check
	// Returns:
	// * integer 1 if the collection is empty, 0 otherwise
	"is-empty": { // **
		Argsn: 1,
		Doc:   "Checks if a collection is empty, returning 1 for empty collections and 0 otherwise.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewBoolean(len(s1.Value) == 0)
			case env.Dict:
				return *env.NewBoolean(len(s1.Data) == 0)
			case env.List:
				return *env.NewBoolean(len(s1.Data) == 0)
			case env.Block:
				return *env.NewBoolean(s1.Series.Len() == 0)
			case env.Table:
				return *env.NewBoolean(len(s1.Rows) == 0)
			case *env.Table:
				return *env.NewBoolean(len(s1.Rows) == 0)
			case env.RyeCtx:
				return *env.NewBoolean(s1.GetWords(*ps.Idx).Series.Len() == 0)
			case env.Vector:
				return *env.NewBoolean(s1.Value.Len() == 0)
			default:
				fmt.Println(s1)
				return MakeArgError(ps, 2, []env.Type{env.StringType, env.DictType, env.ListType, env.BlockType, env.TableType, env.VectorType}, "length?")
			}
		},
	},
	// Tests:
	// equal { { 1 2 3 } .length? } 3
	// equal { length? "abcd" } 4
	// equal { table { 'val } { 1 2 3 4 } |length? } 4
	// equal { vector { 10 20 30 } |length? } 3
	// equal { dict { "a" 1 "b" 2 } |length? } 2
	// Args:
	// * collection: String, block, dict, list, table, context or vector to measure
	// Returns:
	// * integer count of elements in the collection
	"length?": { // **
		Argsn: 1,
		Doc:   "Returns the number of elements in a collection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewInteger(int64(len(s1.Value)))
			case env.Dict:
				return *env.NewInteger(int64(len(s1.Data)))
			case env.List:
				return *env.NewInteger(int64(len(s1.Data)))
			case env.Block:
				return *env.NewInteger(int64(s1.Series.Len()))
			case env.Table:
				return *env.NewInteger(int64(len(s1.Rows)))
			case *env.Table:
				return *env.NewInteger(int64(len(s1.Rows)))
			case env.RyeCtx:
				return *env.NewInteger(int64(s1.GetWords(*ps.Idx).Series.Len()))
			case env.Vector:
				return *env.NewInteger(int64(s1.Value.Len()))
			default:
				fmt.Println(s1)
				return MakeArgError(ps, 2, []env.Type{env.StringType, env.DictType, env.ListType, env.BlockType, env.TableType, env.VectorType}, "length?")
			}
		},
	},
	// Tests:
	// equal { dict { "a" 1 "b" 2 "c" 3 } |keys |length? } 3
	// equal { table { "a" "b" "c" } { 1 2 3 } |keys |length? } 3
	// ; TODO -- doesn't work yet, .header? also has the same problem -- equal { table { 'a 'b 'c } { 1 2 3 } |keys } { 'a 'b 'c }
	// Args:
	// * collection: Dict or table to extract keys from
	// Returns:
	// * block containing all keys from the dictionary or column names from the table
	"keys": {
		Argsn: 1,
		Doc:   "Extracts the keys from a dictionary or column names from a table as a block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:
				keys := make([]env.Object, len(s1.Data))
				i := 0
				for k := range s1.Data {
					keys[i] = *env.NewString(k)
					i++
				}
				return *env.NewBlock(*env.NewTSeries(keys))
			case env.Table:
				return s1.GetColumns()
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},

	// Tests:
	// equal { { 23 34 45 } -> 1 } 34
	// equal { { "a" "b" "c" } -> 0 } "a"
	// equal { dict { "a" 1 "b" 2 } -> "b" } 2
	// Args:
	// * collection: Block, list, dict or other indexable collection
	// * index: Index or key to access
	// Returns:
	// * value at the specified index or key
	"_->": {
		Argsn: 2,
		Doc:   "Accesses a value in a collection by index or key (1-based indexing for blocks and lists).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg0, arg1, false)
		},
	},
	// Tests:
	// equal { 0 <- { 23 34 45 } } 23
	// equal { 2 <- { "a" "b" "c" } } "c"
	// equal { "a" <- dict { "a" 1 "b" 2 } } 1
	// Args:
	// * index: Index or key to access
	// * collection: Block, list, dict or other indexable collection
	// Returns:
	// * value at the specified index or key
	"_<-": {
		Argsn: 2,
		Doc:   "Accesses a value in a collection by index or key, with reversed argument order (0-based indexing for blocks and lists).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg1, arg0, false)
		},
	},
	/* "_<-": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				switch s2 := arg1.(type) {
				case env.Integer:
					idx := s2.Value
					//					if posMode {
					// 	idx--
					//}
					v := s1.Series.PGet(int(idx))
					ok := true
					if ok {
						return v
					} else {
						ps.FailureFlag = true
						return env.NewError1(5) // NOT_FOUND
					}
				}
				//return getFrom(ps, arg1, arg0, false)
				return nil
			}
			return nil
		},
	},*/
	// Tests:
	// equal { 2 <~ { 23 34 45 } } 34
	// equal { 1 <~ { "a" "b" "c" } } "b"
	// Args:
	// * index: Index to access (1-based)
	// * collection: Block, list or other indexable collection
	// Returns:
	// * value at the specified index
	"_<~": {
		Argsn: 2,
		Doc:   "Accesses a value in a collection by index with reversed argument order (1-based indexing).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg1, arg0, true)
		},
	},
	// Tests:
	// equal { { 23 34 45 } ~> 1 } 23
	// equal { { "a" "b" "c" } ~> 1 } "a"
	// Args:
	// * collection: Block, list or other indexable collection
	// * index: Index to access (0-based)
	// Returns:
	// * value at the specified index
	"_~>": {
		Argsn: 2,
		Doc:   "Accesses a value in a collection by index (0-based indexing).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg0, arg1, true)
		},
	},

	// Tests:
	// equal { "abcd" .intersection "cde" } "cd"
	// equal { intersection { 1 2 3 4 } { 2 4 5 } } { 2 4 }
	// equal { intersection { 1 3 5 6 } { 2 3 4 5 } } { 3 5 }
	// equal { intersection { 1 2 3 } { } } {  }
	// equal { intersection { } { 2 3 4  } } { }
	// equal { intersection { 1 2 3 } { 4 5 6 } } { }
	// equal { intersection { } { } } { }
	// equal { intersection list { 1 3 5 6 } list { 2 3 4 5 } } list { 3 5 }
	// equal { intersection list { 1 2 3 } list { } } list {  }
	// equal { intersection list { } list { 2 3 4 } } list { }
	// equal { intersection list { 1 2 3 } list { 4 5 6 } } list { }
	// equal { intersection list { } list { } } list { }
	// Args:
	// * collection1: First string, block or list
	// * collection2: Second string, block or list (same type as first)
	// Returns:
	// * a new collection containing only values that appear in both input collections
	"intersection": {
		Argsn: 2,
		Doc:   "Finds the common elements between two collections, returning only values that appear in both.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					inter := util.IntersectStrings(s1.Value, s2.Value)
					return *env.NewString(inter)
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "intersection")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					inter := util.IntersectBlocks(ps, s1, b2)
					return *env.NewBlock(*env.NewTSeries(inter))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "intersection")
				}
			case env.List:
				switch l2 := arg1.(type) {
				case env.List:
					inter := util.IntersectLists(ps, s1, l2)
					return *env.NewList(inter)
				default:
					return MakeArgError(ps, 2, []env.Type{env.ListType}, "intersection")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.ListType}, "intersection")
			}
		},
	},

	// Tests:
	// equal { intersection\by "foobar" "fbx" fn { a b } { a .contains b } } "fb"
	// equal { intersection\by "fooBar" "Fbx" fn { a b } { a .to-lower .contains to-lower b } } "fB"
	// equal { intersection\by { "foo" 33 } { 33 33 } fn { a b } { a .contains b } } { 33 }
	// equal { intersection\by { "foo" "bar" 33 } { 42 } fn { a b } { map a { .type? } |contains b .type? } } { 33 }
	// equal { intersection\by { { "foo" x } { "bar" y } } { { "bar" z } } fn { a b } { map a { .first } |contains first b } } { { "bar" y } }
	// Args:
	// * collection1: First string or block
	// * collection2: Second string or block
	// * comparator: Function that takes two arguments and returns a truthy value if they should be considered matching
	// Returns:
	// * a new collection containing values from the first collection that match with values from the second collection according to the comparator
	"intersection\\by": {
		Argsn: 3,
		Doc:   "Finds the intersection of two collections using a custom comparison function to determine matching elements.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					switch s3 := arg2.(type) {
					case env.Function:
						inter := IntersectStringsCustom(s1, s2, ps, s3)
						return *env.NewString(inter)
					default:
						return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "intersection\\by")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "intersection\\by")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					switch s3 := arg2.(type) {
					case env.Function:
						inter := IntersectBlocksCustom(s1, b2, ps, s3)
						return *env.NewBlock(*env.NewTSeries(inter))
					default:
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "intersection\\by")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "intersection\\by")
				}
			/* case env.List:
			switch l2 := arg1.(type) {
			case env.List:
				inter := util.IntersectLists(ps, s1, l2)
				return *env.NewList(inter)
			default:
				return MakeArgError(ps, 2, []env.Type{env.ListType}, "intersection")
			} */
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.ListType}, "intersection")
			}
		},
	},

	// Tests:
	// equal { "abcde" .difference "cde" } "ab"
	// equal { difference { 1 2 3 4 } { 2 4 } } { 1 3 }
	// equal { difference list { "Bob" "Sal" "Joe" } list { "Joe" } } list { "Bob" "Sal" }
	// equal { difference "abc" "bc" } "a"
	// equal { difference "abc" "abc" } ""
	// equal { difference "abc" "" } "abc"
	// equal { difference "" "" } ""
	// equal { difference { 1 3 5 6 } { 2 3 4 5 } } { 1 6 }
	// equal { difference { 1 2 3 } {  } } { 1 2 3 }
	// equal { difference { } { 2 3 4  } } { }
	// equal { difference { } { } } { }
	// equal { difference list { 1 3 5 6 } list { 2 3 4 5 } } list { 1 6 }
	// equal { difference list { 1 2 3 } list {  } } list { 1 2 3 }
	// equal { difference list { } list { 2 3 4 } } list { }
	// equal { difference list { } list { } } list { }
	// Args:
	// * collection1: First string, block or list
	// * collection2: Second string, block or list (same type as first)
	// Returns:
	// * a new collection containing values from the first collection that do not appear in the second collection
	"difference": {
		Argsn: 2,
		Doc:   "Creates a new collection containing elements from the first collection that are not present in the second collection.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					diff := util.DiffStrings(s1.Value, s2.Value)
					return *env.NewString(diff)
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "difference")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					diff := util.DiffBlocks(ps, s1, b2)
					return *env.NewBlock(*env.NewTSeries(diff))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "difference")
				}
			case env.List:
				switch l2 := arg1.(type) {
				case env.List:
					diff := util.DiffLists(ps, s1, l2)
					return *env.NewList(diff)
				default:
					return MakeArgError(ps, 2, []env.Type{env.ListType}, "difference")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.ListType}, "difference")
			}
		},
	},

	// add distinct? and count? functions
	// make functions work with list, which column and row can return

	// Tests:
	// equal { { { 1 2 3 } { 4 5 6 } } .transpose } { { 1 4 } { 2 5 } { 3 6 } }
	// equal { { { 1 4 } { 2 5 } { 3 6 } } .transpose } { { 1 2 3 } { 4 5 6 } }
	// Args:
	// * matrix: Block of blocks representing a matrix
	// Returns:
	// * transposed matrix (rows become columns and columns become rows)
	"transpose": {
		Argsn: 1,
		Doc:   "Transposes a matrix (block of blocks), converting rows to columns and columns to rows.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch blk := arg0.(type) {
			case env.Block:

				// Get the number of subblocks
				rows := len(blk.Series.S)

				// get the size of first block
				frstBlk, ok := blk.Series.S[0].(env.Block)
				if !ok {
					return MakeBuiltinError(ps, "First element is not block", "transpose")
				}
				cols := len(frstBlk.Series.S)

				// Create the matrix from which new blocks will be created
				transposed := make([][]env.Object, cols)
				for i := range transposed {
					transposed[i] = make([]env.Object, rows)
				}

				// Fill the contents of new block
				for i := 0; i < rows; i++ {
					subBlk, ok := blk.Series.S[i].(env.Block)
					if !ok {
						return MakeBuiltinError(ps, fmt.Sprintf("Element %d is not block", i), "transpose")
					}
					for j := 0; j < cols; j++ {
						transposed[j][i] = subBlk.Series.S[j]
					}
				}

				// we could probably make this in for loop above ... look at it in the future
				newblk := make([]env.Object, cols)
				for top := 0; top < cols; top++ {
					subblk := make([]env.Object, rows)
					for sub := 0; sub < rows; sub++ {
						subblk[sub] = transposed[top][sub]
					}
					newblk[top] = *env.NewBlock(*env.NewTSeries(subblk))
				}

				return *env.NewBlock(*env.NewTSeries(newblk))
				// From the matrix create the new subblocks and blocks

			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "next")
			}
		},
	},

	// Tests:
	// equal { x: ref { 1 2 3 4 } remove-last! 'x x } { 1 2 3 }
	// equal { x: ref { 1 2 3 4 } remove-last! 'x } { 1 2 3 }
	// Args:
	// * word: Word referring to a block to modify
	// Returns:
	// * the modified block with the last element removed
	"remove-last!": { // **
		Argsn: 1,
		Pure:  false,
		Doc:   "Removes the last element from a block in-place and returns the modified block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg0.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch oldval := val.(type) {
					case *env.Block:
						s := &oldval.Series
						oldval.Series = *s.RmLast()
						ctx.Mod(wrd.Index, oldval)
						return oldval
					default:
						return MakeBuiltinError(ps, "Old value should be Block type.", "remove-last!")
					}
				} else {
					return MakeBuiltinError(ps, "Word not found in context.", "remove-last!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "remove-last!")
			}
		},
	},
	// Tests:
	// ; TODO equal { x: ref { 1 2 3 } append! { 4 } x , x } { 1 2 3 4 }
	// equal { x: ref { 1 2 3 } append! 4 'x , x } { 1 2 3 4 }
	// equal { s: "hello" append! " world" 's , s } "hello world"
	// Args:
	// * value: Value to append
	// * word: Word referring to a block, list or string to modify
	// Returns:
	// * the modified collection with the value appended
	"append!": { // **
		Argsn: 2,
		Doc:   "Appends a value to a block, list or string in-place and returns the modified collection.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch oldval := val.(type) {
					case env.String:
						var newval env.String
						switch s3 := arg0.(type) {
						case env.String:
							newval = *env.NewString(oldval.Value + s3.Value)
						case env.Integer:
							newval = *env.NewString(oldval.Value + strconv.Itoa(int(s3.Value)))
						}
						ctx.Mod(wrd.Index, newval)
						return newval
					case *env.Block: // TODO
						// 	fmt.Println(123)
						s := &oldval.Series
						oldval.Series = *s.Append(arg0)
						ctx.Mod(wrd.Index, oldval)
						return oldval
					case *env.List:
						dataSlice := make([]any, 0)
						switch listData := arg0.(type) {
						case env.List:
							for _, v1 := range oldval.Data {
								dataSlice = append(dataSlice, env.ToRyeValue(v1))
							}
							for _, v2 := range listData.Data {
								dataSlice = append(dataSlice, env.ToRyeValue(v2))
							}
						default:
							return makeError(ps, "Need to pass List of data")
						}
						combineList := make([]any, 0, len(dataSlice))
						for _, v := range dataSlice {
							combineList = append(combineList, env.ToRyeValue(v))
						}
						finalList := *env.NewList(combineList)
						ctx.Mod(wrd.Index, finalList)
						return finalList
					default:
						return makeError(ps, "Type of tagword is not String or Block")
					}
				}
				return makeError(ps, "Tagword not found.")
			case *env.Block:
				dataSlice := make([]env.Object, 0)
				switch blockData := arg0.(type) {
				case env.Block:
					for _, v1 := range wrd.Series.S {
						dataSlice = append(dataSlice, env.ToRyeValue(v1))
					}
					for _, v2 := range blockData.Series.S {
						dataSlice = append(dataSlice, env.ToRyeValue(v2))
					}
				default:
					return makeError(ps, "Need to pass block of data")
				}
				return *env.NewBlock(*env.NewTSeries(dataSlice))
			/* case env.String:
			finalStr := ""
			switch str := arg0.(type) {
			case env.String:
				finalStr = wrd.Value + str.Value
			case env.Integer:
				finalStr = wrd.Value + strconv.Itoa(int(str.Value))
			}
			return *env.NewString(finalStr)*/
			default:
				return makeError(ps, "Value not tagword")
			}
		},
	},
	// Tests:
	// equal { x: ref { 1 2 3 } change\nth! x 2 222 , x } { 1 222 3 }
	// equal { x: ref list { "a" "b" "c" } change\nth! x 1 "X" , x } list { "X" "b" "c" }
	// Args:
	// * collection: Reference to a block or list to modify
	// * position: Position of the element to change (1-based)
	// * value: New value to set at the specified position
	// Returns:
	// * the modified collection with the value changed at the specified position
	"change\\nth!": { // **
		Argsn: 3,
		Doc:   "Changes the value at a specific position in a block or list in-place and returns the modified collection.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case *env.Block:
					if num.Value > int64(s1.Series.Len()) {
						return MakeBuiltinError(ps, fmt.Sprintf("Block has less than %d elements.", num.Value), "change\\nth!")
					}
					s1.Series.S[num.Value-1] = arg2
					return s1
				case *env.List:
					if num.Value > int64(len(s1.Data)) {
						return MakeBuiltinError(ps, fmt.Sprintf("List has less than %d elements.", num.Value), "change\\nth!")
					}
					s1.Data[num.Value-1] = env.RyeToRaw(arg2, ps.Idx)
					return s1
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "change\\nth!")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "change\\nth!")
			}
		},
	},

	// end of collections exploration

	// These are rebol like functions for blocks with carret ... I'm not sure yet it they will be included in long term
	// a carret is an imperative concept, doint blocks on the otheh hand requires a carret

	// Tests:
	// equal { x: { 1 2 3 } peek x } 1
	// Args:
	// * block: Block to peek at
	// Returns:
	// * the current value at the block's cursor position without advancing the cursor
	"peek": {
		Argsn: 1,
		Doc:   "Returns the current value at a block's cursor position without advancing the cursor.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				r := s1.Series.Peek()
				if r != nil {
					return r
				} else {
					return MakeBuiltinError(ps, "past end", "peek")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "peek")
			}
		},
	},
	// Tests:
	// equal { x: { 1 2 3 } pop x } 1
	// Args:
	// * block: Block to pop from
	// Returns:
	// * the current value at the block's cursor position, advancing the cursor
	"pop": {
		Argsn: 1,
		Doc:   "Returns the current value at a block's cursor position and advances the cursor.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Pop()
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "pop")
			}
		},
	},
	// Tests:
	// equal { x: { 1 2 3 } pos x } 0
	// equal { x: { 1 2 3 } next x pos x } 1
	// Args:
	// * block: Block to get position from
	// Returns:
	// * the current cursor position in the block
	"pos": {
		Argsn: 1,
		Doc:   "Returns the current cursor position in a block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return *env.NewInteger(int64(s1.Series.Pos()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "pos")
			}
		},
	},

	// Question: Should blocks really have cursor, like in Rebol? We don't really use them in that stateful way, except maybe when they
	// represent code? Look at it.
	// Tests:
	// equal { x: { 1 2 3 } next x pos x } 1
	// Args:
	// * block: Block to advance cursor in
	// Returns:
	// * the block with its cursor advanced to the next position
	"next": {
		Argsn: 1,
		Doc:   "Advances the cursor position in a block and returns the block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				s1.Series.Next()
				return s1
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "next")
			}
		},
	},

	// TODOC
	// Tests:
	// equal { x: 1 y: 2 vals { x y } } { 1 2 }
	// equal { x: 1 y: 2 vals { 1 y } } { 1 2 }
	// equal { x: 1 y: 2 try { vals { z y } } |type? } 'error
	"vals": { // **
		Argsn: 1,
		Doc:   "Takes a block of Rye values and evaluates each value or expression.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				res := make([]env.Object, 0)
				for ps.Ser.Pos() < ps.Ser.Len() {
					// ps, injnow = EvalExpressionInj(ps, inj, injnow)
					EvalExpression2(ps, false)
					if checkErrorReturnFlag(ps) {
						return ps.Res
					}
					res = append(res, ps.Res)
					// check and raise the flags if needed if true (error) return
					//if checkFlagsAfterBlock(ps, 101) {
					//	return ps
					//}
					// if return flag was raised return ( errorflag I think would return in previous if anyway)
					// ps, injnow = MaybeAcceptComma(ps, inj, injnow)
				}
				ps.Ser = ser
				return *env.NewBlock(*env.NewTSeries(res))
			case env.Word:
				val, found := ps.Ctx.Get(bloc.Index)
				if found {
					return val
				}
				return MakeBuiltinError(ps, "Value not found.", "vals")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.WordType}, "vals")
			}
		},
	},

	// Tests:
	// equal { x: 1 y: 2 vals\with 10 { + x , * y } } { 11 20 }
	// equal { x: 1 y: 2 vals\with 100 { + 10 , * 8.9 } } { 110 890.0 }
	"vals\\with": {
		Argsn: 2,
		Doc:   "Evaluate a block with injecting the first argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				res := make([]env.Object, 0)
				injnow := true
				for ps.Ser.Pos() < ps.Ser.Len() {
					// ps, injnow = EvalExpressionInj(ps, inj, injnow)
					//20231203 EvalExpressionInjectedVALS(ps, arg0, true)
					injnow = EvalExpressionInj(ps, arg0, injnow)
					res = append(res, ps.Res)
					// check and raise the flags if needed if true (error) return
					//if checkFlagsAfterBlock(ps, 101) {
					//	return ps
					//}
					// if return flag was raised return ( errorflag I think would return in previous if anyway)
					//if checkErrorReturnFlag(ps) {
					//	return ps
					//}
					injnow = MaybeAcceptComma(ps, arg0, injnow)
				}
				ps.Ser = ser
				return *env.NewBlock(*env.NewTSeries(res))
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "vals\\with")
			}
		},
	},
}
