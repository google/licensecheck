// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package old

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
	{[]string{"CC-BY-4.0", "UPL-1.0"},
		"This code is licensed by https://creativecommons.org/licenses/BY/4.0/ so have fun" +
			"Also http://opensource.org/licenses/upl is relevant"},
	{[]string{"CC-BY-4.0", "MIT", "UPL-1.0"},
		"This code is licensed by https://creativecommons.org/licenses/BY/4.0/ so have fun" +
			license_MIT +
			"Also http://opensource.org/licenses/upl is relevant"},
	// Special case: concatenated licenses.
	{[]string{"MIT", "MIT"}, license_MIT + license_MIT},
	// There was a bug with a number at EOF. See comments in document.findURLsBetween.
	{[]string{"CC-BY-NC-ND-2.0"}, "See https://creativecommons.org/licenses/by-nc-nd/2.0"},
}

func TestURLMatch(t *testing.T) {
	for _, test := range urlTests {
		cov, ok := Cover([]byte(test.text), Options{})
		if !ok {
			t.Errorf("%q from %.20q... didn't match", test.names, test.text)
			continue
		}
		if len(cov.Match) != len(test.names) {
			t.Log(cov)
			t.Errorf("%q got %d matches; expected %d", test.names, len(cov.Match), len(test.names))
			continue
		}
		for i, m := range cov.Match {
			if test.names[i] != m.Name {
				t.Errorf("%q: match %d is %q; expected %q", test.names, i, m.Name, test.names[i])
			}
			// Since in our test the licenses are literal text and the code assumes URL
			// blocks are fully matched, we should get 100% coverage for the individual licenses,
			// but we do get very close....
			// TODO: Find the gap.
			if m.Percent < 99.0 {
				t.Log(cov)
				t.Errorf("%q: got %.2f%% overall percentage; expected >= 99.00%%", test.names, m.Percent)
			}
		}
		// .. as well as the overall.
		if cov.Percent != 100.0 {
			t.Log(cov)
			t.Errorf("%q: got %.2f%% overall percentage; expected 100.00%%", test.names, cov.Percent)
		}
	}
}
