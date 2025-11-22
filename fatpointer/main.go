package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("=== Fat Pointer Examples in Go ===\n")

	// 1. Thin pointer (regular pointer)
	var x int = 42
	var ptr *int = &x
	fmt.Printf("Thin Pointer (*int):\n")
	fmt.Printf("  Size: %d bytes\n", unsafe.Sizeof(ptr))
	fmt.Printf("  Value: %d\n\n", *ptr)

	// 2. Slice - Fat pointer with ptr + len + cap
	slice := []int{1, 2, 3, 4, 5}
	fmt.Printf("Slice (fat pointer):\n")
	fmt.Printf("  Size: %d bytes (ptr + len + cap)\n", unsafe.Sizeof(slice))
	fmt.Printf("  Len: %d, Cap: %d\n", len(slice), cap(slice))

	// Internal structure of slice
	type sliceHeader struct {
		Data uintptr
		Len  int
		Cap  int
	}
	sh := (*sliceHeader)(unsafe.Pointer(&slice))
	fmt.Printf("  Internal: Data=%#x, Len=%d, Cap=%d\n\n", sh.Data, sh.Len, sh.Cap)

	// 3. String - Fat pointer with ptr + len
	str := "Hello, Go!"
	fmt.Printf("String (fat pointer):\n")
	fmt.Printf("  Size: %d bytes (ptr + len)\n", unsafe.Sizeof(str))
	fmt.Printf("  Len: %d\n", len(str))

	// Internal structure of string
	type stringHeader struct {
		Data uintptr
		Len  int
	}
	strh := (*stringHeader)(unsafe.Pointer(&str))
	fmt.Printf("  Internal: Data=%#x, Len=%d\n\n", strh.Data, strh.Len)

	// 4. Interface - Fat pointer with type + data
	var iface interface{} = 123
	fmt.Printf("Interface (fat pointer):\n")
	fmt.Printf("  Size: %d bytes (type + data)\n", unsafe.Sizeof(iface))

	// Internal structure of interface
	type ifaceHeader struct {
		Type uintptr // pointer to type info
		Data uintptr // pointer to data
	}
	ih := (*ifaceHeader)(unsafe.Pointer(&iface))
	fmt.Printf("  Internal: Type=%#x, Data=%#x\n\n", ih.Type, ih.Data)

	// 5. Demonstrate slice reslicing (sharing underlying array)
	fmt.Println("=== Slice Reslicing Demo ===")
	original := make([]int, 5, 10)
	for i := range original {
		original[i] = i * 10
	}

	sub := original[1:3]
	fmt.Printf("Original: %v (len=%d, cap=%d)\n", original, len(original), cap(original))
	fmt.Printf("Sub[1:3]: %v (len=%d, cap=%d)\n", sub, len(sub), cap(sub))

	// Modifying sub affects original (same underlying array)
	sub[0] = 999
	fmt.Printf("After sub[0]=999:\n")
	fmt.Printf("  Original: %v\n", original)
	fmt.Printf("  Sub:      %v\n\n", sub)

	// 6. Interface type assertion cost
	fmt.Println("=== Interface Type Info ===")
	var animal interface{}

	animal = Dog{Name: "Buddy"}
	printAnimalInfo(animal)

	animal = Cat{Name: "Whiskers"}
	printAnimalInfo(animal)
}

type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return "Woof!"
}

type Cat struct {
	Name string
}

func (c Cat) Speak() string {
	return "Meow!"
}

type Speaker interface {
	Speak() string
}

func printAnimalInfo(v interface{}) {
	// Type switch uses the type pointer in the fat pointer
	switch animal := v.(type) {
	case Dog:
		fmt.Printf("Dog: %s says %s\n", animal.Name, animal.Speak())
	case Cat:
		fmt.Printf("Cat: %s says %s\n", animal.Name, animal.Speak())
	}
}
