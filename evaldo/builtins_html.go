//__ +build b_html

package evaldo

import (
	"fmt"
	"rye/env"

	"golang.org/x/net/html"

	//"fmt"
	"io"
	// "strings"
	//	"bufio"
	//	"io"
	//	"os"
	//	"strconv"
	//	"strings"
)

// { <div> .menu <a> { .content .print } }
// { <div> .menu <a> , .footer { .content .print } }
// { <div> .menu <a> [ .content .print ] <footer> { <a> .external [ .to-upper print ] }

// { <div> .menu <a> when [ [.attr? "href" |includes "www.google.com" ] { <b> [ .content? .collect ] } }

// when we have accessor words
// { <div> .menu <a> when [ [href] .includes "www.google.com" ] { <b> [ [content] .collect ] } }

const TYPE_SUBNODE int = 1
const TYPE_CODE int = 2

type HtmlNavigCondition struct {
	Class string
	Id    string
}

type HtmlDialectNode struct {
	Condition *HtmlNavigCondition
	Code      *env.Block
	SubNode   *env.Dict
	Type      int // 1 - node , 2 - code , 3 - direct_descendant (later)
}

// dmap is a dict of nodes

// tagname is stored in dict that we lookup. On match we look at condition if there.

// tagname could be wildcard in this case we look at condition each time

// xwords after another or conditions mean that we dive deeper, like a block { } would be

func load_html_Dict(es *env.ProgramState, block env.Block) (env.Dict, *env.Error) {

	var key string
	var condition *HtmlNavigCondition
	var keys []string                    // keys
	var conditions []*HtmlNavigCondition // conditions

	data := make(map[string]interface{})
	dmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		object := block.Series.Peek()
		switch obj := object.(type) {
		case env.Void:
			keys = append(keys, "")              // if void append empty string to keys
			conditions = append(conditions, nil) // if void append empty string to keys
			block.Series.Next()
			continue
		case env.Xword:
			key = es.Idx.GetWord(obj.Index)
			block.Series.Next()
			continue
		case env.Tagword: // if tagword append -word- to keys
			key = "-" + es.Idx.GetWord(obj.Index) + "-"
			block.Series.Next()
			continue
		case env.LSetword: // if tagword append -word- to keys
			if condition == nil {
				condition = &HtmlNavigCondition{"", ""}
			}
			condition.Id = es.Idx.GetWord(obj.Index)
			block.Series.Next()
			continue
		case env.Opword: // if tagword append -word- to keys
			if condition == nil {
				condition = &HtmlNavigCondition{"", ""}
			}
			condition.Class = es.Idx.GetWord(obj.Index)
			block.Series.Next()
			continue
		case env.Comma:
			trace4("COMMA")
			// push key and condition
			keys = append(keys, key)
			conditions = append(conditions, condition)
			key = ""
			condition = nil
			block.Series.Next()
			continue
		case env.Block:
			if key != "" || condition != nil {
				keys = append(keys, key)
				conditions = append(conditions, condition) // if void append empty string to keys
				key = ""
				condition = nil
			}
			trace8("KEYS, CONDITIONS")
			// fmt.Println(keys)
			// fmt.Println(conditions)
			block.Series.Next()
			if obj.Mode == 1 { // if block is [ ]
				// if code assign in to keys in Dict
				if len(keys) > 0 {
					for idx, k := range keys {
						dmap.Data[k] = HtmlDialectNode{conditions[idx], &obj, nil, TYPE_CODE}
					}
					keys = make([]string, 0)
					conditions = make([]*HtmlNavigCondition, 0)
				} else {
					dmap.Data["-start-"] = obj
				}
			} else if obj.Mode == 0 { // if block is { }
				subdmap, err := load_html_Dict(es, obj)
				if err != nil {
					return _emptyRM(), err
				}
				trace8(" **** FOR KEYS")
				if len(keys) > 0 {
					// fmt.Println(keys)
					// fmt.Println(subdmap)
					for idx, k := range keys {
						dmap.Data[k] = HtmlDialectNode{conditions[idx], nil, &subdmap, TYPE_SUBNODE}
					}
					keys = make([]string, 0)
					conditions = make([]*HtmlNavigCondition, 0)
				} else {
					return _emptyRM(), env.NewError("no selectors before tag map")
				}
			}
		default:
			// ni Dict ampak blok kode, vrni blok
			return _emptyRM(), env.NewError("unknow type in block parsing TODO")
		}
	}
	trace8("**** RETURNING **** ")
	// fmt.Println(dmap)

	return dmap, nil
}

