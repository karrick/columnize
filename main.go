package main // import "github.com/karrick/columnize"

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/karrick/gobls"
	"github.com/karrick/golf"
)

// fatal prints the error to standard error then exits the program with status
// code 1.
func fatal(err error) {
	stderr("%s\n", err)
	os.Exit(1)
}

// newline returns a string with exactly one terminating newline character.
// More simple than strings.TrimRight.  When input string has multiple newline
// characters, it will strip off all but first one, reusing the same underlying
// string bytes.  When string does not end in a newline character, it returns
// the original string with a newline character appended.
func newline(s string) string {
	l := len(s)
	if l == 0 {
		return "\n"
	}

	// While this is O(length s), it stops as soon as it finds the first non
	// newline character in the string starting from the right hand side of the
	// input string.  Generally this only scans one or two characters and
	// returns.
	for i := l - 1; i >= 0; i-- {
		if s[i] != '\n' {
			if i+1 < l && s[i+1] == '\n' {
				return s[:i+2]
			}
			return s[:i+1] + "\n"
		}
	}

	return s[:1] // all newline characters, so just return the first one
}

// stderr formats and prints its arguments to standard error after prefixing
// them with the program name.
func stderr(f string, args ...interface{}) {
	os.Stderr.Write([]byte(ProgramName + ": " + newline(fmt.Sprintf(f, args...))))
}

// usage prints the error to standard error, prints message how to get help,
// then exits the program with status code 2.
func usage(f string, args ...interface{}) {
	stderr(f, args...)
	golf.Usage()
	os.Exit(2)
}

// verbose formats and prints its arguments to standard error after prefixing
// them with the program name.  This skips printing when optVerbose is false.
func verbose(f string, args ...interface{}) {
	if *optVerbose {
		stderr(f, args...)
	}
}

// warning formats and prints its arguments to standard error after prefixing
// them with the program name.  This skips printing when optQuiet is true.
func warning(f string, args ...interface{}) {
	if !*optQuiet {
		stderr(f, args...)
	}
}

var ProgramName string

func init() {
	var err error
	if ProgramName, err = os.Executable(); err != nil {
		ProgramName = os.Args[0]
	}
	ProgramName = filepath.Base(ProgramName)

	// Rather than display the entire usage information for a parsing error,
	// merely allow golf library to display the error message, then print the
	// command the user may use to show command line usage information.
	golf.Usage = func() {
		stderr("Use `%s --help` for more information.\n", ProgramName)
	}
}

var (
	optForce   = golf.Bool("force", false, "Print errors to stderr, but keep working.")
	optHelp    = golf.BoolP('h', "help", false, "Print command line help and exit.")
	optQuiet   = golf.BoolP('q', "quiet", false, "Do not print intermediate errors to stderr.")
	optVerbose = golf.BoolP('v', "verbose", false, "Print verbose output to stderr.")

	optDelimiter    = golf.StringP('d', "delimiter", "  ", "output column delimiter")
	optFooterLines  = golf.Int("footer", 0, "ignore N lines from footer when formatting columns")
	optHeaderLines  = golf.Int("header", 0, "ignore N lines from header when formatting columns")
	optLeftJustify  = golf.BoolP('l', "left", false, "left-justify all columns")
	optRightJustify = golf.BoolP('r', "right", false, "right-justify all columns")
)

func main() {
	golf.Parse()

	if *optHelp {
		// Show detailed help then exit, ignoring other possibly conflicting
		// options when '--help' is given.
		fmt.Printf(`columnize

Like  'column -t',  but  right  justifies numerical  fields.  Reads input  from
multiple files  specified on the  command line or  from standard input  when no
files are specified.

SUMMARY:  columnize [options] [file1 [file2 ...]] [options]

USAGE: Not all options  may be used with all other  options. See below synopsis
for reference.

    columnize [--quiet | [--force | --verbose]]
              [--header N]
              [--delimiter STRING]
              [--left | --right]
              [--footer N]
              [file1 [file2 ...]]

EXAMPLES:

    columnize < sample.txt
    columnize sample.txt
    columnize benchmarks-a.out benchmarks-b.out
    columnize --header 3 --footer 2 testdata/bench.out

Command line options:`)
		golf.PrintDefaults() // frustratingly, this only prints to stderr, and cannot change because it mimicks flag stdlib package
		return
	}

	if *optQuiet {
		if *optForce {
			usage("cannot use both --quiet and --force")
		}
		if *optVerbose {
			usage("cannot use both --quiet and --verbose")
		}
	}

	err := forEachFile(golf.Args(), func(r io.Reader, w io.Writer) error {
		return process(r, os.Stdout)
	})
	if err != nil {
		fatal(err)
	}
}

// forEachFile invokes callback for each file in files. When files is empty, it
// reads from standard input.
func forEachFile(files []string, callback func(io.Reader, io.Writer) error) error {
	if len(files) == 0 {
		return callback(os.Stdin, os.Stdout)
	}

	for _, file := range files {
		err := withOpenFile(file, func(f *os.File) error {
			return callback(f, os.Stdout)
		})
		if err != nil {
			err = fmt.Errorf("cannot read %q: %s", file, err)
			if !*optForce {
				return err
			}
			warning("%s\n", err)
		}
	}

	return nil
}

func withOpenFile(path string, callback func(*os.File) error) (err error) {
	var fh *os.File

	fh, err = os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		if err2 := fh.Close(); err == nil {
			err = err2
		}
	}()

	// Set err variable so deferred function can inspect it.
	err = callback(fh)
	return
}

func process(ior io.Reader, iow io.Writer) error {
	// Use a cirular buffer, so we are processing the Nth previous line.
	cb, err := newTailBuffer(*optFooterLines)
	if err != nil {
		return err
	}

	var lines [][]string
	widths := make(map[int]int, 16) // pre-allocate 16 columns

	br := gobls.NewScanner(ior)

	var lineNumber int

	for br.Scan() {
		if *optHeaderLines > 0 {
			// Only need to count lines while ignoring headers.
			if lineNumber++; lineNumber <= *optHeaderLines {
				fmt.Fprintf(iow, "%s\n", br.Text())
				continue
			}
			// No reason to count lines any longer.
			*optHeaderLines = 0
		}

		// Recall circular buffer always gives us Nth previous line.
		line := cb.QueueDequeue(br.Text())
		if line == nil {
			continue
		}

		fields := strings.Fields(strings.TrimSpace(line.(string)))
		for i, field := range fields {
			if width := len(field); width > widths[i] { // if width wider than previous width
				widths[i] = width // save this width as new widest width for this column
			}
		}
		lines = append(lines, fields)
	}
	if err := br.Err(); err != nil {
		return err
	}

	// All input has been read (and header has even been printed). Pretty print
	// all lines collected thus far, remembering that there may be N lines left
	// in the circular buffer remaining to be processed.
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
				left(iow, width, field, d)
			} else if *optRightJustify {
				right(iow, width, field, d)
			} else {
				// Right justify if number; otherwise
				// left justify
				if _, err := strconv.ParseFloat(field, 64); err == nil {
					right(iow, width, field, d)
				} else {
					left(iow, width, field, d)
				}
			}
		}
	}
	// Dump remaining contents of circular buffer.
	for _, line := range cb.Drain() {
		fmt.Fprintf(iow, "%s\n", line.(string))
	}
	return nil
}

func left(iow io.Writer, width int, field, delimiter string) {
	fmt.Fprintf(iow, "%-*s%s", width, field, delimiter)
}

func right(iow io.Writer, width int, field, delimiter string) {
	fmt.Fprintf(iow, "%*s%s", width, field, delimiter)
}
