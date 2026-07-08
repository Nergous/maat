# Roadmap

This roadmap describes the public direction for `maat`. It is based on the
detailed internal planning in [`todo.md`](./todo.md), but keeps the GitHub-facing
view shorter and easier to scan.

The order below is a priority signal, not a release promise. The project should
stay dependency-free, idiomatic, and easy to audit.

## Current packages

- `stack`: generic slice-backed LIFO stack.
- `queue`: generic slice-backed FIFO queue.
- `set`: generic map-backed set with unordered iteration and set algebra.

Before the first tagged release, these packages should be treated as the API
baseline for naming, nil behavior, docs, examples, tests, and benchmarks.

## Shared conventions

New packages should follow the same house style:

- `New` and `NewWithCap` where capacity hints are honest.
- `Of` and `From` where literal or slice seeding is natural.
- nil-safe reads: nil receivers behave like empty containers for read-only
  methods.
- comma-ok lookups and removals where an operation can miss.
- `All()` as a non-consuming range-over-func iterator.
- `Slice()` snapshots where a flat element snapshot is meaningful.
- `Clone()` as an independent shallow copy.
- `Clear()` releases backing storage; `Reset()` empties while keeping storage
  for reuse when the backing can support it.
- no internal synchronization; callers wrap containers when sharing them across
  goroutines.
- no external runtime dependencies.

## Near-term

- Stabilize the public API for `stack`, `queue`, and `set`.
- Keep docs, examples, tests, and benchmarks synchronized across packages.
- Add and maintain CI coverage for vetting, tests, and the race detector.
- Add a root README lifecycle matrix for `Clear`, `Reset`, resizing methods, and
  package-specific omissions.
- Review in-place set algebra naming before release. The internal plan prefers
  the `*With` family (`UnionWith`, `IntersectWith`, `DifferenceWith`,
  `SymmetricDifferenceWith`) for consistency with future set-like packages.
- Decide the first tagged release policy once the initial package set feels
  stable.

## Priority P0

These are the most likely next packages after the current baseline stabilizes:

| Package      | Type        | Summary |
| ------------ | ----------- | ------- |
| `binheap`    | `Heap[T]`   | Generic binary heap, min-first by default; type-safe `container/heap` replacement. |
| `bitset`     | `Bitset`    | Dense auto-growing bit vector with word-parallel set algebra. |
| `deque`      | `Deque[T]`  | Growable ring-buffer double-ended queue; O(1) push/pop at both ends. |
| `orderedmap` | `Map[K, V]` | Insertion-order preserving hash map. |

## Priority P1

Useful general-purpose containers that should follow after P0:

| Package      | Type                    | Summary |
| ------------ | ----------------------- | ------- |
| `bloom`      | `Filter[T]`             | Generic Bloom filter via `maphash.Comparable`. |
| `counter`    | `Counter[T]`            | Map-backed multiset / frequency counter. |
| `iheap`      | `IndexedHeap[K, P]`     | Updatable priority queue keyed by unique values. |
| `list`       | `List[T]`, `Element[T]` | Generic doubly linked list with element handles and splicing. |
| `lru`        | `Cache[K, V]`           | Fixed-capacity LRU cache with O(1) operations and eviction callback. |
| `multimap`   | `Multimap[K, V]`        | One key to many values, slice-valued, per-key insertion order. |
| `orderedset` | `Set[T]`                | Insertion-order preserving set. |
| `pqueue`     | `PriorityQueue[E, P]`   | Element/priority queue, built as a thin API over heap behavior. |
| `ring`       | `Ring[T]`               | Fixed-capacity circular FIFO buffer. |
| `sortedmap`  | `SortedMap[K, V]`       | Key-ordered map with navigation, rank/select, and ranges. |
| `sortedset`  | `SortedSet[T]`          | Ordered set with algebra and navigation. |
| `trie`       | `Trie[V]`               | String-keyed path-compressed prefix tree. |
| `unionfind`  | `Dense`                 | Int-indexed disjoint-set. |

## Priority P2

More specialized structures, still in scope when there is a clear API and use
case:

| Package     | Type                | Summary |
| ----------- | ------------------- | ------- |
| `bimap`     | `BiMap[K, V]`       | Bidirectional 1:1 hash map with live inverse view. |
| `fenwick`   | `Tree[T]`           | Fenwick tree for prefix/range sums. |
| `interval`  | `Tree[E, V]`        | Interval tree for stab and overlap queries. |
| `lfu`       | `Cache[K, V]`       | Fixed-capacity LFU cache. |
| `mmheap`    | `MinMaxHeap[T]`     | Double-ended priority queue. |
| `multimap`  | `SetMultimap[K, V]` | Value-deduplicating multimap. |
| `segtree`   | `Tree[T]`           | Segment tree over a caller-supplied monoid. |
| `sparseset` | `SparseSet`         | Sparse set of non-negative ints with O(1) reset. |
| `table`     | `Table[R, C, V]`    | Two-keyed map: row and column to value. |
| `unionfind` | `Keyed[T]`          | Map-backed disjoint-set over comparable keys. |

## Documentation

- Add root-level examples that compare packages side by side.
- Keep godoc comments precise enough to stand alone on pkg.go.dev.
- Document intentional omissions, especially when a package lacks `Cap`, `Grow`,
  `Shrink`, `Clip`, `Clear`, or `Reset`.
- Document performance trade-offs without overpromising exact runtime behavior.

## Considered and rejected

These are intentionally not planned as standalone public packages right now:

- `multiset.Multiset`: merged into `counter.Counter`.
- Skip list: overlaps with sorted B-tree structures, with worse cache locality
  for the single-goroutine use cases here.
- Singly linked list: no clear win over `list.List`, slices, stack, or queue.
- Radix tree as a separate package: folded into `trie` as an implementation
  choice.
- Histogram: statistics rather than a core container; discrete frequencies are
  covered by `counter`.
- d-ary heap: implementation/performance knob for heap-like packages, not a
  separate public type for now.
- TTL caches, TinyLFU, ARC: separate cache families for later consideration, not
  options on `lru`/`lfu`.
- `concurrent/*`: a separate future track, deliberately excluded from this
  single-goroutine roadmap.

## Deliverables per package

Every new package should ship with:

- package godoc covering purpose, complexity, nil-safety, iteration order,
  memory behavior, and concurrency note;
- package `README.md`;
- root `README.md` package-table row;
- runnable `example_test.go`;
- full unit tests, including nil-receiver reads and mutator panics;
- focused benchmarks using `b.Loop()`, `b.ReportAllocs()`, and package-level
  sinks;
- no external runtime dependencies.

## Non-goals

- Replacing Go slices, maps, or the standard library.
- Adding broad utility packages without a concrete container abstraction.
- Adding synchronization internally.
- Adding formatting-only APIs such as `String()` unless a package has a stable,
  deterministic representation.
- Adding external dependencies for core data structures.

## Contributing

Good contributions are small, documented, and tested. New packages should start
with the public API shape, edge-case behavior, examples, tests, and benchmarks,
not just implementation code.
