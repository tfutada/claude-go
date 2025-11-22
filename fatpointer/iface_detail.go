package main

import (
	"fmt"
	"unsafe"
)

// Go runtime internal structure (simplified)
// type iface struct {
//     tab  *itab   // type info + method table
//     data unsafe.Pointer
// }
//
// type eface struct {  // empty interface
//     _type *_type  // type metadata
//     data  unsafe.Pointer
// }

type eface struct {
	Type uintptr // pointer to runtime._type
	Data uintptr // pointer to data
}

func demonstrateInterfaceInternals() {
	fmt.Println("\n=== Interface Internals ===\n")

	// Small types (<=8 bytes) - value stored directly or in static memory
	var i1 interface{} = int(42)
	var i2 interface{} = int64(100)
	var i3 interface{} = true

	// Larger types - pointer to heap
	var i4 interface{} = [100]int{1, 2, 3}
	var i5 interface{} = "hello world"

	// Struct
	type Person struct {
		Name string
		Age  int
	}
	var i6 interface{} = Person{Name: "Alice", Age: 30}

	// All interfaces are 16 bytes
	fmt.Printf("Size of interface{} containing:\n")
	fmt.Printf("  int:       %d bytes\n", unsafe.Sizeof(i1))
	fmt.Printf("  int64:     %d bytes\n", unsafe.Sizeof(i2))
	fmt.Printf("  bool:      %d bytes\n", unsafe.Sizeof(i3))
	fmt.Printf("  [100]int:  %d bytes\n", unsafe.Sizeof(i4))
	fmt.Printf("  string:    %d bytes\n", unsafe.Sizeof(i5))
	fmt.Printf("  struct:    %d bytes\n", unsafe.Sizeof(i6))

	fmt.Println("\n--- Type pointer comparison ---")

	// Same type = same type pointer
	var a interface{} = int(1)
	var b interface{} = int(2)
	var c interface{} = int64(1)

	ea := (*eface)(unsafe.Pointer(&a))
	eb := (*eface)(unsafe.Pointer(&b))
	ec := (*eface)(unsafe.Pointer(&c))

	fmt.Printf("int(1)  -> Type: %#x, Data: %#x\n", ea.Type, ea.Data)
	fmt.Printf("int(2)  -> Type: %#x, Data: %#x\n", eb.Type, eb.Data)
	fmt.Printf("int64(1)-> Type: %#x, Data: %#x\n", ec.Type, ec.Data)
	fmt.Printf("\nSame type (int vs int): %v\n", ea.Type == eb.Type)
	fmt.Printf("Different type (int vs int64): %v\n", ea.Type == ec.Type)

	fmt.Println("\n--- What type pointer contains ---")
	fmt.Println("runtime._type includes:")
	fmt.Println("  - size of the type")
	fmt.Println("  - alignment")
	fmt.Println("  - hash (for type comparison)")
	fmt.Println("  - kind (int, struct, ptr, etc.)")
	fmt.Println("  - string representation")
	fmt.Println("  - garbage collection info")

	fmt.Println("\n--- Interface with methods (iface vs eface) ---")

	var empty interface{} = myInt(42)
	var stringer fmt.Stringer = myInt(42)

	fmt.Printf("empty interface{}:  %d bytes\n", unsafe.Sizeof(empty))
	fmt.Printf("Stringer interface: %d bytes\n", unsafe.Sizeof(stringer))
	fmt.Println("Both 16 bytes, but non-empty interface has itab with method table")
}

type myInt int

func (m myInt) String() string {
	return fmt.Sprintf("myInt(%d)", m)
}

func init() {
	// Register to run after main examples
	defer demonstrateInterfaceInternals()
}
