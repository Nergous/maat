// Package queue provides a generic, slice-backed FIFO (first-in, first-out) queue.
//
// Create a queue with [New] or [NewWithCap]. Elements are added to the back with
// [Queue.Push] or [Queue.PushN] and removed from the front with [Queue.Pop] or
// [Queue.PopN]; [Queue.Peek] inspects the front without removing it.
//
// Push and Pop run in amortized O(1) time (Push may occasionally reallocate
// the backing array as it grows). Peek, Len, Cap and IsEmpty run in O(1).
//
// A nil *Queue[T] behaves as an empty queue for every read-only method and for
// empty-removal methods such as [Queue.Pop] and [Queue.PopN], so the zero
// pointer value is safe to query without first calling [New]. Methods that need
// to store or resize data ([Queue.Push], [Queue.PushN], [Queue.Grow],
// [Queue.Clear], [Queue.Reset], [Queue.Shrink], [Queue.Clip]) require a non-nil
// receiver and panic on a nil one; see each method's documentation.
//
// A Queue is not safe for concurrent use: it performs no internal locking.
// Callers that share a queue across goroutines must provide their own
// synchronization.
package queue

import (
	"iter"
)

// Queue is a generic FIFO queue backed by a slice ring buffer. The front of the
// queue is the oldest element still present, i.e. the next one to be removed. A
// Queue must be created with [New] or [NewWithCap] before methods that store or
// resize data are called; a nil *Queue[T] is treated as a valid empty queue by
// read-only and empty-removal methods.
//
// Queue is not safe for concurrent use.
type Queue[T any] struct {
	data []T
	head int
	len  int
}

// New returns an empty queue with no preallocated capacity.
func New[T any]() *Queue[T] {
	return &Queue[T]{}
}

// NewWithCap returns an empty queue with a backing array preallocated for at
// least n elements. Use it when the maximum size is known up front to avoid
// reallocations while pushing.
func NewWithCap[T any](n int) *Queue[T] {
	if n < 0 {
		n = 0
	}
	return &Queue[T]{data: make([]T, n)}
}

// Len returns the number of elements currently in the queue. A nil receiver
// reports 0.
func (q *Queue[T]) Len() int {
	if q == nil {
		return 0
	}
	return q.len
}

// Cap returns the capacity of the queue's backing buffer, i.e. the total number
// of elements it can hold before the next reallocation. Popping elements does
// not reduce Cap; freed front slots are reused by later pushes. A nil receiver
// reports 0.
func (q *Queue[T]) Cap() int {
	if q == nil {
		return 0
	}
	return len(q.data)
}

// IsEmpty reports whether the queue has no elements. A nil receiver reports
// true.
func (q *Queue[T]) IsEmpty() bool {
	return q == nil || q.len == 0
}

// Peek returns the front element (the oldest, next to be removed) without
// removing it. The boolean result is false when the queue is empty, in which
// case the returned value is the zero value of T. A nil receiver is treated as
// empty and returns (zero, false).
func (q *Queue[T]) Peek() (T, bool) {
	if q == nil || q.len == 0 {
		var zero T
		return zero, false
	}
	return q.data[q.head], true
}

