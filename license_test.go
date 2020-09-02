// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/licensecheck/internal/match"
)

func init() {
	flag.IntVar(&match.TraceDFA, "tracedfa", match.TraceDFA, "trace DFA execution that bails out after `n` non-matching steps")
}

func TestTestdata(t *testing.T) {
	files, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no testdata files found")
	}

	for _, file := range files {
		name := filepath.Base(file)
		if name == "README" {
			continue
		}
		if info, err := os.Stat(file); err == nil && info.IsDir() {
			continue
		}
		if !strings.Contains(file, ".t") {
			t.Errorf("unexpected file: %v", file)
		}
		file := file
		t.Run(name, func(t *testing.T) {
			t.Parallel() // faster and tests for races in parallel usage

			data, err := ioutil.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			// See testdata/README for definition of test data file.
			// Header ends at blank line.
			i := bytes.Index(data, []byte("\n\n"))
			if i < 0 {
				t.Fatalf("%s: invalid test data file: no blank line terminating header", file)
			}
			hdr, data := strings.Split(string(data[:i]), "\n"), data[i+2:]

			lineno := 1
			// Skip leading comment lines.
			for len(hdr) > 0 && strings.HasPrefix(hdr[0], "#") {
				hdr = hdr[1:]
				lineno++
			}

			if len(hdr) > 0 && strings.HasPrefix(hdr[0], "set ") {
				t.Fatalf("%s:%d: set not implemented", file, lineno)
				hdr = hdr[1:]
				lineno++
			}
			if len(hdr) < 1 {
				t.Fatalf("%s: header too short", file)
			}

			linenoStart := lineno
			parseCoverage := func() Coverage {
				var want Coverage
				want.Percent, err = parsePercent(hdr[0])
				if err != nil {
					t.Fatalf("%s:%d: parsing want.Percent: %v", file, lineno, err)
				}
				hdr = hdr[1:]
				lineno++
				for i, line := range hdr {
					f := strings.Fields(line)
					if len(f) != 2 && len(f) != 3 {
						t.Fatalf("%s:%d: bad match field count", file, lineno)
					}
					var m Match
					m.ID = f[0]
					m.Start, m.End, err = parseRange(f[1], len(data))
					if err != nil {
						t.Fatalf("%s:%d: parsing want.Match[%d].Start,End: %v", file, lineno, i, err)
					}
					if len(f) == 3 {
						if f[2] != "URL" {
							t.Fatalf("%s:%d: field 2 should be omitted or should be 'URL'", file, lineno)
						}
						m.IsURL = true
					}
					want.Match = append(want.Match, m)
					lineno++
				}
				return want
			}

			want := parseCoverage()
			linenoEnd := lineno

			cov := Scan(data)
			for _, m := range cov.Match {
				typ := licenseType(m.ID)
				if m.Type != typ {
					t.Errorf("%s: match %s has Type=%s, want %s", file, m.ID, m.Type, typ)
				}
			}

			mismatch := false
			var buf bytes.Buffer
			if !matchPercent(cov.Percent, want.Percent) {
				fmt.Fprintf(&buf, "- %.1f%%\n+ %.1f%%\n", want.Percent, cov.Percent)
				mismatch = true
			} else {
				fmt.Fprintf(&buf, "  %.1f%%\n", cov.Percent)
			}

			covm, wantm := cov.Match, want.Match
			for len(covm) > 0 || len(wantm) > 0 {
				switch {
				case len(covm) > 0 && (len(wantm) == 0 || covm[0].End < wantm[0].Start):
					fmt.Fprintf(&buf, "+ %v\n", fmtMatch(covm[0], len(data)))
					covm = covm[1:]
					mismatch = true

				case len(covm) > 0 && len(wantm) > 0 && matchMatch(covm[0], wantm[0]):
					fmt.Fprintf(&buf, "  %v\n", fmtMatch(covm[0], len(data)))
					covm = covm[1:]
					wantm = wantm[1:]

				default:
					fmt.Fprintf(&buf, "- %v\n", fmtMatch(wantm[0], len(data)))
					wantm = wantm[1:]
					mismatch = true
				}
			}
			if mismatch {
				t.Errorf("%s:%d,%d: diff -want +have:\n%s", file, linenoStart, linenoEnd, buf.Bytes())
			}
		})
	}
}

// fmtMatch formats the match m for printing.
func fmtMatch(m Match, end int) string {
	// Special case for EOF end position.
	var hi string
	if m.End == end {
		hi = "$"
	} else {
		hi = fmt.Sprintf("%d", m.End)
	}
	s := fmt.Sprintf("%s %d,%s", m.ID, m.Start, hi)
	if m.IsURL {
		s += " URL"
	}
	return s
}

// parsePercent parses a percentage (float ending in %).
func parsePercent(s string) (float64, error) {
	if !strings.HasSuffix(s, "%") {
		return 0, fmt.Errorf("missing %% suffix")
	}
	return strconv.ParseFloat(s[:len(s)-len("%")], 64)
}

// parseRange parses a start,end range (two decimals separated by a comma).
// As a special case, the second decimal can be $ meaning end-of-file.
func parseRange(s string, end int) (int, int, error) {
	i := strings.Index(s, ",")
	if i < 0 {
		return 0, 0, fmt.Errorf("malformed range")
	}
	lo, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, 0, err
	}
	var hi int
	if s[i+1:] == "$" {
		hi = end
	} else {
		hi, err = strconv.Atoi(s[i+1:])
		if err != nil {
			return 0, 0, err
		}
	}
	return lo, hi, nil
}

// matchPercent reports whether have matches want.
// We require that they match to within 0.1.
func matchPercent(have, want float64) bool {
	return math.Abs(have-want) < 0.1
}

// matchMatch reports whether have matches want.
func matchMatch(have, want Match) bool {
	return have.ID == want.ID &&
		have.Start == want.Start &&
		have.End == want.End &&
		have.IsURL == want.IsURL
}

var benchdata []byte

func BenchmarkScanTestdata(b *testing.B) {
	if benchdata == nil {
		files, err := filepath.Glob("testdata/*")
		if err != nil {
			b.Fatal(err)
		}
		if len(files) == 0 {
			b.Fatalf("no testdata files found")
		}
		for _, file := range files {
			if info, err := os.Stat(file); err == nil && info.IsDir() {
				continue
			}
			data, err := ioutil.ReadFile(file)
			if err != nil {
				b.Fatal(err)
			}
			benchdata = append(benchdata, data...)
		}
	}

	b.SetBytes(int64(len(benchdata)))
	for i := 0; i < b.N; i++ {
		Scan(benchdata)
	}
}

var trace = flag.String("tr", "", "trace DFA execution on `file` in TestTrace")

func TestTrace(t *testing.T) {
	if *trace == "" {
		t.Skip("-tr not given")
	}
	data, err := ioutil.ReadFile(*trace)
	if err != nil {
		t.Fatal(err)
	}
	match.TraceDFA = 10
	cov := Scan(data)
	match.TraceDFA = 0

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%.1f%%\n", cov.Percent)
	for _, m := range cov.Match {
		fmt.Fprintf(&buf, "%v\n", fmtMatch(m, len(data)))
	}
	t.Logf("coverage:\n%v", buf.String())
}
