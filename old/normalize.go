// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package old

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	blankID       = -1
	unknownWordID = -2
)

// htmlesc unescapes HTML escapes that we've observed,
// especially in Markdown-formatted licenses.
// The replacements must have the same length as the original strings
// to preserve byte offsets.
var htmlesc = strings.NewReplacer(
	"&ldquo;", "   \"   ",
	"&rdquo;", "   \"   ",
	"&amp;", "  &  ",
)

// normalize turns the input byte slice into a slice of normalized words
// as a document, including the indexes required to recover the original.
// Normalized text is all lower case, stripped of punctuation and space.
// The slice of normalized words is a slice of indexes into c.words,
// which is updated to add new words as needed.
// Using integer indexes makes the comparison against input texts faster.
func (c *Checker) normalize(data []byte, updateDict bool) *document {
	var r rune
	var wid int
	pos := 0
	str := toLower(data)
	str = htmlesc.Replace(str)
	next := func() {
		r, wid = utf8.DecodeRuneInString(str[pos:])
		pos += wid
	}
	words := make([]int32, 0, 100)
	indexes := make([]int32, 0, 100)
	// Each iteration adds a word.
	for pos < len(str) {
		start := pos
		const blank = "___" // fill in the blank wildcard
		if strings.HasPrefix(str[pos:], blank) {
			words = append(words, blankID)
			indexes = append(indexes, int32(start))
			pos += len(blank)
			continue
		}
		next()
		// Skip spaces, punctuation, etc. and keep only word characters.
		if !isWordChar(r) {
			continue
		}
		// Now at start of word.
		for pos < len(str) {
			next()
			if !isWordChar(r) {
				pos -= wid // Will skip r next time around.
				break
			}
		}
		if pos > start {
			// Is it a list marker? Longest one is maxListMarkerLength bytes: "viii".
			if pos-start > maxListMarkerLength || !isListMarker(str[start:pos], r) { // If at EOF, r will not be valid punctuation
				word := str[start:pos]
				w, ok := c.dict[word]
				if !ok {
					if updateDict {
						w = int32(len(c.words))
						c.words = append(c.words, word)
						c.dict[word] = w
					} else {
						w = unknownWordID
					}
				}
				words = append(words, w)
				indexes = append(indexes, int32(start))
			}
		}
	}
	return &document{
		text:    data,
		words:   words,
		byteOff: indexes,
	}
}

// toLower returns a lowercased version of the input, guaranteeing
// that the size remains the same so byte offsets between the slice and
// the string created from it, which will be used to locate words, will
// line up. TODO: There is a proposal in Go to provide a UTF-8 handler
// that would make this nicer. Use it if it arrives.
// https://github.com/golang/go/issues/25805
func toLower(b []byte) string {
	var s strings.Builder
	for i, wid := 0, 0; i < len(b); i += wid {
		var r rune
		r, wid = utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && wid == 1 {
			// Trouble. Just copy one byte and make it ASCII.
			s.WriteByte('?')
			continue
		}
		l := unicode.ToLower(r)
		if utf8.RuneLen(l) != wid {
			// More trouble. Just use the original.
			l = r
		}
		s.WriteRune(l)
	}
	return s.String()
}

// isWordChar reports whether r is valid in a word. That means it must
// be a letter, although that definition may change. The rune has already
// been case lowered, although that doesn't matter here.
func isWordChar(r rune) bool {
	return unicode.IsLetter(r)
}

const maxListMarkerLength = 4

var listMarker = func() map[string]bool {
	const allListMarkers = "a b c d e f g h i j k l m n o p q r ii iii iv vi vii viii ix xi xii xiii xiv xv"
	l := map[string]bool{}
	for _, marker := range strings.Split(allListMarkers, " ") {
		if len(marker) > maxListMarkerLength {
			panic("marker too long")
		}
		l[marker] = true
	}
	return l
}()

// isListMarker reports whether s, followed immediately by nextRune, is a potential
// list marker such as "i." or "a)".
func isListMarker(s string, nextRune rune) bool {
	if !listMarker[s] {
		return false
	}
	switch nextRune {
	case '.', ':', ')':
		return true
	}
	return false
}
