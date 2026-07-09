package deque

import "testing"

// Package-level sinks prevent the compiler from optimizing away benchmarked
// results (and the work that produces them).
var (
	intSink   int
	boolSink  bool
	sliceSink []int
)

// BenchmarkPushFront measures the amortized cost of a single PushFront,
// including occasional ring-buffer growth.
func BenchmarkPushFront(b *testing.B) {
	b.ReportAllocs()

	d := New[int]()
	i := 0
	for b.Loop() {
		d.PushFront(i)
		i++
	}
	intSink = d.Len()
}

// BenchmarkPushBack measures the amortized cost of a single PushBack,
// including occasional ring-buffer growth.
func BenchmarkPushBack(b *testing.B) {
	b.ReportAllocs()

	d := New[int]()
	i := 0
	for b.Loop() {
		d.PushBack(i)
		i++
	}
	intSink = d.Len()
}

// BenchmarkPushFrontN measures a fixed-size bulk push at the front. The deque
// is reset each iteration so the benchmark focuses on the batch write path, not
// repeated growth.
func BenchmarkPushFrontN(b *testing.B) {
	const batch = 1024

	vs := make([]int, batch)
	for i := range vs {
		vs[i] = i
	}

	b.ReportAllocs()

	d := NewWithCap[int](batch)
	for b.Loop() {
		d.PushFrontN(vs...)
		b.StopTimer()
		d.Reset()
		b.StartTimer()
	}
	intSink = d.Len()
}

// BenchmarkPushBackN measures a fixed-size bulk push at the back. The deque is
// reset each iteration so the benchmark focuses on the batch write path, not
// repeated growth.
func BenchmarkPushBackN(b *testing.B) {
	const batch = 1024

	vs := make([]int, batch)
	for i := range vs {
		vs[i] = i
	}

	b.ReportAllocs()

	d := NewWithCap[int](batch)
	for b.Loop() {
		d.PushBackN(vs...)
		b.StopTimer()
		d.Reset()
		b.StartTimer()
	}
	intSink = d.Len()
}

// BenchmarkPopFront measures PopFront in isolation. Refilling happens outside
// the timed section so only removals are measured.
func BenchmarkPopFront(b *testing.B) {
	const chunk = 1024

	b.ReportAllocs()

	d := NewWithCap[int](chunk)
	for b.Loop() {
		if d.IsEmpty() {
			b.StopTimer()
			d.Reset()
			for j := range chunk {
				d.PushBack(j)
			}
			b.StartTimer()
		}
		_, boolSink = d.PopFront()
	}
}

// BenchmarkPopBack measures PopBack in isolation. Refilling happens outside the
// timed section so only removals are measured.
func BenchmarkPopBack(b *testing.B) {
	const chunk = 1024

	b.ReportAllocs()

	d := NewWithCap[int](chunk)
	for b.Loop() {
		if d.IsEmpty() {
			b.StopTimer()
			d.Reset()
			for j := range chunk {
				d.PushBack(j)
			}
			b.StartTimer()
		}
		_, boolSink = d.PopBack()
	}
}

// BenchmarkPopFrontN measures draining fixed-size batches from the front.
// Refilling happens outside the timed section so the benchmark focuses on batch
// removal and result-slice allocation.
func BenchmarkPopFrontN(b *testing.B) {
	const batch = 128

	b.ReportAllocs()

	d := NewWithCap[int](batch)
	for b.Loop() {
		if d.Len() < batch {
			b.StopTimer()
			d.Reset()
			for j := range batch {
				d.PushBack(j)
			}
			b.StartTimer()
		}
		sliceSink = d.PopFrontN(batch)
	}
}

// BenchmarkPopBackN measures draining fixed-size batches from the back.
// Refilling happens outside the timed section so the benchmark focuses on batch
// removal and result-slice allocation.
func BenchmarkPopBackN(b *testing.B) {
	const batch = 128

	b.ReportAllocs()

	d := NewWithCap[int](batch)
	for b.Loop() {
		if d.Len() < batch {
			b.StopTimer()
			d.Reset()
			for j := range batch {
				d.PushBack(j)
			}
			b.StartTimer()
		}
		sliceSink = d.PopBackN(batch)
	}
}

// BenchmarkFront measures reading the front element, which must stay O(1) and
// allocation-free.
func BenchmarkFront(b *testing.B) {
	b.ReportAllocs()

	d := New[int]()
	d.PushBackN(1, 2, 3)

	for b.Loop() {
		intSink, boolSink = d.Front()
	}
}

// BenchmarkBack measures reading the back element, which must stay O(1) and
// allocation-free.
func BenchmarkBack(b *testing.B) {
	b.ReportAllocs()

	d := New[int]()
	d.PushBackN(1, 2, 3)

	for b.Loop() {
		intSink, boolSink = d.Back()
	}
}

// BenchmarkPushPopFrontBackChurn measures steady-state work at a fixed live
// size: each iteration removes one end and pushes to the other. The ring-buffer
// backing should keep this allocation-free after initial preallocation.
func BenchmarkPushPopFrontBackChurn(b *testing.B) {
	const live = 16

	d := NewWithCap[int](live)
	for i := range live {
		d.PushBack(i)
	}

	b.ReportAllocs()

	i := live
	for b.Loop() {
		_, boolSink = d.PopFront()
		d.PushBack(i)
		_, boolSink = d.PopBack()
		d.PushFront(i)
		i++
	}
	intSink = d.Len()
}

// BenchmarkAll measures a full iteration over the deque via the
// range-over-func iterator, confirming iteration is allocation-free.
func BenchmarkAll(b *testing.B) {
	const n = 1024

	d := NewWithCap[int](n)
	for i := range n {
		d.PushBack(i)
	}

	b.ReportAllocs()
	for b.Loop() {
		sum := 0
		for v := range d.All() {
			sum += v
		}
		intSink = sum
	}
}

// BenchmarkSlice measures copying the deque out to a plain slice, which
// allocates exactly one backing array per call.
func BenchmarkSlice(b *testing.B) {
	const n = 1024

	d := NewWithCap[int](n)
	for i := range n {
		d.PushBack(i)
	}

	b.ReportAllocs()
	for b.Loop() {
		sliceSink = d.Slice()
	}
}
