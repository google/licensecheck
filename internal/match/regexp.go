// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Exported LRE interface.

package match

import (
	"sync"
)

// An LRE is a compiled license regular expression.
//
// TODO: Move this comment somewhere non-internal later.
//
// A license regular expression (LRE) is a pattern syntax intended for
// describing large English texts such as software licenses, with minor
// allowed variations. The pattern syntax and the matching are word-based
// and case-insensitive; punctuation is ignored in the pattern and in
// the matched text.
//
// The valid LRE patterns are:
//
//	word            - a single case-insensitive word
//	__N__           - any sequence of up to N words
//	expr1 expr2     - concatenation
//	expr1 || expr2  - alternation
//	(( expr ))      - grouping
//	expr??          - zero or one instances of expr
//	//** text **//  - a comment
//
// To make patterns harder to misread in large texts:
//
//	- || must only appear inside (( ))
//	- ?? must only follow (( ))
//	- (( must be at the start of a line, preceded only by spaces
//	- )) must be at the end of a line, followed only by spaces and ??.
//
// For example:
//
//	//** https://en.wikipedia.org/wiki/Filler_text **//
//	Now is
//	((not))??
//	the time for all good
//	((men || women || people))
//	to come to the aid of their __1__.
//
type LRE struct {
	dict   *Dict
	file   string
	syntax *reSyntax

	onceDFA sync.Once
	dfa     reDFA
}

// ParseLRE parses the string s as a license regexp.
// The file name is used in error messages if non-empty.
func ParseLRE(d *Dict, file, s string) (*LRE, error) {
	syntax, err := reParse(d, s, true)
	if err != nil {
		return nil, err
	}
	return &LRE{dict: d, file: file, syntax: syntax}, nil
}

// Dict returns the Dict used by the LRE.
func (re *LRE) Dict() *Dict {
	return re.dict
}

// File returns the file name passed to ParseLRE.
func (re *LRE) File() string {
	return re.file
}

// Match reports whether text matches the license regexp.
func (re *LRE) match(text string) bool {
	re.onceDFA.Do(re.compile)
	match, _ := re.dfa.match(re.dict, text, re.dict.Split(text))
	return match >= 0
}

// compile initializes lre.dfa.
// It is invoked lazily (in Match) because most LREs end up only
// being inputs to a MultiLRE; we never need their DFAs directly.
func (re *LRE) compile() {
	re.dfa = reCompileDFA(re.syntax.compile(nil, 0))
}
