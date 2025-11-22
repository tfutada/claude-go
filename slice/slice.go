package main

import "fmt"

func main() {
	fmt.Println("=== Slice Mutation Demo ===\n")

	// 1. Basic mutation through slice
	fmt.Println("1. Basic mutation:")
	original := []int{1, 2, 3, 4, 5}
	slice := original[1:3] // [2, 3]

	fmt.Printf("   original: %v\n", original)
	fmt.Printf("   slice:    %v\n", slice)

	slice[0] = 99 // Mutates underlying array!

	fmt.Println("   After slice[0] = 99:")
	fmt.Printf("   original: %v  ← Changed!\n", original)
	fmt.Printf("   slice:    %v\n\n", slice)

	// 2. Append surprise
	fmt.Println("2. Append overwrites original:")
	original2 := []int{1, 2, 3, 4, 5}
	slice2 := original2[1:3] // len=2, cap=4

	fmt.Printf("   original2: %v\n", original2)
	fmt.Printf("   slice2:    %v (len=%d, cap=%d)\n", slice2, len(slice2), cap(slice2))

	slice2 = append(slice2, 99) // Has room, overwrites original2[3]!

	fmt.Println("   After append(slice2, 99):")
	fmt.Printf("   original2: %v  ← Surprise!\n", original2)
	fmt.Printf("   slice2:    %v\n\n", slice2)

	// 3. Safe way: copy
	fmt.Println("3. Safe way with copy:")
	original3 := []int{1, 2, 3, 4, 5}
	slice3 := make([]int, 2)
	copy(slice3, original3[1:3]) // Independent copy

	fmt.Printf("   original3: %v\n", original3)
	fmt.Printf("   slice3:    %v (independent)\n", slice3)

	slice3[0] = 99
	slice3 = append(slice3, 88)

	fmt.Println("   After mutations:")
	fmt.Printf("   original3: %v  ← Unchanged!\n", original3)
	fmt.Printf("   slice3:    %v\n\n", slice3)

	// 4. String is safe (immutable)
	fmt.Println("4. String is safe (immutable):")
	s := "hello"
	s2 := s[1:4] // "ell"

	fmt.Printf("   s:  %q\n", s)
	fmt.Printf("   s2: %q\n", s2)
	fmt.Println("   Cannot modify s2's bytes → s stays safe")
	// s2[0] = 'X'  // Compile error: cannot assign to s2[0]
}
