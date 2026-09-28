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

func TestCPathParenWordIsGroup(t *testing.T) {
	// Parentheses are always group delimiters, even around a single word.
	group, idx := loadCPathFirstItem(t, "(word)")
	obj, ok := group.(env.Block)
	if !ok || obj.Series.Len() != 1 {
		t.Fatalf("Expected one-item group, got %T", group)
	}
	wordObj := obj.Series.Get(0)
	if word, ok := wordObj.(env.Word); ok {
		if idx.GetWord(word.Index) != "word" {
			t.Errorf("Expected word 'word', got '%s'", idx.GetWord(word.Index))
		}
	} else {
		t.Errorf("Expected Word in group, got %T (%s)", wordObj, wordObj.Inspect(*idx))
	}

	// The inner tokens retain their ordinary types.
	for _, tc := range []struct {
		code  string
		words []string
	}{
		{"(word/word)", []string{"word", "word"}},
		{"(a/b/c)", []string{"a", "b", "c"}},
	} {
		group, idx = loadCPathFirstItem(t, tc.code)
		obj = group.(env.Block)
		checkCPathWords(t, obj.Series.Get(0), idx, tc.words)
	}

	group, idx = loadCPathFirstItem(t, "(word:)")
	obj = group.(env.Block)
	if sw, ok := obj.Series.Get(0).(env.Setword); ok {
		if idx.GetWord(sw.Index) != "word" {
			t.Errorf("Expected setword 'word', got '%s'", idx.GetWord(sw.Index))
		}
	} else {
		t.Errorf("Expected Setword, got %T (%s)", obj, obj.Inspect(*idx))
	}

	group, idx = loadCPathFirstItem(t, "(word::)")
	obj = group.(env.Block)
	if mw, ok := obj.Series.Get(0).(env.Modword); ok {
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
