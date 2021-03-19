package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("listen address required")
		return
	}
	_, err := start(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
}
