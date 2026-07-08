package queue

import "testing"

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func assertNewEmpty[T any](t *testing.T) {
	t.Helper()

	q := New[T]()
	if got := q.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
	if got := cap(q.data); got != 0 {
		t.Errorf("cap(data) = %d, want 0", got)
	}
}

func TestQueue_New(t *testing.T) {
	t.Run("int", assertNewEmpty[int])
	t.Run("string", assertNewEmpty[string])
	t.Run("float32", assertNewEmpty[float32])
}

func TestQueue_NewWithCap(t *testing.T) {
	tests := []struct {
		name    string
		cap     int
		wantCap int
	}{
		{name: "zero cap", cap: 0, wantCap: 0},
		{name: "one", cap: 1, wantCap: 1},
		{name: "ten", cap: 10, wantCap: 10},
		// A negative argument must be clamped to 0 instead of panicking in
		// make([]T, 0, n) (mirrors stack.NewWithCap).
		{name: "negative clamps to zero", cap: -1, wantCap: 0},
		{name: "large negative clamps to zero", cap: -1000, wantCap: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewWithCap[int](tt.cap)
			if got := q.Len(); got != 0 {
				t.Errorf("Len() = %d, want 0", got)
			}
			if got := cap(q.data); got != tt.wantCap {
				t.Errorf("cap(data) = %d, want %d", got, tt.wantCap)
			}
			// the queue must be usable regardless of the requested capacity.
			q.Push(1)
			if v, ok := q.Pop(); !ok || v != 1 {
				t.Errorf("Pop() = (%d, %v), want (1, true)", v, ok)
			}
		})
	}
}

func TestQueue_PushPop(t *testing.T) {
	tests := []struct {
		name string
		push []int
		want []int
	}{
		{name: "empty", push: nil, want: nil},
		{name: "single element", push: []int{42}, want: []int{42}},
		{name: "fifo order", push: []int{1, 2, 3}, want: []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := New[int]()
			for _, v := range tt.push {
				q.Push(v)
			}

			for i, want := range tt.want {
				got, ok := q.Pop()
				if !ok {
					t.Fatalf("Pop() #%d: ok = false, want true", i)
				}
				if got != want {
					t.Errorf("Pop() #%d = %d, want %d", i, got, want)
				}
			}

			if got, ok := q.Pop(); ok {
				t.Errorf("Pop() on empty: got %d, ok = true, want ok = false", got)
			}
		})
	}
}

func TestQueue_PushPopInterleaved(t *testing.T) {
	// FIFO order must hold even when pushes and pops are interleaved.
	q := New[int]()

	q.Push(1)
	q.Push(2)

	if got, ok := q.Pop(); !ok || got != 1 {
		t.Fatalf("Pop() = (%d, %v), want (1, true)", got, ok)
	}

	q.Push(3)

	for _, want := range []int{2, 3} {
		got, ok := q.Pop()
		if !ok {
			t.Fatalf("Pop(): ok = false, want true")
		}
		if got != want {
			t.Errorf("Pop() = %d, want %d", got, want)
		}
	}

	if got, ok := q.Pop(); ok {
		t.Errorf("Pop() on empty: got %d, ok = true, want ok = false", got)
	}
}

func TestQueue_PopReleasesReference(t *testing.T) {
	// Pop must zero the vacated slot so the backing array no longer retains a
	// reference to the popped element (otherwise it leaks until reallocation).
	q := New[*int]()
	a := 1
	q.Push(&a)

	// alias the same backing array so we can inspect the slot after Pop slices it off.
	backing := q.data[:1]

	if _, ok := q.Pop(); !ok {
		t.Fatalf("Pop() ok = false, want true")
	}
	if backing[0] != nil {
		t.Errorf("backing slot 0 = %p, want nil (Pop must release the reference)", backing[0])
	}
}

func TestQueue_Peek(t *testing.T) {
	t.Run("empty returns zero value and false", func(t *testing.T) {
		q := New[int]()
		got, ok := q.Peek()
		if ok {
			t.Errorf("Peek() ok = true, want false")
		}
		if got != 0 {
			t.Errorf("Peek() = %d, want 0 (zero value)", got)
		}
	})

	t.Run("returns front without removing it", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)

		got, ok := q.Peek()
		if !ok {
			t.Fatalf("Peek() ok = false, want true")
		}
		if got != 1 {
			t.Errorf("Peek() = %d, want 1 (front of queue)", got)
		}
		if got := q.Len(); got != 2 {
			t.Errorf("Len() after Peek() = %d, want 2 (Peek must not modify the queue)", got)
		}
	})
}

