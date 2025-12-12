// Demonstrates 1000 goroutines doing CPU-intensive work (fibonacci).
// Shows how GOMAXPROCS limits true parallelism for CPU-bound tasks.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// fib calculates fibonacci recursively (intentionally slow, CPU-bound)
func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	runtime.GOMAXPROCS(1) // Force single processor

	const numGoroutines = 1000
	const fibN = 30 // ~5ms per call on typical CPU

	fmt.Printf("CPUs: %d, GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("Starting %d goroutines computing fib(%d)...\n\n", numGoroutines, fibN)

	start := time.Now()
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Go(func() {
			result := fib(fibN)
			if i%100 == 0 {
				fmt.Printf("[%6.0fms] G%d: fib(%d) = %d\n",
					time.Since(start).Seconds()*1000, i, fibN, result)
			}
		})
	}

	fmt.Printf("All goroutines launched in %v\n", time.Since(start))
	fmt.Printf("NumGoroutine: %d\n\n", runtime.NumGoroutine())

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\nAll %d goroutines completed in %v\n", numGoroutines, elapsed)
	fmt.Printf("Throughput: %.0f fib/sec\n", float64(numGoroutines)/elapsed.Seconds())
}
