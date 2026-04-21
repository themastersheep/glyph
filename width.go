package glyph

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Single control point for text measurement. Wrapping stdlib and third-party
// calls lets us change width semantics in one place (e.g. emoji table version,
// ambiguous-width handling, terminal-specific overrides) without chasing call
// sites. All wrappers are leaf functions the compiler inlines freely.

// RuneWidth returns the display-cell width of a rune. ASCII is width 1 on the
// fast path; anything above U+1100 defers to the runewidth table. Zero-width
// runes (combining marks) are reported as 1 because terminals treat them as
// advancing the cursor by 1 in most cases.
func RuneWidth(r rune) int {
	if r < 0x1100 {
		return 1
	}
	w := runewidth.RuneWidth(r)
	if w == 0 {
		return 1
	}
	return w
}

// StringWidth returns the total display-cell width of s.
func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

// RuneCount returns the number of runes in s. Use this when you need a
// character count (e.g. indexing into a rune slice), not a display width.
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}