func TestQueue_IsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		pushN int
		want  bool
	}{
		{name: "new queue is empty", pushN: 0, want: true},
		{name: "after push is not empty", pushN: 1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := New[int]()
			for i := 0; i < tt.pushN; i++ {
				q.Push(i)
			}
			if got := q.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueue_Cap(t *testing.T) {
	t.Run("new queue has zero cap", func(t *testing.T) {
		q := New[int]()
		if got := q.Cap(); got != 0 {
			t.Errorf("Cap() = %d, want 0", got)
		}
	})

	t.Run("reflects preallocated capacity", func(t *testing.T) {
		q := NewWithCap[int](8)
		if got := q.Cap(); got != 8 {
			t.Errorf("Cap() = %d, want 8", got)
		}
		if got := q.Len(); got != 0 {
			t.Errorf("Len() = %d, want 0", got)
		}
	})

	t.Run("never less than len while growing", func(t *testing.T) {
		q := New[int]()
		for i := range 100 {
			q.Push(i)
			if c, l := q.Cap(), q.Len(); c < l {
				t.Fatalf("after %d pushes: Cap() = %d < Len() = %d", i+1, c, l)
			}
		}
	})

	t.Run("no realloc while pushing within preallocated cap", func(t *testing.T) {
		q := NewWithCap[int](4)
		for i := range 4 {
			q.Push(i)
		}
		if got := q.Cap(); got != 4 {
			t.Errorf("Cap() = %d, want 4 (no realloc expected within cap)", got)
		}
	})

	t.Run("stays stable when elements are popped", func(t *testing.T) {
		q := NewWithCap[int](4)
		q.PushN(1, 2, 3, 4)
		q.Pop()
		q.Pop()

		if got := q.Cap(); got != 4 {
			t.Errorf("Cap() after two Pop() calls = %d, want 4", got)
		}
	})

	t.Run("reuses front slots after pop without growing", func(t *testing.T) {
		q := NewWithCap[int](4)
		q.PushN(1, 2, 3, 4)
		q.Pop()
		q.Pop()

		q.Push(5)
		q.Push(6)

		if got := q.Cap(); got != 4 {
			t.Errorf("Cap() after wraparound pushes = %d, want 4", got)
		}
		if got, want := q.Slice(), []int{3, 4, 5, 6}; !equalInts(got, want) {
			t.Errorf("Slice() after wraparound pushes = %v, want %v", got, want)
		}
	})
}

func TestQueue_Clear(t *testing.T) {
	q := NewWithCap[int](8)
	for i := range 5 {
		q.Push(i)
	}

	q.Clear()

	if got := q.Len(); got != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", got)
	}
	if !q.IsEmpty() {
		t.Errorf("IsEmpty() after Clear() = false, want true")
	}
	if got := q.Cap(); got != 0 {
		t.Errorf("Cap() after Clear() = %d, want 0 (backing array released)", got)
	}
	if v, ok := q.Pop(); ok {
		t.Errorf("Pop() after Clear() = (%d, true), want ok = false", v)
	}

	// queue stays usable after Clear
	q.Push(99)
	if v, ok := q.Pop(); !ok || v != 99 {
		t.Errorf("Pop() after reuse = (%d, %v), want (99, true)", v, ok)
	}
}

func TestQueue_Reset(t *testing.T) {
	t.Run("empties queue but preserves capacity", func(t *testing.T) {
		q := New[int]()
		for i := range 10 {
			q.Push(i)
		}
		capBefore := q.Cap()

		q.Reset()

		if got := q.Len(); got != 0 {
			t.Errorf("Len() after Reset() = %d, want 0", got)
		}
		if !q.IsEmpty() {
			t.Errorf("IsEmpty() after Reset() = false, want true")
		}
		if got := q.Cap(); got != capBefore {
			t.Errorf("Cap() after Reset() = %d, want %d (capacity must be preserved)", got, capBefore)
		}
	})

	t.Run("zeroes backing array to release references", func(t *testing.T) {
		q := New[*int]()
		a, b, c := 1, 2, 3
		q.Push(&a)
		q.Push(&b)
		q.Push(&c)

		q.Reset()

		// white-box: the array is retained, so inspect every slot up to cap.
		full := q.data[:cap(q.data)]
		for i, p := range full {
			if p != nil {
				t.Errorf("backing slot %d = %p, want nil (Reset must zero elements)", i, p)
			}
		}
	})

	t.Run("queue stays usable after Reset", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Reset()
		q.Push(42)

		if v, ok := q.Pop(); !ok || v != 42 {
			t.Errorf("Pop() after reuse = (%d, %v), want (42, true)", v, ok)
		}
		if v, ok := q.Pop(); ok {
			t.Errorf("Pop() on emptied queue = (%d, true), want ok = false", v)
		}
	})

	t.Run("preserves full capacity after pops", func(t *testing.T) {
		q := NewWithCap[int](8)
		q.PushN(1, 2, 3, 4, 5)
		q.Pop()
		q.Pop()

		q.Reset()

		if got := q.Cap(); got != 8 {
			t.Errorf("Cap() after Pop(), Pop(), Reset() = %d, want 8", got)
		}
	})
}

