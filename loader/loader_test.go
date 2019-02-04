package loader

import (
	"Rejy_go_v1/env"
	//"fmt"
	"testing"
)

func TestLoader_load_integer(t *testing.T) {
	/*loader1 := NewParser()
	input := "{ 123 }"
	val, _ := loader1.ParseAndGetValue(input, nil)
	InspectNode(val)
	block := val.(env.Block)*/
	input := "{ 123 }"
	block, _ := LoadString(input)
	if block.Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_load_integers(t *testing.T) {
	input := "{ 123 342 453 }"
	block, _ := LoadString(input)
	if block.Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_load_word(t *testing.T) {
	input := "{ wowo }"
	block, _ := LoadString(input)
	if block.Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.Series.Get(0).(env.Object).Type() != env.WordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_words(t *testing.T) {
	input := "{ wowowo wawawa yoyoyo }"
	block, _ := LoadString(input)
	if block.Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.WordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setword(t *testing.T) {
	input := "{ wowo: }"
	block, _ := LoadString(input)
	if block.Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.Series.Get(0).(env.Object).Type() != env.SetwordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setwords(t *testing.T) {
	input := "{ wowo: wawa: wiwi: }"
	block, _ := LoadString(input)
	if block.Series.Len() != 3 {
		t.Error("Expected 1 item")
	}
	if block.Series.Get(0).(env.Object).Type() != env.SetwordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setword_check_colon(t *testing.T) {
	input := "{ wowo: }"
	block, _ := LoadString(input)
	if block.Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.Series.Get(0).(env.Object).Type() != env.SetwordType {
		t.Error("Expected type word")
	}

	idx := block.Series.Get(0).(env.Setword).Index

	if genv.GetWord(idx) == "wowo:" {
		t.Error("Collon added to word")
	}

	if genv.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func TestLoader_just_load_various(t *testing.T) {
	input := "{ 123 word 3 { setword: 23 } end 12 word }"
	LoadString(input)
}

func TestLoader_load_mixed(t *testing.T) {
	input := "{ wowo: inc 123 }"
	block, _ := LoadString(input)
	if block.Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.SetwordType {
		t.Error("Expected type setword")
	}
	if block.Series.Get(1).(env.Object).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.Series.Get(2).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_spaces(t *testing.T) {
	input := "{    123	  }"
	block, _ := LoadString(input)
	if block.Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_newlines(t *testing.T) {
	input := "{ \n   123	 \n }"
	block, _ := LoadString(input)
	if block.Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_newlines2(t *testing.T) {
	input := "{ \n\t123	 \nword\nword2\tsetword2:\n\t234 }"
	block, _ := LoadString(input)
	if block.Series.Len() != 5 {
		t.Error("Expected 5 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
	if block.Series.Get(1).(env.Object).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.Series.Get(2).(env.Object).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.Series.Get(3).(env.Object).Type() != env.SetwordType {
		t.Error("Expected type set-word")
	}
	if block.Series.Get(4).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_blocks(t *testing.T) {
	input := "{ \n\t123	{ { 22 } aa } \nword2\tsetword2:\n\t234 }"
	block, _ := LoadString(input)
	if block.Series.Len() != 5 {
		t.Error("Expected 5 items")
	}
	if block.Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
	if block.Series.Get(1).(env.Object).Type() != env.BlockType {
		t.Error("Expected type block")
	}
	if block.Series.Get(1).(env.Block).Series.Get(0).(env.Object).Type() != env.BlockType {
		t.Error("Expected type block")
	}
	if block.Series.Get(1).(env.Block).Series.Get(1).(env.Object).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.Series.Get(1).(env.Block).Series.Get(0).(env.Block).Series.Get(0).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
	if block.Series.Get(2).(env.Object).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.Series.Get(3).(env.Object).Type() != env.SetwordType {
		t.Error("Expected type set-word")
	}
	if block.Series.Get(4).(env.Object).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

/*func TestLoader_load_words(t *testing.T) {
	loader1 := NewLoader()
	input := "{ word word2 }"
	val, _ := loader1.ParseAndGetValue(input, nil)
	if len(val.(env.Block).Series) == 2 {
		t.Error("Expected 2 words")
	}
}

func TestLoader_load_setword(t *testing.T) {
	loader1 := NewLoader()
	input := "{ setword: inv 23 }"
	val, _ := loader1.ParseAndGetValue(input, nil)
	if len(val.(env.Block).Series) == 3 {
		t.Error("Expected 1 item")
	}
}
*/
