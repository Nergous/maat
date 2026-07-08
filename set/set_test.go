package set

import "testing"

func assertSetElements[T comparable](t *testing.T, s *Set[T], want ...T) {
	t.Helper()

	if got := s.Len(); got != len(want) {
		t.Fatalf("Len() = %d, want %d", got, len(want))
	}

	seen := make(map[T]int, len(want))
	for _, v := range want {
		seen[v]++
	}
	for v := range s.All() {
		seen[v]--
	}
	for v, count := range seen {
		if count != 0 {
			t.Fatalf("element %v count delta = %d, want 0", v, count)
		}
	}
}

func assertSliceElements[T comparable](t *testing.T, got []T, want ...T) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(%v) = %d, want %d", got, len(got), len(want))
	}

	seen := make(map[T]int, len(want))
	for _, v := range want {
		seen[v]++
	}
	for _, v := range got {
		seen[v]--
	}
	for v, count := range seen {
		if count != 0 {
			t.Fatalf("element %v count delta = %d, want 0 (got %v)", v, count, got)
		}
	}
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s did not panic, want panic", name)
		}
	}()
	fn()
}

func TestSet_New(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		s := New[int]()
		if s == nil {
			t.Fatalf("New[int]() = nil, want set")
		}
		if got := s.Len(); got != 0 {
			t.Errorf("Len() = %d, want 0", got)
		}
		if s.m == nil {
			t.Errorf("backing map is nil, want allocated map")
		}
	})

	t.Run("string", func(t *testing.T) {
		s := New[string]()
		if s == nil {
			t.Fatalf("New[string]() = nil, want set")
		}
		if !s.IsEmpty() {
			t.Errorf("IsEmpty() = false, want true")
		}
	})
}

func TestSet_NewWithCap(t *testing.T) {
	tests := []struct {
		name string
		cap  int
	}{
		{name: "zero", cap: 0},
		{name: "positive", cap: 8},
		{name: "negative clamps", cap: -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewWithCap[int](tt.cap)
			if got := s.Len(); got != 0 {
				t.Errorf("Len() = %d, want 0", got)
			}
			s.Add(1)
			assertSetElements(t, s, 1)
		})
	}
}

func TestSet_OfFrom(t *testing.T) {
	t.Run("Of de-duplicates values", func(t *testing.T) {
		s := Of(1, 2, 1, 3, 2)
		assertSetElements(t, s, 1, 2, 3)
	})

	t.Run("From builds from a slice", func(t *testing.T) {
		s := From([]string{"go", "go", "maat"})
		assertSetElements(t, s, "go", "maat")
	})
}

func TestSet_AddAddNRemoveContains(t *testing.T) {
	s := New[int]()

	if added := s.Add(1); !added {
		t.Errorf("Add(1) = false, want true")
	}
	if added := s.Add(1); added {
		t.Errorf("Add(1) duplicate = true, want false")
	}

	s.AddN(2, 3, 3)
	assertSetElements(t, s, 1, 2, 3)

	if !s.Contains(2) {
		t.Errorf("Contains(2) = false, want true")
	}
	if s.Contains(4) {
		t.Errorf("Contains(4) = true, want false")
	}

	if removed := s.Remove(2); !removed {
		t.Errorf("Remove(2) = false, want true")
	}
	if removed := s.Remove(2); removed {
		t.Errorf("Remove(2) absent = true, want false")
	}
	assertSetElements(t, s, 1, 3)
}

func TestSet_ZeroValue(t *testing.T) {
	var s Set[int]

	if !s.IsEmpty() {
		t.Errorf("zero-value IsEmpty() = false, want true")
	}
	if added := s.Add(10); !added {
		t.Errorf("zero-value Add(10) = false, want true")
	}
	s.AddN(20, 30)
	assertSetElements(t, &s, 10, 20, 30)

	s.Clear()
	if got := s.Len(); got != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", got)
	}
	s.Add(40)
	assertSetElements(t, &s, 40)
}

