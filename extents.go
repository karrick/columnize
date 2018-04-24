package main

import (
	"strings"
	"unicode"
)

const defaultFieldCount = 16

type extent struct {
	lc, rc int // column boundaries of word
}

func (e extent) width() int { return 1 + e.rc - e.lc }

func extentsFromLine(line string) []extent {
	ee := make([]extent, 0, defaultFieldCount)
	var inWord bool
	var column int // column within line
	var lc int     // column where word starts

	// Loop thru runes in line, splitting into extents.
	for _, r := range line {
		column++
		if unicode.IsSpace(r) != inWord {
			continue // no change; continue slurping
		}
		inWord = !inWord // toggle state
		if inWord {
			// store column and index where word began
			lc = column
		} else {
			// no longer in a word; store field extent
			ee = append(ee, extent{
				lc: lc,
				rc: column - 1,
			})
		}
	}
	if inWord {
		ee = append(ee, extent{
			lc: lc,
			rc: column,
		})
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
	if ee1.rc < ee2.lc {
		return extent{}, false
	}
	if ee2.rc < ee1.lc {
		return extent{}, false
	}
	minL := ee1.lc
	if ee2.lc < minL {
		minL = ee2.lc
	}
	maxR := ee1.rc
	if ee2.rc > maxR {
		maxR = ee2.rc
	}
	return extent{lc: minL, rc: maxR}, true
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
		if ee1[ee1i].lc < ee2[ee2i].lc {
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
	// Walk through the line, rune by rune, tracking the column. Note the byte
	// index into the line may change by more than one byte, depending on the
	// width of a particular rune.
	//
	// When the column number lines up with the left edge of the first extent,
	// track that column position.

	// Recall that extent is column based rather than index of byte.
	var column, wordStart int
	var index int // index into both extents and fields slices

	fields := make([]string, len(extents))

	for range line {
		if column++; column < extents[index].lc {
			continue // before the extent starts
		}
		if column == extents[index].lc {
			wordStart = column
			continue
		}
		if column > extents[index].rc {
			fields[index] = strings.TrimSpace(line[wordStart-1 : column])
			if index++; index == len(extents) {
				break // no need to keep reading line
			}
		}
	}
	if wordStart > 0 && index < len(extents) {
		// If started tracking a word, but did not find the end (because word
		// ends at the end of the line)
		fields[index] = strings.TrimSpace(line[wordStart-1 : column])
	}

	return fields
}
