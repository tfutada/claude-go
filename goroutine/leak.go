// Goroutine memory leak examples - goroutines that never exit.
//
// A goroutine leak occurs when a goroutine is started but never terminates.
// Each leaked goroutine holds memory (~2KB stack minimum) forever.
//
// Common causes:
// - Blocked on channel with no sender/receiver
// - Waiting on channel that is never closed
// - Infinite loop without exit condition
// - Missing cancellation mechanism
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	leakBlockedSend()
	leakBlockedReceive()
	leakForgottenWorker()
	fixedWithDoneChannel()
}

// leakBlockedSend: goroutine blocks forever trying to send.
func leakBlockedSend() {
	fmt.Println("=== LEAK: Blocked Send ===")
	before := runtime.NumGoroutine()

	ch := make(chan int) // unbuffered

	go func() {
		ch <- 42 // blocks forever - no receiver!
		fmt.Println("This never prints")
	}()

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("Goroutines: %d -> %d (leaked: %d)\n\n", before, after, after-before)
	// ch is never read, goroutine stuck forever
}

// leakBlockedReceive: goroutine blocks forever waiting to receive.
func leakBlockedReceive() {
	fmt.Println("=== LEAK: Blocked Receive ===")
	before := runtime.NumGoroutine()

	ch := make(chan int)

	go func() {
		val := <-ch // blocks forever - no sender, never closed!
		fmt.Println("Got:", val)
	}()

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("Goroutines: %d -> %d (leaked: %d)\n\n", before, after, after-before)
	// ch is never written or closed, goroutine stuck forever
}

// leakForgottenWorker: long-running worker without stop mechanism.
func leakForgottenWorker() {
	fmt.Println("=== LEAK: Forgotten Worker ===")
	before := runtime.NumGoroutine()

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			// No way to stop this!
		}
	}()

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("Goroutines: %d -> %d (leaked: %d)\n\n", before, after, after-before)
}

// fixedWithDoneChannel: proper cancellation prevents leak.
func fixedWithDoneChannel() {
	fmt.Println("=== FIXED: With Done Channel ===")
	before := runtime.NumGoroutine()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				fmt.Println("Worker: exiting cleanly")
				return
			default:
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)
	during := runtime.NumGoroutine()

	close(done) // signal worker to stop
	time.Sleep(50 * time.Millisecond)

	after := runtime.NumGoroutine()
	fmt.Printf("Goroutines: %d -> %d -> %d (no leak)\n\n", before, during, after)
}
