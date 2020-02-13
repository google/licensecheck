// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package licensecheck classifies license files and heuristically determines
// how well they correspond to known open source licenses.
package licensecheck

import (
	"regexp"
	"sort"
	"strings"
)

// The order matters here so everything typechecks for the tools, which are fussy.
//go:generate rm -f data.gen.go
//go:generate stringer -type Type
//go:generate go run gen_data.go

// Options allow us to adjust parameters for the matching algorithm.
// TODO: Delete this once the package has been fine-tuned.
type Options struct {
	MinLength int // Minimum length of run, in words, to count as a matching substring.
	Threshold int // Percentage threshold to report a match.
	Slop      int // Maximum allowable gap in a near-contiguous match.
}

var defaults = Options{
	MinLength: 20,
	Threshold: 40,
	Slop:      8,
}

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

// A phrase is a sequence of words used as a key for startIndexes.
// Empirically, two words are best; more is slower.
type phrase [2]int32

type license struct {
	typ          Type
	name         string
	text         string
	doc          *document
	startIndexes map[phrase][]int
}

type document struct {
	text    []byte  // Original text.
	words   []int32 // Normalized words (indexes into c.words)
	byteOff []int32 // ith byteOff is byte offset of ith word in original text.
}

// A Checker matches a set of known licenses.
type Checker struct {
	licenses []license
	urls     map[string]string
	dict     map[string]int32 // dict maps word to index in words
	words    []string         // list of known words
}

// A License describes a single license that can be recognized.
// At least one of the Text or the URL should be set.
type License struct {
	Name string
	Text string
	URL  string
}

// New returns a new Checker that recognizes the given list of licenses.
func New(licenses []License) *Checker {
	c := &Checker{
		licenses: make([]license, 0, len(licenses)),
		urls:     make(map[string]string),
		dict:     make(map[string]int32),
	}
	for _, l := range licenses {
		if l.Text != "" {
			next := len(c.licenses)
			c.licenses = c.licenses[:next+1]
			cl := &c.licenses[next]
			cl.name = l.Name
			cl.typ = licenseType(cl.name)
			cl.text = l.Text
			cl.doc = c.normalize([]byte(cl.text))
			cl.startIndexes = startIndexes(cl.doc.words)
		}
		if l.URL != "" {
			c.urls[l.URL] = l.Name
		}
	}
	return c
}

// Initialized in func init in data.gen.go.
var builtin *Checker
var builtinList []License

