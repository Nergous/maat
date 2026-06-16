package queue

import "testing"

// Package-level sinks prevent the compiler from optimizing away benchmarked
// results (and the work that produces them).
var (
	intSink   int
	boolSink  bool
	sliceSink []int
)

// BenchmarkPush measures the amortized cost of a single Push, including the
// occasional reallocation as the backing array grows. The queue grows across
// the whole run, so the cost of those reallocations is amortized honestly over
// every push rather than hidden behind a one-time prealloc.
func BenchmarkPush(b *testing.B) {
	b.ReportAllocs()

	q := New[int]()
	i := 0
	for b.Loop() {
		q.Push(i)
		i++
	}
	// Keep the final size observable so the pushes are not dead code.
	intSink = q.Len()
}

// BenchmarkPushN measures a single bulk push of a fixed batch, the cheaper
// alternative to repeated Push since it grows the backing array at most once.
// The batch is built once outside the timed loop; the queue is reset (capacity
// retained) each iteration so we measure the bulk append, not repeated growth.
func BenchmarkPushN(b *testing.B) {
	const batch = 1024

	vs := make([]int, batch)
	for i := range vs {
		vs[i] = i
	}

	b.ReportAllocs()

	q := NewWithCap[int](batch)
	for b.Loop() {
		q.PushN(vs...)
		b.StopTimer()
		q.Reset() // retain capacity so the next PushN does not reallocate
		b.StartTimer()
	}
	intSink = q.Len()
}

// BenchmarkPop measures Pop in isolation. Refilling happens outside the timed
// section: whenever the queue drains we push a fresh chunk with the timer
// stopped, so only the Pop calls are measured.
func BenchmarkPop(b *testing.B) {
	const chunk = 1024

	b.ReportAllocs()

	q := NewWithCap[int](chunk)
	for b.Loop() {
		if q.IsEmpty() {
			b.StopTimer()
			q.Reset() // restore the full backing window before refilling
			for j := range chunk {
				q.Push(j)
			}
			b.StartTimer()
		}
		_, boolSink = q.Pop()
	}
}

// BenchmarkPeek measures the cost of reading the front element, which must stay
// O(1) and allocation-free.
func BenchmarkPeek(b *testing.B) {
	b.ReportAllocs()

	q := New[int]()
	q.PushN(1, 2, 3)

	for b.Loop() {
		intSink, boolSink = q.Peek()
	}
}

// BenchmarkPushPopChurn measures steady-state, interleaved Push/Pop at a fixed
// live size: each iteration pops the front and pushes a fresh element, so the
// queue never grows beyond `live` elements. This is the queue-specific workload
// that matters. With the current sliding-slice backing the window marches to the
// end of the array and must periodically copy the live region into a fresh
// allocation, so this benchmark reports recurring allocations per op; a future
// ring-buffer backing would drive that to zero. The numbers here are the
// baseline to beat.
func BenchmarkPushPopChurn(b *testing.B) {
	const live = 16

	q := NewWithCap[int](live)
	for i := range live {
		q.Push(i)
	}

	b.ReportAllocs()

	i := live
	for b.Loop() {
		_, boolSink = q.Pop()
		q.Push(i)
		i++
	}
	intSink = q.Len()
}

// BenchmarkAll measures a full iteration over the queue via the range-over-func
// iterator, confirming iteration is allocation-free.
func BenchmarkAll(b *testing.B) {
	const n = 1024

	q := NewWithCap[int](n)
	for i := range n {
		q.Push(i)
	}

	b.ReportAllocs()
	for b.Loop() {
		sum := 0
		for v := range q.All() {
			sum += v
		}
		intSink = sum
	}
}

// BenchmarkSlice measures copying the queue out to a plain slice, which
// allocates exactly one backing array per call.
func BenchmarkSlice(b *testing.B) {
	const n = 1024

	q := NewWithCap[int](n)
	for i := range n {
		q.Push(i)
	}

	b.ReportAllocs()
	for b.Loop() {
		sliceSink = q.Slice()
	}
}
