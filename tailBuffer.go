package main

// tailBuffer is a non-concurrency safe data structure for storing the N
// previous items, where 0 <= N <= limit.
type tailBuffer struct {
	items  []interface{}
	index  int
	looped bool
}

// newTailBuffer returns a newly initialized tailBuffer..
func newTailBuffer(n uint64) (*tailBuffer, error) {
	switch {
	case n == 0:
		return new(tailBuffer), nil
	default:
		return &tailBuffer{items: make([]interface{}, n)}, nil
	}
}

// QeuueDequeue returns the Nth item back from the head of the queue, storing
// the newly specified item in its place.
func (tb *tailBuffer) QueueDequeue(newItem interface{}) interface{} {
	// Special case when the circular buffer has no capacity: just
	// return item.
	if tb.items == nil {
		return newItem
	}

	// Swap item previously stored at index with new item.
	prevItem := tb.items[tb.index]
	tb.items[tb.index] = newItem

	// Increment index making note whether it wraps.
	if tb.index++; tb.index == cap(tb.items) {
		tb.index = 0
		tb.looped = true
	}

	return prevItem
}

// Drain returns all items from the structure. This implimentation is not
// designed to handle invocation of any other methods after calling Drain.
func (tb *tailBuffer) Drain() []interface{} {
	if tb.looped {
		return append(tb.items[tb.index:], tb.items[:tb.index]...) // f g c d e
	}
	return tb.items[:tb.index] // a b c _ _
}
