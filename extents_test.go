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
	if got, want := ee[0].lc, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].rc, 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsSingleItemWithWhiteSpace(t *testing.T) {
	ee := extentsFromLine(" item ")
	if got, want := len(ee), 1; got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].lc, 2; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].rc, 5; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsTwoItemsWithoutWhiteSpace(t *testing.T) {
	ee := extentsFromLine("one two")
	if got, want := len(ee), 2; got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	// one
	if got, want := ee[0].lc, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].rc, 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// two
	if got, want := ee[1].lc, 5; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[1].rc, 7; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsTwoItemsWithWhitespace(t *testing.T) {
	ee := extentsFromLine(" one two ")
	if got, want := len(ee), 2; got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	// one
	if got, want := ee[0].lc, 2; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[0].rc, 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// two
	if got, want := ee[1].lc, 6; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee[1].rc, 8; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

////////////////////////////////////////

func TestExtentsMergeWithoutOverlap(t *testing.T) {
	ee, ok := attemptMerge(extent{lc: 11, rc: 15}, extent{lc: 6, rc: 8})

	if got, want := ok, false; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.lc, 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.rc, 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestExtentsMergeWithOverlap(t *testing.T) {
	ee, ok := attemptMerge(extent{lc: 1, rc: 3}, extent{lc: 2, rc: 3})

	if got, want := ok, true; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.lc, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee.rc, 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

////////////////////////////////////////

func TestMergeExtentSlices(t *testing.T) {
	ee1 := []extent{extent{lc: 1, rc: 3}, extent{lc: 11, rc: 15}}
	ee2 := []extent{extent{lc: 2, rc: 5}, extent{lc: 6, rc: 8}, extent{lc: 10, rc: 12}}

	ee3 := mergeExtents(ee1, ee2)

	if got, want := len(ee3), 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// first
	if got, want := ee3[0].lc, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[0].rc, 5; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// second
	if got, want := ee3[1].lc, 6; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[1].rc, 8; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// third
	if got, want := ee3[2].lc, 10; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[2].rc, 15; got != want {
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
	if got, want := ee3[0].lc, 1; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[0].rc, 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[0].width(), 4; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// second
	if got, want := ee3[1].lc, 8; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[1].rc, 10; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[1].width(), 3; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// third
	if got, want := ee3[2].lc, 15; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := ee3[2].rc, 21; got != want {
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
