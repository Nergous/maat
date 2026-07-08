package stack

import "testing"

func assertNewEmpty[T any](t *testing.T) {
	t.Helper()

	s := New[T]()
	if got := s.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
	if got := cap(s.data); got != 0 {
		t.Errorf("cap(data) = %d, want 0", got)
	}
}

func TestStack_New(t *testing.T) {
	t.Run("int", assertNewEmpty[int])
	t.Run("string", assertNewEmpty[string])
	t.Run("float32", assertNewEmpty[float32])
}

func TestStack_NewWithCap(t *testing.T) {
	tests := []struct {
		name string
		cap  int
	}{
		{name: "zero cap", cap: 0},
		{name: "one", cap: 1},
		{name: "ten", cap: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewWithCap[int](tt.cap)
			if got := s.Len(); got != 0 {
				t.Errorf("Len() = %d, want 0", got)
			}
			if got := cap(s.data); got != tt.cap {
				t.Errorf("cap(data) = %d, want %d", got, tt.cap)
			}
		})
	}
}

func TestStack_PushPop(t *testing.T) {
	tests := []struct {
		name string
		push []int
		want []int
	}{
		{name: "empty", push: nil, want: nil},
		{name: "single element", push: []int{42}, want: []int{42}},
		{name: "lifo order", push: []int{1, 2, 3}, want: []int{3, 2, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New[int]()
			for _, v := range tt.push {
				s.Push(v)
			}

			for i, want := range tt.want {
				got, ok := s.Pop()
				if !ok {
					t.Fatalf("Pop() #%d: ok = false, want true", i)
				}
				if got != want {
					t.Errorf("Pop() #%d = %d, want %d", i, got, want)
				}
			}

			if got, ok := s.Pop(); ok {
				t.Errorf("Pop() on empty: got %d, ok = true, want ok = false", got)
			}
		})
	}
}

func TestStack_Peek(t *testing.T) {
	t.Run("empty returns zero value and false", func(t *testing.T) {
		s := New[int]()
		got, ok := s.Peek()
		if ok {
			t.Errorf("Peek() ok = true, want false")
		}
		if got != 0 {
			t.Errorf("Peek() = %d, want 0 (zero value)", got)
		}
	})

	t.Run("returns top without removing it", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)

		got, ok := s.Peek()
		if !ok {
			t.Fatalf("Peek() ok = false, want true")
		}
		if got != 2 {
			t.Errorf("Peek() = %d, want 2", got)
		}
		if got := s.Len(); got != 2 {
			t.Errorf("Len() after Peek() = %d, want 2 (Peek must not modify the stack)", got)
		}
	})
}

