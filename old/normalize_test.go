// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package old

import (
	"strings"
	"testing"
)

type listMarkerTest struct {
	text   string
	after  rune
	isList bool
}

var listMarkerTests = []listMarkerTest{
	// No match.
	{"", ' ', false},
	{"a", ' ', false},
	// Match.
	{"b", ')', true},
	{"b", '.', true},
	{"b", '.', true},
	{"i", ')', true},
	{"ii", ')', true},
	{"iii", ')', true},
	{"iv", ':', true},
}

func TestListMarkers(t *testing.T) {
	for _, test := range listMarkerTests {
		// All strings must be listMarkerLength or longer to test anything;
		isList := isListMarker(test.text, test.after)
		if isList != test.isList {
			t.Errorf("%q: got %t; want %t", test.text+string(test.after), isList, test.isList)
		}
	}
}

const apacheStart = `
                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.
`

var apacheWords = strings.Fields(strings.Join([]string{
	"apache license",
	"version january",
	"http www apache org licenses",

	"terms and conditions for use reproduction and distribution",

	"definitions",

	"license shall mean the terms and conditions for use reproduction",
	"and distribution as defined by sections through of this document",

	"licensor shall mean the copyright owner or entity authorized by",
	"the copyright owner that is granting the license",

	"legal entity shall mean the union of the acting entity and all",
	"other entities that control are controlled by or are under common",
	"control with that entity for the purposes of this definition",
	"control means the power direct or indirect to cause the",
	"direction or management of such entity whether by contract or",
	"otherwise or ownership of fifty percent or more of the",
	"outstanding shares or beneficial ownership of such entity",
}, " "))

func TestNormalize(t *testing.T) {
	c := builtin
	doc := c.normalize([]byte(apacheStart), true)
	for i, w := range doc.words {
		if i >= len(apacheWords) {
			t.Fatalf("more words than expected starting at %d: %s", i, c.words[w])
		}
		if c.words[w] != apacheWords[i] {
			t.Fatalf("mismatch at word %d: got %q; want %q", i, c.words[w], apacheWords[i])
		}
	}
}

func TestToLower(t *testing.T) {
	input := "A \xa1\xb0covered work\xa1\xb1 means the Program.\nI am the Α and the Ω.\n"
	output := "a ??covered work?? means the program.\ni am the α and the ω.\n"
	got := toLower([]byte(input))
	if got != output {
		t.Fatalf("expected %q; got %q", output, got)
	}
}

// There was a bug handling bad UTF-8-encoded files caused by the conversion
// from bytes to lower-case string changing the length of the string, leading to
// a mismatch between byte offsets and the words, stored as strings, of the
// document. This snippet, reduced from an instance of a badly encoded GPL
// license, triggered the bug.
const badUTF8 = "\xa1\xb0This License\xa1\xb1 refers to version 3.\n" +
	"\xa1\xb0Copyright\xa1\xb1 also means masks.\n" +
	"\xa1\xb0The Program\xa1\xb1 \xa1\xb0you\xa1\xb1. \xa1\xb0Licensees\xa1\xb1 and \xa1\xb0recipients\xa1\xb1 may be.\n" +
	"To \xa1\xb0modify\xa1\xb1 a work is called a \xa1\xb0modified version\xa1\xb1 of the earlier work or a work \xa1\xb0based on\xa1\xb1 the earlier work.\n" +
	"A \xa1\xb0covered work\xa1\xb1 means the Program.\n" +
	"To \xa1\xb0propagate\xa1\xb1 work means to do anything with it that, without permission, would make you directly or secondarily liable for infringement under applicable copyright law, except executing it on a computer or modifying a private copy.  Propagation includes copying, distribution (with or without modification), making available to the public, and in some countries other activities as well. \n" +
	"To \xa1\xb0convey\xa1\xb1 a work \n" +
	"The \xa1\xb0source code\xa1\xb1 for  \xa1\xb0Object code\xa1\xb1 means work.\n" +
	"A \xa1\xb0Standard Interface\xa1\xb1 means \n" +
	"The \xa1\xb0Corresponding Source\xa1\xb1 for a work\n" +
	"Sign a \xa1\xb0copyright disclaimer\xa1\xb1 for the program, if necessary. For more information on this,GNU GPL, see <http://www.gnu.org/licenses/>.\n" +
	"If this is what you want to do, please read <http://www.gnu.org/philosophy/why-not-lgpl.html>.\n" +
	""

func TestBadUTF8(t *testing.T) {
	_, ok := Cover([]byte(badUTF8), Options{Threshold: 1})
	if !ok {
		t.Fatalf("failed to handle bad UTF-8")
	}
}
