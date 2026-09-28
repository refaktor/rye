package loader

import (
	"strings"
	"testing"

	"github.com/refaktor/rye/env"
)

func TestLoaderNoPEGSpacing(t *testing.T) {
	assertSpacingError := func(t *testing.T, input string) {
		t.Helper()
		result, _ := LoadStringNoPEG(input, false)
		if err, ok := result.(env.Error); !ok || !strings.Contains(err.Message, "Missing space") {
			t.Errorf("expected missing-space error for %q, got %v", input, result)
		}
	}

	for _, input := range []string{"123+123", "123abc", "word;comment", `"text"word`, "foo|bar"} {
		t.Run(input, func(t *testing.T) { assertSpacingError(t, input) })
	}
}

func TestLoaderNoPEGCollectionPrefixes(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int
	}{
		{"L{1 2}", NPEG_TOKEN_LIST_BLOCK_START},
		{"L[1 2]", NPEG_TOKEN_LIST_BBLOCK_START},
		{"D{name 1}", NPEG_TOKEN_DICT_BLOCK_START},
		{"D['name 1]", NPEG_TOKEN_DICT_BBLOCK_START},
	} {
		t.Run(tc.input, func(t *testing.T) {
			lex := NewLexer(tc.input)
			if tok := lex.NextToken(); tok.Type != tc.want {
				t.Errorf("first token of %q: got %v, want %v", tc.input, tok, tc.want)
			}
		})
	}

	// Lowercase is no longer a collection prefix, even without a space.
	for _, prefix := range []string{"l{", "l[", "d{", "d["} {
		t.Run(prefix, func(t *testing.T) {
			lex := NewLexer(prefix)
			if tok := lex.NextToken(); tok.Type != NPEG_TOKEN_WORD || tok.Value != prefix[:1] {
				t.Errorf("first token of %q: got %v, want ordinary word", prefix, tok)
			}
			if tok := lex.NextToken(); tok.Type != NPEG_TOKEN_BLOCK_START && tok.Type != NPEG_TOKEN_BBLOCK_START {
				t.Errorf("second token of %q: got %v, want ordinary delimiter", prefix, tok)
			}
		})
	}

	// Whitespace separates the uppercase prefix from the delimiter too.
	for _, input := range []string{"L {", "L [", "D {", "D ["} {
		lex := NewLexer(input)
		if tok := lex.NextToken(); tok.Type != NPEG_TOKEN_WORD {
			t.Errorf("first token of %q: got %v, want ordinary word", input, tok)
		}
	}
}

func TestLoaderNoPEGStandaloneDelimiters(t *testing.T) {
	// Compare the complete parsed representation, not just whether parsing succeeds.
	pairs := [][2]string{
		{"try(10 + 200)", "try ( 10 + 200 )"},
		{"a{b}c", "a { b } c"},
		{"a[b]c", "a [ b ] c"},
		{"a(b)c", "a ( b ) c"},
		{"(word)", "( word )"},
		{"1234,1231", "1234 , 1231"},
		{`"hi",(2)`, `"hi" , ( 2 )`},
		{"word:({42})", "word: ( { 42 } )"},
		{"person.age,?value", "person.age , ?value"},
		{"foo/bar,10", "foo/bar , 10"},
		{"L{1 2}D[3]", "L{ 1 2 } D[ 3 ]"},
		{".[1].(2).{3}", ".[ 1 ] .( 2 ) .{ 3 }"},
	}
	for _, pair := range pairs {
		t.Run(pair[0], func(t *testing.T) {
			compact, idx := LoadStringNoPEG(pair[0], false)
			spaced, spacedIdx := LoadStringNoPEG(pair[1], false)
			cb, ok := compact.(env.Block)
			if !ok {
				t.Fatalf("compact %q: %v", pair[0], compact)
			}
			sb, ok := spaced.(env.Block)
			if !ok {
				t.Fatalf("spaced %q: %v", pair[1], spaced)
			}
			if cb.Series.Len() != sb.Series.Len() {
				t.Fatalf("different item counts: %d vs %d", cb.Series.Len(), sb.Series.Len())
			}
			for i := 0; i < cb.Series.Len(); i++ {
				got := cb.Series.Get(i).Inspect(*idx)
				want := sb.Series.Get(i).Inspect(*spacedIdx)
				if got != want {
					t.Errorf("item %d: %s != %s", i, got, want)
				}
			}
		})
	}
}
