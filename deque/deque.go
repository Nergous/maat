// Package deque provides a generic, slice-backed double-ended queue.
//
// Create a deque with [New] or [NewWithCap]. Elements can be added to either
// end with [Deque.PushFront], [Deque.PushBack], [Deque.PushFrontN], or
// [Deque.PushBackN], and removed from either end with [Deque.PopFront],
// [Deque.PopBack], [Deque.PopFrontN], or [Deque.PopBackN]. [Deque.Front] and
// [Deque.Back] inspect the ends without removing elements.
//
// Single-element pushes and pops run in amortized O(1) time. Front, Back, Len,
// Cap, and IsEmpty run in O(1). Methods that copy or iterate over every live
// element, such as All, Clone, Slice, Shrink, and Clip, run in O(n).
//
// The zero value of Deque[T] is usable as an empty deque. A nil *Deque[T]
// behaves as an empty deque for every read-only method and for empty-removal
// methods such as [Deque.PopFront], [Deque.PopBack], [Deque.PopFrontN], and
// [Deque.PopBackN], so the zero pointer value is safe to query without first
// calling [New]. Methods that need to store, clear, reset, shrink, or resize
// data require a non-nil receiver and panic on a nil one; see each method's
// documentation.
//
// A Deque is not safe for concurrent use: it performs no internal locking.
// Callers that share a deque across goroutines must provide their own
// synchronization.
package deque

import "iter"

// Deque is a generic double-ended queue backed by a slice ring buffer. The
// front is the first element removed by [Deque.PopFront]; the back is the first
// element removed by [Deque.PopBack]. A Deque can be used as a FIFO queue, a
// LIFO stack, or a mixed-priority queue where either end is significant.
//
// The zero value is a valid empty deque. A nil *Deque[T] is treated as a valid
// empty deque by read-only and empty-removal methods, but methods that store or
// resize data require a non-nil receiver.
//
// Deque is not safe for concurrent use.
type Deque[T any] struct {
	data []T
	head int
	len  int
}

// New returns an empty deque with no preallocated capacity.
func New[T any]() *Deque[T] {
	return &Deque[T]{}
}

// NewWithCap returns an empty deque with a backing array preallocated for at
// least n elements. Use it when the maximum size is known up front to avoid
// reallocations while pushing. A negative n is clamped to zero.
func NewWithCap[T any](n int) *Deque[T] {
	if n < 0 {
		n = 0
	}
	return &Deque[T]{data: make([]T, n)}
}

// Len returns the number of elements currently in the deque. A nil receiver
// reports 0.
func (d *Deque[T]) Len() int {
	if d == nil {
		return 0
	}
	return d.len
}

// Cap returns the capacity of the deque's backing ring buffer, i.e. the total
// number of elements it can hold before the next reallocation. Popping elements
// does not reduce Cap; freed front and back slots are reused by later pushes. A
// nil receiver reports 0.
func (d *Deque[T]) Cap() int {
	if d == nil {
		return 0
	}
	return len(d.data)
}

// IsEmpty reports whether the deque has no elements. A nil receiver reports
// true.
func (d *Deque[T]) IsEmpty() bool {
	return d == nil || d.len == 0
}

// Front returns the front element without removing it. The boolean result is
// false when the deque is empty, in which case the returned value is the zero
// value of T. A nil receiver is treated as empty and returns (zero, false).
func (d *Deque[T]) Front() (T, bool) {
	var zero T
	if d == nil || d.len == 0 {
		return zero, false
	}
	return d.data[d.head], true
}

// Back returns the back element without removing it. The boolean result is
// false when the deque is empty, in which case the returned value is the zero
// value of T. A nil receiver is treated as empty and returns (zero, false).
func (d *Deque[T]) Back() (T, bool) {
	var zero T
	if d == nil || d.len == 0 {
		return zero, false
	}
	return d.data[d.index(d.len-1)], true
}

// All returns an iterator over the deque's elements from front to back, without
// consuming them; the deque is left unchanged. The returned [iter.Seq] is a Go
// 1.23 range-over-func iterator, so it can be used directly in a range loop.
// Breaking out of the loop early stops iteration cleanly. A nil receiver is
// treated as empty and yields no elements.
func (d *Deque[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		if d == nil || d.len == 0 {
			return
		}
		first, second := d.segments()
		for _, v := range first {
			if !yield(v) {
				return
			}
		}
		for _, v := range second {
			if !yield(v) {
				return
			}
		}
	}
}

// Clone returns an independent shallow copy of the deque. The two deques share
// no backing array, so mutating one does not affect the other. Elements are
// copied as-is: if T is a pointer or contains references, the pointed-to data is
// shared between the copies. Runs in O(n).
//
// A nil receiver, treated everywhere as an empty deque, clones to a new, usable
// empty deque (never nil), so the result can be pushed to without a further
// [New] call.
func (d *Deque[T]) Clone() *Deque[T] {
	if d == nil {
		return New[T]()
	}

	c := &Deque[T]{
		data: make([]T, d.len),
		len:  d.len,
	}
	d.copyLiveTo(c.data)
	return c
}

