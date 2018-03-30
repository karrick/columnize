package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/karrick/gobls"
	"github.com/karrick/golf"
	"github.com/karrick/gorill"
)

var (
	delimiter    = golf.StringP('d', "delimiter", " ", "column delimiter")
	ignoreHeader = golf.BoolP('s', "skip-header", false, "skip header when determining justificationn")
	leftJustify  = golf.BoolP('l', "left", false, "left-justify all columns")
	rightJustify = golf.BoolP('r', "right", false, "right-justify all columns")
)

func main() {
	golf.Parse()

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
	var lines [][]string
	widths := make(map[int]int, 16)
	rightJustifys := make(map[int]bool, 16)

	header := *ignoreHeader
	br := gobls.NewScanner(ior)
	for br.Scan() {
		fields := strings.Fields(strings.TrimSpace(br.Text()))
		for i, field := range fields {
			width := len(field)
			previousWidth := widths[i]
			if width > previousWidth {
				widths[i] = width
			}
			if !header && !(*leftJustify || *rightJustify) {
				// NOTE: If either first time this column observed, i.e., likely
				// only for first line of input, or all previous fields in this
				// column have been numbers...
				if rj, ok := rightJustifys[i]; !ok || rj {
					_, err := strconv.ParseFloat(field, 64)
					if err != nil {
						// not a number; mark this column as left justify
						rightJustifys[i] = false
					} else if !ok {
						// first time column observed, and is a number
						rightJustifys[i] = true
					}
				}
			}
		}
		lines = append(lines, fields)
		header = false
	}
	if err := br.Err(); err != nil {
		return err
	}
	for _, line := range lines {
		d := *delimiter
		for i := 0; i < len(line); i++ {
			if i == len(line)-1 {
				d = "" // do not emit trailing delimiter
			}

			field := line[i]
			width := widths[i]

			if *leftJustify {
				left(width, field, d)
			} else if *rightJustify {
				right(width, field, d)
			} else {
				if rightJustifys[i] {
					right(width, field, d)
				} else {
					left(width, field, d)
				}
			}
		}
		fmt.Println()
	}
	return nil
}

func left(width int, field, delimiter string) {
	fmt.Printf("%-*s%s", width, field, delimiter)
}
func right(width int, field, delimiter string) {
	fmt.Printf("%*s%s", width, field, delimiter)
}
