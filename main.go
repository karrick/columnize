package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/karrick/golf"
	"github.com/karrick/gorill"
)

var (
	optHeaderLines  = golf.Int("header", 0, "ignore N lines from header when formatting columns")
	optFooterLines  = golf.Int("footer", 0, "ignore N lines from footer when formatting columns")
	optDelimiter    = golf.StringP('d', "delimiter", "  ", "output column delimiter")
	optLeftJustify  = golf.BoolP('l', "left", false, "left-justify all columns")
	optRightJustify = golf.BoolP('r', "right", false, "right-justify all columns")
)

func main() {
	optHelp := golf.BoolP('h', "help", false, "Print command line help and exit")
	optIgnoreHeader := golf.BoolP('s', "skip-header", false, "Same as `--header 1`")
	golf.Parse()

	if *optHeaderLines == 0 && *optIgnoreHeader {
		*optHeaderLines = 1
	}

	if *optHelp {
		fmt.Fprintf(os.Stderr, "%s\n", filepath.Base(os.Args[0]))
		if *optHelp {
			fmt.Fprintf(os.Stderr, "        Like `column -t`, but right justifies numerical fields.\n\n")
			fmt.Fprintf(os.Stderr, "Reads input from multiple files specified on the command line or from standard\ninput when no files are specified.\n\n")
			golf.Usage()
		}
		exit(nil)
	}

	var ior io.Reader
	if golf.NArg() == 0 {
		ior = os.Stdin
	} else {
		ior = &gorill.FilesReader{Pathnames: golf.Args()}
	}

	exit(extents(ior))
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

