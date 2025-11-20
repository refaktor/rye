//go:build !no_sxml
// +build !no_sxml

package evaldo

import (
	"encoding/xml"

	"github.com/refaktor/rye/env"

	"io"
)

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
					return EmptyRM(), err
				}
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = rm
						keys = []string{}
					}
				} else {
					return EmptyRM(), MakeBuiltinError(ps, "No selectors before tag map.", "reader//do-sxml")
				}
			}
		default:
			// ni Dict ampak blok kode, vrni blok
			return EmptyRM(), MakeBuiltinError(ps, "Unknown type in block parsing TODO.", "reader//do-sxml")
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
			// if !ok {
			//	ob, ok = rmap.Data["-any-"]
			// }
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
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
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
					if ps.ErrorFlag {
						ps.Ser = ser
						return ps.Res
					}
					ps.Ser = ser
				}
			} else {
				inElement := se.Name.Local
				trace5("END")
				//fmt.Println(curtag)
				if true || inElement == curtag { // TODO -- solve the case of same named elements inside <person><person></person></person>
					trace5("END2")
					//fmt.Println(rmap)
					b, ok := rmap.Data["-any-"]
					if ok {
						switch obj := b.(type) {
						case env.Block:
							ser := ps.Ser // TODO -- make helper function that "does" a block
							ps.Ser = obj.Series
							EvalBlockInj(ps, *env.NewNative(ps.Idx, se, "rye-sxml-start"), true)
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							ps.Ser = ser
						default:
							// TODO Err
						}
					}

					//fmt.Println(stack)
					//fmt.Println(tags)
					// n := len(stack) - 1 // Top element
					// rmap = stack[n]
					// stack = stack[:n] // Pop

					// m := len(tags) - 1 // Top element
					// curtag = tags[m]
					// tags = tags[:m] // Pop
				}
			}
		case xml.CharData:
			ob, ok := rmap.Data[""]
			if ok {
				switch obj := ob.(type) {
				case env.Block:
					ser := ps.Ser // TODO -- make helper function that "does" a block
					ps.Ser = obj.Series
					EvalBlockInj(ps, *env.NewString(string(se.Copy())), true)
					if ps.ErrorFlag {
						ps.Ser = ser
						return ps.Res
					}
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
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
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
	// fmt.Println(s)
}

var Builtins_sxml = map[string]*env.Builtin{

	//
	// ##### SXML ##### "Streaming, SAX-like XML processing"
	//
	// { <person> [ .print ] }
	// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }
	//
	// Tests:
	// stdout {
	//   "<scene><bot>C3PO</bot><bot>R2D2</bot><jedi>Luke</jedi></scene>" |reader
	//   .do-sxml { _ [ .prns ] }
	// } "C3PO R2D2 Luke "
	// stdout {
	//   "<scene><bot>C3PO</bot><bot>R2D2</bot><jedi>Luke</jedi></scene>" |reader
	//   .do-sxml { <bot> { _ [ .prns ] } }
	// } "C3PO R2D2 "
	// stdout {
	//   "<scene><ship>XWing</ship><bot>R2D2</bot><jedi>Luke</jedi></scene>" |reader
	//   .do-sxml { <bot> <jedi> { _ [ .prns ] } }
	// } "R2D2 Luke "
	// stdout {
	//   "<scene><xwing><bot>R2D2</bot><person>Luke</person></xwing><destroyer><person>Vader</person></destroyer></scene>" |reader
	//   .do-sxml { <xwing> { <person> { _ [ .prns ] } } }
	// } "Luke "
	// Args:
	// * reader: XML reader object
	// * block: SXML processing block with tag handlers
	// Returns:
	// * result of processing the XML
	"reader//do-sxml": {
		Argsn: 2,
		Doc:   "Processes XML using a streaming SAX-like approach with tag handlers.",
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

	// Tests:
	// stdout {
	//   `<scene><ship type="xwing"><person age="25">Luke</person></ship><ship type="destroyer"><person age="55">Vader</person></ship></scene>` |reader
	//   .do-sxml { <ship> [ .attr? 0 |last |prns	 ] }
	// } "xwing destroyer "
	// stdout {
	//   `<scene><ship type="xwing"><person age="25">Luke</person></ship><ship type="destroyer"><person age="55">Vader</person></ship></scene>` |reader
	//   .do-sxml { <person> [ .attr? 0 |last |prns	 ] }
	// } "25 55 "
	// Args:
	// * element: XML start element
	// * index: Integer index of the attribute to retrieve
	// Returns:
	// * list [ namespace tag value ] of the attribute or void if not found
	"rye-sxml-start//attr?": {
		Argsn: 2,
		Doc:   "Retrieves an attribute by index from an XML start element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					switch n := arg1.(type) {
					case env.Integer:
						if int(n.Value) < len(obj1.Attr) {
							attr := obj1.Attr[int(n.Value)]
							return *env.NewList([]any{
								*env.NewString(attr.Name.Space),
								*env.NewString(attr.Name.Local),
								*env.NewString(attr.Value),
							})
						} else {
							return env.Void{}
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "rye-sxml-start//get-attr")
					}
				default:
					return MakeBuiltinError(ps, "Not xml-start element.", "rye-sxml-start//get-attr")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-sxml-start//get-attr")
			}
		},
	},
	// TODO:
	// stdout {
	//   "<scene><xwing><bot>R2D2</bot><person><name>Luke</name></person></xwing><destroyer><person>Vader</person></destroyer></scene>" |reader
	//   .do-sxml { <xwing> { 'start [ prns "YYY" ] <bot> [ print "***" ] 'any [ .name? .probe ] 'end [ print "xx" ] } }
	// } "bot R2D2 \nperson name Luke \n"
	// Args:
	// * element: XML start element
	// Returns:
	// * string name of the XML element
	"rye-sxml-start//name?": {
		Argsn: 1,
		Doc:   "Returns the name of an XML start element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					return *env.NewString(obj1.Name.Local)
				default:
					return MakeBuiltinError(ps, "Not xml-start element.", "rye-sxml-start//name?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-sxml-start//name?")
			}
		},
	},
}
