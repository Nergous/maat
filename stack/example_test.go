package stack_test

import (
	"fmt"

	"github.com/Nergous/maat/stack"
)

// A stack is LIFO (last in, first out): elements come out in the reverse
// order they went in.
func Example() {
	s := stack.New[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	for !s.IsEmpty() {
		v, _ := s.Pop()
		fmt.Println(v)
	}
	// Output:
	// 3
	// 2
	// 1
}

// NewWithCap preallocates the backing array. Use it when the final size is
// known up front to avoid reallocations during Push.
func ExampleNewWithCap() {
	s := stack.NewWithCap[string](3)
	s.Push("a")
	s.Push("b")
	s.Push("c")

	fmt.Println(s.Len(), s.Cap())
	// Output: 3 3
}

// Pop removes and returns the top element. The second value reports whether
// the stack was non-empty; on an empty stack it is false and the value is the
// zero value.
func ExampleStack_Pop() {
	s := stack.New[string]()
	s.Push("first")
	s.Push("last")

	v, ok := s.Pop()
	fmt.Println(v, ok)

	s.Pop() // removes "first"

	_, ok = s.Pop() // stack is now empty
	fmt.Println(ok)
	// Output:
	// last true
	// false
}

// Peek returns the top element without removing it.
func ExampleStack_Peek() {
	s := stack.New[int]()
	s.Push(10)
	s.Push(20)

	top, ok := s.Peek()
	fmt.Println(top, ok)
	fmt.Println("len unchanged:", s.Len())
	// Output:
	// 20 true
	// len unchanged: 2
}

// All returns a range-over-func iterator over the elements from top to bottom
// without consuming them. The stack is left intact after iteration.
func ExampleStack_All() {
	s := stack.New[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	for v := range s.All() {
		fmt.Println(v)
	}
	fmt.Println("remaining:", s.Len())
	// Output:
	// 3
	// 2
	// 1
	// remaining: 3
}

// PushN adds several elements at once, in argument order, so the last argument
// ends up on top. It is equivalent to calling Push for each value but grows the
// backing array at most once.
func ExampleStack_PushN() {
	s := stack.New[int]()
	s.PushN(1, 2, 3) // 3 ends up on top

	for !s.IsEmpty() {
		v, _ := s.Pop()
		fmt.Println(v)
	}
	// Output:
	// 3
	// 2
	// 1
}

// Clone returns an independent shallow copy. Mutating one stack does not affect
// the other, since they share no backing array.
func ExampleStack_Clone() {
	a := stack.New[int]()
	a.PushN(1, 2, 3)

	b := a.Clone()
	b.Pop() // removes 3 from b only

	fmt.Println(a.Len(), b.Len())

	top, _ := a.Peek()
	fmt.Println("a top:", top)
	// Output:
	// 3 2
	// a top: 3
}

// Slice returns a copy of the elements from top to bottom (LIFO) — the same
// order All yields. The returned slice is independent of the stack.
func ExampleStack_Slice() {
	s := stack.New[int]()
	s.PushN(1, 2, 3)

	out := s.Slice()
	fmt.Println(out)

	// The slice is detached: modifying it leaves the stack untouched.
	out[0] = 99
	top, _ := s.Peek()
	fmt.Println("stack top:", top)

	// An empty stack returns nil.
	empty := stack.New[int]()
	fmt.Println(empty.Slice() == nil)
	// Output:
	// [3 2 1]
	// stack top: 3
	// true
}

// Grow reserves capacity for more elements up front; Shrink and Clip both
// shrink the capacity back down to the current length. Shrink copies into a
// right-sized array and frees the old one immediately, whereas Clip only
// reslices in O(1) and defers reclaiming memory to the next growth.
func ExampleStack_Shrink() {
	s := stack.NewWithCap[int](1024)
	s.PushN(1, 2, 3)

	s.Grow(2000) // reserve room for 2000 more elements
	fmt.Println(s.Cap() >= 2003)

	s.Shrink() // copy into a right-sized array, free the old one now
	fmt.Println(s.Len(), s.Cap())

	s.Clip() // already tight: no-op
	fmt.Println(s.Len(), s.Cap())
	// Output:
	// true
	// 3 3
	// 3 3
}

// Reset empties the stack but keeps the allocated backing array, so it can be
// reused without a new allocation.
func ExampleStack_Reset() {
	s := stack.NewWithCap[int](8)
	s.Push(1)
	s.Push(2)

	s.Reset()
	fmt.Println(s.Len(), s.Cap())
	// Output: 0 8
}

// Clear, unlike Reset, releases the backing array: capacity drops back to zero
// and the memory becomes eligible for garbage collection.
func ExampleStack_Clear() {
	s := stack.NewWithCap[int](8)
	s.Push(1)
	s.Push(2)

	s.Clear()
	fmt.Println(s.Len(), s.Cap())
	// Output: 0 0
}

// The nil zero value of *Stack[T] behaves as an empty stack for every
// read-only method, so it can be queried safely without first calling New.
// Mutating methods (Push, PushN, Clear, Reset, Grow, Shrink, Clip) still
// require a stack created with New.
func Example_nilReceiver() {
	var s *stack.Stack[int] // nil, never initialized

	fmt.Println(s.Len(), s.IsEmpty())

	v, ok := s.Pop()
	fmt.Println(v, ok)
	// Output:
	// 0 true
	// 0 false
}

// A practical example: checking whether brackets are balanced. Push each
// opening bracket; on a closing one, pop and verify the pair matches.
func Example_balancedBrackets() {
	balanced := func(input string) bool {
		pairs := map[rune]rune{')': '(', ']': '[', '}': '{'}
		st := stack.New[rune]()

		for _, r := range input {
			switch r {
			case '(', '[', '{':
				st.Push(r)
			case ')', ']', '}':
				top, ok := st.Pop()
				if !ok || top != pairs[r] {
					return false
				}
			}
		}
		return st.IsEmpty()
	}

	fmt.Println(balanced("([]{})"))
	fmt.Println(balanced("([)]"))
	// Output:
	// true
	// false
}
