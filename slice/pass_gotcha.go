// Slice pass-by-value gotcha - modifying slice inside function.
//
// Slice header (ptr, len, cap) is copied when passed to function.
// - Modifying elements: visible to caller (same backing array)
// - Append: NOT visible to caller (len changed only in copy)
//
// Run: go run ./slice/pass_gotcha.go
package main

import "fmt"

func main() {
	printDiagram()
	gotchaModifyElement()
	gotchaAppend()
	fixReturnSlice()
	fixPointerToSlice()
}

// gotchaModifyElement: modifying elements IS visible to caller.
func gotchaModifyElement() {
	fmt.Println("=== Gotcha 1: Modify Element (visible) ===")

	nums := []int{1, 2, 3}
	fmt.Printf("Before: %v\n", nums)

	modifyFirst(nums)
	fmt.Printf("After:  %v  ← modified!\n\n", nums)
}

func modifyFirst(s []int) {
	s[0] = 999 // modifies backing array, caller sees it
}

// gotchaAppend: append is NOT visible to caller.
func gotchaAppend() {
	fmt.Println("=== Gotcha 2: Append (NOT visible) ===")

	nums := []int{1, 2, 3}
	fmt.Printf("Before: %v, len=%d\n", nums, len(nums))

	appendOne(nums)
	fmt.Printf("After:  %v, len=%d  ← unchanged!\n\n", nums, len(nums))
}

func appendOne(s []int) {
	s = append(s, 4) // s is local copy, caller's len unchanged
	fmt.Printf("Inside: %v, len=%d\n", s, len(s))
}

// fixReturnSlice: return the modified slice.
func fixReturnSlice() {
	fmt.Println("=== Fix 1: Return Slice ===")

	nums := []int{1, 2, 3}
	fmt.Printf("Before: %v\n", nums)

	nums = appendAndReturn(nums) // reassign!
	fmt.Printf("After:  %v  ← fixed!\n\n", nums)
}

func appendAndReturn(s []int) []int {
	s = append(s, 4)
	return s // caller must use returned value
}

// fixPointerToSlice: pass pointer to slice header.
func fixPointerToSlice() {
	fmt.Println("=== Fix 2: Pointer to Slice ===")

	nums := []int{1, 2, 3}
	fmt.Printf("Before: %v\n", nums)

	appendWithPointer(&nums)
	fmt.Printf("After:  %v  ← fixed!\n\n", nums)
}

func appendWithPointer(s *[]int) {
	*s = append(*s, 4) // modify caller's slice header
}

func printDiagram() {
	diagram := `
┌─────────────────────────────────────────────────────────────────────────────┐
│  SLICE PASS-BY-VALUE GOTCHA                                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Caller:                         Function receives COPY of header:          │
│  ┌─────────────────┐             ┌─────────────────┐                        │
│  │ ptr  ─────────────────┐   ┌───│ ptr (copied)    │                        │
│  │ len = 3         │     │   │   │ len = 3 (copied)│                        │
│  │ cap = 3         │     │   │   │ cap = 3 (copied)│                        │
│  └─────────────────┘     │   │   └─────────────────┘                        │
│       nums               │   │        s (local)                             │
│                          ▼   ▼                                              │
│                        ┌─────┬─────┬─────┐                                  │
│                        │  1  │  2  │  3  │  (shared backing array)          │
│                        └─────┴─────┴─────┘                                  │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  s[0] = 999  →  Modifies backing array  →  Caller sees change ✓            │
├─────────────────────────────────────────────────────────────────────────────┤
│  s = append(s, 4)  →  Updates LOCAL s.len  →  Caller's len unchanged ✗     │
│                                                                             │
│  Caller:                         Function (after append):                   │
│  ┌─────────────────┐             ┌─────────────────┐                        │
│  │ ptr  ───────────────────┐ ┌───│ ptr             │                        │
│  │ len = 3  ← still 3!     │ │   │ len = 4  ← only local!                   │
│  │ cap = 3                 │ │   │ cap = ?         │                        │
│  └─────────────────┘       │ │   └─────────────────┘                        │
│                            ▼ ▼                                              │
│                          ┌─────┬─────┬─────┬─────┐                          │
│                          │  1  │  2  │  3  │  4  │                          │
│                          └─────┴─────┴─────┴─────┘                          │
│                                             ▲                               │
│                                        Caller can't see this!               │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  FIXES:                                                                     │
│    1. Return slice:    nums = appendFunc(nums)                              │
│    2. Pointer:         func appendFunc(s *[]int) { *s = append(*s, 4) }     │
└─────────────────────────────────────────────────────────────────────────────┘
`
	fmt.Println(diagram)
}
