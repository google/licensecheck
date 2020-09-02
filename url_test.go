// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"testing"
)

type urlTest struct {
	ids  []string
	text string
}

var urlTests = []urlTest{
	{[]string{"CC-BY-4.0"},
		"This code is licensed by https://creativecommons.org/licenses/BY/4.0/ so have fun"},
	{[]string{"CC-BY-NC-4.0"},
		"This code is licensed under https://creativecommons.org/licenses/by-nc/4.0/legalcode so have fun"},
	{[]string{"CC-BY-SA-4.0", "UPL-1.0"},
		"This code is licensed by https://creativecommons.org/licenses/BY-SA/4.0/ so have fun" +
			"Also http://opensource.org/licenses/upl is relevant"},
	{[]string{"CC-BY-ND-4.0", "MIT", "UPL-1.0"},
		"This code is licensed by https://creativecommons.org/licenses/BY-nd/4.0/ so have fun" +
			license_MIT +
			"Also http://opensource.org/licenses/upl is relevant"},
	// Special case: concatenated licenses.
	{[]string{"MIT", "MIT"}, license_MIT + license_MIT},
	// There was a bug with a number at EOF. See comments in document.findURLsBetween.
	{[]string{"CC-BY-NC-ND-2.0"}, "See https://creativecommons.org/licenses/by-nc-nd/2.0"},
}

func TestURLMatch(t *testing.T) {
	for _, test := range urlTests {
		cov := Scan([]byte(test.text))
		if len(cov.Match) != len(test.ids) {
			t.Log(cov)
			t.Errorf("%q got %d matches; expected %d", test.ids, len(cov.Match), len(test.ids))
			continue
		}
		for i, m := range cov.Match {
			if test.ids[i] != m.ID {
				t.Errorf("%q: match %d is %q; expected %q", test.ids, i, m.ID, test.ids[i])
			}
		}
		if cov.Percent < 40 {
			t.Log(cov)
			t.Errorf("%q: got %.2f%% overall percentage: expected >= 40%%", test.ids, cov.Percent)
		}
	}
}

func TestURLIDs(t *testing.T) {
	have := make(map[string]bool)
	for _, l := range builtinLREs {
		have[l.ID] = true
	}
	for _, l := range builtinURLs {
		if !have[l.ID] {
			t.Errorf("unknown URL license ID: %s", l.ID)
		}
	}
}

var license_MIT = rot13(mitLicenseRot13)

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
