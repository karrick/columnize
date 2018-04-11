package main

import (
	"strings"
	"testing"
)

func TestExtentsEmptyString(t *testing.T) {
	ee := extentsFromLine("")
	if got, want := len(ee), 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsHandlesMoreThanDefaultFieldCount(t *testing.T) {
	ee := extentsFromLine(strings.Repeat(" item ", defaultFieldCount+1))
	if got, want := len(ee), defaultFieldCount+1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsSingleItemNoWhiteSpace(t *testing.T) {
	ee := extentsFromLine("item")
	if got, want := len(ee), 1; got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].l, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].r, 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsSingleItemWithWhiteSpace(t *testing.T) {
	ee := extentsFromLine(" item ")
	if got, want := len(ee), 1; got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].l, 2; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].r, 5; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsTwoItemsWithoutWhiteSpace(t *testing.T) {
	ee := extentsFromLine("one two")
	if got, want := len(ee), 2; got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	// one
	if got, want := ee[0].l, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].r, 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// two
	if got, want := ee[1].l, 5; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[1].r, 7; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsTwoItemsWithWhitespace(t *testing.T) {
	ee := extentsFromLine(" one two ")
	if got, want := len(ee), 2; got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	// one
	if got, want := ee[0].l, 2; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].r, 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// two
	if got, want := ee[1].l, 6; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[1].r, 8; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

////////////////////////////////////////

func TestExtentsMergeWithoutOverlap(t *testing.T) {
	ee, ok := attemptMerge(extent{l: 11, r: 15}, extent{l: 6, r: 8})

	if got, want := ok, false; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.l, 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.r, 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsMergeWithOverlap(t *testing.T) {
	ee, ok := attemptMerge(extent{l: 1, r: 3}, extent{l: 2, r: 3})

	if got, want := ok, true; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.l, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.r, 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

////////////////////////////////////////

func TestMergeExtentSlices(t *testing.T) {
	ee1 := []extent{extent{l: 1, r: 3}, extent{l: 11, r: 15}}
	ee2 := []extent{extent{l: 2, r: 5}, extent{l: 6, r: 8}, extent{l: 10, r: 12}}

	ee3 := mergeExtents(ee1, ee2)

	if got, want := len(ee3), 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// first
	if got, want := ee3[0].l, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[0].r, 5; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// second
	if got, want := ee3[1].l, 6; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[1].r, 8; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// third
	if got, want := ee3[2].l, 10; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[2].r, 15; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

////////////////////////////////////////

func TestTwoLines(t *testing.T) {
	ee1 := extentsFromLine("one    two    three")
	ee2 := extentsFromLine(" one            three")
	ee3 := mergeExtents(ee1, ee2)
	
	if got, want := len(ee3), 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// first
	if got, want := ee3[0].l, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[0].r, 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[0].width(), 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// second
	if got, want := ee3[1].l, 8; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[1].r, 10; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[1].width(), 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// third
	if got, want := ee3[2].l, 15; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[2].r, 21; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[2].width(), 7; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

////////////////////////////////////////

func TestFieldsFromExtents(t *testing.T) {
	line := "one    two    three"
	ee := extentsFromLine(line)

	t.Run("nothing missing", func(t *testing.T) {
		fields := fieldsFromExtents(line, ee)
    	if got, want := len(fields), 3; got != want {
    		t.Fatalf("GOT: %v; WANT: %v", got, want)
    	}
    	// first
    	if got, want := fields[0], "one"; got != want {
    		t.Errorf("GOT: %#v; WANT: %#v", got, want)
    	}
    	// second
    	if got, want := fields[1], "two"; got != want {
    		t.Errorf("GOT: %#v; WANT: %#v", got, want)
    	}
    	// third
    	if got, want := fields[2], "three"; got != want {
    		t.Errorf("GOT: %#v; WANT: %#v", got, want)
    	}
   	})

	t.Run("column missing", func(t *testing.T) {
		fields := fieldsFromExtents("one           three", ee)
    	if got, want := len(fields), 3; got != want {
    		t.Fatalf("GOT: %v; WANT: %v", got, want)
    	}
    	// first
    	if got, want := fields[0], "one"; got != want {
    		t.Errorf("GOT: %v; WANT: %v", got, want)
    	}
    	// second
    	if got, want := fields[1], ""; got != want {
    		t.Errorf("GOT: %v; WANT: %v", got, want)
    	}
    	// third
    	if got, want := fields[2], "three"; got != want {
    		t.Errorf("GOT: %v; WANT: %v", got, want)
    	}
	})
}
