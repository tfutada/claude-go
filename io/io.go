package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("input.txt")
	if err != nil {
		fmt.Println(err)
	}

	var totalCount int

	for {
		data := make([]byte, 10)

		count, err := f.Read(data)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				break
			}
			break
		}
		println(count)
		totalCount += count
	}

	var reader io.Reader = strings.NewReader(" テストデータ")
	var readCloser io.ReadCloser = io.NopCloser(reader)

	fmt.Printf("count %v\n", totalCount)
	fmt.Printf("%#v", readCloser)
}

// in Rust, you need to provide a buffer to receive file contents, but no need to specify the size.
// let mut contents = String::new();
//    file.read_to_string(&mut contents)?;
