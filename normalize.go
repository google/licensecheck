// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

// normalize turns the input byte slice into a slice of normalized words
// as a document, including the indexes required to recover the original.
// Normalized text is all lower case, stripped of punctuation and space.
// The slice of normalized words is a slice of indexes into c.words,
// which is updated to add new words as needed.
// Using integer indexes makes the comparison against input texts faster.
func (c *Checker) normalize(data []byte) *document {
	var r rune
	var wid int
	pos := 0
	data = removeCopyrightLines(data)
	str := toLower(data)
	next := func() {
		r, wid = utf8.DecodeRuneInString(str[pos:])
		pos += wid
	}
	words := make([]int32, 0, 100)
	indexes := make([]int32, 0, 100)
	// Each iteration adds a word.
	for pos < len(str) {
		start := pos
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
					w = int32(len(c.words))
					c.words = append(c.words, word)
					c.dict[word] = w
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

var copyrightText = []byte("\nCopyright ")

// removeCopyrightLines returns its argument text with (nearly) all lines beginning
// with the word Copyright deleted. (The exception is for the lines in the Creative
// Commons licenses that are a definition of Copyright.) Leading spaces are
// significant: the line must start with a 'C'. This cleanup eliminates a common
// difference between standard license text and the form of the license seen in
// practice. If a copyright line is deleted, the return value is a fresh copy to
// avoid overwriting the caller's data.
func removeCopyrightLines(text []byte) []byte {
	copied := false
	for i := 0; ; {
		copyright := copyrightText
		if i == 0 {
			copyright = copyright[1:] // Drop leading newline
		}
		start := bytes.Index(text[i:], copyright)
		if start < 0 {
			break
		}
		if i > 0 {
			start += i + 1 // Skip starting newline.
		}
		newline := bytes.IndexByte(text[start:], '\n')
		if newline < 0 {
			break
		}
		newline = start + newline // Leave trailing newline, making it line of blanks.
		i = newline
		// Special case for the Creative Commons licenses, which define copyright.
		// TODO: Better ideas?
		if bytes.Contains(text[start:newline], []byte(" means copyright ")) {
			continue
		}
		if !copied {
			text = append([]byte(nil), text...)
			copied = true
		}
		// White out the text.
		for j := start; j < newline; j++ {
			text[j] = ' '
		}
	}
	return text
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
