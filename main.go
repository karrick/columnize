package main // import "github.com/karrick/columnize"

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/karrick/gobls"
	"github.com/karrick/gologs"
)

func fatal(err error) {
	log.Error("%s", err)
	os.Exit(1)
}

var log *gologs.Logger
var optArgs []string
var optDelimiter string
var optFooterLines, optHeaderLines uint64
var optForce, optLeftJustify, optRightJustify bool

func init() {
	var optDebug, optHelp, optQuiet, optVerbose bool

	// Initialize the global log variable, which will be used very much like the
	// log standard library would be used.
	var err error
	log, err = gologs.New(os.Stderr, gologs.DefaultCommandFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}

	var errs []error

getNextArgument:
	for ai, am := 1, len(os.Args)-1; ai <= am; ai++ {
		if os.Args[ai][0] != '-' {
			optArgs = append(optArgs, os.Args[ai]) // this argument is not an option
			continue
		}

		ail := len(os.Args[ai])

		if ail == 1 {
			optArgs = append(optArgs, os.Args[ai]) // solitary hyphen: implies standard input
			continue
		}

		if os.Args[ai][1] == '-' {
			switch os.Args[ai] {
			case "--":
				return // double hyphen: stop processing command line arguments
			case "--debug":
				optDebug = true
			case "--delimiter":
				if ai == am {
					errs = append(errs, fmt.Errorf("option missing required argument: %q", os.Args[ai]))
					continue
				}
				ai++
				optDelimiter = os.Args[ai]
			case "--footer":
				if ai == am {
					errs = append(errs, fmt.Errorf("option missing required argument: %q", os.Args[ai]))
					continue
				}
				optFooterLines, err = strconv.ParseUint(os.Args[ai+1], 10, 64)
				if err != nil {
					errs = append(errs, fmt.Errorf("footer option: %s", err))
					continue
				}
				ai++
			case "--force":
				optForce = true
			case "--header":
				if ai == am {
					errs = append(errs, fmt.Errorf("option missing required argument: %q", os.Args[ai]))
					continue
				}
				optHeaderLines, err = strconv.ParseUint(os.Args[ai+1], 10, 64)
				if err != nil {
					errs = append(errs, fmt.Errorf("header option: %s", err))
					continue
				}
				ai++
			case "--help":
				optHelp = true
			case "--left":
				optLeftJustify = true
			case "--quiet":
				optQuiet = true
			case "--right":
				optRightJustify = true
			case "--verbose":
				optVerbose = true
			default:
				errs = append(errs, fmt.Errorf("unknown option: %q", os.Args[ai]))
			}
			continue
		}

		for aii := 1; aii < ail; aii++ {
			switch os.Args[ai][aii] {
			case 'd':
				switch {
				case ail-aii > 1:
					optDelimiter = os.Args[ai][aii+1:] // option argument is rest of this argument
				case ai < am:
					ai++
					optDelimiter = os.Args[ai] // option argument is next argument
				default:
					errs = append(errs, fmt.Errorf("option missing required argument: %q", os.Args[ai]))
				}
				continue getNextArgument
			case 'f':
				optForce = true
			case 'h':
				optHelp = true
			case 'l':
				optLeftJustify = true
			case 'q':
				optQuiet = true
			case 'r':
				optRightJustify = true
			case 'v':
				optVerbose = true
			default:
				errs = append(errs, fmt.Errorf("unknown option prefix: %q", os.Args[ai][aii]))
			}
		}
	}

	if optQuiet {
		if optDebug {
			errs = append(errs, fmt.Errorf("cannot use both --quiet and --debug"))
		}
		if optForce {
			errs = append(errs, fmt.Errorf("cannot use both --quiet and --force"))
		}
		if optVerbose {
			errs = append(errs, fmt.Errorf("cannot use both --quiet and --verbose"))
		}
	}

	if optHelp {
		// Show detailed help then exit, ignoring other possibly conflicting
		// options when '--help' is given.
		fmt.Printf(`columnize

Like  'column -t',  but right  justifies numerical  fields.  Reads  input from
multiple files  specified on the command  line or from standard  input when no
files are specified.

SUMMARY:  columnize [options] [file1 [file2 ...]] [options]

USAGE: Not all options may be used with all other options.  See below synopsis
for reference.

    columnize [--quiet | [--debug | --force | --verbose]]
              [--header N]
              [--delimiter STRING]
              [--left | --right]
              [--footer N]
              [file1 [file2 ...]]

EXAMPLES:

    columnize < testdata/bare
    columnize testdata/bare
    columnize benchmarks-a.out benchmarks-b.out
    columnize --header 3 --footer 2 testdata/ignore-headers-footers

Command line options:
  --force
    Print errors to stderr, but keep working.
  -h, --help
    Print command line help and exit.
  -q, --quiet
    Do not print intermediate errors to stderr.
  -v, --verbose
    Print verbose output to stderr.
  -d, --delimiter string (default: "  ")
    output column delimiter
  --footer int (default: 0)
    ignore N lines from footer when formatting columns
  --header int (default: 0)
    ignore N lines from header when formatting columns
  -l, --left
    left-justify all columns
  -r, --right
    right-justify all columns
`)
		os.Exit(0)
	}

	if len(errs) > 0 {
		// Rather than display the entire usage information for a parsing error,
		// merely allow golf library to display the error message, then print
		// the command the user may use to show command line usage information.
		for _, err := range errs {
			log.Error("%s", err)
		}
		log.Error("Use '--help' for more information.")
		os.Exit(2)
	}

	// Configure log level according to command line flags.
	if optDebug {
		log.SetDebug()
	} else if optVerbose {
		log.SetVerbose()
	} else if optQuiet {
		log.SetError()
	} else {
		log.SetInfo()
	}
}

func main() {
	err := forEachFile(optArgs, func(r io.Reader, w io.Writer) error {
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
		err := withOpenFile(file, func(f io.Reader) error {
			return callback(f, os.Stdout)
		})
		if err != nil {
			if !optForce {
				return err
			}
			log.Warning("cannot read %q: %s", file, err)
		}
	}

	return nil
}

func withOpenFile(path string, callback func(io.Reader) error) (err error) {
	if path == "-" {
		return callback(os.Stdin)
	}

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
	cb, err := newTailBuffer(optFooterLines)
	if err != nil {
		return err
	}

	var lines [][]string
	widths := make(map[int]int, 16) // pre-allocate 16 columns

	br := gobls.NewScanner(ior)

	var lineNumber uint64

	for br.Scan() {
		if optHeaderLines > 0 {
			// Only need to count lines while ignoring headers.
			if lineNumber++; lineNumber <= optHeaderLines {
				fmt.Fprintf(iow, "%s\n", br.Text())
				continue
			}
			// No reason to count lines any longer.
			optHeaderLines = 0
		}

		// Recall circular buffer always gives us Nth previous line.
		line := cb.QueueDequeue(br.Text())
		if line == nil {
			continue
		}

		fields := strings.Fields(line.(string))
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
		d := optDelimiter
		for i := 0; i < len(line); i++ {
			// Print newline instead of delimiter for
			// final column.
			if i == len(line)-1 {
				d = "\n"
			}

			field := line[i]
			width := widths[i]

			if optLeftJustify {
				left(iow, width, field, d)
			} else if optRightJustify {
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
