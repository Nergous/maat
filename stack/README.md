# stack

A generic, slice-backed **LIFO** (last-in, first-out) stack for Go.

It is part of the `maat` data-structures library. The stack is small,
allocation-friendly, and uses Go 1.23 range-over-func iterators
([`iter.Seq`](https://pkg.go.dev/iter#Seq)) for non-consuming iteration.

## Install

```go
import "github.com/Nergous/maat/stack"
```

Requires Go 1.23+ (for range-over-func). The module targets Go 1.26.

## Usage

```go
package main

import (
	"fmt"

	"github.com/Nergous/maat/stack"
)

func main() {
	s := stack.New[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	for !s.IsEmpty() {
		v, _ := s.Pop()
		fmt.Println(v) // 3, 2, 1
	}
}
```

Preallocate when the maximum size is known up front to avoid reallocations
while pushing:

```go
s := stack.NewWithCap[string](3)
s.Push("a")
s.Push("b")
s.Push("c")
fmt.Println(s.Len(), s.Cap()) // 3 3
```

## API reference

| Method / Function       | Signature                              | Description                                                                                         | Complexity      |
| ----------------------- | -------------------------------------- | --------------------------------------------------------------------------------------------------- | --------------- |
| `New`                   | `func New[T any]() *Stack[T]`           | Creates an empty stack with no preallocated capacity.                                                | O(1)            |
| `NewWithCap`            | `func NewWithCap[T any](n int) *Stack[T]` | Creates an empty stack with capacity preallocated for at least `n` elements.                        | O(n)            |
| `Len`                   | `func (s *Stack[T]) Len() int`          | Number of elements currently on the stack.                                                          | O(1)            |
| `Cap`                   | `func (s *Stack[T]) Cap() int`          | Capacity of the backing array before the next reallocation.                                         | O(1)            |
| `IsEmpty`               | `func (s *Stack[T]) IsEmpty() bool`     | Reports whether the stack has no elements.                                                           | O(1)            |
| `Peek`                  | `func (s *Stack[T]) Peek() (T, bool)`   | Returns the top element without removing it; `ok` is `false` and the value is the zero value if empty. | O(1)            |
| `All`                   | `func (s *Stack[T]) All() iter.Seq[T]`  | Iterator over elements from top to bottom (LIFO) without consuming them.                             | O(1) per step   |
| `Push`                  | `func (s *Stack[T]) Push(v T)`          | Adds `v` to the top of the stack.                                                                   | amortized O(1)  |
| `PushN`                 | `func (s *Stack[T]) PushN(vs ...T)`     | Adds `vs` in argument order (last argument ends up on top); grows the backing array at most once.   | amortized O(k)  |
| `Pop`                   | `func (s *Stack[T]) Pop() (T, bool)`    | Removes and returns the top element; `ok` is `false` and the value is the zero value if empty.       | O(1)            |
| `Clone`                 | `func (s *Stack[T]) Clone() *Stack[T]`  | Returns an independent shallow copy; the two stacks share no backing array.                          | O(n)            |
| `Slice`                 | `func (s *Stack[T]) Slice() []T`        | Returns a copy of the elements top to bottom (LIFO, same order as `All`); `nil` if empty.            | O(n)            |
| `Clear`                 | `func (s *Stack[T]) Clear()`            | Removes all elements **and releases** the backing array (`Cap` drops to 0).                         | O(1)            |
| `Reset`                 | `func (s *Stack[T]) Reset()`            | Removes all elements but **keeps** the backing array for reuse (`Cap` preserved).                   | O(n)            |
| `Grow`                  | `func (s *Stack[T]) Grow(n int)`        | Reserves capacity for at least `n` more elements; no-op when `n <= 0` or capacity already suffices.  | O(n)            |
| `Shrink`                | `func (s *Stack[T]) Shrink()`           | Shrink-to-fit: **copies** into a right-sized array and frees the old one now; no-op when `Cap == Len`. | O(n)          |
| `Clip`                  | `func (s *Stack[T]) Clip()`             | Reslices so `Cap == Len` **without copying**; unused memory is reclaimed only on the next growth.    | O(1)            |

### The `(T, bool)` contract

`Peek` and `Pop` return two values. The boolean reports whether the operation
found an element: it is `true` on success and `false` when the stack is empty.
When it is `false`, the first return is the zero value of `T`.

```go
v, ok := s.Pop()
if !ok {
	// stack was empty; v is the zero value
}
```

### Non-consuming iteration with `All`

`All` returns an [`iter.Seq[T]`](https://pkg.go.dev/iter#Seq) — a Go 1.23
range-over-func iterator — that yields elements from **top to bottom** (LIFO
order). It does not modify the stack, and breaking out of the loop early stops
iteration cleanly:

```go
for v := range s.All() {
	fmt.Println(v)
	if v == target {
		break // safe; the stack is untouched
	}
}
fmt.Println(s.Len()) // unchanged
```

### Bulk push with `PushN`

`PushN` adds several elements at once **in argument order**, so the last
argument ends up on top — exactly as if you called `Push` for each in turn,
but it grows the backing array at most once instead of on every element:

```go
s := stack.New[int]()
s.PushN(1, 2, 3) // 3 is now on top, same as Push(1); Push(2); Push(3)

top, _ := s.Peek()
fmt.Println(top) // 3
```

With no arguments it is a no-op.

### Copying out with `Clone` and `Slice`

`Clone` returns an independent stack; `Slice` returns a plain slice. Both are
detached from the original — they share no backing array, so mutating one does
not affect the other:

```go
a := stack.New[int]()
a.PushN(1, 2, 3)

b := a.Clone() // independent *Stack[int]
b.Pop()        // does not touch a
fmt.Println(a.Len(), b.Len()) // 3 2

out := a.Slice() // []int{3, 2, 1} — top to bottom, like All
fmt.Println(out)
```

Both copies are **shallow**: when `T` is a pointer or contains references, the
pointed-to data is shared between the copies. `Slice` returns `nil` for an
empty stack.

## Capacity management

The stack grows automatically, but several methods let you control the backing
array explicitly:

- **Reserve ahead** — `NewWithCap(n)` preallocates at construction; `Grow(n)`
  reserves room for `n` more elements on an existing stack so subsequent pushes
  do not reallocate.
- **Reclaim memory** — `Shrink` and `Clip` shrink the capacity to the current
  length (see below); `Clear` releases the backing array entirely.
- **Reuse memory** — `Reset` empties the stack but keeps the capacity for refilling.

## `Shrink` vs `Clip`

Both reduce the capacity to the current length, but they make opposite
trade-offs between work done now and memory reclaimed now:

| Operation | Copies?            | Cost | Frees unused memory                  | Use when                                                       |
| --------- | ------------------ | ---- | ------------------------------------ | -------------------------------------------------------------- |
| `Shrink`  | yes (new array)    | O(n) | immediately                          | The stack grew large, has shrunk, and you want the memory back now. |
| `Clip`    | no (reslice)       | O(1) | only on the next growth (reallocation) | You want the cheap version and can wait for memory to be reclaimed lazily. |

```go
s := stack.NewWithCap[int](1024)
s.PushN(1, 2, 3)

s.Clip()
fmt.Println(s.Len(), s.Cap()) // 3 3  — O(1) reslice, old array still retained

s.Grow(2000)
s.Shrink()
fmt.Println(s.Len(), s.Cap()) // 3 3  — copied into a right-sized array, old one freed
```

Both are no-ops when `Cap` already equals `Len`.

## `Clear` vs `Reset`

Both empty the stack and drop all references to the removed elements (so they
become eligible for garbage collection), but they differ in what happens to the
allocated backing array:

| Operation | Empties stack | Backing array        | `Cap` afterwards | Use when                                                  |
| --------- | ------------- | -------------------- | ---------------- | --------------------------------------------------------- |
| `Clear`   | yes           | released (eligible for GC) | `0`        | You are done with the stack, or want to free its memory.  |
| `Reset`   | yes           | retained for reuse   | unchanged        | You will refill the stack and want to avoid reallocating. |

```go
s := stack.NewWithCap[int](8)
s.Push(1)
s.Push(2)

s.Reset()
fmt.Println(s.Len(), s.Cap()) // 0 8  — capacity kept for reuse

s.Clear()
fmt.Println(s.Len(), s.Cap()) // 0 0  — backing array released
```

## Nil receiver

The nil zero value of `*Stack[T]` behaves as a valid **empty** stack for every
read-only method, so you can safely query a `nil` stack without first calling
`New`:

```go
var s *stack.Stack[int] // nil

s.Len()        // 0
s.IsEmpty()    // true
s.Cap()        // 0
v, ok := s.Peek() // 0, false
v, ok = s.Pop()   // 0, false (nothing to remove)
s.Slice()      // nil
for range s.All() { /* never runs */ }
```

`Clone` on a nil receiver returns a **new, usable empty stack** (never `nil`),
so the result can be pushed to immediately.

The **mutating** methods — `Push`, `PushN`, `Grow`, `Clear`, `Reset`, `Shrink`
and `Clip` — need a non-nil receiver to store their results, so they **panic**
on a `nil` stack. Create the stack with `New` or `NewWithCap` before mutating it.

## Concurrency

A `Stack` is **not safe for concurrent use**: it performs no internal locking.
If a stack is shared across goroutines, the caller must provide its own
synchronization (for example, a `sync.Mutex`).

## More examples and docs

Runnable, verified examples live in
[`example_test.go`](./example_test.go) — including bracket-balancing — and are
rendered alongside the API on the godoc page.

Benchmarks backing the "amortized O(1)" and "allocation-friendly" claims live in
[`bench_test.go`](./bench_test.go). Run them with:

```sh
go test -bench=. -benchmem ./stack/...
```

View the documentation locally:

```sh
go doc github.com/Nergous/maat/stack          # package overview
go doc github.com/Nergous/maat/stack Stack    # the Stack type and its methods
```

The same comments render on pkg.go.dev-style godoc.
