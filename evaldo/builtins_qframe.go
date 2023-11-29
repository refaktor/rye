//go:build b_qframe
// +build b_qframe

package evaldo

import (
	"fmt"
	"io"
	"rye/env"
	"strconv"

	"github.com/tobgu/qframe"
	"github.com/tobgu/qframe/config/groupby"
)

type colInfo struct {
	Num  int
	Type int
	Name string
}

func eval_new_frame(ps *env.ProgramState, block env.Block) env.Object {
	//var colType = 0 // 1 - ints, 2 - strings
	//var col string
	//var colNum int
	var col colInfo
	mmap := make(map[string]interface{})
	var scol []string
	var icol []int
	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Pop()
		switch obj1 := obj.(type) {
		case env.Tagword:
			if col.Type != 0 {
				if !_addColumToMap(&mmap, icol, scol, &col) {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Problem at adding column to map.", "new-frame")
				}
				scol = make([]string, 0)
				icol = make([]int, 0)
			}
			col.Name = ps.Idx.GetWord(obj1.Index)
		case env.String:
			switch col.Type {
			case 0, 2:
				col.Type = 2
				scol = append(scol, obj1.Value)
			default:
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "String in non-string column.", "new-frame")
			}
		case env.Integer:
			switch col.Type {
			case 0, 1:
				col.Type = 1
				icol = append(icol, int(obj1.Value))
			default:
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Integer in non-integer column.", "new-frame")
			}
		case env.Comma:
			if !_addColumToMap(&mmap, icol, scol, &col) {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Problem at adding column to map.", "new-frame")
			}
		}
	}
	if !_addColumToMap(&mmap, icol, scol, &col) {
		ps.FailureFlag = true
		return MakeBuiltinError(ps, "Problem at adding column to map.", "new-frame")
	}
	//qframe.New(map[string]interface{}{"COL1": []int{1, 2, 3}, "COL2": []string{"a", "b", "c"}})
	return *env.NewNative(ps.Idx, qframe.New(mmap), "rye-frame")
}

// aggregate { sum 'col1 count 'col2 }

type aggregate_state struct {
	fn  *env.Word
	col *env.Tagword
}

func eval_aggregate(ps *env.ProgramState, grouper qframe.Grouper, block env.Block) env.Object {
	// since this is usually not hotcode function we use slice here, we could statically switch for up to N aggregares
	var aggs []qframe.Aggregation
	var state aggregate_state
	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Pop()
		switch obj1 := obj.(type) {
		case env.Tagword:
			state.col = &obj1
		case env.Word:
			state.fn = &obj1
		default:
			return MakeBuiltinError(ps, "Only tag-words and words in this udialect.", "rye-frame-grouper//aggregate")
		}
		if state.col != nil && state.fn != nil {
			aggs = append(aggs, qframe.Aggregation{Fn: ps.Idx.GetWord(state.fn.Index), Column: ps.Idx.GetWord(state.col.Index)})
			state.col = nil
			state.fn = nil
		}
	}
	return *env.NewNative(ps.Idx, grouper.Aggregate(aggs...), "rye-frame")
}

// filter fr { 'col1 > 2 }
// filter fr { or { 'col1 > 2 , col2 = "a" } }

type filter_state struct {
	step int
	fn   *env.Word
	col  *env.Tagword
	val  interface{}
}

func _emptyFC() []qframe.FilterClause {
	return []qframe.FilterClause{}
}

func eval_filter(ps *env.ProgramState, frame qframe.QFrame, block env.Block) ([]qframe.FilterClause, *env.Error) {

	idxOr, _ := ps.Idx.GetIndex("or")
	idxAnd, _ := ps.Idx.GetIndex("and")

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Word:
			block.Series.Next()
			blk := block.Series.Pop()
			switch blk1 := blk.(type) {
			case env.Block:
				switch obj1.Index {
				case idxOr:
					clauses, err := eval_filter(ps, frame, blk1)
					return []qframe.FilterClause{qframe.Or(clauses...)}, err
				case idxAnd:
					clauses, err := eval_filter(ps, frame, blk1)
					return []qframe.FilterClause{qframe.And(clauses...)}, err
				default:
					return _emptyFC(), env.NewError("err 1")
				}
			default:
				return _emptyFC(), env.NewError("err 2")
			}
		case env.Tagword:
			return eval_filter_clauses(ps, frame, block)
		default:
			//fmt.Println(obj1)
			return _emptyFC(), env.NewError("wrong type for frame/filter udialect")
		}
	}
	return _emptyFC(), env.NewError("didn't get all tokens")
}

func trace3(s string) {
	//fmt.Println(s)
}

