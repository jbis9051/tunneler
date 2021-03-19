package main

import (
	"fmt"
	"net"
	"os"
)

const socks5Version = byte(5)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("listen address and tunnel address required")
		return
	}

	address := args[0]
	tunnelerAddr := args[1]

	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go func(conn net.Conn) {
			defer conn.Close()
			err := handleConnection(conn, tunnelerAddr)
			if err != nil {
				fmt.Printf("[ERR] connection failed: %v\n", err)
			}
		}(conn)
	}
}
