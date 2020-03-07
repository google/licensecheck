// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package match

import (
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
