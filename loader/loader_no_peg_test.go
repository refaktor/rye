package loader

import (
	"fmt"
	"testing"

	"github.com/refaktor/rye/env"
)

// Note: LoadStringNoPEG wraps the input in "{ input\n}" automatically.
// Do NOT pass already-wrapped "{ ... }" strings to it - pass raw token sequences.

func TestLoaderNoPEG(t *testing.T) {
	// Test code - no outer braces; LoadStringNoPEG wraps it
	code := `word .op-word <tag> 123 "string asd" { } set-word:`

	// Parse with the non-PEG loader
	result, idx := LoadStringNoPEG(code, false)

	// Check if result is a block
	block, ok := result.(env.Block)
	if !ok {
		t.Fatalf("Expected result to be a Block, got %T", result)
	}

	// Print the block contents
	fmt.Printf("Block contains %d items\n", block.Series.Len())

	// Check each item in the block
	if block.Series.Len() != 7 {
		t.Errorf("Expected 7 items in block, got %d", block.Series.Len())
	}

	// Check the first item (word)
	if word, ok := block.Series.Get(0).(env.Word); ok {
		wordStr := idx.GetWord(word.Index)
		if wordStr != "word" {
			t.Errorf("Expected first item to be 'word', got '%s'", wordStr)
		}
		fmt.Printf("Item 0: Word '%s'\n", wordStr)
	} else {
		t.Errorf("Expected item 0 to be Word, got %T", block.Series.Get(0))
	}

	// Check the second item (.op-word) - in current NoPEG parser, .word is a Dotword (not Opword)
	if dotword, ok := block.Series.Get(1).(env.Dotword); ok {
		wordStr := idx.GetWord(dotword.Index)
		if wordStr != "op-word" {
			t.Errorf("Expected second item to be 'op-word', got '%s'", wordStr)
		}
		fmt.Printf("Item 1: Dotword '%s'\n", wordStr)
	} else {
		t.Errorf("Expected item 1 to be Dotword, got %T", block.Series.Get(1))
	}

	// Check the third item (<tag> as opword - xwords become opwords)
	if opword, ok := block.Series.Get(2).(env.Opword); ok {
		wordStr := idx.GetWord(opword.Index)
		if wordStr != "tag" {
			t.Errorf("Expected third item to be 'tag', got '%s'", wordStr)
		}
		fmt.Printf("Item 2: Opword '%s'\n", wordStr)
	} else {
		t.Errorf("Expected item 2 to be Opword, got %T", block.Series.Get(2))
	}

	// Check the fourth item (123)
	if num, ok := block.Series.Get(3).(env.Integer); ok {
		if num.Value != 123 {
			t.Errorf("Expected fourth item to be 123, got %d", num.Value)
		}
		fmt.Printf("Item 3: Integer %d\n", num.Value)
	} else {
		t.Errorf("Expected item 3 to be Integer, got %T", block.Series.Get(3))
	}

	// Check the fifth item ("string asd")
	if str, ok := block.Series.Get(4).(env.String); ok {
		if str.Value != "string asd" {
			t.Errorf("Expected fifth item to be 'string asd', got '%s'", str.Value)
		}
		fmt.Printf("Item 4: String '%s'\n", str.Value)
	} else {
		t.Errorf("Expected item 4 to be String, got %T", block.Series.Get(4))
	}

	// Check the sixth item ({ })
	if innerBlock, ok := block.Series.Get(5).(env.Block); ok {
		if innerBlock.Series.Len() != 0 {
			t.Errorf("Expected sixth item to be empty block, got block with %d items", innerBlock.Series.Len())
		}
		fmt.Printf("Item 5: Empty Block\n")
	} else {
		t.Errorf("Expected item 5 to be Block, got %T", block.Series.Get(5))
	}

	// Check the seventh item (set-word:)
	if setword, ok := block.Series.Get(6).(env.Setword); ok {
		wordStr := idx.GetWord(setword.Index)
		if wordStr != "set-word" {
			t.Errorf("Expected seventh item to be 'set-word', got '%s'", wordStr)
		}
		fmt.Printf("Item 6: Setword '%s'\n", wordStr)
	} else {
		t.Errorf("Expected item 6 to be Setword, got %T", block.Series.Get(6))
	}
}

func TestLoaderNoPEGEndLineComment(t *testing.T) {
	// Note: this input has no outer braces - LoadStringNEWNoPEG wraps it.
	// Comments are stripped; the result has 4 tokens: x: 1 y: x
	input := "x: 1\n;  if x > 4 {\n    y: x\n;  }"

	ps := env.NewProgramState()
	result := LoadStringNEWNoPEG(input, false, ps)

	block, ok := result.(env.Block)
	if !ok {
		if err, isErr := result.(env.Error); isErr {
			t.Fatalf("Parser returned error: %s", err.Message)
		}
		t.Fatalf("Expected result to be a Block, got %T", result)
	}

	// x: 1 y: x = 4 tokens (comments are stripped)
	if block.Series.Len() != 4 {
		t.Fatalf("Expected 4 items in block (x: 1 y: x), got %d", block.Series.Len())
	}
}

func TestLoaderNoPEGXwordForce(t *testing.T) {
	// No outer braces - LoadStringNoPEG wraps it
	code := `<force*>`

	result, idx := LoadStringNoPEG(code, false)
	block, ok := result.(env.Block)
	if !ok {
		t.Fatalf("Expected result to be a Block, got %T", result)
	}

	if block.Series.Len() != 1 {
		t.Fatalf("Expected 1 item in block, got %d", block.Series.Len())
	}

	opword, ok := block.Series.Get(0).(env.Opword)
	if !ok {
		t.Fatalf("Expected item 0 to be Opword, got %T", block.Series.Get(0))
	}

	wordStr := idx.GetWord(opword.Index)
	if wordStr != "force" {
		t.Errorf("Expected opword 'force', got '%s'", wordStr)
	}
	if opword.Force != 1 {
		t.Errorf("Expected opword force flag to be 1, got %d", opword.Force)
	}
}
