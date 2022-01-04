package evaldo

import (
	"fmt"
	"rye/env"
)

// { <key> [ .print ] }
// { <key> { <more> [ .print ] } }
// { <key> { _ [ .print ] } }
// { <key> <token> [ .print ] }

// both kinds of blocks for a key
// { <person> [ .print ] { <name> print } }

// cpath for traversing deeper into the structure
// { people/author { <name> <surname> [ .print ] } }

// cpath for traversing deeper into the structure
// { people/author { <name> <surname> keyval [ .collect-kv ] } }

// { some { <person> k,v { [1] key , [2] val } } }

// { _ { <person> { * [ -> 1 |print , -> 2 |print ] } } }

func load_structures_Dict(es *env.ProgramState, block env.Block) (env.Dict, *env.Error) {

	var keys []string

	data := make(map[string]interface{})
	rmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Xword:
			// trace5("TAG")
			keys = append(keys, es.Idx.GetWord(obj1.Index))
			block.Series.Next()
			continue
		case env.Tagword:
			keys = append(keys, "-"+es.Idx.GetWord(obj1.Index)+"-")
			block.Series.Next()
			continue
		case env.Void:
			keys = append(keys, "")
			block.Series.Next()
			continue
		case env.Block:
			// trace5("BLO")
			block.Series.Next()
			if obj1.Mode == 1 {
				// if code assign in to keys in Dict
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = obj1
						keys = []string{}
					}
				} else {
					rmap.Data["-start-"] = obj1
				}
			} else if obj1.Mode == 0 {
				rm, err := load_saxml_Dict(es, obj1)
				if err != nil {
					return _emptyRM(), err
				}
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = rm
						keys = []string{}
					}
				} else {
					return _emptyRM(), env.NewError("no selectors before tag map")
				}
			}
		default:
			// ni Dict ampak blok kode, vrni blok
			return _emptyRM(), env.NewError("unknow type in block parsing TODO")
		}
	}
	return rmap, nil
}

func do_structures(es *env.ProgramState, data env.Dict, rmap env.Dict) env.Object { // TODO -- make it work for List too later
	fmt.Println(rmap)
	// fmt.Println("IN DO")
	//	var stack []env.Dict
	for key, val := range data.Data {
		// fmt.Println(key)
		rval0, ok0 := rmap.Data[""]
		if ok0 {
			// trace5("ANY FOUND")
			switch obj := rval0.(type) {
			case env.Dict:
				switch val1 := val.(type) {
				case map[string]interface{}:
					// trace5("RECURSING")
					do_structures(es, *env.NewDict(val1), obj)
					// trace5("OUTCURSING")
				}
			case env.Block:
				//				stack = append(stack, rmap)
				ser := es.Ser // TODO -- make helper function that "does" a block
				es.Ser = obj.Series
				EvalBlockInj(es, JsonToRye(val), true)
				es.Ser = ser
			}
		}
		rval, ok := rmap.Data[key]
		if ok {
			// fmt.Println("found")
			switch obj := rval.(type) {
			case env.Dict:
				switch val1 := val.(type) {
				case map[string]interface{}:
					do_structures(es, *env.NewDict(val1), obj)
				}
			case env.Block:
				//				stack = append(stack, rmap)
				ser := es.Ser // TODO -- make helper function that "does" a block
				es.Ser = obj.Series
				EvalBlockInj(es, JsonToRye(val), true)
				es.Ser = ser
			}

		} else {
			// 	fmt.Println("not found")
		}
	}
	return nil
}

var Builtins_structures = map[string]*env.Builtin{

	"process": {
		Argsn: 2,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_structures_Dict(es, arg1.(env.Block))
			//fmt.Println(rm)
			if err != nil {
				es.FailureFlag = true
				return err
			}
			switch data := arg0.(type) {
			case env.Dict:
				return do_structures(es, data, rm)
			}
			return nil
		},
	},
}
