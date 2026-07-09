package deque_test

import (
	"fmt"

	"github.com/Nergous/maat/deque"
)

// A deque is double-ended: elements can be added and removed from both the
// front and the back.
func Example() {
	d := deque.New[int]()
	d.PushBack(2)
	d.PushFront(1)
	d.PushBack(3)

	for !d.IsEmpty() {
		v, _ := d.PopFront()
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
}

// NewWithCap preallocates the backing array. Use it when the final size is
// known up front to avoid reallocations during pushes.
func ExampleNewWithCap() {
	d := deque.NewWithCap[string](3)
	d.PushBack("a")
	d.PushBack("b")
	d.PushBack("c")

	fmt.Println(d.Len(), d.Cap())
	// Output: 3 3
}

// PopFront removes and returns the front element. The second value reports
// whether the deque was non-empty; on an empty deque it is false and the value
// is the zero value.
func ExampleDeque_PopFront() {
	d := deque.New[string]()
	d.PushBack("first")
	d.PushBack("last")

	v, ok := d.PopFront()
	fmt.Println(v, ok)

	d.PopFront() // removes "last"

	_, ok = d.PopFront() // deque is now empty
	fmt.Println(ok)
	// Output:
	// first true
	// false
}

// PopBack removes and returns the back element. It is useful when the deque is
// acting like a stack at one end.
func ExampleDeque_PopBack() {
	d := deque.New[string]()
	d.PushBack("first")
	d.PushBack("last")

	v, ok := d.PopBack()
	fmt.Println(v, ok)

	d.PopBack() // removes "first"

	_, ok = d.PopBack() // deque is now empty
	fmt.Println(ok)
	// Output:
	// last true
	// false
}

// PushFrontN adds several elements to the front in argument order, so the first
// argument becomes the new front.
func ExampleDeque_PushFrontN() {
	d := deque.New[int]()
	d.PushBackN(4, 5)
	d.PushFrontN(1, 2, 3)

	fmt.Println(d.Slice())
	// Output: [1 2 3 4 5]
}

// Front and Back inspect the ends without removing elements.
func ExampleDeque_Front() {
	d := deque.New[int]()
	d.PushBackN(10, 20, 30)

	front, _ := d.Front()
	back, _ := d.Back()

	fmt.Println(front, back)
	fmt.Println("len unchanged:", d.Len())
	// Output:
	// 10 30
	// len unchanged: 3
}

// PopFrontN removes up to n elements from the front in front-to-back order.
func ExampleDeque_PopFrontN() {
	d := deque.New[int]()
	d.PushBackN(1, 2, 3, 4)

	fmt.Println(d.PopFrontN(3))
	fmt.Println(d.Slice())
	// Output:
	// [1 2 3]
	// [4]
}

// PopBackN removes up to n elements from the back in removal order
// (back-to-front).
func ExampleDeque_PopBackN() {
	d := deque.New[int]()
	d.PushBackN(1, 2, 3, 4)

	fmt.Println(d.PopBackN(3))
	fmt.Println(d.Slice())
	// Output:
	// [4 3 2]
	// [1]
}

// All returns a range-over-func iterator over the elements from front to back
// without consuming them. The deque is left intact after iteration.
func ExampleDeque_All() {
	d := deque.New[int]()
	d.PushBackN(1, 2, 3)

	for v := range d.All() {
		fmt.Println(v)
	}
	fmt.Println("remaining:", d.Len())
	// Output:
	// 1
	// 2
	// 3
	// remaining: 3
}

// Clone returns an independent shallow copy. Mutating one deque does not affect
// the other, since they share no backing array.
func ExampleDeque_Clone() {
	a := deque.New[int]()
	a.PushBackN(1, 2, 3)

	b := a.Clone()
	b.PopFront() // removes 1 from b only
	b.PushBack(4)

	fmt.Println(a.Slice())
	fmt.Println(b.Slice())
	// Output:
	// [1 2 3]
	// [2 3 4]
}

// Slice returns a copy of the elements from front to back. The returned slice
// is independent of the deque.
func ExampleDeque_Slice() {
	d := deque.New[int]()
	d.PushBackN(1, 2, 3)

	out := d.Slice()
	fmt.Println(out)

	out[0] = 99
	front, _ := d.Front()
	fmt.Println("deque front:", front)

	empty := deque.New[int]()
	fmt.Println(empty.Slice() == nil)
	// Output:
	// [1 2 3]
	// deque front: 1
	// true
}

// Grow reserves capacity for more elements up front; Shrink and Clip both
// shrink the capacity back down to the current length. Shrink copies into a
// right-sized array and frees the old one immediately, whereas Clip may reslice
// without copying when live elements are already contiguous.
func ExampleDeque_Shrink() {
	d := deque.NewWithCap[int](1024)
	d.PushBackN(1, 2, 3)

	d.Grow(2000)
	fmt.Println(d.Cap() >= 2003)

	d.Shrink()
	fmt.Println(d.Len(), d.Cap())

	d.Clip()
	fmt.Println(d.Len(), d.Cap())
	// Output:
	// true
	// 3 3
	// 3 3
}

// Reset empties the deque but keeps the allocated backing array, so it can be
// reused without a new allocation.
func ExampleDeque_Reset() {
	d := deque.NewWithCap[int](8)
	d.PushBack(1)
	d.PushBack(2)

	d.Reset()
	fmt.Println(d.Len(), d.Cap())
	// Output: 0 8
}

// Clear, unlike Reset, releases the backing array: capacity drops back to zero
// and the memory becomes eligible for garbage collection.
func ExampleDeque_Clear() {
	d := deque.NewWithCap[int](8)
	d.PushBack(1)
	d.PushBack(2)

	d.Clear()
	fmt.Println(d.Len(), d.Cap())
	// Output: 0 0
}

// The nil zero value of *Deque[T] behaves as an empty deque for every read-only
// method and for empty-removal methods, so it can be queried safely without
// first calling New. Mutating methods still require a deque created with New.
func Example_nilReceiver() {
	var d *deque.Deque[int] // nil, never initialized

	fmt.Println(d.Len(), d.IsEmpty())

	v, ok := d.PopFront()
	fmt.Println(v, ok)
	// Output:
	// 0 true
	// 0 false
}

// A practical example: process urgent items from the front while normal items
// continue to queue at the back.
func Example_urgentQueue() {
	tasks := deque.New[string]()
	tasks.PushBack("write tests")
	tasks.PushBack("run benchmarks")
	tasks.PushFront("fix build")

	for !tasks.IsEmpty() {
		task, _ := tasks.PopFront()
		fmt.Println(task)
	}
	// Output:
	// fix build
	// write tests
	// run benchmarks
}
