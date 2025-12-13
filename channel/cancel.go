// Cancel channel patterns - signaling goroutines to stop.
//
// The "done channel" pattern uses close() to broadcast cancellation.
// Closing a channel wakes ALL receivers simultaneously.
//
// Analogy:
// - Fire alarm: one signal, everyone evacuates
// - close(done) = pull the alarm
// - <-done = hear the alarm and exit
package main

import (
	"fmt"
	"time"
)

func init() {
	// Register demos to run from main
	cancelDemos = []func(){
		doneChannelBasic,
		cancelMultipleWorkers,
		cancelWithCleanup,
	}
}

var cancelDemos []func()

// doneChannelBasic shows simple cancellation pattern.
func doneChannelBasic() {
	fmt.Println("\n=== Done Channel Basic ===")

	done := make(chan struct{}) // empty struct = zero memory

	go func() {
		for {
			select {
			case <-done:
				fmt.Println("Worker: received cancel, exiting")
				return
			default:
				fmt.Println("Worker: working...")
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	time.Sleep(120 * time.Millisecond)
	close(done) // signal cancellation
	time.Sleep(10 * time.Millisecond) // let worker print exit message
}

// cancelMultipleWorkers shows broadcasting cancel to many goroutines.
func cancelMultipleWorkers() {
	fmt.Println("\n=== Cancel Multiple Workers ===")

	done := make(chan struct{})
	workerDone := make(chan int, 3) // collect exit confirmations

	// Start 3 workers
	for i := 1; i <= 3; i++ {
		go func(id int) {
			for {
				select {
				case <-done:
					fmt.Printf("Worker %d: stopping\n", id)
					workerDone <- id
					return
				default:
					time.Sleep(30 * time.Millisecond)
				}
			}
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Main: sending cancel signal")
	close(done) // ONE close wakes ALL workers

	// Wait for all workers to confirm exit
	for i := 0; i < 3; i++ {
		id := <-workerDone
		fmt.Printf("Main: worker %d confirmed exit\n", id)
	}
}

// cancelWithCleanup shows graceful shutdown with cleanup.
func cancelWithCleanup() {
	fmt.Println("\n=== Cancel With Cleanup ===")

	done := make(chan struct{})
	cleaned := make(chan struct{})

	go func() {
		defer close(cleaned) // signal cleanup complete

		// Simulate holding resources
		fmt.Println("Worker: acquired resources")

		<-done // wait for cancel

		// Cleanup
		fmt.Println("Worker: releasing resources...")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Worker: cleanup complete")
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)  // signal stop
	<-cleaned    // wait for cleanup to finish
	fmt.Println("Main: worker fully stopped")
}