// All returns an iterator over the queue's elements from front to back (FIFO
// order), without consuming them; the queue is left unchanged. The returned
// [iter.Seq] is a Go 1.23 range-over-func iterator, so it can be used directly
// in a range loop. Breaking out of the loop early stops iteration cleanly. A nil
// receiver is treated as empty and yields no elements.
func (q *Queue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		if q == nil || q.len == 0 {
			return
		}
		first, second := q.segments()
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

// Clone returns an independent shallow copy of the queue. The two queues share
// no backing array, so mutating one (pushing, popping, clearing) does not affect
// the other. Elements are copied as-is: if T is a pointer or contains
// references, the pointed-to data is shared between the copies. Runs in O(n).
//
// A nil receiver, treated everywhere as an empty queue, clones to a new, usable
// empty queue (never nil), so the result can be pushed to without a further
// [New] call.
func (q *Queue[T]) Clone() *Queue[T] {
	if q == nil {
		return New[T]()
	}

	c := &Queue[T]{
		data: make([]T, q.len),
		len:  q.len,
	}
	q.copyLiveTo(c.data)
	return c
}

// Slice returns a copy of the queue's elements from front to back (FIFO), the
// same order [Queue.All] yields. The returned slice is independent of the queue:
// they share no backing array, so callers may modify it freely without affecting
// the queue, and vice versa. An empty queue returns nil; a nil receiver is
// treated as empty and likewise returns nil. Runs in O(n).
func (q *Queue[T]) Slice() []T {
	if q == nil || q.len == 0 {
		return nil
	}

	out := make([]T, q.len)
	q.copyLiveTo(out)
	return out
}

// Push adds v to the back of the queue. It runs in amortized O(1) time; the
// backing array may be reallocated to grow.
//
// Push requires a non-nil receiver and panics on a nil one: it must store the
// element somewhere. Create the queue with [New] or [NewWithCap] first.
func (q *Queue[T]) Push(v T) {
	q.PushN(v)
}

// PushN adds vs to the back of the queue in argument order, so the last argument
// ends up at the back. It grows the backing array at most once, making it
// cheaper than repeated [Queue.Push] calls. With no arguments it is a no-op.
// Runs in amortized O(k) for k elements.
//
// PushN requires a non-nil receiver and panics on a nil one: it must store the
// elements somewhere. Create the queue with [New] or [NewWithCap] first.
func (q *Queue[T]) PushN(vs ...T) {
	if q == nil {
		panic("queue: PushN on nil receiver")
	}
	if len(vs) == 0 {
		return
	}

	q.growFor(q.len + len(vs))
	for _, v := range vs {
		q.data[q.index(q.len)] = v
		q.len++
	}
}

// Pop removes and returns the front element (the oldest, next to be removed).
// The boolean result is false when the queue is empty, in which case the
// returned value is the zero value of T. The vacated slot is zeroed so it no
// longer retains a reference to the popped element, then the front is advanced.
// It runs in O(1) time.
func (q *Queue[T]) Pop() (T, bool) {
	var zero T
	if q == nil || q.len == 0 {
		return zero, false
	}
	v := q.data[q.head]
	q.data[q.head] = zero
	q.head++
	if q.head == len(q.data) {
		q.head = 0
	}
	q.len--
	if q.len == 0 {
		q.head = 0
	}
	return v, true
}

// PopN removes and returns up to n front elements in FIFO order. If n is
// greater than Len, it drains the queue. If n is non-positive, or the queue is
// empty or nil, PopN returns nil. The queue's capacity is preserved. Runs in
// O(k), where k is the number of elements returned.
func (q *Queue[T]) PopN(n int) []T {
	if q == nil || n <= 0 || q.len == 0 {
		return nil
	}
	if n > q.len {
		n = q.len
	}

	out := make([]T, n)
	for i := range n {
		v, _ := q.Pop()
		out[i] = v
	}
	return out
}

// Shrink reduces the backing array to hold exactly Len elements, copying them
// into a new, right-sized array so the previous (larger) array becomes eligible
// for garbage collection immediately. Use it to reclaim memory after a queue
// that grew large has shrunk and will not grow back.
//
// Shrink runs in O(n) and allocates once; it is a no-op when the capacity
// already equals the length. For the cheaper, non-copying variant that only
// reslices and lets memory be reclaimed lazily on the next growth, use
// [Queue.Clip]. Shrink is the counterpart of [Queue.Grow].
//
// Shrink requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Shrink() {
	if q.len == 0 {
		q.data = nil
		q.head = 0
		return
	}
	if len(q.data) == q.len && q.head == 0 {
		return
	}

	trimmed := make([]T, q.len)
	q.copyLiveTo(trimmed)
	q.data = trimmed
	q.head = 0
}

// Clip reduces the reported capacity to the current length. When the live
// elements are already contiguous in the backing buffer, Clip reslices without
// copying; when they wrap around the end of the buffer, it copies them into a
// compact buffer. Because the original array may be retained by a resliced view,
// unused memory is not guaranteed to be released immediately. For an immediate,
// always-copying reclaim, use [Queue.Shrink].
//
// Clip requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Clip() {
	if q.len == 0 {
		q.data = nil
		q.head = 0
		return
	}
	if len(q.data) == q.len && q.head == 0 {
		return
	}
	if q.head+q.len <= len(q.data) {
		q.data = q.data[q.head : q.head+q.len : q.head+q.len]
		q.head = 0
		return
	}

	clipped := make([]T, q.len)
	q.copyLiveTo(clipped)
	q.data = clipped
	q.head = 0
}

// Grow ensures the queue has spare capacity for at least n more elements without
// reallocating during subsequent pushes, growing the backing array if needed. It
// is the counterpart of [Queue.Shrink]. Grow runs in O(n) when it reallocates and
// is a no-op when n is non-positive or the capacity is already sufficient. It
// panics if the new capacity overflows, the same condition as [slices.Grow].
//
// Grow requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Grow(n int) {
	if q == nil {
		panic("queue: Grow on nil receiver")
	}
	if n <= 0 {
		return
	}
	q.growFor(q.len + n)
}

// Clear removes all elements and releases the backing array, so its memory
// becomes eligible for garbage collection. After Clear, Cap reports 0. To empty
// the queue while keeping its capacity for reuse, use [Queue.Reset] instead.
//
// Clear requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Clear() {
	q.data = nil
	q.head = 0
	q.len = 0
}

// Reset removes all elements but keeps the backing array for reuse, so Cap is
// preserved. The elements are zeroed so the array no longer retains references
// to them. To also release the backing array, use [Queue.Clear] instead.
//
// Reset requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Reset() {
	q.clearLive()
	q.head = 0
	q.len = 0
}

func (q *Queue[T]) index(offset int) int {
	return (q.head + offset) % len(q.data)
}

func (q *Queue[T]) segments() ([]T, []T) {
	if q.len == 0 {
		return nil, nil
	}
	end := q.head + q.len
	if end <= len(q.data) {
		return q.data[q.head:end], nil
	}
	return q.data[q.head:], q.data[:end%len(q.data)]
}

func (q *Queue[T]) copyLiveTo(dst []T) {
	first, second := q.segments()
	n := copy(dst, first)
	copy(dst[n:], second)
}

func (q *Queue[T]) clearLive() {
	first, second := q.segments()
	clear(first)
	clear(second)
}

func (q *Queue[T]) growFor(need int) {
	if need <= len(q.data) {
		return
	}

	newCap := nextCap(len(q.data), need)
	grown := make([]T, newCap)
	q.copyLiveTo(grown)
	q.data = grown
	q.head = 0
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