func do_html(es *env.ProgramState, reader io.Reader, dmap env.Dict) env.Object {

	trace8("**** DO HTML **** ")
	// fmt.Println(dmap)

	var stack []env.Dict // ??_?
	var tags []string    // list of tags ??_?
	var curtag string    // current tag ??_?
	decoder := html.NewTokenizer(reader)

	// check of root node has start (code node)
	b, ok := dmap.Data["-start-"]
	if ok {
		switch obj := b.(type) {
		case env.Block: // if values of start key is of type block, evaluate it
			ser := es.Ser // TODO -- make helper function that "does" a block
			es.Ser = obj.Series
			EvalBlockInj(es, *env.NewNative(es.Idx, "", "rye-html-start"), true)
			if es.ErrorFlag {
				return es.Res
			}
			es.Ser = ser
		default:
			// TODO Err
		}
	}

myloop:
	for {

		//fmt.Println(dmap)

		// Read tokens from the XML document in a stream.
		rawtoken := decoder.Next()
		// Inspect the type of the token just read.
		switch rawtoken {
		case html.StartTagToken:
			trace8("START TOKEN (tag)")
			token := decoder.Token()
			tag := token.Data
			trace8(tag)
			// check if token exists in current dmap
			rawNode, ok := dmap.Data[tag]
			if ok {
				// if it is then
				trace8(" --- is in RMAP")
				switch node := rawNode.(type) { // swith on the type of object from rmap
				case HtmlDialectNode: // if it is Dict

					id := ""
					class := ""
					// get the class and id attributes
					for _, a := range token.Attr {
						if a.Key == "id" {
							id = a.Val
						} else if a.Key == "class" {
							class = a.Val
						}
					}
					trace8("NODE, CONDITION")
					// fmt.Println(node)
					// fmt.Println(node.Condition)
					if (node.Condition == nil) || ((node.Condition.Class == "" || node.Condition.Class == class) &&
						(node.Condition.Id == "" || node.Condition.Id == id)) {
						switch node.Type {
						case TYPE_CODE: // if it is block
							trace4("  THIS IS BLOCK OBJ")
							// stack = append(stack, rmap) // append rmap to stack and evaluate block
							ser := es.Ser // TODO -- make helper function that "does" a block
							es.Ser = node.Code.Series
							EvalBlockInj(es, *env.NewNative(es.Idx, token, "rye-html-start"), true)
							if es.ErrorFlag {
								return es.Res
							}
							es.Ser = ser

						case TYPE_SUBNODE:
							trace8(" --- THIS IS DICT OBJ")
							tags = append(tags, curtag) // append current tag to tags
							stack = append(stack, dmap) // append whole rmap to the stack
							//fmt.Println(stack)
							dmap = *node.SubNode // set the received object as rmap
							curtag = tag         // set tag as curtag
							// if new rmap has -start- run it
							rawCode, ok := node.SubNode.Data["-start-"]
							if ok {
								switch code := rawCode.(type) {
								case env.Block: // if values of start key is of type block, evaluate it
									ser := es.Ser // TODO -- make helper function that "does" a block
									es.Ser = code.Series
									EvalBlockInj(es, *env.NewNative(es.Idx, token, "rye-html-start"), true)
									if es.ErrorFlag {
										return es.Res
									}
									es.Ser = ser
								default:
									// TODO Err
								}
							}
						}
						/*
							case HtmlDialectNode: // if it is Dict
								trace4("  THIS IS DIALECT NODE")
								blk := obj.Value // set the received object as rmap

								id := ""
								class := ""
								// get the class and id attributes
								for _, a := range token.Attr {
									if a.Key == "id" {
										id = a.Val
									} else if a.Key == "class" {
										class = a.Val
									}
								}
								if (obj.Condition.Class == "" || obj.Condition.Class == class) &&
									(obj.Condition.Id == "" || obj.Condition.Id == id) {
									switch blk2 := blk.(type) { // swith on the type of object from rmap
									case env.Dict: // if it is Dict
										trace8(" --- THIS IS DICT OBJ")
										tags = append(tags, curtag) // append current tag to tags
										stack = append(stack, rmap) // append whole rmap to the stack
										//fmt.Println(stack)
										rmap = blk2  // set the received object as rmap
										curtag = tag // set tag as curtag
										// if new rmap has -start- run it
										b, ok := rmap.Data["-start-"]
										if ok {
											switch obj := b.(type) {
											case env.Block: // if values of start key is of type block, evaluate it
												ser := es.Ser // TODO -- make helper function that "does" a block
												es.Ser = obj.Series
												EvalBlockInj(es, *env.NewNative(es.Idx, tok, "rye-html-start"), true)
												es.Ser = ser
											default:
												// TODO Err
											}
										}
									case env.Block:
										ser := es.Ser // TODO -- make helper function that "does" a block
										es.Ser = blk2.Series
										EvalBlockInj(es, *env.NewNative(es.Idx, tok, "rye-html-start"), true)
										es.Ser = ser
									}
								} */
					}
				default:
					// TODO err not HtmlNode
				}
			}
		case html.TextToken:
			trace8("TEXT TOKEN")
			// fmt.Println(dmap)
			tok := decoder.Token()
			rawnode, ok := dmap.Data[""]
			if ok {
				// fmt.Println("IN IF")
				// fmt.Println(rawnode)
				switch node := rawnode.(type) {
				case HtmlDialectNode:
					// fmt.Println("IN BLOCK")
					ser := es.Ser // TODO -- make helper function that "does" a block
					es.Ser = node.Code.Series
					EvalBlockInj(es, env.String{string(tok.Data)}, true)
					if es.ErrorFlag {
						return es.Res
					}
					es.Ser = ser
				}
			}
		case html.EndTagToken:
			tok := decoder.Token()
			inElement := tok.Data
			// inElement := se.Name.Local
			trace8("END TAG (elem, curtag)")
			trace8(inElement)
			trace8(curtag)
			if inElement == curtag { // TODO -- solve the case of same named elements inside <person><person></person></person>
				trace4("END2")
				//fmt.Println(rmap)
				b, ok := dmap.Data["-end-"]
				if ok {
					switch obj := b.(type) {
					case env.Block:
						ser := es.Ser // TODO -- make helper function that "does" a block
						es.Ser = obj.Series
						EvalBlockInj(es, *env.NewNative(es.Idx, tok, "rye-html-start"), true)
						if es.ErrorFlag {
							return es.Res
						}
						es.Ser = ser
					default:
						// TODO Err
					}
				}

				trace8("*** RETURNING ***")
				// fmt.Println(stack)
				// fmt.Println(tags)
				n := len(stack) - 1 // Top element
				// fmt.Println(n)
				dmap = stack[n]
				stack = stack[:n] // Pop

				m := len(tags) - 1 // Top element
				curtag = tags[m]
				tags = tags[:m] // Pop

				// fmt.Println(stack)
				// fmt.Println(tags)

			}
		case html.ErrorToken:
			break myloop
		default:
		}
	}
	// fmt.Println(dmap)
	b1, ok1 := dmap.Data["-return-"]
	if ok1 {
		switch obj := b1.(type) {
		case HtmlDialectNode: // if values of start key is of type block, evaluate it
			ser := es.Ser // TODO -- make helper function that "does" a block
			es.Ser = obj.Code.Series
			EvalBlockInj(es, *env.NewNative(es.Idx, "", "rye-html-start"), true)
			if es.ErrorFlag {
				return es.Res
			}
			es.Ser = ser
			return es.Res
		default:
			// TODO Err
		}
	}
	return nil
}

