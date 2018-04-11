package main

import (
	"strings"
	"unicode"
)

const defaultFieldCount = 16

type extent struct {
	l, r int
}

func (e extent) width() int { return 1 + e.r - e.l }

func extentsFromLine(line string) []extent {
	ee := make([]extent, 0, defaultFieldCount)
	var inWord bool
	var column int   // column within line
	var wordLeft int // column where word starts

	// Loop thru runes in line, splitting into extents.
	for _, r := range line {
		column++
		if unicode.IsSpace(r) != inWord {
			continue // slurping either word or non-word
		}
		inWord = !inWord // toggle state
		if inWord {
			wordLeft = column
		} else {
			ee = append(ee, extent{l: wordLeft, r: column - 1})
		}
	}
	if inWord {
		ee = append(ee, extent{l: wordLeft, r: column})
	}
	return ee
}

// attemptMerge determines whether two extents overlap. If they do, it returns
// the merged extent overlapping both input extents. Otherwise it returns nil.
//
// Ignoring header and footer lines, columnar input would not have
// overlapping extents from different lines. It might very well have missing
// extents for a given line, but if two extents share any columns, they are
// the same extent.
func attemptMerge(ee1, ee2 extent) (extent, bool) {
	if ee1.r < ee2.l {
		return extent{}, false
	}
	if ee2.r < ee1.l {
		return extent{}, false
	}
	minL := ee1.l
	if ee2.l < minL {
		minL = ee2.l
	}
	maxR := ee1.r
	if ee2.r > maxR {
		maxR = ee2.r
	}
	return extent{l: minL, r: maxR}, true
}

// INPUT: two slices of extents, some of which will overlap.
// OUTPUT: consolidated slice of extents, merging the correlated ones
func mergeExtents(ee1, ee2 []extent) []extent {
	var ee1i, ee2i int
	var ee3 []extent

	for {
		if ee1i == len(ee1) {
			ee3 = append(ee3, ee2[ee2i:]...)
			break
		}
		if ee2i == len(ee2) {
			ee3 = append(ee3, ee1[ee1i:]...)
			break
		}
		if ee, ok := attemptMerge(ee1[ee1i], ee2[ee2i]); ok {
			ee3 = append(ee3, ee)
			ee1i++
			ee2i++
			continue
		}
		// not mergeable, so pick smaller one
		if ee1[ee1i].l < ee2[ee2i].l {
			ee3 = append(ee3, ee1[ee1i])
			ee1i++
			continue
		}
		ee3 = append(ee3, ee2[ee2i])
		ee2i++
	}

	return ee3
}

func fieldsFromExtents(line string, extents []extent) []string {
	fields := make([]string, len(extents))

	// recall that extent is column number rather than byte index
	var ei, column, wordStart int

	for _, _ = range line {
		if column++; column < extents[ei].l {
			continue // before the extent starts
		}
		if column == extents[ei].l {
			wordStart = column
			continue
		}
		if column > extents[ei].r {
			fields[ei] = strings.TrimSpace(line[wordStart-1 : column])
			if ei++; ei == len(extents) {
				break // no need keep reading line
			}
		}
	}
	if ei < len(extents) {
		fields[ei] = strings.TrimSpace(line[wordStart-1 : column])
	}

	return fields
}

