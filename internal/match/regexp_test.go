// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package match

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var multiMatchTests = []struct {
	re   string
	in   string
	list []Match
}{
	{"a\n((b || c))\nd", `a b d`, []Match{{0, 0, 3}}},
	{"a\n((b || c))\nd", `a b c d`, nil},
	{"a b c / a\n((c || d))\ne", `a b c x a c e x a d e x`, []Match{{0, 0, 3}, {1, 4, 7}, {1, 8, 11}}},
	{"a b c / a b c d / b c e", `a b c d e a b c b c e`, []Match{{1, 0, 4}, {0, 5, 8}, {2, 8, 11}}},
}

func TestMultiLREMatch(t *testing.T) {
	var d Dict
	for id, tt := range multiMatchTests {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			var list []*LRE
			for _, expr := range strings.Split(tt.re, "/") {
				re, err := ParseLRE(&d, "x", expr)
				if err != nil {
					t.Fatalf("Parse(%q): %v", expr, err)
				}
				list = append(list, re)
			}
			re, err := NewMultiLRE(list)
			if err != nil {
				t.Fatal(err)
			}
			m := re.Match(tt.in)
			var mlist []Match
			if m != nil {
				if m.Text != tt.in {
					t.Errorf("invalid text in m")
				}
				words := d.Split(tt.in)
				if !reflect.DeepEqual(m.Words, words) {
					t.Errorf("invalid words in m")
				}
				mlist = m.List
			}
			if !reflect.DeepEqual(mlist, tt.list) {
				t.Errorf("incorrect match:\nhave %+v\nwant %+v", mlist, tt.list)
			}
		})
	}
}
