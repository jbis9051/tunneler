package main

import (
	"fmt"
	"net"
)

const socks5Version = byte(5)

func main() {
	address := ":8080"
	tunnelerAddr := "192.168.1.186:8081"

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
