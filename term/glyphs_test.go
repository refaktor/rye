package term

import "testing"

// Glyphs a terminal draws as one character. The first few are one rune each;
// the rest are the ones the old zero-width table could not see.
var glyphCases = []struct {
	name  string
	s     string
	width int
}{
	{"ascii", "a", 1},
	{"cjk", "私", 2},
	{"accented letter", "é", 1},
	{"family", "\U0001F468\u200d\U0001F469\u200d\U0001F467\u200d\U0001F466", 2},
	{"rainbow flag", "\U0001F3F3\ufe0f\u200d\U0001F308", 2},
	{"thumb with a skin tone", "\U0001F44D\U0001F3FD", 2},
	{"flag", "\U0001F1EF\U0001F1F5", 2},
}

func TestCountGlyphsWidth(t *testing.T) {
	for _, c := range glyphCases {
		if got := countGlyphs([]rune(c.s)); got != c.width {
			t.Errorf("%s: countGlyphs(%q) = %d, occupies %d", c.name, c.s, got, c.width)
		}
	}
	// Sentences, to check the glyphs still add up alongside ordinary text.
	for _, c := range []struct {
		s     string
		width int
	}{
		{"query", 5},
		{"hello『世界』", 13},
		{"a\U0001F468\u200d\U0001F469\u200d\U0001F467\u200d\U0001F466b", 4},
	} {
		if got := countGlyphs([]rune(c.s)); got != c.width {
			t.Errorf("countGlyphs(%q) = %d, occupies %d", c.s, got, c.width)
		}
	}
}

// Moving over one glyph must move over all of it. Stopping partway leaves the
// cursor inside the character, and the runes either side of it then get
// edited separately.
func TestGlyphMovementKeepsGlyphsWhole(t *testing.T) {
	for _, c := range glyphCases {
		s := []rune(c.s)
		if got := getPrefixGlyphs(s, 1); len(got) != len(s) {
			t.Errorf("%s: prefix of one glyph took %d of %d runes", c.name, len(got), len(s))
		}
		if got := getSuffixGlyphs(s, 1); len(got) != len(s) {
			t.Errorf("%s: suffix of one glyph took %d of %d runes", c.name, len(got), len(s))
		}

		// One glyph of "a<glyph>b" is the "a" at the front and the "b" at
		// the back, never a piece of the glyph between them.
		padded := append(append([]rune("a"), s...), 'b')
		if got := getPrefixGlyphs(padded, 1); len(got) != 1 {
			t.Errorf("%s: prefix of one glyph of %q took %d runes, wanted 1",
				c.name, string(padded), len(got))
		}
		if got := getSuffixGlyphs(padded, 1); len(got) != 1 {
			t.Errorf("%s: suffix of one glyph of %q took %d runes, wanted 1",
				c.name, string(padded), len(got))
		}
	}
}

func TestGlyphMovementEdges(t *testing.T) {
	s := []rune("hello『世界』")
	if got := getPrefixGlyphs(s, 999); len(got) != len(s) {
		t.Errorf("prefix overflow returned %d of %d runes", len(got), len(s))
	}
	if got := getSuffixGlyphs(s, 999); len(got) != len(s) {
		t.Errorf("suffix overflow returned %d of %d runes", len(got), len(s))
	}
	if got := getPrefixGlyphs(s, -3); len(got) != 0 {
		t.Errorf("negative prefix returned %d runes", len(got))
	}
	if got := getSuffixGlyphs(s, -3); len(got) != 0 {
		t.Errorf("negative suffix returned %d runes", len(got))
	}
	if got := getPrefixGlyphs(nil, 1); len(got) != 0 {
		t.Errorf("prefix of nothing returned %d runes", len(got))
	}
	if got := getSuffixGlyphs(nil, 1); len(got) != 0 {
		t.Errorf("suffix of nothing returned %d runes", len(got))
	}
}
