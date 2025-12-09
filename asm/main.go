package main

import "fmt"

//go:noinline
func add(a, b int64) int64 {
	return a + b
}

//go:noinline
func sum(nums []int64) int64 {
	var total int64
	for _, n := range nums {
		total += n
	}
	return total
}

func main() {
	fmt.Println(add(10, 20))
	fmt.Println(sum([]int64{1, 2, 3, 4, 5}))
}
