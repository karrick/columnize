package main

import (
	"strings"
	"testing"
)

func ensureError(tb testing.TB, err error, contains ...string) {
	tb.Helper()
	if len(contains) == 0 || (len(contains) == 1 && contains[0] == "") {
		if err != nil {
			tb.Fatalf("GOT: %v; WANT: %v", err, contains)
		}
	} else if err == nil {
		tb.Errorf("GOT: %v; WANT: %v", err, contains)
	} else {
		for _, stub := range contains {
			if stub != "" && !strings.Contains(err.Error(), stub) {
				tb.Errorf("GOT: %v; WANT: %q", err, stub)
			}
		}
	}
}

func ensureSlicesOfStringsMatch(tb testing.TB, got, want []string) {
	tb.Helper()

	la, lb := len(got), len(want)

	max := la
	if max < lb {
		max = lb
	}

	for i := 0; i < max; i++ {
		if i < la && i < lb {
			if got, want := got[i], want[i]; got != want {
				tb.Errorf("%d: GOT: %q; WANT: %q", i, got, want)
			}
		} else if i < la {
			tb.Errorf("%d: GOT: extra slice: %v", i, got[i])
		} else /* i < lb */ {
			tb.Errorf("%d: WANT: extra slice: %v", i, want[i])
		}
	}
	// if tb.Failed() {
	// 	tb.Logf("GOT: %v; WANT: %v", got, want)
	// }
}

func ensureSlicesOfSlicesOfStringsMatch(tb testing.TB, got, want [][]string) {
	tb.Helper()

	la, lb := len(got), len(want)

	max := la
	if max < lb {
		max = lb
	}

	for i := 0; i < max; i++ {
		if i < la && i < lb {
			ensureSlicesOfStringsMatch(tb, got[i], want[i])
		} else if i < la {
			tb.Errorf("%d: GOT: extra slice: %v", i, got[i])
		} else /* i < lb */ {
			tb.Errorf("%d: WANT: extra slice: %v", i, want[i])
		}
	}
	// if tb.Failed() {
	// 	tb.Logf("GOT: %v; WANT: %v", got, want)
	// }
}