func TestQueue_Wraparound(t *testing.T) {
	q := NewWithCap[int](5)
	q.PushN(1, 2, 3, 4, 5)

	for _, want := range []int{1, 2, 3} {
		if got, ok := q.Pop(); !ok || got != want {
			t.Fatalf("Pop() = (%d, %v), want (%d, true)", got, ok, want)
		}
	}

	q.PushN(6, 7, 8)

	if got := q.Cap(); got != 5 {
		t.Errorf("Cap() after wraparound = %d, want 5", got)
	}
	if got, want := q.Slice(), []int{4, 5, 6, 7, 8}; !equalInts(got, want) {
		t.Errorf("Slice() across wraparound = %v, want %v", got, want)
	}

	var iterated []int
	for v := range q.All() {
		iterated = append(iterated, v)
	}
	if want := []int{4, 5, 6, 7, 8}; !equalInts(iterated, want) {
		t.Errorf("All() across wraparound = %v, want %v", iterated, want)
	}

	c := q.Clone()
	q.Pop()
	q.Push(9)

	if got, want := c.Slice(), []int{4, 5, 6, 7, 8}; !equalInts(got, want) {
		t.Errorf("Clone().Slice() across wraparound = %v, want %v", got, want)
	}
}

func TestQueue_PushN(t *testing.T) {
	t.Run("pushes all in order, last argument at the back", func(t *testing.T) {
		q := New[int]()
		q.PushN(1, 2, 3)

		// FIFO: they come back out in the order they went in.
		for _, want := range []int{1, 2, 3} {
			if got, ok := q.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("no arguments is a no-op", func(t *testing.T) {
		q := New[int]()
		q.PushN()
		if got := q.Len(); got != 0 {
			t.Errorf("Len() after PushN() = %d, want 0", got)
		}
	})

	t.Run("appends onto existing elements", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.PushN(2, 3)

		if got, want := q.Slice(), []int{1, 2, 3}; !equalInts(got, want) {
			t.Errorf("Slice() = %v, want %v", got, want)
		}
	})

	t.Run("grows the backing array at most once", func(t *testing.T) {
		q := New[int]()
		q.PushN(1, 2, 3, 4, 5, 6, 7, 8)
		// A single bulk append should land on a capacity that holds all 8
		// without a second growth step.
		if got := q.Cap(); got < 8 {
			t.Errorf("Cap() after PushN(8 elems) = %d, want >= 8", got)
		}
	})
}

func TestQueue_PopN(t *testing.T) {
	t.Run("pops up to n elements in FIFO order", func(t *testing.T) {
		q := New[int]()
		q.PushN(1, 2, 3, 4)

		if got, want := q.PopN(3), []int{1, 2, 3}; !equalInts(got, want) {
			t.Errorf("PopN(3) = %v, want %v", got, want)
		}
		if got, want := q.Slice(), []int{4}; !equalInts(got, want) {
			t.Errorf("remaining Slice() = %v, want %v", got, want)
		}
	})

	t.Run("larger than len drains queue", func(t *testing.T) {
		q := NewWithCap[int](4)
		q.PushN(1, 2)

		if got, want := q.PopN(10), []int{1, 2}; !equalInts(got, want) {
			t.Errorf("PopN(10) = %v, want %v", got, want)
		}
		if got := q.Len(); got != 0 {
			t.Errorf("Len() after PopN(10) = %d, want 0", got)
		}
		if got := q.Cap(); got != 4 {
			t.Errorf("Cap() after PopN(10) = %d, want 4", got)
		}
	})

	t.Run("non-positive n returns nil", func(t *testing.T) {
		q := New[int]()
		q.PushN(1, 2)

		if got := q.PopN(0); got != nil {
			t.Errorf("PopN(0) = %v, want nil", got)
		}
		if got := q.PopN(-1); got != nil {
			t.Errorf("PopN(-1) = %v, want nil", got)
		}
		if got := q.Len(); got != 2 {
			t.Errorf("Len() after non-positive PopN = %d, want 2", got)
		}
	})

	t.Run("works across wraparound", func(t *testing.T) {
		q := NewWithCap[int](5)
		q.PushN(1, 2, 3, 4, 5)
		q.Pop()
		q.Pop()
		q.PushN(6, 7)

		if got, want := q.PopN(4), []int{3, 4, 5, 6}; !equalInts(got, want) {
			t.Errorf("PopN(4) across wraparound = %v, want %v", got, want)
		}
		if got, want := q.Slice(), []int{7}; !equalInts(got, want) {
			t.Errorf("remaining Slice() = %v, want %v", got, want)
		}
	})

	t.Run("nil receiver returns nil", func(t *testing.T) {
		var q *Queue[int]
		if got := q.PopN(3); got != nil {
			t.Errorf("nil.PopN(3) = %v, want nil", got)
		}
	})
}

func TestQueue_All(t *testing.T) {
	collect := func(q *Queue[int]) []int {
		var out []int
		for v := range q.All() {
			out = append(out, v)
		}
		return out
	}

	t.Run("yields front to back (FIFO)", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)
		q.Push(3)

		if got, want := collect(q), []int{1, 2, 3}; !equalInts(got, want) {
			t.Errorf("All() = %v, want %v", got, want)
		}
	})

	t.Run("empty queue yields nothing", func(t *testing.T) {
		q := New[int]()
		if got := collect(q); len(got) != 0 {
			t.Errorf("All() over empty = %v, want no elements", got)
		}
	})

	t.Run("yields front to back after some Pops", func(t *testing.T) {
		q := New[int]()
		q.PushN(1, 2, 3, 4)
		q.Pop() // drop the 1; front advances

		if got, want := collect(q), []int{2, 3, 4}; !equalInts(got, want) {
			t.Errorf("All() after Pop = %v, want %v", got, want)
		}
	})

	t.Run("does not consume the queue", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)

		_ = collect(q)

		if got := q.Len(); got != 2 {
			t.Errorf("Len() after All() = %d, want 2 (iteration must not consume)", got)
		}
		if v, ok := q.Pop(); !ok || v != 1 {
			t.Errorf("Pop() after All() = (%d, %v), want (1, true)", v, ok)
		}
	})

	t.Run("break stops iteration early and leaves queue intact", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)
		q.Push(3)

		var got []int
		for v := range q.All() {
			got = append(got, v)
			if len(got) == 2 {
				break
			}
		}

		if want := []int{1, 2}; !equalInts(got, want) {
			t.Errorf("All() with early break = %v, want %v", got, want)
		}
		if l := q.Len(); l != 3 {
			t.Errorf("Len() after break = %d, want 3", l)
		}
	})
}

