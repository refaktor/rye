//go:build !no_sxml
// +build !no_sxml

package evaldo

import (
	"encoding/xml"

	"github.com/refaktor/rye/env"

	"fmt"
	"io"
)

// { <person> [ .print ] }
// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }

func load_saxml_Dict(ps *env.ProgramState, block env.Block) (env.Dict, *env.Error) {

	var keys []string

	data := make(map[string]any)
	rmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Xword:
			trace5("TAG")
			keys = append(keys, ps.Idx.GetWord(obj1.Index))
			block.Series.Next()
			continue
		case env.Tagword:
			keys = append(keys, "-"+ps.Idx.GetWord(obj1.Index)+"-")
			block.Series.Next()
			continue
		case env.Void:
			keys = append(keys, "")
			block.Series.Next()
			continue
		case env.Block:
			trace5("BLO")
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
				rm, err := load_saxml_Dict(ps, obj1)
				if err != nil {
					return _emptyRM(), err
				}
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = rm
						keys = []string{}
					}
				} else {
					return _emptyRM(), MakeBuiltinError(ps, "No selectors before tag map.", "rye-reader//do-sxml")
				}
			}
		default:
			// ni Dict ampak blok kode, vrni blok
			return _emptyRM(), MakeBuiltinError(ps, "Unknown type in block parsing TODO.", "rye-reader//do-sxml")
		}
	}
	return rmap, nil
}

func do_sxml(ps *env.ProgramState, reader io.Reader, rmap env.Dict) env.Object {

	var stack []env.Dict
	var tags []string
	var curtag string
	decoder := xml.NewDecoder(reader)
	//total := 0
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			trace5("START")
			tag := se.Name.Local
			ob, ok := rmap.Data[tag]
			if ok {
				trace5("START2")
				tags = append(tags, curtag)
				switch obj := ob.(type) {
				case env.Dict:
					trace5("START3")
					stack = append(stack, rmap)
					//fmt.Println(stack)
					rmap = obj
					curtag = tag

					// if new rmap has -start- run it
					b, ok := rmap.Data["-start-"]
					if ok {
						switch obj := b.(type) {
						case env.Block:
							ser := ps.Ser // TODO -- make helper function that "does" a block
							ps.Ser = obj.Series
							EvalBlockInj(ps, *env.NewNative(ps.Idx, se, "rye-sxml-start"), true)
							ps.Ser = ser
						default:
							// TODO Err
						}
					}
				case env.Block:
					stack = append(stack, rmap)
					ser := ps.Ser // TODO -- make helper function that "does" a block
					ps.Ser = obj.Series
					EvalBlockInj(ps, *env.NewNative(ps.Idx, se, "rye-sxml-start"), true)
					ps.Ser = ser
				}
			}
		case xml.CharData:
			ob, ok := rmap.Data[""]
			if ok {
				switch obj := ob.(type) {
				case env.Block:
					ser := ps.Ser // TODO -- make helper function that "does" a block
					ps.Ser = obj.Series
					EvalBlockInj(ps, env.String{string(se.Copy())}, true)
					ps.Ser = ser
				}
			}
		case xml.EndElement:
			inElement := se.Name.Local
			trace5("END")
			//fmt.Println(curtag)
			if inElement == curtag { // TODO -- solve the case of same named elements inside <person><person></person></person>
				trace5("END2")
				//fmt.Println(rmap)
				b, ok := rmap.Data["-end-"]
				if ok {
					switch obj := b.(type) {
					case env.Block:
						ser := ps.Ser // TODO -- make helper function that "does" a block
						ps.Ser = obj.Series
						EvalBlockInj(ps, *env.NewNative(ps.Idx, se, "rye-sxml-start"), true)
						ps.Ser = ser
					default:
						// TODO Err
					}
				}

				//fmt.Println(stack)
				//fmt.Println(tags)
				n := len(stack) - 1 // Top element
				rmap = stack[n]
				stack = stack[:n] // Pop

				m := len(tags) - 1 // Top element
				curtag = tags[m]
				tags = tags[:m] // Pop
			}
		default:
		}
	}
	return nil
}

func trace5(s string) {
	fmt.Println(s)
}

var Builtins_sxml = map[string]*env.Builtin{

	"rye-reader//do-sxml": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_saxml_Dict(ps, arg1.(env.Block))
			//fmt.Println(rm)
			if err != nil {
				ps.FailureFlag = true
				return err
			}
			return do_sxml(ps, arg0.(env.Native).Value.(io.Reader), rm)
		},
	},

	//se.Attr[1].Value
	"rye-sxml-start//get-attr": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					switch n := arg1.(type) {
					case env.Integer:
						if int(n.Value) < len(obj1.Attr) {
							return env.String{obj1.Attr[int(n.Value)].Value}
						} else {
							return env.Void{}
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "rye-sxml-start//get-attr")
					}
				default:
					return MakeBuiltinError(ps, "Not xml-strat element.", "rye-sxml-start//get-attr")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-sxml-start//get-attr")
			}
		},
	},
	"rye-sxml-start//name?": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					return env.String{obj1.Name.Local}
				default:
					return MakeBuiltinError(ps, "Not xml-strat element.", "rye-sxml-start//name?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-sxml-start//name?")
			}
		},
	},
}
