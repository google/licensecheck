// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package old

import (
	"bytes"
	"testing"
)

// TestSelfCoverage makes sure that the license texts match themselves.
// That is a mininum requirement.
func TestSelfCoverage(t *testing.T) {
	for _, l := range builtin.licenses {
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
	for _, l := range builtin.licenses {
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
	length := len(lic.doc.text)
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
