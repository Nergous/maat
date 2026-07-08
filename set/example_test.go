package set_test

import (
	"fmt"
	"slices"

	"github.com/Nergous/maat/set"
)

// A set stores each value at most once.
func Example() {
	s := set.New[int]()
	s.Add(1)
	s.Add(2)
	s.Add(1)

	fmt.Println(s.Len())
	fmt.Println(s.Contains(2))
	fmt.Println(s.Contains(3))
	// Output:
	// 2
	// true
	// false
}

// Of builds a set from explicit values, removing duplicates.
func ExampleOf() {
	s := set.Of("go", "maat", "go")

	out := s.Slice()
	slices.Sort(out)
	fmt.Println(out)
	// Output: [go maat]
}

// Union returns a new set containing all values from both sets.
func ExampleSet_Union() {
	a := set.Of(1, 2, 3)
	b := set.Of(3, 4)

	out := a.Union(b).Slice()
	slices.Sort(out)
	fmt.Println(out)
	// Output: [1 2 3 4]
}

// Intersection returns a new set containing values present in every input set.
func ExampleSet_Intersection() {
	a := set.Of(1, 2, 3)
	b := set.Of(2, 3, 4)

	out := a.Intersection(b).Slice()
	slices.Sort(out)
	fmt.Println(out)
	// Output: [2 3]
}

// All returns a range-over-func iterator. Set iteration order is unspecified,
// so sort collected values when stable output matters.
func ExampleSet_All() {
	s := set.Of(3, 1, 2)

	var out []int
	for v := range s.All() {
		out = append(out, v)
	}
	slices.Sort(out)
	fmt.Println(out)
	// Output: [1 2 3]
}

// A practical example: de-duplicate a slice while keeping membership checks
// explicit.
func Example_deduplicate() {
	seen := set.New[string]()
	for _, name := range []string{"alice", "bob", "alice", "cora"} {
		seen.Add(name)
	}

	out := seen.Slice()
	slices.Sort(out)
	fmt.Println(out)
	// Output: [alice bob cora]
}
