// Demonstrates 1000 goroutines running concurrently.
// Shows Go's lightweight goroutine model - each starts with ~2KB stack.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	const numGoroutines = 1000

	fmt.Printf("CPUs: %d, GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("Starting %d goroutines...\n\n", numGoroutines)

	start := time.Now()
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Go(func() {
			time.Sleep(100 * time.Millisecond) // simulate work
			if i%100 == 0 {
				fmt.Printf("[%6.0fms] G%d done\n", time.Since(start).Seconds()*1000, i)
			}
		})
	}

	fmt.Printf("All goroutines launched in %v\n", time.Since(start))
	fmt.Printf("NumGoroutine: %d\n\n", runtime.NumGoroutine())

	wg.Wait()

	fmt.Printf("\nAll %d goroutines completed in %v\n", numGoroutines, time.Since(start))
}
