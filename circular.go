package main

import "fmt"

type circularBuffer struct {
	items []interface{}
	i     int
}

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

func (cb *circularBuffer) QueueDequeue(item interface{}) interface{} {
	// Special case when the circular buffer has no capacity: just
	// return item.
	if cb.items == nil {
		return item
	}
	t := cb.items[cb.i]               // Save item at index to be returned later.
	cb.items[cb.i] = item             // Store new item at index.
	cb.i = (cb.i + 1) % cap(cb.items) // Advance index with capacity wrap-around.
	return t
}
