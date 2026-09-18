package loader

import (
	"testing"

	"github.com/refaktor/rye/env"
)

// Note: LoadStringNoPEG wraps input in "{ input\n}" automatically.

func checkCPathWords(t *testing.T, obj env.Object, idx *env.Idxs, expected []string) {
	t.Helper()
	cp, ok := obj.(env.CPath)
	if !ok {
		t.Fatalf("Expected CPath, got %T", obj)
	}
	if len(cp.Words) != len(expected) {
		t.Fatalf("Expected %d words, got %d (%s)", len(expected), len(cp.Words), cp.Inspect(*idx))
	}
	for i, w := range expected {
		if idx.GetWord(cp.Words[i].Index) != w {
			t.Errorf("Word %d: expected %q, got %q", i, w, idx.GetWord(cp.Words[i].Index))
		}
	}
}

func loadCPathFirstItem(t *testing.T, code string) (env.Object, *env.Idxs) {
	t.Helper()
	result, idx := LoadStringNoPEG(code, false)
	block, ok := result.(env.Block)
	if !ok {
		if err, isErr := result.(env.Error); isErr {
			t.Fatalf("Parser returned error for %q: %s", code, err.Message)
		}
		t.Fatalf("Expected result to be a Block for %q, got %T", code, result)
	}
	if block.Series.Len() != 1 {
		t.Fatalf("Expected 1 item in block for %q, got %d", code, block.Series.Len())
	}
	return block.Series.Get(0), idx
}

func TestCPathParenWordStripsParens(t *testing.T) {
	// (word) -> word
	obj, idx := loadCPathFirstItem(t, "(word)")
	if word, ok := obj.(env.Word); ok {
		if idx.GetWord(word.Index) != "word" {
			t.Errorf("Expected word 'word', got '%s'", idx.GetWord(word.Index))
		}
	} else {
		t.Errorf("Expected Word, got %T (%s)", obj, obj.Inspect(*idx))
	}

	// (word/word) -> cpath word/word
	obj, idx = loadCPathFirstItem(t, "(word/word)")
	checkCPathWords(t, obj, idx, []string{"word", "word"})

	// (a/b/c) -> cpath a/b/c
	obj, idx = loadCPathFirstItem(t, "(a/b/c)")
	checkCPathWords(t, obj, idx, []string{"a", "b", "c"})

	// (word:) -> setword word
	obj, idx = loadCPathFirstItem(t, "(word:)")
	if sw, ok := obj.(env.Setword); ok {
		if idx.GetWord(sw.Index) != "word" {
			t.Errorf("Expected setword 'word', got '%s'", idx.GetWord(sw.Index))
		}
	} else {
		t.Errorf("Expected Setword, got %T (%s)", obj, obj.Inspect(*idx))
	}

	// (word::) -> modword word
	obj, idx = loadCPathFirstItem(t, "(word::)")
	if mw, ok := obj.(env.Modword); ok {
		if idx.GetWord(mw.Index) != "word" {
			t.Errorf("Expected modword 'word', got '%s'", idx.GetWord(mw.Index))
		}
	} else {
		t.Errorf("Expected Modword, got %T (%s)", obj, obj.Inspect(*idx))
	}
}

func TestCPathEmptySegmentsRejected(t *testing.T) {
	cases := []string{"word/word/", "word//word", "@/", "word/@/"}
	for _, code := range cases {
		result, _ := LoadStringNoPEG(code, false)
		if _, ok := result.(env.Error); !ok {
			t.Errorf("Expected parse error for %q, got %T", code, result)
		}
	}
}

func TestCPathParentNavigationMidPath(t *testing.T) {
	// word/@/word should be a cpath with _@ in the middle, not an email
	obj, idx := loadCPathFirstItem(t, "word/@/word")
	checkCPathWords(t, obj, idx, []string{"word", "_@", "word"})

	// @/word
	obj, idx = loadCPathFirstItem(t, "@/word")
	checkCPathWords(t, obj, idx, []string{"_@", "word"})

	// @/@/word
	obj, idx = loadCPathFirstItem(t, "@/@/word")
	checkCPathWords(t, obj, idx, []string{"_@", "_@", "word"})
}

func TestCPathNormalStillWorks(t *testing.T) {
	obj, idx := loadCPathFirstItem(t, "word/word")
	checkCPathWords(t, obj, idx, []string{"word", "word"})

	obj, idx = loadCPathFirstItem(t, "a/b/c")
	checkCPathWords(t, obj, idx, []string{"a", "b", "c"})
}
