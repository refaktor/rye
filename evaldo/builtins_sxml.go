package evaldo

import (
	"../env"
	"encoding/xml"

	//"fmt"
	"io"
	"strings"
	//	"bufio"
	//	"io"
	//	"os"
	//	"strconv"
	//	"strings"
)

func _emptyRM() env.RawMap {
	return env.RawMap{}
}

// { <person> [ .print ] }
// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }

func load_saxml_rawmap(es *env.ProgramState, block env.Block) (env.RawMap, *env.Error) {

	var keys []string

	data := make(map[string]interface{})
	rmap := *env.NewRawMap(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Xword:
			trace4("TAG")
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
			trace4("BLO")
			block.Series.Next()
			if obj1.Mode == 1 {
				// if code assign in to keys in rawmap
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = obj1
						keys = []string{}
					}
				} else {
					rmap.Data["-start-"] = obj1
				}
			} else if obj1.Mode == 0 {
				rm, err := load_saxml_rawmap(es, obj1)
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
			// ni rawmap ampak blok kode, vrni blok
			return _emptyRM(), env.NewError("unknow type in block parsing TODO")
		}
	}
	return rmap, nil
}

func do_sxml(es *env.ProgramState, reader io.Reader, rmap env.RawMap) env.Object {

	var stack []env.RawMap
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
			trace4("START")
			tag := se.Name.Local
			ob, ok := rmap.Data[tag]
			if ok {
				trace4("START2")
				tags = append(tags, curtag)
				switch obj := ob.(type) {
				case env.RawMap:
					trace4("START3")
					stack = append(stack, rmap)
					//fmt.Println(stack)
					rmap = obj
					curtag = tag

					// if new rmap has -start- run it
					b, ok := rmap.Data["-start-"]
					if ok {
						switch obj := b.(type) {
						case env.Block:
							ser := es.Ser // TODO -- make helper function that "does" a block
							es.Ser = obj.Series
							EvalBlockInj(es, *env.NewNative(es.Idx, se, "rye-sxml-start"), true)
							es.Ser = ser
						default:
							// TODO Err
						}
					}
				case env.Block:
					stack = append(stack, rmap)
					ser := es.Ser // TODO -- make helper function that "does" a block
					es.Ser = obj.Series
					EvalBlockInj(es, *env.NewNative(es.Idx, se, "rye-sxml-start"), true)
					es.Ser = ser
				}
			}
		case xml.CharData:
			ob, ok := rmap.Data[""]
			if ok {
				switch obj := ob.(type) {
				case env.Block:
					ser := es.Ser // TODO -- make helper function that "does" a block
					es.Ser = obj.Series
					EvalBlockInj(es, env.String{string(se.Copy())}, true)
					es.Ser = ser
				}
			}
		case xml.EndElement:
			inElement := se.Name.Local
			trace4("END")
			//fmt.Println(curtag)
			if inElement == curtag { // TODO -- solve the case of same named elements inside <person><person></person></person>
				trace4("END2")
				//fmt.Println(rmap)
				b, ok := rmap.Data["-end-"]
				if ok {
					switch obj := b.(type) {
					case env.Block:
						ser := es.Ser // TODO -- make helper function that "does" a block
						es.Ser = obj.Series
						EvalBlockInj(es, *env.NewNative(es.Idx, se, "rye-sxml-start"), true)
						es.Ser = ser
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

func trace4(s string) {
	//	fmt.Println(s)
}

var Builtins_sxml = map[string]*env.Builtin{

	"string-reader": {
		Argsn: 1,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(es.Idx, strings.NewReader(arg0.(env.String).Value), "rye-reader")
		},
	},

	"rye-reader//do-sxml": {
		Argsn: 2,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_saxml_rawmap(es, arg1.(env.Block))
			//fmt.Println(rm)
			if err != nil {
				es.FailureFlag = true
				return err
			}
			return do_sxml(es, arg0.(env.Native).Value.(io.Reader), rm)
		},
	},

	//se.Attr[1].Value
	"rye-sxml-start//get-attr": {
		Argsn: 2,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
						return env.NewError("second arg not integer")
					}
				default:
					return env.NewError("Not xml-strat element")
				}
			default:
				return env.NewError("first argument should be native")
			}
		},
	},
	"rye-sxml-start//name?": {
		Argsn: 1,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					return env.String{obj1.Name.Local}
				default:
					return env.NewError("Not xml-strat element")
				}
			default:
				return env.NewError("first argument should be native")
			}
		},
	},
}