func eval_filter_clauses(ps *env.ProgramState, frame qframe.QFrame, block env.Block) ([]qframe.FilterClause, *env.Error) {
	// since this is usually not hotcode function we use slice here, we could statically switch for up to N aggregares
	trace3("filter_clauses")
	var s filter_state
	var clauses []qframe.FilterClause
	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Pop()
		switch obj1 := obj.(type) {
		case env.Tagword:
			if s.step == 0 {
				s.step = 1
				s.col = &obj1
			} else {
				return _emptyFC(), env.NewError("tagword not in right position")
			}
		case env.Word:
			if s.step == 1 {
				s.step = 2
				s.fn = &obj1
			} else {
				return _emptyFC(), env.NewError("word not in right position")
			}
		case env.String, env.Integer, env.Getword:
			if s.step == 2 {
				s.step = 3
				switch obj2 := obj1.(type) {
				case env.Integer:
					s.val = int(obj2.Value)
				case env.String:
					s.val = obj2.Value
				case env.Getword:
					v, e := ps.Ctx.Get(obj2.Index)
					if !e {
						s.val = v
					} else {
						return _emptyFC(), env.NewError("get-word value not found in context")
					}
				}
			} else {
				return _emptyFC(), env.NewError("word not in right position")
			}
		case env.Comma:
			if s.step == 3 {
				clauses = append(clauses, qframe.Filter{Column: ps.Idx.GetWord(s.col.Index), Comparator: ps.Idx.GetWord(s.fn.Index), Arg: s.val})
				s.step = 0
			}
		default:
			return _emptyFC(), env.NewError("wrong type for frame/filter udialect")
		}
		if s.step == 3 {
			clauses = append(clauses, qframe.Filter{Column: ps.Idx.GetWord(s.col.Index), Comparator: ps.Idx.GetWord(s.fn.Index), Arg: s.val})
		}
	}
	return clauses, nil
}

func _addColumToMap(mmap *map[string]interface{}, icol []int, scol []string, col *colInfo) bool {
	var col1 string
	if col.Name == "" {
		col1 = "col" + strconv.Itoa(col.Num)
	} else {
		col1 = col.Name
	}
	switch col.Type {
	case 1:
		(*mmap)[col1] = icol
	case 2:
		(*mmap)[col1] = scol
	default:
		return false
		//return *env.NewError("wrong column type")
	}
	col.Type = 0
	return true
}

var Builtins_qframe = map[string]*env.Builtin{

	// these builtins accept dialects
	// in general
	//  * 'tag-word is column name
	//  * int, string, ... are values
	//  * ...

	// show the surnames, top down, and number of all jims by surname
	// you could do all this interactively, like in shell :)
	//
	// load %data.csv   // based on the extension this becomes a rye-file-csv kind, so only load is enough
	// 	|filter { 'name = "Jim" }
	// 	|group 'surname
	// 	|aggregate { count => 'cnt }
	// 	|select { 'surname 'cnt }
	// 	|order 'cnt
	// 	|print

	// new-frame { 'col1 1 2 3 'col3 "a" "b" "b" }  returns kind: rye-frame
	// new-frame { 1 "jim" 23 , 2 "jane" 33 }
	// new-frame { 'id 'name 'age , 1 "jim" 23 , 2 "jane" 33 }

	"rye-reader//read-csv": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, qframe.ReadCSV(arg0.(env.Native).Value.(io.Reader)), "rye-frame")
		},
	},

	"new-frame": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// for loop ... if tag-word create slice, if values add to last slice
			//ps.Inj = nil
			// create frame by constructor
			return eval_new_frame(ps, arg0.(env.Block))

		},
	},

	"rye-frame//show": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(arg0.(env.Native).Value)
			return arg0
		},
	},
	// group-by fr 'col1
	"rye-frame//group-by": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ft := arg0.(type) {
			case env.Native:
				switch cn := arg1.(type) {
				case env.Tagword:
					return *env.NewNative(ps.Idx, ft.Value.(qframe.QFrame).GroupBy(groupby.Columns(ps.Idx.GetWord(cn.Index))), "rye-frame-grouper")
				default:
					return MakeBuiltinError(ps, "Tagword type is not found.", "rye-frame//group-by")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-frame//group-by")
			}
		},
	},
	// sort fr 'col1
	"rye-frame//sort": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ft := arg0.(type) {
			case env.Native:
				switch cn := arg1.(type) {
				case env.Tagword:
					return *env.NewNative(ps.Idx, ft.Value.(qframe.QFrame).Sort(qframe.Order{Column: ps.Idx.GetWord(cn.Index)}), "rye-frame")
				default:
					return MakeBuiltinError(ps, "Tagword type is not found.", "rye-frame//sort")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-frame//sort")
			}
		},
	},
	// filter fr { 'col1 > 2 }
	// filter fr { or { 'col1 > 2 , col2 = "a" } }
	"rye-frame//filter": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch frame := arg0.(type) {
			case env.Native:
				switch bl := arg1.(type) {
				case env.Block:
					clauses, err := eval_filter(ps, frame.Value.(qframe.QFrame), bl)
					if err == nil {
						//fmt.Println(clauses)
						if len(clauses) == 1 {
							return *env.NewNative(ps.Idx, frame.Value.(qframe.QFrame).Filter(clauses[0]), "rye-frame")
						} else {
							return MakeBuiltinError(ps, "Only one top-level clause allowed.", "rye-frame//filter")
						}
					} else {
						return *err
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "rye-frame//filter")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-frame//filter")
			}
		},
	},
	// aggregate fr { int-sum 'col2 }
	"rye-frame-grouper//aggregate": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch grouper := arg0.(type) {
			case env.Native:
				switch bl := arg1.(type) {
				case env.Block:
					return eval_aggregate(ps, grouper.Value.(qframe.Grouper), bl)
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "rye-frame-grouper//aggregate")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-frame-grouper//aggregate")
			}
		},
	},
}
