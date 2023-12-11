package loader

import (
	"fmt"
	"rye/env"

	//"fmt"

	//"fmt"
	"testing"
)

func _TestLoaderSpruce_empty(t *testing.T) {
	input := ``
	root, _ := LoadSpruceString(input)
	fmt.Println(root)
	if root.Depth != -1 {
		t.Error("Expected depth -1")
	}
	/*
		if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
			t.Error("Expected type integer")
		}*/
}

func _TestLoaderSpruce_one_level(t *testing.T) {
	input := `this
you
they	
`
	root, _ := LoadSpruceString(input)
	fmt.Println(root)
	fmt.Println(len(root.Children))
	if len(root.Children) != 3 {
		t.Error("Expected len 3")
	} /*
		if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
			t.Error("Expected type integer")
		}*/
}

func TestLoaderSpruce_two_levels(t *testing.T) {
	input := `level0
 level1
  level2
`
	root, _ := LoadSpruceString(input)
	fmt.Println(root)
	fmt.Println(len(root.Children))
	if len(root.Children) != 1 {
		t.Error("Expected len 3")
	}
	if root.Children[0].Depth != 0 {
		t.Error("Expected depth 0")
	}
	fmt.Println("depths")
	fmt.Println(root)
	fmt.Println(len(root.Children))
	fmt.Println(len(root.Children[0].Children))
	//fmt.Println(len(root.Children[1].Children))
	//fmt.Println(len(root.Children[2].Children))
	if len(root.Children[0].Children) != 1 {
		t.Error("Expected len 1")
	}
	if len(root.Children[0].Children[0].Children) != 1 {
		t.Error("Expected len 1")
	}
	if root.Children[0].Children[0].Children[0].Depth != 2 {
		t.Error("Expected depth 2")
	}
}
func TestLoaderSpruce_tree_1(t *testing.T) {
	input := `level0
 level1a
  level2
 level1b
more1
 more2a
 more2b
`
	root, _ := LoadSpruceString(input)
	if len(root.Children) != 2 {
		t.Error("Expected len 2")
	}
	if len(root.Children[0].Children) != 2 {
		t.Error("Expected len 2")
	}
	if len(root.Children[0].Children[0].Children) != 1 {
		t.Error("Expected len 1")
	}
	if len(root.Children[1].Children) != 2 {
		t.Error("Expected len 2")
	}
}

func TestLoaderSpruce_branch_block(t *testing.T) {
	input := `level0
 { add 2 3 }
`
	root, _ := LoadSpruceString(input)
	if len(root.Children) != 1 {
		t.Error("Expected len 1")
	}
	if len(root.Children[0].Children) != 1 {
		t.Error("Expected len 1")
	}
	if root.Children[0].Children[0].Value.Type() != env.BlockType {
		t.Error("Expected block type")
	}

	fmt.Println(root.Children[0].Children[0].Value)

	if root.Children[0].Children[0].Value.(env.Block).Series.Len() != 3 {
		t.Error("Expected block length 3")
	}

	/*

		when we load string we also get genv (Environment) ... how do we combine multiple of these
		together so we don't have multiple environments with each registering it's own builtins for example,
		but combine them together. We execute this block at different time, in user-mode, not build mode. so ....

		block, genv := loader.LoadString(input)
		es := env.NewProgramState(block.Series, genv)
		RegisterBuiltins(es)
		es = EvalBlock(es)
	*/
}