func TestQueue_Slice(t *testing.T) {
	t.Run("returns elements front to back (FIFO)", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)
		q.Push(3)

		if got, want := q.Slice(), []int{1, 2, 3}; !equalInts(got, want) {
			t.Errorf("Slice() = %v, want %v", got, want)
		}
	})

	t.Run("returns an independent copy", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)

		sl := q.Slice()
		for i := range sl {
			sl[i] = 999
		}

		// mutating the returned slice must not disturb the queue's contents.
		for _, want := range []int{1, 2} {
			if got, ok := q.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v) after mutating Slice(), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("does not consume the queue", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)

		_ = q.Slice()

		if got := q.Len(); got != 2 {
			t.Errorf("Len() after Slice() = %d, want 2 (must not consume)", got)
		}
	})

	t.Run("empty queue returns nil regardless of how it became empty", func(t *testing.T) {
		drained := NewWithCap[int](8)
		drained.Push(1)
		drained.Pop()

		reset := NewWithCap[int](8)
		reset.Push(1)
		reset.Reset()

		cases := map[string]*Queue[int]{
			"fresh New":       New[int](),
			"preallocated":    NewWithCap[int](8),
			"drained to zero": drained,
			"after Reset":     reset,
		}
		for name, q := range cases {
			if got := q.Slice(); got != nil {
				t.Errorf("%s: Slice() = %v, want nil", name, got)
			}
		}
	})
}