// BuiltinLicenses returns the list of licenses built into the package.
// That is, the built-in checker is equivalent to New(BuiltinLicenses()).
func BuiltinLicenses() []License {
	// Return a copy so caller cannot change list entries.
	return append([]License{}, builtinList...)
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

type submatch struct {
	start      int // Index of starting word.
	end        int // Index of first following word.
	licenseEnd int // Index within license of last matching word.
	// Number of words between start and end that actually match.
	// Because of slop, this can be less than end-start.
	matched int
}

// startIndexes is used during initialization to construct a map from
// the occurrences of each word in the license to their byte offsets.
func startIndexes(words []int32) map[phrase][]int {
	m := make(map[phrase][]int, len(words))
	var p phrase
	const n = len(p)
	for i := 0; i+n <= len(words); i++ {
		copy(p[:], words[i:])
		m[p] = append(m[p], i)
	}
	return m
}

// Cover computes the coverage of the text according to the
// license set compiled into the package.
//
// An input text may match multiple licenses. If that happens,
// Match contains only disjoint matches. If multiple licenses
// match a particular section of the input, the best match
// is chosen so the returned coverage describes at most
// one match for each section of the input.
//
func Cover(input []byte, opts Options) (Coverage, bool) {
	return builtin.Cover(input, opts)
}

// Cover is like the top-level function Cover, but it uses the
// set of licenses in the Checker instead of the built-in license set.
func (c *Checker) Cover(input []byte, opts Options) (Coverage, bool) {
	doc := c.normalize(input)
	// Match the input text against all licenses.
	var matches []Match
	for _, l := range c.licenses {
		// For each license, there may be multiple submatches,
		// usually indicating multiple licenses in a file.
		// Create a separate Match for each.
		for _, s := range l.submatches(doc.words, opts) {
			matches = append(matches, makeMatch(l, s))
		}
	}

	if len(matches) == 0 {
		matches := doc.findURLsBetween(c, nil)
		if len(matches) == 0 {
			return Coverage{}, false
		}
		overallPercent := doc.percent(matches)
		doc.toByteOffsets(c, matches)
		return Coverage{
			Percent: overallPercent,
			Match:   matches,
		}, true
	}

	// Sort into lexical order so Coverage is sequential across the input.
	doc.sort(matches)

	// We have potentially multiple candidate matches and must winnow them
	// down to the best non-overlapping set. Do this by noticing when two
	// overlap, and killing off the one that matches fewer words in the
	// text, including the slop.
	killed := make([]bool, len(matches))
	for i := range matches {
		if killed[i] {
			continue
		}
		mi := &matches[i]
		miWords := mi.Percent * float64(mi.End-mi.Start)
		for j := range matches {
			if killed[j] || i == j {
				continue
			}
			mj := &matches[j]
			if mi.overlaps(mj) {
				victim := i
				if miWords > mj.Percent*float64(mj.End-mj.Start) {
					victim = j
				}
				killed[victim] = true
			}
		}
	}
	result := matches[:0]
	for i := range matches {
		if !killed[i] {
			result = append(result, matches[i])
		}
	}
	matches = result

	// Look for URLs in the gaps.
	if urls := doc.findURLsBetween(c, matches); len(urls) > 0 {
		// Sort again.
		matches = append(matches, urls...)
		doc.sort(matches)
	}

	// Compute this before overwriting offsets.
	overallPercent := doc.percent(matches)

	doc.toByteOffsets(c, matches)

	return Coverage{
		Percent: overallPercent,
		Match:   matches,
	}, true
}

func (doc *document) sort(matches []Match) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Start < matches[j].Start
	})
}

func (doc *document) wordOffset(byteOffset int) int {
	for i, off := range doc.byteOff {
		if int(off) >= byteOffset {
			return i
		}
	}
	return len(doc.words)
}

// toByteOffsets converts in-place the non-URL Matches' word offsets in the document to byte offsets.
func (doc *document) toByteOffsets(c *Checker, matches []Match) {
	for i := range matches {
		m := &matches[i]
		start := m.Start
		if start == 0 {
			m.Start = 0
		} else {
			m.Start = int(doc.byteOff[start])
		}
		end := m.End
		if end == 0 {
			m.End = 0
		} else if end == len(doc.words) {
			m.End = len(doc.text)
		} else {
			m.End = int(doc.byteOff[end-1]) + len(c.words[doc.words[end-1]])
		}
	}
}

// The regular expression is a simplified finder of URLS. We assume licenses are
// going to have fairly simple URLs, and in practice they do. See urls.go.
// Matching is case-insensitive.
const (
	pathRE   = `[-a-z0-9_.#?=]+` // Paths plus queries.
	domainRE = `[-a-z0-9_.]+`
)

var urlRE = regexp.MustCompile(`(?i)https?://(` + domainRE + `)+(\.org|com)(/` + pathRE + `)+/?`)

