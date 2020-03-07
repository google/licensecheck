// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package match

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
