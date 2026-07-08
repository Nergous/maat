// Package set provides a generic, map-backed set of unique comparable values.
//
// Create a set with [New], [NewWithCap], [Of], or [From]. Values are inserted
// with [Set.Add] and [Set.AddN], removed with [Set.Remove], and queried with
// [Set.Contains], [Set.ContainsAll], and [Set.ContainsAny].
//
// Add, Remove, Contains, Len, and IsEmpty run in expected O(1) time. Operations
// that iterate over all elements, such as Slice, Clone, and set algebra, run in
// O(n) or O(n+m), depending on the number of input elements.
//
// Iteration order is unspecified. The set is backed by a Go map, so [Set.All]
// and [Set.Slice] may return values in a different order between calls. Sort the
// result outside the package when stable output is needed.
//
// A nil *Set[T] behaves as an empty set for read-only methods, predicates, and
// empty-removal methods. Methods that need to add or clear receiver state
// ([Set.Add], [Set.AddN], [Set.AddSet], [Set.Clear], and [Set.Reset]) require a
// non-nil receiver and panic on a nil one. The zero value of Set[T] is usable:
// Add and AddN allocate the backing map on first insertion.
//
// A Set is not safe for concurrent use: it performs no internal locking. Callers
// that share a set across goroutines must provide their own synchronization.
package set

import "iter"

// Set is a collection of unique comparable values backed by a Go map.
//
// The zero value is a usable empty set. A nil *Set[T] is treated as an empty set
// by read-only methods and predicates.
type Set[T comparable] struct {
	m map[T]struct{}
}

// New returns an empty set.
func New[T comparable]() *Set[T] {
	return &Set[T]{
		m: make(map[T]struct{}),
	}
}

// NewWithCap returns an empty set with a map allocated for approximately n
// elements. A negative n is clamped to 0.
func NewWithCap[T comparable](n int) *Set[T] {
	if n < 0 {
		n = 0
	}
	return &Set[T]{
		m: make(map[T]struct{}, n),
	}
}

// Of returns a set containing vs. Duplicate values are stored once.
func Of[T comparable](vs ...T) *Set[T] {
	set := NewWithCap[T](len(vs))
	for _, v := range vs {
		set.m[v] = struct{}{}
	}

	return set
}

// From returns a set containing the values in s. Duplicate values are stored
// once.
func From[T comparable](s []T) *Set[T] {
	return Of(s...)
}

// Len returns the number of values in the set. A nil receiver reports 0.
func (s *Set[T]) Len() int {
	if s == nil {
		return 0
	}
	return len(s.m)
}

// IsEmpty reports whether the set has no values. A nil receiver reports true.
func (s *Set[T]) IsEmpty() bool {
	return s == nil || len(s.m) == 0
}

// Contains reports whether v is present in the set. A nil receiver reports
// false.
func (s *Set[T]) Contains(v T) bool {
	if s == nil {
		return false
	}

	_, ok := s.m[v]
	return ok
}

// ContainsAll reports whether every value in vs is present in the set. With no
// arguments it reports true. A nil receiver is treated as an empty set.
func (s *Set[T]) ContainsAll(vs ...T) bool {
	if len(vs) == 0 {
		return true
	}

	if s == nil {
		return false
	}

	for _, v := range vs {
		if !s.Contains(v) {
			return false
		}
	}

	return true
}

// ContainsAny reports whether at least one value in vs is present in the set.
// With no arguments it reports false. A nil receiver is treated as an empty set.
func (s *Set[T]) ContainsAny(vs ...T) bool {
	if len(vs) == 0 {
		return false
	}

	for _, v := range vs {
		if s.Contains(v) {
			return true
		}
	}

	return false
}

// All returns an iterator over the set's values. The order is unspecified and
// not stable between calls. The set is left unchanged. A nil receiver is treated
// as empty and yields no values.
func (s *Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		if s == nil {
			return
		}

		for v := range s.m {
			if !yield(v) {
				return
			}
		}
	}
}

// Slice returns a copy of the set's values. The order is unspecified and not
// stable between calls. The returned slice is independent of the set. An empty
// or nil set returns nil.
func (s *Set[T]) Slice() []T {
	if s == nil {
		return nil
	}

	if len(s.m) == 0 {
		return nil
	}

	sl := make([]T, 0, len(s.m))
	for v := range s.m {
		sl = append(sl, v)
	}
	return sl
}

// Clone returns an independent shallow copy of the set. The two sets share no
// backing map, so mutating one does not affect the other. Values are copied
// as-is: if T is a pointer or contains references, the pointed-to data is shared
// between the copies.
//
// A nil receiver clones to a new, usable empty set.
func (s *Set[T]) Clone() *Set[T] {
	if s == nil {
		return New[T]()
	}

	new := New[T]()
	for v := range s.m {
		new.m[v] = struct{}{}
	}
	return new
}

// Add inserts v into the set and reports whether v was newly added. If v was
// already present, Add leaves the set unchanged and returns false.
//
// Add requires a non-nil receiver and panics on a nil one. A zero-value Set is
// usable: Add allocates its backing map on first insertion.
func (s *Set[T]) Add(v T) bool {
	if s == nil {
		panic("set: Add on nil receiver")
	}
	if s.m == nil {
		s.m = make(map[T]struct{})
	}

	_, ok := s.m[v]
	s.m[v] = struct{}{}
	return !ok
}

