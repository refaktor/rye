package env

import (
	"fmt"
	"strings"
)

// A Spruce tree is a tree of SprNodes
// it is loaded via loader_spruce that parses the text and creates the nodes
// For now it uses Objects of type Word and Block. Word is a path in spruce tree, block is a Rye code that
// gets executed . We will probably also need blocks to define values, it might be ordinary block, or some
// special type of block like {name:type} ... we will see

type SprNode struct {
	Value    Object
	Children []*SprNode
	Depth    int
	Parent   *SprNode
}

func NewSprNode(value Object, depth int, parent *SprNode) *SprNode {
	var s SprNode
	s.Value = value
	s.Depth = depth
	s.Parent = parent
	return &s
}

func (n SprNode) FindChild(idx int) *SprNode {
	for _, child := range n.Children {
		if child.Value.Type() == WordType && child.Value.(Word).Index == idx {
			return child
		}
	}
	return nil
}

func (n SprNode) Print(depth int, idxs Idxs) {
	for _, child := range n.Children {
		fmt.Println(LeftPad(child.Value.Inspect(idxs), " ", depth))
		child.Print(depth+1, idxs)
	}
}

func LeftPad(s string, padStr string, padCount int) string {
	var retStr = strings.Repeat(padStr, padCount) + s
	return retStr
}
