// builtins_trees.go
package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

var builtins_trees = map[string]*env.Builtin{

	//
	// ##### Tree Traversal ##### "Functions for traversing tree-like data structures"
	//
	// Tests:
	// ; TODO: Add tests for for-tree
	// Args:
	// * node: Initial tree node to start traversal from
	// * nodeWord: Word to bind the current node to during traversal
	// * condition: Block that evaluates to a boolean to determine if a node should be processed
	// * branch: Block that returns child nodes for the current node
	// * body: Block of code to execute for each node that meets the condition
	// Returns:
	// * result of the last body block execution
	"for-tree": {
		Argsn: 5,
		Doc:   "Traverses a tree structure, executing a block of code for each node that meets a condition.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// arg0: Initial node
			// arg1: Node word (to store current node)
			// arg2: Condition block
			// arg3: Branch block (returns child nodes)
			// arg4: Body block (executed for each node)

			switch nodeWord := arg1.(type) {
			case env.Word:
				switch condBlock := arg2.(type) {
				case env.Block:
					switch branchBlock := arg3.(type) {
					case env.Block:
						switch bodyBlock := arg4.(type) {
						case env.Block:
							// Save original series
							ser := ps.Ser

							// Create a recursive function to traverse the tree
							var traverseTree func(node env.Object) env.Object
							traverseTree = func(node env.Object) env.Object {
								// Store current node in context
								fmt.Println(ps.FailureFlag)
								ps.Ctx.Mod(nodeWord.Index, node)
								ser := ps.Ser
								// Evaluate condition
								ps.Ser = condBlock.Series
								// ps.Ser.Reset()
								fmt.Println(ps)
								fmt.Println(ps.Ser)
								EvalBlock(ps)
								if ps.ErrorFlag {
									fmt.Println("***1")
									ps.Ser = ser
									return ps.Res
								}
								ps.Ser = ser

								// If condition is true, execute body
								if util.IsTruthy(ps.Res) {
									// Execute body block
									ser = ps.Ser
									ps.Ser = bodyBlock.Series
									ps.Ser.Reset()
									EvalBlock(ps)
									if ps.ErrorFlag || ps.ReturnFlag {
										fmt.Println("***2")
										ps.Ser = ser
										return ps.Res
									}
									ps.Ser = ser

									// Get child nodes using branch block
									ser := ps.Ser
									ps.Ser = branchBlock.Series
									res := make([]env.Object, 0)
									for ps.Ser.Pos() < ps.Ser.Len() {
										// ps, injnow = EvalExpressionInj(ps, inj, injnow)
										EvalExpression_CollectArg(ps, false)
										if ps.ReturnFlag || ps.ErrorFlag {
											return ps.Res
										}
										res = append(res, ps.Res)
										// check and raise the flags if needed if true (error) return
										//if checkFlagsAfterBlock(ps, 101) {
										//	return ps
										//}
										// if return flag was raised return ( errorflag I think would return in previous if anyway)
										// ps, injnow = MaybeAcceptComma(ps, inj, injnow)
									}
									ps.Ser = ser
									ps.Res = *env.NewBlock(*env.NewTSeries(res))

									/* Get child nodes using branch block
									ps.Ser = branchBlock.Series
									ps.Ser.Reset()
									EvalBlock(ps)
									if ps.ErrorFlag {
										fmt.Println("***3")
										fmt.Println(ps.Res)
										return ps.Res
									}*/
									fmt.Println("---")
									// Process child nodes
									switch children := ps.Res.(type) {
									case env.Block:
										for i := 0; i < children.Series.Len(); i++ {
											childNode := children.Series.Get(i)
											traverseTree(childNode)
											if ps.ErrorFlag || ps.ReturnFlag {
												fmt.Println("***4")
												break
											}
										}
									case env.List:
										for i := 0; i < len(children.Data); i++ {
											childNode := env.ToRyeValue(children.Data[i])
											traverseTree(childNode)
											if ps.ErrorFlag || ps.ReturnFlag {
												fmt.Println("***5")
												break
											}
										}
									}
								}

								return ps.Res
							}

							// Start traversal with initial node
							result := traverseTree(arg0)

							// Restore original series
							ps.Ser = ser
							return result
						default:
							return MakeArgError(ps, 4, []env.Type{env.BlockType}, "for-tree")
						}
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "for-tree")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for-tree")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "for-tree")
			}
		},
	},

	// Tests:
	// ; TODO: Add tests for tree-map
	// Args:
	// * node: Initial tree node to start traversal from
	// * transformer: Block or builtin to transform each node
	// Returns:
	// * a new tree with the same structure but transformed nodes
	"tree-map": {
		Argsn: 2,
		Doc:   "Creates a new tree by applying a transformation to each node while preserving the structure.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch node := arg0.(type) {
			case env.Dict:
				switch transformer := arg1.(type) {
				case env.Block, env.Builtin:
					// Save original series
					ser := ps.Ser

					// Create a recursive function to transform the tree
					var transformTree func(node env.Dict) env.Dict
					transformTree = func(node env.Dict) env.Dict {
						// Create a new dictionary for the transformed node
						newData := make(map[string]any)

						// Copy and transform each key-value pair
						for k, v := range node.Data {
							// Special handling for child nodes
							if k == "left" || k == "right" || k == "children" {
								switch childNode := v.(type) {
								case env.Dict:
									newData[k] = transformTree(childNode)
								case env.List:
									// Transform list of child nodes
									childrenList := make([]any, 0)
									for _, child := range childNode.Data {
										if dict, ok := child.(env.Dict); ok {
											childrenList = append(childrenList, transformTree(dict))
										} else {
											childrenList = append(childrenList, child)
										}
									}
									newData[k] = *env.NewList(childrenList)
								default:
									newData[k] = v
								}
							} else {
								// Transform the value using the provided transformer
								switch transformer := transformer.(type) {
								case env.Block:
									ps.Ser = transformer.Series
									ps.Ser.Reset()
									EvalBlockInj(ps, env.ToRyeValue(v), true)
									if ps.ErrorFlag {
										return node // Return original on error
									}
									newData[k] = ps.Res
								case env.Builtin:
									result := DirectlyCallBuiltin(ps, transformer, env.ToRyeValue(v), nil)
									newData[k] = result
								}
							}
						}

						return env.Dict{Data: newData, Kind: node.Kind}
					}

					// Transform the tree starting from the root node
					result := transformTree(node)

					// Restore original series
					ps.Ser = ser
					return result
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "tree-map")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "tree-map")
			}
		},
	},

	// Tests:
	// ; TODO: Add tests for tree-fold
	// Args:
	// * node: Initial tree node to start traversal from
	// * accumWord: Word to bind the accumulator to
	// * initial: Initial value for the accumulator
	// * folder: Block that combines the accumulator with each node value
	// Returns:
	// * final accumulated value after traversing the entire tree
	"tree-fold": {
		Argsn: 3,
		Doc:   "Traverses a tree and accumulates a value by applying a function to each node.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch node := arg0.(type) {
			case env.Dict:
				switch accumWord := arg1.(type) {
				case env.Word:
					switch folder := arg3.(type) {
					case env.Block:
						// Save original series
						ser := ps.Ser

						// Initialize accumulator with the provided initial value
						acc := arg2

						// Create a recursive function to fold the tree
						var foldTree func(node env.Dict) env.Object
						foldTree = func(node env.Dict) env.Object {
							// Get the node value
							nodeValue, hasValue := node.Data["value"]

							if hasValue {
								// Update accumulator with the current node value
								ps.Ctx.Mod(accumWord.Index, acc)
								ps.Ser = folder.Series
								ps.Ser.Reset()
								EvalBlockInj(ps, env.ToRyeValue(nodeValue), true)
								if ps.ErrorFlag {
									return ps.Res
								}
								acc = ps.Res
							}

							// Process left child if present
							if left, hasLeft := node.Data["left"]; hasLeft {
								if leftDict, ok := left.(env.Dict); ok {
									foldTree(leftDict)
								}
							}

							// Process right child if present
							if right, hasRight := node.Data["right"]; hasRight {
								if rightDict, ok := right.(env.Dict); ok {
									foldTree(rightDict)
								}
							}

							// Process children list if present
							if children, hasChildren := node.Data["children"]; hasChildren {
								if childrenList, ok := children.(env.List); ok {
									for _, child := range childrenList.Data {
										if childDict, ok := child.(env.Dict); ok {
											foldTree(childDict)
										}
									}
								}
							}

							return acc
						}

						// Fold the tree starting from the root node
						result := foldTree(node)

						// Restore original series
						ps.Ser = ser
						return result
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "tree-fold")
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.WordType}, "tree-fold")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "tree-fold")
			}
		},
	},
}
