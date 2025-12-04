//go:build !no_html
// +build !no_html

package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"

	"golang.org/x/net/html"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"

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

// { <div> .menu <a> when [ [.Attr? "href" |includes "www.google.com" ] { <b> [ .content? .collect ] } }

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

	data := make(map[string]any)
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
		case env.Word: // if tagword append -word- to keys
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
					return EmptyRM(), err
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
					return EmptyRM(), env.NewError("no selectors before tag map")
				}
			}
		default:
			// ni Dict ampak blok kode, vrni blok
			return EmptyRM(), env.NewError("unknow type in block parsing TODO")
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
			MaybeDisplayFailureOrError(es, es.Idx, "do_html")

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
							MaybeDisplayFailureOrError(es, es.Idx, "do_html\\code")

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
									MaybeDisplayFailureOrError(es, es.Idx, "do_html\\subnode")
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
					EvalBlockInj(es, *env.NewString(tok.Data), true)
					MaybeDisplayFailureOrError(es, es.Idx, "do_html\\text-token")
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
						MaybeDisplayFailureOrError(es, es.Idx, "do_html\\end-tag-token")

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
			MaybeDisplayFailureOrError(es, es.Idx, "do_html\\return")

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

	//
	// ##### HTML ##### "HTML processing functions"
	//
	// Tests:
	// equal { unescape\html "&gt;hello&lt;" } ">hello<"
	// Args:
	// * text: HTML-escaped string
	// Returns:
	// * string with HTML entities converted to their character equivalents
	"unescape\\html": {
		Argsn: 1,
		Doc:   "Converts HTML entities to their character equivalents.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "unescape\\html")
			}
			mkd := html.UnescapeString(text.Value)
			return *env.NewString(mkd)
		},
	},

	// Tests:
	// equal { escape\html "<hello>" } "&lt;hello&gt;"
	// Args:
	// * text: String containing HTML special characters
	// Returns:
	// * string with special characters converted to HTML entities
	"escape\\html": {
		Argsn: 1,
		Doc:   "Converts special characters to HTML entities.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "unescape\\html")
			}
			mkd := html.EscapeString(text.Value)
			return *env.NewString(mkd)
		},
	},

	// Tests:
	// equal { html->markdown "<h1>title</h1><p>para</p>" } "# title\n\npara"
	// Args:
	// * html: HTML string to convert
	// Returns:
	// * string containing markdown equivalent of the HTML
	"html->markdown": {
		Argsn: 1,
		Doc:   "Converts HTML text to markdown format.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			text, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "html->markdown")
			}
			mkd, err := htmltomarkdown.ConvertString(text.Value)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "html->markdown")
			}
			return *env.NewString(mkd)
		},
	},

	// Tests:
	// stdout { "<html><body><div class='menu'><a href='/'>home</a><a href='/about/'>about</a>" |reader
	//   .Parse-html { <a> [ .Attr? 'href |prns ] }
	// } "/ /about/ "
	// Args:
	// * reader: HTML reader object
	// * block: HTML processing block with tag handlers
	// Returns:
	// * result of processing the HTML
	"reader//Parse-html": {
		Argsn: 2,
		Doc:   "Parses HTML using a streaming approach with tag handlers.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_html_Dict(ps, arg1.(env.Block))
			trace8("*** _--- GOT RM ++**")
			// fmt.Println(rm)
			if err != nil {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Error to load html dict.", "reader//parse-html")
			}
			return do_html(ps, arg0.(env.Native).Value.(io.Reader), rm)
		},
	},

	// Tests:
	// stdout { "<div class='menu' id='nav'></div>" |reader .Parse-html { <div> [ .Attr? 'class |prn ] } } "menu"
	// stdout { "<div class='menu' id='nav'></div>" |reader .Parse-html { <div> [ .Attr? 'id |prn ] } } "nav"
	// Args:
	// * element: HTML token element
	// * name-or-index: Attribute name (as word or string) or index (as integer)
	// Returns:
	// * string value of the attribute or void if not found
	"rye-html-start//Attr?": {
		Argsn: 2,
		Doc:   "Retrieves an attribute value by name or index from an HTML element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tok1 := arg0.(type) {
			case env.Native:
				switch tok := tok1.Value.(type) {
				case html.Token:
					switch n := arg1.(type) {
					case env.Integer:
						if int(n.Value) < len(tok.Attr) {
							return *env.NewString(tok.Attr[int(n.Value)].Val)
						} else {
							return MakeBuiltinError(ps, "Attribute index out of bounds.", "rye-html-start//Attr?")
						}
					case env.Word:
						attrName := ps.Idx.GetWord(n.Index)
						for _, a := range tok.Attr {
							if a.Key == attrName {
								return *env.NewString(a.Val)
							}
						}
						return MakeBuiltinError(ps, "Attribute '"+attrName+"' not found.", "rye-html-start//Attr?")
					case env.String:
						attrName := n.Value
						for _, a := range tok.Attr {
							if a.Key == attrName {
								return *env.NewString(a.Val)
							}
						}
						return MakeBuiltinError(ps, "Attribute '"+attrName+"' not found.", "rye-html-start//Attr?")
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.WordType, env.StringType}, "rye-html-start//Attr?")
					}
				default:
					return MakeBuiltinError(ps, "Token value is not matching.", "rye-html-start//Attr?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-html-start//Attr?")
			}
		},
	},

	// Args:
	// * element: HTML token element
	// Returns:
	// * dict with all attributes where keys are attribute names
	"rye-html-start//Attrs?": {
		Argsn: 1,
		Doc:   "Returns a dict of all attributes from an HTML element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tok1 := arg0.(type) {
			case env.Native:
				switch tok := tok1.Value.(type) {
				case html.Token:
					data := make(map[string]any)
					for _, attr := range tok.Attr {
						data[attr.Key] = *env.NewString(attr.Val)
					}
					return *env.NewDict(data)
				default:
					return MakeBuiltinError(ps, "Token value is not matching.", "rye-html-start//Attrs?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-html-start//Attrs?")
			}
		},
	},

	// Tests:
	// stdout { "<div></div>" |reader .Parse-html { <div> [ .Name? |print ] } } "div\n"
	// Args:
	// * element: HTML token element
	// Returns:
	// * string name of the HTML element
	"rye-html-start//Name?": {
		Argsn: 1,
		Doc:   "Returns the name of an HTML element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tok1 := arg0.(type) {
			case env.Native:
				switch tok := tok1.Value.(type) {
				case html.Token:
					return *env.NewString(tok.Data)
				default:
					return MakeBuiltinError(ps, "Not xml-start element.", "rye-html-start//Name?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-html-start//Name?")
			}
		},
	},

	/*	"rye-html-start//name?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					return env.String{obj1.Name.Local}
				default:
					return env.NewError("Not xml-start element")
				}
			default:
				return env.NewError("first argument should be native")
			}
		},
	},*/
}
