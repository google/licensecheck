// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Getcc converts Creative Commons license pages into license regular expressions (LREs).
//
// Usage:
//
//	go run getcc.go [-f] name...
//
// Getcc converts each XHTML file into an LRE file id.lre, where id is a Creative Commons license ID
// (like CC-BY-SA-NC-4.0).
//
// Getcc is only intended to provide a good start for the LRE for a given license.
// The result of the conversion may still need manual adjustment over time to deal
// with real-world variation, although less than the non-CC licenses.
//
// If id.lre already exists, getcc skips the conversion instead of overwriting id.lre.
// If the -f flag is given, getcc overwrites id.lre.
//
// Getcc expects to find the Creative Commons web site checked out in _cc
// and the SPDX database checked out in _spdx, which you can do using:
//
//	git clone https://github.com/creativecommons/creativecommons.org _cc
//
package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var forceOverwrite = flag.Bool("f", false, "force overwrite")

func usage() {
	fmt.Fprintf(os.Stderr, "usage: go run getcc.go [-f] name\n")
	os.Exit(2)
}

var exitStatus int

func main() {
	log.SetPrefix("getcc: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
	}

	if info, err := os.Stat("_spdx"); err != nil || !info.IsDir() {
		log.Fatalf("expected SPDX database in _spdx; check out with:\n\tgit clone https://github.com/spdx/license-list-data _spdx")
	}
	if info, err := os.Stat("_cc"); err != nil || !info.IsDir() {
		log.Fatalf("expected Creative Commons web site in _cc; check out with:\n\tgit clone https://github.com/creativecommons/creativecommons.org _cc")
	}

	for _, file := range args {
		convert(file)
	}

	cmd := exec.Command("go", "generate")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	os.Exit(exitStatus)
}

