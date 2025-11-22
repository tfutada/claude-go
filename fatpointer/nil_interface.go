package main

import (
	"fmt"
	"unsafe"
)

type eface2 struct {
	Type uintptr
	Data uintptr
}

func demonstrateNilInterface() {
	fmt.Println("\n=== Nil Interface Internals ===\n")

	// 1. True nil interface - both pointers are nil
	var nilIface interface{}
	e1 := (*eface2)(unsafe.Pointer(&nilIface))
	fmt.Printf("var nilIface interface{}\n")
	fmt.Printf("  Type: %#x, Data: %#x\n", e1.Type, e1.Data)
	fmt.Printf("  nilIface == nil: %v\n\n", nilIface == nil)

	// 2. Interface holding nil pointer - has type, nil data
	var ptr *int = nil
	var ifaceWithNil interface{} = ptr
	e2 := (*eface2)(unsafe.Pointer(&ifaceWithNil))

	// ptr = 9 is type error - can't assign int to *int
	// Even if we do: nine := 9; ptr = &nine
	// ifaceWithNil still holds the OLD nil pointer (it's a copy)
	nine := 9.8
	nilIface = &nine // this doesn't affect ifaceWithNil!

	fmt.Printf("var ptr *int = nil; var iface interface{} = ptr\n")
	fmt.Printf("  Type: %#x, Data: %#x\n", e2.Type, e2.Data)
	fmt.Printf("  iface == nil: %v  <- SURPRISE!\n\n", ifaceWithNil == nil)

	// 3. This is the famous Go nil interface trap
	fmt.Println("--- The Nil Interface Trap ---")

	// Function returns concrete nil
	err := getErrorBad()
	e3 := (*eface2)(unsafe.Pointer(&err))
	fmt.Printf("\nerr := getErrorBad() // returns (*myError)(nil)\n")
	fmt.Printf("  Type: %#x, Data: %#x\n", e3.Type, e3.Data)
	fmt.Printf("  err == nil: %v\n", err == nil)
	fmt.Printf("  err.(*myError) == nil: %v\n\n", err.(*myError) == nil)

	// Correct way
	err2 := getErrorGood()
	e4 := (*eface2)(unsafe.Pointer(&err2))
	fmt.Printf("err2 := getErrorGood() // returns nil properly\n")
	fmt.Printf("  Type: %#x, Data: %#x\n", e4.Type, e4.Data)
	fmt.Printf("  err2 == nil: %v\n\n", err2 == nil)

	// 4. Visual comparison
	fmt.Println("--- Summary ---")
	fmt.Println("Nil interface:        Type=0x0, Data=0x0  -> == nil is TRUE")
	fmt.Println("Interface with nil:   Type=0xN, Data=0x0  -> == nil is FALSE")
}

type myError struct{}

func (e *myError) Error() string { return "error" }

func getErrorBad() error {
	var err *myError = nil
	return err // returns (*myError)(nil), not nil
}

func getErrorGood() error {
	var err *myError = nil
	if err == nil {
		return nil // return actual nil interface
	}
	return err
}

func init() {
	defer demonstrateNilInterface()
}
