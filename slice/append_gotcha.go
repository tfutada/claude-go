package main

import "fmt"

func main() {
	fmt.Println("=== Append may or may not reallocate ===")
	fmt.Println()
	printDiagram()

	// Create slice with len=3, cap=5
	a := make([]int, 3, 5)
	a[0], a[1], a[2] = 1, 2, 3

	fmt.Printf("Initial state:\n")
	fmt.Printf("a = %v, len=%d, cap=%d, addr=%p\n\n", a, len(a), cap(a), a)

	// Case 1: append fits in capacity - SHARES backing array
	fmt.Println("--- Case 1: append within capacity ---")
	b := append(a, 4) // len becomes 4, still under cap=5

	fmt.Printf("b := append(a, 4)\n")
	fmt.Printf("a = %v, addr=%p\n", a, a)
	fmt.Printf("b = %v, addr=%p\n\n", b, b)

	// Modify b[0] - affects a too!
	b[0] = 999
	fmt.Printf("After b[0] = 999:\n")
	fmt.Printf("a = %v  ← MODIFIED!\n", a)
	fmt.Printf("b = %v\n\n", b)

	// Reset
	a = make([]int, 3, 5)
	a[0], a[1], a[2] = 1, 2, 3

	// Case 2: append exceeds capacity - NEW backing array
	fmt.Println("--- Case 2: append exceeds capacity ---")
	c := append(a, 4, 5, 6) // needs 6 elements, cap=5 → reallocate

	fmt.Printf("c := append(a, 4, 5, 6)\n")
	fmt.Printf("a = %v, addr=%p\n", a, a)
	fmt.Printf("c = %v, addr=%p  ← different address!\n\n", c, c)

	// Modify c[0] - does NOT affect a
	c[0] = 888
	fmt.Printf("After c[0] = 888:\n")
	fmt.Printf("a = %v  ← unchanged\n", a)
	fmt.Printf("c = %v\n\n", c)

	// How to check if reallocation happened
	fmt.Println("=== Safe pattern: force new allocation ===\n")

	original := make([]int, 3, 5)
	original[0], original[1], original[2] = 1, 2, 3

	// Use full slice expression to limit capacity
	safe := original[0:3:3] // len=3, cap=3 (not 5!)

	fmt.Printf("original = %v, cap=%d\n", original, cap(original))
	fmt.Printf("safe := original[0:3:3], cap=%d\n\n", cap(safe))

	safe = append(safe, 4) // must reallocate since cap=3
	safe[0] = 777

	fmt.Printf("After safe = append(safe, 4); safe[0] = 777:\n")
	fmt.Printf("original = %v  ← unchanged (different backing array)\n", original)
	fmt.Printf("safe     = %v\n", safe)
}

func printDiagram() {
	diagram := `
┌─────────────────────────────────────────────────────────────────────────────┐
│  INITIAL STATE: a := make([]int, 3, 5); a = {1, 2, 3}                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Slice Header (stack)          Backing Array (heap)                         │
│  ┌─────────────────┐           ┌─────┬─────┬─────┬─────┬─────┐             │
│  │ ptr  ───────────────────────▶│  1  │  2  │  3  │  ?  │  ?  │             │
│  │ len = 3         │           └─────┴─────┴─────┴─────┴─────┘             │
│  │ cap = 5         │             [0]   [1]   [2]   [3]   [4]                │
│  └─────────────────┘                               ▲     ▲                  │
│         a                                     unused capacity               │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  CASE 1: b := append(a, 4)  ← FITS IN CAPACITY (cap=5)                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐           ┌─────┬─────┬─────┬─────┬─────┐             │
│  │ ptr  ───────────────────────▶│  1  │  2  │  3  │  4  │  ?  │             │
│  │ len = 3         │       ┌───▶└─────┴─────┴─────┴─────┴─────┘             │
│  │ cap = 5         │       │     [0]   [1]   [2]   [3]   [4]                │
│  └─────────────────┘       │                         ▲                      │
│         a                  │                    b added here                │
│                            │                                                │
│  ┌─────────────────┐       │    SAME backing array!                         │
│  │ ptr  ───────────────────┘    Modifying b[0] affects a[0]                 │
│  │ len = 4         │                                                        │
│  │ cap = 5         │                                                        │
│  └─────────────────┘                                                        │
│         b                                                                   │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  CASE 2: c := append(a, 4, 5, 6)  ← EXCEEDS CAPACITY (needs 6, cap=5)       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐           ┌─────┬─────┬─────┬─────┬─────┐             │
│  │ ptr  ───────────────────────▶│  1  │  2  │  3  │  ?  │  ?  │             │
│  │ len = 3         │           └─────┴─────┴─────┴─────┴─────┘             │
│  │ cap = 5         │             (old array, still used by a)              │
│  └─────────────────┘                                                        │
│         a                                                                   │
│                                                                             │
│  ┌─────────────────┐           ┌─────┬─────┬─────┬─────┬─────┬─────┬...    │
│  │ ptr  ───────────────────────▶│  1  │  2  │  3  │  4  │  5  │  6  │       │
│  │ len = 6         │           └─────┴─────┴─────┴─────┴─────┴─────┴...    │
│  │ cap = 10        │             (NEW array, ~2x growth)                    │
│  └─────────────────┘                                                        │
│         c                      DIFFERENT backing array!                     │
│                                Modifying c[0] does NOT affect a[0]          │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  SAFE PATTERN: s := a[0:3:3]  ← Full slice expression, cap = len            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐           ┌─────┬─────┬─────┬─────┬─────┐             │
│  │ ptr  ───────────────────────▶│  1  │  2  │  3  │  ?  │  ?  │             │
│  │ len = 3         │       ┌───▶└─────┴─────┴─────┴─────┴─────┘             │
│  │ cap = 5         │       │                                                │
│  └─────────────────┘       │                                                │
│         a                  │                                                │
│                            │                                                │
│  ┌─────────────────┐       │                                                │
│  │ ptr  ───────────────────┘    Same array BUT cap limited to 3             │
│  │ len = 3         │            Any append MUST reallocate!                 │
│  │ cap = 3  ← key! │                                                        │
│  └─────────────────┘                                                        │
│         s                                                                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
`
	fmt.Println(diagram)
}
