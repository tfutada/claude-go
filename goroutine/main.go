// Package main demonstrates goroutine scheduler basics.
//
// Goroutines are lightweight threads managed by the Go runtime using
// the GMP model (G=Goroutine, M=Machine/OS thread, P=Processor).
//
// Key points:
// - Goroutines start with ~2KB stack (vs 1-8MB for OS threads)
// - M:N scheduling: M goroutines multiplexed onto N OS threads
// - Work-stealing scheduler distributes work across P processors
// - GOMAXPROCS controls the number of P (logical processors)
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	runtime.GOMAXPROCS(2) // Force only 2 parallel executions

	start := time.Now()
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Go(func() {
			fmt.Printf("[%6.0fms] G%d start\n", time.Since(start).Seconds()*1000, i)
			time.Sleep(100 * time.Millisecond) // simulate work
			fmt.Printf("[%6.0fms] G%d end\n", time.Since(start).Seconds()*1000, i)
		})
	}
	wg.Wait()
}
