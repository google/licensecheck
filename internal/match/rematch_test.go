// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package match

import (
	"strings"
	"testing"
)

var compileTests = `
a b c
0	word a
1	word b
2	word c
3	match 0

a b c ??
0	word a
1	word b
2	alt 4
3	word c
4	match 0

a ((b || c)) d
0	word a
1	alt 4
2	word b
3	jump 5
4	word c
5	word d
6	match 0

a __3__ b
0	word a
1	alt 7
2	any
3	alt 7
4	any
5	alt 7
6	any
7	word b
8	match 0

a b c / d e f
0	alt 5
1	word a
2	word b
3	word c
4	match 0
5	word d
6	word e
7	word f
8	match 1

((c __2__))?? d e f
0	alt 6
1	word c
2	alt 6
3	any
4	alt 6
5	any
6	word d
7	word e
8	word f
9	match 0
`

func TestCompile(t *testing.T) {
	var d Dict
	for _, tt := range strings.Split(compileTests, "\n\n") {
		tt = strings.TrimSpace(tt) + "\n"
		i := strings.Index(tt, "\n")
		in, want := tt[:i], tt[i+1:]

		prog := testProg(t, &d, in)
		if prog == nil {
			continue
		}
		out := prog.string(&d)
		if out != want {
			t.Errorf("RE(%q).prog():\nhave:\n%s\nwant:\n%s", in, out, want)
		}
	}
}

func testProg(t *testing.T, dict *Dict, expr string) reProg {
	if strings.Contains(expr, "/") {
		var list []*reSyntax
		for _, str := range strings.Split(expr, "/") {
			re, err := reParse(dict, str, false)
			if err != nil {
				t.Errorf("Parse(%q): %v", expr, err)
				return nil
			}
			list = append(list, re)
		}
		return reCompileMulti(list)
	}

	re, err := reParse(dict, expr, false)
	if err != nil {
		t.Errorf("reParse(%q): %v", expr, err)
		return nil
	}
	return re.compile(nil, 0)
}

var compileDFATests = `
a b c
0 a:3
3 b:6
6 c:9
9 m0

a ((b || c)) d
0 a:3
3 b:8 c:8
8 d:11
11 m0

a b c / a ((c | d)) e
0 a:3
3 b:8 c:13
8 c:11
11 m0
13 d:16
16 e:19
19 m1

((c __2__))?? d e f
0 c:5 d:18
5 *:10 d:31
10 *:15 d:26
15 d:18
18 e:21
21 f:24
24 m0
26 d:18 e:21
31 *:15 d:26 e:38
38 d:18 f:24


a b c / a ((c || d)) e
0 a:3
3 b:10 c:15 d:15
10 c:13
13 m0
15 e:18
18 m1
`

func TestCompileDFA(t *testing.T) {
	var d Dict
	for _, tt := range strings.Split(compileDFATests, "\n\n") {
		tt = strings.TrimSpace(tt) + "\n"
		i := strings.Index(tt, "\n")
		in, want := tt[:i], tt[i+1:]

		prog := testProg(t, &d, in)
		if prog == nil {
			continue
		}
		dfa := reCompileDFA(prog)
		out := dfa.string(&d)
		if out != want {
			t.Errorf("RE(%q).dprog():\nhave:\n%s\nwant:\n%s", in, out, want)
		}
	}
}

var matchTests = []struct {
	re    string
	in    string
	match int32
	end   int
}{
	{`a b c`, `a b c d`, 0, 3},
	{`a ((b || c)) d`, `a b d`, 0, 3},
	{`a ((b || c)) d`, `a b c d`, -1, 0},
	{`a __1__ d`, `a b c d e f g h i j k`, -1, 0},
	{`a __1__ c`, `a b c d e f g h i j k`, 0, 3},
	{`a __1__ b`, `a b c d e f g h i j k`, 0, 2},
	{`a __1__ b`, `a b b d e f g h i j k`, 0, 3},
	{`a __1__ a`, `a b c d e f g h i j k`, -1, 0},
	{`a __5__ k`, `a b c d e f g h i j k`, -1, 0},
	{`a __5__ h`, `a b c d e f g h i j k`, -1, 0},
	{`a __5__ g`, `a b c d e f g h i j k`, 0, 7},
	{`a __5__ f`, `a b c d e f g h i j k`, 0, 6},
	{`a __5__ e`, `a b c d e f g h i j k`, 0, 5},
	{`a __5__ d`, `a b c d e f g h i j k`, 0, 4},
	{`a __5__ c`, `a b c d e f g h i j k`, 0, 3},
	{`a __5__ b`, `a b c d e f g h i j k`, 0, 2},
	{`a __5__ a`, `a b c d e f g h i j k`, -1, 0},
	{`a b c / d e f`, `a b c d`, 0, 3},
	{`a b c / d e f`, `d e f`, 1, 3},
	{`a b c / d e f`, `a b d e f`, -1, 0},
	{`a b c / __3__ d e f`, `a b d e f g`, 1, 5},

	// Spell checking
	// allowedMismatches + singular/plural original(s) via canMisspell
	{`it is good this copy of it matches the original`, `they are good the copies of them matches those originals`, 0, 10},

	// canMisspell
	{`abcdef`, `abdef`, 0, 1},
	{`abcdef`, `bcdef`, 0, 1},
	{`abcdef`, `abcde`, 0, 1},
	{`abcdef`, `abcxdef`, 0, 1},
	{`abcdef`, `xabcdef`, 0, 1},
	{`abcdef`, `abcdefx`, 0, 1},
	{`abcdef`, `abxdef`, 0, 1},
	{`abcdef`, `xbcdef`, 0, 1},
	{`abcdef`, `abcdex`, 0, 1},

	// canMisspellJoin
	{`x noninfringement y`, `x non-infringement y`, 0, 4},

	// misspell split
	{`i non-infringement j`, `i noninfringement j`, 0, 3},
}

func TestReDFAMatch(t *testing.T) {
	var d Dict
	for _, tt := range matchTests {
		prog := testProg(t, &d, tt.re)
		if prog == nil {
			continue
		}
		dfa := reCompileDFA(prog)
		match, end := dfa.match(&d, tt.in, d.Split(tt.in))
		if match != tt.match || end != tt.end {
			t.Errorf("reDFA(%q).match(%v) = %v, %v, want %v, %v", tt.re, tt.in, match, end, tt.match, tt.end)
		}
	}
}
