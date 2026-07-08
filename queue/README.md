# queue

A generic, slice-backed ring-buffer **FIFO** (first-in, first-out) queue for Go.

It is part of the `maat` data-structures library. The queue is small,
allocation-friendly, and follows the same standard-library conventions as the
rest of the collection (the `(value, ok)` contract, explicit memory control,
non-consuming iteration, and a usable nil/zero value for reads).

## Install

```go
import "github.com/Nergous/maat/queue"
```

Requires Go 1.23+. The module targets Go 1.26.

## Usage

```go
package main

import (
	"fmt"

	"github.com/Nergous/maat/queue"
)

func main() {
	q := queue.New[int]()
	q.Push(1)
	q.Push(2)
	q.Push(3)

	for !q.IsEmpty() {
		v, _ := q.Pop()
		fmt.Println(v) // 1, 2, 3
	}
}
```

Preallocate when the maximum size is known up front to avoid reallocations
while pushing:

```go
q := queue.NewWithCap[string](3)
q.Push("a")
q.Push("b")
q.Push("c")
fmt.Println(q.Len(), q.Cap()) // 3 3
```

## API reference

| Method / Function | Signature                                  | Description                                                                                             | Complexity     |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------- | -------------- |
| `New`             | `func New[T any]() *Queue[T]`              | Creates an empty queue with no preallocated capacity.                                                   | O(1)           |
| `NewWithCap`      | `func NewWithCap[T any](n int) *Queue[T]`  | Creates an empty queue with capacity preallocated for at least `n` elements (a negative `n` is clamped to 0). | O(n)           |
| `Len`             | `func (q *Queue[T]) Len() int`             | Number of elements currently in the queue.                                                              | O(1)           |
| `Cap`             | `func (q *Queue[T]) Cap() int`             | Capacity of the backing ring buffer; it does not shrink when elements are popped.                       | O(1)           |
| `IsEmpty`         | `func (q *Queue[T]) IsEmpty() bool`        | Reports whether the queue has no elements.                                                              | O(1)           |
| `Peek`            | `func (q *Queue[T]) Peek() (T, bool)`      | Returns the front element (the oldest) without removing it; `ok` is `false` and the value is the zero value if empty. | O(1)           |
| `All`             | `func (q *Queue[T]) All() iter.Seq[T]`     | Range-over-func iterator over the elements front→back (FIFO), without consuming them.                   | O(n)           |
| `Push`            | `func (q *Queue[T]) Push(v T)`             | Adds `v` to the back of the queue.                                                                      | amortized O(1) |
| `PushN`           | `func (q *Queue[T]) PushN(vs ...T)`        | Bulk-enqueues `vs` in order (last argument at the back), growing the backing array at most once.       | amortized O(k) |
| `Pop`             | `func (q *Queue[T]) Pop() (T, bool)`       | Removes and returns the front element (the oldest); `ok` is `false` and the value is the zero value if empty. | O(1)           |
| `PopN`            | `func (q *Queue[T]) PopN(n int) []T`       | Removes and returns up to `n` front elements in FIFO order; preserves capacity.                        | O(k)           |
| `Clone`           | `func (q *Queue[T]) Clone() *Queue[T]`     | Returns an independent shallow copy sharing no backing array.                                           | O(n)           |
| `Slice`           | `func (q *Queue[T]) Slice() []T`           | Returns an independent copy of the elements front→back (FIFO); an empty queue returns `nil`.            | O(n)           |
| `Grow`            | `func (q *Queue[T]) Grow(n int)`           | Reserves capacity for at least `n` more elements; no-op when `n <= 0`.                                  | O(n)           |
| `Shrink`          | `func (q *Queue[T]) Shrink()`              | Copies the elements into a new right-sized array, releasing the larger one immediately.                 | O(n)           |
| `Clip`            | `func (q *Queue[T]) Clip()`                | Reduces reported capacity to `Len`; copies only when wrapped elements need compaction.                 | O(n) worst case |
| `Clear`           | `func (q *Queue[T]) Clear()`               | Removes all elements **and releases** the backing array (`Cap` drops to 0).                            | O(1)           |
| `Reset`           | `func (q *Queue[T]) Reset()`               | Removes all elements but **keeps** the backing array for reuse (`Cap` preserved).                      | O(n)           |

