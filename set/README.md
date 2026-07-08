# set

A generic, map-backed set of unique comparable values for Go.

It is part of the `maat` data-structures library. The set is small,
dependency-free, and follows the same package conventions as `stack` and
`queue`: nil-safe reads, explicit memory control, non-consuming iteration, and
standard Go boolean contracts.

## Install

```go
import "github.com/Nergous/maat/set"
```

Requires Go 1.23+ (for range-over-func). The module targets Go 1.26.

## Usage

```go
package main

import (
	"fmt"
	"slices"

	"github.com/Nergous/maat/set"
)

func main() {
	s := set.Of("go", "maat", "go")

	fmt.Println(s.Contains("maat"))

	out := s.Slice()
	slices.Sort(out) // set iteration order is unspecified
	fmt.Println(out)
}
```

Preallocate when the expected size is known up front:

```go
s := set.NewWithCap[int](1000)
for i := range 1000 {
	s.Add(i)
}
fmt.Println(s.Len())
```

## API reference

| Method / Function       | Signature                                              | Description                                                                  | Complexity        |
| ----------------------- | ------------------------------------------------------ | ---------------------------------------------------------------------------- | ----------------- |
| `New`                   | `func New[T comparable]() *Set[T]`                     | Creates an empty set.                                                        | O(1)              |
| `NewWithCap`            | `func NewWithCap[T comparable](n int) *Set[T]`         | Creates an empty set with a map allocated for approximately `n` elements.     | O(n)              |
| `Of`                    | `func Of[T comparable](vs ...T) *Set[T]`               | Builds a set from explicit values, de-duplicating them.                      | O(n)              |
| `From`                  | `func From[T comparable](s []T) *Set[T]`               | Builds a set from a slice, de-duplicating values.                            | O(n)              |
| `Len`                   | `func (s *Set[T]) Len() int`                           | Number of values in the set.                                                 | O(1)              |
| `IsEmpty`               | `func (s *Set[T]) IsEmpty() bool`                      | Reports whether the set has no values.                                       | O(1)              |
| `Contains`              | `func (s *Set[T]) Contains(v T) bool`                  | Reports whether `v` is present.                                              | O(1) expected     |
| `ContainsAll`           | `func (s *Set[T]) ContainsAll(vs ...T) bool`           | Reports whether every value in `vs` is present.                              | O(k) expected     |
| `ContainsAny`           | `func (s *Set[T]) ContainsAny(vs ...T) bool`           | Reports whether at least one value in `vs` is present.                       | O(k) expected     |
| `All`                   | `func (s *Set[T]) All() iter.Seq[T]`                   | Iterator over the set's values in unspecified order.                         | O(1) per step     |
| `Slice`                 | `func (s *Set[T]) Slice() []T`                         | Returns an independent slice of values in unspecified order; `nil` if empty. | O(n)              |
| `Clone`                 | `func (s *Set[T]) Clone() *Set[T]`                     | Returns an independent shallow copy.                                         | O(n)              |
| `Add`                   | `func (s *Set[T]) Add(v T) bool`                       | Inserts `v`; returns true when it was newly added.                            | O(1) expected     |
| `AddN`                  | `func (s *Set[T]) AddN(vs ...T)`                       | Inserts all values in `vs`, de-duplicating them.                             | O(k) expected     |
| `Remove`                | `func (s *Set[T]) Remove(v T) bool`                    | Deletes `v`; returns true when it was present.                                | O(1) expected     |
| `AddSet`                | `func (s *Set[T]) AddSet(other *Set[T])`               | In-place union with `other`.                                                 | O(n) expected     |
| `RemoveSet`             | `func (s *Set[T]) RemoveSet(other *Set[T])`            | In-place difference with `other`.                                            | O(n) expected     |
| `RetainAll`             | `func (s *Set[T]) RetainAll(other *Set[T])`            | In-place intersection with `other`.                                          | O(n) expected     |
| `Clear`                 | `func (s *Set[T]) Clear()`                             | Removes all values and releases the backing map.                             | O(1)              |
| `Reset`                 | `func (s *Set[T]) Reset()`                             | Removes all values but keeps the backing map for reuse.                      | O(n)              |
| `Union`                 | `func (s *Set[T]) Union(others ...*Set[T]) *Set[T]`    | Returns a new set containing values from all operands.                       | O(n+k) expected   |
| `Intersection`          | `func (s *Set[T]) Intersection(others ...*Set[T]) *Set[T]` | Returns a new set containing values present in every operand.            | O(n*k) expected   |
| `Difference`            | `func (s *Set[T]) Difference(others ...*Set[T]) *Set[T]` | Returns a new set with values from `s` absent from all `others`.           | O(n*k) expected   |
| `SymmetricDifference`   | `func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T]` | Returns values present in exactly one operand.                         | O(n+k) expected   |
| `Equal`                 | `func (s *Set[T]) Equal(other *Set[T]) bool`           | Reports whether two sets contain the same values.                            | O(n) expected     |
| `IsSubset`              | `func (s *Set[T]) IsSubset(other *Set[T]) bool`        | Reports whether every value in `s` is present in `other`.                    | O(n) expected     |
| `IsSuperset`            | `func (s *Set[T]) IsSuperset(other *Set[T]) bool`      | Reports whether every value in `other` is present in `s`.                    | O(n) expected     |
| `IsDisjoint`            | `func (s *Set[T]) IsDisjoint(other *Set[T]) bool`      | Reports whether two sets have no values in common.                           | O(n) expected     |