func TestStack_IsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		pushN int
		want  bool
	}{
		{name: "new stack is empty", pushN: 0, want: true},
		{name: "after push is not empty", pushN: 1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New[int]()
			for i := 0; i < tt.pushN; i++ {
				s.Push(i)
			}
			if got := s.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_Cap(t *testing.T) {
	t.Run("new stack has zero cap", func(t *testing.T) {
		s := New[int]()
		if got := s.Cap(); got != 0 {
			t.Errorf("Cap() = %d, want 0", got)
		}
	})

	t.Run("reflects preallocated capacity", func(t *testing.T) {
		s := NewWithCap[int](8)
		if got := s.Cap(); got != 8 {
			t.Errorf("Cap() = %d, want 8", got)
		}
		if got := s.Len(); got != 0 {
			t.Errorf("Len() = %d, want 0", got)
		}
	})

	t.Run("never less than len while growing", func(t *testing.T) {
		s := New[int]()
		for i := range 100 {
			s.Push(i)
			if c, l := s.Cap(), s.Len(); c < l {
				t.Fatalf("after %d pushes: Cap() = %d < Len() = %d", i+1, c, l)
			}
		}
	})

	t.Run("no realloc while pushing within preallocated cap", func(t *testing.T) {
		s := NewWithCap[int](4)
		for i := range 4 {
			s.Push(i)
		}
		if got := s.Cap(); got != 4 {
			t.Errorf("Cap() = %d, want 4 (no realloc expected within cap)", got)
		}
	})
}

func TestStack_All(t *testing.T) {
	collect := func(s *Stack[int]) []int {
		var out []int
		for v := range s.All() {
			out = append(out, v)
		}
		return out
	}

	t.Run("yields top to bottom (LIFO)", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)
		s.Push(3)

		if got, want := collect(s), []int{3, 2, 1}; !equalInts(got, want) {
			t.Errorf("All() = %v, want %v", got, want)
		}
	})

	t.Run("empty stack yields nothing", func(t *testing.T) {
		s := New[int]()
		if got := collect(s); len(got) != 0 {
			t.Errorf("All() over empty = %v, want no elements", got)
		}
	})

	t.Run("does not modify the stack", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)

		_ = collect(s)

		if got := s.Len(); got != 2 {
			t.Errorf("Len() after All() = %d, want 2 (iteration must not consume)", got)
		}
		if v, ok := s.Pop(); !ok || v != 2 {
			t.Errorf("Pop() after All() = (%d, %v), want (2, true)", v, ok)
		}
	})

	t.Run("break stops iteration early and leaves stack intact", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)
		s.Push(3)

		var got []int
		for v := range s.All() {
			got = append(got, v)
			if len(got) == 2 {
				break
			}
		}

		if want := []int{3, 2}; !equalInts(got, want) {
			t.Errorf("All() with early break = %v, want %v", got, want)
		}
		if l := s.Len(); l != 3 {
			t.Errorf("Len() after break = %d, want 3", l)
		}
	})
}

func TestStack_Clear(t *testing.T) {
	s := NewWithCap[int](8)
	for i := range 5 {
		s.Push(i)
	}

	s.Clear()

	if got := s.Len(); got != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", got)
	}
	if !s.IsEmpty() {
		t.Errorf("IsEmpty() after Clear() = false, want true")
	}
	if got := s.Cap(); got != 0 {
		t.Errorf("Cap() after Clear() = %d, want 0 (backing array released)", got)
	}
	if v, ok := s.Pop(); ok {
		t.Errorf("Pop() after Clear() = (%d, true), want ok = false", v)
	}

	// stack stays usable after Clear
	s.Push(99)
	if v, ok := s.Pop(); !ok || v != 99 {
		t.Errorf("Pop() after reuse = (%d, %v), want (99, true)", v, ok)
	}
}

func TestStack_Reset(t *testing.T) {
	t.Run("empties stack but preserves capacity", func(t *testing.T) {
		s := New[int]()
		for i := range 10 {
			s.Push(i)
		}
		capBefore := s.Cap()

		s.Reset()

		if got := s.Len(); got != 0 {
			t.Errorf("Len() after Reset() = %d, want 0", got)
		}
		if !s.IsEmpty() {
			t.Errorf("IsEmpty() after Reset() = false, want true")
		}
		if got := s.Cap(); got != capBefore {
			t.Errorf("Cap() after Reset() = %d, want %d (capacity must be preserved)", got, capBefore)
		}
	})

	t.Run("zeroes backing array to release references", func(t *testing.T) {
		s := New[*int]()
		a, b, c := 1, 2, 3
		s.Push(&a)
		s.Push(&b)
		s.Push(&c)

		s.Reset()

		// white-box: the array is retained, so inspect every slot up to cap.
		full := s.data[:cap(s.data)]
		for i, p := range full {
			if p != nil {
				t.Errorf("backing slot %d = %p, want nil (Reset must zero elements)", i, p)
			}
		}
	})

	t.Run("stack stays usable after Reset", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Reset()
		s.Push(42)

		if v, ok := s.Pop(); !ok || v != 42 {
			t.Errorf("Pop() after reuse = (%d, %v), want (42, true)", v, ok)
		}
		if v, ok := s.Pop(); ok {
			t.Errorf("Pop() on emptied stack = (%d, true), want ok = false", v)
		}
	})
}

