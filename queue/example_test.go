package queue_test

import (
	"fmt"

	"github.com/Nergous/maat/queue"
)

// A queue is FIFO (first in, first out): elements come out in the same order
// they went in.
func Example() {
	q := queue.New[int]()
	q.Push(1)
	q.Push(2)
	q.Push(3)

	for !q.IsEmpty() {
		v, _ := q.Pop()
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
}

// NewWithCap preallocates the backing array. Use it when the final size is
// known up front to avoid reallocations during Push.
func ExampleNewWithCap() {
	q := queue.NewWithCap[string](3)
	q.Push("a")
	q.Push("b")
	q.Push("c")

	fmt.Println(q.Len(), q.Cap())
	// Output: 3 3
}

// Pop removes and returns the front element (the oldest). The second value
// reports whether the queue was non-empty; on an empty queue it is false and
// the value is the zero value.
func ExampleQueue_Pop() {
	q := queue.New[string]()
	q.Push("first")
	q.Push("last")

	v, ok := q.Pop()
	fmt.Println(v, ok)

	q.Pop() // removes "last"

	_, ok = q.Pop() // queue is now empty
	fmt.Println(ok)
	// Output:
	// first true
	// false
}

// PopN removes up to n elements from the front and returns them in FIFO order.
func ExampleQueue_PopN() {
	q := queue.New[int]()
	q.PushN(1, 2, 3, 4)

	fmt.Println(q.PopN(3))
	fmt.Println(q.Slice())
	// Output:
	// [1 2 3]
	// [4]
}

// Peek returns the front element without removing it.
func ExampleQueue_Peek() {
	q := queue.New[int]()
	q.Push(10)
	q.Push(20)

	front, ok := q.Peek()
	fmt.Println(front, ok)
	fmt.Println("len unchanged:", q.Len())
	// Output:
	// 10 true
	// len unchanged: 2
}

// Reset empties the queue but keeps the allocated backing array, so it can be
// reused without a new allocation.
func ExampleQueue_Reset() {
	q := queue.NewWithCap[int](8)
	q.Push(1)
	q.Push(2)

	q.Reset()
	fmt.Println(q.Len(), q.Cap())
	// Output: 0 8
}

// Clear, unlike Reset, releases the backing array: capacity drops back to zero
// and the memory becomes eligible for garbage collection.
func ExampleQueue_Clear() {
	q := queue.NewWithCap[int](8)
	q.Push(1)
	q.Push(2)

	q.Clear()
	fmt.Println(q.Len(), q.Cap())
	// Output: 0 0
}

// A practical example: breadth-first (level-order) traversal of a binary tree.
// A queue visits nodes in arrival order, so each level is processed left to
// right before the next level begins.
func Example_breadthFirstTraversal() {
	type node struct {
		value       int
		left, right *node
	}

	//        1
	//       / \
	//      2   3
	//     / \   \
	//    4   5   6
	tree := &node{
		value: 1,
		left: &node{
			value: 2,
			left:  &node{value: 4},
			right: &node{value: 5},
		},
		right: &node{
			value: 3,
			right: &node{value: 6},
		},
	}

	q := queue.New[*node]()
	q.Push(tree)

	for !q.IsEmpty() {
		n, _ := q.Pop()
		fmt.Println(n.value)
		if n.left != nil {
			q.Push(n.left)
		}
		if n.right != nil {
			q.Push(n.right)
		}
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
}
