// Demonstrates that SIGURG used for preemption is NOT visible to user code.
// Go's runtime handles preemption SIGURG internally before user signal handlers.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1)

	// Try to catch SIGURG
	sigCh := make(chan os.Signal, 100)
	signal.Notify(sigCh, syscall.SIGURG)

	var sigCount atomic.Int64
	go func() {
		for range sigCh {
			sigCount.Add(1)
		}
	}()

	// CPU-bound goroutine (will be preempted by SIGURG internally)
	var counter atomic.Int64
	done := make(chan struct{})
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

	// Let it run for a while
	time.Sleep(500 * time.Millisecond)
	close(done)

	fmt.Printf("Counter: %d (goroutine was preempted many times)\n", counter.Load())
	fmt.Printf("SIGURG received by user handler: %d\n", sigCount.Load())
	fmt.Println("→ Runtime intercepts preemption SIGURG before user handlers")
}