func TestQueue_Clone(t *testing.T) {
	t.Run("copies all elements in order", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)
		q.Push(3)

		c := q.Clone()

		if got, want := c.Slice(), []int{1, 2, 3}; !equalInts(got, want) {
			t.Errorf("Clone().Slice() = %v, want %v", got, want)
		}
		if got := c.Len(); got != 3 {
			t.Errorf("Clone().Len() = %d, want 3", got)
		}
	})

	t.Run("is independent of the original", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)

		c := q.Clone()
		c.Push(99) // mutating the clone...
		q.Pop()    // ...and the original...

		// ...must not affect each other
		if got := c.Len(); got != 3 {
			t.Errorf("clone Len() = %d, want 3 (independent of original)", got)
		}
		if got := q.Len(); got != 1 {
			t.Errorf("original Len() = %d, want 1 (independent of clone)", got)
		}
	})

	t.Run("clone of empty queue is empty and usable", func(t *testing.T) {
		q := New[int]()

		c := q.Clone()

		if got := c.Len(); got != 0 {
			t.Errorf("Clone().Len() = %d, want 0", got)
		}
		c.Push(1)
		if v, ok := c.Pop(); !ok || v != 1 {
			t.Errorf("Pop() after push on cloned empty = (%d, %v), want (1, true)", v, ok)
		}
	})
}

func TestQueue_Grow(t *testing.T) {
	t.Run("reserves capacity without changing length or contents", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		q.Push(2)

		q.Grow(100)

		if got := q.Cap(); got < 102 {
			t.Errorf("Cap() after Grow(100) = %d, want >= 102", got)
		}
		if got := q.Len(); got != 2 {
			t.Errorf("Len() after Grow(100) = %d, want 2", got)
		}
		if got, want := q.Slice(), []int{1, 2}; !equalInts(got, want) {
			t.Errorf("Slice() after Grow = %v, want %v", got, want)
		}
	})

	t.Run("no reallocation while pushing within grown capacity", func(t *testing.T) {
		q := New[int]()
		q.Grow(8)
		capAfterGrow := q.Cap()

		for i := range 8 {
			q.Push(i)
		}

		if got := q.Cap(); got != capAfterGrow {
			t.Errorf("Cap() after pushing within grown cap = %d, want %d (no realloc)", got, capAfterGrow)
		}
	})

	t.Run("non-positive n is a no-op", func(t *testing.T) {
		q := New[int]()
		q.Push(1)
		capBefore := q.Cap()

		q.Grow(0)
		q.Grow(-5)

		if got := q.Cap(); got != capBefore {
			t.Errorf("Cap() after Grow(0)/Grow(-5) = %d, want %d (no-op)", got, capBefore)
		}
	})
}

