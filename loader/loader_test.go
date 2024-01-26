package loader

import (
	"fmt"

	"github.com/refaktor/rye/env"

	//"fmt"
	"testing"
)

func TestLoader_load_integer(t *testing.T) {
	input := "123"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	// fmt.Println(block.(env.Block).Series.Get(0).Type())

	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_load_negative_integer(t *testing.T) {
	input := "-123"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	// fmt.Println(block.(env.Block).Series.Get(0).Type())

	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_load_integers(t *testing.T) {
	input := "123 342 453"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_load_decimal(t *testing.T) {
	input := "123.231"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	// fmt.Println(block.(env.Block).Series.Get(0).Type())

	if block.(env.Block).Series.Get(0).Type() != env.DecimalType {
		t.Error("Expected type decimal")
	}
}

func TestLoader_load_negative_decimal(t *testing.T) {
	input := "-123.324"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	// fmt.Println(block.(env.Block).Series.Get(0).Type())

	if block.(env.Block).Series.Get(0).Type() != env.DecimalType {
		t.Error("Expected type decimal")
	}
}

func TestLoader_load_decimals(t *testing.T) {
	input := "-123.1 -342.2 -453.3"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.DecimalType {
		t.Error("Expected type decimal")
	}
}

func TestLoader_load_word(t *testing.T) {
	input := "wowo"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.WordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_words(t *testing.T) {
	input := "wowowo wawawa yoyoyo"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.WordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setword(t *testing.T) {
	input := "wowo:"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.SetwordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setwords(t *testing.T) {
	input := "wowo: wawa: wiwi:"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.SetwordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setword_check_colon(t *testing.T) {
	input := "wowo:"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.SetwordType {
		t.Error("Expected type word")
	}

	idx := block.(env.Block).Series.Get(0).(env.Setword).Index

	if wordIndex.GetWord(idx) == "wowo:" {
		t.Error("Collon added to word")
	}

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func TestLoader_load_opword_1(t *testing.T) {
	input := ".wowo"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.OpwordType {
		t.Error("Expected type Opword")
	}

	idx := block.(env.Block).Series.Get(0).(env.Opword).Index

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func TestLoader_load_pipeword_1(t *testing.T) {
	input := "|wowo"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.PipewordType {
		t.Error("Expected type Pipeword")
	}

	idx := block.(env.Block).Series.Get(0).(env.Pipeword).Index

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func TestLoader_just_load_various(t *testing.T) {
	input := "123 word 3 { setword: 23 } end 12 word"
	LoadString(input, false)
}

func TestLoader_load_mixed(t *testing.T) {
	input := "wowo: inc 123"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.SetwordType {
		t.Error("Expected type setword")
	}
	if block.(env.Block).Series.Get(1).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.(env.Block).Series.Get(2).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_spaces(t *testing.T) {
	input := "   123	 "
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_newlines(t *testing.T) {
	input := "\n   123	 \n"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_newlines2(t *testing.T) {
	input := "\n\t123	 \nword\nword2\tsetword2:\n\t234"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 5 {
		t.Error("Expected 5 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
	if block.(env.Block).Series.Get(1).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.(env.Block).Series.Get(2).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.(env.Block).Series.Get(3).Type() != env.SetwordType {
		t.Error("Expected type set-word")
	}
	if block.(env.Block).Series.Get(4).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_blocks(t *testing.T) {
	input := "\n\t123	{ { 22 } aa } \nword2\tsetword2:\n\t234"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 5 {
		t.Error("Expected 5 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
	if block.(env.Block).Series.Get(1).Type() != env.BlockType {
		t.Error("Expected type block")
	}
	if block.(env.Block).Series.Get(1).(env.Block).Series.Get(0).Type() != env.BlockType {
		t.Error("Expected type block")
	}
	if block.(env.Block).Series.Get(1).(env.Block).Series.Get(1).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.(env.Block).Series.Get(1).(env.Block).Series.Get(0).(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
	if block.(env.Block).Series.Get(2).Type() != env.WordType {
		t.Error("Expected type word")
	}
	if block.(env.Block).Series.Get(3).Type() != env.SetwordType {
		t.Error("Expected type set-word")
	}
	if block.(env.Block).Series.Get(4).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_load_string_1(t *testing.T) {
	input := "\" wowo 123 !._' \""
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.StringType {
		t.Error("Expected type String")
	} else {
		fmt.Println(block.(env.Block).Series.Get(0).(env.String).Value)
		if block.(env.Block).Series.Get(0).(env.String).Value != " wowo 123 !._' " {
			t.Error("Not correct string content")
		}
	}
}

func TestLoader_load_void_comma(t *testing.T) {
	input := ", _"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 2 {
		t.Error("Expected 2 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.CommaType {
		t.Error("Expected type Comma")
	}
	if block.(env.Block).Series.Get(1).Type() != env.VoidType {
		t.Error("Expected type Void")
	}
}

/*func TestLoader_load_words(t *testing.T) {
	loader1 := NewLoader()
	input := "word word2"
	val, _ := loader1.ParseAndGetValue(input, nil)
	if len(val.(env.Block).Series) == 2 {
		t.Error("Expected 2 words")
	}
}

func TestLoader_load_setword(t *testing.T) {
	loader1 := NewLoader()
	input := "setword: inv 23"
	val, _ := loader1.ParseAndGetValue(input, nil)
	if len(val.(env.Block).Series) == 3 {
		t.Error("Expected 1 item")
	}
}
*/

func TestLoader_load_argword(t *testing.T) {
	input := "{somename:somekind}"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	fmt.Println(block.(env.Block).Series.Get(0).Inspect(*wordIndex))

	if block.(env.Block).Series.Get(0).Type() != env.ArgwordType {
		t.Error("Expected type Argword")
	}
	idx, _ := wordIndex.GetIndex("somename")
	if block.(env.Block).Series.Get(0).(env.Argword).Name.Index != idx {
		t.Error("Expected name somename")
	}
	idx2, _ := wordIndex.GetIndex("somekind")
	if block.(env.Block).Series.Get(0).(env.Argword).Kind.Index != idx2 {
		t.Error("Expected kind somekind")
	}
}

func TestLoader_load_group(t *testing.T) {
	input := "( 1 2 , sada )"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	fmt.Println(block.(env.Block).Series.Get(0).Inspect(*wordIndex))

	if block.(env.Block).Series.Get(0).Type() != env.BlockType {
		t.Error("Expected type Block")
	}
}

func TestLoader_load_lsetword(t *testing.T) {
	input := "123 :lsetword1"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 2 {
		t.Error("Expected 1 items")
	}

	fmt.Println(block.(env.Block).Series.Get(0).Inspect(*wordIndex))

	if block.(env.Block).Series.Get(1).Type() != env.LSetwordType {
		t.Error("Expected type LSetword")
	}
	idx, _ := wordIndex.GetIndex("lsetword1")
	if block.(env.Block).Series.Get(1).(env.LSetword).Index != idx {
		t.Error("Expected name lsetword")
	}
}

func TestLoader_load_uri_min(t *testing.T) {
	input := "sqlite://db"
	block, _ := LoadString(input, false)
	block.Trace("BLOCK URI ....")
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	//fmt.Println(block.(env.Block).Series.Get(0).Inspect(wordIndex))

	if block.(env.Block).Series.Get(0).Type() != env.UriType {
		t.Error("Expected type Uri")
	}
	idx, _ := wordIndex.GetIndex("sqlite")
	if block.(env.Block).Series.Get(0).(env.Uri).Scheme.Index != idx {
		t.Error("Expected scheme sqlite")
	}

	if block.(env.Block).Series.Get(0).(env.Uri).Path != "db" { // todo later return just the path part ... but there are more components to URI, so we do it later
		t.Error("Expected path sqlite://db")
	}
}

func TestLoader_cpath(t *testing.T) {
	input := "user/check/user"
	block, _ := LoadString(input, false)
	block.Trace("CPATH")
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	//fmt.Println(block.(env.Block).Series.Get(0).Inspect(wordIndex))

	if block.(env.Block).Series.Get(0).Type() != env.CPathType {
		t.Error("Expected type CPath")
	}

	fmt.Println(block.(env.Block).Series.Get(0))

	block.(env.Block).Series.Get(0).(env.CPath).Word1.Probe(*wordIndex)

	idx, _ := wordIndex.GetIndex("user")
	if block.(env.Block).Series.Get(0).(env.CPath).Word1.Index != idx {
		t.Error("Expected context user")
	}

	idx2, _ := wordIndex.GetIndex("check")
	if block.(env.Block).Series.Get(0).(env.CPath).Word2.Index != idx2 { // todo later return just the path part ... but there are more components to URI, so we do it later
		t.Error("Expected word1 check")
	}

	idx3, _ := wordIndex.GetIndex("user")
	if block.(env.Block).Series.Get(0).(env.CPath).Word3.Index != idx3 { // todo later return just the path part ... but there are more components to URI, so we do it later
		t.Error("Expected word1 user")
	}
}

func TestLoader_load_tagword_1(t *testing.T) {
	input := "'wowo"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.TagwordType {
		t.Error("Expected type Pipeword")
	}

	idx := block.(env.Block).Series.Get(0).(env.Tagword).Index

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func TestLoader_load_xword_1(t *testing.T) {
	input := "<wowo>"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.XwordType {
		t.Error("Expected type Xword")
	}

	idx := block.(env.Block).Series.Get(0).(env.Xword).Index

	fmt.Println(wordIndex.GetWord(idx))

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func DISABLED__TestLoader_load_exword_1(t *testing.T) {
	input := "<wowo>"
	block, _ := LoadString(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.EXwordType {
		t.Error("Expected type EXword")
	}

	idx := block.(env.Block).Series.Get(0).(env.EXword).Index

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}