Read-only methods (`Len`, `Cap`, `IsEmpty`, `Peek`, `All`, `Slice`, `Clone`)
and empty-removal methods (`Pop`, `PopN`) treat a `nil *Queue[T]` as a valid
empty queue and never panic. Methods that store or resize data (`Push`, `PushN`,
`Grow`, `Shrink`, `Clip`, `Clear`, `Reset`) require a non-nil receiver and panic
on a nil one.

### The `(T, bool)` contract

`Peek` and `Pop` return two values. The boolean reports whether the operation
found an element: it is `true` on success and `false` when the queue is empty.
When it is `false`, the first return is the zero value of `T`.

```go
v, ok := q.Pop()
if !ok {
	// queue was empty; v is the zero value
}
```

### FIFO order

A queue removes elements in the same order they were added: `Push` appends to the
back, `Pop`, `PopN`, and `Peek` operate on the front (the oldest element still
present).

```go
q := queue.New[string]()
q.Push("first")
q.Push("last")

v, _ := q.Pop()
fmt.Println(v) // first
```

### Non-consuming iteration with `All`

`All` returns an [`iter.Seq[T]`](https://pkg.go.dev/iter#Seq) — a Go 1.23
range-over-func iterator — that yields elements from **front to back** (FIFO
order). It does not modify the queue, and breaking out of the loop early stops
iteration cleanly:

```go
for v := range q.All() {
	fmt.Println(v)
	if v == target {
		break // safe; the queue is untouched
	}
}
fmt.Println(q.Len()) // unchanged
```

### Bulk push with `PushN`

`PushN` adds several elements at once **in argument order**, so the last
argument ends up at the back — exactly as if you called `Push` for each in turn,
but it grows the backing array at most once instead of on every element:

```go
q := queue.New[int]()
q.PushN(1, 2, 3) // same as Push(1); Push(2); Push(3)

front, _ := q.Peek()
fmt.Println(front) // 1
```

With no arguments it is a no-op.

### Batch pop with `PopN`

`PopN` removes up to `n` elements from the front and returns them in FIFO order.
If `n` is larger than the current length, it drains the queue. If `n <= 0`, or
the queue is empty or nil, it returns `nil`.

```go
q := queue.New[int]()
q.PushN(1, 2, 3, 4)

batch := q.PopN(3)
fmt.Println(batch) // [1 2 3]
fmt.Println(q.Slice()) // [4]
```

`PopN` preserves the backing capacity, so it works well for batch consumers that
reuse the same queue.

### Copying out with `Clone` and `Slice`

`Clone` returns an independent queue; `Slice` returns a plain slice. Both are
detached from the original — they share no backing array, so mutating one does
not affect the other:

```go
a := queue.New[int]()
a.PushN(1, 2, 3)

b := a.Clone() // independent *Queue[int]
b.Pop()        // does not touch a
fmt.Println(a.Len(), b.Len()) // 3 2

out := a.Slice() // []int{1, 2, 3} — front to back, like All
fmt.Println(out)
```

Both copies are **shallow**: when `T` is a pointer or contains references, the
pointed-to data is shared between the copies. `Slice` returns `nil` for an
empty queue.

## Capacity management

The queue grows automatically, but several methods let you control the backing
ring buffer explicitly:

- **Reserve ahead** — `NewWithCap(n)` preallocates at construction; `Grow(n)`
  reserves room for `n` more elements on an existing queue so subsequent pushes
  do not reallocate.
- **Reclaim memory** — `Shrink` and `Clip` shrink the capacity to the current
  length (see below); `Clear` releases the backing array entirely.
- **Reuse memory** — `Reset` empties the queue but keeps the capacity for refilling.
- **Stable capacity** — `Pop` and `PopN` free slots for reuse, but they do not
  reduce the value reported by `Cap`.

## `Shrink` vs `Clip`

Both reduce the capacity to the current length, but they make opposite
trade-offs between work done now and memory reclaimed now:

| Operation | Copies?                         | Cost          | Frees unused memory                  | Use when                                                       |
| --------- | ------------------------------- | ------------- | ------------------------------------ | -------------------------------------------------------------- |
| `Shrink`  | yes (new array)                 | O(n)          | immediately                          | The queue grew large, has shrunk, and you want the memory back now. |
| `Clip`    | only if wrapped elements require compaction | O(n) worst case | not guaranteed immediately           | You want reported capacity trimmed to `Len` and can accept lazy reclaim. |

```go
q := queue.NewWithCap[int](1024)
q.PushN(1, 2, 3)

q.Clip()
fmt.Println(q.Len(), q.Cap()) // 3 3  — compact reported capacity

q.Grow(2000)
q.Shrink()
fmt.Println(q.Len(), q.Cap()) // 3 3  — copied into a right-sized array, old one freed
```

Both are no-ops when `Cap` already equals `Len`.

## `Clear` vs `Reset`

Both empty the queue and zero the removed elements (so the backing array no
longer retains references to them), but they differ in what happens to the
allocated memory:

| Operation | Empties queue | Backing array              | `Cap` afterwards | Use when                                                  |
| --------- | ------------- | -------------------------- | ---------------- | --------------------------------------------------------- |
| `Clear`   | yes           | released (eligible for GC) | `0`              | You are done with the queue, or want to free its memory.  |
| `Reset`   | yes           | retained for reuse         | unchanged        | You will refill the queue and want to avoid reallocating. |

```go
q := queue.NewWithCap[int](8)
q.Push(1)
q.Push(2)

q.Reset()
fmt.Println(q.Len(), q.Cap()) // 0 8  — capacity kept for reuse

q.Clear()
fmt.Println(q.Len(), q.Cap()) // 0 0  — backing array released
```

### Ring-buffer capacity

The queue stores elements in a ring buffer. When elements are popped, their
slots become available for later pushes; the buffer does not have to slide or
shrink. This keeps `Cap` stable across `Pop`/`PopN` and lets `Reset` preserve the
full allocation even after many removals.

## Nil receiver

The nil zero value of `*Queue[T]` behaves as a valid **empty** queue for every
read-only method, so you can safely query a `nil` queue without first calling
`New`:

```go
var q *queue.Queue[int] // nil

q.Len()        // 0
q.IsEmpty()    // true
q.Cap()        // 0
v, ok := q.Peek() // 0, false
v, ok = q.Pop()   // 0, false (nothing to remove)
q.PopN(3)      // nil
q.Slice()      // nil
for range q.All() { /* never runs */ }
```

`Clone` on a nil receiver returns a **new, usable empty queue** (never `nil`),
so the result can be pushed to immediately.

Methods that store or resize data — `Push`, `PushN`, `Grow`, `Clear`, `Reset`,
`Shrink` and `Clip` — need a non-nil receiver, so they **panic** on a `nil`
queue. Create the queue with `New` or `NewWithCap` before calling them.

## Concurrency

A `Queue` is **not safe for concurrent use**: it performs no internal locking.
If a queue is shared across goroutines, the caller must provide its own
synchronization (for example, a `sync.Mutex`).

## More examples and docs

Runnable, verified examples live in
[`example_test.go`](./example_test.go) — including breadth-first (level-order)
tree traversal — and are rendered alongside the API on the godoc page.

Benchmarks backing the "amortized O(1)" and "allocation-friendly" claims live in
[`bench_test.go`](./bench_test.go). Run them with:

```sh
go test -bench=. -benchmem ./queue/...
```

View the documentation locally:

```sh
go doc github.com/Nergous/maat/queue          # package overview
go doc github.com/Nergous/maat/queue Queue    # the Queue type and its methods
```

The same comments render on pkg.go.dev-style godoc.