func convert(id string) {
	// Convert CC name to URL basename.
	// Lowercase and convert - to _ starting just before the number.
	file := strings.TrimPrefix(id, "CC-")
	i := strings.IndexAny(file, "0123456789")
	if i > 0 {
		file = file[:i-1] + strings.ReplaceAll(file[i-1:], "-", "_")
	}
	file = strings.ToLower(file)

	i = strings.Index(file, "_")
	j := strings.Index(file[i+1:], "_")
	if j >= 0 {
		j += i + 1
	} else {
		j = len(file)
	}
	var url string
	if i >= 0 {
		url = "https://creativecommons.org/licenses/" + file[:i] + "/" + file[i+1:j]
		if j < len(file) {
			url += "/" + file[j+1:]
		}
	}

	data, err := ioutil.ReadFile("_cc/docroot/legalcode/" + file + ".html")
	if err != nil {
		log.Print(err)
		exitStatus = 1
		return
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "//**\n%s\n%s\n", id, url)
	fmt.Fprintf(&buf, "**//\n\n((Creative Commons))??\n\n")

	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	// find div id=deed
	// id=deed-head
	//	id=deed-license h2
	// deed.deed-head.deed-license
	// deed.deed-main.deed-main-content
	//	blockquote optional
	//	h3 optional
	//	ol/li turns into annotated wildcard
	//	look for numbers at start of lines?
	//	blockquote at end optional
	//
	// deed.p align center heading
	//	deed.text fineprint optional
	//	License optional
	//	deed.text findprint optional

	wanted := []string{
		"div id=deed-license",
		"div id=deed-main-content",
		"h1", "h2",
	}
	if !bytes.Contains(data, []byte(`id="deed-`)) {
		wanted = append(wanted,
			"p align=center > b",
			"p align=center > a",
			"div class=text",
		)
	}

	var stack []string
	var undoStack []func()
	want := func() bool {
		context := strings.Join(stack, " > ")
		for _, w := range wanted {
			if strings.Contains(context, w) {
				return true
			}
		}
		return false
	}
	for {
		t, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("reading %s: %v", file, err)
			exitStatus = 1
			return
		}
		switch t := t.(type) {
		case xml.StartElement:
			undo := func() {}
			desc := t.Name.Local
			var nl bool
			var optional bool
			extra := ""
			switch t.Name.Local {
			case "h1", "h2", "h3":
				nl = true
				optional = true
				if url != "" {
					extra = "((" + url + "\n((/legalcode))??\n))??\n"
				}
			case "div":
				if id := find(t.Attr, "id"); id != "" {
					desc += " id=" + id
				}
				if class := find(t.Attr, "class"); class != "" {
					desc += " class=" + class
					optional = class == "fineprint" || class == "shaded" || class == "summary"
				}
				nl = true
			case "blockquote":
				nl = true
				optional = true
			case "p":
				nl = true
				if align := find(t.Attr, "align"); align != "" {
					desc += " align=" + align
					if align == "center" {
						optional = true
					}
				}
				optional = optional || find(t.Attr, "class") == "shaded"
			case "li":
				if !want() {
					break
				}
				buf.WriteString("\n__1__ ")
				undo = func() {
					buf.WriteString("\n")
				}
			}
			stack = append(stack, desc)
			if nl && want() {
				buf.WriteString("\n")
				if optional {
					buf.WriteString("((\n")
				}
				undo = func() {
					if optional {
						buf.WriteString("\n))??")
					}
					buf.WriteString("\n")
					buf.WriteString(extra)
				}
			}
			undoStack = append(undoStack, undo)
		case xml.EndElement:
			switch t.Name.Local {
			case "p", "li", "div":
				buf.WriteString("\n")
			}
			stack = stack[:len(stack)-1]
			undo := undoStack[len(undoStack)-1]
			undo()
			undoStack = undoStack[:len(undoStack)-1]
		case xml.CharData:
			s := string(t)
			if s == "" || !want() {
				continue
			}

			// A handful of the licenses are formatted like this:
			//
			//	1. You must:
			//
			//		a. One thing.
			//		b. Another thing.
			//		c. Yet another thing.
			//
			//	   For the avoidance of doubt, you need not something else.
			//
			//	2. More conditions.
			//
			// For whatever reason, some text copies make the "For the avoidance of doubt"
			// part another entry in the sublist (d. in this case).
			// So let it have a list item prefix, even though it isn't one.
			if strings.HasPrefix(strings.TrimSpace(s), "For the avoidance of doubt") &&
				!strings.HasSuffix(strings.TrimSpace(buf.String()), "__1__") {
				s = "__1__ " + strings.TrimLeft(s, " \n")
			}

			s = strings.ReplaceAll(s, "((", " ( ( ")
			s = strings.ReplaceAll(s, "))", " ) ) ")
			if s == "License" {
				s = "((License))??"
			}
			wrap(&buf, s)
		}
	}

	target := id + ".lre"
	if _, err := os.Stat(target); err == nil && !*forceOverwrite {
		return
	}
	text := regexp.MustCompile(`[ \t]*\n[ \t]*`).ReplaceAll(buf.Bytes(), []byte("\n"))
	text = regexp.MustCompile(`\n\n\n+`).ReplaceAll(text, []byte("\n\n"))
	text = regexp.MustCompile(`\(\(\n\n+`).ReplaceAll(text, []byte("((\n"))
	text = regexp.MustCompile(`\n\n+\)\)`).ReplaceAll(text, []byte("\n))"))
	text = bytes.TrimSpace(text)
	text = append(text, '\n')

	if len(text) < 1000 {
		log.Printf("%s: conversion too short\n", id)
		exitStatus = 1
		return
	}

	if err := ioutil.WriteFile(target, text, 0666); err != nil {
		log.Print(err)
		exitStatus = 1
		return
	}

	text, err = ioutil.ReadFile("_cc/docroot/legalcode/" + file + ".txt")
	if err != nil {
		// In absence of a canonical form, take whatever SPDX has.
		text, err = ioutil.ReadFile("_spdx/text/" + id + ".txt")
		if err != nil {
			log.Printf("cannot find text for %s in _cc or _spdx", id)
			exitStatus = 1
			return
		}
	}

	if err := ioutil.WriteFile("../testdata/licenses/"+id+".txt", text, 0666); err != nil {
		log.Print(err)
		exitStatus = 1
		return
	}
	if _, err := os.Stat("../testdata/" + id + ".t1"); err != nil {
		data := []byte(fmt.Sprintf("0%%\nscan\n100%%\n%s 100%% 0,$\n\n%s", id, text))
		if err := ioutil.WriteFile("../testdata/"+id+".t1", data, 0666); err != nil {
			log.Print(err)
			exitStatus = 1
			return
		}
	}
}

func find(list []xml.Attr, key string) string {
	for _, a := range list {
		if a.Name.Local == key {
			return a.Value
		}
	}
	return ""
}

// wrap adds literal text to the buffer buf, wrapping long lines.
// Wrapping is important for reading future diffs in the LRE files.
func wrap(buf *bytes.Buffer, text string) {
	all := buf.Bytes()
	i := len(all)
	for i > 0 && all[i-1] != '\n' {
		i--
	}
	buf.Truncate(i)
	lines := strings.SplitAfter(text, "\n")
	lines[0] = string(all[i:]) + lines[0]
	for _, line := range lines {
		i := 0
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		indent := line[:i]
		line = line[i:]
		const Target = 80
		for len(line) > Target-len(indent) {
			j := Target - len(indent)
			for j >= 0 && line[j] != ' ' && line[j] != '\t' {
				j--
			}
			if j < 0 {
				j = Target - len(indent)
				for j < len(line) && line[j] != ' ' && line[j] != '\t' {
					j++
				}
				if j == len(line) {
					break
				}
			}
			buf.WriteString(indent)
			buf.WriteString(line[:j])
			buf.WriteString("\n")
			for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
				j++
			}
			line = line[j:]
		}
		buf.WriteString(indent)
		buf.WriteString(line)
	}
}
