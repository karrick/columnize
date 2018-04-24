package main

import "fmt"

// tailBuffer is a non-concurrency safe data structure for storing the N
// previous items. While concurrency safety could be added, the overhead is not
// warranted as this particular program has no need of it.
type tailBuffer struct {
	i      int
	items  []interface{}
	looped bool
}

// newTailBuffer returns a newly initialized tailBuffer.
func newTailBuffer(n int) (*tailBuffer, error) {
	switch {
	case n < 0:
		return nil, fmt.Errorf("cannot create tailBuffer with negative item count: %d", n)
	case n == 0:
		return new(tailBuffer), nil
	default:
		return &tailBuffer{items: make([]interface{}, n)}, nil
	}
}

// QueueDequeue returns the Nth item back from the head of the queue, storing
// the newly specified item in its place. N was specified at the time the
// tailBuffer was created.
func (cb *tailBuffer) QueueDequeue(item interface{}) interface{} {
	// Special case when buffer has no capacity: just return item.
	if cb.items == nil {
		return item
	}
	tmp := cb.items[cb.i] // Save item at index to be returned later.
	cb.items[cb.i] = item // Store new item at index.

	// Increment index making note when we wrap-around.
	if cb.i++; cb.i == cap(cb.items) {
		cb.i = 0
		cb.looped = true
	}

	return tmp
}

// Drain returns remaining items. This implimentation is kept simple and
// therefore not designed to properly handle invocation of any other methods
// after calling Drain.
func (cb *tailBuffer) Drain() []interface{} {
	if cb.looped {
		return append(cb.items[cb.i:], cb.items[:cb.i]...) // f g c d e
	}
	return cb.items[:cb.i] // a b c _ _
}
