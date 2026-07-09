package deque

import "testing"

func TestDeque_PushPopBothEnds(t *testing.T) {
	d := New[int]()
	d.PushBack(2)
	d.PushFront(1)
	d.PushBack(3)

	if got, want := d.Slice(), []int{1, 2, 3}; !equalInts(got, want) {
		t.Fatalf("Slice() = %v, want %v", got, want)
	}
	if got, ok := d.Front(); !ok || got != 1 {
		t.Fatalf("Front() = (%d, %v), want (1, true)", got, ok)
	}
	if got, ok := d.Back(); !ok || got != 3 {
		t.Fatalf("Back() = (%d, %v), want (3, true)", got, ok)
	}
	if got, ok := d.PopFront(); !ok || got != 1 {
		t.Fatalf("PopFront() = (%d, %v), want (1, true)", got, ok)
	}
	if got, ok := d.PopBack(); !ok || got != 3 {
		t.Fatalf("PopBack() = (%d, %v), want (3, true)", got, ok)
	}
	if got, want := d.Slice(), []int{2}; !equalInts(got, want) {
		t.Fatalf("remaining Slice() = %v, want %v", got, want)
	}
}

func TestDeque_BulkOperations(t *testing.T) {
	d := New[int]()
	d.PushBackN(4, 5)
	d.PushFrontN(1, 2, 3)

	if got, want := d.Slice(), []int{1, 2, 3, 4, 5}; !equalInts(got, want) {
		t.Fatalf("Slice() after bulk pushes = %v, want %v", got, want)
	}
	if got, want := d.PopFrontN(2), []int{1, 2}; !equalInts(got, want) {
		t.Fatalf("PopFrontN(2) = %v, want %v", got, want)
	}
	if got, want := d.PopBackN(2), []int{5, 4}; !equalInts(got, want) {
		t.Fatalf("PopBackN(2) = %v, want %v", got, want)
	}
	if got, want := d.Slice(), []int{3}; !equalInts(got, want) {
		t.Fatalf("remaining Slice() = %v, want %v", got, want)
	}
}

func TestDeque_PopNDrainsAndHandlesEmpty(t *testing.T) {
	d := NewWithCap[int](8)
	d.PushBackN(1, 2, 3)

	if got, want := d.PopFrontN(10), []int{1, 2, 3}; !equalInts(got, want) {
		t.Fatalf("PopFrontN(10) = %v, want %v", got, want)
	}
	if got := d.Len(); got != 0 {
		t.Fatalf("Len() after draining = %d, want 0", got)
	}
	if got := d.Cap(); got != 8 {
		t.Fatalf("Cap() after draining = %d, want 8", got)
	}
	if got := d.PopBackN(1); got != nil {
		t.Fatalf("PopBackN on empty = %v, want nil", got)
	}
	if got := d.PopFrontN(0); got != nil {
		t.Fatalf("PopFrontN(0) = %v, want nil", got)
	}
	if got := d.PopBackN(-1); got != nil {
		t.Fatalf("PopBackN(-1) = %v, want nil", got)
	}
}

func TestDeque_WrapAround(t *testing.T) {
	d := NewWithCap[int](4)
	d.PushBackN(1, 2, 3, 4)
	d.PopFront()
	d.PopFront()
	d.PushBackN(5, 6)

	if got, want := d.Slice(), []int{3, 4, 5, 6}; !equalInts(got, want) {
		t.Fatalf("wrapped Slice() = %v, want %v", got, want)
	}
	if got, want := d.PopBackN(3), []int{6, 5, 4}; !equalInts(got, want) {
		t.Fatalf("wrapped PopBackN(3) = %v, want %v", got, want)
	}
	if got, want := d.Slice(), []int{3}; !equalInts(got, want) {
		t.Fatalf("remaining wrapped Slice() = %v, want %v", got, want)
	}
}

func TestDeque_AllCloneAndSlice(t *testing.T) {
	d := New[int]()
	d.PushBackN(1, 2, 3)

	var all []int
	for v := range d.All() {
		all = append(all, v)
	}
	if got, want := all, []int{1, 2, 3}; !equalInts(got, want) {
		t.Fatalf("All() = %v, want %v", got, want)
	}

	cloned := d.Clone()
	d.PopFront()
	cloned.PushBack(4)

	if got, want := d.Slice(), []int{2, 3}; !equalInts(got, want) {
		t.Fatalf("original Slice() = %v, want %v", got, want)
	}
	if got, want := cloned.Slice(), []int{1, 2, 3, 4}; !equalInts(got, want) {
		t.Fatalf("clone Slice() = %v, want %v", got, want)
	}
}