func trace4(s string) {
	// fmt.Println(s)
}
func trace8(s string) {
	if false {
		fmt.Println(s)
	}
}

var Builtins_html = map[string]*env.Builtin{

	"rye-reader//parse-html": {
		Argsn: 2,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_html_Dict(es, arg1.(env.Block))
			trace8("*** _--- GOT RM ++**")
			// fmt.Println(rm)
			if err != nil {
				es.FailureFlag = true
				return err
			}
			return do_html(es, arg0.(env.Native).Value.(io.Reader), rm)
		},
	},

	"rye-html-start//attr?": {
		Argsn: 2,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tok1 := arg0.(type) {
			case env.Native:
				switch tok := tok1.Value.(type) {
				case html.Token:
					switch n := arg1.(type) {
					case env.Integer:
						if int(n.Value) < len(tok.Attr) {
							return env.String{tok.Attr[int(n.Value)].Val}
						} else {
							return env.Void{}
						}
					case env.Tagword:
						for _, a := range tok.Attr {
							if a.Key == es.Idx.GetWord(n.Index) {
								return env.String{a.Val}
							}
						}
						return env.Void{}
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

	"rye-html-start//name?": {
		Argsn: 1,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tok1 := arg0.(type) {
			case env.Native:
				switch tok := tok1.Value.(type) {
				case html.Token:
					return env.String{tok.Data}
				default:
					return env.NewError("Not xml-strat element")
				}
			default:
				return env.NewError("first argument should be native")
			}
		},
	},

	/*	"rye-html-start//name?": {
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
	},*/
}
