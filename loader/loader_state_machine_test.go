//go:build stm_loader

package loader

import (
	"fmt"
	"testing"

	"github.com/refaktor/rye/env"
)

func TestLoaderStateMachine(t *testing.T) {
	// Test code
	code := `{ word .op-word 123 "string asd" { } set-word: }`

	// Parse with the state machine parser
	result, idx := LoadStringStateMachine(code, false)

	// Check if result is a block
	block, ok := result.(env.Block)
	if !ok {
		t.Fatalf("Expected result to be a Block, got %T", result)
	}

	// Print the block contents
	fmt.Printf("Block contains %d items\n", block.Series.Len())

	// Check each item in the block
	if block.Series.Len() != 6 {
		t.Errorf("Expected 6 items in block, got %d", block.Series.Len())
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

	// Check the second item (.op-word)
	if opword, ok := block.Series.Get(1).(env.Opword); ok {
		wordStr := idx.GetWord(opword.Index)
		fmt.Printf("Item 1: Opword '%s'\n", wordStr)
	} else {
		t.Errorf("Expected item 1 to be Opword, got %T", block.Series.Get(1))
	}

	// Check the third item (123)
	if num, ok := block.Series.Get(2).(env.Integer); ok {
		if num.Value != 123 {
			t.Errorf("Expected third item to be 123, got %d", num.Value)
		}
		fmt.Printf("Item 2: Integer %d\n", num.Value)
	} else {
		t.Errorf("Expected item 2 to be Integer, got %T", block.Series.Get(2))
	}

	// Check the fourth item ("string asd")
	if str, ok := block.Series.Get(3).(env.String); ok {
		if str.Value != "string asd" {
			t.Errorf("Expected fourth item to be 'string asd', got '%s'", str.Value)
		}
		fmt.Printf("Item 3: String '%s'\n", str.Value)
	} else {
		t.Errorf("Expected item 3 to be String, got %T", block.Series.Get(3))
	}

	// Check the fifth item ({ })
	if innerBlock, ok := block.Series.Get(4).(env.Block); ok {
		if innerBlock.Series.Len() != 0 {
			t.Errorf("Expected fifth item to be empty block, got block with %d items", innerBlock.Series.Len())
		}
		fmt.Printf("Item 4: Empty Block\n")
	} else {
		t.Errorf("Expected item 4 to be Block, got %T", block.Series.Get(4))
	}

	// Check the sixth item (set-word:)
	if setword, ok := block.Series.Get(5).(env.Setword); ok {
		wordStr := idx.GetWord(setword.Index)
		if wordStr != "set-word" {
			t.Errorf("Expected sixth item to be 'set-word', got '%s'", wordStr)
		}
		fmt.Printf("Item 5: Setword '%s'\n", wordStr)
	} else {
		t.Errorf("Expected item 5 to be Setword, got %T", block.Series.Get(5))
	}
}
