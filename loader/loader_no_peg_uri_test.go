package loader

import (
	"testing"

	"github.com/refaktor/rye/env"
)

// Note: inputs have NO outer braces; LoadStringNoPEG wraps them.

func TestLoaderNoPEGURI(t *testing.T) {
	t.Run("URI parsing", func(t *testing.T) {
		input := `http://example.com`
		t.Logf("Testing input: %s", input)
		result, _ := LoadStringNoPEG(input, false)
		t.Logf("Result type: %T", result)

		block, ok := result.(env.Block)
		if !ok {
			t.Errorf("Expected result to be a Block, got %T", result)
			return
		}
		if block.Series.Len() != 1 {
			t.Errorf("Expected block to contain 1 item, got %d", block.Series.Len())
			return
		}
		uri, ok := block.Series.Get(0).(env.Uri)
		if !ok {
			t.Errorf("Expected item to be a URI, got %T", block.Series.Get(0))
			return
		}
		// uri.Path is the part after ://
		if uri.Path != "example.com" {
			t.Errorf("Expected URI path to be 'example.com', got '%s'", uri.Path)
		}
	})

	t.Run("Word followed by URI", func(t *testing.T) {
		input := `word http://example.com`
		t.Logf("Testing input: %s", input)
		result, _ := LoadStringNoPEG(input, false)

		block, ok := result.(env.Block)
		if !ok {
			t.Errorf("Expected result to be a Block, got %T", result)
			return
		}
		if block.Series.Len() != 2 {
			t.Errorf("Expected block to contain 2 items, got %d", block.Series.Len())
			return
		}
		uri, ok := block.Series.Get(1).(env.Uri)
		if !ok {
			t.Errorf("Expected second item to be a URI, got %T", block.Series.Get(1))
			return
		}
		if uri.Path != "example.com" {
			t.Errorf("Expected URI path to be 'example.com', got '%s'", uri.Path)
		}
	})
}
