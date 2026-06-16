// Package queue provides a generic, slice-backed FIFO (first-in, first-out) queue.
//
// Create a queue with [New] or [NewWithCap]. Elements are added to the back with
// [Queue.Push] and removed from the front with [Queue.Pop]; [Queue.Peek]
// inspects the front without removing it.
//
// Push and Pop run in amortized O(1) time (Push may occasionally reallocate
// the backing array as it grows). Peek, Len, Cap and IsEmpty run in O(1).
//
// A nil *Queue[T] behaves as an empty queue for every read-only method, so the
// zero pointer value is safe to query without first calling [New]. The mutating
// methods ([Queue.Push], [Queue.PushN], [Queue.Grow], [Queue.Clear],
// [Queue.Reset], [Queue.Shrink], [Queue.Clip]) require a non-nil receiver and
// panic on a nil one; see each method's documentation.
//
// A Queue is not safe for concurrent use: it performs no internal locking.
// Callers that share a queue across goroutines must provide their own
// synchronization.
package queue

import (
	"iter"
	"slices"
)

// Queue is a generic FIFO queue backed by a slice. The front of the queue is the
// oldest element still present, i.e. the next one to be removed. A Queue must be
// created with [New] or [NewWithCap] before any mutating method is called; a nil
// *Queue[T] is treated as a valid empty queue by the read-only methods only.
//
// Queue is not safe for concurrent use.
type Queue[T any] struct {
	data []T
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
	return &Queue[T]{data: make([]T, 0, n)}
}

// Len returns the number of elements currently in the queue. A nil receiver
// reports 0.
func (q *Queue[T]) Len() int {
	if q == nil {
		return 0
	}
	return len(q.data)
}

// Cap returns the capacity of the queue's backing array measured from the
// current front, i.e. the number of elements it can hold before the next
// reallocation. Because [Queue.Pop] advances the front past the consumed slot,
// Cap shrinks as elements are popped. A nil receiver reports 0.
func (q *Queue[T]) Cap() int {
	if q == nil {
		return 0
	}
	return cap(q.data)
}

// IsEmpty reports whether the queue has no elements. A nil receiver reports
// true.
func (q *Queue[T]) IsEmpty() bool {
	return q == nil || len(q.data) == 0
}

// Peek returns the front element (the oldest, next to be removed) without
// removing it. The boolean result is false when the queue is empty, in which
// case the returned value is the zero value of T. A nil receiver is treated as
// empty and returns (zero, false).
func (q *Queue[T]) Peek() (T, bool) {
	if q == nil || len(q.data) == 0 {
		var zero T
		return zero, false
	}
	return q.data[0], true
}

// All returns an iterator over the queue's elements from front to back (FIFO
// order), without consuming them; the queue is left unchanged. The returned
// [iter.Seq] is a Go 1.23 range-over-func iterator, so it can be used directly
// in a range loop. Breaking out of the loop early stops iteration cleanly. A nil
// receiver is treated as empty and yields no elements.
func (q *Queue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		if q == nil {
			return
		}
		for i := 0; i < len(q.data); i++ {
			if !yield(q.data[i]) {
				return
			}
		}
	}
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
	q.data = append(q.data, vs...)
}

// Pop removes and returns the front element (the oldest, next to be removed).
// The boolean result is false when the queue is empty, in which case the
// returned value is the zero value of T. The vacated slot is zeroed so it no
// longer retains a reference to the popped element, then the front is advanced.
// It runs in O(1) time. Note that advancing the front reduces the capacity
// reported by [Queue.Cap].
func (q *Queue[T]) Pop() (T, bool) {
	var zero T
	if q == nil || len(q.data) == 0 {
		return zero, false
	}
	v := q.data[0]
	q.data[0] = zero
	q.data = q.data[1:]
	return v, true
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

	return &Queue[T]{
		data: append([]T(nil), q.data...),
	}
}

// Slice returns a copy of the queue's elements from front to back (FIFO), the
// same order [Queue.All] yields. The returned slice is independent of the queue:
// they share no backing array, so callers may modify it freely without affecting
// the queue, and vice versa. An empty queue returns nil; a nil receiver is
// treated as empty and likewise returns nil. Runs in O(n).
func (q *Queue[T]) Slice() []T {
	if q == nil || len(q.data) == 0 {
		return nil
	}

	out := slices.Clone(q.data)
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
	if cap(q.data) == len(q.data) {
		return
	}

	trimmed := make([]T, len(q.data))
	copy(trimmed, q.data)
	q.data = trimmed
}

// Clip reduces the reported capacity to the current length without copying, by
// reslicing the backing array (see [slices.Clip]). It runs in O(1) and does not
// allocate. Because the backing array is retained, the unused tail memory is not
// returned to the garbage collector until a later growth reallocates. For an
// immediate, copying reclaim, use [Queue.Shrink].
//
// Clip requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Clip() {
	q.data = slices.Clip(q.data)
}

// Grow ensures the queue has spare capacity for at least n more elements without
// reallocating during subsequent pushes, growing the backing array if needed. It
// is the counterpart of [Queue.Shrink]. Grow runs in O(n) when it reallocates and
// is a no-op when n is non-positive or the capacity is already sufficient. It
// panics if the new capacity overflows, the same condition as [slices.Grow].
//
// Grow requires a non-nil receiver and panics on a nil one (unless n is
// non-positive, in which case it returns without touching the receiver). Create
// the queue with [New] or [NewWithCap] first.
func (q *Queue[T]) Grow(n int) {
	if n <= 0 {
		return
	}
	q.data = slices.Grow(q.data, n)
}

// Clear removes all elements and releases the backing array, so its memory
// becomes eligible for garbage collection. After Clear, Cap reports 0. To empty
// the queue while keeping its capacity for reuse, use [Queue.Reset] instead.
//
// Clear requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Clear() {
	q.data = nil
}

// Reset removes all elements but keeps the backing array for reuse, so Cap is
// preserved. The elements are zeroed so the array no longer retains references
// to them. To also release the backing array, use [Queue.Clear] instead.
//
// Reset requires a non-nil receiver and panics on a nil one. Create the queue
// with [New] or [NewWithCap] first.
func (q *Queue[T]) Reset() {
	clear(q.data)
	q.data = q.data[:0]
}