func TestStack_Shrink(t *testing.T) {
	t.Run("reduces capacity to length and preserves contents", func(t *testing.T) {
		s := NewWithCap[int](64)
		s.Push(1)
		s.Push(2)
		s.Push(3)

		s.Shrink()

		if got := s.Cap(); got != 3 {
			t.Errorf("Cap() after Shrink() = %d, want 3 (trimmed to len)", got)
		}
		if got := s.Len(); got != 3 {
			t.Errorf("Len() after Shrink() = %d, want 3", got)
		}
		for _, want := range []int{3, 2, 1} {
			if got, ok := s.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("allocates a new backing array", func(t *testing.T) {
		s := NewWithCap[int](64)
		s.Push(1)
		s.Push(2)
		s.Push(3)
		old := &s.data[0]

		s.Shrink()

		if &s.data[0] == old {
			t.Errorf("Shrink() reused the backing array, want a new right-sized one")
		}
		if got := s.Cap(); got != 3 {
			t.Errorf("Cap() after Shrink() = %d, want 3", got)
		}
	})

	t.Run("empty stack drops to zero capacity", func(t *testing.T) {
		s := NewWithCap[int](8)

		s.Shrink()

		if got := s.Cap(); got != 0 {
			t.Errorf("Cap() after Shrink() on empty = %d, want 0", got)
		}
		if got := s.Len(); got != 0 {
			t.Errorf("Len() after Shrink() on empty = %d, want 0", got)
		}
	})

	t.Run("stack stays usable after Shrink", func(t *testing.T) {
		s := NewWithCap[int](16)
		s.Push(1)
		s.Shrink()
		s.Push(2)

		if v, ok := s.Pop(); !ok || v != 2 {
			t.Errorf("Pop() after reuse = (%d, %v), want (2, true)", v, ok)
		}
		if v, ok := s.Pop(); !ok || v != 1 {
			t.Errorf("Pop() = (%d, %v), want (1, true)", v, ok)
		}
	})
}

func TestStack_Clip(t *testing.T) {
	t.Run("reduces capacity to length and preserves contents", func(t *testing.T) {
		s := NewWithCap[int](64)
		s.Push(1)
		s.Push(2)
		s.Push(3)

		s.Clip()

		if got := s.Cap(); got != 3 {
			t.Errorf("Cap() after Clip() = %d, want 3 (trimmed to len)", got)
		}
		if got := s.Len(); got != 3 {
			t.Errorf("Len() after Clip() = %d, want 3", got)
		}
		for _, want := range []int{3, 2, 1} {
			if got, ok := s.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("does not allocate a new backing array", func(t *testing.T) {
		s := NewWithCap[int](64)
		s.Push(1)
		s.Push(2)
		s.Push(3)
		old := &s.data[0]

		s.Clip()

		if &s.data[0] != old {
			t.Errorf("Clip() copied to a new array, want the same backing array (no copy)")
		}
		if got := s.Cap(); got != 3 {
			t.Errorf("Cap() after Clip() = %d, want 3", got)
		}
	})

	t.Run("empty stack drops to zero capacity", func(t *testing.T) {
		s := NewWithCap[int](8)

		s.Clip()

		if got := s.Cap(); got != 0 {
			t.Errorf("Cap() after Clip() on empty = %d, want 0", got)
		}
		if got := s.Len(); got != 0 {
			t.Errorf("Len() after Clip() on empty = %d, want 0", got)
		}
	})

	t.Run("stack stays usable after Clip", func(t *testing.T) {
		s := NewWithCap[int](16)
		s.Push(1)
		s.Clip()
		s.Push(2)

		if v, ok := s.Pop(); !ok || v != 2 {
			t.Errorf("Pop() after reuse = (%d, %v), want (2, true)", v, ok)
		}
		if v, ok := s.Pop(); !ok || v != 1 {
			t.Errorf("Pop() = (%d, %v), want (1, true)", v, ok)
		}
	})
}

func TestStack_PushN(t *testing.T) {
	t.Run("pushes all in order, last argument on top", func(t *testing.T) {
		s := New[int]()
		s.PushN(1, 2, 3)

		for _, want := range []int{3, 2, 1} {
			if got, ok := s.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("no arguments is a no-op", func(t *testing.T) {
		s := New[int]()
		s.PushN()
		if got := s.Len(); got != 0 {
			t.Errorf("Len() after PushN() = %d, want 0", got)
		}
	})

	t.Run("appends onto existing elements", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.PushN(2, 3)

		if got, want := s.Slice(), []int{3, 2, 1}; !equalInts(got, want) {
			t.Errorf("Slice() = %v, want %v", got, want)
		}
	})
}

func TestStack_Clone(t *testing.T) {
	t.Run("copies all elements in order", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)
		s.Push(3)

		c := s.Clone()

		if got, want := c.Slice(), []int{3, 2, 1}; !equalInts(got, want) {
			t.Errorf("Clone().Slice() = %v, want %v", got, want)
		}
		if got := c.Len(); got != 3 {
			t.Errorf("Clone().Len() = %d, want 3", got)
		}
	})

	t.Run("is independent of the original", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)

		c := s.Clone()
		c.Push(99) // mutating the clone...
		s.Pop()    // ...and the original...

		// ...must not affect each other
		if got := c.Len(); got != 3 {
			t.Errorf("clone Len() = %d, want 3 (independent of original)", got)
		}
		if got := s.Len(); got != 1 {
			t.Errorf("original Len() = %d, want 1 (independent of clone)", got)
		}
	})

	t.Run("clone of empty stack is empty and usable", func(t *testing.T) {
		s := New[int]()

		c := s.Clone()

		if got := c.Len(); got != 0 {
			t.Errorf("Clone().Len() = %d, want 0", got)
		}
		c.Push(1)
		if v, ok := c.Pop(); !ok || v != 1 {
			t.Errorf("Pop() after push on cloned empty = (%d, %v), want (1, true)", v, ok)
		}
	})
}

func TestStack_Grow(t *testing.T) {
	t.Run("reserves capacity without changing length or contents", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)

		s.Grow(100)

		if got := s.Cap(); got < 102 {
			t.Errorf("Cap() after Grow(100) = %d, want >= 102", got)
		}
		if got := s.Len(); got != 2 {
			t.Errorf("Len() after Grow(100) = %d, want 2", got)
		}
		if got, want := s.Slice(), []int{2, 1}; !equalInts(got, want) {
			t.Errorf("Slice() after Grow = %v, want %v", got, want)
		}
	})

	t.Run("no reallocation while pushing within grown capacity", func(t *testing.T) {
		s := New[int]()
		s.Grow(8)
		capAfterGrow := s.Cap()

		for i := range 8 {
			s.Push(i)
		}

		if got := s.Cap(); got != capAfterGrow {
			t.Errorf("Cap() after pushing within grown cap = %d, want %d (no realloc)", got, capAfterGrow)
		}
	})

	t.Run("non-positive n is a no-op", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		capBefore := s.Cap()

		s.Grow(0)
		s.Grow(-5)

		if got := s.Cap(); got != capBefore {
			t.Errorf("Cap() after Grow(0)/Grow(-5) = %d, want %d (no-op)", got, capBefore)
		}
	})
}

func TestStack_Slice(t *testing.T) {
	t.Run("returns elements top to bottom (LIFO)", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)
		s.Push(3)

		if got, want := s.Slice(), []int{3, 2, 1}; !equalInts(got, want) {
			t.Errorf("Slice() = %v, want %v", got, want)
		}
	})

	t.Run("returns an independent copy", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)

		sl := s.Slice()
		for i := range sl {
			sl[i] = 999
		}

		for _, want := range []int{2, 1} {
			if got, ok := s.Pop(); !ok || got != want {
				t.Errorf("Pop() = (%d, %v) after mutating Slice(), want (%d, true)", got, ok, want)
			}
		}
	})

	t.Run("does not consume the stack", func(t *testing.T) {
		s := New[int]()
		s.Push(1)
		s.Push(2)

		_ = s.Slice()

		if got := s.Len(); got != 2 {
			t.Errorf("Len() after Slice() = %d, want 2 (must not consume)", got)
		}
	})

	t.Run("empty stack returns nil regardless of how it became empty", func(t *testing.T) {
		drained := NewWithCap[int](8)
		drained.Push(1)
		drained.Pop()

		reset := NewWithCap[int](8)
		reset.Push(1)
		reset.Reset()

		cases := map[string]*Stack[int]{
			"fresh New":       New[int](),
			"preallocated":    NewWithCap[int](8),
			"drained to zero": drained,
			"after Reset":     reset,
		}
		for name, s := range cases {
			if got := s.Slice(); got != nil {
				t.Errorf("%s: Slice() = %v, want nil", name, got)
			}
		}
	})
}

