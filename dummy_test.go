package main

import (
	"io"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/karrick/gobls"
	"github.com/karrick/goutfs"
)

type tuple struct {
	start, end int
}

func foo(ior io.Reader) ([][]string, error) {
	nonSpaces := make(map[int]struct{})
	var lines []*goutfs.String
	var maxColumn int

	scanner := gobls.NewScanner(ior)
	for scanner.Scan() {
		line := goutfs.NewString(scanner.Text())

		// Look for non-spaces in this line
		l := line.Len()
		for i := 0; i < l; i++ {
			if maxColumn < i {
				maxColumn = i
			}
			r, _ := utf8.DecodeRune(line.Char(i))
			if !unicode.IsSpace(r) {
				nonSpaces[i] = struct{}{}
			}
		}

		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	_, inField := nonSpaces[0]
	var columns []tuple
	var start int

	for i := 0; i <= maxColumn; i++ {
		if _, ok := nonSpaces[i]; ok {
			if !inField {
				start = i
				inField = true
			}
		} else if inField {
			columns = append(columns, tuple{start, i})
			inField = false
		}
	}
	if inField {
		columns = append(columns, tuple{start, maxColumn + 1})
	}

	var linesWithFields [][]string
	for _, line := range lines {
		var fields []string
		for _, column := range columns {
			field := string(line.Slice(column.start, column.end))
			// field = strings.TrimSpace(field)
			fields = append(fields, field)
		}
		linesWithFields = append(linesWithFields, fields)
	}

	return linesWithFields, nil
}

func TestFoo(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, err := foo(strings.NewReader(""))
		ensureError(t, err)
	})

	t.Run("empty lines", func(t *testing.T) {
		_, err := foo(strings.NewReader("\n\n\n"))
		ensureError(t, err)
	})

	t.Run("spaces", func(t *testing.T) {
		blob := `BenchmarkLowBpool-8             5   283987573 ns/op   149665651 B/op   10255 allocs/op
BenchmarkMedBpool-8             1  2207387112 ns/op  1197784512 B/op   82816 allocs/op
BenchmarkHighBpool-8            1 15027765198 ns/op  9582583872 B/op  661065 allocs/op
BenchmarkHighBpoolRuthless-8    1 16900663489 ns/op 14508565536 B/op 3278937 allocs/op
BenchmarkLowChanPool-8          5   320440990 ns/op   149652630 B/op   10247 allocs/op
BenchmarkMedChanPool-8          1  2344110330 ns/op  1197240032 B/op   82359 allocs/op
BenchmarkHighChanPool-8         1 15506092351 ns/op  9577960560 B/op  659007 allocs/op
BenchmarkHighChanPoolRuthless-8 1 10045489866 ns/op   139116288 B/op   56196 allocs/op
BenchmarkLowLockPool-8          5   334747027 ns/op   149652832 B/op   10251 allocs/op
BenchmarkMedLockPool-8          1  2629362657 ns/op  1197265984 B/op   82689 allocs/op
BenchmarkHighLockPool-8         1 20958793958 ns/op  9578683520 B/op  671025 allocs/op
BenchmarkHighLockPoolRuthless-8 1 10754887930 ns/op    97871408 B/op   38699 allocs/op
BenchmarkLowNotPool-8           5   357442963 ns/op   209977731 B/op   38921 allocs/op
BenchmarkMedNotPool-8           1  2599181484 ns/op  1679837984 B/op  311614 allocs/op
BenchmarkHighNotPool-8          1 16489452179 ns/op 13438974864 B/op 2496261 allocs/op
BenchmarkHighNotPoolRuthless-8  1 12639344861 ns/op 14498563856 B/op 3254607 allocs/op
BenchmarkLowPreAllocatedPool-8  5   352396686 ns/op   168501171 B/op   10518 allocs/op
BenchmarkMedPreAllocatedPool-8  1  2018839349 ns/op  1359448768 B/op   84443 allocs/op
BenchmarkHighPreAllocatedPool-8 1 17677792159 ns/op 11372325008 B/op  682934 allocs/op
BenchmarkLowSyncPool-8          5   303020333 ns/op   149756040 B/op   10390 allocs/op
BenchmarkMedSyncPool-8          1  2043986948 ns/op  1197755448 B/op   82853 allocs/op
BenchmarkHighSyncPool-8         1 14928328269 ns/op  9579632616 B/op  661306 allocs/op
BenchmarkHighSyncPoolRuthless-8 1  9319163392 ns/op   521904008 B/op  241995 allocs/op

`

		linesWithFields, err := foo(strings.NewReader(blob))
		ensureError(t, err)
		ensureSlicesOfSlicesOfStringsMatch(t, linesWithFields, [][]string{
			[]string{
				"BenchmarkLowBpool-8            ", "5", "  283987573", "ns/op", "  149665651", "B/op", "  10255", "allocs/op",
			},
			[]string{
				"BenchmarkMedBpool-8            ", "1", " 2207387112", "ns/op", " 1197784512", "B/op", "  82816", "allocs/op",
			},
			[]string{
				"BenchmarkHighBpool-8           ", "1", "15027765198", "ns/op", " 9582583872", "B/op", " 661065", "allocs/op",
			},
			[]string{
				"BenchmarkHighBpoolRuthless-8   ", "1", "16900663489", "ns/op", "14508565536", "B/op", "3278937", "allocs/op",
			},
			[]string{
				"BenchmarkLowChanPool-8         ", "5", "  320440990", "ns/op", "  149652630", "B/op", "  10247", "allocs/op",
			},
			[]string{
				"BenchmarkMedChanPool-8         ", "1", " 2344110330", "ns/op", " 1197240032", "B/op", "  82359", "allocs/op",
			},
			[]string{
				"BenchmarkHighChanPool-8        ", "1", "15506092351", "ns/op", " 9577960560", "B/op", " 659007", "allocs/op",
			},
			[]string{
				"BenchmarkHighChanPoolRuthless-8", "1", "10045489866", "ns/op", "  139116288", "B/op", "  56196", "allocs/op",
			},
			[]string{
				"BenchmarkLowLockPool-8         ", "5", "  334747027", "ns/op", "  149652832", "B/op", "  10251", "allocs/op",
			},
			[]string{
				"BenchmarkMedLockPool-8         ", "1", " 2629362657", "ns/op", " 197265984", "B/op ", "   82689", "allocs/op",
			},
			[]string{
				"BenchmarkHighLockPool-8        ", "1", " 20958793958", "ns/op", "9578683520", "B/op ", "  671025", "allocs/op",
			},
			[]string{
				"BenchmarkHighLockPoolRuthless-8", "1", " 10754887930", "ns/op", "  97871408", "B/op   ", "   38699", "allocs/op",
			},
			[]string{
				"BenchmarkLowNotPool-8          ", "5", "   357442963", "ns/op ", "209977731", "B/op  ", "   38921", "allocs/op",
			},
			[]string{
				"BenchmarkMedNotPool-8          ", "1", "  2599181484", "ns/op ", "1679837984", "B/op ", "  311614", "allocs/op",
			},
			[]string{
				"BenchmarkHighNotPool-8         ", "1", " 16489452179", "ns/op", "13438974864", "B/op", " 2496261", "allocs/op",
			},
			[]string{
				"BenchmarkHighNotPoolRuthless-8 ", "1", " 12639344861", "ns/op", "14498563856", "B/op", " 3254607", "allocs/op",
			},
			[]string{
				"BenchmarkLowPreAllocatedPool-8 ", "5", "   352396686", "ns/op ", " 168501171", "B/op  ", "   10518", "allocs/op",
			},
			[]string{
				"BenchmarkMedPreAllocatedPool-8 ", "1", "  2018839349", "ns/op ", "1359448768", "B/op ", "   84443", "allocs/op",
			},
			[]string{
				"BenchmarkHighPreAllocatedPool-8", "1", " 17677792159", "ns/op", "11372325008", "B/op", "  682934", "allocs/op",
			},
			[]string{
				"BenchmarkLowSyncPool-8         ", "5", "   303020333", "ns/op ", " 149756040", "B/op  ", "   10390", "allocs/op",
			},
			[]string{
				"BenchmarkMedSyncPool-8         ", "1", "  2043986948", "ns/op ", "1197755448", "B/op ", "   82853", "allocs/op",
			},
			[]string{
				"BenchmarkHighSyncPool-8        ", "1", " 14928328269", "ns/op", " 9579632616", "B/op ", "  661306", "allocs/op",
			},
			[]string{
				"BenchmarkHighSyncPoolRuthless-8", "1", "  9319163392", "ns/op ", " 521904008", "B/op  ", "  241995", "allocs/op",
			},
		})
		// for _, fields := range linesWithFields {
		// 	t.Logf("FIELDS: %#v\n", fields)
		// }
	})

	t.Run("missing-cell", func(t *testing.T) {
		t.Skip()
		blob := `BenchmarkLowBpool-8                       	       5	 283987573 ns/op	149665651 B/op	   10255 allocs/op
BenchmarkMedBpool-8                       	       1	2207387112 ns/op	1197784512 B/op	   82816 allocs/op
BenchmarkHighBpool-8                      	       1	15027765198 ns/op	9582583872 B/op	  661065 allocs/op
BenchmarkHighBpoolRuthless-8              	       1	16900663489 ns/op	14508565536 B/op	 3278937 allocs/op
BenchmarkLowChanPool-8                    	       5	 320440990 ns/op	149652630 B/op	   10247 allocs/op
BenchmarkMedChanPool-8                    	       1	2344110330 ns/op	1197240032 B/op	   82359 allocs/op
BenchmarkHighChanPool-8                   	       1	15506092351 ns/op	9577960560 B/op	  659007 allocs/op
BenchmarkHighChanPoolRuthless-8           	       1	10045489866 ns/op	139116288 B/op	   56196 allocs/op
BenchmarkLowLockPool-8                    	       5	 334747027 ns/op	149652832 B/op	   10251 allocs/op
BenchmarkMedLockPool-8                    	       1	2629362657 ns/op	1197265984 B/op	   82689 allocs/op
BenchmarkHighLockPool-8                   	       1	20958793958 ns/op	9578683520 B/op	  671025 allocs/op
BenchmarkHighLockPoolRuthless-8           	       1	10754887930 ns/op	97871408 B/op	   38699 allocs/op
BenchmarkLowNotPool-8                     	       5	                 	357442963 ns/op	209977731 B/op	   38921 allocs/op
BenchmarkMedNotPool-8                     	       1	2599181484 ns/op	1679837984 B/op	  311614 allocs/op
BenchmarkHighNotPool-8                    	       1	16489452179 ns/op	13438974864 B/op	 2496261 allocs/op
BenchmarkHighNotPoolRuthless-8            	       1	12639344861 ns/op	14498563856 B/op	 3254607 allocs/op
BenchmarkLowPreAllocatedPool-8            	       5	 352396686 ns/op	168501171 B/op	   10518 allocs/op
BenchmarkMedPreAllocatedPool-8            	       1	2018839349 ns/op	1359448768 B/op	   84443 allocs/op
BenchmarkHighPreAllocatedPool-8           	       1	17677792159 ns/op	11372325008 B/op	  682934 allocs/op
BenchmarkLowSyncPool-8                    	       5	 303020333 ns/op	149756040 B/op	   10390 allocs/op
BenchmarkMedSyncPool-8                    	       1	2043986948 ns/op	1197755448 B/op	   82853 allocs/op
BenchmarkHighSyncPool-8                   	       1	14928328269 ns/op	9579632616 B/op	  661306 allocs/op
BenchmarkHighSyncPoolRuthless-8           	       1	9319163392 ns/op	521904008 B/op	  241995 allocs/op
`

		linesWithFields, err := foo(strings.NewReader(blob))
		ensureError(t, err)
		for _, fields := range linesWithFields {
			t.Logf("FIELDS: %#v\n", fields)
		}
	})
}

