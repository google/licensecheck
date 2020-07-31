// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package match

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

var reParseTests = []struct {
	in  string
	err string
	out string
}{
	{in: "abc", out: "abc"},
	{in: "abc//**text**//def", out: "abc def"},
	{in: "a b c", out: "a b c"},
	{in: " (( abc )) ??", out: "((abc))??"},
	{in: "a b (( c ))??", out: "a b\n((c))??"},
	{in: "(a b ((c) ))??", out: "a b\n((c))??"},
	{in: "(( a b c )) ??", out: "((a b c))??"},
	{in: "z \n(( w ))\n(( a b c )) ??\n", out: "z w\n((a b c))??"},
	{in: "(( a __123__ c )) ??", out: "((a __123__ c))??"},
	{in: "a b ((c ||| d e)) f", out: "a b\n((c || d e))\nf"},
}

func TestReParse(t *testing.T) {
	var d Dict
	for _, tt := range reParseTests {
		re, err := reParse(&d, tt.in, strings.Contains(tt.in, "\n"))
		if err != nil {
			if tt.err == "" {
				t.Errorf("reParse(%q): %v", tt.in, err)
			} else if !strings.Contains(err.Error(), tt.err) {
				t.Errorf("reParse(%q): have error %q, want %q", tt.in, err.Error(), tt.err)
			}
			continue
		}
		if tt.err != "" {
			t.Errorf("reParse(%q): success but want error %q", tt.in, tt.err)
			continue
		}
		out := re.string(&d)
		if out != tt.out {
			t.Errorf("reParse(%q) = %q, want %q", tt.in, out, tt.out)
		}

		// Should be able to parse reprinted output in strict mode.
		_, err = reParse(&d, out, true)
		if err != nil {
			t.Errorf("reParse(%q): %v", out, err)
		}
	}
}

var reParseErrorTests = []struct {
	in  string
	err string
}{
	{"a ((b))", "(( not at beginning of line"},
	{"a || b", "|| outside (( ))"},
	{"((b)) c", ")) not at end of line"},
	{"a??", "?? not preceded by ))"},
	{"((a))\n??", "?? not preceded by ))"},
}

func TestReParseError(t *testing.T) {
	var d Dict
	for _, tt := range reParseErrorTests {
		_, err := reParse(&d, tt.in, true)
		if err == nil {
			t.Errorf("reParse(%q): unexpected success, want error: %s", tt.in, tt.err)
			continue
		}
		if !strings.Contains(err.Error(), tt.err) {
			t.Errorf("reParse(%q): %v, want error: %s", tt.in, err, tt.err)
		}
	}
}

var phraseRETests = []struct {
	in  string
	out string
}{
	{in: "abc", out: "[[abc]]"},
	{in: "a b c", out: "[[a b]]"},
	{in: "abc ??", out: "[[] [abc]]"},
	{in: "a b c ??", out: "[[a b]]"},
	{in: "(a b c) ??", out: "[[a b]]"},
	{in: "(( a b c )) ??", out: "[[] [a b]]"},
	{in: "(( a b c )) ?? d e f", out: "[[a b] [d e]]"},
	{in: "(( a __123__ c )) ??", out: "[[] [a ?] [a c]]"},
	{in: "a b ((c ||| d e)) f", out: "[[a b]]"},
	{in: "((a || b)) ((c || d))", out: "[[a c] [a d] [b c] [b d]]"},
	{in: "a?? b c", out: "[[a b] [b c]]"},
	{in: "((a __1__))?? b c", out: "[[a ?] [a b] [b c]]"},
	{in: "a __20__", out: "[[a ?] [a]]"},
}

func TestLeadingPhrases(t *testing.T) {
	var d Dict
	for _, tt := range phraseRETests {
		re, err := reParse(&d, tt.in, false)
		if err != nil {
			t.Errorf("reParse(%q): %v", tt.in, err)
			continue
		}
		phrases := re.leadingPhrases()
		sort.Slice(phrases, func(i, j int) bool {
			pi := phrases[i]
			pj := phrases[j]
			if pi[0] != pj[0] {
				return pi[0] < pj[0]
			}
			return pi[1] < pj[1]
		})
		var b strings.Builder
		words := d.Words()
		toText := func(w WordID) string {
			if w == AnyWord {
				return "?"
			}
			if w == BadWord {
				return "!"
			}
			return words[w]
		}
		fmt.Fprintf(&b, "[")
		for i, p := range phrases {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString("[")
			if p[1] != BadWord {
				b.WriteString(toText(p[0]))
				b.WriteString(" ")
				b.WriteString(toText(p[1]))
			} else if p[0] != BadWord {
				b.WriteString(toText(p[0]))
			}
			b.WriteString("]")
		}
		b.WriteString("]")
		out := b.String()
		if out != tt.out {
			t.Errorf("reParse(%q).leadingPhrases() = %v, want %v", tt.in, out, tt.out)
		}
	}
}
