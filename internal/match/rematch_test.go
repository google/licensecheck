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

a __3__ b c d
0	word a
1	alt 7
2	any
3	alt 7
4	any
5	alt 7
6	any
7	word b
8	word c
9	word d
10	match 0

a __4__ b c d
0	word a
1	alt 9
2	any
3	alt 9
4	any
5	alt 9
6	any
7	alt 9
8	any
9	word b
10	word c
11	word d
12	cut [1, 9]
13	match 0

a __4__ b c d e
0	word a
1	alt 9
2	any
3	alt 9
4	any
5	alt 9
6	any
7	alt 9
8	any
9	word b
10	word c
11	word d
12	cut [1, 9]
13	word e
14	match 0

a __5__ b c d e f g
0	word a
1	alt 11
2	any
3	alt 11
4	any
5	alt 11
6	any
7	alt 11
8	any
9	alt 11
10	any
11	word b
12	word c
13	word d
14	cut [1, 11]
15	word e
16	word f
17	word g
18	match 0

a __5__ ((b c))?? d e f g
0	word a
1	alt 11
2	any
3	alt 11
4	any
5	alt 11
6	any
7	alt 11
8	any
9	alt 11
10	any
11	alt 14
12	word b
13	word c
14	word d
15	word e
16	word f
17	cut [1, 11]
18	word g
19	match 0

a __5__ ((b c d))?? e f g h
0	word a
1	alt 11
2	any
3	alt 11
4	any
5	alt 11
6	any
7	alt 11
8	any
9	alt 11
10	any
11	alt 16
12	word b
13	word c
14	word d
15	cut [1, 11]
16	word e
17	word f
18	word g
19	cut [1, 11]
20	word h
21	match 0

a __5__ ((b || c x)) d e f g
0	word a
1	alt 11
2	any
3	alt 11
4	any
5	alt 11
6	any
7	alt 11
8	any
9	alt 11
10	any
11	alt 14
12	word b
13	jump 16
14	word c
15	word x
16	word d
17	word e
18	cut [1, 11]
19	word f
20	word g
21	match 0

a b ((__5__ c d))??
0	word a
1	word b
2	alt 16
3	alt 13
4	any
5	alt 13
6	any
7	alt 13
8	any
9	alt 13
10	any
11	alt 13
12	any
13	word c
14	word d
15	cut [3, 13]
16	match 0

a b ((__5__ c d))??
0	word a
1	word b
2	alt 16
3	alt 13
4	any
5	alt 13
6	any
7	alt 13
8	any
9	alt 13
10	any
11	alt 13
12	any
13	word c
14	word d
15	cut [3, 13]
16	match 0

a b __5__ c d ((x?? y?? z z z || y w w w))
0	word a
1	word b
2	alt 12
3	any
4	alt 12
5	any
6	alt 12
7	any
8	alt 12
9	any
10	alt 12
11	any
12	word c
13	word d
14	alt 26
15	alt 18
16	word x
17	cut [2, 12]
18	alt 21
19	word y
20	cut [2, 12]
21	word z
22	cut [2, 12]
23	word z
24	word z
25	jump 31
26	word y
27	cut [2, 12]
28	word w
29	word w
30	word w
31	match 0

a b __5__ c __5__ d
0	word a
1	word b
2	alt 12
3	any
4	alt 12
5	any
6	alt 12
7	any
8	alt 12
9	any
10	alt 12
11	any
12	word c
13	cut [2, 12]
14	alt 24
15	any
16	alt 24
17	any
18	alt 24
19	any
20	alt 24
21	any
22	alt 24
23	any
24	word d
25	cut [14, 24]
26	match 0

The name __10__ may not be used.
0	word the
1	word name
2	alt 22
3	any
4	alt 22
5	any
6	alt 22
7	any
8	alt 22
9	any
10	alt 22
11	any
12	alt 22
13	any
14	alt 22
15	any
16	alt 22
17	any
18	alt 22
19	any
20	alt 22
21	any
22	word may
23	word not
24	word be
25	cut [2, 22]
26	word used
27	match 0
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
		var list []reProg
		for _, str := range strings.Split(expr, "/") {
			re, err := reParse(dict, str, false)
			if err != nil {
				t.Errorf("Parse(%q): %v", expr, err)
				return nil
			}
			prog, err := re.compile(nil, 0)
			if err != nil {
				t.Errorf("compile(%q): %v", expr, err)
				return nil
			}
			list = append(list, prog)
		}
		return reCompileMulti(list)
	}

	re, err := reParse(dict, expr, false)
	if err != nil {
		t.Errorf("reParse(%q): %v", expr, err)
		return nil
	}
	prog, err := re.compile(nil, 0)
	if err != nil {
		t.Errorf("compile(%q): %v", expr, err)
		return nil
	}
	return prog
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

