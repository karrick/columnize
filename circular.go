package main

import "fmt"

// circularBuffer is a non-concurrency safe data structure for storing the N
// previous items.
type circularBuffer struct {
	items  []interface{}
	i      int
	looped bool
}

// newCircularBuffer returns a newly initialized circularBuffer.
func newCircularBuffer(n int) (*circularBuffer, error) {
	switch {
	case n < 0:
		return nil, fmt.Errorf("cannot create round buffer with negative item count: %d", n)
	case n == 0:
		return new(circularBuffer), nil
	default:
		return &circularBuffer{items: make([]interface{}, n)}, nil
	}
}

// QeuueDequeue returns the Nth item back from the head of the queue, storing
// the newly specified item in its place.
func (cb *circularBuffer) QueueDequeue(item interface{}) interface{} {
	// Special case when the circular buffer has no capacity: just
	// return item.
	if cb.items == nil {
		return item
	}
	t := cb.items[cb.i]   // Save item at index to be returned later.
	cb.items[cb.i] = item // Store new item at index.

	// increment index making note if we wrap-around
	if cb.i++; cb.i == cap(cb.items) {
		cb.i = 0
		cb.looped = true
	}

	return t
}

// Drain returns remaining items. This implimentation is not designed to
// properly handle invocation of any other methods after calling Drain.
func (cb *circularBuffer) Drain() []interface{} {
	if cb.looped {
		return append(cb.items[cb.i:], cb.items[:cb.i]...) // f g c d e
	}
	return cb.items[:cb.i] // a b c _ _
}
