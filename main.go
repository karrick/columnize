package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/karrick/gobls"
	"github.com/karrick/golf"
	"github.com/karrick/gorill"
)

var (
	optDelimiter    = golf.StringP('d', "delimiter", "  ", "output column delimiter")
	optHeaderLines  = golf.Int("header", 0, "ignore N lines from header when formatting columns")
	optFooterLines  = golf.Int("footer", 0, "ignore N lines from footer when formatting columns")
	optLeftJustify  = golf.BoolP('l', "left", false, "left-justify all columns")
	optRightJustify = golf.BoolP('r', "right", false, "right-justify all columns")
)

func main() {
	optHelp := golf.BoolP('h', "help", false, "Print command line help and exit")
	optIgnoreHeader := golf.BoolP('s', "skip-header", false, "Same as --ignore-head 1")
	golf.Parse()
	if *optHeaderLines == 0 && *optIgnoreHeader {
		*optHeaderLines = 1
	}

	if *optHelp {
		fmt.Fprintf(os.Stderr, "%s\n", filepath.Base(os.Args[0]))
		if *optHelp {
			fmt.Fprintln(os.Stderr, "        Like `column -t`, but right justifies columns that are all numbers.")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Reads input from multiple files specified on the command line or from standard input when no files are specified.")
			fmt.Fprintln(os.Stderr)
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

	exit(process(ior))
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func process(ior io.Reader) error {
	cb, err := newCircularBuffer(*optFooterLines)
	if err != nil {
		return err
	}

	var lines [][]string
	widths := make(map[int]int, 16) // pre-allocate 16 columns

	br := gobls.NewScanner(ior)

	var lineNumber int

	for br.Scan() {
		if *optHeaderLines > 0 {
			// only need to count lines while ignoring headers
			if lineNumber++; lineNumber <= *optHeaderLines {
				fmt.Printf("%s\n", br.Text())
				continue
			}
		}

		// Use a cirular buffer, so we are processing the Nth
		// previous line.
		line := cb.QueueDequeue(br.Text())
		if line == nil {
			continue
		}

		fields := strings.Fields(strings.TrimSpace(line.(string)))
		for i, field := range fields {
			if width := len(field); width > widths[i] { // if width wider than previous width
				widths[i] = width // save this width as new widest width
			}
		}
		lines = append(lines, fields)
	}
	if err := br.Err(); err != nil {
		return err
	}
	// All input has been read (and header has even been printed).
	for _, line := range lines {
		d := *optDelimiter
		for i := 0; i < len(line); i++ {
			// Print newline instead of delimiter for
			// final column.
			if i == len(line)-1 {
				d = "\n"
			}

			field := line[i]
			width := widths[i]

			if *optLeftJustify {
				left(width, field, d)
			} else if *optRightJustify {
				right(width, field, d)
			} else {
				// Right justify if number; otherwise
				// left justify
				if _, err := strconv.ParseFloat(field, 64); err == nil {
					right(width, field, d)
				} else {
					left(width, field, d)
				}
			}
		}
	}
	// Dump remaining contents of circular buffer.
	for {
		line := cb.QueueDequeue(nil)
		if line == nil {
			break
		}
		fmt.Printf("%s\n", line.(string))
	}
	return nil
}

func left(width int, field, delimiter string) {
	fmt.Printf("%-*s%s", width, field, delimiter)
}
func right(width int, field, delimiter string) {
	fmt.Printf("%*s%s", width, field, delimiter)
}