// func TestDummyWithTabs(t *testing.T) {
// 	t.Skip()
// 	blob := `
// BenchmarkLowBpool-8                       	       5	 283987573 ns/op	149665651 B/op	   10255 allocs/op
// BenchmarkMedBpool-8                       	       1	2207387112 ns/op	1197784512 B/op	   82816 allocs/op
// BenchmarkHighBpool-8                      	       1	15027765198 ns/op	9582583872 B/op	  661065 allocs/op
// BenchmarkHighBpoolRuthless-8              	       1	16900663489 ns/op	14508565536 B/op	 3278937 allocs/op
// BenchmarkLowChanPool-8                    	       5	 320440990 ns/op	149652630 B/op	   10247 allocs/op
// BenchmarkMedChanPool-8                    	       1	2344110330 ns/op	1197240032 B/op	   82359 allocs/op
// BenchmarkHighChanPool-8                   	       1	15506092351 ns/op	9577960560 B/op	  659007 allocs/op
// BenchmarkHighChanPoolRuthless-8           	       1	10045489866 ns/op	139116288 B/op	   56196 allocs/op
// BenchmarkLowLockPool-8                    	       5	 334747027 ns/op	149652832 B/op	   10251 allocs/op
// BenchmarkMedLockPool-8                    	       1	2629362657 ns/op	1197265984 B/op	   82689 allocs/op
// BenchmarkHighLockPool-8                   	       1	20958793958 ns/op	9578683520 B/op	  671025 allocs/op
// BenchmarkHighLockPoolRuthless-8           	       1	10754887930 ns/op	97871408 B/op	   38699 allocs/op
// BenchmarkLowNotPool-8                     	       5	 357442963 ns/op	209977731 B/op	   38921 allocs/op
// BenchmarkMedNotPool-8                     	       1	2599181484 ns/op	1679837984 B/op	  311614 allocs/op
// BenchmarkHighNotPool-8                    	       1	16489452179 ns/op	13438974864 B/op	 2496261 allocs/op
// BenchmarkHighNotPoolRuthless-8            	       1	12639344861 ns/op	14498563856 B/op	 3254607 allocs/op
// BenchmarkLowPreAllocatedPool-8            	       5	 352396686 ns/op	168501171 B/op	   10518 allocs/op
// BenchmarkMedPreAllocatedPool-8            	       1	2018839349 ns/op	1359448768 B/op	   84443 allocs/op
// BenchmarkHighPreAllocatedPool-8           	       1	17677792159 ns/op	11372325008 B/op	  682934 allocs/op
// BenchmarkLowSyncPool-8                    	       5	 303020333 ns/op	149756040 B/op	   10390 allocs/op
// BenchmarkMedSyncPool-8                    	       1	2043986948 ns/op	1197755448 B/op	   82853 allocs/op
// BenchmarkHighSyncPool-8                   	       1	14928328269 ns/op	9579632616 B/op	  661306 allocs/op
// BenchmarkHighSyncPoolRuthless-8           	       1	9319163392 ns/op	521904008 B/op	  241995 allocs/op
// `
// 	ensureError(t, foo(strings.NewReader(blob)))
// }
