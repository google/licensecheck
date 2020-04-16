// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/google/licensecheck/internal/match"
)

var (
	builtinListLRE []License
	builtinScanner *Scanner
)

type Scanner struct {
	licenses []License
	re       *match.MultiLRE
}

func NewScanner(licenses []License) (*Scanner, error) {
	d := new(match.Dict)
	s := new(Scanner)
	var list []*match.LRE
	for _, l := range licenses {
		if l.Text == "" {
			continue
		}
		s.licenses = append(s.licenses, l)
		re, err := match.ParseLRE(d, l.Name, l.Text)
		if err != nil {
			return nil, fmt.Errorf("parsing %v: %v", l.Name, err)
		}
		list = append(list, re)
	}
	re, err := match.NewMultiLRE(list)
	if err != nil {
		return nil, err
	}
	if re == nil {
		return nil, errors.New("missing lre")
	}
	s.re = re
	return s, nil
}

const maxCopyrightWords = 50

func Scan(text []byte) Coverage {
	return builtinScanner.Scan(text)
}

func (s *Scanner) Scan(text []byte) Coverage {
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
