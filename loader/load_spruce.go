// loader.go
package loader

import (
	"fmt"
	"rye/env"
)

/*

We need to parse a textual identation based representation of a tree into
a tree of nodes, which are Rye nodes (words, blocks, ...)

We do it character by character
if character is newline we create and add the current node to the tree ...
it depends if it's a sibling, child, or aunt to the current parent node.

(We probably need to always store the parent node, and through it we can get to it's parents
if we need to)

Parsing lines first won't work nicely, since we can have multiline blocks.


*/

const NOTHING = 0
const STARTL = 1
const SPACES = 2
const INWORD = 3
const INBLOCK = 4

const example1 = `this
 is
  a
   test
    { block of words }`

func LoadSpruceString(input string) (*env.SprNode, *env.Idxs) {
	var state = STARTL
	var space_cnt = 0
	var curr_str = ""
	var depth = 0
	var lastParent = env.NewSprNode(nil, -1, nil) // root node
	var root = lastParent

	// for each character
	for i, c := range input {
		fmt.Println(i, " => ", string(c))

		if state != INBLOCK {
			if c == ' ' {
				if state == STARTL {
					space_cnt += 1
				} else if state == INWORD {
					curr_str += string(c)
				}
			} else if c == '{' {
				if state == STARTL {
					state = INBLOCK
				}
				curr_str = ""
				depth = space_cnt
			} else if c == '\n' {
				if state == INWORD {
					block, _ := LoadString(curr_str, false)
					curr_str = ""
					var obj = block.(env.Block).Series.Get(0)
					fmt.Println("ADDING")
					fmt.Println(lastParent)
					fmt.Println(depth)
					var parent = findParentNode(lastParent, depth)
					//					lastParent = parent
					//					parent.Children = append(parent.Children, *env.NewSprNode(obj, depth, parent))
					fmt.Println("ADDING CHILD")
					fmt.Println(parent)
					var node = *env.NewSprNode(obj, depth, parent)
					parent.Children = append(parent.Children, &node)
					fmt.Println("CHILD ADDED")
					fmt.Println(parent)
					lastParent = &node
				}
				state = STARTL
				// DO if cur_node is not null create node with depth
				// parse current line by Rye loader, if there is only one element
				// create SprNode with it as an object
				// find appropriate parent node and add it as a child
				space_cnt = 0
				depth = 0
			} else {
				if state == STARTL {
					state = INWORD
					depth = space_cnt
				}
				curr_str += string(c)
			}
		} else {
			if c == '}' {
				state = NOTHING
				block, _ := LoadString(curr_str, false)
				fmt.Println("PARSED BLOCK:")
				fmt.Println(curr_str)
				var parent = findParentNode(lastParent, depth)
				var node = *env.NewSprNode(block, depth, parent)
				parent.Children = append(parent.Children, &node)
				curr_str = ""
				lastParent = &node
				// DO parse curr_str as Rye code, get the block Object back
				// create node block
				// find appropriate parent node and add it as a child
			} else {
				curr_str += string(c)
			}
		}
	}
	return root, wordIndex
}

func findParentNode(parent *env.SprNode, depth int) *env.SprNode {
	var cnt = 0
	for true {
		fmt.Println("LOOP")
		fmt.Println(parent)
		fmt.Println(depth)
		if parent.Depth == depth-1 {
			return parent
		}
		if parent.Depth > -1 {
			parent = parent.Parent
		}
		cnt += 1
		if cnt > 30 {
			return nil
		}
	}
	return nil
}
