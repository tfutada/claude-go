// https://medium.com/@matrixorigin-database/optimizing-golang-performance-1-memory-related-dafff15b955a
// Part 1: Compiler Memory Escape Analysis
//
// Run with: go build -gcflags="-m" ./escape/
// This shows escape analysis output

package main

import "fmt"

//go:noinline
func makeBuffer() []byte {
	return make([]byte, 1024) // escapes to heap - returned from non-inlined function
}

func main() {
	buf := makeBuffer()
	for i := range buf {
		buf[i] = buf[i] + 1
	}
	fmt.Printf("%v\n", buf)
}