### Boolean contracts

`Add` and `Remove` return whether the set changed:

```go
s := set.New[int]()
fmt.Println(s.Add(1))    // true
fmt.Println(s.Add(1))    // false, already present
fmt.Println(s.Remove(1)) // true
fmt.Println(s.Remove(1)) // false, already absent
```

`ContainsAll()` with no arguments returns true; `ContainsAny()` with no
arguments returns false.

### Unordered iteration

`All` and `Slice` follow Go map iteration order. That order is unspecified and
may change between calls:

```go
out := s.Slice()
slices.Sort(out) // caller chooses ordering when needed
fmt.Println(out)
```

### Set algebra

Non-mutating algebra returns fresh sets:

```go
a := set.Of(1, 2, 3)
b := set.Of(3, 4)

fmt.Println(a.Union(b).Len())                // 4
fmt.Println(a.Intersection(b).Contains(3))   // true
fmt.Println(a.Difference(b).Contains(1))     // true
fmt.Println(a.SymmetricDifference(b).Len())  // 3
```

For hot paths, `AddSet`, `RemoveSet`, and `RetainAll` mutate the receiver in
place.

## Capacity and memory

Go maps do not expose readable capacity, so `set` intentionally has no `Cap`,
`Grow`, `Shrink`, or `Clip`. Use `NewWithCap(n)` as a construction-time size
hint when the expected size is known.

## `Clear` vs `Reset`

Both empty the set, but they differ in what happens to the backing map:

| Operation | Empties set | Backing map                 | Use when                                               |
| --------- | ----------- | --------------------------- | ------------------------------------------------------ |
| `Clear`   | yes         | released (eligible for GC)  | You are done with the set, or want to free its memory. |
| `Reset`   | yes         | retained for reuse          | You will refill the set and want to avoid reallocating. |

`Reset` uses Go's built-in `clear(m)` behavior to keep the allocated map buckets
for reuse.

## Nil receiver and zero value

A nil `*Set[T]` behaves as a valid empty set for read-only methods, predicates,
and empty-removal methods:

```go
var s *set.Set[int]

fmt.Println(s.Len(), s.IsEmpty()) // 0 true
fmt.Println(s.Contains(1))        // false
fmt.Println(s.Remove(1))          // false
fmt.Println(s.Equal(set.New[int]())) // true
```

`Clone` on a nil receiver returns a new, usable empty set.

Methods that add or clear receiver state (`Add`, `AddN`, `AddSet`, `Clear`,
`Reset`) require a non-nil receiver and panic on a nil one. The zero value of
`Set[T]` is usable:

```go
var s set.Set[int]
s.Add(1)
fmt.Println(s.Contains(1)) // true
```

## Comparable values

`Set[T]` requires `T comparable` because values are stored as Go map keys. When
`T` is an interface type, inserting a dynamic value that is not comparable (for
example, a slice, map, or function) panics just like assigning that value as a
map key.

Float NaN values behave like Go map keys: because `NaN != NaN`, a NaN value may
not be found or removed by a later lookup.

## Concurrency

A `Set` is not safe for concurrent use: it performs no internal locking. If a
set is shared across goroutines, the caller must provide its own synchronization
(for example, a `sync.Mutex`).

## More examples and docs

Runnable, verified examples live in [`example_test.go`](./example_test.go) and
are rendered alongside the API on the godoc page.

Benchmarks live in [`bench_test.go`](./bench_test.go). Run them with:

```sh
go test -bench=. -benchmem ./set/...
```

View the documentation locally:

```sh
go doc github.com/Nergous/maat/set          # package overview
go doc github.com/Nergous/maat/set Set      # the Set type and its methods
```