// findURLsBetween returns a slice of Matches holding URLs of licenses, to be
// inserted into the total list of Matches.
func (doc *document) findURLsBetween(c *Checker, matches []Match) []Match {
	var out []Match
	for i, startWord, nextStartWord := 0, 0, 0; startWord < len(doc.words); i, startWord = i+1, nextStartWord {
		endWord := len(doc.words)
		nextStartWord = endWord
		if i+1 < len(matches) {
			endWord = matches[i+1].Start
			nextStartWord = matches[i+1].End
		}
		// If there's not enough words here for a URL, like http://b.co, then don't try.
		if endWord < startWord+3 {
			continue
		}
		start := int(doc.byteOff[startWord])
		// Since doc.words excludes numerals, the last "word" might not actually
		// be the last text in the file. Make sure to run to EOF if we're at the end.
		// Otherwise, the end will go right up to the start of the next match, and
		// that will include all the text in the gap.
		end := len(doc.text)
		if endWord < len(doc.words) {
			end = int(doc.byteOff[endWord-1]) + len(c.words[doc.words[endWord-1]])
		}
		urlIndexes := urlRE.FindAllIndex(doc.text[start:end], -1)
		if len(urlIndexes) == 0 {
			continue
		}
		for _, u := range urlIndexes {
			u0, u1 := u[0]+start, u[1]+start
			if name, ok := c.licenseURL(string(doc.text[u0:u1])); ok {
				out = append(out, Match{
					Name:    name,
					Type:    licenseType(name),
					Percent: 100.0, // 100% of Start:End is a license URL.
					Start:   doc.wordOffset(u0),
					End:     doc.wordOffset(u1),
					IsURL:   true,
				})
			}
		}
	}
	return out
}

// licenseURL reports whether url is a known URL, and returns its name if it is.
func (c *Checker) licenseURL(url string) (string, bool) {
	// We need to canonicalize the text for lookup.
	// First, trim the leading http:// or https:// and the trailing /.
	// Then we lower-case it.
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/legalcode") // Common for CC licenses.
	url = strings.ToLower(url)
	name, ok := c.urls[url]
	return name, ok
}

// percent returns the total percentage of words in the input matched by matches.
// When it is called, matches (except for URLs) are in units of words.
func (doc *document) percent(matches []Match) float64 {
	matchLength := 0
	for i, m := range matches {
		if m.IsURL {
			matchLength += doc.endPos(matches, i) - doc.startPos(matches, i)
		} else {
			matchLength += m.End - m.Start
			continue
		}
	}
	return 100 * float64(matchLength) / float64(len(doc.words))
}

// startPos returns the starting position of match i for purposes of computing
// coverage percentage. For URLs, it's tricky because Start and End refer to the
// URL itself, so we presume the match covers the whole gap.
func (doc *document) startPos(matches []Match, i int) int {
	m := matches[i]
	if !m.IsURL {
		return m.Start
	}
	// This is a URL match.
	if i == 0 {
		return 0
	}
	// Is the previous match a URL? If so, split the gap.
	// If not, take the whole gap.
	prev := matches[i-1]
	if !prev.IsURL {
		return prev.End
	}
	return (m.Start + prev.End) / 2
}

// endPos is the complement of startPos.
func (doc *document) endPos(matches []Match, i int) int {
	m := matches[i]
	if !m.IsURL {
		return m.End
	}
	if i == len(matches)-1 {
		return len(doc.words)
	}
	next := matches[i+1]
	if !next.IsURL {
		return next.Start
	}
	return (m.End + next.Start) / 2
}

func makeMatch(l license, s submatch) Match {
	var match Match
	match.Name = licenseName(l.name)
	match.Type = l.typ
	match.Percent = 100 * float64(s.matched) / float64(len(l.doc.words))
	match.Start = s.start
	match.End = match.Start + (s.end - s.start)
	return match
}

// licenseName does any renaming required for licenses with multiple texts.
func licenseName(name string) string {
	switch name {
	case "Apache-2.0-User":
		// Apache-2.0 has two forms.
		return "Apache-2.0"
	}
	return name
}

// overlaps reports whether the two matches represent at least part of the same text.
func (m *Match) overlaps(n *Match) bool {
	return m.Start < n.End && n.Start < m.End
}

