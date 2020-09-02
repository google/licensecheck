// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import "testing"

func TestTypeString(t *testing.T) {
	for _, b := range typeBits {
		s := b.t.String()
		if s != b.s {
			t.Errorf("%s.String() = %q, want %q", b.s, s, b.s)
		}
	}

	numError := 0
	for typ := Type(0); typ < Discouraged+100; typ++ {
		s := typ.String()
		ptyp, err := ParseType(s)
		if err != nil {
			t.Errorf("Type(%#x).String = %q; ParseType(...): %v", uint(typ), s, err)
			if numError++; numError > 20 {
				t.FailNow()
			}
			continue
		}
		if ptyp != typ {
			t.Errorf("Type(%#x).String = %q; ParseType(...) = %#x", uint(typ), s, uint(ptyp))
			if numError++; numError > 20 {
				t.FailNow()
			}
			continue
		}
	}
}

var typeMergeTests = []struct {
	t, u Type
	out  Type
}{
	{Unknown, Notice, Unknown},
	{Unknown, NonCommercial, Unknown},
	{Unknown, Discouraged, Unknown},
	{Notice, NonCommercial, Notice | NonCommercial},
	{Notice, ShareProgram, ShareProgram},
}

func TestTypeMerge(t *testing.T) {
	for _, tt := range typeMergeTests {
		if out := tt.t.Merge(tt.u); out != tt.out {
			t.Errorf("(%v).Merge(%v) = %v, want %v", tt.t, tt.u, out, tt.out)
		}
	}
}

func licenseType(id string) Type {
	for _, l := range builtinLREs {
		if l.ID == id {
			return l.Type
		}
	}
	return Unknown
}

var licenseTypeTests = map[string]Type{
	"WTFPL": Discouraged,
}

func TestLicenseType(t *testing.T) {
	for _, l := range BuiltinLicenses() {
		typ, ok := licenseTypeTests[l.ID]
		if !ok {
			typ = Unknown
		}
		if l.Type != typ {
			if l.LRE != "" {
				l.LRE = "..."
			}
			t.Errorf("License%+v: expected Type:%v", l, typ)
		}
	}
}
