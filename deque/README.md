# deque

A generic, slice-backed ring-buffer **double-ended queue** for Go.

It is part of the `maat` data-structures library. The deque is small,
allocation-friendly, and follows the same standard-library conventions as the
rest of the collection: `(value, ok)` lookups and removals, explicit memory
control, non-consuming iteration, and safe nil/zero-value reads.

## Install

```go
import "github.com/Nergous/maat/deque"
```

Requires Go 1.23+ (for range-over-func). The module targets Go 1.26.

## Usage

```go
package main

import (
	"fmt"

	"github.com/Nergous/maat/deque"
)

func main() {
	d := deque.New[int]()
	d.PushBack(2)
	d.PushFront(1)
	d.PushBack(3)

	for !d.IsEmpty() {
		v, _ := d.PopFront()
		fmt.Println(v) // 1, 2, 3
	}
}
```

Preallocate when the maximum size is known up front to avoid reallocations
while pushing:

```go
d := deque.NewWithCap[string](3)
d.PushBack("a")
d.PushBack("b")
d.PushBack("c")
fmt.Println(d.Len(), d.Cap()) // 3 3
```

## API reference

| Method / Function | Signature                                      | Description                                                                                                  | Complexity      |
| ----------------- | ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | --------------- |
| `New`             | `func New[T any]() *Deque[T]`                  | Creates an empty deque with no preallocated capacity.                                                        | O(1)            |
| `NewWithCap`      | `func NewWithCap[T any](n int) *Deque[T]`      | Creates an empty deque with capacity preallocated for at least `n` elements; negative `n` is clamped to 0.    | O(n)            |
| `Len`             | `func (d *Deque[T]) Len() int`                 | Number of elements currently in the deque.                                                                   | O(1)            |
| `Cap`             | `func (d *Deque[T]) Cap() int`                 | Capacity of the backing ring buffer; it does not shrink when elements are popped.                            | O(1)            |
| `IsEmpty`         | `func (d *Deque[T]) IsEmpty() bool`            | Reports whether the deque has no elements.                                                                   | O(1)            |
| `Front`           | `func (d *Deque[T]) Front() (T, bool)`         | Returns the front element without removing it; `ok` is false and the value is zero if empty.                 | O(1)            |
| `Back`            | `func (d *Deque[T]) Back() (T, bool)`          | Returns the back element without removing it; `ok` is false and the value is zero if empty.                  | O(1)            |
| `All`             | `func (d *Deque[T]) All() iter.Seq[T]`         | Range-over-func iterator over elements from front to back, without consuming them.                           | O(n)            |
| `PushFront`       | `func (d *Deque[T]) PushFront(v T)`            | Adds `v` to the front of the deque.                                                                          | amortized O(1)  |
| `PushBack`        | `func (d *Deque[T]) PushBack(v T)`             | Adds `v` to the back of the deque.                                                                           | amortized O(1)  |
| `PushFrontN`      | `func (d *Deque[T]) PushFrontN(vs ...T)`       | Bulk-adds `vs` to the front in argument order, so the first argument becomes the new front.                  | amortized O(k)  |
| `PushBackN`       | `func (d *Deque[T]) PushBackN(vs ...T)`        | Bulk-adds `vs` to the back in argument order, so the last argument becomes the new back.                     | amortized O(k)  |
| `PopFront`        | `func (d *Deque[T]) PopFront() (T, bool)`      | Removes and returns the front element; `ok` is false and the value is zero if empty.                         | O(1)            |
| `PopBack`         | `func (d *Deque[T]) PopBack() (T, bool)`       | Removes and returns the back element; `ok` is false and the value is zero if empty.                          | O(1)            |
| `PopFrontN`       | `func (d *Deque[T]) PopFrontN(n int) []T`      | Removes and returns up to `n` front elements in front-to-back order; preserves capacity.                     | O(k)            |
| `PopBackN`        | `func (d *Deque[T]) PopBackN(n int) []T`       | Removes and returns up to `n` back elements in removal order (back-to-front); preserves capacity.            | O(k)            |
| `Clone`           | `func (d *Deque[T]) Clone() *Deque[T]`         | Returns an independent shallow copy sharing no backing array.                                                | O(n)            |
| `Slice`           | `func (d *Deque[T]) Slice() []T`               | Returns an independent copy of the elements from front to back; an empty deque returns `nil`.                | O(n)            |
| `Grow`            | `func (d *Deque[T]) Grow(n int)`               | Reserves capacity for at least `n` more elements; no-op when `n <= 0` or capacity already suffices.          | O(n)            |
| `Shrink`          | `func (d *Deque[T]) Shrink()`                  | Copies the elements into a new right-sized array, releasing the larger one immediately.                      | O(n)            |
| `Clip`            | `func (d *Deque[T]) Clip()`                    | Reduces reported capacity to `Len`; copies only when wrapped elements need compaction.                       | O(n) worst case |
| `Clear`           | `func (d *Deque[T]) Clear()`                   | Removes all elements and releases the backing array (`Cap` drops to 0).                                      | O(1)            |
| `Reset`           | `func (d *Deque[T]) Reset()`                   | Removes all elements but keeps the backing array for reuse (`Cap` preserved).                               | O(n)            |