// Slice returns a copy of the deque's elements from front to back, the same
// order [Deque.All] yields. The returned slice is independent of the deque:
// they share no backing array, so callers may modify it freely without
// affecting the deque, and vice versa. An empty deque returns nil; a nil
// receiver is treated as empty and likewise returns nil. Runs in O(n).
func (d *Deque[T]) Slice() []T {
	if d == nil || d.len == 0 {
		return nil
	}

	out := make([]T, d.len)
	d.copyLiveTo(out)
	return out
}

// PushFront adds v to the front of the deque. It runs in amortized O(1) time;
// the backing array may be reallocated to grow.
//
// PushFront requires a non-nil receiver and panics on a nil one: it must store
// the element somewhere. Create the deque with [New] or [NewWithCap] first.
func (d *Deque[T]) PushFront(v T) {
	d.PushFrontN(v)
}

// PushBack adds v to the back of the deque. It runs in amortized O(1) time; the
// backing array may be reallocated to grow.
//
// PushBack requires a non-nil receiver and panics on a nil one: it must store
// the element somewhere. Create the deque with [New] or [NewWithCap] first.
func (d *Deque[T]) PushBack(v T) {
	d.PushBackN(v)
}

// PushFrontN adds vs to the front of the deque in argument order, so the first
// argument becomes the new front. It grows the backing array at most once,
// making it cheaper than repeated [Deque.PushFront] calls. With no arguments it
// is a no-op. Runs in amortized O(k) for k elements.
//
// PushFrontN requires a non-nil receiver and panics on a nil one: it must store
// the elements somewhere. Create the deque with [New] or [NewWithCap] first.
func (d *Deque[T]) PushFrontN(vs ...T) {
	if d == nil {
		panic("deque: PushFrontN on nil receiver")
	}
	if len(vs) == 0 {
		return
	}

	d.growFor(d.len + len(vs))
	capacity := len(d.data)
	d.head = (d.head - len(vs) + capacity) % capacity
	for i, v := range vs {
		d.data[(d.head+i)%capacity] = v
	}
	d.len += len(vs)
}

// PushBackN adds vs to the back of the deque in argument order, so the last
// argument becomes the new back. It grows the backing array at most once,
// making it cheaper than repeated [Deque.PushBack] calls. With no arguments it
// is a no-op. Runs in amortized O(k) for k elements.
//
// PushBackN requires a non-nil receiver and panics on a nil one: it must store
// the elements somewhere. Create the deque with [New] or [NewWithCap] first.
func (d *Deque[T]) PushBackN(vs ...T) {
	if d == nil {
		panic("deque: PushBackN on nil receiver")
	}
	if len(vs) == 0 {
		return
	}

	d.growFor(d.len + len(vs))
	for i, v := range vs {
		d.data[d.index(d.len+i)] = v
	}
	d.len += len(vs)
}

// PopFront removes and returns the front element. The boolean result is false
// when the deque is empty, in which case the returned value is the zero value of
// T. The vacated slot is zeroed so it no longer retains a reference to the
// popped element, then the front is advanced. It runs in O(1) time.
//
// A nil receiver is treated as empty and returns (zero, false).
func (d *Deque[T]) PopFront() (T, bool) {
	var zero T
	if d == nil || d.len == 0 {
		return zero, false
	}

	v := d.data[d.head]
	d.data[d.head] = zero
	d.head = (d.head + 1) % len(d.data)
	d.len--
	if d.len == 0 {
		d.head = 0
	}
	return v, true
}

// PopBack removes and returns the back element. The boolean result is false
// when the deque is empty, in which case the returned value is the zero value of
// T. The vacated slot is zeroed so it no longer retains a reference to the
// popped element. It runs in O(1) time.
//
// A nil receiver is treated as empty and returns (zero, false).
func (d *Deque[T]) PopBack() (T, bool) {
	var zero T
	if d == nil || d.len == 0 {
		return zero, false
	}

	idx := d.index(d.len - 1)
	v := d.data[idx]
	d.data[idx] = zero
	d.len--
	if d.len == 0 {
		d.head = 0
	}
	return v, true
}

// PopFrontN removes and returns up to n front elements in front-to-back order.
// If n is greater than Len, it drains the deque. Non-positive n, nil receivers,
// and empty deques return nil. The deque's capacity is preserved. Runs in O(k),
// where k is the number of elements returned.
func (d *Deque[T]) PopFrontN(n int) []T {
	if d == nil || d.len == 0 || n <= 0 {
		return nil
	}
	if n > d.len {
		n = d.len
	}

	out := make([]T, n)
	for i := range n {
		v, _ := d.PopFront()
		out[i] = v
	}
	return out
}

