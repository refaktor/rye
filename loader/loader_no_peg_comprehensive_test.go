package loader

import (
	"fmt"
	"testing"

	"github.com/refaktor/rye/env"
)

// Note: LoadStringNoPEG wraps input in "{ input\n}" automatically.
// Do NOT pass already-wrapped "{ ... }" strings - pass raw token sequences.

func TestLoaderNoPEGComprehensive(t *testing.T) {
	// Test code with various token types - NO outer braces
	code := `
		; This is a comment
		word                ; A simple word
		.op-word            ; A dotword (. prefix = dotword, not opword)
		|pipe-word          ; A pipe-word
		|one-char-pipe      ; A pipe-word with one char
		123                 ; An integer
		-456                ; A negative integer
		3.14                ; A decimal
		-2.718              ; A negative decimal
		"string"            ; A string
		'tag-word           ; A tag-word
		?get-word           ; A get-word
		set-word:           ; A set-word
		:lset-word          ; A left set-word
		mod-word::          ; A mod-word
		::lmod-word         ; A left mod-word
		<xword>             ; An x-word (becomes opword)
		</exword>           ; An ex-word
		http://example.com  ; A URI
		user@example.com    ; An email
		%file/path          ; A file path (URI with file scheme)
		context/path        ; A context path
		.op/context/path    ; An op context path (dotword + slash + chars = opcpath)
		|pipe/context/path  ; A pipe context path
		{ nested block }    ; A nested block
		[ bracket block ]   ; A bracket block
		( group )           ; A group
		,                   ; A comma
		_                   ; Void
`

	result, idx := LoadStringNoPEG(code, false)

	block, ok := result.(env.Block)
	if !ok {
		if err, isErr := result.(env.Error); isErr {
			t.Fatalf("Parser returned error: %s", err.Message)
		} else {
			t.Fatalf("Expected result to be a Block, got %T", result)
		}
	}

	fmt.Printf("Block contains %d items\n", block.Series.Len())

	expectedItems := 28
	if block.Series.Len() != expectedItems {
		t.Errorf("Expected %d items in block, got %d", expectedItems, block.Series.Len())
	}

	checkItem := func(index int, expectedType string, expectedValue string) {
		t.Helper()
		if index >= block.Series.Len() {
			t.Errorf("Block doesn't have item at index %d", index)
			return
		}
		item := block.Series.Get(index)
		fmt.Printf("Item %d: ", index)
		switch expectedType {
		case "Word":
			if word, ok := item.(env.Word); ok {
				wordStr := idx.GetWord(word.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Word '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Word '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Word, got %T", index, item)
			}
		case "Dotword":
			if dotword, ok := item.(env.Dotword); ok {
				wordStr := idx.GetWord(dotword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Dotword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Dotword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Dotword, got %T", index, item)
			}
		case "Opword":
			if opword, ok := item.(env.Opword); ok {
				wordStr := idx.GetWord(opword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Opword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Opword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Opword, got %T", index, item)
			}
		case "Pipeword":
			if pipeword, ok := item.(env.Pipeword); ok {
				wordStr := idx.GetWord(pipeword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Pipeword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Pipeword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Pipeword, got %T", index, item)
			}
		case "Integer":
			if num, ok := item.(env.Integer); ok {
				fmt.Printf("Integer %d\n", num.Value)
			} else {
				t.Errorf("Expected item %d to be Integer, got %T", index, item)
			}
		case "Decimal":
			if dec, ok := item.(env.Decimal); ok {
				fmt.Printf("Decimal %f\n", dec.Value)
			} else {
				t.Errorf("Expected item %d to be Decimal, got %T", index, item)
			}
		case "String":
			if str, ok := item.(env.String); ok {
				if str.Value != expectedValue {
					t.Errorf("Expected item %d to be String '%s', got '%s'", index, expectedValue, str.Value)
				}
				fmt.Printf("String '%s'\n", str.Value)
			} else {
				t.Errorf("Expected item %d to be String, got %T", index, item)
			}
		case "Tagword":
			if tagword, ok := item.(env.Tagword); ok {
				wordStr := idx.GetWord(tagword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Tagword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Tagword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Tagword, got %T", index, item)
			}
		case "Getword":
			if getword, ok := item.(env.Getword); ok {
				wordStr := idx.GetWord(getword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Getword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Getword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Getword, got %T", index, item)
			}
		case "Setword":
			if setword, ok := item.(env.Setword); ok {
				wordStr := idx.GetWord(setword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Setword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Setword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Setword, got %T", index, item)
			}
		case "LSetword":
			if lsetword, ok := item.(env.LSetword); ok {
				wordStr := idx.GetWord(lsetword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be LSetword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("LSetword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be LSetword, got %T", index, item)
			}
		case "Modword":
			if modword, ok := item.(env.Modword); ok {
				wordStr := idx.GetWord(modword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Modword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Modword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Modword, got %T", index, item)
			}
		case "LModword":
			if lmodword, ok := item.(env.LModword); ok {
				wordStr := idx.GetWord(lmodword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be LModword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("LModword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be LModword, got %T", index, item)
			}
		case "Xword":
			// xwords are parsed as Opword
			if opword, ok := item.(env.Opword); ok {
				wordStr := idx.GetWord(opword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be Opword (from x-word) '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("Opword (xword) '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be Opword (from x-word), got %T", index, item)
			}
		case "EXword":
			if exword, ok := item.(env.EXword); ok {
				wordStr := idx.GetWord(exword.Index)
				if wordStr != expectedValue {
					t.Errorf("Expected item %d to be EXword '%s', got '%s'", index, expectedValue, wordStr)
				}
				fmt.Printf("EXword '%s'\n", wordStr)
			} else {
				t.Errorf("Expected item %d to be EXword, got %T", index, item)
			}
		case "Uri":
			if _, ok := item.(env.Uri); ok {
				fmt.Printf("Uri\n")
			} else {
				t.Errorf("Expected item %d to be Uri, got %T", index, item)
			}
		case "Email":
			if _, ok := item.(env.Email); ok {
				fmt.Printf("Email\n")
			} else {
				t.Errorf("Expected item %d to be Email, got %T", index, item)
			}
		case "Uri-File":
			if _, ok := item.(env.Uri); ok {
				fmt.Printf("File Uri\n")
			} else {
				t.Errorf("Expected item %d to be Uri (file), got %T", index, item)
			}
		case "CPath":
			if _, ok := item.(env.CPath); ok {
				fmt.Printf("CPath\n")
			} else {
				t.Errorf("Expected item %d to be CPath, got %T", index, item)
			}
		case "Block":
			if blk, ok := item.(env.Block); ok {
				fmt.Printf("Block with %d items\n", blk.Series.Len())
			} else {
				t.Errorf("Expected item %d to be Block, got %T", index, item)
			}
		case "Comma":
			if _, ok := item.(env.Comma); ok {
				fmt.Printf("Comma\n")
			} else {
				t.Errorf("Expected item %d to be Comma, got %T", index, item)
			}
		case "Void":
			if _, ok := item.(env.Void); ok {
				fmt.Printf("Void\n")
			} else {
				t.Errorf("Expected item %d to be Void, got %T", index, item)
			}
		default:
			fmt.Printf("Unknown type: %T\n", item)
		}
	}

	checkItem(0, "Word", "word")
	checkItem(1, "Dotword", "op-word") // .op-word → Dotword
	checkItem(2, "Pipeword", "pipe-word")
	checkItem(3, "Pipeword", "one-char-pipe")
	checkItem(4, "Integer", "123")
	checkItem(5, "Integer", "-456")
	checkItem(6, "Decimal", "3.14")
	checkItem(7, "Decimal", "-2.718")
	checkItem(8, "String", "string")
	checkItem(9, "Tagword", "tag-word")
	checkItem(10, "Getword", "get-word")
	checkItem(11, "Setword", "set-word")
	checkItem(12, "LSetword", "lset-word")
	checkItem(13, "Modword", "mod-word")
	checkItem(14, "LModword", "lmod-word")
	checkItem(15, "Xword", "xword")
	checkItem(16, "EXword", "exword")
	checkItem(17, "Uri", "http://example.com")
	checkItem(18, "Email", "user@example.com")
	checkItem(19, "Uri-File", "file://file/path")
	checkItem(20, "CPath", "context/path")
	checkItem(21, "CPath", ".op/context/path") // .word/path → opcpath (has word chars before /)
	checkItem(22, "CPath", "|pipe/context/path")
	checkItem(23, "Block", "")
	checkItem(24, "Block", "")
	checkItem(25, "Block", "")
	checkItem(26, "Comma", "")
	checkItem(27, "Void", "")
}
