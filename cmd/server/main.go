package main

import (
	"fmt"
)

func main() {
	_, err := start(":8081")
	if err != nil {
		fmt.Println(err)
		return
	}
}