a __1__ b c d e
0 a:3
3 *:8 b:22
8 b:11
11 c:14
14 d:17
17 e:20
20 m0
22 b:11 c:14

a __2__ b c d e
0 a:3
3 *:8 b:32
8 *:13 b:27
13 b:16
16 c:19
19 d:22
22 e:25
25 m0
27 b:16 c:19
32 *:13 b:27 c:39
39 b:16 d:22

a __3__ b c d e
0 a:3
3 *:8 b:49
8 *:13 b:37
13 *:18 b:32
18 b:21
21 c:24
24 d:27
27 e:30
30 m0
32 b:21 c:24
37 *:18 b:32 c:44
44 b:21 d:27
49 *:13 b:37 c:56
56 *:18 b:32 d:63
63 b:21 e:30

a __4__ b c d e
0 a:3
3 *:8 b:68
8 *:13 b:54
13 *:18 b:42
18 *:23 b:37
23 b:26
26 c:29
29 d:32
32 e:35
35 m0
37 b:26 c:29
42 *:23 b:37 c:49
49 b:26 d:32
54 *:18 b:42 c:61
61 *:23 b:37 d:32
68 *:13 b:54 c:75
75 *:18 b:42 d:32

a __5__ b c d e f g
0 a:3
3 *:8 b:93
8 *:13 b:79
13 *:18 b:65
18 *:23 b:53
23 *:28 b:48
28 b:31
31 c:34
34 d:37
37 e:40
40 f:43
43 g:46
46 m0
48 b:31 c:34
53 *:28 b:48 c:60
60 b:31 d:37
65 *:23 b:53 c:72
72 *:28 b:48 d:37
79 *:18 b:65 c:86
86 *:23 b:53 d:37
93 *:13 b:79 c:100
100 *:18 b:65 d:37

a __5__ b __5__ c
0 a:3
3 *:8 b:31
8 *:13 b:31
13 *:18 b:31
18 *:23 b:31
23 *:28 b:31
28 b:31
31 *:36 c:59
36 *:41 c:59
41 *:46 c:59
46 *:51 c:59
51 *:56 c:59
56 c:59
59 m0

The name __10__ may not be used
0 the:3
3 name:6
6 *:11 may:185
11 *:16 may:171
16 *:21 may:157
21 *:26 may:143
26 *:31 may:129
31 *:36 may:115
36 *:41 may:101
41 *:46 may:87
46 *:51 may:75
51 *:56 may:70
56 may:59
59 not:62
62 be:65
65 used:68
68 m0
70 may:59 not:62
75 *:56 may:70 not:82
82 may:59 be:65
87 *:51 may:75 not:94
94 *:56 may:70 be:65
101 *:46 may:87 not:108
108 *:51 may:75 be:65
115 *:41 may:101 not:122
122 *:46 may:87 be:65
129 *:36 may:115 not:136
136 *:41 may:101 be:65
143 *:31 may:129 not:150
150 *:36 may:115 be:65
157 *:26 may:143 not:164
164 *:31 may:129 be:65
171 *:21 may:157 not:178
178 *:26 may:143 be:65
185 *:16 may:171 not:192
192 *:21 may:157 be:65
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

	// canMisspell - c vs (c) (≡ copyright)
	{`a b c d`, `a b c d`, 0, 4},
	{`a b (c) d`, `a b c d`, 0, 4},
	{`a b copyright d`, `a b c d`, 0, 4},
	{`a b © d`, `a b c d`, 0, 4},
	{`a b c d`, `a b (c) d`, 0, 4},
	{`a b c d`, `a b copyright d`, 0, 4},
	{`a b c d`, `a b © d`, 0, 4},

	// canMisspellJoin
	{`x noninfringement y`, `x non-infringement y`, 0, 4},

	// misspell split
	{`i non-infringement j`, `i noninfringement j`, 0, 3},

	{`a b ((__5__ c d))??`, `a b X X X c d`, 0, 7},
	{`a b ((__5__ c d e))??`, `a b X X X c d e`, 0, 8},

	{`a b __5__ c d ((x?? y?? z z z || y w w w))`, `a b X X X c d y z z z`, 0, 11},
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
