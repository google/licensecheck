// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/google/licensecheck/internal/match"
)

var (
	builtinListLRE []License

	// builtinScanner is initialized lazily,
	// because init is fairly expensive,
	// and delaying it lets us see the init
	// in test cpu profiles.
	builtinScanner     = new(Scanner)
	builtinScannerOnce sync.Once
)

// A Scanner matches a set of known licenses.
type Scanner struct {
	licenses []License
	re       *match.MultiLRE
}

// NewScanner returns a new Scanner that recognizes the given set of licenses.
func NewScanner(licenses []License) (*Scanner, error) {
	s := new(Scanner)
	err := s.init(licenses)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Scanner) init(licenses []License) error {
	d := new(match.Dict)
	var list []*match.LRE
	for _, l := range licenses {
		if l.Text == "" {
			continue
		}
		s.licenses = append(s.licenses, l)
		re, err := match.ParseLRE(d, l.Name, l.Text)
		if err != nil {
			return fmt.Errorf("parsing %v: %v", l.Name, err)
		}
		list = append(list, re)
	}
	re, err := match.NewMultiLRE(list)
	if err != nil {
		return err
	}
	if re == nil {
		return errors.New("missing lre")
	}
	s.re = re
	return nil
}

const maxCopyrightWords = 50

// Scan computes the coverage of the text according to the
// license set compiled into the package.
//
// An input text may match multiple licenses. If that happens,
// Match contains only disjoint matches. If multiple licenses
// match a particular section of the input, the best match
// is chosen so the returned coverage describes at most
// one match for each section of the input.
//
func Scan(text []byte) Coverage {
	return builtinScanner.Scan(text)
}

// Scan is like the top-level function Scan,
// but it uses the set of licenses in the Scanner instead of the built-in license set.
func (s *Scanner) Scan(text []byte) Coverage {
	if s == builtinScanner {
		builtinScannerOnce.Do(func() {
			if err := builtinScanner.init(builtinListLRE); err != nil {
				panic("licensecheck: initializing Scan: " + err.Error())
			}
		})
	}

	matches := s.re.Match(string(text)) // TODO remove conversion
	if matches == nil {
		return Coverage{}
	}

	var c Coverage
	words := matches.Words
	total := 0
	lastEnd := 0
	copyright := s.re.Dict().Lookup("copyright")
	for _, m := range matches.List {
		if lastEnd < m.Start && copyright >= 0 {
			limit := m.Start - maxCopyrightWords
			if limit < lastEnd {
				limit = lastEnd
			}
			for i := limit; i < m.Start; i++ {
				if words[i].ID == copyright {
					m.Start = i
					break
				}
			}
		}

		start := int(words[m.Start].Lo) // byte offset (unlike m.Start)
		if m.Start == 0 {
			start = 0
		} else {
			prev := int(words[m.Start-1].Hi)
			if i := bytes.LastIndexByte(text[prev:start], '\n'); i >= 0 {
				start = prev + i + 1
			}
		}
		end := int(words[m.End-1].Hi) // byte offset (unlike m.End)
		if m.End == len(words) {
			end = len(text)
		} else {
			next := int(words[m.End].Lo)
			if i := bytes.IndexByte(text[end:next], '\n'); i >= 0 {
				end = end + i + 1
			}
		}
		c.Match = append(c.Match, Match{
			Name:    s.licenses[m.ID].Name,
			Percent: 100, // TODO
			Start:   start,
			End:     end,
		})
		total += m.End - m.Start
		lastEnd = m.End
	}
	if len(words) > 0 { // len(words)==0 should be impossible, but avoid NaN
		c.Percent = 100.0 * float64(total) / float64(len(words))
	}

	return c
}
