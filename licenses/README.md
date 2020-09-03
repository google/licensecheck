# Licensecheck: Built-In Licenses

This directory contains the definitions of the licenses built into [github.com/google/licensecheck](../README.md).
To add new licenses, see the “[Adding new built-in licenses](#adding-new-built-in-licenses)” section below.

It is a goal to incorporate the entire [SPDX license list and IDs](https://spdx.dev/licenses/)
(excluding deprecated license IDs)
as closely as possible while still preserving match accuracy.
It is also a goal to recognize commonly-used non-SPDX or non-open-source licenses
in order to provide accurate information when scanning a large tree of projects.
The result is a strict superset of SPDX.

## Known Licenses

The shortest way to define the known licenses is by reference to the
[SPDX license list v3.10](https://github.com/spdx/license-list-data/tree/v3.10),
with deprecated license IDs omitted.
This section describes the deviations from that base set of licenses.

Many of the license definitions used by licensecheck
have been converted from the SPDX regular expressions
into [LRE patterns](#lre), with the patterns corrected to fix SPDX errors
or revised to match variants found in real-world use.
Those pattern revisions are too numerous to document here.
Instead, this document focuses on the supported licenses and IDs themselves.

### Aladdin Free Public Licnense

SPDX defines the ID `Aladdin` for the Aladdin Free Public License version 8.
Licensecheck defines the non-SPDX ID `Aladdin-9` for version 9,
which was used by the final non-AGPL version of Ghostscript.

_Delta from SPDX_:

 - added `Aladdin-9`

### Anti-996 License

The [Anti-996 license](https://github.com/kattgu7/Anti-996-License)
is a non-open-source license that purports to impose restrictions on
a company's employment practices in exchange for use of the software.
It is used by a variety of GitHub repositories.

_Delta from SPDX_:

 - added `Anti996`

### BSD Licenses

SPDX distinguishes many BSD license variants, which reduce to different subsets of the following clauses:

 - Header (“Redistribution and use ... conditions are met:”)
 - Source (“Redistribution of source code must retain the copyright ...”)
 - Binary (“Redistribution in binary form must reproduce the copyright ... in the documentation”)
 - No-Endorse (“The names of the authors may nobe used to endorse or promote products ...”)
 - Advertising (“All advertising materials mentioning features or use of this software must display the following acknowledgement: ...”)
 - Attribution (“Redistributions of any form whatsoever must retain the following acknowledgement: ...”)
 - No-Trademark (“No license is granted to the trademarks ...”)
 - No-Patent (“No express or implied licenses to any party's patent rights are granted ...”)
 - No-IP (“The copyright holders provide no reassurances ... does not infringe ... intellectual property rights of third parties.”)
 - Patent-Grant (“Subject to the terms ... hereby grants ... patent license ...”)
 - Disclaimer (“The software is provided as is ...”)
 - Views (“The views and conclusions contained in this software ... should not be interpreted as representing official policies of ...”)
 - Bug-Fix ([Published bug fixes imply a license to incorporate unless otherwise noted.])
 - No-Nuclear (“You acknowledge that this software is not designed or intended for use in ... nuclear facility.”)
 - No-Nuclear-License (“You acknowledge that this software is not designed, licensed, or intended for use in ... nuclear facility.”)

SPDX defines the following BSD variants, which licensecheck recognizes and distinguishes:

 - Header, Source, Disclaimer (`BSD-1-Clause`)
 - Header, Source, No-Endorse, Disclaimer (`BSD-Source-Code`)
 - Header, Source, Binary, Disclaimer (`BSD-2-Clause`)
 - Header, Source, Binary, Disclaimer, Views (`BSD-2-Clause-Views`)
 - Header, Source, Binary, Patent-Grant, Disclaimer (`BSD-2-Clause-Patent`)
 - Header, Source, Binary, No-Endorse, Disclaimer (`BSD-3-Clause`)
 - Header, Source, Binary, No-Endorse, Disclaimer, Bug-Fix (`BSD-3-Clause-LBNL`)
 - Header, Source, Binary, No-Endorse, Attribution, Disclaimer (`BSD-3-Clause-Attribution`)
 - Header, Source, Binary, No-Endorse, No-Patent, Disclaimer (`BSD-3-Clause-Clear`)
 - Header, Source, Binary, No-Endorse, No-IP, Disclaimer (`BSD-3-Clause-Open-MPI`)
 - Header, Source, Binary, No-Endorse, Advertising, Disclaimer (`BSD-4-Clause`)
 - Header, Source, Binary, No-Endorse, Advertising, Disclaimer issued by University of California (`BSD-4-Clause-UC`)
 - Header, Source, Binary, No-Endorse, custom Disclaimer, No-Nuclear (`BSD-3-No-Nuclear-Warranty`)
 - Header, Source, Binary, No-Endorse, custom Disclaimer, No-Nuclear-License (`BSD-3-No-Nuclear-License`)
 - Header, Source, Binary, No-Endorse, Disclaimer, No-Nuclear-License (`BSD-3-Clause-No-Nuclear-License-2014`)

In addition to the above, licensecheck recognizes the following non-SPDX variants:

 - Header, Source, No-Patent, Disclaimer (`BSD-1-Clause-Clear`)
 - Header, Source, Binary, No-Endorse, No-Trademark, Disclaimer (`BSD-3-Clause-NoTrademark`)

[SPDX defines a `BSD-Protection` license, recognized by licensecheck,
which is a different license entirely (approximately “BSD made viral”)
and is therefore omitted from the above discussion.]

_Delta from SPDX_:

 - added `BSD-1-Clause-Clear`
 - added `BSD-3-Clause-NoTrademark`

### Cryptography Autonomy License

The Cryptographic Autonomy License version 1.0 allows source files to be
marked as subject to the “combined work exception,” which stops certain conditions
from applying to those files.
SPDX defines `CAL-1.0` and `CAL-1.0-Combined-Work-Exception`
to distinguish the two license types for a given file.
This per-file annotation is not visible in the license text,
so licensecheck can only report `CAL-1.0`.

_Delta from SPDX_:

 - never reports `CAL-1.0-Combined-Work-Exception`

### Commons Clause

The [Commons Clause](https://commonsclause.com/) is a license condition
that introduces a commercial-use restriction on top of an otherwise open-source license.
That is, the presence of the Commons Clause changes an open-source license
into a non-open license that does not permit commercial use.
It is used by Redis Labs and other companies.
Licensecheck reports it as `CommonsClause`.

_Delta from SPDX_:

 - added `CommonsClause`

### Creative Commons

Licensecheck supports all the Creative Commons licenses with assigned SPDX identifiers
and then adds the United States port of the Attribution-NonCommercial-ShareAlike 3.0 license
(`CC-BY-NC-SA-3.0-US`),
which is used by a variety of GitHub repositories.

_Delta from SPDX_:

 - added `CC-BY-NC-SA-3.0-US`

### GNU General Public Licenses (AGPL, GPL, LGPL)

For each version of each of these GNU licenses,
SPDX defines a pair of IDs defining whether a newer
version of the license may be chosen by the licensee:
`$LICENSE-only` and `$LICENSE-or-later`.
For example, the two IDs for AGPL-3.0 are `AGPL-3.0-only` and `AGPL-3.0-or-later`.
These distinguish the two different possible file headers, which say one of:

 - ... version 3 of the License.
 - ... either version 3 of the License, or (at your option) any later version.

But another common situation is finding the text of, say,
the AGPL 3.0 in a `LICENSE` file but without any file header.
In this case, it is unclear whether the license is `AGPL-3.0-only` or `AGPL-3.0-or-later`.
Licensecheck exposes this ambiguity as a new license type: `AGPL-3.0` (unsuffixed).
The same holds for all the other AGPL, GPL, and LGPL versions.

Another common variation found in the wild is license notices permitting
LGPL version 2.0 or 3.0 (not 2.0 only; not 2.0 or later).
For that, licensecheck defines `LGPL-2.0-or-3.0`.

_Delta from SPDX_:

 - added `AGPL-1.0`, `AGPL-3.0` for license text (not header)
 - added `GPL-1.0`, `GPL-2.0`, `GPL-3.0` for license text (not header)
 - added `LGPL-2.0`, `LGPL-2.1`, `LGPL-3.0` for license text (not header)
 - added `LGPL-2.0-or-3.0`

### GNU Free Documentation License (GFDL)

SPDX splits each GFDL version into six different variants.
It first splits each into two groups: “exact version only” versus “or later.”
It then splits each group into three variants:
“has invariant text,” “does not have invariant text,” and “don't know about invariant text or not.”

For example, there are six SPDX IDs for GFDL-1.3:
`GFDL-1.3-or-later`,
`GFDL-1.3-invariants-or-later`,
`GFDL-1.3-no-invariants-or-later`,
`GFDL-1.3-only`,
`GFDL-1.3-invariants-only`,
and
`GFDL-1.3-no-invariants-only`.

Licensecheck distinguishes whether or not invariant text is present,
so it always reports either the `-invariants-` or `-no-invariants-` forms.
It never reports the non-specific `GFDL-1.3-or-later` or `GFDL-1.3-only` forms.

As happened with the GPL variants, it is unclear which SPDX ID to use when
presented with the text of the license itself instead of a header.
Licensecheck adds the IDs `GFDL-1.1`, `GFDL-1.2`, and `GFDL-1.3`
to denote finding the license itself, not a header.

_Delta from SPDX_:

 - never reports `GFDL-1.1-only`, `GFDL-1.1-or-later`; added `GFDL-1.1` for license text (not header)
 - never reports `GFDL-1.2-only`, `GFDL-1.2-or-later`; added `GFDL-1.2` for license text (not header)
 - never reports `GFDL-1.3-only`, `GFDL-1.3-or-later`; added `GFDL-1.3` for license text (not header)

### Google Patents

For a brief time, Go's LICENSE file contained a BSD license followed by
an additional explicit patent grant similar to the one used by the Apache-2.0 license.
Licensecheck reports that patent grant text as `GooglePatentClause`.

Go then deleted the grant from the LICENSE file and moved it to a separate PATENTS file,
updating the language.
This file is used by a variety of other BSD-license Google projects as well.
Licensecheck reports that updated language as `GooglePatentsFile`.

Other companies have adopted the PATENTS file for their own BSD-licensed code,
replacing “Google” with (or adding) their own company name.
These variants are also reported as `GooglePatentsFile`, even when the company is not Google
(just as the MIT license need not cover code from MIT).

Delta from SPDX:

 - added `GooglePatentClause`, `GooglePatentsFile`

### MIT License Variants

Licensecheck supports all the SPDX-defined MIT variants: `MIT`, `MIT-0`, `MITNFA`, and `0BSD` [_sic_].

Additionally, licensecheck defines a variant `MIT-NoAd` that adds a non-advertising clause similar to the BSD “No-Promote” clause.

_Delta from SPDX_:

 - added `MIT-NoAd`

### Prosperity

The [Prosperity Public License](https://prosperitylicense.com/)
by [License Zero](https://licensezero.com/)
is a non-commercial companion to the [Parity license](https://paritylicense.com/).
It is a non-open-source license but found in use on GitHub.

Licensecheck assigns the Prosperity Public License 3.0.0 the ID `Prosperity-3.0.0`.

_Delta from SPDX_:

 - added `Prosperity-3.0.0`

### SIL Open Font License (OFL)

SPDX splits OFL-1.0 and OFL-1.1 into three variants each,
depending on whether a reserved font name applies.
The IDs OFL-1.0 and OFL-1.1 make no claim whether a reserved font name applies;
OFL-1.0-RFN and OFL-1.1-RFN say it does;
OFL-1.0-no-RFN and OFL-1.1-no-RFN say it doesn't.

Once again, the SPDX patterns for these variants do not accurately capture
the intended distinction, making any result from a license scanner using those patterns
suspect.

To avoid that confusion, licensecheck does not attempt to use the `-RFN` and `-no-RFN` variants.
It only defines and reports `OFL-1.0` and `OFL-1.1`.

_Delta from SPDX_:

 - never reports `OFL-1.0-RFN`, `OFL-1.0-no-RFN`; always uses `OFL-1.0`
 - never reports `OFL-1.1-RFN` and `OFL-1.1-no-RFN`; always uses `OFL-1.1`

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

## Adding new built-in licenses

This package has an extensive set of built-in licenses,
defined by the `*.lre` files in this directory.
(See [README.md](README.md) for details about the choice of licenses.)

The content of each file is
[text/template](https://pkg.go.dev/text/template) input
that generates LRE output,
so that common pieces can be factored out
(see, for example, [BSD.lre](BSD.lre)).

After editing files in this directory, run `go generate` in the licensecheck (parent) directory.

Note that when using
[licensecheck.NewScanner](https://pkg.go.dev/github.com/google/licensecheck/#NewScanner),
the input is plain LRE, not template text.
