// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"bytes"
	"strings"
	"testing"
)

// TestSelfCoverage makes sure that the license texts match themselves.
// That is a mininum requirement.
func TestSelfCoverage(t *testing.T) {
	for _, l := range licenses {
		cov, ok := Cover(l.doc.text, Options{})
		if !ok {
			t.Errorf("no coverage for %s", l.name)
			continue
		}
		if cov.Percent != 100.0 {
			t.Errorf("%s got %.2f%% coverage", l.name, cov.Percent)
		}
	}
}

// TestApache_2_0_User tests that the "user" form of Apache 2.0 reports the license correctly.
func TestApache_2_0_User(t *testing.T) {
	text := findLicense("Apache-2.0-User").doc.text
	cov, ok := Cover(text, Options{})
	if !ok {
		t.Error("no coverage")
	}
	// Coverage should be 100%
	if cov.Percent != 100 {
		t.Errorf("coverage is %.2f%%, should be 100%%", cov.Percent)
	}
	if len(cov.Match) == 0 {
		t.Fatalf("expected a match; got none")
	}
	for _, m := range cov.Match {
		if m.Name != "Apache-2.0" || m.Percent != 100 {
			t.Errorf("expected license name %q at 100%%; got %q %.2f%%", "Apache-2.0", m.Name, m.Percent)
		}
	}
}

// TestMultiCoverage makes sure that concatenated license texts match themselves in sequence.
func TestMultiCoverage(t *testing.T) {
	mit := findLicense("MIT")
	apache := findLicense("Apache-2.0")
	bsd2 := findLicense("BSD-2-Clause")
	text := bytes.Join([][]byte{mit.doc.text, apache.doc.text, bsd2.doc.text},
		[]byte("\nHere is some intervening text\n"))
	cov, ok := Cover(text, Options{})
	if !ok {
		t.Error("no coverage")
	}
	// Coverage should be >=98% (it's actually about 99%) but not 100% - almost the entire file is a license.
	if cov.Percent == 100 {
		t.Errorf("coverage is %.2g%%, should be less than 100%%", cov.Percent)
	}
	if cov.Percent < 98.0 {
		t.Errorf("coverage is %.2g%%, should be closer to 100%%", cov.Percent)
	}
	// But all three should match by 100% individually.
	for _, match := range cov.Match {
		if match.Percent != 100.0 {
			t.Errorf("%s got only %.2g%% coverage", match.Name, cov.Percent)
		}
	}
	// Matches should be in the right order and not overlap.
	if len(cov.Match) != 3 {
		t.Fatalf("got %d matches; expect 3", len(cov.Match))
	}
	checkMatch(t, cov.Match[0], "MIT", -1)
	checkMatch(t, cov.Match[1], "Apache-2.0", cov.Match[0].End)
	checkMatch(t, cov.Match[2], "BSD-2-Clause", cov.Match[1].End)
}

func findLicense(name string) license {
	for _, l := range licenses {
		if l.name == name {
			return l
		}
	}
	panic("no license named " + name)
}

func checkMatch(t *testing.T, m Match, name string, prevEnd int) {
	t.Helper()
	if m.Name != name {
		t.Errorf("got %q; expected %q", m.Name, name)
		return
	}
	lic := findLicense(name)
	// Skip leading white space, almost certainly trimmed copyright lines.
	length := len(lic.doc.text)
	for _, c := range lic.doc.text {
		if c != ' ' && c != '\n' {
			break
		}
		length--
	}
	// There is some fudge factor in the match lengths because of terminal spaces, so be forgiving.
	min, max := length-5, length
	if n := m.End - m.Start; n < min || max < n {
		t.Errorf("match for %s is %d bytes long; expected %d", name, m.End-m.Start, length)
	}
	if m.End <= m.Start {
		t.Errorf("match for %s starts at %d after it ends at %d", name, m.Start, m.End)
	}
	if m.Start <= prevEnd {
		t.Errorf("match for %s starts at %d before previous ends at %d", name, m.Start, prevEnd)
	}
}

func TestWordOffset(t *testing.T) {
	mit := findLicense("MIT") // A reasonably short one.
	doc := mit.doc
	for i, byteOff := range doc.byteOff {
		wordOff := mit.doc.wordOffset(int(byteOff))
		if wordOff != i {
			t.Fatalf("%d: got word %d; expected %d", i, wordOff, i)
		}
	}
	wordOff := doc.wordOffset(len(doc.text))
	if wordOff != len(doc.words) {
		t.Fatalf("%d: got word %d; expected %d", len(doc.words), wordOff, len(doc.words))
	}
}

const testText = `Copyright (c) 2010 The Walk Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:
1. Redistributions of source code must retain the above copyright
   notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright
   notice, this list of conditions and the following disclaimer in the
   documentation and/or other materials provided with the distribution.
3. The names of the authors may not be used to endorse or promote products
   derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE AUTHORS "AS IS" AND ANY EXPRESS OR
IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
`

// Verify known offsets etc., protecting against
// https://github.com/google/licensecheck/issues/8
func TestMatch(t *testing.T) {
	cover, ok := Cover([]byte(testText), Options{})
	if !ok {
		t.Fatal("Cover failed")
	}
	match := cover.Match[0]
	if expect := "BSD-3-Clause"; match.Name != expect {
		t.Errorf("name is %q; should be %q", match.Name, expect)
	}
	if expect := BSD; match.Type != expect {
		t.Errorf("Type is %q; should be %q", match.Type, expect)
	}
	if match.IsURL {
		t.Errorf("match is URL")
	}
	if expect := strings.Index(testText, "Redistribution"); match.Start != expect {
		t.Errorf("start is %d; should be %d", match.Start, expect)
	}
	if expect := len(testText) - 2; match.End != expect { // -2 for newline and terminal period.
		t.Errorf("end is %d; should be %d", match.End, expect)
	}
}
