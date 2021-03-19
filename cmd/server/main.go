package main

import (
	"fmt"
)

const tunnelerVersion = 1

func main() {
	_, err := start(":8081")
	if err != nil {
		fmt.Println(err)
		return
	}
}
