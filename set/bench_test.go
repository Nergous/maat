package set

import (
	"fmt"
	"testing"
)

// Package-level sinks prevent the compiler from optimizing away benchmarked
// results (and the work that produces them).
var (
	intSink   int
	boolSink  bool
	sliceSink []int
	setSink   *Set[int]
)

// BenchmarkAdd measures the expected amortized cost of inserting new values,
// including occasional map growth.
func BenchmarkAdd(b *testing.B) {
	b.ReportAllocs()

	s := New[int]()
	i := 0
	for b.Loop() {
		s.Add(i)
		i++
	}
	intSink = s.Len()
}

// BenchmarkAddExisting measures the membership-update path when values already
// exist and the map does not grow.
func BenchmarkAddExisting(b *testing.B) {
	const n = 1024

	s := NewWithCap[int](n)
	for i := range n {
		s.Add(i)
	}

	b.ReportAllocs()

	i := 0
	for b.Loop() {
		boolSink = s.Add(i % n)
		i++
	}
}

// BenchmarkContains measures lookup cost for hits and misses.
func BenchmarkContains(b *testing.B) {
	const n = 1024

	s := NewWithCap[int](n)
	for i := range n {
		s.Add(i)
	}

	b.Run("hit", func(b *testing.B) {
		b.ReportAllocs()

		i := 0
		for b.Loop() {
			boolSink = s.Contains(i % n)
			i++
		}
	})

	b.Run("miss", func(b *testing.B) {
		b.ReportAllocs()

		i := n
		for b.Loop() {
			boolSink = s.Contains(i)
			i++
		}
	})
}

// BenchmarkRemove measures Remove in isolation. Refilling happens outside the
// timed section: whenever the set drains we add a fresh chunk with the timer
// stopped, so only removals are measured.
func BenchmarkRemove(b *testing.B) {
	const chunk = 1024

	b.ReportAllocs()

	s := NewWithCap[int](chunk)
	next := 0
	for b.Loop() {
		if s.IsEmpty() {
			b.StopTimer()
			for j := range chunk {
				s.Add(next + j)
			}
			next += chunk
			b.StartTimer()
		}
		boolSink = s.Remove(next - s.Len())
	}
}

// BenchmarkUnion measures non-mutating union across several input sizes.
func BenchmarkUnion(b *testing.B) {
	for _, size := range []int{16, 1024, 65536} {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			a := NewWithCap[int](size)
			c := NewWithCap[int](size)
			for i := range size {
				a.Add(i)
				c.Add(i + size/2)
			}

			b.ReportAllocs()
			for b.Loop() {
				setSink = a.Union(c)
			}
		})
	}
}

// BenchmarkIntersection measures non-mutating intersection across several input
// sizes.
func BenchmarkIntersection(b *testing.B) {
	for _, size := range []int{16, 1024, 65536} {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			a := NewWithCap[int](size)
			c := NewWithCap[int](size)
			for i := range size {
				a.Add(i)
				c.Add(i + size/2)
			}

			b.ReportAllocs()
			for b.Loop() {
				setSink = a.Intersection(c)
			}
		})
	}
}

// BenchmarkAll measures full iteration over the set via the range-over-func
// iterator.
func BenchmarkAll(b *testing.B) {
	const n = 1024

	s := NewWithCap[int](n)
	for i := range n {
		s.Add(i)
	}

	b.ReportAllocs()
	for b.Loop() {
		sum := 0
		for v := range s.All() {
			sum += v
		}
		intSink = sum
	}
}

// BenchmarkSlice measures copying the set out to a plain slice, which allocates
// exactly one backing array per call.
func BenchmarkSlice(b *testing.B) {
	const n = 1024

	s := NewWithCap[int](n)
	for i := range n {
		s.Add(i)
	}

	b.ReportAllocs()
	for b.Loop() {
		sliceSink = s.Slice()
	}
}
