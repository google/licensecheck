// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package licensecheck

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
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
	urls     map[string]License
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
	d.Insert("copyright")
	d.Insert("http")
	var list []*match.LRE
	s.urls = make(map[string]License)
	for _, l := range licenses {
		if l.URL != "" {
			s.urls[l.URL] = l
		}
		if l.Text != "" {
			s.licenses = append(s.licenses, l)
			re, err := match.ParseLRE(d, l.Name, l.Text)
			if err != nil {
				return fmt.Errorf("parsing %v: %v", l.Name, err)
			}
			list = append(list, re)
		}
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

var urlScanRE = regexp.MustCompile(`^(?i)https?://[-a-z0-9_.]+\.(org|com)(/[-a-z0-9_.#?=]+)+/?`)

// Scan is like the top-level function Scan,
// but it uses the set of licenses in the Scanner instead of the built-in license set.
func (s *Scanner) Scan(text []byte) Coverage {
	if s == builtinScanner {
		builtinScannerOnce.Do(func() {
			if err := builtinScanner.init(append(builtinListLRE, builtinURLs...)); err != nil {
				panic("licensecheck: initializing Scan: " + err.Error())
			}
		})
	}

	matches := s.re.Match(string(text)) // TODO remove conversion

	var c Coverage
	words := matches.Words
	total := 0
	lastEnd := 0
	copyright := s.re.Dict().Lookup("copyright")
	http := s.re.Dict().Lookup("http")

	// Add sentinel match trigger URL scan from last match to end of text.
	matches.List = append(matches.List, match.Match{Start: len(words), ID: -1})

	for _, m := range matches.List {
		if m.Start < len(words) && lastEnd < m.Start && copyright >= 0 {
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

		// Pick up any URLs before m.Start.
		for i := lastEnd; i < m.Start; i++ {
			w := &words[i]
			if w.ID == http {
				// Potential URL match.
				// urlRE only considers a match at the start of the input string.
				// Only accept URLs that end before the next scan match.
				if u := urlScanRE.FindIndex(text[w.Lo:]); u != nil && (m.Start == len(words) || int(w.Lo)+u[1] <= int(words[m.Start].Lo)) {
					u0, u1 := int(w.Lo)+u[0], int(w.Lo)+u[1]
					if l, ok := s.licenseURL(string(text[u0:u1])); ok {
						c.Match = append(c.Match, Match{
							Name:    l.Name,
							Percent: 100.0, // TODO
							Start:   u0,
							End:     u1,
							IsURL:   true,
						})
						start := i
						for i < m.Start && int(words[i].Hi) <= u1 {
							i++
						}
						total += i - start
						i-- // counter loop i++
					}
				}
			}
		}

		if m.ID < 0 { // sentinel added above
			break
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
			Percent: 100.0, // TODO
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

// licenseURL reports whether url is a known URL, and returns its name if it is.
func (s *Scanner) licenseURL(url string) (License, bool) {
	// We need to canonicalize the text for lookup.
	// First, trim the leading http:// or https:// and the trailing /.
	// Then we lower-case it.
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/legalcode") // Common for CC licenses.
	url = strings.ToLower(url)
	l, ok := s.urls[url]
	if ok {
		return l, true
	}

	// Try trimming one more path element, so that the ported URL
	//	https://creativecommons.org/licenses/by/3.0/us/
	// is recognized as the known unported URL
	//	https://creativecommons.org/licenses/by/3.0
	if i := strings.LastIndex(url, "/"); i >= 0 {
		if l, ok = s.urls[url[:i]]; ok {
			return l, true
		}
	}

	return License{}, false
}
