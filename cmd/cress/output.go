package main

import (
	"fmt"
	"io"
)

// fprintf and fprintln write cress's CLI output. A failed write to the terminal
// is unrecoverable and not worth surfacing, so the error is ignored on purpose
// — keeping the write error out of errcheck's way without a config exclusion.
func fprintf(w io.Writer, format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }

func fprintln(w io.Writer, a ...any) { _, _ = fmt.Fprintln(w, a...) }
