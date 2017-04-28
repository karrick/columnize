package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	var lines [][]string
	maxs := make(map[int]int, 16)
	br := bufio.NewScanner(os.Stdin)
	for br.Scan() {
		fields := strings.Fields(strings.TrimSpace(br.Text()))
		for i, field := range fields {
			length := len(field)
			if previousMax := maxs[i]; length > previousMax {
				maxs[i] = length
			}
		}
		lines = append(lines, fields)
	}
	if err := br.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	for _, line := range lines {
		for i, field := range line {
			length := maxs[i]
			format := fmt.Sprintf("% %%ds  ", length)
			fmt.Printf(format, field)
		}
		fmt.Println()
	}
}
