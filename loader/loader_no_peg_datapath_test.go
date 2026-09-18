package loader

import (
	"testing"

	"github.com/refaktor/rye/env"
)

// Note: LoadStringNoPEG wraps input in "{ input\n}" automatically.
// Do NOT pass already-wrapped "{ ... }" strings - pass raw token sequences.

func checkDataPathSegments(t *testing.T, obj env.Object, idx *env.Idxs, expected []env.Object) {
	t.Helper()
	dp, ok := obj.(env.DataPath)
	if !ok {
		t.Fatalf("Expected DataPath, got %T", obj)
	}
	if len(dp.Path) != len(expected) {
		t.Fatalf("Expected %d segments, got %d (%s)", len(expected), len(dp.Path), dp.Inspect(*idx))
	}
	for i, seg := range dp.Path {
		if !seg.Equal(expected[i]) {
			t.Errorf("Segment %d: expected %s (%T), got %s (%T)",
				i, expected[i].Inspect(*idx), expected[i], seg.Inspect(*idx), seg)
		}
	}
}

func loadFirstItem(t *testing.T, code string) (env.Object, *env.Idxs) {
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

func TestDataPathParsing(t *testing.T) {
	// word.word -> subject word + word key
	obj, idx := loadFirstItem(t, "word.name")
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("word")),
		*env.NewWord(idx.IndexWord("name")),
	})

	// word.integer -> subject word + integer index
	obj, idx = loadFirstItem(t, "person.0")
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewInteger(0),
	})

	// word.integer."string" -> the full example from the request
	obj, idx = loadFirstItem(t, `person.0."age"`)
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewInteger(0),
		*env.NewString("age"),
	})

	// word.word.integer."string" -> mixed segments
	obj, idx = loadFirstItem(t, `word.lit-word.5."string"`)
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("word")),
		*env.NewWord(idx.IndexWord("lit-word")),
		*env.NewInteger(5),
		*env.NewString("string"),
	})

	// word."string" -> accessing by string key (distinct from word key)
	obj, idx = loadFirstItem(t, `person."name"`)
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewString("name"),
	})

	// deep nesting
	obj, idx = loadFirstItem(t, `person.0.name."x"`)
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewInteger(0),
		*env.NewWord(idx.IndexWord("name")),
		*env.NewString("x"),
	})

	// negative index
	obj, idx = loadFirstItem(t, "person.-1")
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewInteger(-1),
	})

	// string segment containing a dot must not be split
	obj, idx = loadFirstItem(t, `person."a.b"`)
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewString("a.b"),
	})

	// string segment with spaces
	obj, idx = loadFirstItem(t, `person."full name"`)
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewString("full name"),
	})

	// backtick string segment
	obj, idx = loadFirstItem(t, "person.`age`")
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("person")),
		*env.NewString("age"),
	})

	// three word segments
	obj, idx = loadFirstItem(t, "a.b.c")
	checkDataPathSegments(t, obj, idx, []env.Object{
		*env.NewWord(idx.IndexWord("a")),
		*env.NewWord(idx.IndexWord("b")),
		*env.NewWord(idx.IndexWord("c")),
	})
}

func TestDataPathDoesNotBreakExistingTokens(t *testing.T) {
	// decimals are unaffected
	obj, idx := loadFirstItem(t, "3.14")
	if _, ok := obj.(env.Decimal); !ok {
		t.Errorf("Expected Decimal, got %T (%s)", obj, obj.Inspect(*idx))
	}

	// range operator '..' is still an opword
	obj, idx = loadFirstItem(t, "..")
	if op, ok := obj.(env.Opword); ok {
		if idx.GetWord(op.Index) != "_.." {
			t.Errorf("Expected opword '_..', got '%s'", idx.GetWord(op.Index))
		}
	} else {
		t.Errorf("Expected Opword, got %T (%s)", obj, obj.Inspect(*idx))
	}

	// context paths are unaffected
	obj, idx = loadFirstItem(t, "word/word")
	if _, ok := obj.(env.CPath); !ok {
		t.Errorf("Expected CPath, got %T (%s)", obj, obj.Inspect(*idx))
	}

	// dotwords are unaffected
	result, _ := LoadStringNoPEG("1 .add 2", false)
	block := result.(env.Block)
	if block.Series.Len() != 3 {
		t.Errorf("Expected 3 items for '1 .add 2', got %d", block.Series.Len())
	} else {
		if _, ok := block.Series.Get(0).(env.Integer); !ok {
			t.Errorf("Expected Integer, got %T", block.Series.Get(0))
		}
		if _, ok := block.Series.Get(1).(env.Dotword); !ok {
			t.Errorf("Expected Dotword, got %T", block.Series.Get(1))
		}
		if _, ok := block.Series.Get(2).(env.Integer); !ok {
			t.Errorf("Expected Integer, got %T", block.Series.Get(2))
		}
	}

	// a lone word is still a plain word (not a data path)
	obj, idx = loadFirstItem(t, "word")
	if _, ok := obj.(env.Word); !ok {
		t.Errorf("Expected Word, got %T (%s)", obj, obj.Inspect(*idx))
	}
}

func TestDataPathErrors(t *testing.T) {
	// trailing dot followed by nothing is not a data path: the '.' is an opword
	// so "person." becomes word + opword (2 items), not a data path.
	result, _ := LoadStringNoPEG("person.", false)
	if block, ok := result.(env.Block); ok {
		if block.Series.Len() != 2 {
			t.Errorf("Expected 2 items for 'person.', got %d", block.Series.Len())
		}
	} else if err, isErr := result.(env.Error); isErr {
		t.Errorf("Unexpected error for 'person.': %s", err.Message)
	}

	// invalid segment after '.' should produce an error
	result, _ = LoadStringNoPEG("person.@", false)
	if _, ok := result.(env.Error); !ok {
		t.Errorf("Expected error for 'person.@', got %T", result)
	}
}
