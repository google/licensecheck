// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"testing"
)

type urlTest struct {
	names []string
	text  string
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
		if len(cov.Match) != len(test.names) {
			t.Log(cov)
			t.Errorf("%q got %d matches; expected %d", test.names, len(cov.Match), len(test.names))
			continue
		}
		for i, m := range cov.Match {
			if test.names[i] != m.Name {
				t.Errorf("%q: match %d is %q; expected %q", test.names, i, m.Name, test.names[i])
			}
		}
		if cov.Percent < 40 {
			t.Log(cov)
			t.Errorf("%q: got %.2f%% overall percentage: expected >= 40%%", test.names, cov.Percent)
		}
	}
}
