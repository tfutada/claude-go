// Package main demonstrates Go channel fundamentals.
//
// Channels are typed conduits for communication between goroutines.
// They provide synchronization and data transfer in one mechanism.
//
// Delivery Analogy:
// - Unbuffered = Hand-to-hand delivery (no mailbox)
//   Sender waits until receiver takes the package directly.
//   Both must be present at the same time.
//
// - Buffered = Delivery with mailbox (capacity N)
//   Sender drops package in mailbox and leaves (if not full).
//   Receiver picks up later. No need to meet.
//
// Key concepts:
// - Unbuffered channels block until both sender and receiver are ready
// - Buffered channels block only when buffer is full (send) or empty (recv)
// - Closing a channel signals no more values will be sent
// - Select enables waiting on multiple channel operations
package main

import (
	"fmt"
	"time"
)

func main() {
	basicChannel()
	bufferedChannel()
	channelDirections()
	selectDemo()
	rangeOverChannel()
}

// basicChannel shows unbuffered channel synchronization.
// Hand-to-hand delivery: sender and receiver must meet.
func basicChannel() {
	fmt.Println("=== Basic Unbuffered Channel ===")

	ch := make(chan string) // no mailbox - hand-to-hand only

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- "hello" // blocks until receiver is ready
	}()

	msg := <-ch // blocks until sender sends
	fmt.Println("Received:", msg)
}

// bufferedChannel shows non-blocking sends until buffer is full.
// Mailbox delivery: drop and go (until mailbox is full).
func bufferedChannel() {
	fmt.Println("\n=== Buffered Channel ===")

	ch := make(chan int, 3) // mailbox holds 3 packages

	// These don't block - buffer has space
	ch <- 1
	ch <- 2
	ch <- 3
	// ch <- 4 would block here (buffer full)

	fmt.Println("Buffer len:", len(ch), "cap:", cap(ch))

	// Receive all
	fmt.Println(<-ch, <-ch, <-ch)
}

// channelDirections shows send-only and receive-only channel types.
func channelDirections() {
	fmt.Println("\n=== Channel Directions ===")

	ch := make(chan int)

	go sender(ch)   // ch is send-only in sender
	receiver(ch)    // ch is receive-only in receiver
}

func sender(ch chan<- int) {   // send-only
	ch <- 42
}

func receiver(ch <-chan int) { // receive-only
	fmt.Println("Received:", <-ch)
}

// selectDemo shows multiplexing multiple channels.
func selectDemo() {
	fmt.Println("\n=== Select Statement ===")

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "from ch1"
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "from ch2"
	}()

	// Wait for both messages
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Println(msg)
		case msg := <-ch2:
			fmt.Println(msg)
		}
	}
}

// rangeOverChannel shows iterating until channel is closed.
func rangeOverChannel() {
	fmt.Println("\n=== Range Over Channel ===")

	ch := make(chan int)

	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i
		}
		close(ch) // signal no more values
	}()

	// Range exits when channel is closed
	for v := range ch {
		fmt.Println("Got:", v)
	}

	// Check if channel is closed
	v, ok := <-ch
	fmt.Printf("After close: value=%d, ok=%v\n", v, ok)
}
