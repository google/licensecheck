// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package licensecheck classifies license files and heuristically determines
// how well they correspond to known open source licenses.
//
// Scanning
//
// A text (a slice of bytes) can be scanned for known licenses by calling Scan.
// The resulting Coverage structure describes all the matches found as well
// as what percentage of the file was covered by known matches.
//
//	cov := licensecheck.Scan(text)
//	fmt.Printf("%.1f%% of text covered by licenses:\n", cov.Percent)
//	for _, m := range cov.Match {
//		fmt.Printf("%s at [%d:%d] IsURL=%v\n", m.Name, m.Start, m.End, m.IsURL)
//	}
//
// The Scan function uses a built-in license set, which is the known SPDX licenses
// augmented with some other commonly seen licenses.
// (See licenses/README.md for details about the license set.)
//
// A custom scanner can be created using NewScanner, passing in a set of
// license patterns to scan for. The license patterns are written as license regular expressions (LREs).
// BuiltinLicenses returns the set of license patterns used by Scan.
//
// License Regular Expressions
// Each license to be recognized is specified by writing a license regular expression (LRE) for it.
// The pattern syntax and the matching are word-based and case-insensitive;
// punctuation is ignored in the pattern and in the matched text.
//
// The valid LRE patterns are:
//
//  - word, a single case-insensitive word
//  - __N__, any sequence of up to N words
//  - expr1 expr2, concatenation of two expressions
//  - expr1 || expr2, alternation of two expressions
//  - (( expr )), grouping
//  - (( expr ))??, zero or one instances of the grouped expression
//  - //** text **//, a comment ignored by the parser
//
// To make patterns harder to misread in large texts:
// (( must only appear at the start of a line (possibly indented);
// )) and ))?? must only appear at the end of a line (with possible trailing spaces);
// and || must only appear inside a (( )) or (( ))?? group.
//
// For example:
//
// 	//** https://en.wikipedia.org/wiki/Filler_text **//
// 	Now is
// 	((not))??
// 	the time for all good
// 	((men || women || people))
// 	to come to the aid of their __1__.
//
// The Old Cover and Checker API
//
// An older, less precise matcher using the names Cover, New, and Checker
// was removed from this package.
// Use v0.1.0 for the final version of that API.
//
package licensecheck

import (
	"strings"
)

// The order matters here so everything typechecks for the tools, which are fussy.
//go:generate rm -f data.gen.go
//go:generate stringer -type Type
//go:generate go run gen_data.go

// Type groups the licenses into various classifications.
// TODO: This list is clearly incomplete.
type Type int

const (
	AGPL Type = iota
	Apache
	BSD
	CC
	GPL
	JSON
	MIT
	Unlicense
	Zlib
	Other
	NumTypes = Other
)

func licenseType(name string) Type {
	for l := Type(0); l < NumTypes; l++ {
		if strings.HasPrefix(name, l.String()) {
			return l
		}
	}
	return Other
}

// A License describes a single license that can be recognized.
// At least one of the Text or the URL should be set.
type License struct {
	Name string
	Text string
	URL  string
}

// Coverage describes how the text matches various licenses.
type Coverage struct {
	// Percent is the fraction of the total text, in normalized words, that
	// matches any valid license, expressed as a percentage across all of the
	// licenses matched.
	Percent float64

	// Match describes, in sequential order, the matches of the input text
	// across the various licenses. Typically it will be only one match long,
	// but if the input text is a concatenation of licenses it will contain
	// a match value for each element of the concatenation.
	Match []Match
}

// When we build the Match, Start and End are word offsets,
// but they are converted to byte offsets in the original
// before being passed back to the caller.

// Match describes how a section of the input matches a license.
type Match struct {
	Name    string  // The (file) name of the license it matches.
	Type    Type    // The type of the license: BSD, MIT, etc.
	Percent float64 // The fraction of words between Start and End that are matched.
	Start   int     // The byte offset of the first word in the input that matches.
	End     int     // The byte offset of the end of the last word in the input that matches.
	// IsURL reports that the matched text identifies a license by indirection
	// through a URL. If set, Start and End specify the location of the URL
	// itself, and Percent is always 100.0.
	IsURL bool
}
