// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package match

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

func TestDict(t *testing.T) {
	words := []string{"a", "b", "c", "b", "d", "a"}
	nonwords := []string{"abcdef"}
	indexes := []WordID{0, 1, 2, 1, 3, 0}

	var d Dict
	for j := 0; j < 2; j++ {
		for i, w := range words {
			id := d.Insert(w)
			if id != indexes[i] {
				t.Errorf("#%d: Insert(%q) = %d, want %d", i, w, id, indexes[i])
			}
		}
	}

	for i, w := range words {
		id := d.Lookup(w)
		if id != indexes[i] {
			t.Errorf("#%d: Lookup(%q) = %d, want %d", i, w, id, indexes[i])
		}
	}

	for i, w := range nonwords {
		id := d.Lookup(w)
		if id != BadWord {
			t.Errorf("#%d: Lookup(%q) = %d, want BadWord", i, w, id)
		}
	}

	list := d.Words()
	want := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(list, want) {
		t.Errorf("List() = %v, want %v", list, want)
	}
}

var htmlTagSizeTests = []struct {
	in  string
	out int
}{
	{"<x>y", 3},
	{"</x>y", 4},
	{"<https://www.google.com>", 0},
	{"<me@example.com>", 0},
	{"<a\nhref=golang.org>", 19},
	{"<a\n\n\ntoo many lines>", 0},
}

func TestHTMLTagSize(t *testing.T) {
	for _, tt := range htmlTagSizeTests {
		out := htmlTagSize(tt.in)
		if out != tt.out {
			t.Errorf("htmlTagSize(%q) = %d want %d", tt.in, out, tt.out)
		}
	}
}

var htmlEntitySizeTests = []struct {
	in  string
	out int
}{
	{"&x y; z", 0},
	{"&x; y z", 3},
	{"&#xabcxyz;", 0},
	{"&#xabc;xyz", 7},
	{"&#123;abc", 6},
	{"&#123abc;", 0},
}

func TestHTMLEntitySize(t *testing.T) {
	for _, tt := range htmlEntitySizeTests {
		out := htmlEntitySize(tt.in)
		if out != tt.out {
			t.Errorf("htmlEntitySize(%q) = %d want %d", tt.in, out, tt.out)
		}
	}
}

var markdownAnchorSizeTests = []struct {
	in  string
	out int
}{
	{"{#abc}", 6},
	{"{abc}", 0},
	{"{#abc", 0},
	{"{#abc def}", 0},
	{"{#abc\ndef}", 0},
	{"{#abc\rdef}", 0},
}

func TestMarkdownAnchorSize(t *testing.T) {
	for _, tt := range markdownAnchorSizeTests {
		out := markdownAnchorSize(tt.in)
		if out != tt.out {
			t.Errorf("markdownAnchorSize(%q) = %d want %d", tt.in, out, tt.out)
		}
	}
}

var insertSplitTests = []struct {
	in  string
	out string
}{
	{"ABC abc AbC", "abc abc abc"},
	{"ÀÁàáÈÉèéÌÍìíÒÓòóÙÚùú", "aaaaeeeeiiiioooouuuu"},
	{"&#34;abc&#34; 12 abc 13", "abc 12 abc 13"},
	{"&#x34;abc&#x34; 12 abc &amp; 13", "abc 12 abc 13"},
	{"&#34 &#x34 &amp abc", "34 x34 amp abc"},
	{"abc<", "abc"},
	{"abc<def", "abc def"},
	{"abc<1>def", "abc 1 def"},
	{"abc<p<>def", "abc p def"},
	{"abc<p\r>def", "abc def"},
	{"abc<p\n>def", "abc def"},
	{"abc<p\n\n>def", "abc def"},
	{"abc<p\n\n\n>def", "abc p def"},
	{"<p>abc</p>def", "abc def"},
	{"<p\n>abc", "abc"},
	{"<http://golang.org>", "http golang org"},
	{"heading {#head} more", "heading more"},
	{"[text](", "text"},
	{"[text](http", "text http"},
	{"[text](http://link", "text http link"},
	{"[text](http://link)MORE", "text more"},
	{"[text](https://link)MORE", "text more"},
	{"[text](https://link )MORE", "text http link more"},
	{"[text](https://link\t)MORE", "text http link more"},
	{"[text](https://link\r)MORE", "text http link more"},
	{"[text](https://link\n)MORE", "text http link more"},
	{"[text](#anchor) more", "text more"},
	{"Copyright 2020 Gopher®", "copyright 2020 gopher"},
	{"Copyright © 2020 Gopher®", "copyright 2020 gopher"},
	{"(c) 2020 Gopher®", "copyright 2020 gopher"},
	{"(C) 2020 Gopher®", "copyright 2020 gopher"},
	{"© 2020 Gopher®", "copyright 2020 gopher"},
	{"&copy; 2020 Gopher®", "copyright 2020 gopher"},
	{"a b c (c) d", "a b c copyright d"},
	{"a b c copyright d", "a b c copyright d"},
	{"a b c © d", "a b c copyright d"},

	{"http://golang.org", "http golang org"},
	{"https://golang.org", "http golang org"},
	{"the notice(s) must", "the notices must"},
}