// AddN inserts all values in vs. Duplicate values are stored once. With no
// arguments it is a no-op.
//
// AddN requires a non-nil receiver and panics on a nil one, even when called
// with no arguments.
func (s *Set[T]) AddN(vs ...T) {
	if s == nil {
		panic("set: AddN on nil receiver")
	}

	if len(vs) == 0 {
		return
	}
	if s.m == nil {
		s.m = make(map[T]struct{}, len(vs))
	}

	for _, v := range vs {
		s.m[v] = struct{}{}
	}
}

// Remove deletes v from the set and reports whether v was present. Removing
// from an empty or nil set is not an error and returns false.
func (s *Set[T]) Remove(v T) bool {
	if s == nil {
		return false
	}

	_, ok := s.m[v]
	delete(s.m, v)
	return ok
}

// AddSet inserts every value from other into s. A nil other set is treated as
// empty and makes AddSet a no-op.
//
// AddSet requires a non-nil receiver and panics on a nil one.
func (s *Set[T]) AddSet(other *Set[T]) {
	if s == nil {
		panic("set: AddSet on nil receiver")
	}

	if other == nil {
		return
	}
	if s.m == nil {
		s.m = make(map[T]struct{}, len(other.m))
	}

	for v := range other.m {
		s.m[v] = struct{}{}
	}
}

// RemoveSet removes every value in other from s. A nil receiver or nil other
// set is treated as empty and makes RemoveSet a no-op.
func (s *Set[T]) RemoveSet(other *Set[T]) {
	if s == nil || other == nil {
		return
	}

	for v := range other.m {
		delete(s.m, v)
	}
}

// RetainAll keeps only values that are also present in other. It is the
// in-place form of intersection. A nil other set is treated as empty and clears
// s. A nil receiver is treated as empty and makes RetainAll a no-op.
func (s *Set[T]) RetainAll(other *Set[T]) {
	if s == nil {
		return
	}

	if other == nil {
		clear(s.m)
		return
	}

	for v := range s.m {
		if !other.Contains(v) {
			delete(s.m, v)
		}
	}
}

// Clear removes all values and releases the backing map, so its memory becomes
// eligible for garbage collection. To empty the set while keeping map buckets
// for reuse, use [Set.Reset] instead.
//
// Clear requires a non-nil receiver and panics on a nil one.
func (s *Set[T]) Clear() {
	if s == nil {
		panic("set: Clear on nil receiver")
	}

	s.m = nil
}

// Reset removes all values but keeps the allocated backing map for reuse. To
// release the backing map, use [Set.Clear] instead.
//
// Reset requires a non-nil receiver and panics on a nil one.
func (s *Set[T]) Reset() {
	if s == nil {
		panic("set: Reset on nil receiver")
	}

	clear(s.m)
}

// Union returns a new set containing all values from s and others. Nil operands
// are treated as empty sets.
func (s *Set[T]) Union(others ...*Set[T]) *Set[T] {
	res := s.Clone()
	for _, other := range others {
		if other == nil {
			continue
		}
		for v := range other.m {
			res.m[v] = struct{}{}
		}
	}

	return res
}

// Intersection returns a new set containing values present in s and every set
// in others. Nil operands are treated as empty sets.
func (s *Set[T]) Intersection(others ...*Set[T]) *Set[T] {
	res := s.Clone()
	for _, other := range others {
		for v := range res.m {
			if !other.Contains(v) {
				delete(res.m, v)
			}
		}
	}

	return res
}

// Difference returns a new set containing values present in s and absent from
// every set in others. Nil operands are treated as empty sets.
func (s *Set[T]) Difference(others ...*Set[T]) *Set[T] {
	res := s.Clone()
	for _, other := range others {
		for v := range res.m {
			if other.Contains(v) {
				delete(res.m, v)
			}
		}
	}

	return res
}

// SymmetricDifference returns a new set containing values that are present in
// exactly one of s or other. A nil operand is treated as an empty set.
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	res := s.Clone()
	if other == nil {
		return res
	}

	for v := range other.m {
		if res.Contains(v) {
			delete(res.m, v)
		} else {
			res.m[v] = struct{}{}
		}
	}

	return res
}

// Equal reports whether s and other contain the same values. Nil sets are
// treated as empty sets.
func (s *Set[T]) Equal(other *Set[T]) bool {
	if s.Len() != other.Len() {
		return false
	}
	if s == nil {
		return true
	}
	for k := range s.m {
		if !other.Contains(k) {
			return false
		}
	}
	return true
}

// IsSubset reports whether every value in s is present in other. Nil sets are
// treated as empty sets.
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	if s.Len() > other.Len() {
		return false
	}
	if s == nil {
		return true
	}
	for k := range s.m {
		if !other.Contains(k) {
			return false
		}
	}
	return true
}

// IsSuperset reports whether every value in other is present in s. Nil sets are
// treated as empty sets.
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	return other.IsSubset(s)
}

// IsDisjoint reports whether s and other have no values in common. Nil sets are
// treated as empty sets.
func (s *Set[T]) IsDisjoint(other *Set[T]) bool {
	if s == nil || other == nil {
		return true
	}

	for k := range s.m {
		if other.Contains(k) {
			return false
		}
	}
	return true
}
