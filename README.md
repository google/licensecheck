The licensecheck package classifies license files and heuristically
determines how well they correspond to known open source licenses.

The design aims never to give a false positive. It uses a soft text
matching algorithm, though, which means it can miss valid licenses
that have been modified more than usual. It also recognizes licenses
identified by URL.

To use this package, the client first identifies a file or text
that might contain a license, then calls the Cover function for a
description of what licenses are present in the text. The description
returned by Cover lists the licenses in the order they appear in
the text. By convention in most open source repositories, the first
license present describes the main body of the work, while subsequent
licenses apply to borrowed subcomponents.

The output of `go doc Cover` and `go doc Coverage` provides more
information:

```
func Cover(input []byte, opts Options) (Coverage, bool)
    Cover computes the coverage of the text according to the license set
    compiled into the package.

    An input text may match multiple licenses. If that happens, Match contains
    only disjoint matches. If multiple licenses match a particular section of
    the input, the best match is chosen so the returned coverage describes at
    most one match for each section of the input.

type Coverage struct {
	// Percent is the fraction of the total text, in normalized words, that
	// matches any valid license, expressed as a percentage across all of the
	// licenses matched.
	Percent float64

	// Match describes, in sequential order, the matches of the input text
	// across the various licenses. Typically it will be only one match long,
	// but if the input text is a concatenation of licenses it will contain
	// a match value for each element of the concatenation.
	Match []Match
}
    Coverage describes how the text matches various licenses.
```

The licenses subdirectory contains canonical forms of the licenses
recognized by the package. The list covers most open source software
but is not comprehensive. More licenses may be added by copying
their text to the subdirectory as individual files and running `go
generate` in the root.
