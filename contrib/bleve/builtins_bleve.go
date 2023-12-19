//go:build b_bleve
// +build b_bleve

package bleve

import (
	"encoding/json"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/blevesearch/bleve/v2/search/query"
	index "github.com/blevesearch/bleve_index_api"

	"fmt"
)

var Builtins_bleve = map[string]*env.Builtin{

	"new-bleve": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch mpi := arg0.(type) {
			case env.Native:
				switch s := arg1.(type) {
				case env.Uri:
					path := strings.Split(s.Path, "://")
					iindex, err := bleve.New(path[1], mpi.Value.(mapping.IndexMapping))
					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}
					return *env.NewNative(ps.Idx, iindex, "bleve-index")
				default:
					return evaldo.MakeError(ps, "Arg 2 not file Uri.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"open-bleve": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				path := strings.Split(s.Path, "://")
				iindex, err := bleve.Open(path[1])
				if err != nil {
					return evaldo.MakeError(ps, err.Error())
				}
				return *env.NewNative(ps.Idx, iindex, "bleve-index")
			default:
				return evaldo.MakeError(ps, "Arg 1 not file Uri.")
			}
		},
	},

	"new-bleve-text-field-mapping": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			mapping := bleve.NewTextFieldMapping()
			return *env.NewNative(ps.Idx, mapping, "bleve-text-field-mapping")
		},
	},
	"new-bleve-document-mapping": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			mapping := bleve.NewDocumentMapping()
			return *env.NewNative(ps.Idx, mapping, "bleve-document-mapping")
		},
	},
	"new-bleve-index-mapping": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			mapping := bleve.NewIndexMapping()
			return *env.NewNative(ps.Idx, mapping, "bleve-index-mapping")
		},
	},

	"bleve-index-mapping//add-document-mapping": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch docmap := arg0.(type) {
			case env.Native:
				switch name := arg1.(type) {
				case env.String:
					switch fmapping := arg2.(type) {
					case env.Native:
						docmap.Value.(*mapping.IndexMappingImpl).AddDocumentMapping(name.Value, fmapping.Value.(*mapping.DocumentMapping))
						return arg2
					default:
						return evaldo.MakeError(ps, "Arg 3 not native")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 not String")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not Native.")
			}
		},
	},

	"bleve-document-mapping//add-field-mapping-at": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch docmap := arg0.(type) {
			case env.Native: // *mapping.DocumentMapping
				switch name := arg1.(type) {
				case env.String:
					switch fmapping := arg2.(type) {
					case env.Native:
						docmap.Value.(*mapping.DocumentMapping).AddFieldMappingsAt(name.Value, fmapping.Value.(*mapping.FieldMapping))
						return arg0
					default:
						return evaldo.MakeError(ps, "Arg 3 not native")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 not String")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not Native.")
			}
		},
	},

	"bleve-index//index": {
		Argsn: 3,
		Doc:   "[ ses-session* gomail-message from-email recipients ]",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch idx := arg0.(type) {
			case env.Native:
				switch ident := arg1.(type) { // gomail-message
				case env.String:
					switch text := arg2.(type) { // recipients
					case env.String:
						var doc any
						json.Unmarshal([]byte(text.Value), &doc)
						err := idx.Value.(bleve.Index).Index(ident.Value, doc)
						if err != nil {
							return evaldo.MakeError(ps, err.Error())
						}
						return arg0
					case env.Dict:
						err := idx.Value.(bleve.Index).Index(ident.Value, text.Data)
						if err != nil {
							return evaldo.MakeError(ps, err.Error())
						}
						return arg0
					default:
						return evaldo.MakeError(ps, "A3 not String")
					}
				default:
					return evaldo.MakeError(ps, "A2 not String")
				}
			default:
				return evaldo.MakeError(ps, "A1 not Native")
			}
			return nil
		},
	},

	"new-match-query": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch text := arg0.(type) {
			case env.String:
				query := bleve.NewMatchQuery(text.Value)
				return *env.NewNative(ps.Idx, query, "bleve-query")
			default:
				return evaldo.MakeError(ps, "Arg 1 not String..")
			}
		},
	},
	"bleve-query//new-search-request": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch qry := arg0.(type) {
			case env.Native:
				search := bleve.NewSearchRequest(qry.Value.(query.Query))
				return *env.NewNative(ps.Idx, search, "bleve-search")
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},
	"bleve-search//search": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch search := arg0.(type) {
			case env.Native:
				switch _index := arg1.(type) {
				case env.Native:
					iindex := _index.Value.(bleve.Index)
					searchResults, err := iindex.Search(search.Value.(*bleve.SearchRequest))
					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}

					//fmt.Println(searchResults)
					// for _, hit := range searchResults.Hits {
					/*						fmt.Println(hit.ID)
											fmt.Println(hit.Index)
											fmt.Println(hit.Expl)
											fmt.Println(hit.Locations)
											fmt.Println("*")
					*/
					// hhit, _ := iindex.Document(hit.ID)

					/*						for fragmentField, fragments := range hhit.Fragments {
												fmt.Println("-")
												fmt.Printf("\t%s***\n", fragmentField)
												for _, fragment := range fragments {
													fmt.Printf("\t\t%s\n", fragment)
												}
											}
					*/
					//						fmt.Println("****")
					//						fmt.Println(hhit.Size())

					//hhit.VisitFields(func(ff index.Field) {
					//	fmt.Println(ff.Name())
					//	fmt.Println(string(ff.Value()))
					//})

					//for otherFieldName, otherFieldValue := range hit.Fields {
					//	fmt.Println("+")
					//	if _, ok := hit.Fragments[otherFieldName]; !ok {
					//		fmt.Printf("\t%s\n", otherFieldName)
					//		fmt.Printf("\t\t%v\n", otherFieldValue)
					//	}
					// }
					//fmt.Println(searchResults.Hits[sr])
					//dm := searchResults.Hits[sr] // range searchResults.Hits[sr] {
					//fmt.Println(dm.Fields)

					//}
					return *env.NewNative(ps.Idx, searchResults, "bleve-results")
				default:
					return evaldo.MakeError(ps, "Arg 1 not native.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},
	"bleve-results//summary": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sr := arg0.(type) {
			case env.Native:
				res := fmt.Sprint(sr.Value.(*bleve.SearchResult))
				return env.String{res}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},
	"bleve-results//to-list": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sr := arg0.(type) {
			case env.Native:
				switch _index := arg1.(type) {
				case env.Native:
					iindex := _index.Value.(bleve.Index)
					sr_ := sr.Value.(*bleve.SearchResult)
					data := make([]any, sr_.Hits.Len())
					for i, hit := range sr_.Hits {

						item := make(map[string]any)

						item["_id"] = hit.ID
						// item["_score"] = strconv.Itoa(hit.)

						hhit, _ := iindex.Document(hit.ID)
						hhit.VisitFields(func(ff index.Field) {
							item[ff.Name()] = string(ff.Value())
						})

						data[i] = *env.NewDict(item)

					}

					return env.List{data, env.Word{0}}
				default:
					return evaldo.MakeError(ps, "Arg 2 not native.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},
}
