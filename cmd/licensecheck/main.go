// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command licensecheck provides a command-line interface to the licensecheck
// package. Given a file, it prints any licenses found in it, one per line,
// along with the percentage of the license text that matched. It exits with a
// non-zero status code on error; finding no licenses is not considered an
// error.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/licensecheck"
)

var verbose = flag.Bool("v", false, "print additional details for license matches")

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	filename := flag.Arg(0)
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "licensecheck: %v\n", err)
		os.Exit(1)
	}
	options := licensecheck.Options{
		MinLength: 10,
		Threshold: 40,
		Slop:      8,
	}
	coverage, ok := licensecheck.Cover(contents, options)
	if ok {
		for _, m := range coverage.Match {
			if *verbose {
				fmt.Printf("%s\t%s\t%f%%\t%d\t%d\t%t\n", m.Name, m.Type, m.Percent, m.Start, m.End, m.IsURL)
			} else {
				fmt.Printf("%s\t%f%%\n", m.Name, m.Percent)
			}
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: licensecheck [-v] <file>\n")
	flag.PrintDefaults()
}
