# Known Licenses

This directory contains the definitions of the licenses known to [github.com/google/licensecheck](..).

## License Regular Expressions (LREs)

Each license to be recognized is specified by writing a license regular expression (LRE) for it.
The pattern syntax and the matching are word-based and case-insensitive;
punctuation is ignored in the pattern and in the matched text.

The valid LRE patterns are:

 - `word`, a single case-insensitive word
 - `__N__`, any sequence of up to N words
 - `expr1 expr2`, concatenation of two expressions
 - `expr1 || expr2`, alternation of two expressions
 - `(( expr ))`, grouping
 - `(( expr ))??`, zero or one instances of the grouped expression
 - `//** text **//`, a comment ignored by the parser

To make patterns harder to misread in large texts:
`((` must only appear at the start of a line (possibly indented);
`))` and `))??` must only appear at the end of a line (with possible trailing spaces);
and `||` must only appear inside a `(( ))` or `(( ))??` group.

For example:

	//** https://en.wikipedia.org/wiki/Filler_text **//
	Now is
	((not))??
	the time for all good
	((men || women || people))
	to come to the aid of their __1__.

## Built-in licenses & adding new licenses

This package has an extensive set of built-in licenses

## Comparison with SPDX identifiers

Each matched license is identified by a short string, such as `Apache-2.0` for the Apache 2.0 license.
Wherever possible, this package uses the same identifiers that SPDX does.

This section documents the places where this package's identifiers diverge from SPDX.
The exact license patterns used may also differ from SPDX, based on observed real-world licenses.

### AGPL, GFDL, GPL, LGPL

TODO

### BSD

TODO
FreeBSD
UC


### MIT

TODO

### CAL

CAL combined work exception


### Non-SPDX Additions

*Aladdin-9*

*Anti996*

*CC-BY-NC-SA-3.0-US*

*CommonsClause*

*GooglePatentClause*

*GooglePatentsFile*

*NIST*

