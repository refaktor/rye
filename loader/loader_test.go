package loader

import (
	"fmt"
	"strconv"

	"github.com/refaktor/rye/env"

	//"fmt"
	"testing"
)

// LoadString was the old API. The current API is LoadStringNoPEG.
// All tests in this file have been updated to use LoadStringNoPEG.

func TestLoader_load_integer(t *testing.T) {
	input := "123"
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_load_decimal(t *testing.T) {
	input := "123.231"
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.DecimalType {
		t.Error("Expected type decimal")
	}
}

func TestLoader_load_word(t *testing.T) {
	input := "wowo"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.WordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_words(t *testing.T) {
	input := "wowowo wawawa yoyoyo"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.WordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setword(t *testing.T) {
	input := "wowo:"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.SetwordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setwords(t *testing.T) {
	input := "wowo: wawa: wiwi:"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.SetwordType {
		t.Error("Expected type word")
	}
}

func TestLoader_load_setword_check_colon(t *testing.T) {
	input := "wowo:"
	block, _ := LoadStringNoPEG(input, false)
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

// In the current NoPEG parser, .word tokens are DOTWORDS (env.DotwordType), not opwords.
// Only symbolic operators like +, -, *, / are opwords.
// The old PEG parser treated .wowo as an opword; the new parser has a dedicated dotword type.
func TestLoader_load_dotword_1(t *testing.T) {
	input := ".wowo"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.DotwordType {
		t.Error("Expected type Dotword (not Opword — dotwords are now a distinct type)")
	}

	idx := block.(env.Block).Series.Get(0).(env.Dotword).Index

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func TestLoader_load_pipeword_1(t *testing.T) {
	input := "|wowo"
	block, _ := LoadStringNoPEG(input, false)
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
	LoadStringNoPEG(input, false)
}

func TestLoader_load_mixed(t *testing.T) {
	input := "wowo: inc 123"
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_newlines(t *testing.T) {
	input := "\n   123	 \n"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}
	if block.(env.Block).Series.Get(0).Type() != env.IntegerType {
		t.Error("Expected type integer")
	}
}

func TestLoader_multiple_newlines2(t *testing.T) {
	input := "\n\t123	 \nword\nword2\tsetword2:\n\t234"
	block, _ := LoadStringNoPEG(input, false)
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

func TestLoader_bblock(t *testing.T) {
	input := "a: 1 [ a 2 ]"
	block, _ := LoadStringNoPEG(input, false)
	// expect setword, integer, block
	if block.(env.Block).Series.Len() != 3 {
		t.Error("Expected 3 items")
	}
	innerBlock := block.(env.Block).Series.Get(2)
	if innerBlock.Type() != env.BlockType {
		t.Error("Expected type block")
	}
	// block is not evaluated yet, only loaded
	if innerBlock.(env.Block).Series.Get(0).Type() != env.WordType {
		t.Error("Expected first item to be evaluated to type integer but got type " + strconv.Itoa(int(innerBlock.(env.Block).Series.Get(0).Type())))
	}
	if innerBlock.(env.Block).Series.Get(1).Type() != env.IntegerType {
		t.Error("Expected second item to be type integer")
	}
}

func TestLoader_multiple_blocks(t *testing.T) {
	input := "\n\t123	{ { 22 } aa } \nword2\tsetword2:\n\t234"
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
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

// TestLoader_load_argword is DISABLED: The old PEG parser supported {name:kind} as a special
// "argword" token in one pass. The current NoPEG parser requires whitespace around delimiters
// and does not emit argword tokens. The env.Argword type still exists but is no longer produced
// by source parsing. Argwords are now constructed programmatically if needed.
func DISABLED__TestLoader_load_argword(t *testing.T) {
	input := "{somename:somekind}"
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
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
	block, _ := LoadStringNoPEG(input, false)
	block.Trace("CPATH")
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 items")
	}

	//fmt.Println(block.(env.Block).Series.Get(0).Inspect(wordIndex))

	if block.(env.Block).Series.Get(0).Type() != env.CPathType {
		t.Error("Expected type CPath")
	}

	fmt.Println(block.(env.Block).Series.Get(0))

	block.(env.Block).Series.Get(0).(env.CPath).Word1.Print(*wordIndex)

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
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.TagwordType {
		t.Error("Expected type Tagword")
	}

	idx := block.(env.Block).Series.Get(0).(env.Tagword).Index

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

// In the current NoPEG parser, <word> xwords are internally converted to opwords.
// parseToken() for NPEG_TOKEN_XWORD returns env.NewOpword, so <wowo> has type OpwordType.
func TestLoader_load_xword_1(t *testing.T) {
	input := "<wowo>"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	// xwords are parsed as opwords in the current NoPEG parser
	if block.(env.Block).Series.Get(0).Type() != env.OpwordType {
		t.Error("Expected type Opword (xwords <word> are converted to opwords in current parser)")
	}

	idx := block.(env.Block).Series.Get(0).(env.Opword).Index

	fmt.Println(wordIndex.GetWord(idx))

	if wordIndex.GetWord(idx) != "wowo" {
		t.Error("Word spelling not correct")
	}
}

func DISABLED__TestLoader_load_exword_1(t *testing.T) {
	input := "<wowo>"
	block, _ := LoadStringNoPEG(input, false)
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

// TestLoader_load_dotslash verifies the fix: ./ should parse as a dotword (for division operator),
// not as an op-context-path. Reported in step02-results.md, fixed in loader_no_peg.go readOpWord().
func TestLoader_load_dotslash(t *testing.T) {
	input := "./"
	block, _ := LoadStringNoPEG(input, false)
	if block.(env.Block).Series.Len() != 1 {
		t.Error("Expected 1 item")
	}
	if block.(env.Block).Series.Get(0).Type() != env.DotwordType {
		t.Error("Expected type Dotword (./ should be a dotword for the division operator, not a cpath)")
	}

	idx := block.(env.Block).Series.Get(0).(env.Dotword).Index
	word := wordIndex.GetWord(idx)
	if word != "_/" {
		t.Errorf("Expected word '_/' (dotword for division op), got '%s'", word)
	}
}

// TestLoader_load_comments verifies that comments are correctly skipped inside blocks.
func TestLoader_load_comments(t *testing.T) {
	input := "x: 1\n; this is a comment\ny: 2"
	block, _ := LoadStringNoPEG(input, false)
	// Should have: setword x, integer 1, setword y, integer 2
	if block.(env.Block).Series.Len() != 4 {
		t.Errorf("Expected 4 items (comments skipped), got %d", block.(env.Block).Series.Len())
	}
}
