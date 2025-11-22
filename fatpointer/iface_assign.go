package main

import (
	"fmt"
	"unsafe"
)

type eface3 struct {
	Type uintptr
	Data uintptr
}

func demonstrateInterfaceAssignment() {
	fmt.Println("\n=== Interface Assignment Behavior ===\n")

	// 1. Assigning pointer to interface - the pointer VALUE is copied
	x := 42
	ptr := &x
	var iface interface{} = ptr

	fmt.Printf("x = 42, ptr = &x, iface = ptr\n")
	fmt.Printf("  ptr addr:  %p\n", ptr)
	fmt.Printf("  iface:     %v (same pointer value)\n\n", iface)

	// Changing what ptr points to doesn't affect iface
	y := 100
	ptr = &y
	fmt.Printf("After ptr = &y:\n")
	fmt.Printf("  ptr now:   %p (points to y)\n", ptr)
	fmt.Printf("  iface:     %v (still points to x)\n", iface)
	fmt.Printf("  *iface.(*int): %d\n\n", *iface.(*int))

	// But changing the value through the original pointer affects both
	x = 999
	fmt.Printf("After x = 999:\n")
	fmt.Printf("  *iface.(*int): %d (sees the change)\n\n", *iface.(*int))

	// 2. Assigning value type - a COPY is made
	fmt.Println("--- Value types are copied ---")
	val := 50
	var iface2 interface{} = val

	e := (*eface3)(unsafe.Pointer(&iface2))
	fmt.Printf("val = 50, iface2 = val\n")
	fmt.Printf("  val addr:   %p\n", &val)
	fmt.Printf("  iface Data: %#x (different address - it's a copy)\n\n", e.Data)

	val = 200
	fmt.Printf("After val = 200:\n")
	fmt.Printf("  val:    %d\n", val)
	fmt.Printf("  iface2: %d (unchanged - it's a copy)\n\n", iface2)

	// 3. To modify through interface, assign pointer
	fmt.Println("--- To modify through interface, use pointer ---")
	num := 10
	var iface3 interface{} = &num

	fmt.Printf("num = 10, iface3 = &num\n")
	*iface3.(*int) = 77
	fmt.Printf("After *iface3.(*int) = 77:\n")
	fmt.Printf("  num: %d (changed through interface)\n\n", num)

	// 4. Reassigning interface
	fmt.Println("--- Reassigning interface ---")
	var iface4 interface{} = 1
	e4 := (*eface3)(unsafe.Pointer(&iface4))
	fmt.Printf("iface4 = 1:       Type=%#x, Data=%#x\n", e4.Type, e4.Data)

	iface4 = "hello"
	fmt.Printf("iface4 = \"hello\": Type=%#x, Data=%#x\n", e4.Type, e4.Data)

	iface4 = []int{1, 2, 3}
	fmt.Printf("iface4 = []int:   Type=%#x, Data=%#x\n", e4.Type, e4.Data)
}

func init() {
	defer demonstrateInterfaceAssignment()
}
