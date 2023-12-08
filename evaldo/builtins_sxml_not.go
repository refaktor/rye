//go:build !b_sxml
// +build !b_sxml

package evaldo

import (
	"rye/env"
)

//

var Builtins_sxml = map[string]*env.Builtin{}

func load_saxml_Dict(es *env.ProgramState, block env.Block) (env.Dict, *env.Error) {
	var keys []string

	data := make(map[string]interface{})
	rmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Xword:
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
