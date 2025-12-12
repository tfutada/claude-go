// Benchmark comparing goroutine throughput across different GOMAXPROCS and goroutine counts.
// Creates a matrix showing how parallelism scales with CPU cores.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// fib calculates fibonacci recursively (CPU-bound)
func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func benchmark(numGoroutines, maxProcs, fibN int) time.Duration {
	runtime.GOMAXPROCS(maxProcs)

	var wg sync.WaitGroup
	start := time.Now()

	for range numGoroutines {
		wg.Go(func() {
			fib(fibN)
		})
	}

	wg.Wait()
	return time.Since(start)
}

func main() {
	const fibN = 30

	numCPU := runtime.NumCPU()
	goroutineCounts := []int{100, 500, 1000, 2000}

	// Build proc counts: 1, 2, 4, ..., numCPU, numCPU+1
	var procCounts []int
	for p := 1; p <= numCPU; p *= 2 {
		procCounts = append(procCounts, p)
	}
	if procCounts[len(procCounts)-1] != numCPU {
		procCounts = append(procCounts, numCPU)
	}
	procCounts = append(procCounts, numCPU+1) // CPU cores + 1

	fmt.Printf("CPU Cores: %d, fib(%d)\n\n", numCPU, fibN)

	// Header
	fmt.Printf("%-12s", "Goroutines")
	for _, p := range procCounts {
		fmt.Printf("%12s", fmt.Sprintf("P=%d", p))
	}
	fmt.Printf("%12s\n", "Speedup")
	fmt.Println(string(make([]byte, 12+12*len(procCounts)+12)))

	// Benchmark matrix
	for _, g := range goroutineCounts {
		fmt.Printf("%-12d", g)

		var singleCore time.Duration
		var maxCore time.Duration

		for i, p := range procCounts {
			elapsed := benchmark(g, p, fibN)
			fmt.Printf("%10dms", elapsed.Milliseconds())

			if i == 0 {
				singleCore = elapsed
			}
			maxCore = elapsed
		}

		// Speedup ratio (single core vs max cores)
		speedup := float64(singleCore) / float64(maxCore)
		fmt.Printf("%10.2fx\n", speedup)
	}

	fmt.Println()
	fmt.Println("Speedup = P=1 time / P=max time (ideal = num cores)")
}