func TestSet_ContainsAllContainsAny(t *testing.T) {
	s := Of(1, 2, 3)

	if !s.ContainsAll() {
		t.Errorf("ContainsAll() = false, want true")
	}
	if !s.ContainsAll(1, 3) {
		t.Errorf("ContainsAll(1, 3) = false, want true")
	}
	if s.ContainsAll(1, 4) {
		t.Errorf("ContainsAll(1, 4) = true, want false")
	}

	if s.ContainsAny() {
		t.Errorf("ContainsAny() = true, want false")
	}
	if !s.ContainsAny(4, 2) {
		t.Errorf("ContainsAny(4, 2) = false, want true")
	}
	if s.ContainsAny(4, 5) {
		t.Errorf("ContainsAny(4, 5) = true, want false")
	}
}

func TestSet_All(t *testing.T) {
	s := Of(1, 2, 3)

	var got []int
	for v := range s.All() {
		got = append(got, v)
	}
	assertSliceElements(t, got, 1, 2, 3)
	if got := s.Len(); got != 3 {
		t.Errorf("Len() after All() = %d, want 3", got)
	}

	count := 0
	for range s.All() {
		count++
		break
	}
	if count != 1 {
		t.Errorf("early break count = %d, want 1", count)
	}
	if got := s.Len(); got != 3 {
		t.Errorf("Len() after early break = %d, want 3", got)
	}
}

func TestSet_Slice(t *testing.T) {
	t.Run("returns unordered independent copy", func(t *testing.T) {
		s := Of(1, 2, 3)
		sl := s.Slice()
		assertSliceElements(t, sl, 1, 2, 3)

		for i := range sl {
			sl[i] = 99
		}
		assertSetElements(t, s, 1, 2, 3)
	})

	t.Run("empty returns nil", func(t *testing.T) {
		if got := New[int]().Slice(); got != nil {
			t.Errorf("empty Slice() = %v, want nil", got)
		}
	})
}

func TestSet_Clone(t *testing.T) {
	t.Run("copies contents independently", func(t *testing.T) {
		s := Of(1, 2, 3)
		c := s.Clone()

		assertSetElements(t, c, 1, 2, 3)
		c.Add(4)
		s.Remove(1)

		assertSetElements(t, c, 1, 2, 3, 4)
		assertSetElements(t, s, 2, 3)
	})

	t.Run("nil clone returns usable empty set", func(t *testing.T) {
		var s *Set[int]
		c := s.Clone()
		if c == nil {
			t.Fatalf("nil.Clone() = nil, want usable empty set")
		}
		c.Add(1)
		assertSetElements(t, c, 1)
	})
}

func TestSet_ClearReset(t *testing.T) {
	t.Run("Clear releases map and remains reusable", func(t *testing.T) {
		s := Of(1, 2, 3)
		s.Clear()

		if got := s.Len(); got != 0 {
			t.Errorf("Len() after Clear() = %d, want 0", got)
		}
		if s.m != nil {
			t.Errorf("backing map after Clear() is non-nil, want nil")
		}
		s.Add(4)
		assertSetElements(t, s, 4)
	})

	t.Run("Reset keeps allocated map and remains reusable", func(t *testing.T) {
		s := Of(1, 2, 3)
		s.Reset()

		if got := s.Len(); got != 0 {
			t.Errorf("Len() after Reset() = %d, want 0", got)
		}
		if s.m == nil {
			t.Errorf("backing map after Reset() is nil, want retained map")
		}
		s.Add(4)
		assertSetElements(t, s, 4)
	})
}

func TestSet_Algebra(t *testing.T) {
	a := Of(1, 2, 3)
	b := Of(3, 4)
	c := Of(2, 3, 5)

	assertSetElements(t, a.Union(b, c), 1, 2, 3, 4, 5)
	assertSetElements(t, a.Intersection(c), 2, 3)
	assertSetElements(t, a.Difference(b, c), 1)
	assertSetElements(t, a.SymmetricDifference(b), 1, 2, 4)

	assertSetElements(t, a, 1, 2, 3)
	assertSetElements(t, b, 3, 4)
}

func TestSet_AlgebraNilAndAlias(t *testing.T) {
	var nilSet *Set[int]

	assertSetElements(t, nilSet.Union(Of(1, 2)), 1, 2)
	assertSetElements(t, Of(1, 2).Union(nil), 1, 2)
	assertSetElements(t, Of(1, 2).Intersection(nil))
	assertSetElements(t, Of(1, 2).Difference(nil), 1, 2)
	assertSetElements(t, nilSet.SymmetricDifference(Of(1, 2)), 1, 2)

	s := Of(1, 2)
	u := s.Union()
	u.Add(3)
	assertSetElements(t, s, 1, 2)
	assertSetElements(t, u, 1, 2, 3)

	assertSetElements(t, s.Intersection(s), 1, 2)
	assertSetElements(t, s.Difference(s))
	assertSetElements(t, s.SymmetricDifference(s))
}