func TestDictInsertSplit(t *testing.T) {
	var d Dict
	for _, tt := range insertSplitTests {
		words := d.InsertSplit(tt.in)
		var out string
		for i, w := range words {
			if i > 0 {
				out += " "
			}
			if w.ID == BadWord {
				out += "?"
			} else {
				out += d.Words()[w.ID]
			}
		}
		if out != tt.out {
			t.Errorf("Words(d, %q) = %q, want %q", tt.in, out, tt.out)
		}
	}
}

var splitTests = []struct {
	in  string
	out string
}{
	{"distribute,\n<p>sublicense, extort,  and/or\rsell ", "distribute sublicense ? and or sell"},
}

func TestDictSplit(t *testing.T) {
	var d Dict
	for _, w := range regexp.MustCompile(`\pL+`).FindAllString(rot13(mitLicenseRot13), -1) {
		d.Insert(w)
	}

	for _, tt := range splitTests {
		words := d.Split(tt.in)
		var out string
		for i, w := range words {
			if i > 0 {
				out += " "
			}
			if w.ID == BadWord {
				out += "?"
			} else {
				out += d.Words()[w.ID]
			}
		}
		if out != tt.out {
			t.Errorf("Words(d, %q) = %q, want %q", tt.in, out, tt.out)
		}
	}
}

var mitLicenseRot13 = ` // MIT License, rot13 to hide from license scanners
pbclevtug 2020 gur evtug tbcure

crezvffvba vf urerol tenagrq, serr bs punetr, gb nal crefba bognvavat n pbcl
bs guvf fbsgjner naq nffbpvngrq qbphzragngvba svyrf (gur "fbsgjner"), gb qrny
va gur fbsgjner jvgubhg erfgevpgvba, vapyhqvat jvgubhg yvzvgngvba gur evtugf
gb hfr, pbcl, zbqvsl, zretr, choyvfu, qvfgevohgr, fhoyvprafr, naq/be fryy
pbcvrf bs gur fbsgjner, naq gb crezvg crefbaf gb jubz gur fbsgjner vf
sheavfurq gb qb fb, fhowrpg gb gur sbyybjvat pbaqvgvbaf:

gur nobir pbclevtug abgvpr naq guvf crezvffvba abgvpr funyy or vapyhqrq va nyy
pbcvrf be fhofgnagvny cbegvbaf bs gur fbsgjner.

gur fbsgjner vf cebivqrq "nf vf", jvgubhg jneenagl bs nal xvaq, rkcerff be
vzcyvrq, vapyhqvat ohg abg yvzvgrq gb gur jneenagvrf bs zrepunagnovyvgl,
svgarff sbe n cnegvphyne checbfr naq abavasevatrzrag. va ab rirag funyy gur
nhgubef be pbclevtug ubyqref or yvnoyr sbe nal pynvz, qnzntrf be bgure
yvnovyvgl, jurgure va na npgvba bs pbagenpg, gbeg be bgurejvfr, nevfvat sebz,
bhg bs be va pbaarpgvba jvgu gur fbsgjner be gur hfr be bgure qrnyvatf va gur
fbsgjner.
`

func rot13(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'a' <= c && c <= 'm' {
			b[i] = c + 13
		}
		if 'n' <= c && c <= 'z' {
			b[i] = c - 13
		}
	}
	return string(b)
}

var bench struct {
	data []byte
	str  string
	dict Dict
}

func benchSetup(b *testing.B) {
	if bench.data != nil {
		return
	}
	files, err := filepath.Glob("../testdata/*")
	if err != nil {
		b.Fatal(err)
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			b.Fatal(err)
		}
		bench.data = append(bench.data, data...)
	}
	bench.str = string(bench.data)
	bench.dict.InsertSplit(bench.str)
	b.ResetTimer()
}

func BenchmarkInsertSplit(b *testing.B) {
	benchSetup(b)
	b.SetBytes(int64(len(bench.str)))
	for i := 0; i < b.N; i++ {
		var dict Dict
		dict.InsertSplit(bench.str)
	}
}

func BenchmarkSplit(b *testing.B) {
	benchSetup(b)
	b.SetBytes(int64(len(bench.str)))
	for i := 0; i < b.N; i++ {
		bench.dict.Split(bench.str)
	}
}