// submatches returns a list describing the runs of words in text
// that match the license. Its algorithm is a heuristic and can be
// defeated, but seems to work well in practice.
func (l *license) submatches(text []int32, opts Options) (s []submatch) {
	if len(text) == 0 || len(l.doc.words) == 0 {
		return s
	}
	if opts.MinLength <= 0 {
		opts.MinLength = defaults.MinLength
	}
	if opts.Slop <= 0 {
		opts.Slop = defaults.Slop
	}
	// For each word of the input, look to see if a sequence starting there
	// matches a sequence in the license.
	var p phrase
	for k := 0; k+len(p) <= len(text); { // k updated in loop.
		copy(p[:], text[k:])
		// Find longest match starting with that word.
		startIndexes := l.startIndexes[p]
		matchLength := 0
		matchIndex := 0
		for _, index := range startIndexes {
			start := k
			j := k + len(p)
			for _, w := range l.doc.words[index+len(p):] {
				if j == len(text) || w != text[j] {
					break
				}
				j++
			}
			if j-start > matchLength {
				matchLength = j - start
				matchIndex = index
			}
		}

		if matchLength < opts.MinLength {
			k++
			continue
		}

		// We have a long match. Remember it and advance the location in
		// the text. Note that we do not do anything to advance the license
		// text, which means that certain reorderings will match, perhaps
		// erroneously. This has not appeared in practice, while handling
		// things this way means the algorithm can identify multiple
		// appearances of a license within a single file.
		start := k
		end := start + matchLength
		k = end // The last word is not part of the match, but might be part of the next.

		// The blank (wildcard) ___ maps to word ID -1.
		// If we see a blank, we allow it to be filled in by up to 70 words.
		// This allows recognizing quite a few specialized copyright lines
		// (see for example testdata/MIT.t2) while not being large enough
		// to jump over an entire other license (our shortest is Apache-2.0-User
		// at 80 words).
		const blank = -1
		const blankMax = 70

		// Does this fit onto the previous match, or is it close
		// enough to consider? The slop allows text like
		//	Copyright (c) 2009 Snarfboodle Inc. All rights reserved.
		// to match
		// 	Copyright (c) <YEAR> <COMPANY>. All rights reserved.
		// and be considered a single span.
		if len(s) > 0 {
			prev := &s[len(s)-1]
			textGap := opts.Slop
			if prev.licenseEnd < len(l.doc.words) && l.doc.words[prev.licenseEnd] == blank {
				textGap = blankMax
			}
			if prev.end+textGap >= start && matchIndex >= prev.licenseEnd {
				if textGap == blankMax {
					prev.matched++ // matched the blank
				}
				prev.end = end
				prev.matched += matchLength
				prev.licenseEnd = matchIndex + matchLength
				continue
			}
		}

		// Does this match immediately follow an early blank in the license text?
		// If so, see if we can extend it backward.
		// The most common case needing this is licenses that start with "Copyright ___".
		// The text before the blank is too short to be its own match but it can be
		// part of this one.
		if matchIndex >= 2 && l.doc.words[matchIndex-1] == blank && l.doc.words[matchIndex-2] != blank {
			min := start - blankMax
			if min < 0 {
				min = 0
			}
			if len(s) > 0 && min < s[len(s)-1].end {
				min = s[len(s)-1].end
			}
			for i := start - 1; i >= min; i-- {
				if text[i] == l.doc.words[matchIndex-2] {
					// Found a match across the gap.
					start = i
					matchIndex -= 2
					matchLength += 2
					// Extend backward if possible.
					for start > 0 && matchIndex > 0 && text[start-1] == l.doc.words[matchIndex-1] {
						start--
						matchIndex--
						matchLength++
					}
					break
				}
			}
		}

		s = append(s, submatch{
			start:      start,
			end:        end,
			matched:    matchLength,
			licenseEnd: matchIndex + matchLength,
		})
	}
	return s
}
