//__ +build b_html

package evaldo

import (
	"rye/env"

	"golang.org/x/net/html"

	"fmt"
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

type HtmlNavigCondition struct {
	Class string
	Id    string
}

type HtmlDialectNode struct {
	Condition *HtmlNavigCondition
	Value     env.Block
}

// tagname is stored in dict that we lookup. On match we look at condition if there.

// tagname could be wildcard in this case we look at condition each time

// xwords after another or conditions mean that we dive deeper, like a block { } would be

func load_html_Dict(es *env.ProgramState, block env.Block) (env.Dict, *env.Error) {

	var key string
	var condition *HtmlNavigCondition
	var keys []string                    // keys
	var conditions []*HtmlNavigCondition // conditions

	data := make(map[string]interface{}) //
	rmap := *env.NewDict(data)           //

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Void:
			trace4("VOID")
			keys = append(keys, "")              // if void append empty string to keys
			conditions = append(conditions, nil) // if void append empty string to keys
			block.Series.Next()
			continue
		case env.Xword:
			trace4("XWORD")
			key = es.Idx.GetWord(obj1.Index)
			// keys = append(keys, ) // if xword append to keys
			block.Series.Next()
			continue
		case env.Tagword: // if tagword append -word- to keys
			trace4("TAGWORD")
			key = "-" + es.Idx.GetWord(obj1.Index) + "-"
			//keys = append(keys, "-"+es.Idx.GetWord(obj1.Index)+"-")
			block.Series.Next()
			continue
		case env.LSetword: // if tagword append -word- to keys
			if condition == nil {
				condition = &HtmlNavigCondition{"", ""}
			}
			condition.Id = es.Idx.GetWord(obj1.Index)
			//keys = append(keys, "-"+es.Idx.GetWord(obj1.Index)+"-")
			block.Series.Next()
			continue
		case env.Opword: // if tagword append -word- to keys
			if condition == nil {
				condition = &HtmlNavigCondition{"", ""}
			}
			condition.Class = es.Idx.GetWord(obj1.Index)
			//keys = append(keys, "-"+es.Idx.GetWord(obj1.Index)+"-")
			block.Series.Next()
			continue
		case env.Comma:
			trace4("COMMA")
			keys = append(keys, key)
			conditions = append(conditions, condition) // if void append empty string to keys
			key = ""
			condition = nil
			block.Series.Next()
			continue
		case env.Block:
			trace4("BLOCK")
			block.Series.Next()
			if obj1.Mode == 1 {
				// if code assign in to keys in Dict
				if len(keys) > 0 {
					for idx, k := range keys {
						if conditions[idx] != nil {
							rmap.Data[k] = HtmlDialectNode{conditions[idx], obj1}
						} else {
							rmap.Data[k] = obj1
						}
						keys = []string{}
					}
				} else {
					rmap.Data["-start-"] = obj1
				}
			} else if obj1.Mode == 0 {
				rm, err := load_html_Dict(es, obj1)
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

func do_html(es *env.ProgramState, reader io.Reader, rmap env.Dict) env.Object {

	var stack []env.Dict // ??_?
	var tags []string    // list of tags ??_?
	var curtag string    // current tag ??_?
	decoder := html.NewTokenizer(reader)
	//total := 0
	for {
		// Read tokens from the XML document in a stream.
		t := decoder.Next()
		// Inspect the type of the token just read.
		switch t {
		case html.StartTagToken:
			trace4("START TOKEN")
			tok := decoder.Token()
			tag := tok.Data
			// check if token exists in current rmap
			ob, ok := rmap.Data[tag]
			if ok {
				// if it is then
				trace4("is in RMAP")
				tags = append(tags, curtag) // append current tag to tags
				switch obj := ob.(type) {   // swith on the type of object from rmap
				case env.Dict: // if it is Dict
					trace4("THIS IS DICT OBJ")
					stack = append(stack, rmap) // append whole rmap to the stack
					//fmt.Println(stack)
					rmap = obj   // set the received object as rmap
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
				case HtmlDialectNode: // if it is Dict
					trace4("THIS IS DIALECT NODE")
					blk := obj.Value // set the received object as rmap
					if obj.Condition.Class == "title" {
						stack = append(stack, rmap) // append rmap to stack and evaluate block
						ser := es.Ser               // TODO -- make helper function that "does" a block
						es.Ser = blk.Series
						EvalBlockInj(es, *env.NewNative(es.Idx, tok, "rye-html-start"), true)
						es.Ser = ser
					}

				case env.Block: // if it is block
					trace4("THIS IS BLOCK OBJ")
					stack = append(stack, rmap) // append rmap to stack and evaluate block
					ser := es.Ser               // TODO -- make helper function that "does" a block
					es.Ser = obj.Series
					EvalBlockInj(es, *env.NewNative(es.Idx, tok, "rye-html-start"), true)
					es.Ser = ser
				}
			}
		case html.TextToken:
			tok := decoder.Token()
			ob, ok := rmap.Data[""]
			if ok {
				switch obj := ob.(type) {
				case env.Block:
					ser := es.Ser // TODO -- make helper function that "does" a block
					es.Ser = obj.Series
					EvalBlockInj(es, env.String{string(tok.Data)}, true)
					es.Ser = ser
				}
			}
		case html.EndTagToken:
			tok := decoder.Token()
			inElement := tok.Data
			// inElement := se.Name.Local
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
						EvalBlockInj(es, *env.NewNative(es.Idx, tok, "rye-html-start"), true)
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
		case html.ErrorToken:
			return nil
		default:
		}
	}
	return nil
}

func trace4(s string) {
	//	fmt.Println(s)
}

var Builtins_html = map[string]*env.Builtin{

	"rye-reader//do-html": {
		Argsn: 2,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_html_Dict(es, arg1.(env.Block))
			fmt.Println(rm)
			if err != nil {
				es.FailureFlag = true
				return err
			}
			return do_html(es, arg0.(env.Native).Value.(io.Reader), rm)
		},
	},

	/*	"rye-html-start//attr?": {
			Argsn: 2,
			Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch obj := arg0.(type) {
				case env.Native:
					switch obj1 := obj.Value.(type) {
					case html.StartTagToken:
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

		"rye-html-start//name?": {
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
		}, */
}