Read-only methods (`Len`, `Cap`, `IsEmpty`, `Front`, `Back`, `All`, `Slice`,
`Clone`) and empty-removal methods (`PopFront`, `PopBack`, `PopFrontN`,
`PopBackN`) treat a `nil *Deque[T]` as a valid empty deque and never panic.
Methods that store, clear, reset, shrink, or resize data (`PushFront`,
`PushBack`, `PushFrontN`, `PushBackN`, `Grow`, `Shrink`, `Clip`, `Clear`,
`Reset`) require a non-nil receiver and panic on a nil one.

### The `(T, bool)` contract

`Front`, `Back`, `PopFront`, and `PopBack` return two values. The boolean
reports whether the operation found an element: it is `true` on success and
`false` when the deque is empty. When it is `false`, the first return is the
zero value of `T`.

```go
v, ok := d.PopFront()
if !ok {
	// deque was empty; v is the zero value
}
```

### Double-ended order

A deque can operate on either end:

- `PushFront` and `PopFront` work at the front.
- `PushBack` and `PopBack` work at the back.
- `Front`, `Back`, `All`, and `Slice` read the current logical order from front
  to back.

```go
d := deque.New[string]()
d.PushBack("normal")
d.PushFront("urgent")

front, _ := d.PopFront()
back, _ := d.PopBack()
fmt.Println(front, back) // urgent normal
```

### Non-consuming iteration with `All`