func TestSet_Predicates(t *testing.T) {
	a := Of(1, 2)
	b := Of(1, 2, 3)
	c := Of(4)

	if !a.Equal(Of(2, 1)) {
		t.Errorf("Equal() = false, want true")
	}
	if a.Equal(b) {
		t.Errorf("Equal() with different sizes = true, want false")
	}
	if !a.IsSubset(b) {
		t.Errorf("IsSubset() = false, want true")
	}
	if !b.IsSuperset(a) {
		t.Errorf("IsSuperset() = false, want true")
	}
	if !a.IsDisjoint(c) {
		t.Errorf("IsDisjoint() = false, want true")
	}
	if a.IsDisjoint(b) {
		t.Errorf("IsDisjoint() overlapping = true, want false")
	}
}

func TestSet_PredicatesNilAsEmpty(t *testing.T) {
	var nilSet *Set[int]
	empty := New[int]()
	nonEmpty := Of(1)

	if !nilSet.Equal(empty) {
		t.Errorf("nil.Equal(empty) = false, want true")
	}
	if !nilSet.IsSubset(nonEmpty) {
		t.Errorf("nil.IsSubset(nonEmpty) = false, want true")
	}
	if !nonEmpty.IsSuperset(nilSet) {
		t.Errorf("nonEmpty.IsSuperset(nil) = false, want true")
	}
	if !nilSet.IsDisjoint(nonEmpty) {
		t.Errorf("nil.IsDisjoint(nonEmpty) = false, want true")
	}
}

func TestSet_InPlaceMutators(t *testing.T) {
	t.Run("AddSet", func(t *testing.T) {
		s := Of(1)
		s.AddSet(Of(2, 3))
		s.AddSet(nil)
		assertSetElements(t, s, 1, 2, 3)
	})

	t.Run("RemoveSet", func(t *testing.T) {
		s := Of(1, 2, 3)
		s.RemoveSet(Of(2, 4))
		s.RemoveSet(nil)
		assertSetElements(t, s, 1, 3)
	})

	t.Run("RetainAll", func(t *testing.T) {
		s := Of(1, 2, 3)
		s.RetainAll(Of(2, 4))
		assertSetElements(t, s, 2)

		s.RetainAll(nil)
		assertSetElements(t, s)
	})
}

func TestSet_NilReceiver(t *testing.T) {
	var s *Set[int]

	if got := s.Len(); got != 0 {
		t.Errorf("nil.Len() = %d, want 0", got)
	}
	if !s.IsEmpty() {
		t.Errorf("nil.IsEmpty() = false, want true")
	}
	if s.Contains(1) {
		t.Errorf("nil.Contains(1) = true, want false")
	}
	if !s.ContainsAll() {
		t.Errorf("nil.ContainsAll() = false, want true")
	}
	if s.ContainsAll(1) {
		t.Errorf("nil.ContainsAll(1) = true, want false")
	}
	if s.ContainsAny(1) {
		t.Errorf("nil.ContainsAny(1) = true, want false")
	}
	if s.Remove(1) {
		t.Errorf("nil.Remove(1) = true, want false")
	}
	if got := s.Slice(); got != nil {
		t.Errorf("nil.Slice() = %v, want nil", got)
	}

	count := 0
	for range s.All() {
		count++
	}
	if count != 0 {
		t.Errorf("nil.All() yielded %d elements, want 0", count)
	}

	s.RemoveSet(Of(1))
	s.RetainAll(Of(1))
}

func TestSet_NilReceiverMutatorsPanic(t *testing.T) {
	var s *Set[int]

	assertPanics(t, "Add", func() { s.Add(1) })
	assertPanics(t, "AddN", func() { s.AddN(1, 2) })
	assertPanics(t, "AddN with no arguments", func() { s.AddN() })
	assertPanics(t, "AddSet", func() { s.AddSet(Of(1)) })
	assertPanics(t, "Clear", func() { s.Clear() })
	assertPanics(t, "Reset", func() { s.Reset() })
}
