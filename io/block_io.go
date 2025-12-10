package main

import (
	"fmt"
	"os"
)

// BlockSize is typical filesystem/disk block size
const BlockSize = 4096 // 4KB

func main() {
	file, err := os.OpenFile("data.bin", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	//defer os.Remove("data.bin") // cleanup

	// Block addressing: block number × block size = byte offset
	blockNum := int64(10) // block #10
	offset := blockNum * BlockSize

	// Write a block at specific offset (no seek needed)
	writeBuf := make([]byte, BlockSize)
	copy(writeBuf, []byte("hello block device style I/O"))
	_, err = file.WriteAt(writeBuf, offset)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote 4KB at offset", offset)

	// Read the block back (thread-safe, no shared file pointer)
	readBuf := make([]byte, BlockSize)
	_, err = file.ReadAt(readBuf, offset)
	if err != nil {
		panic(err)
	}
	fmt.Println("Read data:", string(readBuf[:32]))
}
