package main

import (
	"fmt"
	"unsafe"
)

// Bad: fields ordered poorly - wastes memory due to padding
type BadOrder struct {
	a bool // 1 byte
	// 7 bytes padding to align next field
	b int64 // 8 bytes
	c bool  // 1 byte
	// 7 bytes padding (to align next struct)
}

// Good: fields ordered by size (largest first) - minimal padding
type GoodOrder struct {
	b int64 // 8 bytes
	a bool  // 1 byte
	c bool  // 1 byte
	// 6 bytes padding
}

// Real-world example: User struct
type UserBad struct {
	Active   bool   // 1 + 7 padding
	ID       int64  // 8
	Verified bool   // 1 + 3 padding
	Age      int32  // 4
	Name     string // 16 (string header: ptr + len)
	Admin    bool   // 1 + 7 padding
}

type UserGood struct {
	ID       int64  // 8
	Name     string // 16
	Age      int32  // 4
	Active   bool   // 1
	Verified bool   // 1
	Admin    bool   // 1
	// 1 padding
}

func main() {
	fmt.Println("=== Struct Alignment Demo ===\n")

	// Simple example
	fmt.Println("Simple struct:")
	fmt.Printf("  BadOrder  size: %d bytes\n", unsafe.Sizeof(BadOrder{}))
	fmt.Printf("  GoodOrder size: %d bytes\n", unsafe.Sizeof(GoodOrder{}))
	fmt.Printf("  Savings: %d bytes (%.0f%%)\n\n",
		unsafe.Sizeof(BadOrder{})-unsafe.Sizeof(GoodOrder{}),
		float64(unsafe.Sizeof(BadOrder{})-unsafe.Sizeof(GoodOrder{}))/float64(unsafe.Sizeof(BadOrder{}))*100)

	// Real-world example
	fmt.Println("User struct:")
	fmt.Printf("  UserBad  size: %d bytes\n", unsafe.Sizeof(UserBad{}))
	fmt.Printf("  UserGood size: %d bytes\n", unsafe.Sizeof(UserGood{}))
	fmt.Printf("  Savings: %d bytes (%.0f%%)\n\n",
		unsafe.Sizeof(UserBad{})-unsafe.Sizeof(UserGood{}),
		float64(unsafe.Sizeof(UserBad{})-unsafe.Sizeof(UserGood{}))/float64(unsafe.Sizeof(UserBad{}))*100)

	// Show field offsets for BadOrder
	fmt.Println("BadOrder field offsets:")
	fmt.Printf("  a (bool)  at offset %d\n", unsafe.Offsetof(BadOrder{}.a))
	fmt.Printf("  b (int64) at offset %d\n", unsafe.Offsetof(BadOrder{}.b))
	fmt.Printf("  c (bool)  at offset %d\n", unsafe.Offsetof(BadOrder{}.c))

	// Show field offsets for GoodOrder
	fmt.Println("\nGoodOrder field offsets:")
	fmt.Printf("  b (int64) at offset %d\n", unsafe.Offsetof(GoodOrder{}.b))
	fmt.Printf("  a (bool)  at offset %d\n", unsafe.Offsetof(GoodOrder{}.a))
	fmt.Printf("  c (bool)  at offset %d\n", unsafe.Offsetof(GoodOrder{}.c))

	// Memory impact with slices
	fmt.Println("\n=== Memory Impact with 1 Million Items ===")
	const count = 1_000_000
	badSize := uint64(unsafe.Sizeof(BadOrder{})) * count
	goodSize := uint64(unsafe.Sizeof(GoodOrder{})) * count
	fmt.Printf("  []BadOrder:  %.2f MB\n", float64(badSize)/1024/1024)
	fmt.Printf("  []GoodOrder: %.2f MB\n", float64(goodSize)/1024/1024)
	fmt.Printf("  Savings:     %.2f MB\n", float64(badSize-goodSize)/1024/1024)
}