func TestDeque_MemoryControls(t *testing.T) {
	t.Run("Reset keeps capacity", func(t *testing.T) {
		d := NewWithCap[int](8)
		d.PushBackN(1, 2, 3)
		d.Reset()

		if got := d.Len(); got != 0 {
			t.Fatalf("Len() after Reset() = %d, want 0", got)
		}
		if got := d.Cap(); got != 8 {
			t.Fatalf("Cap() after Reset() = %d, want 8", got)
		}
		d.PushBack(4)
		if got, ok := d.PopFront(); !ok || got != 4 {
			t.Fatalf("PopFront() after Reset/PushBack = (%d, %v), want (4, true)", got, ok)
		}
	})

	t.Run("Clear releases capacity", func(t *testing.T) {
		d := NewWithCap[int](8)
		d.PushBackN(1, 2, 3)
		d.Clear()

		if got := d.Len(); got != 0 {
			t.Fatalf("Len() after Clear() = %d, want 0", got)
		}
		if got := d.Cap(); got != 0 {
			t.Fatalf("Cap() after Clear() = %d, want 0", got)
		}
	})

	t.Run("Shrink and Clip trim to length", func(t *testing.T) {
		d := NewWithCap[int](16)
		d.PushBackN(1, 2, 3)
		d.Shrink()
		if got := d.Cap(); got != 3 {
			t.Fatalf("Cap() after Shrink() = %d, want 3", got)
		}
		d.Grow(10)
		if got := d.Cap(); got < 13 {
			t.Fatalf("Cap() after Grow(10) = %d, want >= 13", got)
		}
		d.Clip()
		if got := d.Cap(); got != 3 {
			t.Fatalf("Cap() after Clip() = %d, want 3", got)
		}
		if got, want := d.Slice(), []int{1, 2, 3}; !equalInts(got, want) {
			t.Fatalf("Slice() after memory controls = %v, want %v", got, want)
		}
	})
}

func TestDeque_NilReceiver(t *testing.T) {
	var d *Deque[int]

	if got := d.Len(); got != 0 {
		t.Fatalf("nil.Len() = %d, want 0", got)
	}
	if got := d.Cap(); got != 0 {
		t.Fatalf("nil.Cap() = %d, want 0", got)
	}
	if !d.IsEmpty() {
		t.Fatalf("nil.IsEmpty() = false, want true")
	}
	if got, ok := d.Front(); ok || got != 0 {
		t.Fatalf("nil.Front() = (%d, %v), want (0, false)", got, ok)
	}
	if got, ok := d.Back(); ok || got != 0 {
		t.Fatalf("nil.Back() = (%d, %v), want (0, false)", got, ok)
	}
	if got, ok := d.PopFront(); ok || got != 0 {
		t.Fatalf("nil.PopFront() = (%d, %v), want (0, false)", got, ok)
	}
	if got, ok := d.PopBack(); ok || got != 0 {
		t.Fatalf("nil.PopBack() = (%d, %v), want (0, false)", got, ok)
	}
	if got := d.PopFrontN(1); got != nil {
		t.Fatalf("nil.PopFrontN(1) = %v, want nil", got)
	}
	if got := d.PopBackN(1); got != nil {
		t.Fatalf("nil.PopBackN(1) = %v, want nil", got)
	}
	if got := d.Slice(); got != nil {
		t.Fatalf("nil.Slice() = %v, want nil", got)
	}
	if got := d.Clone(); got == nil || got.Len() != 0 {
		t.Fatalf("nil.Clone() = %#v, want usable empty deque", got)
	}
}

func TestDeque_NilReceiverMutatorsPanic(t *testing.T) {
	tests := []struct {
		name string
		call func(d *Deque[int])
	}{
		{name: "PushFront", call: func(d *Deque[int]) { d.PushFront(1) }},
		{name: "PushBack", call: func(d *Deque[int]) { d.PushBack(1) }},
		{name: "PushFrontN", call: func(d *Deque[int]) { d.PushFrontN(1, 2) }},
		{name: "PushBackN", call: func(d *Deque[int]) { d.PushBackN(1, 2) }},
		{name: "Grow", call: func(d *Deque[int]) { d.Grow(1) }},
		{name: "Clear", call: func(d *Deque[int]) { d.Clear() }},
		{name: "Reset", call: func(d *Deque[int]) { d.Reset() }},
		{name: "Shrink", call: func(d *Deque[int]) { d.Shrink() }},
		{name: "Clip", call: func(d *Deque[int]) { d.Clip() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d *Deque[int]
			assertPanics(t, tt.name, func() { tt.call(d) })
		})
	}
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s did not panic, want panic", name)
		}
	}()
	fn()
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
