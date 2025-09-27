package term

import (
	"strings"
	"unicode"
)

const bright = "\x1b[1m"
const dim = "\x1b[2m"
const black = "\x1b[30m"
const red = "\x1b[31m"
const green = "\x1b[32m"
const yellow = "\x1b[33m"
const blue = "\x1b[34m"
const magenta = "\x1b[35m"
const cyan = "\x1b[36m"
const white = "\x1b[37m"
const reset = "\x1b[0m"
const reset2 = "\033[39;49m"

const color_word1 = cyan
const color_word2 = yellow
const color_num2 = magenta
const color_string2 = green
const color_comment = dim + white
const color_emph = bright

type HighlightedStringBuilder struct {
	b strings.Builder
}

func (h *HighlightedStringBuilder) WriteRune(c rune) {
	h.b.WriteRune(c)
}

func (h *HighlightedStringBuilder) String() string {
	return h.b.String()
}

func (h *HighlightedStringBuilder) ColoredString(inStr bool) string {
	return h.getColor(inStr) + h.b.String() + reset
}

func (h *HighlightedStringBuilder) Reset() {
	h.b.Reset()
}

func (h *HighlightedStringBuilder) getColor(inStr bool) string {
	s := h.b.String()
	if len(s) == 0 {
		return ""
	}

	// If we're explicitly told we're in a string, always color as string
	if inStr {
		return color_string2
	}

	// Not in a string context, check token content
	if strings.HasPrefix(s, ";") {
		return color_comment
	}
	if hasPrefixMultiple(s, "\"", "`") {
		return color_string2
	}
	if strings.HasPrefix(s, "%") && len(s) != 1 {
		return color_string2
	}
	if hasPrefixMultiple(s, "?", "~", "|", "\\", ".", "'", "<") {
		if len(s) != 1 {
			return color_word2
		}
	}
	if strings.HasPrefix(s, ":") {
		if strings.HasPrefix(s, "::") {
			if len(s) != 2 {
				return color_emph + color_word1
			}
		} else if len(s) != 1 {
			return color_word1
		}
	}
	if strings.HasSuffix(s, ":") {
		if strings.HasSuffix(s, "::") {
			if len(s) != 2 {
				return color_emph + color_word1
			}
		} else if len(s) != 1 {
			return color_word1
		}
	}
	if unicode.IsNumber(rune(s[0])) {
		return color_num2
	}
	if unicode.IsLetter(rune(s[0])) {
		if strings.Contains(s, "://") {
			return color_string2
		}
		if strings.HasSuffix(s, "!") || strings.HasPrefix(s, "set-") {
			return color_emph + color_word2
		}
		return color_word2
	}
	return ""
}

func hasPrefixMultiple(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func RyeHighlight(s string, inStr1X bool, inStr2X bool, columns int) (string, bool, bool) {
	var fullB strings.Builder
	var hb HighlightedStringBuilder

	var inComment, inStr1, inStr2 bool
	inStr1 = inStr1X
	inStr2 = inStr2X

	for _, c := range s {
		//if (i+2)%columns == 0 {
		//	hb.WriteRune('\n')
		// hb.WriteRune('\r')
		// }
		if inComment {
			hb.WriteRune(c)
		} else if c == ';' && !inStr1 && !inStr2 {
			inComment = true
			hb.WriteRune(c)
		} else if c == '"' {
			if inStr1 {
				// Closing quote - add quote and finish the string
				hb.WriteRune(c)
				fullB.WriteString(hb.ColoredString(inStr1 || inStr2))
				inStr1 = false
				hb.Reset()
			} else {
				// Opening quote - finish any current token first, then start string
				if hb.b.Len() > 0 {
					fullB.WriteString(hb.ColoredString(inStr1 || inStr2))
					hb.Reset()
				}
				hb.WriteRune(c)
				inStr1 = true
			}
		} else if c == '`' {
			if inStr2 {
				// Closing quote - add quote and finish the string
				hb.WriteRune(c)
				fullB.WriteString(hb.ColoredString(inStr1 || inStr2))
				inStr2 = false
				hb.Reset()
			} else {
				// Opening quote - finish any current token first, then start string
				if hb.b.Len() > 0 {
					fullB.WriteString(hb.ColoredString(inStr1 || inStr2))
					hb.Reset()
				}
				hb.WriteRune(c)
				inStr2 = true
			}
		} else if unicode.IsSpace(c) && !inComment && !inStr1 && !inStr2 {
			fullB.WriteString(hb.ColoredString(inStr1 || inStr2))
			hb.Reset()

			fullB.WriteRune(c)
		} else {
			hb.WriteRune(c)
		}
	}
	fullB.WriteString(hb.ColoredString(inStr1 || inStr2))
	hb.Reset()
	return fullB.String(), inStr1, inStr2
}
