package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	leftJustify  = flag.Bool("l", false, "left-justify all columns")
	rightJustify = flag.Bool("r", false, "right-justify all columns")
	delimiter    = flag.String("d", " ", "column delimiter")
	ignoreHeader = flag.Bool("h", false, "ignore header when determining justification")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		if err := process(os.Stdin); err != nil {
			bail(err)
		}
		return
	}
	for _, arg := range args {
		fh, err := os.Open(arg)
		if err != nil {
			bail(err)
		}
		if err := process(fh); err != nil {
			bail(err)
		}
		if err := fh.Close(); err != nil {
			bail(err)
		}
	}
}

func bail(err error) {
	fmt.Fprintf(os.Stderr, "%s", err)
	os.Exit(1)
}

func process(ior io.Reader) error {
	var lines [][]string
	widths := make(map[int]int, 16)
	rightJustifys := make(map[int]bool, 16)

	header := *ignoreHeader
	br := bufio.NewScanner(ior)
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
						fmt.Println("not a number: %v\n", field)
						// not a number; mark this column as left justify
						rightJustifys[i] = false
					} else if !ok {
						fmt.Println("number: %v\n", field)
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
				if rj := rightJustifys[i]; rj {
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
