// Demonstrates preemption timing in Go.
// A CPU-bound goroutine will be preempted every ~10-20ms by sysmon.
//
// Run with: go run preempt.go
package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1) // Single P to force scheduling contention

	var counter atomic.Int64
	done := make(chan struct{})

	// CPU-bound goroutine (no cooperative yields)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				counter.Add(1)
			}
		}
	}()

	// Observer goroutine - can only run when CPU-bound one is preempted
	start := time.Now()
	lapStart := time.Now()
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Millisecond) // Will actually wait until preemption
		now := time.Now()
		lap := now.Sub(lapStart)
		fmt.Printf("[%6.0fms] Observer ran, lap: %6.2fms, counter: %d\n",
			time.Since(start).Seconds()*1000,
			lap.Seconds()*1000,
			counter.Load())
		lapStart = now
	}

	close(done)
}