// PopBackN removes and returns up to n back elements in removal order
// (back-to-front). If n is greater than Len, it drains the deque.
// Non-positive n, nil receivers, and empty deques return nil. The deque's
// capacity is preserved. Runs in O(k), where k is the number of elements
// returned.
func (d *Deque[T]) PopBackN(n int) []T {
	if d == nil || d.len == 0 || n <= 0 {
		return nil
	}
	if n > d.len {
		n = d.len
	}

	out := make([]T, n)
	for i := range n {
		v, _ := d.PopBack()
		out[i] = v
	}
	return out
}

// Shrink reduces the backing array to hold exactly Len elements, copying them
// into a new, right-sized array so the previous (larger) array becomes eligible
// for garbage collection immediately. Use it to reclaim memory after a deque
// that grew large has shrunk and will not grow back.
//
// Shrink runs in O(n) and allocates once; it is a no-op when the capacity
// already equals the length and live elements start at the front of the backing
// buffer. For the cheaper variant that can avoid copying when live elements are
// contiguous, use [Deque.Clip]. Shrink is the counterpart of [Deque.Grow].
//
// Shrink requires a non-nil receiver and panics on a nil one. Create the deque
// with [New] or [NewWithCap] first.
func (d *Deque[T]) Shrink() {
	if d.len == 0 {
		d.data = nil
		d.head = 0
		return
	}
	if len(d.data) == d.len && d.head == 0 {
		return
	}

	trimmed := make([]T, d.len)
	d.copyLiveTo(trimmed)
	d.data = trimmed
	d.head = 0
}

// Clip reduces the reported capacity to the current length, copying only when
// live elements wrap around the end of the backing buffer. Because a
// non-copying reslice may still retain the original array, unused memory is not
// guaranteed to be released immediately. For an immediate, always-copying
// reclaim, use [Deque.Shrink].
//
// Clip requires a non-nil receiver and panics on a nil one. Create the deque
// with [New] or [NewWithCap] first.
func (d *Deque[T]) Clip() {
	if d.len == 0 {
		d.data = nil
		d.head = 0
		return
	}
	if len(d.data) == d.len && d.head == 0 {
		return
	}
	if d.head+d.len <= len(d.data) {
		d.data = d.data[d.head : d.head+d.len : d.head+d.len]
		d.head = 0
		return
	}

	clipped := make([]T, d.len)
	d.copyLiveTo(clipped)
	d.data = clipped
	d.head = 0
}

// Grow ensures the deque has spare capacity for at least n more elements
// without reallocating during subsequent pushes, growing the backing array if
// needed. It is the counterpart of [Deque.Shrink]. Grow runs in O(n) when it
// reallocates and is a no-op when n is non-positive or the capacity is already
// sufficient.
//
// Grow requires a non-nil receiver and panics on a nil one. Create the deque
// with [New] or [NewWithCap] first.
func (d *Deque[T]) Grow(n int) {
	if d == nil {
		panic("deque: Grow on nil receiver")
	}
	if n <= 0 {
		return
	}
	d.growFor(d.len + n)
}

// Clear removes all elements and releases the backing array, so its memory
// becomes eligible for garbage collection. After Clear, Cap reports 0. To empty
// the deque while keeping its capacity for reuse, use [Deque.Reset] instead.
//
// Clear requires a non-nil receiver and panics on a nil one. Create the deque
// with [New] or [NewWithCap] first.
func (d *Deque[T]) Clear() {
	d.data = nil
	d.head = 0
	d.len = 0
}

// Reset removes all elements but keeps the backing array for reuse, so Cap is
// preserved. The elements are zeroed so the array no longer retains references
// to them. To also release the backing array, use [Deque.Clear] instead.
//
// Reset requires a non-nil receiver and panics on a nil one. Create the deque
// with [New] or [NewWithCap] first.
func (d *Deque[T]) Reset() {
	d.clearLive()
	d.head = 0
	d.len = 0
}

func (d *Deque[T]) index(offset int) int {
	return (d.head + offset) % len(d.data)
}

func (d *Deque[T]) segments() ([]T, []T) {
	if d.len == 0 {
		return nil, nil
	}
	end := d.head + d.len
	if end <= len(d.data) {
		return d.data[d.head:end], nil
	}
	return d.data[d.head:], d.data[:end%len(d.data)]
}

func (d *Deque[T]) copyLiveTo(dst []T) {
	first, second := d.segments()
	n := copy(dst, first)
	copy(dst[n:], second)
}

func (d *Deque[T]) clearLive() {
	first, second := d.segments()
	clear(first)
	clear(second)
}

func (d *Deque[T]) growFor(need int) {
	if need <= len(d.data) {
		return
	}

	grown := make([]T, nextCap(len(d.data), need))
	d.copyLiveTo(grown)
	d.data = grown
	d.head = 0
}

func nextCap(current, need int) int {
	if current <= 0 {
		current = 1
	}
	for current < need {
		if current < 1024 {
			current *= 2
			continue
		}
		current += current / 4
	}
	return current
}