`All` returns an [`iter.Seq[T]`](https://pkg.go.dev/iter#Seq) range-over-func
iterator that yields elements from **front to back**. It does not modify the
deque, and breaking out of the loop early stops iteration cleanly:

```go
for v := range d.All() {
	fmt.Println(v)
	if v == target {
		break // safe; the deque is untouched
	}
}
fmt.Println(d.Len()) // unchanged
```

### Bulk pushes

`PushFrontN` and `PushBackN` add several elements at once and grow the backing
array at most once:

```go
d := deque.New[int]()
d.PushBackN(4, 5)
d.PushFrontN(1, 2, 3)

fmt.Println(d.Slice()) // [1 2 3 4 5]
```

`PushFrontN` keeps argument order at the front: the first argument becomes the
new front. `PushBackN` keeps argument order at the back: the last argument
becomes the new back. With no arguments, both are no-ops.

### Batch pops

`PopFrontN` removes up to `n` elements from the front and returns them in
front-to-back order. `PopBackN` removes up to `n` elements from the back and
returns them in removal order, which is back-to-front.

```go
d := deque.New[int]()
d.PushBackN(1, 2, 3, 4)

fmt.Println(d.PopFrontN(2)) // [1 2]
fmt.Println(d.PopBackN(2))  // [4 3]
```

If `n` is larger than the current length, the method drains the deque. If
`n <= 0`, or the deque is empty or nil, it returns `nil`. Batch pops preserve
the backing capacity.

### Copying out with `Clone` and `Slice`

`Clone` returns an independent deque; `Slice` returns a plain slice. Both are
detached from the original: they share no backing array, so mutating one does
not affect the other.

```go
a := deque.New[int]()
a.PushBackN(1, 2, 3)

b := a.Clone()
b.PopFront()
b.PushBack(4)

fmt.Println(a.Slice()) // [1 2 3]
fmt.Println(b.Slice()) // [2 3 4]
```

Both copies are shallow: when `T` is a pointer or contains references, the
pointed-to data is shared between the copies. `Slice` returns `nil` for an
empty deque.

## Capacity management

The deque grows automatically, but several methods let you control the backing
ring buffer explicitly:

- **Reserve ahead**: `NewWithCap(n)` preallocates at construction; `Grow(n)`
  reserves room for `n` more elements on an existing deque so later pushes do
  not reallocate.
- **Reclaim memory**: `Shrink` and `Clip` shrink the capacity to the current
  length; `Clear` releases the backing array entirely.
- **Reuse memory**: `Reset` empties the deque but keeps the capacity for
  refilling.
- **Stable capacity**: `PopFront`, `PopBack`, `PopFrontN`, and `PopBackN` free
  slots for reuse, but they do not reduce the value reported by `Cap`.

## `Shrink` vs `Clip`

Both reduce the capacity to the current length, but they make different
trade-offs between work done now and memory reclaimed now:

| Operation | Copies?                                      | Cost          | Frees unused memory        | Use when                                                           |
| --------- | -------------------------------------------- | ------------- | -------------------------- | ------------------------------------------------------------------ |
| `Shrink`  | yes (new array)                              | O(n)          | immediately                | The deque grew large, has shrunk, and you want the memory back now. |
| `Clip`    | only if wrapped elements require compaction  | O(n) worst case | not guaranteed immediately | You want reported capacity trimmed to `Len` and can accept lazy reclaim. |

```go
d := deque.NewWithCap[int](1024)
d.PushBackN(1, 2, 3)

d.Clip()
fmt.Println(d.Len(), d.Cap()) // 3 3

d.Grow(2000)
d.Shrink()
fmt.Println(d.Len(), d.Cap()) // 3 3
```

Both are no-ops when the deque is already tight.

## `Clear` vs `Reset`

Both empty the deque and zero the removed elements so the backing array no
longer retains references to them, but they differ in what happens to the
allocated memory:

| Operation | Empties deque | Backing array              | `Cap` afterwards | Use when                                                  |
| --------- | ------------- | -------------------------- | ---------------- | --------------------------------------------------------- |
| `Clear`   | yes           | released (eligible for GC) | `0`              | You are done with the deque, or want to free its memory.  |
| `Reset`   | yes           | retained for reuse         | unchanged        | You will refill the deque and want to avoid reallocating. |

```go
d := deque.NewWithCap[int](8)
d.PushBack(1)
d.PushBack(2)

d.Reset()
fmt.Println(d.Len(), d.Cap()) // 0 8

d.Clear()
fmt.Println(d.Len(), d.Cap()) // 0 0
```

### Ring-buffer capacity

The deque stores elements in a ring buffer. When elements are popped from either
end, their slots become available for later pushes. The buffer does not slide or
shrink on removal, so `Cap` stays stable across pops and `Reset` preserves the
full allocation after many removals.

## Nil receiver and zero value

The zero value of `Deque[T]` is usable:

```go
var d deque.Deque[int]
d.PushBack(1)
fmt.Println(d.Len()) // 1
```

A nil `*Deque[T]` behaves as a valid **empty** deque for read-only methods and
empty-removal methods:

```go
var d *deque.Deque[int] // nil

d.Len()          // 0
d.IsEmpty()      // true
d.Cap()          // 0
v, ok := d.Front()    // 0, false
v, ok = d.Back()      // 0, false
v, ok = d.PopFront()  // 0, false
v, ok = d.PopBack()   // 0, false
d.PopFrontN(3)   // nil
d.PopBackN(3)    // nil
d.Slice()        // nil
for range d.All() { /* never runs */ }
```

`Clone` on a nil receiver returns a new, usable empty deque (never `nil`), so
the result can be pushed to immediately.

Methods that store, clear, reset, shrink, or resize data need a non-nil
receiver, so they panic on a nil deque. Create the deque with `New` or
`NewWithCap`, or use a zero-value `Deque[T]` value, before calling them.

## Concurrency

A `Deque` is not safe for concurrent use: it performs no internal locking. If a
deque is shared across goroutines, the caller must provide its own
synchronization (for example, a `sync.Mutex`).

## More examples and docs

Runnable, verified examples live in [`example_test.go`](./example_test.go) and
are rendered alongside the API on the godoc page.

Benchmarks backing the "amortized O(1)" and "allocation-friendly" claims live in
[`bench_test.go`](./bench_test.go). Run them with:

```sh
go test -bench=. -benchmem ./deque/...
```

View the documentation locally:

```sh
go doc github.com/Nergous/maat/deque          # package overview
go doc github.com/Nergous/maat/deque Deque    # the Deque type and its methods
```

The same comments render on pkg.go.dev-style godoc.