func TestStack_NilReceiver(t *testing.T) {
	// A nil *Stack[T] must behave as a valid empty stack for every read-only
	// method, without panicking.
	t.Run("Len is 0", func(t *testing.T) {
		var s *Stack[int]
		if got := s.Len(); got != 0 {
			t.Errorf("nil.Len() = %d, want 0", got)
		}
	})

	t.Run("Cap is 0", func(t *testing.T) {
		var s *Stack[int]
		if got := s.Cap(); got != 0 {
			t.Errorf("nil.Cap() = %d, want 0", got)
		}
	})

	t.Run("IsEmpty is true", func(t *testing.T) {
		var s *Stack[int]
		if !s.IsEmpty() {
			t.Errorf("nil.IsEmpty() = false, want true")
		}
	})

	t.Run("Peek returns zero, false", func(t *testing.T) {
		var s *Stack[int]
		if got, ok := s.Peek(); ok || got != 0 {
			t.Errorf("nil.Peek() = (%d, %v), want (0, false)", got, ok)
		}
	})

	t.Run("Pop returns zero, false", func(t *testing.T) {
		var s *Stack[int]
		if got, ok := s.Pop(); ok || got != 0 {
			t.Errorf("nil.Pop() = (%d, %v), want (0, false)", got, ok)
		}
	})

	t.Run("Slice is empty", func(t *testing.T) {
		var s *Stack[int]
		if got := s.Slice(); len(got) != 0 {
			t.Errorf("nil.Slice() = %v, want no elements", got)
		}
	})

	t.Run("All yields nothing", func(t *testing.T) {
		var s *Stack[int]
		count := 0
		for range s.All() {
			count++
		}
		if count != 0 {
			t.Errorf("nil.All() yielded %d elements, want 0", count)
		}
	})

	t.Run("Clone returns a usable empty stack", func(t *testing.T) {
		var s *Stack[int]

		c := s.Clone()
		if c == nil {
			t.Fatalf("nil.Clone() = nil, want a usable empty stack")
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

func TestStack_NilReceiver_MutatorsPanic(t *testing.T) {
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
		call func(s *Stack[int])
	}{
		{name: "Push", call: func(s *Stack[int]) { s.Push(1) }},
		{name: "PushN", call: func(s *Stack[int]) { s.PushN(1, 2) }},
		{name: "PushN with no arguments", call: func(s *Stack[int]) { s.PushN() }},
		{name: "Grow", call: func(s *Stack[int]) { s.Grow(8) }},
		{name: "Grow zero", call: func(s *Stack[int]) { s.Grow(0) }},
		{name: "Grow negative", call: func(s *Stack[int]) { s.Grow(-1) }},
		{name: "Clear", call: func(s *Stack[int]) { s.Clear() }},
		{name: "Reset", call: func(s *Stack[int]) { s.Reset() }},
		{name: "Shrink", call: func(s *Stack[int]) { s.Shrink() }},
		{name: "Clip", call: func(s *Stack[int]) { s.Clip() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *Stack[int]
			assertPanics(t, tt.name, func() { tt.call(s) })
		})
	}
}

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