func TestQueue_Shrink(t *testing.T) {
	t.Run("reduces capacity to length and preserves contents", func(t *testing.T) {
		q := NewWithCap[int](64)
		q.Push(1)
		q.Push(2)
		q.Push(3)

		q.Shrink()

		if got := q.Cap(); got != 3 {
			t.Errorf("Cap() after Shrink() = %d, want 3 (trimmed to len)", got)
		}
		if got := q.Len(); got != 3 {
			t.Errorf("Len() after Shrink() = %d, want 3", got)
		}
		for _, want := range []int{1, 2, 3} {
			if got, ok := q.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("allocates a new backing array", func(t *testing.T) {
		q := NewWithCap[int](64)
		q.Push(1)
		q.Push(2)
		q.Push(3)
		old := &q.data[0]

		q.Shrink()

		if &q.data[0] == old {
			t.Errorf("Shrink() reused the backing array, want a new right-sized one")
		}
		if got := q.Cap(); got != 3 {
			t.Errorf("Cap() after Shrink() = %d, want 3", got)
		}
	})

	t.Run("empty queue drops to zero capacity", func(t *testing.T) {
		q := NewWithCap[int](8)

		q.Shrink()

		if got := q.Cap(); got != 0 {
			t.Errorf("Cap() after Shrink() on empty = %d, want 0", got)
		}
		if got := q.Len(); got != 0 {
			t.Errorf("Len() after Shrink() on empty = %d, want 0", got)
		}
	})

	t.Run("queue stays usable after Shrink", func(t *testing.T) {
		q := NewWithCap[int](16)
		q.Push(1)
		q.Shrink()
		q.Push(2)

		if v, ok := q.Pop(); !ok || v != 1 {
			t.Errorf("Pop() after reuse = (%d, %v), want (1, true)", v, ok)
		}
		if v, ok := q.Pop(); !ok || v != 2 {
			t.Errorf("Pop() = (%d, %v), want (2, true)", v, ok)
		}
	})
}

func TestQueue_Clip(t *testing.T) {
	t.Run("reduces capacity to length and preserves contents", func(t *testing.T) {
		q := NewWithCap[int](64)
		q.Push(1)
		q.Push(2)
		q.Push(3)

		q.Clip()

		if got := q.Cap(); got != 3 {
			t.Errorf("Cap() after Clip() = %d, want 3 (trimmed to len)", got)
		}
		if got := q.Len(); got != 3 {
			t.Errorf("Len() after Clip() = %d, want 3", got)
		}
		for _, want := range []int{1, 2, 3} {
			if got, ok := q.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("does not allocate a new backing array", func(t *testing.T) {
		q := NewWithCap[int](64)
		q.Push(1)
		q.Push(2)
		q.Push(3)
		old := &q.data[0]

		q.Clip()

		if &q.data[0] != old {
			t.Errorf("Clip() copied to a new array, want the same backing array (no copy)")
		}
		if got := q.Cap(); got != 3 {
			t.Errorf("Cap() after Clip() = %d, want 3", got)
		}
	})

	t.Run("empty queue drops to zero capacity", func(t *testing.T) {
		q := NewWithCap[int](8)

		q.Clip()

		if got := q.Cap(); got != 0 {
			t.Errorf("Cap() after Clip() on empty = %d, want 0", got)
		}
		if got := q.Len(); got != 0 {
			t.Errorf("Len() after Clip() on empty = %d, want 0", got)
		}
	})

	t.Run("queue stays usable after Clip", func(t *testing.T) {
		q := NewWithCap[int](16)
		q.Push(1)
		q.Clip()
		q.Push(2)

		if v, ok := q.Pop(); !ok || v != 1 {
			t.Errorf("Pop() after reuse = (%d, %v), want (1, true)", v, ok)
		}
		if v, ok := q.Pop(); !ok || v != 2 {
			t.Errorf("Pop() = (%d, %v), want (2, true)", v, ok)
		}
	})
}

func TestQueue_SustainedChurn(t *testing.T) {
	// Interleave Push/Pop at a fixed live size over many iterations. FIFO order
	// must hold throughout and the capacity must stay bounded (it must not grow
	// without limit). With the current sliding-slice backing, Cap is measured
	// from the moving front and resets when the window slides and reallocates,
	// so the assertion is a generous bound rather than a tight equality.
	const live = 16
	const iterations = 100000

	q := New[int]()
	for i := range live {
		q.Push(i)
	}

	next := live
	maxCap := 0
	for i := range iterations {
		got, ok := q.Pop()
		if !ok {
			t.Fatalf("Pop() at iteration %d: ok = false, want true", i)
		}
		if want := next - live; got != want {
			t.Fatalf("Pop() at iteration %d = %d, want %d (FIFO order broken)", i, got, want)
		}
		q.Push(next)
		next++

		if c := q.Cap(); c > maxCap {
			maxCap = c
		}
		if got := q.Len(); got != live {
			t.Fatalf("Len() at iteration %d = %d, want %d (live size must stay fixed)", i, got, live)
		}
	}

	// The window only ever holds `live` elements, so the backing capacity must
	// stay a small multiple of that, never proportional to the iteration count.
	if maxCap > live*64 {
		t.Errorf("max Cap() during churn = %d, want bounded near live size %d", maxCap, live)
	}
}

func TestQueue_NilReceiver(t *testing.T) {
	// A nil *Queue[T] must behave as a valid empty queue for every read-only
	// method, without panicking.
	t.Run("Len is 0", func(t *testing.T) {
		var q *Queue[int]
		if got := q.Len(); got != 0 {
			t.Errorf("nil.Len() = %d, want 0", got)
		}
	})

	t.Run("Cap is 0", func(t *testing.T) {
		var q *Queue[int]
		if got := q.Cap(); got != 0 {
			t.Errorf("nil.Cap() = %d, want 0", got)
		}
	})

	t.Run("IsEmpty is true", func(t *testing.T) {
		var q *Queue[int]
		if !q.IsEmpty() {
			t.Errorf("nil.IsEmpty() = false, want true")
		}
	})

	t.Run("Peek returns zero, false", func(t *testing.T) {
		var q *Queue[int]
		if got, ok := q.Peek(); ok || got != 0 {
			t.Errorf("nil.Peek() = (%d, %v), want (0, false)", got, ok)
		}
	})

	t.Run("Pop returns zero, false", func(t *testing.T) {
		var q *Queue[int]
		if got, ok := q.Pop(); ok || got != 0 {
			t.Errorf("nil.Pop() = (%d, %v), want (0, false)", got, ok)
		}
	})

	t.Run("Slice is empty", func(t *testing.T) {
		var q *Queue[int]
		if got := q.Slice(); len(got) != 0 {
			t.Errorf("nil.Slice() = %v, want no elements", got)
		}
	})

	t.Run("All yields nothing", func(t *testing.T) {
		var q *Queue[int]
		count := 0
		for range q.All() {
			count++
		}
		if count != 0 {
			t.Errorf("nil.All() yielded %d elements, want 0", count)
		}
	})

	t.Run("Clone returns a usable empty queue", func(t *testing.T) {
		var q *Queue[int]

		c := q.Clone()
		if c == nil {
			t.Fatalf("nil.Clone() = nil, want a usable empty queue")
		}
		if got := c.Len(); got != 0 {
			t.Errorf("nil.Clone().Len() = %d, want 0", got)
		}

		// the clone must be usable for mutation without a further New call.
		c.Push(42)
		if v, ok := c.Pop(); !ok || v != 42 {
			t.Errorf("Pop() after push on cloned nil = (%d, %v), want (42, true)", v, ok)
		}
	})
}

func TestQueue_NilReceiver_MutatorsPanic(t *testing.T) {
	// The mutating methods need a non-nil receiver to store their results; a nil
	// receiver is a programming error and must panic.
	assertPanics := func(t *testing.T, name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s on nil receiver did not panic, want panic", name)
			}
		}()
		fn()
	}

	tests := []struct {
		name string
		call func(q *Queue[int])
	}{
		{name: "Push", call: func(q *Queue[int]) { q.Push(1) }},
		{name: "PushN", call: func(q *Queue[int]) { q.PushN(1, 2) }},
		{name: "PushN with no arguments", call: func(q *Queue[int]) { q.PushN() }},
		{name: "Grow", call: func(q *Queue[int]) { q.Grow(8) }},
		{name: "Grow zero", call: func(q *Queue[int]) { q.Grow(0) }},
		{name: "Grow negative", call: func(q *Queue[int]) { q.Grow(-1) }},
		{name: "Clear", call: func(q *Queue[int]) { q.Clear() }},
		{name: "Reset", call: func(q *Queue[int]) { q.Reset() }},
		{name: "Shrink", call: func(q *Queue[int]) { q.Shrink() }},
		{name: "Clip", call: func(q *Queue[int]) { q.Clip() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var q *Queue[int]
			assertPanics(t, tt.name, func() { tt.call(q) })
		})
	}
}
